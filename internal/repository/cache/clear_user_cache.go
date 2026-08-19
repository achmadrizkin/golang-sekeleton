package cache

import (
	"context"

	apperrors "github.com/fauzie/golang-sekeleton/pkg/errors"
)

// ClearUserCache removes every cached user entry, scanning in batches so it
// is safe to run against a large keyspace without blocking Redis.
func (rw *cacheReadWriter) ClearUserCache(ctx context.Context) error {
	var cursor uint64
	for {
		keys, next, err := rw.client.Scan(ctx, cursor, UserCachePrefix+"*", 100).Result()
		if err != nil {
			return apperrors.NewDataAccessError("failed to scan user cache keys", err)
		}
		if len(keys) > 0 {
			if err := rw.client.Del(ctx, keys...).Err(); err != nil {
				return apperrors.NewDataAccessError("failed to clear user cache", err)
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}
