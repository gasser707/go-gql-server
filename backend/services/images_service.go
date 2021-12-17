package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"github.com/gasser707/go-gql-server/cloud"
	"github.com/gasser707/go-gql-server/custom"
	dbModels "github.com/gasser707/go-gql-server/databases/models"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/gasser707/go-gql-server/graph/model"
	"github.com/gasser707/go-gql-server/helpers"
	"github.com/jmoiron/sqlx"
	"github.com/twinj/uuid"
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
	DB              *sqlx.DB
	storageOperator cloud.StorageOperatorInterface
	visionOperator  cloud.VisionOperatorInterface
}

func NewImagesService(ctx context.Context, db *sqlx.DB, storageOperator cloud.StorageOperatorInterface) *imagesService {
	vo, err := cloud.NewVisionOperator(ctx)
	if err != nil {
		panic(err)
	}
	return &imagesService{DB: db, storageOperator: storageOperator, visionOperator: vo}
}

func (s *imagesService) UploadImages(ctx context.Context, input []*model.NewImageInput) ([]*custom.Image, error) {
	userId, ok := ctx.Value(helpers.UserIdKey).(intUserID)
	if !ok {
		return nil, customErr.Internal(ctx, "userId not found in ctx")
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
	userId, ok := ctx.Value(helpers.UserIdKey).(intUserID)
	if !ok {
		return false, customErr.Internal(ctx, "userId not found in ctx")
	}
	errs, ctx := errgroup.WithContext(ctx)
	for _, delImg := range input {
		i := delImg
		errs.Go(
			func() error {
				return s.processDeleteImage(ctx, i, userId)
			})
	}
	err := errs.Wait()
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
		CreatedAt: time.Now(),
	}
	result, err := s.DB.NamedExec(`INSERT INTO images(title, description, private, forSale, price, user_id, created_at)
		VALUES(:title, :description, :private, :forSale, :price, :user_id, :created_at)`, dbImg)
	if err != nil {
		return customErr.DB(ctx, err)
	}
	imgId, _ := result.LastInsertId()

	err = s.insertLabels(ctx, inputImg.Labels, int(imgId))
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

func (s *imagesService) processDeleteImage(ctx context.Context, input *model.DeleteImageInput, userId intUserID) (err error) {

	delImgId, err := strconv.Atoi(input.ID)
	if err != nil  {
		return customErr.BadRequest(ctx, err.Error())
	} 
	img := dbModels.Image{}
	err = s.DB.Get(&img,"SELECT * FROM images WHERE id=? AND user_id=?", delImgId, userId)
	if err != nil  {
		return customErr.DB(ctx, err)
	} 

	if img.UserID != int(userId) {
		return customErr.Forbidden(ctx, err.Error())
	}
	c:= 0
	err = s.DB.Get(&c,"SELECT COUNT(*) FROM sales WHERE image_id=?", delImgId)
	if err!= nil {
		return customErr.DB(ctx, err)
	}
	if c != 0 {
		img.Archived = true
		s.DB.Exec("UPDATE images SET archived=True WHERE id=?", delImgId)
		return nil
	}
	err = s.storageOperator.DeleteImage(ctx, img.URL)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec("DELETE FROM images WHERE id=?", delImgId)
	if err != nil {
		return customErr.DB(ctx, err)
	}

	return nil
}

func (s *imagesService) GetImages(ctx context.Context, input *model.ImageFilterInput) ([]*custom.Image, error) {
	userId, ok := ctx.Value(helpers.UserIdKey).(intUserID)
	if !ok {
		return nil, customErr.Internal(ctx, "userId not found in ctx")
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
	userId, ok := ctx.Value(helpers.UserIdKey).(intUserID)
	if !ok {
		return nil, customErr.Internal(ctx, "userId not found in ctx")
	}
	inputId, err := strconv.Atoi(ID)
	if err != nil {
		return nil, customErr.BadRequest(ctx, err.Error())
	}

	img := dbModels.Image{}
	err = s.DB.Get(&img,"SELECT * FROM users WHERE id=?", inputId)
	if err != nil || (img.UserID != int(userId) && img.Private) {
		return nil, customErr.Forbidden(ctx, err.Error())
	}

	labels:= []string{}
	err = s.DB.Select(&labels, "SELECT tag FROM labels WHERE image_id=?", inputId)
	if err != nil {
		return nil, customErr.DB(ctx, err)

	}

	return &custom.Image{
		ID: ID, UserID: fmt.Sprintf("%v", img.UserID), Created: &img.CreatedAt,
		Title: img.Title, URL: img.URL, Description: img.Description, Private: img.Private,
		ForSale: img.ForSale, Price: img.Price, Labels: labels,
	}, nil
}

func (s *imagesService) GetImagesByFilter(ctx context.Context, userID intUserID, filter string) ([]*custom.Image, error) {
	dbImgs := []dbModels.Image{}
	err := s.DB.Select(&dbImgs, filter)
	if err != nil {
		return nil, customErr.DB(ctx, err)

	}
	imgList := []*custom.Image{}
	for _, img := range dbImgs {
		labels:= []string{}
		err := s.DB.Select(&labels, "SELECT tag FROM labels WHERE image_id=?", img.ID)
		if err != nil {
			return nil, customErr.DB(ctx, err)

		}
		imgList = append(imgList, &custom.Image{
			ID: fmt.Sprintf("%v", img.ID), UserID: fmt.Sprintf("%v", img.UserID), Created: &img.CreatedAt,
			Title: img.Title, URL: img.URL, Description: img.Description, Private: img.Private,
			ForSale: img.ForSale, Price: img.Price, Labels: labels,
		})
	}
	return imgList, nil
}

func (s *imagesService) GetAllPublicImgs(ctx context.Context) ([]*custom.Image, error) {
	dbImgs:= []dbModels.Image{}
	err := s.DB.Select(&dbImgs,"SELECT * FROM images WHERE private=False AND archived=False")
		if err != nil {
		return nil, customErr.DB(ctx, err)

	}
	imgList := []*custom.Image{}
	for _, img := range dbImgs {
		labels:= []string{}
		err := s.DB.Select(&labels, "SELECT tag FROM labels WHERE image_id=?", img.ID)
		if err != nil {
			return nil, customErr.DB(ctx, err)

		}
		imgList = append(imgList, &custom.Image{
			ID: fmt.Sprintf("%v", img.ID), UserID: fmt.Sprintf("%v", img.UserID), Created: &img.CreatedAt,
			Title: img.Title, URL: img.URL, Description: img.Description, Private: img.Private,
			ForSale: img.ForSale, Price: img.Price, Labels: labels,
		})
	}
	return imgList, nil
}

func (s *imagesService) UpdateImage(ctx context.Context, input *model.UpdateImageInput) (*custom.Image, error) {
	userId, ok := ctx.Value(helpers.UserIdKey).(intUserID)
	if !ok {
		return nil, customErr.Internal(ctx, "userId not found in ctx")
	}
	img := dbModels.Image{}
	err := s.DB.Get(&img,"SELECT * FROM images WHERE id=? AND user_id=?", input.ID, userId)
	if err != nil  {
		return nil,	customErr.DB(ctx, err)
	} 

	img.Title = input.Title
	img.ForSale = input.ForSale
	img.Private = input.Private
	img.Description = input.Description
	img.Price = input.Price

	_, err = s.DB.NamedExec(fmt.Sprintf(`UPDATE images SET title= :title, forSale= :forSale, private= :private, 
	description= :description, price= :price WHERE id=%d`, img.ID), img)
	if err != nil {
		return nil, customErr.DB(ctx, err)
	}

	if input.Labels != nil {
		_, err = s.DB.Exec("DELETE FROM labels WHERE image_id=?", img.ID)
		if err != nil {
			return nil, customErr.DB(ctx, err)
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
	insertedLabels:= []dbModels.Label{}
	for _, l:= range labels{
		insertedLabels= append(insertedLabels, dbModels.Label{ImageID: imgId, Tag: strings.ToLower(l)})
	}

	_, err := s.DB.NamedExec("INSERT INTO labels(image_id, tag) VALUES(:ImagedID, :Tag)", &insertedLabels)
	if err != nil {
		return customErr.DB(ctx, err)
	}
	return nil
}

func (s *imagesService) AutoGenerateLabels(ctx context.Context, imageId string) ([]string, error) {
	userId, ok := ctx.Value(helpers.UserIdKey).(intUserID)
	if !ok {
		return nil, customErr.Internal(ctx, "userId not found in ctx")
	}
	img := dbModels.Image{}
	err := s.DB.Get(&img, "SELECT * FROM images where id=? AND user_id=?", imageId, userId)
	if err != nil {
		return nil, customErr.DB(ctx, err)
	}
	generatedLabels, err := s.visionOperator.DetectImgProps(ctx, img.URL)
	if err != nil {
		return nil, customErr.Internal(ctx, err.Error())
	}
	oldLabels:= []string{}
	err = s.DB.Select(&oldLabels,"SELECT tag FROM labels WHERE image_id=?", imageId)
	if err != nil {
		return nil, customErr.DB(ctx, err)
	}
	newLabels := helpers.RemoveDuplicate(generatedLabels, oldLabels)
	err = s.insertLabels(ctx, newLabels, img.ID)
	if err != nil {
		return nil, err
	}
	return append(newLabels, oldLabels...), nil
}
