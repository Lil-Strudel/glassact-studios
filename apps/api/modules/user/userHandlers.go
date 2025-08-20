package user

import (
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
)

type userModule struct {
	*app.Application
}

func NewUserModule(app *app.Application) *userModule {
	return &userModule{
		app,
	}
}

func (userModule *userModule) HandleGetUserSelf(w http.ResponseWriter, r *http.Request) {
	user := userModule.ContextGetUser(r)

	userModule.WriteJSON(w, r, http.StatusOK, user)
}
