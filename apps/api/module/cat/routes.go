package cat

import "net/http"

func SetupRoutes(parentMux *http.ServeMux) {
	catMux := http.NewServeMux()

	catMux.HandleFunc("GET /", GetCats)
	catMux.HandleFunc("POST /", PostCat)

	parentMux.Handle("/cat/", http.StripPrefix("/cat", catMux))
}
