package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/fauzie/golang-sekeleton/pkg/claims"
)

// UnaryServerClaims extracts a Bearer token from the "authorization"
// metadata key, validates it, and stores the resulting claims.Claims on the
// context for downstream handlers/interceptors. Mirrors ClaimsMiddleware's
// HTTP behaviour: it does not itself reject missing/invalid tokens (that is
// UnaryServerAuthInterceptor's job, since some methods are public).
func UnaryServerClaims(cfg *claims.JWTConfig, validateJWT bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		token := bearerFromMetadata(ctx)
		if token != "" {
			c, err := claims.ParseAndValidate(token, *cfg)
			if err == nil {
				ctx = claims.SetClaims(ctx, c)
			} else if validateJWT {
				return nil, status.Error(codes.Unauthenticated, "invalid token: "+err.Error())
			}
		} else if validateJWT {
			return nil, status.Error(codes.Unauthenticated, "missing bearer token")
		}
		return handler(ctx, req)
	}
}

func bearerFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get("authorization")
	if len(values) == 0 {
		return ""
	}
	const prefix = "Bearer "
	v := values[0]
	if len(v) > len(prefix) && v[:len(prefix)] == prefix {
		return v[len(prefix):]
	}
	return ""
}
