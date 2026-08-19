package usecase

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/fauzie/golang-sekeleton/internal/domain"
)

// Update validates and persists changes to an existing user, then
// refreshes (not just invalidates) its cache entry so the next read is
// still a cache hit with fresh data.
func (uc *UserUseCase) Update(ctx context.Context, user *domain.User) error {
	logger := uc.logger.WithContext(ctx)

	if err := uc.validator.Validate(user); err != nil {
		return fmt.Errorf("usecase.Update.Validate: %w", err)
	}

	if err := uc.repo.DBReadWriter.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("usecase.Update: %w", err)
	}

	uc.setCacheWithBestEffort(ctx, user, "update")

	logger.Info("user updated successfully", zap.String("user_id", user.ID))
	return nil
}
