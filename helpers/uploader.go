package helpers

import (
	"context"
	"fmt"
	"io"
	"os"

	"cloud.google.com/go/storage"
	"github.com/99designs/gqlgen/graphql"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/api/option"
)

type UploaderInterface interface {
	UploadImage(ctx context.Context, img *graphql.Upload, imgName string, path string) (url string, err error)
}

//UsersService implements the usersServiceInterface
var _ UploaderInterface = &uploader{}

var bucketName = "BUCKET_NAME"

type uploader struct {
	storageClient *storage.Client
}

func NewUploader(ctx context.Context) (*uploader, error) {

	storageClient, err:= storage.NewClient(ctx, option.WithCredentialsFile("bucket-keys.json"))
	if(err!=nil){
		return nil, err
	}
	return &uploader{
		storageClient: storageClient,
	}, nil
}

func (u *uploader) UploadImage(ctx context.Context, img *graphql.Upload, imgName string, path string) (url string, err error) {
	bucket :=  os.Getenv(bucketName)

	sw := u.storageClient.Bucket(bucket).Object(path+"/"+imgName).NewWriter(ctx)
	if _, err = io.Copy(sw, img.File); err != nil {
			return "", fmt.Errorf("io.Copy: %v", err)
	}
	if err := sw.Close(); err != nil {
			return "", fmt.Errorf("Writer.Close: %v", err)
	}

	url = "/" + bucket + "/" + sw.Attrs().Name
	return url, nil
}
