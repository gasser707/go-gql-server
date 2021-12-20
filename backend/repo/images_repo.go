package repo

import (
	"context"
	"fmt"

	dbModels "github.com/gasser707/go-gql-server/databases/models"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/jmoiron/sqlx"
)

type ImagesRepoInterface interface {
	GetById(ctx context.Context, imgId int, userId int) (*dbModels.Image, []string ,error)
	GetAllPublic(ctx context.Context) ([]*dbModels.Image, error)
	GetByFilter(ctx context.Context, filter string)([]*dbModels.Image, error)
	GetImageIfOwner(ctx context.Context, imgId int, userId int) (*dbModels.Image, error)
	Create(ctx context.Context, dbImg *dbModels.Image) (imgId int64,err error)
	Update(ctx context.Context, id int, img *dbModels.Image) error
	Delete(ctx context.Context, imgId int, userId int) error
	InsertImageLabels(ctx context.Context, imgId int, labels []*dbModels.Label)(error)
	GetImageLabels(ctx context.Context, imgId int) ([]string, error)
	DeleteImageLabels(ctx context.Context, imgId int) (error)
	CountImageSales(ctx context.Context, imgId int) (int, error)
}

var _ ImagesRepoInterface = &imagesRepo{}

type imagesRepo struct {
	db *sqlx.DB
}

func NewImagesRepo(db *sqlx.DB) *imagesRepo {
	return &imagesRepo{
		db,
	}
}

func (r *imagesRepo) GetById(ctx context.Context, imgId int, userId int) (*dbModels.Image, []string ,error) {
	img := dbModels.Image{}
	err := r.db.Get(&img, "SELECT * FROM images WHERE id=?", imgId)
	if err != nil {
		return nil, nil, customErr.DB(ctx, err)
	}else if ok:= r.checkUserBought(ctx, imgId, int(userId)); img.UserID != int(userId) && (img.Private || img.Archived)&& !ok {
		return nil, nil, customErr.Forbidden(ctx, err.Error())
	}
	labels,err := r.GetImageLabels(ctx, imgId)
	if err != nil  {
		return nil, nil, err
	}
	return &img, labels, nil
}

func (r *imagesRepo) GetImageIfOwner(ctx context.Context, imgId int, userId int) (*dbModels.Image, error) {
	img := dbModels.Image{}
	err := r.db.Get(&img, "SELECT * FROM images WHERE id=? AND user_id=?", imgId, userId)
	if err != nil {
		return nil, customErr.DB(ctx, err)
	}
	return &img, nil
}

func (r *imagesRepo) GetAllPublic(ctx context.Context) ([]*dbModels.Image, error){
	dbImgs := []*dbModels.Image{}
	err := r.db.Select(&dbImgs, "SELECT * FROM images WHERE private=False AND archived=False")
	if err != nil {
		return nil, customErr.DB(ctx, err)
	}
	return dbImgs, nil
}

func (r *imagesRepo) GetByFilter(ctx context.Context, filter string)([]*dbModels.Image, error){
	dbImgs := []*dbModels.Image{}
	err := r.db.Select(&dbImgs, filter)
	if err != nil {
		return nil, customErr.DB(ctx, err)
	}
	return dbImgs, nil
}



func (r *imagesRepo) Create(ctx context.Context, dbImg *dbModels.Image) (imgId int64,err error) {

	result, err := r.db.NamedExec(`INSERT INTO images(title, description, private, forSale, price, user_id, created_at)
		VALUES(:title, :description, :private, :forSale, :price, :user_id, :created_at)`, dbImg)
	if err != nil {
		return -1,customErr.DB(ctx, err)
	}
	imgId, _ = result.LastInsertId()

	return imgId, nil
}


func (r *imagesRepo) Update(ctx context.Context, id int, img *dbModels.Image) error {
	
	_, err := r.db.NamedExec(fmt.Sprintf(`UPDATE images SET title= :title, forSale= :forSale, private= :private, 
	description= :description, price= :price, archived= :archived WHERE id=%d`, id), img)
	if err != nil {
		return customErr.DB(ctx, err)
	}

	return nil
}


func (r *imagesRepo) Delete(ctx context.Context, imgId int, userId int) error {

	img,err:= r.GetImageIfOwner(ctx, imgId, userId)
	if err != nil {
		return err
	}
	c, err:= r.CountImageSales(ctx, img.ID)
	if err != nil {
		return err
	}
	if c != 0 {
		img.Archived = true
		err = r.Update(ctx, img.ID, img)
		if err != nil {
			return err
		}
		return nil
	}
	_, err = r.db.Exec("DELETE FROM images WHERE id=?", imgId)
	if err != nil {
		return customErr.DB(ctx, err)
	}
	err = r.DeleteImageLabels(ctx, imgId)
	if err != nil {
		return err
	}
	return nil
}


func (r *imagesRepo) GetImageLabels(ctx context.Context, imgId int) ([]string, error) {

	labels := []string{}
	err := r.db.Select(&labels, "SELECT tag FROM labels WHERE image_id=?", imgId)
	if err != nil {
		return nil, customErr.DB(ctx, err)
	}
	return labels,nil
}

func (r *imagesRepo) InsertImageLabels(ctx context.Context, imgId int, labels []*dbModels.Label)(error) {

	_, err := r.db.NamedExec("INSERT INTO labels(image_id, tag) VALUES(:ImagedID, :Tag)", &labels)
	if err != nil {
		return customErr.DB(ctx, err)
	}
	return nil
}

func (r *imagesRepo) CountImageSales(ctx context.Context, imgId int) (int, error){

	c := 0
	err := r.db.Get(&c, "SELECT COUNT(*) FROM sales WHERE image_id=?", imgId)
	if err != nil {
		return -1, customErr.DB(ctx, err)
	}
	return c, nil
}

func (r *imagesRepo) DeleteImageLabels(ctx context.Context, imgId int) (error){

	_, err := r.db.Exec("DELETE FROM labels WHERE image_id=?", imgId)
	if err != nil {
		return customErr.DB(ctx, err)
	}
	return nil
}


func (r *imagesRepo) checkUserBought(ctx context.Context, imgId int, userId int) (bool){
	id:=-1
	err := r.db.Get(&id,"SELECT id FROM sales WHERE image_id=? AND buyer_id=?", imgId, userId)
	return err==nil
}

