package repo

import (
	"fmt"
	dbModels "github.com/gasser707/go-gql-server/databases/models"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/jmoiron/sqlx"
)

type UsersRepoInterface interface {
	GetById(id int) (*dbModels.User, error)
	GetByEmail(email string) (*dbModels.User, error)
	GetByUsername(username string) ([]dbModels.User, error)
	GetAll() ([]dbModels.User, error)
	CountByEmail(email string) (int, error)
	Create(insertedUser *dbModels.User) (int64, error)
	Update(id int, updatedUser *dbModels.User) error
}

var _ UsersRepoInterface = &usersRepo{}
var _ UsersRepoInterface = &mysqlUsersRepo{}

type mysqlUsersRepo struct {
	db *sqlx.DB
}

type usersRepo struct {
	repo UsersRepoInterface
}

func NewUsersRepo(db *sqlx.DB) *usersRepo {
	mysqlRepo := &mysqlUsersRepo{
		db,
	}
	return &usersRepo{
		repo: mysqlRepo,
	}
}

func (r *usersRepo) GetById(id int) (*dbModels.User, error) {
	return r.repo.GetById(id)

}

func (r *usersRepo) GetByEmail(email string) (*dbModels.User, error) {
	return r.repo.GetByEmail(email)
}

func (r *usersRepo) GetByUsername(username string) ([]dbModels.User, error) {
	return r.repo.GetByUsername(username)
}

func (r *usersRepo) GetAll() ([]dbModels.User, error) {
	return r.repo.GetAll()
}

func (r *usersRepo) CountByEmail(email string) (int, error) {
	return r.repo.CountByEmail(email)

}

func (r *usersRepo) Create(insertedUser *dbModels.User) (id int64, err error) {
	return r.repo.Create(insertedUser)
}

func (r *usersRepo) Update(id int, updatedUser *dbModels.User) error {
	return r.repo.Update(id, updatedUser)
}

func (r *mysqlUsersRepo) GetById(id int) (*dbModels.User, error) {
	user := dbModels.User{}
	err := r.db.Get(&user, "SELECT * FROM users WHERE id=?", id)
	if err != nil {
		return nil, customErr.DB(err)
	}
	return &user, nil
}

func (r *mysqlUsersRepo) GetByEmail(email string) (*dbModels.User, error) {
	user := dbModels.User{}
	err := r.db.Get(&user, "SELECT * FROM users WHERE email=?", email)
	if err != nil {
		return nil, customErr.DB(err)
	}
	return &user, nil
}

func (r *mysqlUsersRepo) GetByUsername(username string) ([]dbModels.User, error) {
	users := []dbModels.User{}
	err := r.db.Get(&users, "SELECT * FROM users WHERE username=?", username)
	if err != nil {
		return nil, customErr.DB(err)
	}
	return users, nil
}

func (r *mysqlUsersRepo) GetAll() ([]dbModels.User, error) {
	users := []dbModels.User{}
	err := r.db.Select(&users, "SELECT * FROM users")
	if err != nil {
		return nil, customErr.DB(err)
	}
	return users, nil
}

func (r *mysqlUsersRepo) CountByEmail(email string) (int, error) {
	c := 0
	err := r.db.Get(&c, "SELECT COUNT(*) FROM users WHERE email=?", email)
	if err != nil {
		return -1, customErr.DB(err)
	}
	return c, nil
}

func (r *mysqlUsersRepo) Create(insertedUser *dbModels.User) (id int64, err error) {

	result, err := r.db.NamedExec(`INSERT INTO users(email, password, username, bio, role, created_at) VALUES(
		:email, :password, :username, :bio, :role, :created_at)`, insertedUser)
	if err != nil {
		return -1, customErr.DB(err)
	}
	userId, _ := result.LastInsertId()
	return userId, nil
}

func (r *mysqlUsersRepo) Update(id int, updatedUser *dbModels.User) error {
	_, err := r.db.NamedExec(fmt.Sprintf(`UPDATE users SET username=:username, bio=:bio, email=:email, 
	avatar=:avatar WHERE id = %d`, id), &updatedUser)
	if err != nil {
		return customErr.DB(err)
	}
	return nil
}
