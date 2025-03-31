package main

import (
	"net/http"
)

type errorType string

const (
	routeNotFound       = errorType("route-not-found")
	recordNotFound      = errorType("record-not-found")
	serverError         = errorType("server-error")
	authenticationError = errorType("authentication-error")
)

type errorConfig struct {
	status   int
	message  string
	expected bool
}

var typeConfig = map[errorType]errorConfig{
	routeNotFound: {
		status:   http.StatusNotFound,
		message:  "The route you have requested was not found or does not exist.",
		expected: true,
	},
	recordNotFound: {
		status:   http.StatusNotFound,
		message:  "The record you have requested was not found or does not exist.",
		expected: true,
	},
	serverError: {
		status:   http.StatusInternalServerError,
		message:  "The server encountered an unknown issue while processing your request.",
		expected: false,
	},
	authenticationError: {
		status:   http.StatusUnauthorized,
		message:  `Invalid authentication credentials. Please provide an "Authorization" header in "Bearer $access_token" format.`,
		expected: true,
	},
}

func (app *application) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	app.log.Error(err.Error(), "method", method, "uri", uri)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, errorType errorType, err error) {
	if err != nil {
		app.logError(r, err)
	}

	errorConfig := typeConfig[errorType]

	resBody := map[string]any{
		"type":    errorType,
		"message": errorConfig.message,
	}

	app.writeJSON(w, r, errorConfig.status, resBody)
}
