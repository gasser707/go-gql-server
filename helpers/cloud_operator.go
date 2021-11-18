package helpers

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/99designs/gqlgen/graphql"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/api/option"
)

type CloudOperatorInterface interface {
	UploadImage(ctx context.Context, img *graphql.Upload, imgName string, path string) (url string, err error)
	DeleteImage(ctx context.Context, path string) error
}

//UsersService implements the usersServiceInterface
var _ CloudOperatorInterface = &cloudOperator{}

var bucketName = "BUCKET_NAME"

type cloudOperator struct {
	storageClient *storage.Client
}

func NewCloudOperator(ctx context.Context) (*cloudOperator, error) {

	storageClient, err := storage.NewClient(ctx, option.WithCredentialsFile("bucket-keys.json"))
	if err != nil {
		return nil, err
	}
	return &cloudOperator{
		storageClient: storageClient,
	}, nil
}

func (co *cloudOperator) UploadImage(ctx context.Context, img *graphql.Upload, imgName string, path string) (url string, err error) {
	bucket := os.Getenv(bucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	sw := co.storageClient.Bucket(bucket).Object(path + "/" + imgName).NewWriter(ctx)
	if _, err = io.Copy(sw, img.File); err != nil {
		return "", fmt.Errorf("io.Copy: %v", err)
	}
	if err := sw.Close(); err != nil {
		return "", fmt.Errorf("Writer.Close: %v", err)
	}

	url = "/" + bucket + "/" + sw.Attrs().Name
	return url, nil
}



// deleteFile removes specified object.
func (co *cloudOperator) DeleteImage(ctx context.Context, path string ) error {
	bucket := os.Getenv(bucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	o := co.storageClient.Bucket(bucket).Object(path)
	if err := o.Delete(ctx); err != nil {
			return fmt.Errorf("Object(%q).Delete: %v", path, err)
	}

	return nil
}