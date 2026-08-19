package usecase

import (
	"context"

	"go.uber.org/zap"
)

// ProcessUserDeleted is invoked by the user.deleted message consumer.
// Cache eviction is naturally idempotent (deleting an already-absent key
// is a no-op), so redelivery is safe.
func (uc *UserUseCase) ProcessUserDeleted(ctx context.Context, id string) error {
	logger := uc.logger.WithContext(ctx)

	uc.deleteCacheWithBestEffort(ctx, id, "process_user_deleted")

	logger.Info("processed user.deleted event", zap.String("user_id", id))
	return nil
}
