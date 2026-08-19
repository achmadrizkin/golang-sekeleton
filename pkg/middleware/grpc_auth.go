package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fauzie/golang-sekeleton/pkg/claims"
)

// UnaryServerAuthInterceptor rejects any request that reached it without
// authenticated claims, unless info.FullMethod is in publicMethods. Add new
// public RPCs to the list passed at server.New — see the
// "[INJECTION POINT: Public gRPC Methods]" marker in internal/server/server.go.
func UnaryServerAuthInterceptor(publicMethods []string) grpc.UnaryServerInterceptor {
	public := make(map[string]bool, len(publicMethods))
	for _, m := range publicMethods {
		public[m] = true
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if public[info.FullMethod] {
			return handler(ctx, req)
		}
		if !claims.GetClaims(ctx).IsAuthenticated() {
			return nil, status.Error(codes.Unauthenticated, "authentication required")
		}
		return handler(ctx, req)
	}
}
