package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"ncbi-tool-server/controllers"
	"ncbi-tool-server/utils"
	"net/http"
	"os"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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
	ctx.Bucket = os.Getenv("BUCKET")
	ctx.Port = "80"
	if os.Getenv("PORT") != "" {
		ctx.Port = os.Getenv("PORT")
	}
	ctx.Store = s3.New(session.Must(session.NewSession()))
	var err error
	ctx.SetupDatabase()
	defer ctx.Db.Close()

	// Routing
	router := mux.NewRouter()
	fileController := controllers.NewFileController(ctx)
	fileController.Register(router)
	directoryController := controllers.NewDirectoryController(ctx)
	directoryController.Register(router)
	router.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "Welcome to the NCBI data tool.")
		})
	router.NotFoundHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "Page not found.")
		})

	// Start server
	log.Print("Starting listener...")
	err = http.ListenAndServe(":" + ctx.Port, router)
	if err != nil {
		log.Print(err.Error())
		log.Fatal("Error in running listen and serve.")
	}
}
