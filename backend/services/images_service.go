package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	dbModels "github.com/gasser707/go-gql-server/databases/models"
	customErr "github.com/gasser707/go-gql-server/errors"
	email_svc"github.com/gasser707/go-gql-server/services/email"
	"github.com/gasser707/go-gql-server/graphql/custom"
	"github.com/gasser707/go-gql-server/graphql/model"
	"github.com/gasser707/go-gql-server/helpers"
	"github.com/gasser707/go-gql-server/repo"
	"github.com/gasser707/go-gql-server/utils/cloud"
	"github.com/jmoiron/sqlx"
	"github.com/twinj/uuid"
	"golang.org/x/sync/errgroup"
)

type ImagesServiceInterface interface {
	UploadImages(ctx context.Context, input []*model.NewImageInput) ([]*custom.Image, error)
	DeleteImages(ctx context.Context, input []string) (bool, error)
	GetImages(ctx context.Context, input *model.ImageFilterInput) ([]*custom.Image, error)
	GetImageById(ctx context.Context, ID string) (*custom.Image, error)
	UpdateImage(ctx context.Context, input *model.UpdateImageInput) (*custom.Image, error)
	AutoGenerateLabels(ctx context.Context, imageId string) ([]string, error)
}

//UsersService implements the usersServiceInterface
var _ ImagesServiceInterface = &imagesService{}

type imagesService struct {
	repo            repo.ImagesRepoInterface
	storageOperator cloud.StorageOperatorInterface
	visionOperator  cloud.VisionOperatorInterface
	emailAdaptor    email_svc.EmailAdaptorInterface
}

func NewImagesService(ctx context.Context, db *sqlx.DB, storageOperator cloud.StorageOperatorInterface,
	emailAdaptor email_svc.EmailAdaptorInterface) *imagesService {
	vo, err := cloud.NewVisionOperator(ctx)
	if err != nil {
		panic(err)
	}
	return &imagesService{repo: repo.NewImagesRepo(db), storageOperator: storageOperator,
		visionOperator: vo, emailAdaptor: emailAdaptor}
}

