package middleware

import (
	"net/http"
	"vpod/internal/env"
)

func BasicAuth(env *env.Env, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		wanted := env.GetAuth()
		if !ok || !validateCredentials(user, pass, wanted) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func validateCredentials(username string, password string, wanted env.AuthInfo) bool {
	return username == wanted.User && password == wanted.Pass
}
