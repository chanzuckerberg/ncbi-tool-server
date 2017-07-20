package controllers

import (
	"github.com/gorilla/mux"
	"ncbi-tool-server/models"
	"ncbi-tool-server/utils"
	"net/http"
	"time"
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
	router.HandleFunc("/directory", dc.Show)
	router.HandleFunc("/directory/compare", dc.Compare)
	router.HandleFunc("/directory/at-time", dc.AtTime)
}

// Show handles requests for showing a directory listing
func (dc *DirectoryController) Show(w http.ResponseWriter,
	r *http.Request) {
	dir := models.NewDirectory(dc.ctx)
	pathName := utils.GetDirPath(r)
	output := r.URL.Query().Get("output")
	now := time.Now().Format("2006-01-02T03:04:05")
	result, err := dir.GetPast(pathName, now, output)
	dc.DefaultResponse(w, result, err)
}

// Compare handles requests for comparing directory states at different times
func (dc *DirectoryController) Compare(w http.ResponseWriter,
	r *http.Request) {
	dir := models.NewDirectory(dc.ctx)
	pathName := utils.GetDirPath(r)
	startDate := r.URL.Query().Get("start-date")
	endDate := r.URL.Query().Get("end-date")
	result, err := dir.CompareListing(pathName, startDate, endDate)
	dc.DefaultResponse(w, result, err)
}

// AtTime handles requests for a directory listing at a given time
func (dc *DirectoryController) AtTime(w http.ResponseWriter,
	r *http.Request) {
	dir := models.NewDirectory(dc.ctx)
	pathName := utils.GetDirPath(r)
	inputTime := r.URL.Query().Get("input-time")
	output := r.URL.Query().Get("output")
	result, err := dir.GetPast(pathName, inputTime, output)
	dc.DefaultResponse(w, result, err)
}
