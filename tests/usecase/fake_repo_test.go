// Package usecase_test exercises internal/usecase from outside the
// package (only its public API), against hand-written fakes for the
// repository interfaces — demonstrating why usecase depending only on
// interfaces.DBReadWriter/CacheReadWriter/MessagePublisher (never a
// concrete driver) matters: none of these tests touch a real database,
// Redis, or message broker.
package usecase_test

import (
	"context"
	"database/sql"

	"github.com/fauzie/golang-sekeleton/internal/domain"
)

// fakeDB implements repointerface.DBReadWriter with overridable funcs, so
// each test only wires up the methods it actually exercises.
type fakeDB struct {
	addUserFn       func(ctx context.Context, user *domain.User) error
	getUserByIDFn   func(ctx context.Context, id string) (*domain.User, error)
	getAllUsersFn   func(ctx context.Context) ([]*domain.User, error)
	getAllPaginated func(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error)
	updateUserFn    func(ctx context.Context, user *domain.User) error
	deleteUserFn    func(ctx context.Context, id string) error
	searchUsersFn   func(ctx context.Context, query string) ([]*domain.User, error)
}

func (f *fakeDB) DB() *sql.DB                { return nil }
func (f *fakeDB) Close() error               { return nil }
func (f *fakeDB) Ping(context.Context) error { return nil }

func (f *fakeDB) AddUser(ctx context.Context, user *domain.User) error {
	return f.addUserFn(ctx, user)
}
func (f *fakeDB) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	return f.getAllUsersFn(ctx)
}
func (f *fakeDB) GetAllUsersPaginated(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error) {
	return f.getAllPaginated(ctx, page, pageSize)
}
func (f *fakeDB) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return f.getUserByIDFn(ctx, id)
}
func (f *fakeDB) UpdateUser(ctx context.Context, user *domain.User) error {
	return f.updateUserFn(ctx, user)
}
func (f *fakeDB) DeleteUser(ctx context.Context, id string) error {
	return f.deleteUserFn(ctx, id)
}
func (f *fakeDB) SearchUsers(ctx context.Context, query string) ([]*domain.User, error) {
	return f.searchUsersFn(ctx, query)
}

// fakeValidator always succeeds, so tests can focus on usecase behaviour
// instead of the (separately tested) validation rules.
type fakeValidator struct{}

func (fakeValidator) Validate(*domain.User) error { return nil }
