package middleware

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/fauzie/golang-sekeleton/pkg/logger"
)

// UnaryServerLogger logs one line per RPC (method, gRPC status code,
// duration), skipping methods in excludedPaths (health checks).
func UnaryServerLogger(log *logger.Logger, excludedPaths map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if excludedPaths[info.FullMethod] {
			return handler(ctx, req)
		}

		start := time.Now()
		resp, err := handler(ctx, req)

		log.WithContext(ctx).Info("grpc request",
			zap.String("method", info.FullMethod),
			zap.String("code", status.Code(err).String()),
			zap.Duration("duration", time.Since(start)),
		)
		return resp, err
	}
}
