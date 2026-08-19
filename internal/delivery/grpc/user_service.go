// Package grpc adapts internal/endpoint into the generated gRPC service
// interface. Handlers here do exactly three things: map proto -> domain,
// call the endpoint, map the result -> proto. No business logic lives here.
package grpc

import (
	"context"

	"go.uber.org/zap"

	"github.com/fauzie/golang-sekeleton/internal/endpoint"
	"github.com/fauzie/golang-sekeleton/internal/mapper"
	"github.com/fauzie/golang-sekeleton/pkg/logger"
	pb "github.com/fauzie/golang-sekeleton/proto"
)

// UserService implements pb.UserGrpcServiceServer.
type UserService struct {
	pb.UnimplementedUserGrpcServiceServer
	endpoints endpoint.UserEndpoints
	mapper    *mapper.UserMapper
	logger    *logger.Logger
}

// NewUserService builds a UserService.
func NewUserService(endpoints endpoint.UserEndpoints, m *mapper.UserMapper, log *logger.Logger) *UserService {
	return &UserService{endpoints: endpoints, mapper: m, logger: log}
}

// callEndpoint standardizes error logging around an endpoint invocation.
func (s *UserService) callEndpoint(ctx context.Context, ep endpoint.Endpoint, request interface{}, operation string) (interface{}, error) {
	resp, err := ep(ctx, request)
	if err != nil {
		s.logger.WithContext(ctx).Error(operation, zap.Error(err))
		return nil, err
	}
	return resp, nil
}
