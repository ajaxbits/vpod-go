package handlers

import (
	"net/http"
)

func IndexHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "internal/views/index.html")
	}
	return http.HandlerFunc(fn)
}

func StaticHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		staticDir := "./internal/views/static"
		http.FileServer(http.Dir(staticDir)).ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
