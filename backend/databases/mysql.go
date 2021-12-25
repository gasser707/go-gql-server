package databases

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
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

func NewMysqlClient() *sqlx.DB {

	//username:password@protocol(address)/dbname?param=value
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&multiStatements=true",
		username, password, host, schema,
	)

	var mysqlClient *sqlx.DB
	var err error
	connected := false

	log.Println("trying to connect to db")
	for i:=0; i<1; i++{
		mysqlClient, err = sqlx.Connect("mysql", dataSourceName)
		if err == nil {
			connected = true
			break
		} else {
			log.Println("failed will try again in 30 secs!")
			time.Sleep(30*time.Second)
		}
	}

	if (!connected){
		log.Println(err)
		log.Println("Couldn't connect to db will exit")
		os.Exit(1)
	}

	log.Println("database successfully configured")

	return mysqlClient

}
