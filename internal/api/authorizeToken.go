package api

import (
	"net/http"
	"strings"
)

func TokenMiddleware(user string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			tokenRaw := r.Header.Get("Authorization")
			token := strings.TrimPrefix(tokenRaw, "Bearer ")
			if token != desiredToken(user) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func desiredToken(user string) string {
	return "TODO"
}
