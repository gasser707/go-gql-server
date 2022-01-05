package middleware

import (
	"context"
	"net/http"
	"os"
	"time"

	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	_ "github.com/joho/godotenv/autoload"
)

var env = "ENV"
var cookieKey = "cookie-name"

type CookieAccess struct {
	Writer        http.ResponseWriter
	EncodedCookie string
}

// method to write cookie
func (ca *CookieAccess) SetCookie(at string, rt string, sm *securecookie.SecureCookie) {

	value := map[string]string{
		"access_token":  at,
		"refresh_token": rt,
	}

	if encoded, err := sm.Encode("cookie-name", value); err == nil {
		cookie := &http.Cookie{
			Name:  "cookie-name",
			Value: encoded,
			Path:  "/*",
			// Secure:   true,
			SameSite: http.SameSiteLaxMode,
			HttpOnly: true,
			Expires:  time.Now().Add(time.Hour * 24 * 7),
		}
		if os.Getenv(env) == "dev" {
			cookie.Secure = false
		}
		http.SetCookie(ca.Writer, cookie)
		ca.EncodedCookie = encoded
	} else {
		println(err.Error())
	}

}

func setValInCtx(ctx *gin.Context, key string, val interface{}) {
	newCtx := context.WithValue(ctx.Request.Context(), key, val)
	ctx.Request = ctx.Request.WithContext(newCtx)
}

func GetCookieAccess(ctx context.Context) (*CookieAccess, error) {
	cookieKey := cookieKey
	ca, ok := ctx.Value(cookieKey).(*CookieAccess)
	if !ok {
		return nil, customErr.NoAuth(ctx, "cookie not found")
	}
	return ca, nil
}

func CookieMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		encodedCookie := ""
		ca, err := ctx.Request.Cookie(string(cookieKey))
		if err == nil {
			encodedCookie = ca.Value
		}
		cookieA := CookieAccess{
			Writer:        ctx.Writer,
			EncodedCookie: encodedCookie,
		}

		// &cookieA is a pointer so any changes in future is changing cookieA is context
		setValInCtx(ctx, cookieKey, &cookieA)

		// calling the actual resolver
		ctx.Next()
	}
}
