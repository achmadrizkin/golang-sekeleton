package middleware

import (
	"net/http"
	"strings"

	"github.com/fauzie/golang-sekeleton/pkg/claims"
)

// ClaimsMiddleware extracts and validates the Bearer token from the
// Authorization header, storing the resulting claims.Claims on the request
// context. When validateJWT is false (e.g. local dev), an Authorization
// header is still parsed into claims if present, but its absence does not
// reject the request — that decision belongs to the per-method auth
// bypass list, not this middleware.
func ClaimsMiddleware(cfg *claims.JWTConfig, validateJWT bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r.Header.Get("Authorization"))
			if token != "" {
				c, err := claims.ParseAndValidate(token, *cfg)
				if err == nil {
					r = r.WithContext(claims.SetClaims(r.Context(), c))
				} else if validateJWT {
					http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
					return
				}
			} else if validateJWT {
				http.Error(w, "unauthorized: missing bearer token", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if strings.HasPrefix(header, prefix) {
		return strings.TrimPrefix(header, prefix)
	}
	return ""
}
