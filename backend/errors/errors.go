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

func NewError(message string, code int) *gqlerror.Error {
	newErr := &gqlerror.Error{
		Message: errCodeMap[code],
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
	if os.Getenv(env) == "dev" {
		newErr.Path = graphql.GetPath(context.Background())
		newErr.Message = message
	}
	return newErr
}

func Internal(message string) *gqlerror.Error {
	code := http.StatusInternalServerError
	newErr := &gqlerror.Error{
		Message: errCodeMap[code],
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
	if os.Getenv(env) == "dev" {
		newErr.Path = graphql.GetPath(context.Background())
		newErr.Message = message
	}
	return newErr
}

func NoAuth(message string) *gqlerror.Error {
	code := http.StatusUnauthorized
	newErr := &gqlerror.Error{
		Message: errCodeMap[code],
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
	if os.Getenv(env) == "dev" {
		newErr.Path = graphql.GetPath(context.Background())
		newErr.Message = message
	}
	return newErr
}

func BadRequest(message string) *gqlerror.Error {
	code := http.StatusBadRequest
	newErr := &gqlerror.Error{
		Message: errCodeMap[code],
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
	if os.Getenv(env) == "dev" {
		newErr.Path = graphql.GetPath(context.Background())
		newErr.Message = message
	}
	return newErr
}

func UnProcessable(message string) *gqlerror.Error {
	code := http.StatusUnprocessableEntity
	newErr := &gqlerror.Error{
		Message: errCodeMap[code],
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
	if os.Getenv(env) == "dev" {
		newErr.Path = graphql.GetPath(context.Background())
		newErr.Message = message
	}
	return newErr
}

func Forbidden(message string) *gqlerror.Error {
	code := http.StatusForbidden
	newErr := &gqlerror.Error{
		Message: errCodeMap[code],
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
	if os.Getenv(env) == "dev" {
		newErr.Path = graphql.GetPath(context.Background())
		newErr.Message = message
	}
	return newErr
}

func NotFound(message string) *gqlerror.Error {
	code := http.StatusNotFound
	newErr := &gqlerror.Error{
		Message: errCodeMap[code],
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
	if os.Getenv(env) == "dev" {
		newErr.Path = graphql.GetPath(context.Background())
		newErr.Message = message
	}
	return newErr
}

func DB(err error) *gqlerror.Error {
	if err == sql.ErrNoRows {
		return NotFound(err.Error())
	}
	return Internal(err.Error())
}
