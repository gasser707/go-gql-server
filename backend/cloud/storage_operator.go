package cloud

import (
	"context"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/99designs/gqlgen/graphql"
	customErr "github.com/gasser707/go-gql-server/errors"
)

type StorageOperatorInterface interface {
	UploadImage(ctx context.Context, img *graphql.Upload, imgName string, path string) (url string, err error)
	DeleteImage(ctx context.Context, path string) error
}

//UsersService implements the usersServiceInterface
var _ StorageOperatorInterface = &storageOperator{}

var bucketName = "BUCKET_NAME"

type storageOperator struct {
	storageClient *storage.Client
}

func NewStorageOperator(ctx context.Context) (*storageOperator, error) {

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, customErr.Internal(ctx, err.Error())
	}
	return &storageOperator{
		storageClient: storageClient,
	}, nil
}

func (s *storageOperator) UploadImage(ctx context.Context, img *graphql.Upload, imgName string, userId string) (url string, err error) {
	bucket := os.Getenv(bucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()
	sw := s.storageClient.Bucket(bucket).Object(userId + "/" + imgName).NewWriter(ctx)
	if _, err = io.Copy(sw, img.File); err != nil {
		return "", customErr.Internal(ctx, err.Error())
	}
	if err := sw.Close(); err != nil {
		return "", customErr.Internal(ctx, err.Error())
	}

	url = sw.Attrs().Name
	return url, nil
}

// deleteFile removes specified object.
func (s *storageOperator) DeleteImage(ctx context.Context, path string) error {
	bucket := os.Getenv(bucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	o := s.storageClient.Bucket(bucket).Object(path)
	if err := o.Delete(ctx); err != nil {
		return customErr.Internal(ctx, err.Error())
	}

	return nil
}
