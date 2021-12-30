package services

import (
	"context"
	"net/smtp"
	"os"

	// dbModels "github.com/gasser707/go-gql-server/databases/models"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/gasser707/go-gql-server/utils/emails"
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
)

const (
	emailAddress = "EMAIL_ADDRESS"
	emailHost    = "EMAIL_HOST"
)

var (
	addr      = os.Getenv(emailAddress)
	host      = os.Getenv(emailHost)
	emailAuth = smtp.PlainAuth("", "", "", host)
)

type EmailServiceInterface interface {
	SendEmail(ctx context.Context, email emails.EmailInterface) error
}

//UsersService implements the usersServiceInterface
var _ EmailServiceInterface = &emailService{}

type emailService struct {
	db      *sqlx.DB
	factory emails.EmailFactoryInterface
}

func NewEmailService(db *sqlx.DB) *emailService {
	return &emailService{
		db:      db,
		factory: emails.NewEmailFactory(),
	}
}

func (s *emailService) SendEmail(ctx context.Context, email emails.EmailInterface) error {

	emailContent := s.factory.GenerateEmailContent(email)
	emailContent =  "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n" + emailContent

	err := smtp.SendMail(addr, emailAuth, email.GetSender(), email.GetTo(), []byte(emailContent))
	if err != nil {
		return customErr.Internal(ctx, err.Error())
	}

	return nil
}
