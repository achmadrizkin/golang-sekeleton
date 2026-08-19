package cache

import (
	"context"
	"encoding/json"

	"github.com/fauzie/golang-sekeleton/internal/domain"
	apperrors "github.com/fauzie/golang-sekeleton/pkg/errors"
)

// SetUser writes a user into cache with UserCacheTTL.
func (rw *cacheReadWriter) SetUser(ctx context.Context, user *domain.User) error {
	key := UserCachePrefix + user.ID

	data, err := json.Marshal(user)
	if err != nil {
		return apperrors.NewDataAccessError("failed to marshal user", err)
	}

	if err := rw.client.Set(ctx, key, data, UserCacheTTL).Err(); err != nil {
		return apperrors.NewDataAccessError("failed to set user in cache", err)
	}
	return nil
}
