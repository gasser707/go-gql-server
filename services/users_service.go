package services

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/gasser707/go-gql-server/custom"
	dbModels "github.com/gasser707/go-gql-server/databases/models"
	"github.com/gasser707/go-gql-server/graph/model"
	"github.com/gasser707/go-gql-server/helpers"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type UsersServiceInterface interface {
	UpdateUser(ctx context.Context, input model.UpdateUserInput) (*custom.User, error)
	RegisterUser(ctx context.Context, input model.NewUserInput) (*custom.User, error)
	GetUsers(ctx context.Context, input model.UserFilterInput) ([]*custom.User, error)
	GetUserById(ctx context.Context, ID string) (*custom.User, error)
}

//UsersService implements the usersServiceInterface
var _ UsersServiceInterface = &usersService{}

type usersService struct {
	DB            *sql.DB
	AuthService   AuthServiceInterface
	cloudOperator helpers.CloudOperatorInterface
}

func NewUsersService(db *sql.DB, authSrv AuthServiceInterface, cloudOperator helpers.CloudOperatorInterface) *usersService {
	return &usersService{DB: db, AuthService: authSrv, cloudOperator: cloudOperator}
}

func (s *usersService) UpdateUser(ctx context.Context, input model.UpdateUserInput) (*custom.User, error) {

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
	if input.Avatar != nil {
		newAvatarUrl, err = s.cloudOperator.UploadImage(ctx, input.Avatar, "avatar", fmt.Sprintf("%v", userId))
		if err != nil {
			return nil, err
		}
	}
	if newAvatarUrl != "" {
		user.Avatar = newAvatarUrl
	}
	_, err = user.Update(ctx, s.DB, boil.Infer())
	if err != nil {
		return nil, err
	}

	returnUser := &custom.User{Avatar: user.Avatar, Email: user.Email,
		Username: user.Username, Bio: user.Bio, ID: fmt.Sprintf("%v", userId)}
	return returnUser, nil
}

func (s *usersService) RegisterUser(ctx context.Context, input model.NewUserInput) (*custom.User, error) {

	c, _ := dbModels.Users(Where("email=?", input.Email)).Count(ctx, s.DB)

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
		Role:     model.RoleUser.String(),
	}

	err = insertedUser.Insert(ctx, s.DB, boil.Infer())
	if err != nil {
		return nil, err
	}

	avatarUrl, err := s.cloudOperator.UploadImage(ctx, &input.Avatar, "avatar", fmt.Sprintf("%v", insertedUser.ID))
	if err != nil {
		return nil, err
	}
	insertedUser.Avatar = avatarUrl
	insertedUser.Update(ctx, s.DB, boil.Infer())

	returnedUser := &custom.User{
		Username: input.Username,
		Email:    input.Email,
		Bio:      input.Bio,
		Avatar:   avatarUrl,
		Joined:   &insertedUser.CreatedAt,
	}

	return returnedUser, nil
}

func (s *usersService) GetUsers(ctx context.Context, input model.UserFilterInput) ([]*custom.User, error) {
	_, _, err := s.AuthService.validateCredentials(ctx)
	if err != nil {
		return nil, err
	}

	if input.ID != nil {
		user, err := s.GetUserById(ctx, *input.ID)
		if(err!=nil){
			return nil, err
		}
		return []*custom.User{user},nil
	}

	if input.Email != nil {
		return s.GetUserByEmail(ctx, *input.Email)
	}
	if input.Username != nil {
		return s.GetUsersByUserName(ctx, *input.Username)
	}
	return s.GetAllUsers(ctx)

}


func (s *usersService) GetUserById(ctx context.Context, ID string) (*custom.User, error) {
	
	inputId, err := strconv.Atoi(ID)
	if err != nil {
		return nil, err
	}

	user, err := dbModels.FindUser(ctx, s.DB, inputId)
	if err != nil {
		return nil, err
	}
	return &custom.User{
		ID: fmt.Sprintf("%v", user.ID), Username: user.Username, 
		Email: user.Email, Avatar: user.Avatar, Joined: &user.CreatedAt,
	}, nil
}


func (s *usersService) GetUserByEmail(ctx context.Context, email string) ([]*custom.User, error) {

	user, err := dbModels.Users(Where("email = ?", email)).One(ctx, s.DB)
		if err != nil {
			return nil, err
		}
		return []*custom.User{
			{ID: fmt.Sprintf("%v", user.ID),
				Username: user.Username, Email: user.Email, Avatar: user.Avatar, Joined: &user.CreatedAt,
			},
		}, nil
}


func (s *usersService) GetUsersByUserName(ctx context.Context, username string) ([]*custom.User, error) {

	userList := []*custom.User{}
	users, err := dbModels.Users(Where("username = ?", username)).All(ctx, s.DB)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("no users with this username found")
	}
	for _, user := range users {
		userList = append(userList, &custom.User{ID: fmt.Sprintf("%v", user.ID),
			Username: user.Username, Email: user.Email, Avatar: user.Avatar, Joined: &user.CreatedAt})
	}
	return userList, nil
}

func (s *usersService) GetAllUsers(ctx context.Context) ([]*custom.User, error) {
	userList := []*custom.User{}
	users, err := dbModels.Users().All(ctx, s.DB)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		userList = append(userList, &custom.User{ID: fmt.Sprintf("%v", user.ID),
			Username: user.Username, Email: user.Email, Avatar: user.Avatar, Joined: &user.CreatedAt})
	}
	return userList, nil
}