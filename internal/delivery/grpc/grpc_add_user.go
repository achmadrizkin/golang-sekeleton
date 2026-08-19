package grpc

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/fauzie/golang-sekeleton/internal/endpoint"
	pb "github.com/fauzie/golang-sekeleton/proto"
)

// Add adds a new user.
func (s *UserService) Add(ctx context.Context, req *pb.UserModel) (*pb.ResUserMessage, error) {
	s.logger.Info("add user request received", zap.String("username", req.GetUsername().GetValue()))

	user := s.mapper.ProtoToDomain(req)

	resp, err := s.callEndpoint(ctx, s.endpoints.AddUserEndpoint,
		endpoint.AddUserRequest{User: user}, "failed to add user")
	if err != nil {
		return nil, err
	}

	addResp, ok := resp.(endpoint.AddUserResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	if addResp.Err != nil {
		return nil, addResp.Err
	}

	return &pb.ResUserMessage{Message: "OK"}, nil
}
