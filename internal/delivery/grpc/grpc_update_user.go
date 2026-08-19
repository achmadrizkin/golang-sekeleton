package grpc

import (
	"context"
	"fmt"

	"github.com/fauzie/golang-sekeleton/internal/endpoint"
	pb "github.com/fauzie/golang-sekeleton/proto"
)

// Update updates an existing user.
func (s *UserService) Update(ctx context.Context, req *pb.UpdateUserRequest) (*pb.ResUserMessage, error) {
	user := s.mapper.ProtoToDomain(req.GetUser())
	user.ID = req.GetId()

	resp, err := s.callEndpoint(ctx, s.endpoints.UpdateUserEndpoint,
		endpoint.UpdateUserRequest{User: user}, "failed to update user")
	if err != nil {
		return nil, err
	}

	updateResp, ok := resp.(endpoint.UpdateUserResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	if updateResp.Err != nil {
		return nil, updateResp.Err
	}

	return &pb.ResUserMessage{Message: "OK"}, nil
}
