package cloud

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	gcs "cloud.google.com/go/storage"
    customErr "github.com/gasser707/go-gql-server/errors"
)

type StorageOperatorInterface interface {
	UploadImage(img io.Reader, imgName string, productId string) (url string, err error)
	DeleteImage(path string) error
}

type GcsClient struct {
	client *gcs.Client
}

type storageOperator struct {
	storageClient StorageOperatorInterface
}

//UsersService implements the usersServiceInterface
var _ StorageOperatorInterface = &storageOperator{}

var bucketName = os.Getenv("BUCKET_NAME")

const baseGcsUrl = "https://storage.googleapis.com"

func NewGcsClient() (*GcsClient, error) {

	storageClient, err := gcs.NewClient(context.Background())
	if err != nil {
		return nil, customErr.Internal(err.Error())
	}

	return &GcsClient{client: storageClient}, nil
}

func NewStorageOperator(client StorageOperatorInterface) *storageOperator {

	return &storageOperator{
		storageClient: client,
	}
}

func (c *GcsClient) UploadImage(img io.Reader, imgName string,
	productId string) (url string, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	sw := c.client.Bucket(bucketName).Object(productId + "/" + imgName).NewWriter(ctx)
	if _, err = io.Copy(sw, img); err != nil {
		return "", customErr.Internal(err.Error())
	}
	if err := sw.Close(); err != nil {
		return "", customErr.Internal(err.Error())
	}

	url = fmt.Sprintf("%s/%s/%s", baseGcsUrl, bucketName, sw.Attrs().Name)
	return url, nil
}

func (c *GcsClient) DeleteImage(path string) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	o := c.client.Bucket(bucketName).Object(path)
	if err := o.Delete(ctx); err != nil {
		return customErr.Internal(err.Error())
	}

	return nil

}

func (s *storageOperator) UploadImage(img io.Reader, imgName string, productId string) (url string, err error) {
	return s.storageClient.UploadImage(img, imgName, productId)
}

// deleteFile removes specified object.
func (s *storageOperator) DeleteImage(path string) error {
	return s.storageClient.DeleteImage(path)
}
