package utils

import (
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

// Context contains general state variables for the server
type Context struct {
	Db         *sql.DB
	Server     string `yaml:"Server"`
	Port       string `yaml:"Port"`
	Username   string `yaml:"Username"`
	Password   string `yaml:"Password"`
	SourcePath string `yaml:"SourcePath"`
	LocalPath  string `yaml:"LocalPath"`
	LocalTop   string `yaml:"LocalTop"`
	Bucket     string `yaml:"Bucket"`
	Store      s3iface.S3API
}

// NewContext initializes new general state variables
func NewContext() *Context {
	ctx := Context{}
	ctx.loadConfig()
	ctx.connectAWS()
	return &ctx
}

// Loads the configuration file.
func (ctx *Context) loadConfig() *Context {
	file, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatal("Error loading config: " + err.Error())
		return nil
	}

	err = yaml.Unmarshal(file, ctx)
	if err != nil {
		log.Fatal("Error loading config: " + err.Error())
		return nil
	}

	return ctx
}

// Creates a new AWS client session.
func (ctx *Context) connectAWS() *Context {
	ctx.Store = s3.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})))
	return ctx
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
		log.Println("DB connection string: " + sourceName)
		ctx.Db, err = sql.Open("mysql", sourceName)
	}
	if err != nil {
		log.Println(err)
		log.Fatal("Failed to set up database opener.")
	}
	err = ctx.Db.Ping()
	if err != nil {
		log.Println(err)
		log.Fatal("Failed to ping database.")
	}
	rows, err := ctx.Db.Query("show tables like 'entries'")
	if err != nil || !rows.Next() {
		log.Fatal("Table 'entiries' does not exist.")
	}
	log.Println("Successfully connected database.")
}
