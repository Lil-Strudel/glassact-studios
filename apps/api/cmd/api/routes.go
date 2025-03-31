package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	unprotected := alice.New()
	mux.Handle("GET /api/auth/google", unprotected.ThenFunc(app.handleGetGoogleAuth))
	mux.Handle("GET /api/auth/google/callback", unprotected.ThenFunc(app.handleGetGoogleAuthCallback))
	mux.Handle("POST /api/auth/token/access", unprotected.ThenFunc(app.handlePostTokenAccess))
	mux.Handle("GET /api/auth/logout", unprotected.ThenFunc(app.handleGetLogout))
	mux.Handle("/", unprotected.ThenFunc(app.handleNotFound))

	protected := alice.New(app.authenticate)
	mux.Handle("GET /api/user/self", protected.ThenFunc(app.handleGetUserSelf))

	standard := alice.New(app.recoverPanic, app.logRequest)
	return standard.Then(mux)
}
