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

func (r *mutationResolver) LogoutAll(ctx context.Context, input *bool) (bool, error) {
	return r.AuthService.LogoutAll(ctx)
}

func (r *mutationResolver) Refresh(ctx context.Context, input *bool) (bool, error) {
	return r.AuthService.RefreshCredentials(ctx)
}

func (r *mutationResolver) ValidateUser(ctx context.Context, validationToken string) (bool, error) {
	return r.AuthService.ValidateUser(ctx, validationToken)
}

func (r *mutationResolver) RequestPasswordReset(ctx context.Context, email string) (bool, error) {
	return r.AuthService.RequestPasswordReset(ctx, email)
}

func (r *mutationResolver) ProcessPasswordReset(ctx context.Context, resetToken string, newPassword string) (bool, error) {
	return r.AuthService.ProcessPasswordReset(ctx, resetToken, newPassword)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
