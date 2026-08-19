package middleware

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/fauzie/golang-sekeleton/pkg/logger"
)

// RecoveryMiddleware converts a panic anywhere downstream into a 500
// response instead of crashing the process, and logs the panic value plus
// a stack trace for diagnosis. It must be the outermost-executing (i.e.
// innermost-wrapped, applied last) middleware so nothing downstream of it
// can panic past it.
func RecoveryMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.WithContext(r.Context()).Error("panic recovered",
						zap.Any("panic", rec), zap.Stack("stack"))
					http.Error(w, fmt.Sprintf("internal server error: %v", rec), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
