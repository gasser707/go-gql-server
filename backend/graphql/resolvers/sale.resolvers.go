package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

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
	return r.ImagesService.GetImageById(ctx, sale.ImageID)
}

func (r *saleResolver) Buyer(ctx context.Context, sale *custom.Sale) (*custom.User, error) {
	return r.UsersService.GetUserById(sale.BuyerID)
}

func (r *saleResolver) Seller(ctx context.Context, sale *custom.Sale) (*custom.User, error) {
	return r.UsersService.GetUserById(sale.SellerID)
}

// Sale returns generated.SaleResolver implementation.
func (r *Resolver) Sale() generated.SaleResolver { return &saleResolver{r} }

type saleResolver struct{ *Resolver }
