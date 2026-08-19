package cache

import (
	"context"
	"encoding/json"

	"github.com/fauzie/golang-sekeleton/internal/domain"
	apperrors "github.com/fauzie/golang-sekeleton/pkg/errors"
)

// GetUser fetches a user from cache. A miss is reported as NotFoundError,
// which usecase code treats as a signal to fall through to the database —
// not as a request-failing error.
func (rw *cacheReadWriter) GetUser(ctx context.Context, id string) (*domain.User, error) {
	key := UserCachePrefix + id

	val, err := rw.client.Get(ctx, key).Result()
	if err != nil {
		return nil, apperrors.NewNotFoundError("User cache", id)
	}

	var user domain.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, apperrors.NewDataAccessError("failed to unmarshal user from cache", err)
	}
	return &user, nil
}
