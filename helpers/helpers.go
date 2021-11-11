package helpers

import (
	"crypto/rand"

	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
)


var SecureCookieManager *securecookie.SecureCookie

func init(){
	// Hash keys should be at least 32 bytes long

key := make([]byte, 32)
_, err := rand.Read(key)
if err != nil {
	// handle error here
	println(err.Error())
}

var hashKey = key
// Block keys should be 16 bytes (AES-128) or 32 bytes (AES-256) long.
// Shorter keys may weaken the encryption used.
var blockKey = key

SecureCookieManager = securecookie.New(hashKey, blockKey)
}


func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
