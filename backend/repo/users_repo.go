package repo

import (
	"context"
	"fmt"
	dbModels "github.com/gasser707/go-gql-server/databases/models"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/jmoiron/sqlx"
)

type UsersRepoInterface interface {
	GetById(ctx context.Context, id int) (*dbModels.User, error)
	GetByEmail(ctx context.Context, email string) (*dbModels.User, error)
	GetByUsername(ctx context.Context, username string) ([]dbModels.User, error)
	GetAll(ctx context.Context) ([]dbModels.User, error)
	CountByEmail(ctx context.Context, email string) (int, error)
	Create(ctx context.Context, insertedUser *dbModels.User) (int64 ,error)
	Update(ctx context.Context, id int, updatedUser *dbModels.User) error
}

var _ UsersRepoInterface = &usersRepo{}

type usersRepo struct {
	db *sqlx.DB
}

func NewUsersRepo(db *sqlx.DB) *usersRepo {
	return &usersRepo{
		db,
	}
}

func (r *usersRepo) GetById(ctx context.Context, id int) (*dbModels.User, error) {
	user := dbModels.User{}
	err := r.db.Get(&user, "SELECT * FROM users WHERE id=?", id)
	if err != nil {
		return nil, customErr.DB(ctx, err)
	}
	return &user, nil
}

func (r *usersRepo) GetByEmail(ctx context.Context, email string) (*dbModels.User, error) {
	user := dbModels.User{}
	err := r.db.Get(&user, "SELECT * FROM users WHERE email=?", email)
	if err != nil {
		return nil, customErr.DB(ctx, err)
	}
	return &user, nil
}

func (r *usersRepo) GetByUsername(ctx context.Context, username string) ([]dbModels.User, error) {
	users := []dbModels.User{}
	err := r.db.Get(&users, "SELECT * FROM users WHERE username=?", username)
	if err != nil {
		return nil, customErr.DB(ctx, err)
	}
	return users, nil
}

func (r *usersRepo) GetAll(ctx context.Context) ([]dbModels.User, error) {
	users := []dbModels.User{}
	err := r.db.Select(&users, "SELECT * FROM users")
	if err != nil {
		return nil, customErr.DB(ctx, err)
	}
	return users, nil
}

func (r *usersRepo) CountByEmail(ctx context.Context, email string) (int, error) {
	c := 0
	err := r.db.Get(&c, "SELECT COUNT(*) FROM users WHERE email=?", email)
	if err != nil {
		return -1, customErr.DB(ctx, err)
	}
	return c, nil
}

func (r *usersRepo) Create(ctx context.Context, insertedUser *dbModels.User) (id int64, err error) {

	result, err := r.db.NamedExec(`INSERT INTO users(email, password, username, bio, role, created_at) VALUES(
		:email, :password, :username, :bio, :role, :created_at)`, insertedUser)
	if err != nil {
		return -1, customErr.DB(ctx, err)
	}
	userId, _ := result.LastInsertId()
	return userId, nil
}


func (r *usersRepo) Update(ctx context.Context, id int, updatedUser *dbModels.User) error {
	_, err := r.db.NamedExec(fmt.Sprintf(`UPDATE users SET username=:username, bio=:bio, email=:email, 
	avatar=:avatar WHERE id = %d`, id), &updatedUser)
	if err != nil {
		return customErr.DB(ctx, err)
	}
	return nil
}


