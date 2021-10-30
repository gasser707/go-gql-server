package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
)
type TokenDetails struct {
    userId string
    role string
    AccessToken  string
    RefreshToken string
    AccessUuid   string
    RefreshUuid  string
    AtExpires    int64
    RtExpires    int64
  }

type CookieAccess struct {
    Writer     http.ResponseWriter
    UserId     uint64
    IsLoggedIn bool
}
// method to write cookie
func (this *CookieAccess) SetToken(token string) {
    http.SetCookie(this.Writer, &http.Cookie{
        Name:     "shotify-cookie",
        Value:    token,
        HttpOnly: true,
        Path:     "/",
        Expires:  time.Now().Add(token_expire),
    })
}


func extractUserId(ctx *gin.Context) (uint64, error) {
    c, err := ctx.Request.Cookie(cookieName)
    if err != nil {
        return 0, errors.New("There is no token in cookies")
    }

    userId, err := ParseToken(c.Value)
    if err != nil {
        return 0, err
    }
    return userId, nil
}

func setValInCtx(ctx *gin.Context, val interface{}) {
    newCtx := context.WithValue(ctx.Request.Context(), cookieAccessKeyCtx, val)
    ctx.Request = ctx.Request.WithContext(newCtx)
}

func Middleware() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        cookieA := CookieAccess{
            Writer: ctx.Writer,
        }

        // &cookieA is a pointer so any changes in future is changing cookieA is context
        setValInCtx(ctx, &cookieA)

        userId, err := extractUserId(ctx)
        if err != nil {
            cookieA.IsLoggedIn = false
            ctx.Next()
            return
        }

        cookieA.UserId = userId
        cookieA.IsLoggedIn = true

       // calling the actual resolver
        ctx.Next()
       // here will execute after resolver and all other middlewares was called
       // so &cookieA is safe from garbage collector
    }
}