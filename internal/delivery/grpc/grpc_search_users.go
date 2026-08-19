package grpc

import (
	"context"
	"fmt"

	"github.com/fauzie/golang-sekeleton/internal/endpoint"
	pb "github.com/fauzie/golang-sekeleton/proto"
)

// Search matches users against username, email, and full_name.
func (s *UserService) Search(ctx context.Context, req *pb.SearchUsersRequest) (*pb.ResUserList, error) {
	resp, err := s.callEndpoint(ctx, s.endpoints.SearchUsersEndpoint,
		endpoint.SearchUsersRequest{Query: req.GetQuery()}, "failed to search users")
	if err != nil {
		return nil, err
	}

	searchResp, ok := resp.(endpoint.SearchUsersResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	if searchResp.Err != nil {
		return nil, searchResp.Err
	}

	return &pb.ResUserList{Items: s.mapper.DomainListToProto(searchResp.Users)}, nil
}
