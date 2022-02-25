package repo

import (
	"context"
	"fmt"

	dbModels "github.com/gasser707/go-gql-server/databases/models"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/jmoiron/sqlx"
)

type ImagesRepoInterface interface {
	GetById(imgId int, userId int) (*dbModels.Image, []string, error)
	GetAllPublic(ctx context.Context) ([]*dbModels.Image, error)
	GetByFilter(filter string) ([]*dbModels.Image, error)
	GetImageIfOwner(imgId int, userId int) (*dbModels.Image, error)
	Create(dbImg *dbModels.Image) (imgId int64, err error)
	Update(id int, img *dbModels.Image) error
	Delete(imgId int, userId int) error
	InsertImageLabels(imgId int, labels []*dbModels.Label) error
	GetImageLabels(imgId int) ([]string, error)
	DeleteImageLabels(imgId int) error
	CountImageSales(imgId int) (int, error)
	checkUserBought(imgId int, userId int) bool
}

var _ ImagesRepoInterface = &imagesRepo{}

var _ ImagesRepoInterface = &mysqlImagesRepo{}

type imagesRepo struct {
	repo ImagesRepoInterface
}

type mysqlImagesRepo struct {
	db *sqlx.DB
}

func NewImagesRepo(db *sqlx.DB) *imagesRepo {
	mysqlRepo := &mysqlImagesRepo{
		db,
	}
	return &imagesRepo{
		repo: mysqlRepo,
	}
}

func (r *imagesRepo) GetById(imgId int, userId int) (*dbModels.Image, []string, error) {
	return r.repo.GetById(imgId, userId)
}

func (r *imagesRepo) GetImageIfOwner(imgId int, userId int) (*dbModels.Image, error) {
	return r.repo.GetImageIfOwner(imgId, userId)
}

func (r *imagesRepo) GetAllPublic(ctx context.Context) ([]*dbModels.Image, error) {
	return r.repo.GetAllPublic(ctx)

}

func (r *imagesRepo) GetByFilter(filter string) ([]*dbModels.Image, error) {
	return r.repo.GetByFilter(filter)

}

func (r *imagesRepo) Create(dbImg *dbModels.Image) (imgId int64, err error) {
	return r.repo.Create(dbImg)

}

func (r *imagesRepo) Update(id int, img *dbModels.Image) error {
	return r.repo.Update(id, img)

}

func (r *imagesRepo) Delete(imgId int, userId int) error {
	return r.repo.Delete(imgId, userId)

}

func (r *imagesRepo) GetImageLabels(imgId int) ([]string, error) {
	return r.repo.GetImageLabels(imgId)

}

func (r *imagesRepo) InsertImageLabels(imgId int, labels []*dbModels.Label) error {
	return r.repo.InsertImageLabels(imgId, labels)

}

func (r *imagesRepo) CountImageSales(imgId int) (int, error) {
	return r.repo.CountImageSales(imgId)

}

func (r *imagesRepo) DeleteImageLabels(imgId int) error {
	return r.repo.DeleteImageLabels(imgId)

}

func (r *imagesRepo) checkUserBought(imgId int, userId int) bool {
	return r.repo.checkUserBought(imgId, userId)
}

func (r *mysqlImagesRepo) GetById(imgId int, userId int) (*dbModels.Image, []string, error) {
	img := dbModels.Image{}
	err := r.db.Get(&img, "SELECT * FROM images WHERE id=?", imgId)
	if err != nil {
		return nil, nil, customErr.DB(err)
	} else if ok := r.checkUserBought(imgId, int(userId)); img.UserID != int(userId) && (img.Private || img.Archived) && !ok {
		return nil, nil, customErr.Forbidden(err.Error())
	}
	labels, err := r.GetImageLabels(imgId)
	if err != nil {
		return nil, nil, err
	}
	return &img, labels, nil
}

