package main

import (
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	
	app.logger.Errorw("interal server error", "method", r.Method, "path", r.URL.Path, "error", err.Error() )

	writeJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *application) forbiddenResponse(w http.ResponseWriter, r *http.Request, err error) {
	
	app.logger.Warnw("forbidden", "method", r.Method, "path", r.URL.Path, "error")

	writeJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	
	app.logger.Warnf("Bad request error", "method", r.Method, "path", r.URL.Path, "error", err.Error() )

	writeJSONError(w, http.StatusBadRequest, "the server encountered a problem")
}

func (app *application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	
	app.logger.Errorw("Confict respose error", "method", r.Method, "path", r.URL.Path, "error", err.Error() )

	writeJSONError(w, http.StatusConflict, "the server encountered a problem")
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	
	app.logger.Warnf("not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error() )

	writeJSONError(w, http.StatusNotFound, "the server encountered a problem")
}

func (app *application) unauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	
	app.logger.Warnf("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error() )

	writeJSONError(w, http.StatusUnauthorized, "the server encountered a problem")
}

func (app *application) unauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	
	app.logger.Warnf("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error() )

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	writeJSONError(w, http.StatusUnauthorized, "the server encountered a problem")
}

