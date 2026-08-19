package middleware

import (
	"net/http"
	"time"
)

// TimeoutMiddleware cancels the request context after d, except for paths
// in excludedPaths — used for long-lived connections (e.g. SSE streams)
// that must not be cut off. Wraps http.TimeoutHandler for the standard
// "504 after deadline" behaviour.
func TimeoutMiddleware(d time.Duration, excludedPaths map[string]bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		timeoutHandler := http.TimeoutHandler(next, d, "request timed out")
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if excludedPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}
			timeoutHandler.ServeHTTP(w, r)
		})
	}
}
