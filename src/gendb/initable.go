package gendb

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func initVars() (string, string, string, string, string) {
	dataHost, exist := os.LookupEnv("MYSQL_HOST")

	var sdataHost string = ""
	if exist {
		sdataHost = dataHost
	}

	dataPort, exist := os.LookupEnv("MYSQL_PORT")
	var sdataPort string = ""
	if exist {
		sdataPort = dataPort
	}

	dataUser, exist := os.LookupEnv("MYSQL_USER")
	var sdataUser string = ""
	if exist {
		sdataUser = dataUser
	}

	dataPass, exist := os.LookupEnv("MYSQL_PASS")
	var sdataPass string = ""
	if exist {
		sdataPass = dataPass
	}
	dataName, exist := os.LookupEnv("MYSQL_DBNAME")
	var sdataName string = ""
	if exist {
		sdataName = dataName
	}

	return sdataHost, sdataPort, sdataUser, sdataPass, sdataName
}

func InitDb() (dbs *sql.DB, er error) {
	dbHost, dbPort, dbUser, dbPassword, dbName := initVars()
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName))

	if err != nil {
		slog.Error(err.Error())
		return nil, fmt.Errorf("unable to connect to the database")
	}

	db.Prepare("")
	return db, nil
}
