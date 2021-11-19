package services

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/gasser707/go-gql-server/cloud"
	"github.com/gasser707/go-gql-server/custom"
	dbModels "github.com/gasser707/go-gql-server/databases/models"
	"github.com/gasser707/go-gql-server/graph/model"
	"github.com/twinj/uuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/sync/errgroup"
	// . "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type ImagesServiceInterface interface {
	UploadImages(ctx context.Context, input []*model.NewImageInput) ([]*custom.Image, error)
	DeleteImages(ctx context.Context, input []*model.DeleteImageInput) (bool, error)
}

//UsersService implements the usersServiceInterface
var _ ImagesServiceInterface = &imagesService{}

type imagesService struct {
	DB            *sql.DB
	AuthService   AuthServiceInterface
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
		fmt.Println(2)

	if err != nil {
		return err
	}
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
	image := &custom.Image{
		ID: fmt.Sprintf("%v", dbImg.ID), Title: dbImg.Title, Description: dbImg.Description,
		URL: dbImg.URL, Private: dbImg.Private,
		ForSale: dbImg.ForSale, Price: dbImg.Price, UserID: fmt.Sprintf("%v", userId),
		Created: &dbImg.CreatedAt,
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
