package controllers

import (
	"encoding/json"
	"log"
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

// ErrorResponse contains error information for the client
type ErrorResponse struct {
	Code  int
	Error string
}

// InternalError sends an error for server errors to the client
func (ac *ApplicationController) InternalError(w http.ResponseWriter,
	err error) {
	res := ErrorResponse{http.StatusInternalServerError,
		"Error: " + err.Error()}
	ac.ErrorOutput(w, res)
}

// BadRequest sends an error for badly formed requests to the client
func (ac *ApplicationController) BadRequest(w http.ResponseWriter,
	err error) {
	res := ErrorResponse{http.StatusBadRequest,
		"Request error: " + err.Error()}
	ac.ErrorOutput(w, res)
}

// ErrorOutput writes a error response to the client
func (ac *ApplicationController) ErrorOutput(w http.ResponseWriter,
	result ErrorResponse) {
	js, err := json.Marshal(result)
	if err != nil {
		log.Print("Error with JSON marshal: " + err.Error())
		http.Error(w, "Error: "+err.Error(),
			http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(result.Code)
	_, err = w.Write(js)
	if err != nil {
		log.Print("Error writing JSON output: " + err.Error())
	}
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
	_, err = w.Write(js)
	if err != nil {
		log.Print("Error writing JSON output: " + err.Error())
	}
}
