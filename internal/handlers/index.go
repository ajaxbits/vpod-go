package handlers

import (
	"fmt"
	"net/http"
)

func Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "internal/views/index.html")
	}
}

func Static() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL)
		staticDir := "./internal/views/static"
		http.FileServer(http.Dir(staticDir)).ServeHTTP(w, r)
	}
}
