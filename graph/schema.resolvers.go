package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/gasser707/go-gql-server/custom"
	"github.com/gasser707/go-gql-server/graph/generated"
	"github.com/gasser707/go-gql-server/graph/model"
)

func (r *imageResolver) User(ctx context.Context, img *custom.Image) (*custom.User, error) {
	return r.UsersService.GetUserById(ctx, img.UserID)
}

func (r *mutationResolver) RegisterUser(ctx context.Context, input model.NewUserInput) (*custom.User, error) {
	return r.UsersService.RegisterUser(ctx, input)
}

func (r *mutationResolver) UpdateUser(ctx context.Context, input model.UpdateUserInput) (*custom.User, error) {
	return r.UsersService.UpdateUser(ctx, input)
}

func (r *mutationResolver) UploadImages(ctx context.Context, input []*model.NewImageInput) ([]*custom.Image, error) {
	return r.ImagesService.UploadImages(ctx, input)
}

func (r *mutationResolver) DeleteImages(ctx context.Context, input []*model.DeleteImageInput) (bool, error) {
	return r.ImagesService.DeleteImages(ctx, input)
}

func (r *mutationResolver) UpdateImage(ctx context.Context, input model.UpdateImageInput) (*custom.Image, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) BuyImage(ctx context.Context, input *model.BuyImageInput) (*custom.Sale, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (bool, error) {
	return r.AuthService.Login(ctx, input)
}

func (r *mutationResolver) Logout(ctx context.Context, input *bool) (bool, error) {
	return r.AuthService.Logout(ctx)
}

func (r *queryResolver) Images(ctx context.Context, input *model.ImageFilterInput) ([]*custom.Image, error) {
	return r.ImagesService.GetImages(ctx, input)
}

func (r *queryResolver) Users(ctx context.Context, input *model.UserFilterInput) ([]*custom.User, error) {
	return r.UsersService.GetUsers(ctx, input)
}

func (r *saleResolver) Image(ctx context.Context, obj *custom.Sale) (*custom.Image, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *saleResolver) Buyer(ctx context.Context, sale *custom.Sale) (*custom.User, error) {
	return r.UsersService.GetUserById(ctx, sale.BuyerID)
}

func (r *saleResolver) Seller(ctx context.Context, sale *custom.Sale) (*custom.User, error) {
	return r.UsersService.GetUserById(ctx, sale.SellerID)
}

func (r *userResolver) Role(ctx context.Context, obj *custom.User) (model.Role, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *userResolver) Images(ctx context.Context, obj *custom.User) ([]*custom.Image, error) {
	panic(fmt.Errorf("not implemented"))
}

// Image returns generated.ImageResolver implementation.
func (r *Resolver) Image() generated.ImageResolver { return &imageResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Sale returns generated.SaleResolver implementation.
func (r *Resolver) Sale() generated.SaleResolver { return &saleResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type imageResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type saleResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
