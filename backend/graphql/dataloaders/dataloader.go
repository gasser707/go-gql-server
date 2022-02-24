package dataloaders

import (
	"context"
	"fmt"
	"time"

	dbModels "github.com/gasser707/go-gql-server/databases/models"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/gasser707/go-gql-server/graphql/custom"
	"github.com/jmoiron/sqlx"
)

//go:generate go run github.com/vektah/dataloaden UserLoader int *github.com/gasser707/go-gql-server/graphql/custom.User

//go:generate go run github.com/vektah/dataloaden ImageLoader int []*github.com/gasser707/go-gql-server/graphql/custom.Image

//go:generate go run github.com/vektah/dataloaden SaleImageLoader int *github.com/gasser707/go-gql-server/graphql/custom.Image

type contextKey string

const Key = contextKey("dataloaders")

// Loaders holds references to the individual dataloaders.
type loaders struct {
	// individual loaders will be defined here
	UserByID       *UserLoader
	ImagesByUserID *ImageLoader
	ImageByID      *SaleImageLoader
}

func NewLoaders(ctx context.Context, db *sqlx.DB) *loaders {
	return &loaders{
		UserByID:       newUserByID(ctx, db),
		ImagesByUserID: newImagesByUserID(ctx, db),
		ImageByID:      newImageByID(ctx, db),
	}
}

// RetrieverInterface retrieves dataloaders from the request context.
type RetrieverInterface interface {
	Retrieve(context.Context) *loaders
}

type retriever struct {
	key contextKey
}

func (r *retriever) Retrieve(ctx context.Context) *loaders {
	return ctx.Value(r.key).(*loaders)
}

// NewRetriever instantiates a new implementation of Retriever.
func NewRetriever() RetrieverInterface {
	return &retriever{key: Key}
}

func newUserByID(ctx context.Context, db *sqlx.DB) *UserLoader {
	return NewUserLoader(UserLoaderConfig{
		MaxBatch: 100,
		Wait:     5 * time.Millisecond,
		Fetch: func(ids []int) ([]*custom.User, []error) {
			dbUsers := []*dbModels.User{}
			query, args, err := sqlx.In("SELECT * FROM users WHERE id IN (?)", ids)
			if err != nil {
				return nil, []error{customErr.DB(err)}
			}
			query = db.Rebind(query)
			err = db.Select(&dbUsers, query, args...)
			if err != nil {
				return nil, []error{customErr.DB(err)}
			}

			m := make(map[int]*custom.User, len(dbUsers))

			for _, user := range dbUsers {
				m[user.ID] = &custom.User{
					ID:       fmt.Sprintf("%v", user.ID),
					Username: user.Username,
					Email:    user.Email,
					Avatar:   user.Avatar,
					Joined:   &user.CreatedAt,
					Bio:      user.Bio,
				}
			}

			result := make([]*custom.User, len(ids))

			for i, id := range ids {
				if val, ok := m[id]; ok {
					result[i] = val
				}
			}
			return result, nil
		},
	})
}

func newImagesByUserID(ctx context.Context, db *sqlx.DB) *ImageLoader {
	return NewImageLoader(ImageLoaderConfig{
		MaxBatch: 100,
		Wait:     5 * time.Millisecond,
		Fetch: func(ids []int) ([][]*custom.Image, []error) {
			dbImages := []*dbModels.Image{}
			query, args, err := sqlx.In("SELECT * FROM images WHERE user_id IN (?)", ids)
			if err != nil {
				return nil, []error{customErr.DB(err)}
			}
			query = db.Rebind(query)
			err = db.Select(&dbImages, query, args...)
			if err != nil {
				return nil, []error{customErr.DB(err)}
			}

			m := make(map[int][]*custom.Image, len(dbImages))

			for _, id := range ids {
				for _, img := range dbImages {
					if id == img.UserID {
						m[id] = append(m[id], &custom.Image{
							ID:              fmt.Sprintf("%v", img.ID),
							UserID:          fmt.Sprintf("%v", img.UserID),
							Created:         &img.CreatedAt,
							Title:           img.Title,
							URL:             img.URL,
							Description:     img.Description,
							Private:         img.Private,
							ForSale:         img.ForSale,
							Price:           img.Price,
							DiscountPercent: img.DiscountPercent,
							Archived:        img.Archived,
						})
					}
				}
			}

			result := make([][]*custom.Image, len(ids))

			for i, id := range ids {
				result[i] = m[id]
			}

			return result, nil
		},
	})

}

func newImageByID(ctx context.Context, db *sqlx.DB) *SaleImageLoader {
	return NewSaleImageLoader(SaleImageLoaderConfig{
		MaxBatch: 100,
		Wait:     5 * time.Millisecond,
		Fetch: func(ids []int) ([]*custom.Image, []error) {
			dbImages := []*dbModels.Image{}
			query, args, err := sqlx.In("SELECT * FROM images WHERE id IN (?)", ids)
			if err != nil {
				return nil, []error{customErr.DB(err)}
			}
			query = db.Rebind(query)
			err = db.Select(&dbImages, query, args...)
			if err != nil {
				return nil, []error{customErr.DB(err)}
			}

			m := make(map[int]*custom.Image, len(dbImages))

			for _, img := range dbImages {
				m[img.ID] = &custom.Image{
					ID:              fmt.Sprintf("%v", img.ID),
					UserID:          fmt.Sprintf("%v", img.UserID),
					Created:         &img.CreatedAt,
					Title:           img.Title,
					URL:             img.URL,
					Description:     img.Description,
					Private:         img.Private,
					ForSale:         img.ForSale,
					Price:           img.Price,
					DiscountPercent: img.DiscountPercent,
					Archived:        img.Archived,
				}
			}

			result := make([]*custom.Image, len(ids))
			for i, id := range ids {
				if val, ok := m[id]; ok {
					result[i] = val
				}
			}

			return result, nil
		},
	})

}
