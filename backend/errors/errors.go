package errors

import (
	"context"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql"
	_ "github.com/joho/godotenv/autoload"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

var env = "ENV"

var errCodeMap = map[int]string{
	400: "There is a problem with the input you sent",
	401:"Your auth credentials are invalid",
	404: "No results found",
	403:"You are trying to access a resource that doesn't belong to you",
	500: "Sorry! There seems to be a problem on our end",
}

func NewError(ctx context.Context,message string, code int) *gqlerror.Error {
	newErr := &gqlerror.Error{
		Message: errCodeMap[code],
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
	if os.Getenv(env) == "dev" {
		newErr.Path = graphql.GetPath(ctx)
		newErr.Message = message
	}
	return newErr
}


func Internal(ctx context.Context,message string) *gqlerror.Error {
	code := http.StatusInternalServerError
	newErr := &gqlerror.Error{
		Message: errCodeMap[code],
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
	if os.Getenv(env) == "dev" {
		newErr.Path = graphql.GetPath(ctx)
		newErr.Message = message
	}
	return newErr
}


func NoAuth(ctx context.Context,message string) *gqlerror.Error {
	code := http.StatusUnauthorized
	newErr := &gqlerror.Error{
		Message: errCodeMap[code],
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
	if os.Getenv(env) == "dev" {
		newErr.Path = graphql.GetPath(ctx)
		newErr.Message = message
	}
	return newErr
}


func BadRequest(ctx context.Context,message string) *gqlerror.Error {
	code := http.StatusBadRequest
	newErr := &gqlerror.Error{
		Message: errCodeMap[code],
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
	if os.Getenv(env) == "dev" {
		newErr.Path = graphql.GetPath(ctx)
		newErr.Message = message
	}
	return newErr
}

func Forbidden(ctx context.Context,message string) *gqlerror.Error {
	code := http.StatusForbidden
	newErr := &gqlerror.Error{
		Message: errCodeMap[code],
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
	if os.Getenv(env) == "dev" {
		newErr.Path = graphql.GetPath(ctx)
		newErr.Message = message
	}
	return newErr
}

func NotFound(ctx context.Context,message string) *gqlerror.Error {
	code := http.StatusNotFound
	newErr := &gqlerror.Error{
		Message: errCodeMap[code],
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
	if os.Getenv(env) == "dev" {
		newErr.Path = graphql.GetPath(ctx)
		newErr.Message = message
	}
	return newErr
}