func (s *imagesService) UploadImages(ctx context.Context, input []*model.NewImageInput) ([]*custom.Image, error) {
	userId, ok := ctx.Value(helpers.UserIdKey).(IntUserID)
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

func (s *imagesService) DeleteImages(ctx context.Context, input []string) (bool, error) {
	userId, ok := ctx.Value(helpers.UserIdKey).(IntUserID)
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
	userId IntUserID) (err error) {

	url, err := s.storageOperator.UploadImage(ctx, &inputImg.File, uuid.NewV4().String(), fmt.Sprintf("%v", userId))
	if err != nil {
		return err
	}
	dbImg := dbModels.Image{
		Title:           inputImg.Title,
		Description:     inputImg.Description,
		Private:         inputImg.Private,
		ForSale:         inputImg.ForSale,
		Price:           inputImg.Price,
		DiscountPercent: inputImg.DiscountPercent,
		UserID:          int(userId),
		URL:             url,
		CreatedAt:       time.Now(),
	}
	imgId, err := s.repo.Create(ctx, &dbImg)
	if err != nil {
		return err
	}
	err = s.insertLabels(ctx, inputImg.Labels, int(imgId))
	if err != nil {
		return err
	}

	image := &custom.Image{
		ID:              fmt.Sprintf("%v", imgId),
		Title:           dbImg.Title,
		Description:     dbImg.Description,
		URL:             dbImg.URL,
		Private:         dbImg.Private,
		ForSale:         dbImg.ForSale,
		Price:           dbImg.Price,
		DiscountPercent: dbImg.DiscountPercent,
		UserID:          fmt.Sprintf("%v", userId),
		Created:         &dbImg.CreatedAt,
		Labels:          inputImg.Labels,
	}

	ch <- image
	return nil
}

func (s *imagesService) processDeleteImage(ctx context.Context, ID string, userId IntUserID) (err error) {

	delImgId, err := strconv.Atoi(ID)
	if err != nil {
		return customErr.BadRequest(ctx, err.Error())
	}
	img, err := s.repo.GetImageIfOwner(ctx, delImgId, int(userId))
	if err != nil {
		return err
	}
	err = s.repo.Delete(ctx, delImgId, int(userId))
	if err != nil {
		return err
	}
	url := img.URL
	err = s.storageOperator.DeleteImage(ctx, url)
	if err != nil {
		return err
	}
	return nil
}

func (s *imagesService) GetImages(ctx context.Context, input *model.ImageFilterInput) ([]*custom.Image, error) {
	userId, ok := ctx.Value(helpers.UserIdKey).(IntUserID)
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
	userId, ok := ctx.Value(helpers.UserIdKey).(IntUserID)
	if !ok {
		return nil, customErr.Internal(ctx, "userId not found in ctx")
	}
	inputId, err := strconv.Atoi(ID)
	if err != nil {
		return nil, customErr.BadRequest(ctx, err.Error())
	}
	img, labels, err := s.repo.GetById(ctx, inputId, int(userId))
	if err != nil {
		return nil, err

	}

	return &custom.Image{
		ID:              ID,
		UserID:          fmt.Sprintf("%v", img.UserID),
		Created:         &img.CreatedAt,
		Title:           img.Title,
		URL:             img.URL,
		Description:     img.Description,
		Private:         img.Private,
		ForSale:         img.ForSale,
		Price:           img.Price,
		DiscountPercent: img.DiscountPercent,
		Labels:          labels,
		Archived:        img.Archived,
	}, nil
}

func (s *imagesService) GetImagesByFilter(ctx context.Context, userID IntUserID, filter string) ([]*custom.Image, error) {
	dbImgs, err := s.repo.GetByFilter(ctx, filter)
	if err != nil {
		return nil, err
	}
	imgList := []*custom.Image{}
	for _, img := range dbImgs {
		labels, err := s.repo.GetImageLabels(ctx, img.ID)
		if err != nil {
			return nil, err

		}
		imgList = append(imgList, &custom.Image{
			ID:              fmt.Sprintf("%v", img.ID),
			UserID:          fmt.Sprintf("%v", img.UserID),
			Created:         &img.CreatedAt,
			Title:           img.Title,
			URL:             img.URL,
			Description:     img.Description,
			Private:         img.Private,
			ForSale:         img.ForSale,
			Price:           img.Price,
			DiscountPercent: img.DiscountPercent,
			Labels:          labels,
			Archived:        img.Archived,
		})
	}
	return imgList, nil
}

func (s *imagesService) GetAllPublicImgs(ctx context.Context) ([]*custom.Image, error) {
	dbImgs, err := s.repo.GetAllPublic(ctx)
	if err != nil {
		return nil, err
	}
	imgList := []*custom.Image{}
	for _, img := range dbImgs {
		labels, err := s.repo.GetImageLabels(ctx, img.ID)
		if err != nil {
			return nil, err

		}
		imgList = append(imgList, &custom.Image{
			ID:              fmt.Sprintf("%v", img.ID),
			UserID:          fmt.Sprintf("%v", img.UserID),
			Created:         &img.CreatedAt,
			Title:           img.Title,
			URL:             img.URL,
			Description:     img.Description,
			Private:         img.Private,
			ForSale:         img.ForSale,
			Price:           img.Price,
			DiscountPercent: img.DiscountPercent,
			Labels:          labels,
			Archived:        img.Archived,
		})
	}
	return imgList, nil
}

func (s *imagesService) UpdateImage(ctx context.Context, input *model.UpdateImageInput) (*custom.Image, error) {
	userId, ok := ctx.Value(helpers.UserIdKey).(IntUserID)
	if !ok {
		return nil, customErr.Internal(ctx, "userId not found in ctx")
	}
	imgId, err := strconv.Atoi(input.ID)
	if err != nil {
		return nil, customErr.BadRequest(ctx, err.Error())
	}
	img, err := s.repo.GetImageIfOwner(ctx, imgId, int(userId))
	if err != nil {
		return nil, err
	}

	img.Title = input.Title
	img.ForSale = input.ForSale
	img.Private = input.Private
	img.Description = input.Description
	img.Price = input.Price
	img.DiscountPercent = input.DiscountPercent
	img.Archived = input.Archived

	err = s.repo.Update(ctx, img.ID, img)
	if err != nil {
		return nil, err
	}

	if input.Labels != nil {
		err = s.repo.DeleteImageLabels(ctx, imgId)
		if err != nil {
			return nil, err
		}
		err = s.insertLabels(ctx, input.Labels, img.ID)
		if err != nil {
			return nil, err
		}
	}

	return &custom.Image{
		Title:           img.Title,
		Description:     img.Description,
		ForSale:         img.ForSale,
		Private:         img.Private,
		UserID:          fmt.Sprintf("%v", img.UserID),
		Price:           img.Price,
		DiscountPercent: img.DiscountPercent,
		ID:              input.ID,
		Archived:        input.Archived,
	}, nil
}

func (s *imagesService) insertLabels(ctx context.Context, labels []string, imgId int) error {
	if len(labels) == 0 {
		return nil
	}
	insertedLabels := []*dbModels.Label{}
	for _, l := range labels {
		insertedLabels = append(insertedLabels, &dbModels.Label{ImageID: imgId, Tag: strings.ToLower(l)})
	}

	err := s.repo.InsertImageLabels(ctx, imgId, insertedLabels)
	if err != nil {
		return err
	}
	return nil
}

func (s *imagesService) AutoGenerateLabels(ctx context.Context, imageId string) ([]string, error) {
	userId, ok := ctx.Value(helpers.UserIdKey).(IntUserID)
	if !ok {
		return nil, customErr.Internal(ctx, "userId not found in ctx")
	}
	imgId, err := strconv.Atoi(imageId)
	if err != nil {
		return nil, customErr.BadRequest(ctx, err.Error())
	}
	img, err := s.repo.GetImageIfOwner(ctx, imgId, int(userId))
	if err != nil {
		return nil, customErr.DB(ctx, err)
	}
	generatedLabels, err := s.visionOperator.DetectImgProps(ctx, img.URL)
	if err != nil {
		return nil, err
	}
	oldLabels, err := s.repo.GetImageLabels(ctx, imgId)
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