func (r *mysqlImagesRepo) GetImageIfOwner(imgId int, userId int) (*dbModels.Image, error) {
	img := dbModels.Image{}
	err := r.db.Get(&img, "SELECT * FROM images WHERE id=? AND user_id=?", imgId, userId)
	if err != nil {
		return nil, customErr.DB(err)
	}
	return &img, nil
}

func (r *mysqlImagesRepo) GetAllPublic(ctx context.Context) ([]*dbModels.Image, error) {
	dbImgs := []*dbModels.Image{}
	err := r.db.Select(&dbImgs, "SELECT * FROM images WHERE private=False AND archived=False")
	if err != nil {
		return nil, customErr.DB(err)
	}
	return dbImgs, nil
}

func (r *mysqlImagesRepo) GetByFilter(filter string) ([]*dbModels.Image, error) {
	dbImgs := []*dbModels.Image{}
	err := r.db.Select(&dbImgs, filter)
	if err != nil {
		return nil, customErr.DB(err)
	}
	return dbImgs, nil
}

func (r *mysqlImagesRepo) Create(dbImg *dbModels.Image) (imgId int64, err error) {

	result, err := r.db.NamedExec(`INSERT INTO images(title, description, private, forSale, price, discountPercent, user_id, 
	created_at, url) VALUES(:title, :description, :private, :forSale, :price, :discountPercent ,:user_id, :created_at, :url)`, dbImg)
	if err != nil {
		return -1, customErr.DB(err)
	}
	imgId, _ = result.LastInsertId()

	return imgId, nil
}

func (r *mysqlImagesRepo) Update(id int, img *dbModels.Image) error {

	_, err := r.db.NamedExec(fmt.Sprintf(`UPDATE images SET title= :title, forSale= :forSale, private= :private, 
	description= :description, price= :price, discountPercent= :discountPercent, archived= :archived, url= :url WHERE id=%d`, id), img)
	if err != nil {
		return customErr.DB(err)
	}

	return nil
}

func (r *mysqlImagesRepo) Delete(imgId int, userId int) error {

	img, err := r.GetImageIfOwner(imgId, userId)
	if err != nil {
		return err
	}
	c, err := r.CountImageSales(img.ID)
	if err != nil {
		return err
	}
	if c != 0 {
		img.Archived = true
		err = r.Update(img.ID, img)
		if err != nil {
			return err
		}
		return nil
	}
	_, err = r.db.Exec("DELETE FROM images WHERE id=?", imgId)
	if err != nil {
		return customErr.DB(err)
	}
	err = r.DeleteImageLabels(imgId)
	if err != nil {
		return err
	}
	return nil
}

func (r *mysqlImagesRepo) GetImageLabels(imgId int) ([]string, error) {

	labels := []string{}
	err := r.db.Select(&labels, "SELECT tag FROM labels WHERE image_id=?", imgId)
	if err != nil {
		return nil, customErr.DB(err)
	}
	return labels, nil
}

func (r *mysqlImagesRepo) InsertImageLabels(imgId int, labels []*dbModels.Label) error {

	_, err := r.db.NamedExec("INSERT INTO labels(image_id, tag) VALUES(:image_id, :tag)", labels)
	if err != nil {
		return customErr.DB(err)
	}
	return nil
}

func (r *mysqlImagesRepo) CountImageSales(imgId int) (int, error) {

	c := 0
	err := r.db.Get(&c, "SELECT COUNT(*) FROM sales WHERE image_id=?", imgId)
	if err != nil {
		return -1, customErr.DB(err)
	}
	return c, nil
}

func (r *mysqlImagesRepo) DeleteImageLabels(imgId int) error {

	_, err := r.db.Exec("DELETE FROM labels WHERE image_id=?", imgId)
	if err != nil {
		return customErr.DB(err)
	}
	return nil
}

func (r *mysqlImagesRepo) checkUserBought(imgId int, userId int) bool {
	id := -1
	err := r.db.Get(&id, "SELECT id FROM sales WHERE image_id=? AND buyer_id=?", imgId, userId)
	return err == nil
}
