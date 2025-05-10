package handlers

import (
	"fmt"
	"net/http"
)

func Index() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "internal/views/index.html")
	}
	return http.HandlerFunc(fn)
}

func Static() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL)
		staticDir := "./internal/views/static"
		http.FileServer(http.Dir(staticDir)).ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
