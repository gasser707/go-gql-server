package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gasser707/go-gql-server/cloud"
	"github.com/gasser707/go-gql-server/custom"
	dbModels "github.com/gasser707/go-gql-server/databases/models"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/gasser707/go-gql-server/graph/model"
	"github.com/gasser707/go-gql-server/helpers"
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
	DB              *sqlx.DB
	storageOperator cloud.StorageOperatorInterface
}

func NewUsersService(db *sqlx.DB, storageOperator cloud.StorageOperatorInterface) *usersService {
	return &usersService{DB: db, storageOperator: storageOperator}
}

func (s *usersService) UpdateUser(ctx context.Context, input model.UpdateUserInput) (*custom.User, error) {

	userId, ok := ctx.Value(helpers.UserIdKey).(intUserID)
	if !ok {
		return nil, customErr.Internal(ctx, "userId not found in ctx")
	}
	user := &dbModels.User{}
	err := s.DB.Get(user, "SELECT * FROM users WHERE id = ?", userId)
	if err != nil {
		return nil, customErr.Internal(ctx, err.Error())
	}
	// if err != nil && err != sql.ErrNoRows {
	// 	return nil, customErr.Internal(ctx, err.Error())
	// } else if err == sql.ErrNoRows {
	// 	return nil, customErr.NotFound(ctx, err.Error())
	// }

	user.Username = input.Username
	user.Bio = input.Bio

	if input.Email != user.Email {
		c := 0
		s.DB.Get(&c, "SELECT COUNT(*) FROM users WHERE email=?", input.Email)
		if c != 0 {
			return nil, customErr.BadRequest(ctx, "A user with this email already exists")
		}
	}
	user.Email = input.Email

	var newAvatarUrl string
	if input.Avatar != nil {
		newAvatarUrl, err = s.storageOperator.UploadImage(ctx, input.Avatar, "avatar", fmt.Sprintf("%v", userId))
		if err != nil {
			return nil, err
		}
	}
	if newAvatarUrl != "" {
		user.Avatar = newAvatarUrl
	}
	_, err = s.DB.NamedExec(fmt.Sprintf(`UPDATE users SET username=:username, bio=:bio, email=:email, 
	avatar=:avatar WHERE id = %d`, userId), user)
	if err != nil {
		return nil, customErr.Internal(ctx, err.Error())
	}

	returnUser := &custom.User{Avatar: user.Avatar, Email: user.Email,
		Username: user.Username, Bio: user.Bio, ID: fmt.Sprintf("%v", userId)}
	return returnUser, nil
}

func (s *usersService) RegisterUser(ctx context.Context, input model.NewUserInput) (*custom.User, error) {

	c := 0
	s.DB.Get(&c, "SELECT COUNT(*) FROM users WHERE email=?", input.Email)
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

	result, err := s.DB.NamedExec(`INSERT INTO users(email, password, username, bio, role, created_at) VALUES(
		:email, :password, :username, :bio, :role, :created_at)`, insertedUser)
	if err != nil {
		return nil, customErr.Internal(ctx, err.Error())
	}
	userId, _ := result.LastInsertId()

	avatarUrl, err := s.storageOperator.UploadImage(ctx, &input.Avatar, "avatar", fmt.Sprintf("%v", userId))
	if err != nil {
		return nil, err
	}
	insertedUser.Avatar = avatarUrl
	result, err = s.DB.NamedExec(`UPDATE users SET avatar=:avatar`, insertedUser)
	if err != nil {
		return nil, customErr.Internal(ctx, err.Error())
	}
	returnedUser := &custom.User{
		Username: input.Username,
		Email:    input.Email,
		Bio:      input.Bio,
		Avatar:   avatarUrl,
		Joined:   &insertedUser.CreatedAt,
		ID:       fmt.Sprintf("%v", insertedUser.ID),
	}

	return returnedUser, nil
}

func (s *usersService) GetUsers(ctx context.Context, input *model.UserFilterInput) ([]*custom.User, error) {
	_, ok := ctx.Value(helpers.UserIdKey).(intUserID)
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
		return nil, customErr.Internal(ctx, err.Error())
	}
	user := dbModels.User{}
	err = s.DB.Get(&user, "SELECT * FROM users WHERE id=?", inputId)
	if err != nil {
		return nil, customErr.Internal(ctx, err.Error())
	}

	return &custom.User{
		ID: fmt.Sprintf("%v", user.ID), Username: user.Username,
		Email: user.Email, Avatar: user.Avatar, Joined: &user.CreatedAt,
	}, nil
}

func (s *usersService) GetUserByEmail(ctx context.Context, email string) ([]*custom.User, error) {

	user := dbModels.User{}
	err := s.DB.Get(&user, "SELECT * FROM users WHERE email=?", email)
	if err != nil {
		return nil, customErr.Internal(ctx, err.Error())
	}
	return []*custom.User{
		{ID: fmt.Sprintf("%v", user.ID),
			Username: user.Username, Email: user.Email, Avatar: user.Avatar, Joined: &user.CreatedAt,
		},
	}, nil
}

func (s *usersService) GetUsersByUserName(ctx context.Context, username string) ([]*custom.User, error) {

	userList := []*custom.User{}
	users := []dbModels.User{}
	err := s.DB.Get(&users, "SELECT * FROM users WHERE id=?", username)
	if err != nil {
		return nil, customErr.Internal(ctx, err.Error())
	}
	for _, user := range users {
		userList = append(userList, &custom.User{ID: fmt.Sprintf("%v", user.ID),
			Username: user.Username, Email: user.Email, Avatar: user.Avatar, Joined: &user.CreatedAt})
	}
	return userList, nil
}

func (s *usersService) GetAllUsers(ctx context.Context) ([]*custom.User, error) {
	userList := []*custom.User{}
	users := []dbModels.User{}
	err := s.DB.Select(&users, "SELECT * FROM users")
	if err != nil {
		return nil, customErr.Internal(ctx, err.Error())
	}
	for _, user := range users {
		userList = append(userList, &custom.User{ID: fmt.Sprintf("%v", user.ID),
			Username: user.Username, Email: user.Email, Avatar: user.Avatar, Joined: &user.CreatedAt})
	}
	return userList, nil
}
