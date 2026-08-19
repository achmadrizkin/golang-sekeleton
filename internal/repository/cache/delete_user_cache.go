package cache

import (
	"context"

	apperrors "github.com/fauzie/golang-sekeleton/pkg/errors"
)

// DeleteUser removes a user's cache entry (used after update/delete so
// stale data isn't served on the next cache-aside read).
func (rw *cacheReadWriter) DeleteUser(ctx context.Context, id string) error {
	if err := rw.client.Del(ctx, UserCachePrefix+id).Err(); err != nil {
		return apperrors.NewDataAccessError("failed to delete user from cache", err)
	}
	return nil
}
