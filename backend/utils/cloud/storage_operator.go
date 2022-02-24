package cloud

import (
	"context"
	"fmt"
	"io"
	"time"

	gcs "cloud.google.com/go/storage"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/gasser707/go-gql-server/utils"
)

type StorageOperatorInterface interface {
	UploadImage(img io.Reader, imgName string, userId string) (url string, err error)
	DeleteImage(path string) error
	ChangeImagePath(oldPath string, newPath string) (newUrl string, err error)
}

type GcsClient struct {
	client *gcs.Client
}

type storageOperator struct {
	storageClient StorageOperatorInterface
}

//storageOperator implements the StorageOperatorInterfaceInterface
var _ StorageOperatorInterface = &storageOperator{}

func NewGcsClient() (*GcsClient, error) {

	client, err := gcs.NewClient(context.Background())
	if err != nil {
		return nil, customErr.Internal(err.Error())
	}

	return &GcsClient{client: client}, nil
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
	sw := c.client.Bucket(utils.BucketName).Object(productId + "/" + imgName).NewWriter(ctx)
	if _, err = io.Copy(sw, img); err != nil {
		return "", customErr.Internal(err.Error())
	}
	if err := sw.Close(); err != nil {
		return "", customErr.Internal(err.Error())
	}

	url = fmt.Sprintf("%s/%s/%s", utils.BaseGcsUrl, utils.BucketName, sw.Attrs().Name)
	return url, nil
}

func (c *GcsClient) DeleteImage(path string) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	o := c.client.Bucket(utils.BucketName).Object(path)
	if err := o.Delete(ctx); err != nil {
		return customErr.Internal(err.Error())
	}

	return nil

}

func (c *GcsClient) ChangeImagePath(oldPath string, newPath string) (newUrl string, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	src := c.client.Bucket(utils.BucketName).Object(oldPath)
	dst := c.client.Bucket(utils.BucketName).Object(newPath)

	if _, err := dst.CopierFrom(src).Run(ctx); err != nil {
		return "", customErr.Internal(fmt.Sprintf("Object(%q).CopierFrom(%q).Run: %v", newPath, oldPath, err))
	}
	if err := src.Delete(ctx); err != nil {
		return "", customErr.Internal(fmt.Sprintf("Object(%q).Delete: %v", oldPath, err))
	}
	newUrl = fmt.Sprintf("%s/%s/%s", utils.BaseGcsUrl, utils.BucketName, newPath)
	return newUrl, nil
}

func (s *storageOperator) UploadImage(img io.Reader, imgName string, productId string) (url string, err error) {
	return s.storageClient.UploadImage(img, imgName, productId)
}

// deleteFile removes specified object.
func (s *storageOperator) DeleteImage(path string) error {
	return s.storageClient.DeleteImage(path)
}

func (s *storageOperator) ChangeImagePath(oldPath string, newPath string) (newUrl string, err error) {
	return s.storageClient.ChangeImagePath(oldPath, newPath)

}
