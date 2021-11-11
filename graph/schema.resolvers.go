package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"github.com/gasser707/go-gql-server/auth"
	"github.com/gasser707/go-gql-server/custom"
	db "github.com/gasser707/go-gql-server/databases"
	dbModels "github.com/gasser707/go-gql-server/databases/models"
	"github.com/gasser707/go-gql-server/graph/generated"
	"github.com/gasser707/go-gql-server/graph/model"
	"github.com/gasser707/go-gql-server/helpers"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func (r *imageResolver) User(ctx context.Context, obj *custom.Image) (*custom.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) RegisterUser(ctx context.Context, input model.NewUserInput) (*custom.User, error) {
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
	err = insertedUser.Insert(ctx, db.MysqlDB, boil.Infer())
	if(err!=nil){
		return nil, err
	}

	returnedUser := &custom.User{
		Username: input.Username,
		Email:    input.Email,
		Bio:      input.Bio,
		Avatar:   input.Avatar,
		Joined: &insertedUser.CreatedAt,
	}
	return returnedUser, nil
}

func (r *mutationResolver) UpdateUser(ctx context.Context, input model.UpdateUserInput) (*custom.User, error) {
	userId, err := auth.AuthService.GetCredentials(ctx)
	if err != nil {
		return nil, err
	}

	user, err := dbModels.FindUser(ctx, db.MysqlDB, int(userId))

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

	if err != nil {
		return nil, err
	}

	returnUser := &custom.User{Avatar: user.Avatar, Email: user.Email, Username: user.Username, Bio: user.Bio}
	return returnUser, nil
}

func (r *mutationResolver) UploadImages(ctx context.Context, input []*model.NewImageInput) ([]*custom.Image, error) {
	userId, err := auth.AuthService.GetCredentials(ctx)
	if err != nil {
		return nil, err
	}
	dbImages := []dbModels.Image{}
	for _, inputImg := range input {
		image := dbModels.Image{
		  Title: inputImg.Title, Description: inputImg.Description,
		  URL: inputImg.URL, Private: inputImg.Private,
		  ForSale: inputImg.ForSale, Price: inputImg.Price, UserID: int(userId),
		}
		err:= image.Insert(ctx, db.MysqlDB, boil.Infer())
		if(err!= nil ){
			return nil, err
		}

		for _, inputLabel:= range inputImg.Labels{
			label:= dbModels.Label{
				Tag: inputLabel,
				ImageID: image.ID,
			}
			err:= label.Insert(ctx, db.MysqlDB, boil.Infer())
			if(err!= nil){
				return nil, err
			}
		}
		dbImages = append(dbImages, image)
	}

	images := []*custom.Image{}
	for _, dbImg := range dbImages {
		imgId:= fmt.Sprintf("%v", dbImg.ID)
		image := &custom.Image{
		  ID: imgId, Title: dbImg.Title, Description: dbImg.Description,
		  URL: dbImg.URL, Private: dbImg.Private,
		  ForSale: dbImg.ForSale, Price: dbImg.Price, UserID: string(userId),
		  Created: &dbImg.CreatedAt,
		}
		images = append(images, image)
	}

	return images, nil
}

func (r *mutationResolver) DeleteImages(ctx context.Context, input []*model.DeleteImageInput) (bool, error) {
	userId, err := auth.AuthService.GetCredentials(ctx)
	if err != nil {
		return false, err
	}
	
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
