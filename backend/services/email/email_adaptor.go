package services

import (
	"context"
	"github.com/gasser707/go-gql-server/utils/emails"
	"log"
)

type EmailAdaptorInterface interface {
	SendWelcomeEmail(ctx context.Context, sender string, to []string, name string)
	SendResetPassEmail(ctx context.Context, sender string, to []string, name string, resetLink string)
	SendReceiptEmail(ctx context.Context, sender string, to []string, sellerName string,
		buyerName string, imageID string, imageTitle string, paymentMethod string)
}

//emailAdaptor implements the EmailAdaptorInterface
var _ EmailAdaptorInterface = &emailAdaptor{}

type emailAdaptor struct {
	emailService EmailServiceInterface
}

func NewEmailAdaptor(emailService EmailServiceInterface) *emailAdaptor {
	return &emailAdaptor{emailService: emailService}
}

func (ea *emailAdaptor) SendWelcomeEmail(ctx context.Context, sender string, to []string, name string) {
	email := &emails.Email{
		Type:   emails.Welcome,
		Sender: sender,
		To:     to,
		Name:   name,
	}

	err := ea.emailService.SendEmail(ctx, email)
	if err != nil {
		log.Println("couldn't send email\n", err.Error())
	}
}

func (ea *emailAdaptor) SendResetPassEmail(ctx context.Context, sender string, to []string, name string, resetLink string) {

	email := &emails.ResetPassEmail{
		ResetLink: resetLink,
		Email: emails.Email{
			Type:   emails.ResetPassword,
			Sender: sender,
			To:     to,
			Name:   name,
		},
	}

	err := ea.emailService.SendEmail(ctx, email)
	if err != nil {
		log.Println("couldn't send email\n", err.Error())
	}

}

func (ea *emailAdaptor) SendReceiptEmail(ctx context.Context, sender string, to []string, sellerName string,
	buyerName string, imageID string, imageTitle string, paymentMethod string) {

	email := &emails.ReceiptEmail{
		SellerName:    sellerName,
		ImageID:       imageID,
		ImageTitle:    imageTitle,
		PaymentMethod: paymentMethod,
		Email: emails.Email{
			Type:   emails.ResetPassword,
			Sender: sender,
			To:     to,
			Name:   buyerName,
		},
	}

	err := ea.emailService.SendEmail(ctx, email)
	if err != nil {
		log.Println("couldn't send email\n", err.Error())
	}

}
