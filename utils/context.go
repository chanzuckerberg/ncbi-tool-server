package utils

import (
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"log"
	"os"
)

// Context contains general state variables for the server
type Context struct {
	Db         *sql.DB
	Bucket     string
	Store      s3iface.S3API
	Port       string
}

// NewContext initializes new general state variables
func NewContext() *Context {
	ctx := Context{}
	return &ctx
}

// SetupDatabase sets up the db and checks connection conditions
func (ctx *Context) SetupDatabase() {
	var err error
	isDevelopment := os.Getenv("ENVIRONMENT") == "development"
	if isDevelopment {
		ctx.Db, err = sql.Open("mysql",
			"dev:password@tcp(127.0.0.1:3306)/testdb")
		if err != nil {
			log.Fatal("Failed to set up database opener: " + err.Error())
		}
		_, err = ctx.Db.Exec("create table if not exists entries")
		if err != nil {
			log.Fatal("Failed to create table entries: " + err.Error())
		}
	} else {
		// Setup RDS db from env variables
		rdsHostname := os.Getenv("RDS_HOSTNAME")
		rdsPort := os.Getenv("RDS_PORT")
		rdsDbName := os.Getenv("RDS_DB_NAME")
		rdsUsername := os.Getenv("RDS_USERNAME")
		rdsPassword := os.Getenv("RDS_PASSWORD")
		sourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			rdsUsername, rdsPassword, rdsHostname, rdsPort, rdsDbName)
		log.Print("DB connection string: " + sourceName)
		ctx.Db, err = sql.Open("mysql", sourceName)
	}
	if err != nil {
		log.Print(err)
		log.Fatal("Failed to set up database opener.")
	}
	err = ctx.Db.Ping()
	if err != nil {
		log.Print(err)
		log.Fatal("Failed to ping database.")
	}
	rows, err := ctx.Db.Query("show tables like 'entries'")
	if err != nil || !rows.Next() {
		log.Fatal("Table 'entries' does not exist.")
	}
	log.Print("Successfully connected database.")
}
