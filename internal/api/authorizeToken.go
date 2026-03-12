package api

import (
	"net/http"
	"strings"
)

// TokenMiddleware validates Bearer tokens in the Authorization header.
// WARNING: desiredToken is not yet implemented. Do not wire this
// middleware into routes until a real token source is provided.
func TokenMiddleware(user string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			tokenRaw := r.Header.Get("Authorization")
			token := strings.TrimPrefix(tokenRaw, "Bearer ")
			if token == "" || token != desiredToken(user) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func desiredToken(_ string) string {
	// TODO: replace with a real token source (e.g., config file, env var,
	// or database lookup). This function intentionally panics to prevent
	// accidental use before implementation.
	panic("api: TokenMiddleware is not implemented — do not wire into routes")
}
