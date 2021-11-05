package main

import (
	// "context"
	// "net/http"
	// "github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gasser707/go-gql-server/graph"
	"github.com/gasser707/go-gql-server/graph/generated"
	"github.com/gasser707/go-gql-server/auth"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"

)


// Defining the Graphql handler
func graphqlHandler() gin.HandlerFunc {
	// NewExecutableSchema and Config are in the generated.go file
	// Resolver is in the resolver.go file

	c :=  generated.Config{Resolvers: &graph.Resolver{}} 
	// c.Directives.Authorize = Authorize
		
	h := handler.NewDefaultServer(generated.NewExecutableSchema(c))

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}


// Defining the Playground handler
func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/query")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}


func main() {
	//redis details



		// Setting up Gin
		r := gin.Default()

		r.Use(auth.Middleware())

		// r.Use(WithCookie())

		r.POST("/query", graphqlHandler())
		r.GET("/", playgroundHandler())
		r.Run()

}

// func Authorize (ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
// 	c := ctx.Value(ctx, auth_middleware.CookieKey)
// 	if c == nil {
// 		return nil, er.ErrInvalidToken.Log(ctx)
// 	}
// 		cookie := c.(*http.Cookie)

// 	jwtToken := ctx.Value(ctx, auth.TokenKey).([]byte)
		 
// 	   //... decrypt, validate, set claims etc
	  
// 	   next(ctx)
// }
