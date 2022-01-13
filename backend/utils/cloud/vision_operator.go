package cloud

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	vision "cloud.google.com/go/vision/apiv1"
	customErr "github.com/gasser707/go-gql-server/errors"
	"golang.org/x/sync/errgroup"
	visionpb "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

type VisionOperatorInterface interface {
	DetectImgProps(ctx context.Context, source string) (labels []string, err error)
	DetectLocalImgProps(ctx context.Context, imgReader io.Reader) (labels []string, err error) 
}

//UsersService implements the usersServiceInterface
var _ VisionOperatorInterface = &visionOperator{}

type visionOperator struct {
	visionClient *vision.ImageAnnotatorClient
}

func NewVisionOperator(ctx context.Context) (*visionOperator, error) {

	visionClient, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, customErr.Internal(ctx, err.Error())
	}
	return &visionOperator{
		visionClient: visionClient,
	}, nil
}

func (v *visionOperator) DetectImgProps(ctx context.Context, source string) (labels []string, err error) {
	imgUrl := fmt.Sprintf("gs://%s/%s", os.Getenv(bucketName), source)
	image := vision.NewImageFromURI(imgUrl)
	errs, ctx := errgroup.WithContext(ctx)
	ch := make(chan string)
	errs.Go(
		func() error {
			return v.getLabels(ctx, ch, image, 6)
		})

	errs.Go(
		func() error {
			return v.getLandMarks(ctx, ch, image, 5)
		})
	errs.Go(
		func() error {
			return v.getLogos(ctx, ch, image, 3)
		})
	errs.Go(
		func() error {
			return v.getObjects(ctx, ch, image)
		})

	go func() {
		errs.Wait()
		close(ch)
	}()

	// read from channel as they come in until its closed
	labels = []string{}
	for label := range ch {
		labels = append(labels, label)
	}

	return labels, errs.Wait()

}

func (v *visionOperator) DetectLocalImgProps(ctx context.Context, imgReader io.Reader) (labels []string, err error) {

	image, err := vision.NewImageFromReader(imgReader)
	if err != nil {
		return nil, err
	}
	errs, ctx := errgroup.WithContext(ctx)
	ch := make(chan string)
	errs.Go(
		func() error {
			return v.getLabels(ctx, ch, image, 6)
		})

	errs.Go(
		func() error {
			return v.getLandMarks(ctx, ch, image, 5)
		})
	errs.Go(
		func() error {
			return v.getLogos(ctx, ch, image, 3)
		})
	errs.Go(
		func() error {
			return v.getObjects(ctx, ch, image)
		})

	go func() {
		errs.Wait()
		close(ch)
	}()

	// read from channel as they come in until its closed
	labels = []string{}
	for label := range ch {
		labels = append(labels, label)
	}

	return labels, errs.Wait()

}

func (v *visionOperator) getLabels(ctx context.Context, ch chan string, img *visionpb.Image, limit int) (err error) {

	annotations, err := v.visionClient.DetectLabels(ctx, img, nil, limit)
	if err != nil {
		return customErr.Internal(ctx, err.Error())
	}
	for _, annotation := range annotations {
		ch <- strings.ToLower(annotation.Description)
	}

	return nil

}

func (v *visionOperator) getLandMarks(ctx context.Context, ch chan string, img *visionpb.Image, limit int) (err error) {

	annotations, err := v.visionClient.DetectLandmarks(ctx, img, nil, limit)
	if err != nil {
		return customErr.Internal(ctx, err.Error())
	}
	for _, annotation := range annotations {
		ch <- strings.ToLower(annotation.Description)
	}

	return nil

}

func (v *visionOperator) getLogos(ctx context.Context, ch chan string, img *visionpb.Image, limit int) (err error) {

	annotations, err := v.visionClient.DetectLogos(ctx, img, nil, limit)
	if err != nil {
		return customErr.Internal(ctx, err.Error())
	}
	for _, annotation := range annotations {
		ch <- strings.ToLower(annotation.Description)
	}

	return nil

}

func (v *visionOperator) getObjects(ctx context.Context, ch chan string, img *visionpb.Image) (err error) {

	annotations, err := v.visionClient.LocalizeObjects(ctx, img, nil)
	if err != nil {
		return customErr.Internal(ctx, err.Error())
	}
	for _, annotation := range annotations {
		ch <- strings.ToLower(annotation.Name)
	}

	return nil
}
