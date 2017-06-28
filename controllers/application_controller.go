package controllers

import (
	"encoding/json"
	"ncbi-tool-server/utils"
	"net/http"
)

// ApplicationController for general application functions
type ApplicationController struct {
	ctx *utils.Context
}

// NewApplicationController creates a new app controller instance
func NewApplicationController(
	ctx *utils.Context) *ApplicationController {
	return &ApplicationController{ctx: ctx}
}

// InternalError sends an error for server errors to the client
func (ac *ApplicationController) InternalError(w http.ResponseWriter,
	err error) {
	http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
}

// BadRequest sends an error for badly formed requests to the client
func (ac *ApplicationController) BadRequest(w http.ResponseWriter) {
	http.Error(w, "Invalid request.", http.StatusBadRequest)
}

// Output marshals a struct into JSON format for output
func (ac *ApplicationController) Output(w http.ResponseWriter,
	result interface{}) {
	js, err := json.Marshal(result)
	if err != nil {
		ac.InternalError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(js)
}
