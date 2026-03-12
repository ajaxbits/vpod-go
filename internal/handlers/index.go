package handlers

import (
	"bytes"
	"html/template"
	"io/fs"
	"net/http"
	"time"
	"vpod/internal/views"
)

type indexData struct {
	Version string
}

// Index returns a handler that serves the main page with the app version
// rendered into the footer.
func Index(version string) http.HandlerFunc {
	tmpl := template.Must(template.ParseFS(views.ViewFS, "index.html"))

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, indexData{Version: version})
	if err != nil {
		panic(err)
	}
	rendered := buf.Bytes()

	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "index.html", time.Now(), bytes.NewReader(rendered))
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
