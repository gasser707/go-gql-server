package middleware

import (
	"context"

	"github.com/gasser707/go-gql-server/graphql/dataloaders"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Middleware stores Loaders as a request-scoped context value.
func DataLoaderMiddleware(db *sqlx.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		loaders := dataloaders.NewLoaders(ctx.Request.Context(), db)
		augmentedCtx := context.WithValue(ctx.Request.Context(), dataloaders.Key, loaders)
		ctx.Request = ctx.Request.WithContext(augmentedCtx)
		ctx.Next()
	}
}
