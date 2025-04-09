package app

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Lil-Strudel/glassact-studios/libs/database"
)

func (app *Application) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if excp := recover(); excp != nil {
				w.Header().Set("Connection", "close")

				var err error
				switch v := excp.(type) {
				case string:
					err = errors.New(v)
				case error:
					err = v
				default:
					err = errors.New(fmt.Sprint(v))
				}

				app.WriteError(w, r, app.Err.ServerError, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *Application) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip     = r.RemoteAddr
			proto  = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
		)

		app.Log.Info("received request", "ip", ip, "proto", proto, "method", method, "uri", uri)

		next.ServeHTTP(w, r)
	})
}

func (app *Application) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			app.WriteError(w, r, app.Err.AuthenticationError, nil)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.WriteError(w, r, app.Err.AuthenticationError, nil)
			return
		}

		token := headerParts[1]

		err := app.Validate.Var(token, "required,len=26")
		if err != nil {
			app.WriteError(w, r, app.Err.AuthenticationError, nil)
			return
		}

		user, found, err := app.Db.Users.GetForToken(database.ScopeAccess, token)
		if err != nil {
			app.WriteError(w, r, app.Err.ServerError, err)
			return
		}

		if !found {
			app.WriteError(w, r, app.Err.AuthenticationError, nil)
			return
		}

		r = app.ContextSetUser(r, user)

		next.ServeHTTP(w, r)
	})
}
