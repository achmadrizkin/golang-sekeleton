package grpc

import (
	"context"
	"fmt"

	"github.com/fauzie/golang-sekeleton/internal/endpoint"
	pb "github.com/fauzie/golang-sekeleton/proto"
)

// Delete removes a user.
func (s *UserService) Delete(ctx context.Context, req *pb.UserByIdRequest) (*pb.ResUserMessage, error) {
	resp, err := s.callEndpoint(ctx, s.endpoints.DeleteUserEndpoint,
		endpoint.DeleteUserRequest{ID: req.GetId()}, "failed to delete user")
	if err != nil {
		return nil, err
	}

	deleteResp, ok := resp.(endpoint.DeleteUserResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	if deleteResp.Err != nil {
		return nil, deleteResp.Err
	}

	return &pb.ResUserMessage{Message: "OK"}, nil
}
