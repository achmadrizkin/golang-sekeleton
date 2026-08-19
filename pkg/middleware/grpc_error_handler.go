package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	apperrors "github.com/fauzie/golang-sekeleton/pkg/errors"
)

// UnaryServerErrorHandler translates the typed errors from pkg/errors into
// the matching gRPC status code, so handlers can just return an *AppError
// and never think about codes.* directly. Any error already carrying a
// gRPC status (e.g. produced by an earlier interceptor) passes through
// unchanged.
func UnaryServerErrorHandler() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}
		if _, ok := status.FromError(err); ok {
			// err is already a gRPC status error (e.g. raised by an earlier
			// interceptor) — pass it through unchanged.
			return resp, err
		}
		return resp, status.Error(mapCode(apperrors.CodeOf(err)), err.Error())
	}
}

func mapCode(code apperrors.Code) codes.Code {
	switch code {
	case apperrors.CodeValidation:
		return codes.InvalidArgument
	case apperrors.CodeNotFound:
		return codes.NotFound
	case apperrors.CodeDuplicateKey:
		return codes.AlreadyExists
	case apperrors.CodeUnauthorized:
		return codes.Unauthenticated
	case apperrors.CodeForbidden:
		return codes.PermissionDenied
	case apperrors.CodeConflict:
		return codes.FailedPrecondition
	case apperrors.CodeDataAccess, apperrors.CodeInternal:
		return codes.Internal
	default:
		return codes.Unknown
	}
}
