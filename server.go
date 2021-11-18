package main

//go:generate sqlboiler --wipe --no-tests -o databases/models -p databases mysql

import (
	"context"
	"log"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gasser707/go-gql-server/auth"
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
	co,err:= helpers.NewCloudOperator(ctx)
	if(err!=nil){
		log.Panic(err)
	}

	sc := helpers.NewSecureCookie()
	tk:= auth.NewTokenService(sc)
	rd:= auth.NewRedisStore()
	mysqlDB:= databases.NewMysqlClient()
	authSrv:= services.NewAuthService(rd, tk, mysqlDB, sc )
	usrSrv := services.NewUsersService(mysqlDB, authSrv, co)
	imgSrv := services.NewImagesService(mysqlDB, authSrv, co)
	
	c :=  generated.Config{Resolvers: &graph.Resolver{AuthService: authSrv, ImagesService: imgSrv, UsersService: usrSrv}} 
			
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
