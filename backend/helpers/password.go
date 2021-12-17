package helpers

import (
	"context"
	customErr "github.com/gasser707/go-gql-server/errors"
	"golang.org/x/crypto/bcrypt"
)
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), customErr.BadRequest(context.Background(), err.Error())
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

const UserIdKey = "userId"
