package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gasser707/go-gql-server/auth"
	"github.com/gasser707/go-gql-server/custom"
	db "github.com/gasser707/go-gql-server/databases"
	dbModels "github.com/gasser707/go-gql-server/databases/models"
	"github.com/gasser707/go-gql-server/graph/model"
	"github.com/gasser707/go-gql-server/helpers"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type usersServiceInterface interface {
	UpdateUser(ctx context.Context, input model.UpdateUserInput)(*custom.User, error)
	RegisterUser(ctx context.Context, input model.NewUserInput) (*custom.User, error)
}

//UsersService implements the usersServiceInterface
var _ usersServiceInterface = &UsersService{}
type UsersService struct{
	DB *sql.DB

}

func (s *UsersService) UpdateUser(ctx context.Context, input model.UpdateUserInput)(*custom.User, error){

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


func (s *UsersService) RegisterUser(ctx context.Context, input model.NewUserInput) (*custom.User, error){

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
	if err != nil {
		return nil, err
	}

	returnedUser := &custom.User{
		Username: input.Username,
		Email:    input.Email,
		Bio:      input.Bio,
		Avatar:   input.Avatar,
		Joined:   &insertedUser.CreatedAt,
	}
	return returnedUser, nil


} 