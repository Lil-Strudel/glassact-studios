package modules

import (
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/auth"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/user"
	"github.com/justinas/alice"
)

func GetRoutes(app *app.Application) http.Handler {
	mux := http.NewServeMux()

	unprotected := alice.New()
	protected := alice.New(app.Authenticate)
	standard := alice.New(app.RecoverPanic, app.LogRequest)

	authModule := auth.NewAuthModule(app)
	mux.Handle("GET /api/auth/google", unprotected.ThenFunc(authModule.HandleGetGoogleAuth))
	mux.Handle("GET /api/auth/google/callback", unprotected.ThenFunc(authModule.HandleGetGoogleAuthCallback))
	mux.Handle("POST /api/auth/token/access", unprotected.ThenFunc(authModule.HandlePostTokenAccess))
	mux.Handle("GET /api/auth/logout", unprotected.ThenFunc(authModule.HandleGetLogout))

	userModule := user.NewUserModule(app)
	mux.Handle("GET /api/user/self", protected.ThenFunc(userModule.HandleGetUserSelf))

	mux.Handle("/", unprotected.ThenFunc(app.HandleNotFound))
	return standard.Then(mux)
}
