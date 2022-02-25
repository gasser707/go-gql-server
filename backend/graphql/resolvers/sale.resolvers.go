package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"strconv"

	"github.com/gasser707/go-gql-server/graphql/custom"
	"github.com/gasser707/go-gql-server/graphql/generated"
)

func (r *mutationResolver) BuyImage(ctx context.Context, id string) (*custom.Sale, error) {
	return r.SaleService.BuyImage(ctx, id)
}

func (r *queryResolver) Sales(ctx context.Context) ([]*custom.Sale, error) {
	return r.SaleService.GetSales(ctx)
}

func (r *saleResolver) Image(ctx context.Context, sale *custom.Sale) (*custom.Image, error) {
	imgId, _ := strconv.Atoi(sale.ImageID)
	return r.DataLoaders.Retrieve(ctx).ImageByID.Load(imgId)
}

func (r *saleResolver) Buyer(ctx context.Context, sale *custom.Sale) (*custom.User, error) {
	buyerId, _ := strconv.Atoi(sale.BuyerID)
	return r.DataLoaders.Retrieve(ctx).UserByID.Load(buyerId)
}

func (r *saleResolver) Seller(ctx context.Context, sale *custom.Sale) (*custom.User, error) {
	sellerId, _ := strconv.Atoi(sale.SellerID)
	return r.DataLoaders.Retrieve(ctx).UserByID.Load(sellerId)
}

// Sale returns generated.SaleResolver implementation.
func (r *Resolver) Sale() generated.SaleResolver { return &saleResolver{r} }

type saleResolver struct{ *Resolver }
