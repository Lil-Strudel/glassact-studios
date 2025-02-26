package auth

import "net/http"

func SetupRoutes(parentMux *http.ServeMux) {
	authMux := http.NewServeMux()

	authMux.HandleFunc("GET /google", GetGoogleAuth)
	authMux.HandleFunc("GET /google/callback", GetGoogleAuthCallback)

	parentMux.Handle("/auth/", http.StripPrefix("/auth", authMux))
}
