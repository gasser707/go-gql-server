package databases

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
)

const (
	mysqlUsersUsername = "MYSQL_USER"
	mysqlUsersPassword = "MYSQL_PASS"
	mysqlUsersHost     = "MYSQL_HOST"
	mysqlUsersSchema   = "MYSQL_DBNAME"
)

var (

	username = os.Getenv(mysqlUsersUsername)
	password = os.Getenv(mysqlUsersPassword)
	host     = os.Getenv(mysqlUsersHost)
	schema   = os.Getenv(mysqlUsersSchema)
)

func NewMysqlClient() *sql.DB {

	//username:password@protocol(address)/dbname?param=value
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
		username, password, host, schema,
	)

	mysqlClient, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}
	if err = mysqlClient.Ping(); err != nil {
		panic(err)
	}

	log.Println("database successfully configured")

	return mysqlClient

}
