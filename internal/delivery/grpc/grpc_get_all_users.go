package grpc

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/fauzie/golang-sekeleton/internal/endpoint"
	pb "github.com/fauzie/golang-sekeleton/proto"
)

// GetAll retrieves every user.
func (s *UserService) GetAll(ctx context.Context, _ *emptypb.Empty) (*pb.ResUserList, error) {
	resp, err := s.callEndpoint(ctx, s.endpoints.GetAllUsersEndpoint,
		endpoint.GetAllUsersRequest{}, "failed to get all users")
	if err != nil {
		return nil, err
	}

	getResp, ok := resp.(endpoint.GetAllUsersResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	if getResp.Err != nil {
		return nil, getResp.Err
	}

	return &pb.ResUserList{Items: s.mapper.DomainListToProto(getResp.Users)}, nil
}
