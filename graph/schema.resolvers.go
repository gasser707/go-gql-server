package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gasser707/go-gql-server/auth"
	db "github.com/gasser707/go-gql-server/databases"
	dbModels "github.com/gasser707/go-gql-server/databases/models"
	"github.com/gasser707/go-gql-server/graph/generated"
	"github.com/gasser707/go-gql-server/graph/model"
	"github.com/gasser707/go-gql-server/helpers"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func (r *imageResolver) User(ctx context.Context, obj *model.Image) (*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) RegisterUser(ctx context.Context, input model.NewUserInput) (*model.User, error) {
	c, _ := dbModels.Users(Where("email = ?", input.Email)).Count(ctx, db.MysqlDB)

	if c != 0 {
		return nil, fmt.Errorf("user already exists")
	}

	pwd, err := helpers.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}
	insertedUser := &dbModels.User{
		Email:    input.Email,
		Password: pwd,
		Username: input.Username,
		Bio:      input.Bio,
		Avatar:   input.Avatar,
		Role:     model.RoleUser.String(),
	}
	insertedUser.Insert(ctx, db.MysqlDB, boil.Infer())

	returnedUser := &model.User{
		Username: input.Username,
		Email:    input.Email,
		Bio:      input.Bio,
		Avatar:   input.Avatar,
	}
	return returnedUser, nil
}

func (r *mutationResolver) UpdateUser(ctx context.Context, input model.UpdateUserInput) (*model.User, error) {

	err := auth.AuthService.IsSelf(ctx, input)
	if(err!= nil){
		return nil, err
	}

	id, err := strconv.Atoi(input.ID)
	if err != nil {
		return nil, err
	}

	user, err := dbModels.FindUser(ctx, db.MysqlDB, id)

	if err != nil {
		return nil, err
	}

	if input.Avatar != nil {
		user.Avatar = *input.Avatar
	}
	if input.Bio != nil {
		user.Bio = *input.Bio
	}
	if input.Username != nil {
		user.Username = *input.Username
	}
	if input.Email != nil {
		user.Username = *input.Email
	}

	_, err = user.Update(ctx, db.MysqlDB, boil.Infer())

	if(err != nil){
		return nil, err
	}

	returnUser := &model.User{Avatar: user.Avatar, Email: user.Email, Username: user.Username, Bio: user.Bio}

	return returnUser, nil
}

func (r *mutationResolver) UploadImages(ctx context.Context, input []*model.NewImageInput) ([]*model.Image, error) {
	userId := auth.AuthService.GetCredentials(ctx)
	id, err := strconv.Atoi(string(userId))
	if err != nil {
		return nil, err
	}

	images := []model.Image{}

	for _,value:= range input{
		image:= model.Image{Title: value.Title, Description: value.Description, URL: value.URL, Private: value.Private,
			ForSale: value.ForSale, Price: value.Price,
			User: &model.User{
				Username: model.Image.User.Username,
			},
		}
		images = append(images, image)
	}

}

func (r *mutationResolver) DeleteImages(ctx context.Context, input []*model.DeleteImageInput) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpdateImage(ctx context.Context, input model.UpdateImageInput) (*model.Image, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) BuyImage(ctx context.Context, input *model.BuyImageInput) (*model.Sale, error) {
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

func (r *queryResolver) Images(ctx context.Context, input *model.ImageFilterInput) ([]*model.Image, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Users(ctx context.Context, input *model.UserFilterInput) ([]*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *saleResolver) Image(ctx context.Context, obj *model.Sale) (*model.Image, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *saleResolver) Buyer(ctx context.Context, obj *model.Sale) (*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *saleResolver) Seller(ctx context.Context, obj *model.Sale) (*model.User, error) {
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

// Sale returns generated.SaleResolver implementation.
func (r *Resolver) Sale() generated.SaleResolver { return &saleResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type imageResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type saleResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
