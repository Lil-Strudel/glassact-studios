package router

import (
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/module/auth"
	"github.com/Lil-Strudel/glassact-studios/apps/api/module/cat"
)

func SetupRoutes(parentMux *http.ServeMux) {
	apiMux := http.NewServeMux()

	auth.SetupRoutes(apiMux)
	cat.SetupRoutes(apiMux)

	parentMux.Handle("/api/", http.StripPrefix("/api", apiMux))
	parentMux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})
}
