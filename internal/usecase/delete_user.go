package usecase

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/fauzie/golang-sekeleton/internal/domain"
)

// Delete removes a user (must succeed), then evicts its cache entry and
// publishes user.deleted (both best effort).
func (uc *UserUseCase) Delete(ctx context.Context, id string) error {
	logger := uc.logger.WithContext(ctx)

	if err := uc.repo.DBReadWriter.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("usecase.Delete: %w", err)
	}

	uc.deleteCacheWithBestEffort(ctx, id, "delete")

	if uc.repo.MessagePublisher != nil {
		payload := &domain.UserDeletedPayload{ID: id, DeletedAt: timeNow()}
		if err := uc.repo.MessagePublisher.PublishUserDeleted(ctx, payload); err != nil {
			logger.Error("failed to publish user.deleted event", zap.Error(err), zap.String("user_id", id))
		}
	}

	logger.Info("user deleted successfully", zap.String("user_id", id))
	return nil
}
