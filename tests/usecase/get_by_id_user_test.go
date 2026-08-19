package usecase_test

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fauzie/golang-sekeleton/internal/domain"
	"github.com/fauzie/golang-sekeleton/internal/repository"
	"github.com/fauzie/golang-sekeleton/internal/usecase"
	apperrors "github.com/fauzie/golang-sekeleton/pkg/errors"
)

// fakeCache implements repointerface.CacheReadWriter with overridable funcs.
type fakeCache struct {
	getUserFn func(ctx context.Context, id string) (*domain.User, error)
	setUserFn func(ctx context.Context, user *domain.User) error
}

func (f *fakeCache) Close() error                             { return nil }
func (f *fakeCache) Ping(context.Context) error               { return nil }
func (f *fakeCache) GetClient() redis.UniversalClient         { return nil }
func (f *fakeCache) DeleteUser(context.Context, string) error { return nil }
func (f *fakeCache) ClearUserCache(context.Context) error     { return nil }

func (f *fakeCache) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return f.getUserFn(ctx, id)
}
func (f *fakeCache) SetUser(ctx context.Context, user *domain.User) error {
	if f.setUserFn == nil {
		return nil
	}
	return f.setUserFn(ctx, user)
}

func TestGetByID_ReturnsCacheHitWithoutTouchingDB(t *testing.T) {
	cached := &domain.User{ID: "u1", Username: "cached"}
	cache := &fakeCache{getUserFn: func(context.Context, string) (*domain.User, error) { return cached, nil }}
	db := &fakeDB{getUserByIDFn: func(context.Context, string) (*domain.User, error) {
		t.Fatal("GetByID should not hit the database on a cache hit")
		return nil, nil
	}}

	uc := usecase.NewUserUseCase(&repository.Repository{DBReadWriter: db, Cache: cache}, fakeValidator{}, newTestLogger(t))

	got, err := uc.GetByID(context.Background(), "u1")
	require.NoError(t, err)
	assert.Same(t, cached, got)
}

func TestGetByID_FallsThroughToDBOnCacheMiss(t *testing.T) {
	fromDB := &domain.User{ID: "u1", Username: "from-db"}
	var cachedAfter *domain.User

	cache := &fakeCache{
		getUserFn: func(context.Context, string) (*domain.User, error) {
			return nil, apperrors.NewNotFoundError("User cache", "u1")
		},
		setUserFn: func(_ context.Context, user *domain.User) error {
			cachedAfter = user
			return nil
		},
	}
	db := &fakeDB{getUserByIDFn: func(context.Context, string) (*domain.User, error) { return fromDB, nil }}

	uc := usecase.NewUserUseCase(&repository.Repository{DBReadWriter: db, Cache: cache}, fakeValidator{}, newTestLogger(t))

	got, err := uc.GetByID(context.Background(), "u1")
	require.NoError(t, err)
	assert.Same(t, fromDB, got)
	assert.Same(t, fromDB, cachedAfter, "a DB hit should populate the cache best-effort")
}
