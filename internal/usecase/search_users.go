package usecase

import (
	"context"
	"fmt"

	"github.com/fauzie/golang-sekeleton/internal/domain"
)

// Search matches query against username, email, and full_name.
func (uc *UserUseCase) Search(ctx context.Context, query string) ([]*domain.User, error) {
	users, err := uc.repo.DBReadWriter.SearchUsers(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("usecase.Search: %w", err)
	}
	return users, nil
}
