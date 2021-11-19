package helpers

import (
	"os"
	"github.com/gorilla/securecookie"
	_ "github.com/joho/godotenv/autoload"
)



func NewSecureCookie()  *securecookie.SecureCookie{
	// Hash keys should be at least 32 bytes long

var hashKey = "COOKIE_HASH_KEY"
// Block keys should be 16 bytes (AES-128) or 32 bytes (AES-256) long.
// Shorter keys may weaken the encryption used.
var blockKey = "COOKIE_BLOCK_KEY"

SecureCookieManager := securecookie.New([]byte(os.Getenv(hashKey)), []byte(os.Getenv(blockKey)))
return SecureCookieManager
}
