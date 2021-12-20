package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	dbModels "github.com/gasser707/go-gql-server/databases/models"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/gasser707/go-gql-server/graphql/custom"
	"github.com/gasser707/go-gql-server/helpers"
	"github.com/gasser707/go-gql-server/repo"
	"github.com/jmoiron/sqlx"
)

type SalesServiceInterface interface {
	BuyImage(ctx context.Context, id string) (*custom.Sale, error)
	GetSales(ctx context.Context) ([]*custom.Sale, error)
}

//SalesService implements the usersServiceInterface
var _ SalesServiceInterface = &salesService{}

type salesService struct {
	repo         repo.SalesRepoInterface
}

func NewSalesService(db *sqlx.DB) *salesService {
	return &salesService{repo: repo.NewSalesRepo(db)}
}

func (s *salesService) BuyImage(ctx context.Context, id string) (*custom.Sale, error) {
	userId, ok := ctx.Value(helpers.UserIdKey).(intUserID)
	if !ok {
		return nil, customErr.Internal(ctx, "userId not found in ctx")
	}
	imgId, err := strconv.Atoi(id)
	if err != nil {
		return nil, customErr.BadRequest(ctx, err.Error())
	}
	img, err := s.repo.GetImageById(ctx, imgId, int(userId))
	if err != nil || !img.ForSale || img.UserID == int(userId) {
		return nil, customErr.Forbidden(ctx, err.Error())
	}
	sale := dbModels.Sale{
		Price:     img.Price,
		ImageID:   imgId,
		BuyerID:   int(userId),
		SellerID:  img.UserID,
		CreatedAt: time.Now(),
	}
	saleId,err := s.repo.Create(ctx, &sale)
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

func (s *salesService) GetSales(ctx context.Context) ([]*custom.Sale, error) {
	userId, ok := ctx.Value(helpers.UserIdKey).(intUserID)
	if !ok {
		return nil, customErr.Internal(ctx, "userId not found in ctx")
	}
	dbSales,err := s.repo.GetAll(ctx, int(userId))
	if err != nil {
		return nil, customErr.DB(ctx, err)
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
