package services

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/gasser707/go-gql-server/cloud"
	"github.com/gasser707/go-gql-server/custom"
	dbModels "github.com/gasser707/go-gql-server/databases/models"
	"github.com/gasser707/go-gql-server/graph/model"
	"github.com/twinj/uuid"
	"github.com/volatiletech/sqlboiler/queries"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/sync/errgroup"
)

type ImagesServiceInterface interface {
	UploadImages(ctx context.Context, input []*model.NewImageInput) ([]*custom.Image, error)
	DeleteImages(ctx context.Context, input []*model.DeleteImageInput) (bool, error)
	GetImages(ctx context.Context, input *model.ImageFilterInput) ([]*custom.Image, error)
	GetImageById(ctx context.Context, ID string) (*custom.Image, error)
	UpdateImage(ctx context.Context, input *model.UpdateImageInput)(*custom.Image, error)
}

//UsersService implements the usersServiceInterface
var _ ImagesServiceInterface = &imagesService{}

type imagesService struct {
	DB              *sql.DB
	AuthService     AuthServiceInterface
	storageOperator cloud.StorageOperatorInterface
}

func NewImagesService(db *sql.DB, authSrv AuthServiceInterface, storageOperator cloud.StorageOperatorInterface) *imagesService {
	return &imagesService{DB: db, AuthService: authSrv, storageOperator: storageOperator}
}

func (s *imagesService) UploadImages(ctx context.Context, input []*model.NewImageInput) ([]*custom.Image, error) {
	userId, _, err := s.AuthService.validateCredentials(ctx)
	if err != nil {
		return nil, err
	}
	errs, ctx := errgroup.WithContext(ctx)
	ch := make(chan *custom.Image)
	for _, inputImg := range input {
		i := inputImg
		errs.Go(
			func() error {
				return s.processUploadImage(ctx, ch, i, userId)
			})
	}
	go func() {
		errs.Wait()
		close(ch)
	}()

	// read from channel as they come in until its closed
	images := []*custom.Image{}
	for img := range ch {
		images = append(images, img)
	}

	return images, errs.Wait()
}

func (s *imagesService) DeleteImages(ctx context.Context, input []*model.DeleteImageInput) (bool, error) {
	userId, _, err := s.AuthService.validateCredentials(ctx)
	if err != nil {
		return false, err
	}
	errs, ctx := errgroup.WithContext(ctx)
	for _, delImg := range input {
		i := delImg
		errs.Go(
			func() error {
				return s.processDeleteImage(ctx, i, userId)
			})
	}
	err = errs.Wait()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *imagesService) processUploadImage(ctx context.Context, ch chan *custom.Image, inputImg *model.NewImageInput,
	userId intUserID) (err error) {

	url, err := s.storageOperator.UploadImage(ctx, &inputImg.File, uuid.NewV4().String(), fmt.Sprintf("%v", userId))
	if err != nil {
		return err
	}
	dbImg := dbModels.Image{
		Title: inputImg.Title, Description: inputImg.Description,
		Private: inputImg.Private,
		ForSale: inputImg.ForSale, Price: inputImg.Price,
		UserID: int(userId), URL: url,
	}
	err = dbImg.Insert(ctx, s.DB, boil.Infer())
	if err != nil {
		return err
	}
	err = s.insertLabels(inputImg.Labels, dbImg.ID)
	if err != nil {
		return err
	}

	image := &custom.Image{
		ID: fmt.Sprintf("%v", dbImg.ID), Title: dbImg.Title, Description: dbImg.Description,
		URL: dbImg.URL, Private: dbImg.Private,
		ForSale: dbImg.ForSale, Price: dbImg.Price, UserID: fmt.Sprintf("%v", userId),
		Created: &dbImg.CreatedAt, Labels: inputImg.Labels,
	}

	ch <- image
	return nil
}

