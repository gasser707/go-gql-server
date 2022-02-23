package repo

import (
	"context"

	dbModels "github.com/gasser707/go-gql-server/databases/models"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/jmoiron/sqlx"
)

type AuthRepoInterface interface {
	GetUserByEmail(ctx context.Context, email string) (*dbModels.User, error)
	UpdatePassword(ctx context.Context, id string, password string) error
	UpdateVerified(ctx context.Context, id string) error
}

var _ AuthRepoInterface = &authRepo{}

type authRepo struct {
	db *sqlx.DB
}

func NewAuthRepo(db *sqlx.DB) *authRepo {
	return &authRepo{
		db,
	}
}

func (r *authRepo) GetUserByEmail(ctx context.Context, email string) (*dbModels.User, error) {
	user := dbModels.User{}
	err := r.db.Get(&user, "SELECT * FROM users WHERE email=?", email)
	if err != nil {
		return nil, customErr.BadRequest(err.Error())
	}
	if !user.Verfied {
		return nil, customErr.UnProcessable("your account in unverified! go to http://localhost:8025 to verify it")
	}
	return &user, nil
}

func (r *authRepo) UpdatePassword(ctx context.Context, id string, password string) error {
	_, err := r.db.Exec(`UPDATE users SET password=? WHERE id=?`, password, id)
	if err != nil {
		return customErr.DB(err)
	}
	return nil
}

func (r *authRepo) UpdateVerified(ctx context.Context, id string) error {
	_, err := r.db.Exec(`UPDATE users SET verified=true WHERE id=?`, id)
	if err != nil {
		return customErr.DB(err)
	}
	return nil
}
