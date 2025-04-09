package app

import (
	"net/http"
)

func (app *Application) HandleNotFound(w http.ResponseWriter, r *http.Request) {
	app.WriteError(w, r, app.Err.RouteNotFound, nil)
}
