package user

import (
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
)

type application struct {
	*app.Application
}

func NewUserModule(app *app.Application) *application {
	return &application{
		app,
	}
}

func (app *application) HandleGetUserSelf(w http.ResponseWriter, r *http.Request) {
	user := app.ContextGetUser(r)

	app.WriteJSON(w, r, http.StatusOK, user)
}
