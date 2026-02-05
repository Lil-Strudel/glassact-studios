package app

import (
	"net/http"
)

func (app *Application) RequirePermission(action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := app.ContextGetUser(r)

			if !user.Can(action) {
				app.WriteError(w, r, app.Err.Forbidden, nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (app *Application) RequireRole(roles ...string) func(http.Handler) http.Handler {
	roleMap := make(map[string]bool)
	for _, role := range roles {
		roleMap[role] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := app.ContextGetUser(r)

			if !roleMap[user.GetRole()] {
				app.WriteError(w, r, app.Err.Forbidden, nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
