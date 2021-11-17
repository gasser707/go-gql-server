package services

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gasser707/go-gql-server/custom"
	dbModels "github.com/gasser707/go-gql-server/databases/models"
	"github.com/gasser707/go-gql-server/graph/model"
	"github.com/gasser707/go-gql-server/helpers"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type UsersServiceInterface interface {
	UpdateUser(ctx context.Context, input model.UpdateUserInput)(*custom.User, error)
	RegisterUser(ctx context.Context, input model.NewUserInput) (*custom.User, error)
}


//UsersService implements the usersServiceInterface
var _ UsersServiceInterface = &usersService{}
type usersService struct{
	DB *sql.DB
	AuthService AuthServiceInterface

}
func NewUsersService( db *sql.DB, authSrv AuthServiceInterface ) *usersService {
	return &usersService{DB: db, AuthService: authSrv}
}

func (s *usersService) UpdateUser(ctx context.Context, input model.UpdateUserInput)(*custom.User, error){

	userId, _, err := s.AuthService.validateCredentials(ctx)
	if err != nil {
		return nil, err
	}

	user, err := dbModels.FindUser(ctx, s.DB, int(userId))
	if err != nil {
		return nil, err
	}

	user.Username = input.Username
	user.Bio = input.Bio
	user.Email = input.Email

	var newAvatarUrl string
	if (input.Avatar != nil){
		//upload avatar to google image
	}

	if(newAvatarUrl != ""){
		user.Avatar = newAvatarUrl
	}

	_, err = user.Update(ctx, s.DB, boil.Infer())
	if(err !=nil){
		return nil, err
	}

	returnUser := &custom.User{Avatar: user.Avatar, Email: user.Email, Username: user.Username, Bio: user.Bio}
	return returnUser, nil
}


func (s *usersService) RegisterUser(ctx context.Context, input model.NewUserInput) (*custom.User, error){

	c, _ := dbModels.Users(Where("email = ?", input.Email)).Count(ctx, s.DB)

	if c != 0 {
		return nil, fmt.Errorf("user already exists")
	}

	pwd, err := helpers.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	var avatarUrl string

	//upload to google cloud

	insertedUser := &dbModels.User{
		Email:    input.Email,
		Password: pwd,
		Username: input.Username,
		Bio:      input.Bio,
		Avatar:   avatarUrl,
		Role:     model.RoleUser.String(),
	}

	err = insertedUser.Insert(ctx, s.DB, boil.Infer())
	if err != nil {
		return nil, err
	}

	returnedUser := &custom.User{
		Username: input.Username,
		Email:    input.Email,
		Bio:      input.Bio,
		Avatar:   avatarUrl,
		Joined:   &insertedUser.CreatedAt,
	}

	return returnedUser, nil
} 