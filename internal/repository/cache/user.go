package cache

import "time"

const (
	// UserCachePrefix prefixes every user cache key.
	UserCachePrefix = "user:"
	// UserCacheTTL is how long a cached user entry lives.
	UserCacheTTL = 30 * time.Minute
)