func (s *imagesService) processDeleteImage(ctx context.Context, input *model.DeleteImageInput,
	userId intUserID) (err error) {

	delImgId, _ := strconv.Atoi(input.ID)
	img, err := dbModels.FindImage(ctx, s.DB, delImgId)
	if err != nil {
		return err
	}
	if img.UserID != int(userId) {
		return fmt.Errorf("an image from the list doesn't belong to you")
	}
	err = s.storageOperator.DeleteImage(ctx, img.URL)
	if err != nil {
		return err
	}
	_, err = img.Delete(ctx, s.DB)
	if err != nil {
		return err
	}

	return nil
}

func (s *imagesService) GetImages(ctx context.Context, input *model.ImageFilterInput) ([]*custom.Image, error) {
	userId, _, err := s.AuthService.validateCredentials(ctx)
	if err != nil {
		return nil, err
	}
	if(input == nil ){
		return s.GetAllPublicImgs(ctx)
	}

	if input.ID != nil {
		img, err := s.GetImageById(ctx, *input.ID)
		if err != nil {
			return nil, err
		}
		return []*custom.Image{img}, nil
	}

	filter:= parseFilter(input, userId)
	return s.GetImagesByFilter(ctx, userId, filter)

}

func (s *imagesService) GetImageById(ctx context.Context, ID string) (*custom.Image, error) {
	userId, _, err := s.AuthService.validateCredentials(ctx)
	if err != nil {
		return nil, err
	}
	inputId, err := strconv.Atoi(ID)
	if err != nil {
		return nil, err
	}

	img, err := dbModels.FindImage(ctx, s.DB, inputId)
	if err != nil || (img.UserID!=int(userId) && img.Private) {
		return nil, fmt.Errorf("image not available")
	}

	ls, err := img.Labels().All(ctx, s.DB)
	if err != nil {
		return nil, err
	}
	labels := labelSliceToString(ls)

	return &custom.Image{
		ID: ID, UserID: fmt.Sprintf("%v", img.UserID), Created: &img.CreatedAt,
		Title: img.Title, URL: img.URL, Description: img.Description, Private: img.Private,
		ForSale: img.ForSale, Price: img.Price, Labels: labels,
	}, nil
}

func (s *imagesService) GetImagesByFilter(ctx context.Context, userID intUserID, filter string) ([]*custom.Image, error) {
	var dbImgs  dbModels.ImageSlice
	err := queries.Raw(filter).Bind(ctx, s.DB, &dbImgs)
	if err != nil {
		return nil, err
	}
	imgList:= []*custom.Image{}
	for _,img := range dbImgs{
		ls, err := img.Labels().All(ctx, s.DB)
		if err != nil {
			return nil, err
		}
		labels := labelSliceToString(ls)
		imgList = append(imgList, &custom.Image{
			ID: fmt.Sprintf("%v",img.ID), UserID: fmt.Sprintf("%v", img.UserID), Created: &img.CreatedAt,
			Title: img.Title, URL: img.URL, Description: img.Description, Private: img.Private,
			ForSale: img.ForSale, Price: img.Price, Labels: labels,
		})
	}
	return imgList,nil
}

func (s *imagesService) GetAllPublicImgs(ctx context.Context) ([]*custom.Image, error) {

	dbImgs, err := dbModels.Images(Where("images.private = False")).All(ctx, s.DB)
	if err != nil {
		return nil, err
	}
	imgList:= []*custom.Image{}
	for _,img := range dbImgs{
		ls, err := img.Labels().All(ctx, s.DB)
		if err != nil {
			return nil, err
		}
		labels := labelSliceToString(ls)
		imgList = append(imgList, &custom.Image{
			ID: fmt.Sprintf("%v",img.ID), UserID: fmt.Sprintf("%v", img.UserID), Created: &img.CreatedAt,
			Title: img.Title, URL: img.URL, Description: img.Description, Private: img.Private,
			ForSale: img.ForSale, Price: img.Price, Labels: labels,
		})
	}
	return imgList,nil
}

