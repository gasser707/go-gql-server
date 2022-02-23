package middleware

import (
	"context"
	"net/http"
	customErr "github.com/gasser707/go-gql-server/errors"
	"github.com/gin-gonic/gin"
)

const headerKey = "header-name" 

type HeaderAccess struct {
    Writer     http.ResponseWriter
    CsrfToken   string
}
// method to write headers
func (ha *HeaderAccess) SetCsrfToken(csrfTk string) {
	ha.Writer.Header().Set("X-CSRF-Token", csrfTk)    
}

func GetHeaderAccess(ctx context.Context) (*HeaderAccess, error) {
    ha, ok :=  ctx.Value(headerKey).(*HeaderAccess)
    if(!ok){
        return nil, customErr.NoAuth("csrf token not found")
    }
    return ha, nil
}

func HeaderMiddleware() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        token:= ctx.Request.Header.Get("X-CSRF-Token")
        headerAccess:= HeaderAccess{
            Writer: ctx.Writer,
            CsrfToken: token,
        }

        // &headerAccess is a pointer so any changes in future is changing cookieA is context
        setValInCtx(ctx, headerKey ,&headerAccess)

       // calling the actual resolver
        ctx.Next()
    }
}