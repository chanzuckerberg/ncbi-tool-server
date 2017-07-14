package controllers

import (
	"errors"
	"github.com/gorilla/mux"
	"ncbi-tool-server/models"
	"ncbi-tool-server/utils"
	"net/http"
)

// FileController is for handling file actions
type FileController struct {
	ApplicationController
	ctx *utils.Context
}

// NewFileController returns a new controller instance
func NewFileController(ctx *utils.Context) *FileController {
	return &FileController{
		ctx: ctx,
	}
}

// Register registers the file endpoint with the router
func (fc *FileController) Register(router *mux.Router) {
	router.HandleFunc("/file", fc.Show)
	router.HandleFunc("/file/history", fc.History)
	router.HandleFunc("/file/at-time", fc.AtTime)
}

// Show handles requests for showing file information
func (fc *FileController) Show(w http.ResponseWriter,
	r *http.Request) {
	// Setup
	file := models.NewFile(fc.ctx)
	pathName := r.URL.Query().Get("path-name")
	var err error
	var result models.Entry
	versionNum := r.URL.Query().Get("version-num")

	// Dispatch operations
	switch {
	case pathName == "":
		fc.BadRequest(w, errors.New("empty pathName"))
		return
	case versionNum != "":
		// Serve up file version
		result, err = file.GetVersion(pathName, versionNum)
	default:
		// Serve up the file, latest version
		result, err = file.GetVersion(pathName, "0")
	}
	fc.DefaultResponse(w, result, err)
}

// History handles requests for showing file version history.
func (fc *FileController) History(w http.ResponseWriter,
	r *http.Request) {
	file := models.NewFile(fc.ctx)
	pathName := r.URL.Query().Get("path-name")
	result, err := file.GetHistory(pathName)
	fc.DefaultResponse(w, result, err)
}

// AtTime handles requests for getting a file version at a point in time.
func (fc *FileController) AtTime(w http.ResponseWriter,
	r *http.Request) {
	file := models.NewFile(fc.ctx)
	pathName := r.URL.Query().Get("path-name")
	inputTime := r.URL.Query().Get("input-time")
	result, err := file.GetAtTime(pathName, inputTime)
	fc.DefaultResponse(w, result, err)
}