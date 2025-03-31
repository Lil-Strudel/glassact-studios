package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) writeJSON(w http.ResponseWriter, r *http.Request, status int, data any) {
	js, err := json.Marshal(data)
	if err != nil {
		app.errorResponse(w, r, serverError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
}
