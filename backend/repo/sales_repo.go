package repo

import (
	"context"

	dbModels "github.com/gasser707/go-gql-server/databases/models"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/jmoiron/sqlx"
)

type SalesRepoInterface interface {
	GetAll(ctx context.Context, userId int) ([]dbModels.Sale, error)
	Create(ctx context.Context, sale *dbModels.Sale) (int64 ,error)
	GetImageById(ctx context.Context, imgId int, userId int) (*dbModels.Image, error)
}

var _ SalesRepoInterface = &salesRepo{}

type salesRepo struct {
	db *sqlx.DB
}

func NewSalesRepo(db *sqlx.DB) *salesRepo {
	return &salesRepo{
		db,
	}
}


func (r *salesRepo) GetAll(ctx context.Context, userId int) ([]dbModels.Sale, error) {
	dbSales := []dbModels.Sale{}
	err := r.db.Select(&dbSales,"SELECT * FROM sales WHERE buyer_id=? OR seller_id=?", userId, userId)
	if err != nil {
		return nil, customErr.DB(ctx, err)
	}
	return dbSales, nil
}


func (r *salesRepo) Create(ctx context.Context, sale *dbModels.Sale) (id int64, err error) {
	result, err := r.db.NamedExec(`INSERT INTO sales(image_id, buyer_id, seller_id, price, created_at) VALUES (:image_id,
		:buyer_id, :seller_id, :price, :created_at)`, sale)
   if err != nil {
	   return -1, customErr.DB(ctx, err)
   }
   saleId, _ := result.LastInsertId()
	return saleId, nil
}

func (r *salesRepo) GetImageById(ctx context.Context, imgId int, userId int) (*dbModels.Image, error) {
	img := dbModels.Image{}
	err := r.db.Get(&img, "SELECT * FROM images WHERE id=?", imgId)
	if err != nil {
		return nil, customErr.DB(ctx, err)
	}else if img.UserID != int(userId) && (img.Private|| img.Archived) {
		return nil, customErr.Forbidden(ctx, err.Error())
	}
	return &img, nil
}
