// Package repository assembles the concrete Repository from a list of
// factories, one per infrastructure dependency (database, cache, message
// broker). Nothing outside this package and internal/server knows or cares
// which SQL driver, Redis topology, or message broker backs a given
// Repository — usecase code only ever sees the interfaces in
// internal/repository/interface.
package repository

import (
	"database/sql"
	"fmt"

	repointerface "github.com/fauzie/golang-sekeleton/internal/repository/interface"
	"github.com/fauzie/golang-sekeleton/pkg/logger"
)

// Kind identifies which Repository field a factory's output belongs in.
// Two factories can build the exact same concrete type (e.g. the master
// cache and a read-only replica are both *cache.Conf producing the same
// struct) so the factory declares its role explicitly instead of relying
// on a type switch.
type Kind string

const (
	KindDB           Kind = "db"
	KindCache        Kind = "cache"
	KindCacheReplica Kind = "cache_replica"
	KindMQPublisher  Kind = "mq_publisher"
)

// RepoFactory builds one piece of the Repository. Config types
// (database.Conf, cache.Conf, mq.PublisherConf, ...) implement this.
type RepoFactory interface {
	Kind() Kind
	Build() (interface{}, error)
}

// Repository holds every infrastructure dependency the service was wired
// with. Any field may be nil when that dependency wasn't configured
// (e.g. no cache in a given environment) — callers must check, which is
// exactly what ReadCache()/DBHandle() below do so they only have to check
// once.
type Repository struct {
	DBReadWriter     repointerface.DBReadWriter
	Cache            repointerface.CacheReadWriter
	CacheReplica     repointerface.CacheReader
	MessagePublisher repointerface.MessagePublisher
}

// NewRepository runs every factory and assembles the results into one
// Repository. A failure in any factory aborts the whole assembly — a
// service should not start half-wired.
func NewRepository(factories []RepoFactory, log *logger.Logger) (*Repository, error) {
	repo := &Repository{}

	for _, f := range factories {
		built, err := f.Build()
		if err != nil {
			return nil, fmt.Errorf("repository: build %s: %w", f.Kind(), err)
		}

		switch f.Kind() {
		case KindDB:
			v, ok := built.(repointerface.DBReadWriter)
			if !ok {
				return nil, fmt.Errorf("repository: %s factory returned %T, want DBReadWriter", f.Kind(), built)
			}
			repo.DBReadWriter = v
		case KindCache:
			v, ok := built.(repointerface.CacheReadWriter)
			if !ok {
				return nil, fmt.Errorf("repository: %s factory returned %T, want CacheReadWriter", f.Kind(), built)
			}
			repo.Cache = v
		case KindCacheReplica:
			v, ok := built.(repointerface.CacheReader)
			if !ok {
				return nil, fmt.Errorf("repository: %s factory returned %T, want CacheReader", f.Kind(), built)
			}
			repo.CacheReplica = v
		case KindMQPublisher:
			v, ok := built.(repointerface.MessagePublisher)
			if !ok {
				return nil, fmt.Errorf("repository: %s factory returned %T, want MessagePublisher", f.Kind(), built)
			}
			repo.MessagePublisher = v
		default:
			return nil, fmt.Errorf("repository: unknown factory kind %q", f.Kind())
		}

		log.Info(fmt.Sprintf("repository: %s wired", f.Kind()))
	}

	return repo, nil
}

// ReadCache returns the cache to use for reads: the read-only replica when
// one is configured, otherwise the master cache (which may itself be nil
// when no cache is wired at all). Centralizing this means no call site has
// to repeat the "replica if present, else master" check.
func (r *Repository) ReadCache() repointerface.CacheReader {
	if r.CacheReplica != nil {
		return r.CacheReplica
	}
	if r.Cache != nil {
		return r.Cache
	}
	return nil
}

// DBHandle returns the shared *sql.DB, or nil when no database is wired.
func (r *Repository) DBHandle() *sql.DB {
	if r.DBReadWriter == nil {
		return nil
	}
	return r.DBReadWriter.DB()
}

// Close shuts down every wired dependency, collecting (not stopping on)
// individual failures so one broken connection doesn't prevent the others
// from closing during shutdown.
func (r *Repository) Close() error {
	var errs []error
	closers := []struct {
		name string
		c    interface{ Close() error }
	}{
		{"database", r.DBReadWriter},
		{"cache", r.Cache},
		{"cache_replica", r.CacheReplica},
		{"message_publisher", r.MessagePublisher},
	}
	for _, entry := range closers {
		if entry.c == nil {
			continue
		}
		if err := entry.c.Close(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", entry.name, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("repository: close errors: %v", errs)
	}
	return nil
}
