package cat

import "net/http"

func SetupRoutes(parentMux *http.ServeMux) {
	parentMux.HandleFunc("GET /cat", GetCats)
	parentMux.HandleFunc("POST /cat", PostCat)
}
