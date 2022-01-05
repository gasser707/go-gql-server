package errors

import (
	"context"
	"database/sql"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql"
	_ "github.com/joho/godotenv/autoload"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

var env = "ENV"

var errCodeMap = map[int]string{
	http.StatusBadRequest:          "There is a problem with the input you sent",
	http.StatusUnauthorized:        "Your auth credentials are invalid",
	http.StatusNotFound:            "No results found",
	http.StatusForbidden:           "You are trying to access a resource that doesn't belong to you",
	http.StatusInternalServerError: "Sorry! There seems to be a problem on our end",
}

func NewError(ctx context.Context, message string, code int) *gqlerror.Error {
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

func Internal(ctx context.Context, message string) *gqlerror.Error {
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

func NoAuth(ctx context.Context, message string) *gqlerror.Error {
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

func BadRequest(ctx context.Context, message string) *gqlerror.Error {
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

func UnProcessable(ctx context.Context, message string) *gqlerror.Error {
	code := http.StatusUnprocessableEntity
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

func Forbidden(ctx context.Context, message string) *gqlerror.Error {
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

func NotFound(ctx context.Context, message string) *gqlerror.Error {
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

func DB(ctx context.Context, err error) *gqlerror.Error {
	if err == sql.ErrNoRows {
		return NotFound(ctx, err.Error())
	}
	return Internal(ctx, err.Error())
}
