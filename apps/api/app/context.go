package app

import (
	"context"
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *Application) ContextSetAuthUser(r *http.Request, user data.AuthUser) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *Application) ContextGetUser(r *http.Request) data.AuthUser {
	user, ok := r.Context().Value(userContextKey).(data.AuthUser)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}

func (app *Application) ContextGetDealershipUser(r *http.Request) *data.DealershipUser {
	user := app.ContextGetUser(r)
	dealershipUser, ok := user.(*data.DealershipUser)
	if !ok {
		panic("user is not a dealership user")
	}
	return dealershipUser
}

func (app *Application) ContextGetInternalUser(r *http.Request) *data.InternalUser {
	user := app.ContextGetUser(r)
	internalUser, ok := user.(*data.InternalUser)
	if !ok {
		panic("user is not an internal user")
	}
	return internalUser
}
