package errors

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	_ "github.com/joho/godotenv/autoload"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"os"
)

var env = "ENV"
var devEnv = "dev"

func NewError(ctx context.Context, err error, message string, code string) *gqlerror.Error {
	newErr := &gqlerror.Error{
		Message: message,
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
	if os.Getenv(env) == devEnv {
		newErr.Path = graphql.GetPath(ctx)
		newErr.Message = err.Error()
	}
	return newErr
}
