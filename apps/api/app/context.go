package app

import (
	"context"
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *Application) ContextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *Application) ContextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
