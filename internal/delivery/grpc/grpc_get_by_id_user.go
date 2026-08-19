package grpc

import (
	"context"
	"fmt"

	"github.com/fauzie/golang-sekeleton/internal/endpoint"
	pb "github.com/fauzie/golang-sekeleton/proto"
)

// GetById gets a user by ID.
func (s *UserService) GetById(ctx context.Context, req *pb.UserByIdRequest) (*pb.UserModel, error) {
	resp, err := s.callEndpoint(ctx, s.endpoints.GetUserByIDEndpoint,
		endpoint.GetUserByIDRequest{ID: req.GetId()}, "failed to get user")
	if err != nil {
		return nil, err
	}

	getResp, ok := resp.(endpoint.GetUserByIDResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	if getResp.Err != nil {
		return nil, getResp.Err
	}

	return s.mapper.DomainToProto(getResp.User), nil
}
