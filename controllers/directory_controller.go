package controllers

import (
	"github.com/gorilla/mux"
	"ncbi-tool-server/models"
	"ncbi-tool-server/utils"
	"net/http"
	"errors"
)

// DirectoryController is for handling directory actions
type DirectoryController struct {
	ApplicationController
	ctx *utils.Context
}

// NewDirectoryController returns a new controller instance
func NewDirectoryController(ctx *utils.Context) *DirectoryController {
	return &DirectoryController{
		ctx: ctx,
	}
}

// Register registers the directory endpoint with the router
func (dc *DirectoryController) Register(router *mux.Router) {
	router.HandleFunc("/directory/compare", dc.Compare)
	router.HandleFunc("/directory", dc.Show)
}

// Show handles requests for showing directory listing
func (dc *DirectoryController) Show(w http.ResponseWriter,
	r *http.Request) {
	// Setup
	dir := models.NewDirectory(dc.ctx)
	inputTime := r.URL.Query().Get("input-time")
	op := r.URL.Query().Get("op")
	pathName := r.URL.Query().Get("path-name")
	output := r.URL.Query().Get("output")
	var err error
	var result interface{}

	// Dispatch operations
	switch {
	case pathName == "":
		dc.BadRequest(w, errors.New("empty pathName"))
		return
	case op == "at-time":
		// Serve up folder at a given time
		result, err = dir.GetPast(pathName, inputTime, output)
	default:
		// Serve up latest version of the folder
		result, err = dir.GetLatest(pathName, output)
	}

	if err != nil {
		dc.BadRequest(w, err)
		return
	}
	dc.Output(w, result)
}

func (dc *DirectoryController) Compare(w http.ResponseWriter,
	r *http.Request) {
	// Setup
	dir := models.NewDirectory(dc.ctx)
	pathName := r.URL.Query().Get("path-name")
	startDate := r.URL.Query().Get("start-date")
	endDate := r.URL.Query().Get("end-date")
	var err error
	var result interface{}

	result, err = dir.CompareListing(pathName, startDate, endDate)

	if err != nil {
		dc.BadRequest(w, err)
		return
	}
	dc.Output(w, result)
}