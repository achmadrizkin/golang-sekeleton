// Package repointerface defines the contracts the usecase layer depends on.
// Nothing in here knows about MySQL vs PostgreSQL, Redis vs a stub, or
// Kafka vs RabbitMQ vs Pub/Sub — that is exactly the point: usecase codes
// against these interfaces, and internal/repository/{database,cache,mq}
// provide the concrete implementations wired up in internal/server.
package repointerface

import (
	"context"
	"database/sql"
	"io"

	"github.com/redis/go-redis/v9"

	"github.com/fauzie/golang-sekeleton/internal/domain"
)

// DBReadWriter is the single gateway to the database: it owns the
// connection lifecycle (Ping/Close), exposes the shared *sql.DB via DB()
// for anything that needs it directly, and carries every entity operation
// this service needs. Adding a new entity means adding methods here, not a
// new interface.
type DBReadWriter interface {
	io.Closer
	// DB returns the underlying *sql.DB.
	DB() *sql.DB
	// Ping verifies connectivity for readiness checks.
	Ping(ctx context.Context) error

	// User operations
	AddUser(ctx context.Context, user *domain.User) error
	GetAllUsers(ctx context.Context) ([]*domain.User, error)
	GetAllUsersPaginated(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	DeleteUser(ctx context.Context, id string) error
	SearchUsers(ctx context.Context, query string) ([]*domain.User, error)
}

// CacheReader is implemented by both the master cache and any read-only
// replica, so usecase code can read through Repository.ReadCache() without
// caring which one it got.
type CacheReader interface {
	io.Closer
	GetUser(ctx context.Context, id string) (*domain.User, error)
	GetClient() redis.UniversalClient
	Ping(ctx context.Context) error
}

// CacheWriter is implemented only by the master cache — a read-only replica
// has no reason to satisfy it, which is what makes misdirected writes a
// compile error instead of a runtime surprise.
type CacheWriter interface {
	io.Closer
	SetUser(ctx context.Context, user *domain.User) error
	DeleteUser(ctx context.Context, id string) error
	ClearUserCache(ctx context.Context) error
}

// CacheReadWriter is the master cache's full interface.
type CacheReadWriter interface {
	CacheReader
	CacheWriter
}

// MessagePublisher exposes one typed method per event this service
// publishes, regardless of which broker backs a given topic. Add new
// entities' publish methods below the marker.
type MessagePublisher interface {
	io.Closer
	Ping(ctx context.Context) error

	PublishUserCreated(ctx context.Context, payload *domain.UserCreatedPayload) error
	PublishUserDeleted(ctx context.Context, payload *domain.UserDeletedPayload) error
	// [INJECTION POINT: MessagePublisher Methods]
}
