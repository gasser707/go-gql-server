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
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/gasser707/go-gql-server/graph/model"
	"github.com/gasser707/go-gql-server/helpers"
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
	UpdateImage(ctx context.Context, input *model.UpdateImageInput) (*custom.Image, error)
	AutoGenerateLabels(ctx context.Context, imageId string) ([]string, error)
}

//UsersService implements the usersServiceInterface
var _ ImagesServiceInterface = &imagesService{}

type imagesService struct {
	DB              *sql.DB
	AuthService     AuthServiceInterface
	storageOperator cloud.StorageOperatorInterface
	visionOperator  cloud.VisionOperatorInterface
}

func NewImagesService(ctx context.Context, db *sql.DB, authSrv AuthServiceInterface, storageOperator cloud.StorageOperatorInterface) *imagesService {
	vo, err := cloud.NewVisionOperator(ctx)
	if err != nil {
		panic(err)
	}
	return &imagesService{DB: db, AuthService: authSrv, storageOperator: storageOperator, visionOperator: vo}
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
		return customErr.Internal(ctx, err.Error())
	}
	err = s.insertLabels(ctx, inputImg.Labels, dbImg.ID)
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
		return customErr.Internal(ctx, err.Error())
	}
	if img.UserID != int(userId) {
		return customErr.Forbidden(ctx, err.Error())
	}
	c, err := img.Sales().Count(ctx, s.DB)
	if err != nil {
		return customErr.Internal(ctx, err.Error())
	}
	if c != 0 {
		img.Archived = true
		img.Update(ctx, s.DB, boil.Infer())
		return nil
	}
	err = s.storageOperator.DeleteImage(ctx, img.URL)
	if err != nil {
		return err
	}
	_, err = img.Delete(ctx, s.DB)
	if err != nil {
		return customErr.Internal(ctx, err.Error())
	}

	return nil
}

func (s *imagesService) GetImages(ctx context.Context, input *model.ImageFilterInput) ([]*custom.Image, error) {
	userId, _, err := s.AuthService.validateCredentials(ctx)
	if err != nil {
		return nil, err
	}
	if input == nil {
		return s.GetAllPublicImgs(ctx)
	}

	if input.ID != nil {
		img, err := s.GetImageById(ctx, *input.ID)
		if err != nil {
			return nil, err
		}
		return []*custom.Image{img}, nil
	}

	filter := helpers.ParseFilter(input, int(userId))
	return s.GetImagesByFilter(ctx, userId, filter)

}

func (s *imagesService) GetImageById(ctx context.Context, ID string) (*custom.Image, error) {
	userId, _, err := s.AuthService.validateCredentials(ctx)
	if err != nil {
		return nil, err
	}
	inputId, err := strconv.Atoi(ID)
	if err != nil {
		return nil, customErr.Internal(ctx, err.Error())
	}

	img, err := dbModels.FindImage(ctx, s.DB, inputId)
	if err != nil || (img.UserID != int(userId) && img.Private) {
		return nil, customErr.Forbidden(ctx, err.Error())
	}

	ls, err := img.Labels().All(ctx, s.DB)
	if err != nil {
		return nil, customErr.Internal(ctx, err.Error())
	}
	labels := helpers.LabelSliceToString(ls)

	return &custom.Image{
		ID: ID, UserID: fmt.Sprintf("%v", img.UserID), Created: &img.CreatedAt,
		Title: img.Title, URL: img.URL, Description: img.Description, Private: img.Private,
		ForSale: img.ForSale, Price: img.Price, Labels: labels,
	}, nil
}

func (s *imagesService) GetImagesByFilter(ctx context.Context, userID intUserID, filter string) ([]*custom.Image, error) {
	var dbImgs dbModels.ImageSlice
	err := queries.Raw(filter).Bind(ctx, s.DB, &dbImgs)
	if err != nil {
		return nil, customErr.Internal(ctx, err.Error())
	}
	imgList := []*custom.Image{}
	for _, img := range dbImgs {
		ls, err := img.Labels().All(ctx, s.DB)
		if err != nil {
			return nil, customErr.Internal(ctx, err.Error())
		}
		labels := helpers.LabelSliceToString(ls)
		imgList = append(imgList, &custom.Image{
			ID: fmt.Sprintf("%v", img.ID), UserID: fmt.Sprintf("%v", img.UserID), Created: &img.CreatedAt,
			Title: img.Title, URL: img.URL, Description: img.Description, Private: img.Private,
			ForSale: img.ForSale, Price: img.Price, Labels: labels,
		})
	}
	return imgList, nil
}

