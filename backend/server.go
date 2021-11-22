package main

//go:generate sqlboiler --wipe --no-tests -o databases/models -p databases mysql

import (
	"context"
	"database/sql"
	"log"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gasser707/go-gql-server/auth"
	"github.com/gasser707/go-gql-server/cloud"
	"github.com/gasser707/go-gql-server/databases"
	"github.com/gasser707/go-gql-server/graph"
	"github.com/gasser707/go-gql-server/graph/generated"
	"github.com/gasser707/go-gql-server/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/joho/godotenv/autoload"
)

// Defining the Graphql handler
func graphqlHandler(mysqlDB *sql.DB) gin.HandlerFunc {
	// NewExecutableSchema and Config are in the generated.go file
	// Resolver is in the resolver.go file

	ctx := context.Background()
	so, err := cloud.NewStorageOperator(ctx)
	if err != nil {
		log.Panic(err)
	}
	authSrv := services.NewAuthService(mysqlDB)
	usrSrv := services.NewUsersService(mysqlDB, authSrv, so)
	imgSrv := services.NewImagesService(mysqlDB, authSrv, so)
	saleSrv := services.NewSalesService(mysqlDB, authSrv)

	c := generated.Config{Resolvers: &graph.Resolver{AuthService: authSrv,
		ImagesService: imgSrv, UsersService: usrSrv, SaleService: saleSrv}}

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

	mysqlDB := databases.NewMysqlClient()
	driver, _ := mysql.WithInstance(mysqlDB, &mysql.Config{})
    m, err := migrate.NewWithDatabaseInstance(
        "file://databases/migrations",
        "mysql", 
        driver,
    )
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil && err!=migrate.ErrNoChange{
		log.Fatal(err)
	}

	// Setting up Gin
	r := gin.Default()

	r.Use(auth.Middleware())

	r.POST("/query", graphqlHandler(mysqlDB))
	r.GET("/", playgroundHandler())
	r.Run()

}