func (s *imagesService) UpdateImage(ctx context.Context, input *model.UpdateImageInput) (*custom.Image, error){
	userId, _, err := s.AuthService.validateCredentials(ctx)
	if err != nil {
		return nil, err
	}

	img, err:= dbModels.Images(Where("id=? And images.user_Id=?",input.ID, userId )).One(ctx, s.DB)
	if err != nil {
		return nil, fmt.Errorf("image doesn't exist or doesn't belong to you")
	}

	img.Title = input.Title
	img.ForSale = input.ForSale
	img.Private = input.Private
	img.Description = input.Description
	img.Price = input.Price

	_,err = img.Update(ctx,s.DB,boil.Infer())
	if err != nil {
		return nil, err
	}
	
	_,err = dbModels.Labels(Where("image_id=?",img.ID)).DeleteAll(ctx,s.DB)
	if err != nil {
		return nil, err
	}
	err = s.insertLabels(input.Labels, img.ID)
	if err != nil {
		return nil, err
	}	
	return &custom.Image{
		Title: img.Title,
		Description: img.Description,
		ForSale: img.ForSale,
		Private: img.Private,
		Labels: input.Labels,
		UserID: fmt.Sprintf("%v",img.UserID),
		Price: img.Price,
		ID: input.ID,		
	}, nil
}

func labelSliceToString(slice dbModels.LabelSlice) []string {
	strArr := []string{}
	for _, l := range slice {
		strArr = append(strArr, l.Tag)
	}
	return strArr
}


func parseFilter(input *model.ImageFilterInput, userID intUserID) string{
	queryStr := []string{}
	filterStart := ""
	filterStr:=""
	filterAdded:=false
	if(input.Labels!=nil && len(input.Labels)>0){
		filterStr = "select images.id, created_at, url, description, user_id, title, price, forSale, private From labels join images on images.id = labels.image_id where "
		filterStr = filterStr+ "labels.tag in "+ parseLabels(input.Labels)
		queryStr = append(queryStr, filterStr)
	}else{
		filterStart = "select * from images where "
	}
	if input.UserID!=nil{
		filterStr = "images.user_id = "+ *input.UserID
		if(fmt.Sprintf("%v", userID)!= *input.UserID){
			filterStr = filterStr+"And images.private= False"
		}
		queryStr = append(queryStr, filterStr)
		filterAdded=true
	}
	if input.UserID!=nil && input.Private !=nil && fmt.Sprintf("%v", userID) == *input.UserID{
		filterStr = "images.private = "+ fmt.Sprintf("%t", *input.Private)
		queryStr = append(queryStr, filterStr)
		filterAdded=true
	}
	if(input.ForSale!=nil){
		filterStr= "images.forSale= "+ fmt.Sprintf("%t", *input.ForSale)
		queryStr = append(queryStr, filterStr)
		filterAdded=true
	}
	if(input.PriceLimit!=nil){
		filterStr= "images.price<= " +fmt.Sprintf("%v", input.PriceLimit)
		queryStr = append(queryStr, filterStr)
		filterAdded=true
	}
	if(input.Title!=nil){
		filterStr= "images.title= "+ fmt.Sprintf("'%s'", *input.Title)
		queryStr = append(queryStr, filterStr)
		filterAdded=true
	}
	if(!filterAdded){
		return filterStart+"images.private=False"
	}

	return filterStart + strings.Join(queryStr[:], " And ")
}

func parseLabels(labels []string) string{
	str:=""
	for _, l:= range labels{
	str= str+ fmt.Sprintf("'%s',", l)
	}

	str = str[0:len(str)-1]
	str= "("+ str+")"
	return str
}

func (s *imagesService) insertLabels(labels []string, imgId int) error{
	insertStr:="insert into labels (tag,image_id) values "
	for _, tag := range labels {
		insertStr= insertStr+fmt.Sprintf("('%s', %v),",tag, imgId)
	}
	insertStr = insertStr[0:len(insertStr)-1]
	_,err:=s.DB.Exec(insertStr)
	if err != nil {
		return err
	}
	return nil
}