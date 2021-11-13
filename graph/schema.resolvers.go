package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"github.com/gasser707/go-gql-server/auth"
	"github.com/gasser707/go-gql-server/custom"
	"github.com/gasser707/go-gql-server/graph/generated"
	"github.com/gasser707/go-gql-server/graph/model"
	"github.com/gasser707/go-gql-server/services"

)

func (r *imageResolver) User(ctx context.Context, obj *custom.Image) (*custom.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) RegisterUser(ctx context.Context, input model.NewUserInput) (*custom.User, error) {
	return services.UsersService.RegisterUser(ctx, input)
}

func (r *mutationResolver) UpdateUser(ctx context.Context, input model.UpdateUserInput) (*custom.User, error) {

}

func (r *mutationResolver) UploadImages(ctx context.Context, input []*model.NewImageInput) ([]*custom.Image, error) {
	return services.ImagesService.UploadImages(ctx, input)
}

func (r *mutationResolver) DeleteImages(ctx context.Context, input []*model.DeleteImageInput) (bool, error) {
	return services.ImagesService.UploadImages(ctx, input)

}

func (r *mutationResolver) UpdateImage(ctx context.Context, input model.UpdateImageInput) (*custom.Image, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) BuyImage(ctx context.Context, input *model.BuyImageInput) (*custom.Sale, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (bool, error) {
	ok, err := auth.AuthService.Login(ctx, input)

	if ok {
		return ok, nil
	} else {
		return false, fmt.Errorf(err.Error())
	}
}

func (r *mutationResolver) Logout(ctx context.Context, input *bool) (bool, error) {
	ok, err := auth.AuthService.Logout(ctx)

	if ok {
		return ok, nil
	} else {
		return false, fmt.Errorf(err.Error())
	}
}

func (r *queryResolver) Images(ctx context.Context, input *model.ImageFilterInput) ([]*custom.Image, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Users(ctx context.Context, input *model.UserFilterInput) ([]*custom.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *saleResolver) Image(ctx context.Context, obj *custom.Sale) (*custom.Image, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *saleResolver) Buyer(ctx context.Context, obj *custom.Sale) (*custom.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *saleResolver) Seller(ctx context.Context, obj *custom.Sale) (*custom.User, error) {
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

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *userResolver) Role(ctx context.Context, obj *custom.User) (model.Role, error) {
	panic(fmt.Errorf("not implemented"))
}
