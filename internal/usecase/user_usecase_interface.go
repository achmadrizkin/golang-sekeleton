package usecase

import (
	"context"

	"github.com/fauzie/golang-sekeleton/internal/domain"
)

// UserUseCaseInterface is what internal/endpoint depends on, so endpoint
// factories can be built (and unit-tested) against a stub without pulling
// in the real UserUseCase's infrastructure dependencies.
type UserUseCaseInterface interface {
	Add(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetAll(ctx context.Context) ([]*domain.User, error)
	GetAllPaginated(ctx context.Context, req *domain.PaginationRequest) (*domain.PaginatedUsers, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string) ([]*domain.User, error)

	// ProcessUserCreated / ProcessUserDeleted are invoked by the message
	// consumers (internal/delivery/messaging) — kept on the same
	// interface as everything else so they get singleflight for free
	// where wired in internal/endpoint.
	ProcessUserCreated(ctx context.Context, id string) error
	ProcessUserDeleted(ctx context.Context, id string) error
}
