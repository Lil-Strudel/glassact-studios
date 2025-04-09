package app

import (
	"encoding/json"
	"net/http"
)

func (app *Application) WriteJSON(w http.ResponseWriter, r *http.Request, status int, data any) {
	js, err := json.Marshal(data)
	if err != nil {
		app.WriteError(w, r, app.Err.ServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
}

func (app *Application) WriteError(w http.ResponseWriter, r *http.Request, errorType ErrorType, err error) {
	if err != nil {
		app.Log.Error(err.Error(), "method", r.Method, "uri", r.URL.RequestURI())
	}

	errorConfig := ErrorMap[errorType]

	resBody := map[string]any{
		"error-type": errorType,
		"message":    errorConfig.Message,
	}

	app.WriteJSON(w, r, errorConfig.Status, resBody)
}
