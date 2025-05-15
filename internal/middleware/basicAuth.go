package middleware

import (
	"net/http"
)

func NewBasicAuth(wantedUser *string, wantedPass *string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			gotUser, gotPass, ok := r.BasicAuth()
			if !ok || !validateCredentials(gotUser, gotPass, *wantedUser, *wantedPass) {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func validateCredentials(gotUser string, gotPass string, wantedUser string, wantedPass string) bool {
	return gotUser == wantedUser && gotPass == wantedPass
}
