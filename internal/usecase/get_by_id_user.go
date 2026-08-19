package usecase

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/fauzie/golang-sekeleton/internal/domain"
)

// GetByID gets a user by ID using the cache-aside pattern: try cache (read
// from a replica when configured, via repo.ReadCache()) first, fall
// through to the database on a miss, then populate the cache best-effort.
func (uc *UserUseCase) GetByID(ctx context.Context, id string) (*domain.User, error) {
	logger := uc.logger.WithContext(ctx)

	if cacheRepo := uc.repo.ReadCache(); cacheRepo != nil {
		if user, err := cacheRepo.GetUser(ctx, id); err == nil {
			logger.Debug("user found in cache", zap.String("user_id", id))
			return user, nil
		}
	}

	user, err := uc.repo.DBReadWriter.GetUserByID(ctx, id)
	if err != nil {
		logger.Error("failed to get user by ID from DB", zap.Error(err), zap.String("user_id", id))
		return nil, fmt.Errorf("usecase.GetByID: %w", err)
	}

	uc.setCacheWithBestEffort(ctx, user, "get")

	return user, nil
}
