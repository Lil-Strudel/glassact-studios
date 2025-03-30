package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /api/auth/google", app.handleGetGoogleAuth)
	router.HandleFunc("GET /api/auth/google/callback", app.handleGetGoogleAuthCallback)

	router.HandleFunc("/", app.handleNotFound)

	return app.recoverPanic(app.logRequest(router))
}
