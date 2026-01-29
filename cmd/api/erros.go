package main

import (
	"fmt"
	"net/http"
)

func (app *application) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)
	app.logger.Error(err.Error(), "method", method, "uri", uri)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {

	env := envelope{"error": message}
	//func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error
	err := app.writeJSON(w, status, env, nil)

	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

// This will be used to send a 505 internal server error
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	message := " the server encountered a problem could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// This will be used to send a 404 Not found Status
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {

	message := "the requested resource could not be found"

	app.errorResponse(w, r, http.StatusNotFound, message)
}

// This will be used to send a 405 Method  Not Allowed
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf(" the %s method is not supported for this resource ", r.Method)

	app.errorResponse(w, r, http.StatusNotFound, message)
}
