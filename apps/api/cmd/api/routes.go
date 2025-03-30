package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /api/cat", app.handleTest)
	router.HandleFunc("/", app.handleNotFound)

	return app.recoverPanic(app.logRequest(router))
}
