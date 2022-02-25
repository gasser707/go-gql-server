package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"strconv"

	"github.com/gasser707/go-gql-server/graphql/custom"
	"github.com/gasser707/go-gql-server/graphql/generated"
	"github.com/gasser707/go-gql-server/graphql/model"
)

func (r *imageResolver) User(ctx context.Context, img *custom.Image) (*custom.User, error) {
	userId, _ := strconv.Atoi(img.UserID)
	return r.DataLoaders.Retrieve(ctx).UserByID.Load(userId)
}

func (r *mutationResolver) UploadImages(ctx context.Context, input []*model.NewImageInput) ([]*custom.Image, error) {
	return r.ImagesService.UploadImages(ctx, input)
}

func (r *mutationResolver) DeleteImages(ctx context.Context, input []string) (bool, error) {
	return r.ImagesService.DeleteImages(ctx, input)
}

func (r *mutationResolver) UpdateImage(ctx context.Context, input model.UpdateImageInput) (*custom.Image, error) {
	return r.ImagesService.UpdateImage(ctx, &input)
}

func (r *mutationResolver) AutoGenerateLabels(ctx context.Context, id string) ([]string, error) {
	return r.ImagesService.AutoGenerateLabels(ctx, id)
}

func (r *queryResolver) Images(ctx context.Context, input *model.ImageFilterInput) ([]*custom.Image, error) {
	return r.ImagesService.GetImages(ctx, input)
}

// Image returns generated.ImageResolver implementation.
func (r *Resolver) Image() generated.ImageResolver { return &imageResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type imageResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
