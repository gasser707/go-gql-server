package middleware

import (
	"context"
	"net/http"
	"time"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
)

const cookieKey = "cookie-name" 

type CookieAccess struct {
    Writer     http.ResponseWriter
    EncodedCookie   string
}
// method to write cookie
func (ca *CookieAccess) SetToken(at string, rt string, sm *securecookie.SecureCookie) {

    value := map[string]string{
		"access_token": at,
        "refresh_token": rt,
	}

	if encoded, err := sm.Encode("cookie-name", value); err == nil {
		cookie := &http.Cookie{
			Name:  "cookie-name",
			Value: encoded,
			Path:  "/*",
			// Secure: true,
			HttpOnly: true,
            Expires: time.Now().Add(time.Hour*24*7),
		}
        http.SetCookie(ca.Writer,cookie)
        ca.EncodedCookie = encoded
	} else{
        println(err.Error())
    }
    
}


func setValInCtx(ctx *gin.Context, val interface{}) {
    cookieKey := cookieKey
    newCtx := context.WithValue(ctx.Request.Context(), cookieKey, val)
    ctx.Request = ctx.Request.WithContext(newCtx)
}

func GetCookieAccess(ctx context.Context) (*CookieAccess, error) {
    cookieKey := cookieKey
    ca, ok :=  ctx.Value(cookieKey).(*CookieAccess)
    if(!ok){
        return nil, customErr.NoAuth(ctx, "cookie not found")
    }
    return ca, nil
}

func Middleware() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        encodedCookie := ""
        cookieKey := cookieKey
        ca,err := ctx.Request.Cookie(string(cookieKey))
        if(err==nil){
            encodedCookie = ca.Value
        }
        cookieA := CookieAccess{
            Writer: ctx.Writer,
            EncodedCookie: encodedCookie,
        }

        // &cookieA is a pointer so any changes in future is changing cookieA is context
        setValInCtx(ctx, &cookieA)

       // calling the actual resolver
        ctx.Next()
    }
}