package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	dbModels "github.com/gasser707/go-gql-server/databases/models"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/gasser707/go-gql-server/graphql/custom"
	"github.com/gasser707/go-gql-server/graphql/model"
	"github.com/gasser707/go-gql-server/helpers"
	"github.com/gasser707/go-gql-server/repo"
	email_svc "github.com/gasser707/go-gql-server/services/email"
	authUtils "github.com/gasser707/go-gql-server/utils/auth"
	"github.com/gasser707/go-gql-server/utils/cloud"
	"github.com/jmoiron/sqlx"
)

type UsersServiceInterface interface {
	UpdateUser(ctx context.Context, input model.UpdateUserInput) (*custom.User, error)
	RegisterUser(ctx context.Context, input model.NewUserInput) (*custom.User, error)
	GetUsers(ctx context.Context, input *model.UserFilterInput) ([]*custom.User, error)
	GetUserById(ctx context.Context, ID string) (*custom.User, error)
}

//UsersService implements the usersServiceInterface
var _ UsersServiceInterface = &usersService{}

type usersService struct {
	repo            repo.UsersRepoInterface
	storageOperator cloud.StorageOperatorInterface
	emailAdaptor    email_svc.EmailAdaptorInterface
	ValTokenMaker   authUtils.TokenOperatorInterface
}

func NewUsersService(db *sqlx.DB, storageOperator cloud.StorageOperatorInterface,
	emailAdaptor email_svc.EmailAdaptorInterface) *usersService {

	return &usersService{repo: repo.NewUsersRepo(db), storageOperator: storageOperator, emailAdaptor: emailAdaptor,
		ValTokenMaker: authUtils.NewTokenOperator(nil)}
}

func (s *usersService) RegisterUser(ctx context.Context, input model.NewUserInput) (*custom.User, error) {

	c, err := s.repo.CountByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if c != 0 {
		return nil, customErr.BadRequest(ctx, "A user with this email already exists")
	}
	pwd, err := helpers.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	insertedUser := &dbModels.User{
		Email:     input.Email,
		Password:  pwd,
		Username:  input.Username,
		Bio:       input.Bio,
		Role:      model.RoleUser.String(),
		CreatedAt: time.Now(),
	}
	userId, err := s.repo.Create(ctx, insertedUser)
	if err != nil {
		return nil, err
	}
	avatarUrl := ""
	if input.Avatar != nil {
		avatarUrl, err = s.storageOperator.UploadImage(input.Avatar.File, "avatar", fmt.Sprintf("%v", userId))
		if err != nil {
			return nil, err
		}
	}
	insertedUser.Avatar = avatarUrl

	err = s.repo.Update(ctx, int(userId), insertedUser)
	if err != nil {
		return nil, err
	}
	returnedUser := &custom.User{
		Username: input.Username,
		Email:    input.Email,
		Bio:      input.Bio,
		Avatar:   avatarUrl,
		Joined:   &insertedUser.CreatedAt,
		ID:       fmt.Sprintf("%v", userId),
	}
	token, err := s.ValTokenMaker.CreateStatelessToken(fmt.Sprintf("%v", userId), authUtils.ValidateUserToken)
	if err != nil {
		return nil, err
	}
	go s.emailAdaptor.SendResetPassEmail(ctx, "auth@shotify.com", []string{input.Email},
		input.Username, fmt.Sprintf("http://%s/validate?token=%s", domain, token))

	return returnedUser, nil
}

func (s *usersService) GetUsers(ctx context.Context, input *model.UserFilterInput) ([]*custom.User, error) {
	_, ok := ctx.Value(helpers.UserIdKey).(IntUserID)
	if !ok {
		return nil, customErr.Internal(ctx, "userId not found in ctx")
	}
	if input == nil {
		return s.GetAllUsers(ctx)
	}

	if input.ID != nil {
		user, err := s.GetUserById(ctx, *input.ID)
		if err != nil {
			return nil, err
		}
		return []*custom.User{user}, nil
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
		return nil, customErr.BadRequest(ctx, err.Error())
	}
	user, err := s.repo.GetById(ctx, inputId)
	if err != nil {
		return nil, err
	}
	return &custom.User{
		ID:       fmt.Sprintf("%v", user.ID),
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Joined:   &user.CreatedAt,
		Bio:      user.Bio,
	}, nil
}

func (s *usersService) GetUserByEmail(ctx context.Context, email string) ([]*custom.User, error) {

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return []*custom.User{
		{ID: fmt.Sprintf("%v", user.ID),
			Username: user.Username,
			Email:    user.Email,
			Avatar:   user.Avatar,
			Joined:   &user.CreatedAt,
			Bio:      user.Bio,
		},
	}, nil
}

func (s *usersService) GetUsersByUserName(ctx context.Context, username string) ([]*custom.User, error) {

	userList := []*custom.User{}
	users, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		userList = append(userList, &custom.User{
			ID:       fmt.Sprintf("%v", user.ID),
			Username: user.Username,
			Email:    user.Email,
			Avatar:   user.Avatar,
			Joined:   &user.CreatedAt,
			Bio:      user.Bio})
	}
	return userList, nil
}

func (s *usersService) GetAllUsers(ctx context.Context) ([]*custom.User, error) {
	userList := []*custom.User{}
	users, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		userList = append(userList, &custom.User{
			ID:       fmt.Sprintf("%v", user.ID),
			Username: user.Username,
			Email:    user.Email,
			Avatar:   user.Avatar,
			Joined:   &user.CreatedAt,
			Bio:      user.Bio})
	}
	return userList, nil
}

func (s *usersService) UpdateUser(ctx context.Context, input model.UpdateUserInput) (*custom.User, error) {

	userId, ok := ctx.Value(helpers.UserIdKey).(IntUserID)
	if !ok {
		return nil, customErr.Internal(ctx, "userId not found in ctx")
	}
	user, err := s.repo.GetById(ctx, int(userId))
	if err != nil {
		return nil, err
	}

	user.Username = input.Username
	user.Bio = input.Bio

	if input.Email != user.Email {
		c, err := s.repo.CountByEmail(ctx, input.Email)
		if err != nil {
			return nil, err
		}
		if c != 0 {
			return nil, customErr.BadRequest(ctx, "A user with this email already exists")
		}
	}
	user.Email = input.Email

	var newAvatarUrl string
	if input.Avatar != nil {
		newAvatarUrl, err = s.storageOperator.UploadImage( input.Avatar.File, "avatar", fmt.Sprintf("%v", userId))
		if err != nil {
			return nil, err
		}
	}
	if newAvatarUrl != "" {
		user.Avatar = newAvatarUrl
	}

	err = s.repo.Update(ctx, int(userId), user)
	if err != nil {
		return nil, err
	}

	returnUser := &custom.User{
		Avatar:   user.Avatar,
		Email:    user.Email,
		Username: user.Username,
		Bio:      user.Bio,
		ID:       fmt.Sprintf("%v", userId)}
	return returnUser, nil
}
