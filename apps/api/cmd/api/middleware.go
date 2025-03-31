package main

import (
	"net/http"
	"strings"

	"github.com/Lil-Strudel/glassact-studios/libs/database"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.errorResponse(w, r, serverError, nil)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip     = r.RemoteAddr
			proto  = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
		)

		app.log.Info("received request", "ip", ip, "proto", proto, "method", method, "uri", uri)

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			app.errorResponse(w, r, authenticationError, nil)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.errorResponse(w, r, authenticationError, nil)
			return
		}

		token := headerParts[1]

		err := app.validate.Var(token, "required,len=26")
		if err != nil {
			app.errorResponse(w, r, authenticationError, nil)
			return
		}

		user, found, err := app.db.Users.GetForToken(database.ScopeAccess, token)
		if err != nil {
			app.errorResponse(w, r, serverError, err)
			return
		}

		if !found {
			app.errorResponse(w, r, authenticationError, nil)
			return
		}

		r = app.contextSetUser(r, user)

		next.ServeHTTP(w, r)
	})
}
