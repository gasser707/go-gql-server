package cloud

import (
	"context"
	"fmt"
	"os"
	"strings"

	vision "cloud.google.com/go/vision/apiv1"
	"golang.org/x/sync/errgroup"
	visionpb "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

type VisionOperatorInterface interface {
	DetectImgProps(ctx context.Context, source string, limit int) (labels []string, err error)
}

//UsersService implements the usersServiceInterface
var _ VisionOperatorInterface = &visionOperator{}

type visionOperator struct {
	visionClient *vision.ImageAnnotatorClient
}

func NewVisionOperator(ctx context.Context) (*visionOperator, error) {

	visionClient, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, err
	}
	return &visionOperator{
		visionClient: visionClient,
	}, nil
}

func (v *visionOperator) DetectImgProps(ctx context.Context, source string, limit int) (labels []string, err error) {
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
			return v.getLandMarks(ctx, ch, image, 2)
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
		return err
	}
	for _, annotation := range annotations {
		ch <- strings.ToLower(annotation.Description)
	}

	return nil

}

func (v *visionOperator) getLandMarks(ctx context.Context, ch chan string, img *visionpb.Image, limit int) (err error) {

	annotations, err := v.visionClient.DetectLandmarks(ctx, img, nil, limit)
	if err != nil {
		return err
	}
	for _, annotation := range annotations {
		ch <- strings.ToLower(annotation.Description)
	}

	return nil

}

func (v *visionOperator) getLogos(ctx context.Context, ch chan string, img *visionpb.Image, limit int) (err error) {

	annotations, err := v.visionClient.DetectLogos(ctx, img, nil, limit)
	if err != nil {
		return err
	}
	for _, annotation := range annotations {
		ch <- strings.ToLower(annotation.Description)
	}

	return nil

}

func (v *visionOperator) getObjects(ctx context.Context, ch chan string, img *visionpb.Image) (err error) {

	annotations, err := v.visionClient.LocalizeObjects(ctx, img, nil)
	if err != nil {
		return err
	}
	for _, annotation := range annotations {
		ch <- strings.ToLower( annotation.Name)
	}

	return nil
}
