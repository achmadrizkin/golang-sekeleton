package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/fauzie/golang-sekeleton/internal/domain"
	apperrors "github.com/fauzie/golang-sekeleton/pkg/errors"
)

// GetUserByID retrieves a user by its ID.
func (rw *dbReadWriter) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, rw.queryTimeout)
	defer cancel()

	query := fmt.Sprintf(`SELECT %s FROM %s WHERE id = %s`, UserFields, UserTableName, rw.dialect.arg(1))

	user := &domain.User{}
	err := rw.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.FullName,
		&user.Avatar, &user.Role, &user.IsActive, &user.EmailVerified,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.NewNotFoundError("User", id)
	}
	if err != nil {
		return nil, apperrors.NewDataAccessError("failed to query user", err)
	}
	return user, nil
}
