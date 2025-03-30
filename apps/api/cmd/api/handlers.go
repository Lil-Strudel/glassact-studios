package main

import "net/http"

func (app *application) handleNotFound(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"message": "route not found",
	}

	app.writeJSON(w, http.StatusNotFound, data)
}

func (app *application) handleTest(w http.ResponseWriter, r *http.Request) {
	type cat struct {
		id   int
		name string
	}

	hi := make([]cat, 1, 1)

	hi[0].id = 1
	hi[0].name = "hello"

	app.writeJSON(w, http.StatusOK, hi)
}
