package databases

import (
	"database/sql"
	// "fmt"
	"log"
	"os"
)

const (
	mysqlUsersUsername = "mysql_users_username"
	mysqlUsersPassword = "mysql_users_password"
	mysqlUsersHost     = "mysql_users_host"
	mysqldb   = "testdb"
)

var (
	Client *sql.DB

	username = os.Getenv(mysqlUsersUsername)
	password = os.Getenv(mysqlUsersPassword)
	host     = os.Getenv(mysqlUsersHost)
	db   = os.Getenv(mysqldb)
)

func init() {

	// dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8",
	// 	username, password, host, db,
	// )

	var err error
	dataSourceName := "user7:s$cret@tcp(127.0.0.1:8083)/testdb"

	Client, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}
	if err = Client.Ping(); err != nil {
		panic(err)
	}

	log.Println("database successfully configured")
}
