package handlers

import (
	"bytes"
	"io/fs"
	"net/http"
	"time"
	"vpod/internal/views"
)

func Index() http.HandlerFunc {
	indexHTML, err := views.ViewFS.ReadFile("index.html")
	if err != nil { // should never happen if the embed worked
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "index.html", time.Now(), bytes.NewReader(indexHTML))
	}
}

func Static() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		fs, err := fs.Sub(views.ViewFS, "static")
		if err != nil { // should never happen if the embed worked
			panic(err)
		}

		http.FileServer(http.FS(fs)).ServeHTTP(w, r)
	}
	return http.StripPrefix("/ui/static/", http.HandlerFunc(fn))
}
