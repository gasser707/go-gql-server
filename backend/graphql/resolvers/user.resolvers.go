package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	
	"github.com/gasser707/go-gql-server/graphql/custom"
	"github.com/gasser707/go-gql-server/graphql/generated"
	"github.com/gasser707/go-gql-server/graphql/model"
)


func (r *mutationResolver) RegisterUser(ctx context.Context, input model.NewUserInput) (*custom.User, error) {
	return r.UsersService.RegisterUser(ctx, input)
}

func (r *mutationResolver) UpdateUser(ctx context.Context, input model.UpdateUserInput) (*custom.User, error) {
	return r.UsersService.UpdateUser(ctx, input)
}

func (r *queryResolver) Users(ctx context.Context, input *model.UserFilterInput) ([]*custom.User, error) {
	return r.UsersService.GetUsers(ctx, input)
}

func (r *userResolver) Role(ctx context.Context, user *custom.User) (model.Role, error) {
	return model.Role(user.Role), nil
}


func (r *userResolver) Images(ctx context.Context, user *custom.User) ([]*custom.Image, error) {
	return r.ImagesService.GetImages(ctx, &model.ImageFilterInput{UserID: &user.ID})
}

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type userResolver struct{ *Resolver }
