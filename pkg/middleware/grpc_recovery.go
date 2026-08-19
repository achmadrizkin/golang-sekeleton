package middleware

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fauzie/golang-sekeleton/pkg/logger"
)

// UnaryServerRecovery converts a panic in a handler into codes.Internal
// instead of crashing the process.
func UnaryServerRecovery(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if rec := recover(); rec != nil {
				log.WithContext(ctx).Error("panic recovered",
					zap.Any("panic", rec), zap.String("method", info.FullMethod), zap.Stack("stack"))
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}
