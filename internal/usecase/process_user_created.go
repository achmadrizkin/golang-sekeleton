package usecase

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// ProcessUserCreated is invoked by the user.created message consumer. It
// must be idempotent — every broker here guarantees at-least-once
// delivery, so this may run more than once for the same ID (retry,
// rebalance, redeliver after a crash). Re-fetching and re-caching is safe
// to repeat; this is also where you'd hook side effects like sending a
// welcome email, as long as that side effect is itself idempotent or
// de-duplicated upstream.
func (uc *UserUseCase) ProcessUserCreated(ctx context.Context, id string) error {
	logger := uc.logger.WithContext(ctx)

	user, err := uc.repo.DBReadWriter.GetUserByID(ctx, id)
	if err != nil {
		return fmt.Errorf("usecase.ProcessUserCreated: %w", err)
	}

	uc.setCacheWithBestEffort(ctx, user, "process_user_created")

	logger.Info("processed user.created event", zap.String("user_id", id))
	return nil
}
