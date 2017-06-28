package utils

import (
	"database/sql"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"fmt"
	"log"
)

// General state variables for the server
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

func NewContext() *Context {
	ctx := Context{}
	ctx.loadConfig()
	ctx.connectAWS()
	return &ctx
}

// Loads the configuration file.
func (ctx *Context) loadConfig() (*Context, error) {
	file, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(file, ctx)
	if err != nil {
		return nil, err
	}

	return ctx, err
}

// Creates a new AWS client session.
func (ctx *Context) connectAWS() *Context {
	ctx.Store = s3.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})))
	return ctx
}

func (ctx *Context) SetupDatabase() {
	var err error
	isDevelopment := os.Getenv("ENVIRONMENT") == "development"
	if isDevelopment {
		ctx.Db, err = sql.Open("mysql",
			"dev:password@tcp(127.0.0.1:3306)/testdb")
		ctx.Db.Exec("create table if not exists entries")
	} else {
		// Setup RDS db from env variables
		RDS_HOSTNAME := os.Getenv("RDS_HOSTNAME")
		RDS_PORT := os.Getenv("RDS_PORT")
		RDS_DB_NAME := os.Getenv("RDS_DB_NAME")
		RDS_USERNAME := os.Getenv("RDS_USERNAME")
		RDS_PASSWORD := os.Getenv("RDS_PASSWORD")
		sourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			RDS_USERNAME, RDS_PASSWORD, RDS_HOSTNAME, RDS_PORT, RDS_DB_NAME)
		log.Println("RDS connection string: " + sourceName)
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
	log.Println("Successfully connected database.")
}