func (s *imagesService) GetAllPublicImgs(ctx context.Context) ([]*custom.Image, error) {

	dbImgs, err := dbModels.Images(Where("images.private=False And images.archived=False")).All(ctx, s.DB)
	if err != nil {
		return nil, customErr.Internal(ctx, err.Error())
	}
	imgList := []*custom.Image{}
	for _, img := range dbImgs {
		ls, err := img.Labels().All(ctx, s.DB)
		if err != nil {
			return nil, customErr.Internal(ctx, err.Error())
		}
		labels := helpers.LabelSliceToString(ls)
		imgList = append(imgList, &custom.Image{
			ID: fmt.Sprintf("%v", img.ID), UserID: fmt.Sprintf("%v", img.UserID), Created: &img.CreatedAt,
			Title: img.Title, URL: img.URL, Description: img.Description, Private: img.Private,
			ForSale: img.ForSale, Price: img.Price, Labels: labels,
		})
	}
	return imgList, nil
}

func (s *imagesService) UpdateImage(ctx context.Context, input *model.UpdateImageInput) (*custom.Image, error) {
	userId, _, err := s.AuthService.validateCredentials(ctx)
	if err != nil {
		return nil, err
	}

	img, err := dbModels.Images(Where("id=? And images.user_Id=?", input.ID, userId)).One(ctx, s.DB)
	if err != nil {
		return nil, customErr.Forbidden(ctx, err.Error())
	}

	img.Title = input.Title
	img.ForSale = input.ForSale
	img.Private = input.Private
	img.Description = input.Description
	img.Price = input.Price

	_, err = img.Update(ctx, s.DB, boil.Infer())
	if err != nil {
		return nil, customErr.Internal(ctx, err.Error())
	}

	if input.Labels != nil {
		_, err = dbModels.Labels(Where("image_id=?", img.ID)).DeleteAll(ctx, s.DB)
		if err != nil {
			return nil, customErr.Internal(ctx, err.Error())
		}
		err = s.insertLabels(ctx, input.Labels, img.ID)
		if err != nil {
			return nil, err
		}
	}

	return &custom.Image{
		Title:       img.Title,
		Description: img.Description,
		ForSale:     img.ForSale,
		Private:     img.Private,
		UserID:      fmt.Sprintf("%v", img.UserID),
		Price:       img.Price,
		ID:          input.ID,
	}, nil
}

func (s *imagesService) insertLabels(ctx context.Context, labels []string, imgId int) error {
	if len(labels) == 0 {
		return nil
	}
	insertStr := "insert into labels (tag,image_id) values "
	for _, tag := range labels {
		insertStr = insertStr + fmt.Sprintf("('%s', %v),", strings.ToLower(tag), imgId)
	}
	insertStr = insertStr[0 : len(insertStr)-1]
	_, err := s.DB.Exec(insertStr)
	if err != nil {
		return customErr.Internal(ctx, err.Error())
	}
	return nil
}

func (s *imagesService) AutoGenerateLabels(ctx context.Context, imageId string) ([]string, error) {
	userId, _, err := s.AuthService.validateCredentials(ctx)
	if err != nil {
		return nil, err
	}
	img, err := dbModels.Images(Where("images.id=? And images.user_id=?", imageId, userId)).One(ctx, s.DB)
	if err != nil {
		return nil, err
	}
	generatedLabels, err := s.visionOperator.DetectImgProps(ctx, img.URL)
	dbLabels, err := img.Labels().All(ctx, s.DB)
	if err != nil {
		return nil, customErr.Internal(ctx, err.Error())
	}
	oldLabels := helpers.LabelSliceToString(dbLabels)
	newLabels := helpers.RemoveDuplicate(generatedLabels, oldLabels)
	err = s.insertLabels(ctx ,newLabels, img.ID)
	if err != nil {
		return nil, err
	}
	return append(newLabels, oldLabels...), nil
}
