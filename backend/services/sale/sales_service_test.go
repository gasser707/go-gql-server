package services

import (
	"context"
	"testing"
	"time"

	dbModels "github.com/gasser707/go-gql-server/databases/models"
	"github.com/gasser707/go-gql-server/graphql/custom"
	mocks "github.com/gasser707/go-gql-server/mocks/repo"
	"github.com/gasser707/go-gql-server/services"
	"github.com/gasser707/go-gql-server/utils"
	"github.com/stretchr/testify/suite"
)

type SalesServiceTestSuite struct {
	suite.Suite
}

func (suite *SalesServiceTestSuite) TestBuyImage() {
	utils.Now = func() time.Time {
		return time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
	}

	mockSalesRepo := mocks.SalesRepoInterface{}

	//Setup expectations
	img := &dbModels.Image{
		ID:              1,
		CreatedAt:       utils.Now(),
		URL:             "foo.com",
		Description:     "bar",
		UserID:          2,
		Title:           "foo",
		Price:           20,
		ForSale:         true,
		Private:         false,
		Archived:        false,
		DiscountPercent: 5,
	}

	sale := &dbModels.Sale{
		ID:        0,
		ImageID:   1,
		BuyerID:   1,
		SellerID:  2,
		CreatedAt: img.CreatedAt,
		Price:     20,
	}
	ctx := context.Background()
	ctx = setValInCtx(ctx, "userId", services.IntUserID(1))

	mockSalesRepo.On("GetImageById", ctx, 1, 1).Return(img, nil)
	mockSalesRepo.On("Create", ctx, sale).Return(int64(1), nil)

	salesService := SalesService{&mockSalesRepo}

	//buy image with id 1
	result, err := salesService.BuyImage(ctx, "1")

	mockSalesRepo.AssertExpectations(suite.T())

	suite.EqualValues(&custom.Sale{
		Price:    img.Price,
		ImageID:  "1",
		BuyerID:  "1",
		SellerID: "2",
		Time:     &img.CreatedAt,
		ID:       "1",
	}, result)
	suite.Nil(err)
}
func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(SalesServiceTestSuite))
}

func setValInCtx(ctx context.Context, key string, val interface{}) context.Context {
	newCtx := context.WithValue(ctx, key, val)
	return newCtx
}
