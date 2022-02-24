package main

import (
	"context"
	"log"
	"os"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gasser707/go-gql-server/databases"
	"github.com/gasser707/go-gql-server/graphql/dataloaders"
	"github.com/gasser707/go-gql-server/graphql/generated"
	"github.com/gasser707/go-gql-server/graphql/resolvers"
	"github.com/gasser707/go-gql-server/helpers"
	"github.com/gasser707/go-gql-server/middleware"
	"github.com/gasser707/go-gql-server/services"
	email_svc "github.com/gasser707/go-gql-server/services/email"
	sales_svc "github.com/gasser707/go-gql-server/services/sale"
	"github.com/gasser707/go-gql-server/utils/cloud"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
)

var playgroundUrl = os.Getenv("DOMAIN_NAME")

// Defining the Graphql handler
func graphqlHandler(mysqlDB *sqlx.DB, dl dataloaders.RetrieverInterface) gin.HandlerFunc {
	// NewExecutableSchema and Config are in the generated.go file
	// Resolver is in the resolver.go file

	ctx := context.Background()
	gcsClient, err := cloud.NewGcsClient()
	if err != nil {
		log.Panic(err)
	}

	so := cloud.NewStorageOperator(gcsClient)
	if err != nil {
		log.Panic(err)
	}
	emailSrv := email_svc.NewEmailService()
	emailAdaptor := email_svc.NewEmailAdaptor(emailSrv)

	authSrv := services.NewAuthService(mysqlDB, emailAdaptor)
	userSrv := services.NewUsersService(mysqlDB, so, emailAdaptor)
	imgSrv := services.NewImagesService(ctx, mysqlDB, so, emailAdaptor)
	saleSrv := sales_svc.NewSalesService(mysqlDB)

	c := generated.Config{Resolvers: &resolvers.Resolver{AuthService: authSrv,
		ImagesService: imgSrv, UsersService: userSrv, SaleService: saleSrv, EmailService: emailSrv,
		DataLoaders: dl,
	}}

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

	mysqlDB := databases.NewMysqlClient()
	dl := dataloaders.NewRetriever()                         // <- here we initialize the dataloader.Retriever
	dlMiddleware := middleware.DataLoaderMiddleware(mysqlDB) // <- here we initialize the middleware

	r.Use(middleware.CookieMiddleware())
	r.Use(middleware.HeaderMiddleware())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
		AllowMethods:     []string{"PUT", "PATCH", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Cookie", "Set-Cookie", "X-CSRF-TOKEN"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "Cookie", "Set-Cookie", "X-CSRF-TOKEN"},
	}))

	r.Use(dlMiddleware)

	r.POST("/query", graphqlHandler(mysqlDB, dl))
	r.GET("/query/playground", playgroundHandler())
	r.Run()

}
