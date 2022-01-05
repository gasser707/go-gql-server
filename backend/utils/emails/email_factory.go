package emails

import (
	"github.com/matcornic/hermes/v2"
)

type EmailFactoryInterface interface {
	GenerateEmailContent(email EmailInterface) string
}

//UsersService implements the usersServiceInterface
var _ EmailFactoryInterface = &emailFactory{}

type emailFactory struct {
	maker *hermes.Hermes
}

func NewEmailFactory() *emailFactory {

	return &emailFactory{
		maker: &hermes.Hermes{
			// Optional Theme
			// Theme: new(Default)
			Product: hermes.Product{
				// Appears in header & footer of e-mails
				Name: "Shotify",
				Link: "https://example-hermes.com/",
				// Optional product logo
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
	switch email.GetType() {
	case Welcome:
		return f.generateWelcomeEmail(email)
	case ResetPassword:
		return f.generateResetPasswordEmail(email.(ResetPassEmailInterface))
	case Receipt:
		return f.generateReceiptEmail(email.(ReceiptEmailInterface))
	case Promotion:
		return f.generatePromotionEmail(email)
	}

	return f.generateWelcomeEmail(email)

}

func (f *emailFactory) generateWelcomeEmail(email EmailInterface) string {
	emailContent := hermes.Email{
		Body: hermes.Body{
			Name: email.GetName(),
			Intros: []string{
				"Welcome to Shotify! We're very excited to have you on board.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "To get started with Hermes, please click here:",
					Button: hermes.Button{
						Color: "#22BC66", // Optional action button color
						Text:  "Confirm your account",
						Link:  "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010",
					},
				},
			},
			Outros: []string{
				"Need help, or have questions? Just reply to this email, we'd love to help.",
			},
		},
	}

	body, _ := f.maker.GenerateHTML(emailContent)
	return body
}

func (f *emailFactory) generateResetPasswordEmail(email ResetPassEmailInterface) string {
	emailContent := hermes.Email{
		Body: hermes.Body{
			Name: email.GetTo()[0],
			Intros: []string{
				"Welcome to Shotify! We're very excited to have you on board.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "To get started with Hermes, please click here:",
					Button: hermes.Button{
						Color: "#22BC66", // Optional action button color
						Text:  "Confirm your account",
						Link:  email.GetResetLink(),
					},
				},
			},
			Outros: []string{
				"Need help, or have questions? Just reply to this email, we'd love to help.",
			},
		},
	}

	body, _ := f.maker.GenerateHTML(emailContent)
	return body
}

func (f *emailFactory) generateReceiptEmail(email ReceiptEmailInterface) string {
	emailContent := hermes.Email{
		Body: hermes.Body{
			Name: email.GetTo()[0],
			Intros: []string{
				"Welcome to Shotify! We're very excited to have you on board.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "To get started with Hermes, please click here:",
					Button: hermes.Button{
						Color: "#22BC66", // Optional action button color
						Text:  "Confirm your account",
						Link:  "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010",
					},
				},
			},
			Outros: []string{
				"Need help, or have questions? Just reply to this email, we'd love to help.",
			},
		},
	}

	body, _ := f.maker.GenerateHTML(emailContent)
	return body
}

func (f *emailFactory) generatePromotionEmail(email EmailInterface) string {
	emailContent := hermes.Email{
		Body: hermes.Body{
			Intros: []string{
				"Welcome to Shotify! We're very excited to have you on board.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "To get started with Hermes, please click here:",
					Button: hermes.Button{
						Color: "#22BC66", // Optional action button color
						Text:  "Confirm your account",
						Link:  "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010",
					},
				},
			},
			Outros: []string{
				"Need help, or have questions? Just reply to this email, we'd love to help.",
			},
		},
	}

	body, _ := f.maker.GenerateHTML(emailContent)
	return body
}
