package repo

import (
	dbModels "github.com/gasser707/go-gql-server/databases/models"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/jmoiron/sqlx"
)

type AuthRepoInterface interface {
	GetUserByEmail(email string) (*dbModels.User, error)
	UpdatePassword(id string, password string) error
	UpdateVerified(id string) error
}

var _ AuthRepoInterface = &authRepo{}
var _ AuthRepoInterface = &mysqlAuthRepo{}

type authRepo struct {
	repo AuthRepoInterface
}

type mysqlAuthRepo struct {
	db *sqlx.DB
}

func NewAuthRepo(db *sqlx.DB) *authRepo {
	mysqlRepo := &mysqlAuthRepo{
		db,
	}
	return &authRepo{
		repo: mysqlRepo,
	}
}

func (ar *authRepo) GetUserByEmail(email string) (*dbModels.User, error) {
	return ar.repo.GetUserByEmail(email)
}

func (ar *authRepo) UpdatePassword(id string, password string) error {
	return ar.repo.UpdatePassword(id, password)
}

func (ar *authRepo) UpdateVerified(id string) error {
	return ar.repo.UpdateVerified(id)
}

func (r *mysqlAuthRepo) GetUserByEmail(email string) (*dbModels.User, error) {
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

func (r *mysqlAuthRepo) UpdatePassword(id string, password string) error {
	_, err := r.db.Exec(`UPDATE users SET password=? WHERE id=?`, password, id)
	if err != nil {
		return customErr.DB(err)
	}
	return nil
}

func (r *mysqlAuthRepo) UpdateVerified(id string) error {
	_, err := r.db.Exec(`UPDATE users SET verified=true WHERE id=?`, id)
	if err != nil {
		return customErr.DB(err)
	}
	return nil
}
