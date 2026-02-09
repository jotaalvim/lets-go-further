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

// Generic function to send error messeges to the client
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {

	env := envelope{"error": message}
	//func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error
	err := app.writeJSON(w, status, env, nil)

	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	// we are using an WWW-Authenticate header to remind the user to authenticate a bearer token
	w.Header().Set("WWW-Authenticate", "Bearer")
	message := "invalid or missing authentication token"
	app.errorResponse(w, r, http.StatusBadRequest, message)
}

// This will be used to send a 400 BAD request
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// This will be used to send a 401 Invalid Credentials
func (app *application) invalidCredentialResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// This will be used to send a 401 must be authenticated
func (app *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you mus be authenticated to acess this resource"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// This will be used to send a 403  Forbiden
func (app *application) incativeAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "you r account must be activated"
	app.errorResponse(w, r, http.StatusForbidden, message)
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

// This will be used to send a 409 Conflict
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record dut to an edit conflit, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}

// This will be used to send a 422 Unprocessable entity
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// This will be used to send a 429 Too Many Requests
func (app *application) rateLimitExceedResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}

// This will be used to send a 505 internal server error
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	message := " the server encountered a problem could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}
