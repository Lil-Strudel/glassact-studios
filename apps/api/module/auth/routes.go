package auth

import "net/http"

func SetupRoutes(parentMux *http.ServeMux) {
	parentMux.HandleFunc("GET /auth/google", GetGoogleAuth)
	parentMux.HandleFunc("GET /auth/google/callback", GetGoogleAuthCallback)
}
