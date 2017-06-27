package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"log"
	_ "github.com/go-sql-driver/mysql"
	"ncbi-tool-server/controllers"
	"ncbi-tool-server/utils"
	"net/http"
	"os"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// General setup procedure
func main() {
	// Setup
	ctx := utils.NewContext()
	var err error

	isDevelopment := os.Getenv("ENVIRONMENT") == "development"
	if isDevelopment {
		// File is created if absent
		ctx.Db, err = sql.Open("sqlite3",
			"versions.db")
	} else {
		// Setup RDS db from env variables
		RDS_HOSTNAME := os.Getenv("RDS_HOSTNAME")
		RDS_PORT := os.Getenv("RDS_PORT")
		RDS_DB_NAME := os.Getenv("RDS_DB_NAME")
		RDS_USERNAME := os.Getenv("RDS_USERNAME")
		RDS_PASSWORD := os.Getenv("RDS_PASSWORD")
		sourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", RDS_USERNAME, RDS_PASSWORD, RDS_HOSTNAME, RDS_PORT, RDS_DB_NAME)
		ctx.Db, err = sql.Open("mysql", sourceName)
	}

	if err != nil {
		log.Println("Failed to set up database opener.")
		log.Fatal(err)
	}
	defer ctx.Db.Close()
	err = ctx.Db.Ping()
	if err != nil {
		log.Println("Failed to ping database.")
		log.Fatal(err)
	}

	// Routing
	router := mux.NewRouter()
	fileController := controllers.NewFileController(ctx)
	fileController.Register(router)
	directoryController := controllers.NewDirectoryController(ctx)
	directoryController.Register(router)
	router.NotFoundHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "Page not found.")
		})

	// Start server
	log.Println("Starting listener...")
	err = http.ListenAndServe(":8000", router)
	if err != nil {
		log.Println(err.Error())
		log.Fatal("Error in running listen and serve.")
	}
}
