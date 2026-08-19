package usecase

import (
	"context"
	"fmt"

	"github.com/fauzie/golang-sekeleton/internal/domain"
)

// GetAll retrieves every user. Unbounded on purpose (see GetAllPaginated
// for the normal listing path) — intended for small admin/reporting use.
func (uc *UserUseCase) GetAll(ctx context.Context) ([]*domain.User, error) {
	users, err := uc.repo.DBReadWriter.GetAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("usecase.GetAll: %w", err)
	}
	return users, nil
}
