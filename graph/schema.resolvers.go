package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/gasser707/go-gql-server/graph/generated"
	"github.com/gasser707/go-gql-server/graph/model"
	handlers "github.com/gasser707/go-gql-server/handler"
)

func (r *imageResolver) User(ctx context.Context, obj *model.Image) (*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateUser(ctx context.Context, input model.NewUser) (*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateImage(ctx context.Context, input model.NewImage) (*model.Image, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (*bool, error) {
	ok, err := handlers.AuthService.Login(ctx)

	if ok {
		return &ok, nil
	} else {
		return new(bool), fmt.Errorf(err.Error())
	}
}

func (r *mutationResolver) Logout(ctx context.Context, input *bool) (*bool, error) {
	ok, err := handlers.AuthService.Logout(ctx)

	if ok {
		return &ok, nil
	} else {
		return new(bool), fmt.Errorf(err.Error())
	}}

func (r *queryResolver) Images(ctx context.Context) ([]*model.Image, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *userResolver) Images(ctx context.Context, obj *model.User) ([]*model.Image, error) {
	panic(fmt.Errorf("not implemented"))
}

// Image returns generated.ImageResolver implementation.
func (r *Resolver) Image() generated.ImageResolver { return &imageResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type imageResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type userResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *queryResolver) User(ctx context.Context) (*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}
