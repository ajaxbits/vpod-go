package middleware

import (
	"net/http"
)

// MaxConcurrent returns middleware that limits the number of concurrent
// requests being processed. Requests beyond the limit receive 503.
func MaxConcurrent(n int) func(http.Handler) http.Handler {
	sem := make(chan struct{}, n)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
				next.ServeHTTP(w, r)
			default:
				http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			}
		})
	}
}
