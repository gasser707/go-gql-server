package main

//go:generate sqlboiler --wipe --no-tests -o databases/models -p databases mysql

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gasser707/go-gql-server/auth"
	"github.com/gasser707/go-gql-server/graph"
	"github.com/gasser707/go-gql-server/graph/generated"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

// Defining the Graphql handler
func graphqlHandler() gin.HandlerFunc {
	// NewExecutableSchema and Config are in the generated.go file
	// Resolver is in the resolver.go file

	c :=  generated.Config{Resolvers: &graph.Resolver{}} 
			
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

		// Setting up Gin
		r := gin.Default()

		r.Use(auth.Middleware())

		r.POST("/query", graphqlHandler())
		r.GET("/", playgroundHandler())
		r.Run()

}
