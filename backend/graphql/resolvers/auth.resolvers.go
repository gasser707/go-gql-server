package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/gasser707/go-gql-server/graphql/generated"
	"github.com/gasser707/go-gql-server/graphql/model"
)

func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (bool, error) {
	return r.AuthService.Login(ctx, input)
}

func (r *mutationResolver) Logout(ctx context.Context, input *bool) (bool, error) {
	return r.AuthService.Logout(ctx)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
