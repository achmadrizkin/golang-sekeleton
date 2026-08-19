package middleware

import (
	"net/http"
	"strings"
)

// CorsConfig controls which origins/methods/headers are allowed.
type CorsConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

// CORSMiddleware applies cfg to every response and short-circuits
// preflight (OPTIONS) requests.
func CORSMiddleware(cfg *CorsConfig) func(http.Handler) http.Handler {
	allowOrigins := joinOr(cfg.AllowedOrigins, "*")
	allowMethods := joinOr(cfg.AllowedMethods, "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	allowHeaders := joinOr(cfg.AllowedHeaders, "Authorization,Content-Type")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && originAllowed(origin, cfg.AllowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if len(cfg.AllowedOrigins) == 0 {
				w.Header().Set("Access-Control-Allow-Origin", allowOrigins)
			}
			w.Header().Set("Access-Control-Allow-Methods", allowMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func originAllowed(origin string, allowed []string) bool {
	for _, o := range allowed {
		if o == "*" || o == origin {
			return true
		}
	}
	return false
}

func joinOr(values []string, fallback string) string {
	if len(values) == 0 {
		return fallback
	}
	return strings.Join(values, ",")
}
