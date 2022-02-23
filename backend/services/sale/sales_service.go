package services

import (
	"context"
	"fmt"
	"strconv"

	dbModels "github.com/gasser707/go-gql-server/databases/models"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/gasser707/go-gql-server/graphql/custom"
	"github.com/gasser707/go-gql-server/helpers"
	"github.com/gasser707/go-gql-server/repo"
	"github.com/gasser707/go-gql-server/services"
	"github.com/gasser707/go-gql-server/utils"
	"github.com/jmoiron/sqlx"
)

type SalesServiceInterface interface {
	BuyImage(ctx context.Context, id string) (*custom.Sale, error)
	GetSales(ctx context.Context) ([]*custom.Sale, error)
}

//SalesService implements the usersServiceInterface
var _ SalesServiceInterface = &SalesService{}

type SalesService struct {
	Repo repo.SalesRepoInterface
}

func NewSalesService(db *sqlx.DB) *SalesService {
	return &SalesService{Repo: repo.NewSalesRepo(db)}
}

func (s *SalesService) BuyImage(ctx context.Context, id string) (*custom.Sale, error) {
	userId, ok := ctx.Value(helpers.UserIdKey).(services.IntUserID)
	if !ok {
		return nil, customErr.Internal("userId not found in ctx")
	}
	imgId, err := strconv.Atoi(id)
	if err != nil {
		return nil, customErr.BadRequest(err.Error())
	}
	img, err := s.Repo.GetImageById(ctx, imgId, int(userId))
	if err != nil || !img.ForSale || img.UserID == int(userId) {
		return nil, customErr.Forbidden("you can't buy an image you own")
	}
	sale := dbModels.Sale{
		Price:     img.Price,
		ImageID:   imgId,
		BuyerID:   int(userId),
		SellerID:  img.UserID,
		CreatedAt: utils.Now(),
	}
	saleId, err := s.Repo.Create(ctx, &sale)
	if err != nil {
		return nil, err
	}

	return &custom.Sale{
		Price:    sale.Price,
		ImageID:  id,
		BuyerID:  fmt.Sprintf("%v", userId),
		SellerID: fmt.Sprintf("%v", sale.SellerID),
		Time:     &sale.CreatedAt,
		ID:       fmt.Sprintf("%d", saleId),
	}, nil
}

func (s *SalesService) GetSales(ctx context.Context) ([]*custom.Sale, error) {
	userId, ok := ctx.Value(helpers.UserIdKey).(services.IntUserID)
	if !ok {
		return nil, customErr.Internal("userId not found in ctx")
	}
	dbSales, err := s.Repo.GetAll(ctx, int(userId))
	if err != nil {
		return nil, customErr.DB(err)
	}

	sales := []*custom.Sale{}
	for _, s := range dbSales {
		sale := &custom.Sale{
			ID:       fmt.Sprintf("%v", s.ID),
			Time:     &s.CreatedAt,
			ImageID:  fmt.Sprintf("%v", s.ImageID),
			BuyerID:  fmt.Sprintf("%v", s.BuyerID),
			SellerID: fmt.Sprintf("%v", s.SellerID),
			Price:    s.Price,
		}
		sales = append(sales, sale)
	}

	return sales, nil

}
