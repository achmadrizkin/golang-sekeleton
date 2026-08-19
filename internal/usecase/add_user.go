package usecase

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/fauzie/golang-sekeleton/internal/domain"
)

// Add adds a new user: validates it, writes it to the database (must
// succeed), then caches it and publishes user.created (both best effort —
// their failure must not fail an otherwise-successful write).
func (uc *UserUseCase) Add(ctx context.Context, user *domain.User) error {
	logger := uc.logger.WithContext(ctx)

	if user.ID == "" {
		user.ID = generateUUID()
	}

	if err := uc.validator.Validate(user); err != nil {
		return fmt.Errorf("usecase.Add.Validate: %w", err)
	}

	if err := uc.repo.DBReadWriter.AddUser(ctx, user); err != nil {
		return fmt.Errorf("usecase.Add: %w", err)
	}

	uc.setCacheWithBestEffort(ctx, user, "add")

	if uc.repo.MessagePublisher != nil {
		payload := &domain.UserCreatedPayload{
			ID: user.ID, Username: user.Username, Email: user.Email,
			FullName: user.FullName, Role: user.Role,
			IsActive: user.IsActive, CreatedAt: user.CreatedAt,
		}
		if err := uc.repo.MessagePublisher.PublishUserCreated(ctx, payload); err != nil {
			logger.Error("failed to publish user.created event",
				zap.Error(err), zap.String("user_id", user.ID))
		}
	}

	logger.Info("user added successfully", zap.String("user_id", user.ID))
	return nil
}
