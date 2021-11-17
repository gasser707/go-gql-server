package services

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"github.com/gasser707/go-gql-server/custom"
	dbModels "github.com/gasser707/go-gql-server/databases/models"
	"github.com/gasser707/go-gql-server/graph/model"
	"github.com/volatiletech/sqlboiler/v4/boil"
	// . "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type ImagesServiceInterface interface {
	UploadImages(ctx context.Context, input []*model.NewImageInput) ([]*custom.Image, error)
	DeleteImages(ctx context.Context, input []*model.DeleteImageInput) (bool, error)
}

//UsersService implements the usersServiceInterface
var _ ImagesServiceInterface = &imagesService{}
type imagesService struct{
	DB *sql.DB
	AuthService AuthServiceInterface
}

func NewImagesService( db *sql.DB, authSrv AuthServiceInterface ) *imagesService {
	return &imagesService{DB: db, AuthService: authSrv}
}

func (s *imagesService) UploadImages(ctx context.Context, input []*model.NewImageInput) ([]*custom.Image, error) {
	userId, err := s.AuthService.GetCredentials(ctx)
	if err != nil {
		return nil, err
	}
	dbImages := []dbModels.Image{}
	for _, inputImg := range input {
		image := dbModels.Image{
			Title: inputImg.Title, Description: inputImg.Description,
		    Private: inputImg.Private,
			ForSale: inputImg.ForSale, Price: inputImg.Price, UserID: int(userId),
		}
		fmt.Println(inputImg.File)
		err := image.Insert(ctx, s.DB, boil.Infer())
		if err != nil {
			return nil, err
		}

		for _, inputLabel := range inputImg.Labels {
			label := dbModels.Label{
				Tag:     inputLabel,
				ImageID: image.ID,
			}
			err := label.Insert(ctx, s.DB, boil.Infer())
			if err != nil {
				return nil, err
			}
		}
		dbImages = append(dbImages, image)
	}

	images := []*custom.Image{}
	for _, dbImg := range dbImages {
		imgId := fmt.Sprintf("%v", dbImg.ID)
		image := &custom.Image{
			ID: imgId, Title: dbImg.Title, Description: dbImg.Description,
			URL: dbImg.URL, Private: dbImg.Private,
			ForSale: dbImg.ForSale, Price: dbImg.Price, UserID: string(userId),
			Created: &dbImg.CreatedAt,
		}
		images = append(images, image)
	}

	return images, nil
}


func (s *imagesService)  DeleteImages(ctx context.Context, input []*model.DeleteImageInput) (bool, error) {
	userId, err := s.AuthService.GetCredentials(ctx)
	if err != nil {
		return false, err
	}

	for _, delImg := range input {
		delImgId, _ := strconv.Atoi(delImg.ID)
		img, err := dbModels.FindImage(ctx, s.DB, delImgId)
		if err != nil {
			return false, err
		}
		if img.UserID != int(userId) {
			return false, fmt.Errorf("an image from the list doesn't belong to you")
		}
		_, err = img.Delete(ctx, s.DB)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}
