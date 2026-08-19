package usecase

import (
	"context"

	"go.uber.org/zap"

	"github.com/fauzie/golang-sekeleton/internal/domain"
	"github.com/fauzie/golang-sekeleton/pkg/util"
)

// GetAllPaginated retrieves one normalized page of users (page >= 1,
// 1 <= page_size <= 100 — see domain.PaginationRequest.Validate).
func (uc *UserUseCase) GetAllPaginated(ctx context.Context, req *domain.PaginationRequest) (*domain.PaginatedUsers, error) {
	logger := uc.logger.WithContext(ctx)

	if err := req.Validate(); err != nil {
		return nil, err
	}

	users, totalCount, err := uc.repo.DBReadWriter.GetAllUsersPaginated(ctx, req.Page, req.PageSize)
	if err != nil {
		logger.Error("failed to get paginated users", zap.Error(err))
		return nil, err
	}

	return &domain.PaginatedUsers{
		Items:      users,
		Pagination: util.NewPaginationMeta(req.Page, req.PageSize, totalCount),
	}, nil
}
