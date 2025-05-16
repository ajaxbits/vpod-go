package handlers

import (
	"net/http"
)

func Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "internal/views/index.html")
	}
}

func Static() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// TODO embed
		staticDir := "./internal/views/static/"
		fs := http.FileServer(http.Dir(staticDir))
		fs.ServeHTTP(w, r)
	}
	return http.StripPrefix("/ui/static/", http.HandlerFunc(fn))
}
