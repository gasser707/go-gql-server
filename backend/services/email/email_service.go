package services

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	// dbModels "github.com/gasser707/go-gql-server/databases/models"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/gasser707/go-gql-server/utils/emails"
	_ "github.com/joho/godotenv/autoload"
)

var env = "ENV"

const (
	emailAddress = "EMAIL_ADDRESS"
	emailHost    = "EMAIL_HOST"
	sendGridKey  = "SENDGRID_API_KEY"
	sendGridFrom = "SENDGRID_FROM"
)

var (
	addr           = os.Getenv(emailAddress)
	host           = os.Getenv(emailHost)
	emailAuth      = smtp.CRAMMD5Auth("","")
	sendGridApiKey = os.Getenv(sendGridKey)
	sendGridEmail  = os.Getenv(sendGridFrom)
)

type EmailServiceInterface interface {
	SendEmail(ctx context.Context, email emails.EmailInterface) error
}

type EmailClientInterface interface {
	SendEmail(ctx context.Context, email emails.EmailInterface, emailContent string) error
}

type devEmailClient struct{}
type prodEmailClient struct{}

type emailService struct {
	factory     emails.EmailFactoryInterface
	emailClient EmailClientInterface
}

func (d *devEmailClient) SendEmail(ctx context.Context, email emails.EmailInterface, emailContent string) error {

	emailContent =
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n" +
			fmt.Sprintf("From: %s\r\n", email.GetSender()) +
			fmt.Sprintf("Subject: %s\r\n\r\n", email.GetType()) +
			emailContent

	err := smtp.SendMail(addr, emailAuth, email.GetSender(), email.GetTo(), []byte(emailContent))
	if err != nil {
		fmt.Println(err.Error())
		return customErr.Internal(ctx, err.Error())
	}

	return nil
}

func (p *prodEmailClient) SendEmail(ctx context.Context, email emails.EmailInterface, emailContent string) error {

	from := mail.NewEmail(email.GetSender(), os.Getenv(sendGridEmail))
	subject := string(email.GetType())
	// mail.
	fmt.Println(email.GetTo()[0])
	to := mail.NewEmail(email.GetName(), email.GetTo()[0])

	message := mail.NewSingleEmail(from, subject, to, emailContent, emailContent)
	client := sendgrid.NewSendClient(os.Getenv(sendGridApiKey))
	_, err := client.Send(message)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

//UsersService implements the usersServiceInterface
var _ EmailServiceInterface = &emailService{}

func NewEmailService() *emailService {
	var client EmailClientInterface
	if os.Getenv(env) == "dev" {
		client = &devEmailClient{}

	} else {
		client = &prodEmailClient{}
	}
	return &emailService{
		factory:     emails.NewEmailFactory(),
		emailClient: client,
	}
}

func (s *emailService) SendEmail(ctx context.Context, email emails.EmailInterface) error {

	emailContent := s.factory.GenerateEmailContent(email)
	err := s.emailClient.SendEmail(ctx, email, emailContent)
	if err != nil {
		return customErr.Internal(ctx, err.Error())
	}
	return nil
}
