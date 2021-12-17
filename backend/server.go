package main

import (
	"context"
	"log"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gasser707/go-gql-server/auth"
	"github.com/gasser707/go-gql-server/cloud"
	"github.com/gasser707/go-gql-server/databases"
	"github.com/gasser707/go-gql-server/graph"
	"github.com/gasser707/go-gql-server/graph/generated"
	"github.com/gasser707/go-gql-server/helpers"
	"github.com/gasser707/go-gql-server/services"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

// Defining the Graphql handler
func graphqlHandler() gin.HandlerFunc {
	// NewExecutableSchema and Config are in the generated.go file
	// Resolver is in the resolver.go file

	ctx := context.Background()
	so, err := cloud.NewStorageOperator(ctx)
	if err != nil {
		log.Panic(err)
	}
	mysqlDB := databases.NewMysqlClient()
	authSrv := services.NewAuthService(mysqlDB)
	userSrv := services.NewUsersService(mysqlDB, so)
	imgSrv := services.NewImagesService(ctx, mysqlDB, so)
	saleSrv := services.NewSalesService(mysqlDB, authSrv)

	c := generated.Config{Resolvers: &graph.Resolver{AuthService: authSrv,
		ImagesService: imgSrv, UsersService: userSrv, SaleService: saleSrv}}

	c.Directives.IsLoggedIn = func(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
		userId, _, err := authSrv.ValidateCredentials(ctx)
		if err != nil {
			return nil, err
		}
		newCtx := context.WithValue(ctx, helpers.UserIdKey, userId)
		return next(newCtx)
	}

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
