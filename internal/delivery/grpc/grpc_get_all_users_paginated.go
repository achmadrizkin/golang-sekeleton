package grpc

import (
	"context"
	"fmt"

	"github.com/fauzie/golang-sekeleton/internal/endpoint"
	pb "github.com/fauzie/golang-sekeleton/proto"
)

// GetAllPaginated retrieves one page of users.
func (s *UserService) GetAllPaginated(ctx context.Context, req *pb.PaginationRequest) (*pb.ResUserPaginated, error) {
	resp, err := s.callEndpoint(ctx, s.endpoints.GetAllUsersPaginatedEndpoint,
		endpoint.GetAllUsersPaginatedRequest{Page: int(req.GetPage()), PageSize: int(req.GetPageSize())},
		"failed to get paginated users")
	if err != nil {
		return nil, err
	}

	getResp, ok := resp.(endpoint.GetAllUsersPaginatedResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}
	if getResp.Err != nil {
		return nil, getResp.Err
	}

	return &pb.ResUserPaginated{
		Items: s.mapper.DomainListToProto(getResp.PaginatedUsers.Items),
		Pagination: &pb.PaginationMeta{
			Page:       int32(getResp.PaginatedUsers.Pagination.Page),
			PageSize:   int32(getResp.PaginatedUsers.Pagination.PageSize),
			TotalItems: getResp.PaginatedUsers.Pagination.TotalItems,
			TotalPages: int32(getResp.PaginatedUsers.Pagination.TotalPages),
		},
	}, nil
}
