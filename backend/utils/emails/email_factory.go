package emails

import (
	"github.com/matcornic/hermes/v2"
)

type EmailFactoryInterface interface {
	GenerateEmailContent(email EmailInterface) string
}

//emailFactory implements the EmailFactoryInterface
var _ EmailFactoryInterface = &emailFactory{}

type emailFactory struct {
	maker *hermes.Hermes
}

func NewEmailFactory() *emailFactory {

	return &emailFactory{
		maker: &hermes.Hermes{
			Product: hermes.Product{
				Name:      "Shotify",
				Link:      "https://example-hermes.com/",
				Logo:      "http://www.duchess-france.org/wp-content/uploads/2016/01/gopher.png",
				Copyright: "Copyright © 2022 Gasser Aly. All rights reserved.",
			},
		}}
}

type EmailType string

const (
	Welcome       EmailType = "Welcome"
	ResetPassword EmailType = "ResetPassword"
	Receipt       EmailType = "Receipt"
	Promotion     EmailType = "Promotion"
)

type EmailInterface interface {
	GetType() EmailType
	GetSender() string
	GetTo() []string
	GetName() string
	GetVerificationLink() string
}

type ResetPassEmailInterface interface {
	EmailInterface
	GetResetLink() string
}

type ReceiptEmailInterface interface {
	EmailInterface
	GetSellerName() string
	GetImageTitle() string
	GetImageID() string
	GetPaymentMethod() string
}

func (e Email) GetType() EmailType {
	return e.Type
}

func (e Email) GetSender() string {
	return e.Sender
}

func (e Email) GetTo() []string {
	return e.To
}

func (e Email) GetName() string {
	return e.Name
}

func (e Email) GetVerificationLink() string {
	return e.Link
}

func (e ResetPassEmail) GetResetLink() string {
	return e.ResetLink
}

func (e ReceiptEmail) GetSellerName() string {
	return e.SellerName
}

func (e ReceiptEmail) GetImageTitle() string {
	return e.ImageTitle
}
func (e ReceiptEmail) GetImageID() string {
	return e.ImageID
}
func (e ReceiptEmail) GetPaymentMethod() string {
	return e.PaymentMethod
}

type Email struct {
	Type   EmailType
	Sender string
	To     []string
	Name   string
	Link   string
}

type ResetPassEmail struct {
	Email
	ResetLink string
}

type ReceiptEmail struct {
	Email
	SellerName    string
	ImageTitle    string
	ImageID       string
	PaymentMethod string
}

func (f *emailFactory) GenerateEmailContent(email EmailInterface) string {
	var emailContent hermes.Email
	switch email.GetType() {
	case Welcome:
		emailContent = f.generateWelcomeEmail(email)
	case ResetPassword:
		emailContent = f.generateResetPasswordEmail(email.(ResetPassEmailInterface))
	default:
		emailContent = f.generateWelcomeEmail(email)
	}

	emailBody, _ := f.maker.GenerateHTML(emailContent)
	return emailBody
}

func (f *emailFactory) generateWelcomeEmail(email EmailInterface) hermes.Email {
	emailContent := hermes.Email{
		Body: hermes.Body{
			Name: email.GetName(),
			Intros: []string{
				"Welcome to Shotify! We're very excited to have you on board.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "To get started with Hermes, please click here to verify your account:",
					Button: hermes.Button{
						Color: "#22BC66",
						Text:  "Confirm your account",
						Link:  email.GetVerificationLink(),
					},
				},
			},
			Outros: []string{
				"Need help, or have questions? Just reply to this email, we'd love to help.",
			},
		},
	}

	return emailContent
}

func (f *emailFactory) generateResetPasswordEmail(email ResetPassEmailInterface) hermes.Email {
	emailContent := hermes.Email{
		Body: hermes.Body{
			Name: email.GetTo()[0],
			Intros: []string{
				"You have asked to reset your password",
			},
			Actions: []hermes.Action{
				{
					Instructions: "To reset click here",
					Button: hermes.Button{
						Color: "#22BC66",
						Text:  "Reset your password",
						Link:  email.GetResetLink(),
					},
				},
			},
			Outros: []string{
				"Need help, or have questions? Just reply to this email, we'd love to help.",
			},
		},
	}
	return emailContent
}
