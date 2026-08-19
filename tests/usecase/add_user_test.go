package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fauzie/golang-sekeleton/internal/domain"
	"github.com/fauzie/golang-sekeleton/internal/repository"
	"github.com/fauzie/golang-sekeleton/internal/usecase"
	"github.com/fauzie/golang-sekeleton/pkg/logger"
)

func newTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	log, err := logger.New(logger.Config{Environment: "development", Level: "error"})
	require.NoError(t, err)
	return log
}

func TestAdd_GeneratesIDAndPersists(t *testing.T) {
	var persisted *domain.User

	db := &fakeDB{
		addUserFn: func(_ context.Context, user *domain.User) error {
			persisted = user
			return nil
		},
	}

	uc := usecase.NewUserUseCase(&repository.Repository{DBReadWriter: db}, fakeValidator{}, newTestLogger(t))

	user := &domain.User{Username: "johndoe", Email: "john@example.com", Password: "password123", FullName: "John Doe"}
	err := uc.Add(context.Background(), user)

	require.NoError(t, err)
	assert.NotEmpty(t, user.ID, "Add should generate an ID when none is given")
	require.NotNil(t, persisted)
	assert.Equal(t, user.ID, persisted.ID)
}

func TestAdd_KeepsCallerSuppliedID(t *testing.T) {
	db := &fakeDB{addUserFn: func(context.Context, *domain.User) error { return nil }}
	uc := usecase.NewUserUseCase(&repository.Repository{DBReadWriter: db}, fakeValidator{}, newTestLogger(t))

	user := &domain.User{ID: "existing-id", Username: "johndoe", Email: "john@example.com", Password: "password123", FullName: "John Doe"}
	require.NoError(t, uc.Add(context.Background(), user))

	assert.Equal(t, "existing-id", user.ID)
}
