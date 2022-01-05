package resolvers

//go:generate go run github.com/99designs/gqlgen

import (
	"github.com/gasser707/go-gql-server/services"
	sale_svc"github.com/gasser707/go-gql-server/services/sale"
	email_svc"github.com/gasser707/go-gql-server/services/email"

)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	UsersService  services.UsersServiceInterface
	ImagesService services.ImagesServiceInterface
	AuthService   services.AuthServiceInterface
	SaleService   sale_svc.SalesServiceInterface
	EmailService  email_svc.EmailServiceInterface
}
