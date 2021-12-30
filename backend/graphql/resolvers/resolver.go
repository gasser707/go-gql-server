package resolvers

import (
	"github.com/gasser707/go-gql-server/services"
	email_svc"github.com/gasser707/go-gql-server/services/email"

)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	UsersService  services.UsersServiceInterface
	ImagesService services.ImagesServiceInterface
	AuthService   services.AuthServiceInterface
	SaleService   services.SalesServiceInterface
	EmailService  email_svc.EmailServiceInterface
}
