package usecase

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/fauzie/golang-sekeleton/internal/domain"
	"github.com/fauzie/golang-sekeleton/internal/repository"
	"github.com/fauzie/golang-sekeleton/internal/validator"
	"github.com/fauzie/golang-sekeleton/pkg/logger"
	"github.com/fauzie/golang-sekeleton/pkg/util"
)

// UserUseCase implements UserUseCaseInterface. It depends only on
// repository interfaces (via *repository.Repository) and a validator
// interface — never on proto, delivery, or a concrete driver, which is
// what lets it be unit-tested without any real infrastructure.
type UserUseCase struct {
	repo      *repository.Repository
	validator validator.UserValidatorInterface
	logger    *logger.Logger
}

// NewUserUseCase builds a UserUseCase.
func NewUserUseCase(repo *repository.Repository, v validator.UserValidatorInterface, log *logger.Logger) *UserUseCase {
	return &UserUseCase{repo: repo, validator: v, logger: log}
}

// generateUUID is the single place a new User ID is generated.
func generateUUID() string { return util.NewUUID() }

// timeNow is the single place "now" is read, kept as a var so tests can
// override it deterministically.
var timeNow = func() time.Time { return time.Now().UTC() }

// setCacheWithBestEffort caches a user, logging (never returning) failures:
// a cache write failing must not fail the request that triggered it.
func (uc *UserUseCase) setCacheWithBestEffort(ctx context.Context, user *domain.User, operation string) {
	if uc.repo.Cache == nil {
		return
	}
	if err := uc.repo.Cache.SetUser(ctx, user); err != nil {
		uc.logger.WithContext(ctx).Warn(
			fmt.Sprintf("failed to set user cache after %s", operation),
			zap.Error(err), zap.String("user_id", user.ID))
	}
}

// deleteCacheWithBestEffort evicts a user's cache entry, logging (never
// returning) failures.
func (uc *UserUseCase) deleteCacheWithBestEffort(ctx context.Context, id, operation string) {
	if uc.repo.Cache == nil {
		return
	}
	if err := uc.repo.Cache.DeleteUser(ctx, id); err != nil {
		uc.logger.WithContext(ctx).Warn(
			fmt.Sprintf("failed to delete user cache after %s", operation),
			zap.Error(err), zap.String("user_id", id))
	}
}
