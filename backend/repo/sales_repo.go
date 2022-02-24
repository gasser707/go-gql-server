package repo

import (
	dbModels "github.com/gasser707/go-gql-server/databases/models"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/jmoiron/sqlx"
)

type SalesRepoInterface interface {
	GetAll(userId int) ([]dbModels.Sale, error)
	Create(sale *dbModels.Sale) (int64, error)
	GetImageById(imgId int, userId int) (*dbModels.Image, error)
}

var _ SalesRepoInterface = &mysqlSalesRepo{}
var _ SalesRepoInterface = &salesRepo{}

type mysqlSalesRepo struct {
	db *sqlx.DB
}

type salesRepo struct {
	repo SalesRepoInterface
}

func NewSalesRepo(db *sqlx.DB) *salesRepo {
	mysqlRepo := &mysqlSalesRepo{
		db,
	}
	return &salesRepo{
		repo: mysqlRepo,
	}
}

func (sr *salesRepo) GetAll(userId int) ([]dbModels.Sale, error) {
	return sr.repo.GetAll(userId)
}

func (sr *salesRepo) Create(sale *dbModels.Sale) (id int64, err error) {
	return sr.repo.Create(sale)
}

func (sr *salesRepo) GetImageById(imgId int, userId int) (*dbModels.Image, error) {
	return sr.repo.GetImageById(imgId, userId)
}

func (r *mysqlSalesRepo) GetAll(userId int) ([]dbModels.Sale, error) {
	dbSales := []dbModels.Sale{}
	err := r.db.Select(&dbSales, "SELECT * FROM sales WHERE buyer_id=? OR seller_id=?", userId, userId)
	if err != nil {
		return nil, customErr.DB(err)
	}
	return dbSales, nil
}

func (r *mysqlSalesRepo) Create(sale *dbModels.Sale) (id int64, err error) {
	result, err := r.db.NamedExec(`INSERT INTO sales(image_id, buyer_id, seller_id, price, created_at) VALUES (:image_id,
		:buyer_id, :seller_id, :price, :created_at)`, sale)
	if err != nil {
		return -1, customErr.DB(err)
	}
	saleId, _ := result.LastInsertId()
	return saleId, nil
}

func (r *mysqlSalesRepo) GetImageById(imgId int, userId int) (*dbModels.Image, error) {
	img := dbModels.Image{}
	err := r.db.Get(&img, "SELECT * FROM images WHERE id=?", imgId)
	if err != nil {
		return nil, customErr.DB(err)
	} else if img.UserID != int(userId) && (img.Private || img.Archived) {
		return nil, customErr.Forbidden(err.Error())
	}
	return &img, nil
}
