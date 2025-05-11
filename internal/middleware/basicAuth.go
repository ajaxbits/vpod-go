package middleware

import (
	"net/http"
)

type AuthInfo struct {
	User string
	Pass string
}

func NewBasicAuth(wanted AuthInfo) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return basicAuth(wanted, next)
	}
}

func basicAuth(wanted AuthInfo, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || !validateCredentials(user, pass, wanted) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func validateCredentials(username string, password string, wanted AuthInfo) bool {
	return username == wanted.User && password == wanted.Pass
}
