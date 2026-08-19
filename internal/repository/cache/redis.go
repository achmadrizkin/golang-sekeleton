// Package cache implements interfaces.CacheReadWriter on top of Redis
// (go-redis v9), supporting both single-node and cluster mode. The same
// concrete type also satisfies interfaces.CacheReader on its own, which is
// how a read-only replica (see Conf.Replica in internal/config) can be
// wired in as Repository.CacheReplica without a second implementation.
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"

	"github.com/fauzie/golang-sekeleton/internal/repository"
	repointerface "github.com/fauzie/golang-sekeleton/internal/repository/interface"
)

// Conf is the RepoFactory input for a Redis connection (master or replica).
// Set Replica: true to build a read-only replica connection, wired into
// Repository.CacheReplica instead of Repository.Cache.
type Conf struct {
	Mode     string // single|cluster
	Addrs    []string
	Password string
	DB       int
	Replica  bool
}

// Kind implements repository.RepoFactory.
func (c *Conf) Kind() repository.Kind {
	if c.Replica {
		return repository.KindCacheReplica
	}
	return repository.KindCache
}

type cacheReadWriter struct {
	client redis.UniversalClient
}

// Build implements repository.RepoFactory: connects to Redis (with OTel
// instrumentation) and returns interfaces.CacheReadWriter.
func (c *Conf) Build() (interface{}, error) {
	var client redis.UniversalClient
	if c.Mode == "cluster" {
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    c.Addrs,
			Password: c.Password,
		})
	} else {
		addr := "localhost:6379"
		if len(c.Addrs) > 0 {
			addr = c.Addrs[0]
		}
		client = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: c.Password,
			DB:       c.DB,
		})
	}

	if err := redisotel.InstrumentTracing(client); err != nil {
		return nil, fmt.Errorf("cache: instrument tracing: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cache: ping: %w", err)
	}

	rw := &cacheReadWriter{client: client}
	var _ repointerface.CacheReadWriter = rw // compile-time interface check
	return rw, nil
}

func (rw *cacheReadWriter) GetClient() redis.UniversalClient { return rw.client }

func (rw *cacheReadWriter) Close() error { return rw.client.Close() }

func (rw *cacheReadWriter) Ping(ctx context.Context) error {
	return rw.client.Ping(ctx).Err()
}
