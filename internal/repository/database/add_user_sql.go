package database

import (
	"context"
	"fmt"

	"github.com/fauzie/golang-sekeleton/internal/domain"
	apperrors "github.com/fauzie/golang-sekeleton/pkg/errors"
)

// AddUser inserts a new user row.
func (rw *dbReadWriter) AddUser(ctx context.Context, user *domain.User) error {
	ctx, cancel := context.WithTimeout(ctx, rw.queryTimeout)
	defer cancel()

	query := fmt.Sprintf(
		`INSERT INTO %s (id, username, email, password, full_name, avatar, role, is_active, email_verified, created_at, updated_at)
		 VALUES (%s, NOW(), NOW())`,
		UserTableName, rw.dialect.args(1, 9))

	_, err := rw.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.Password, user.FullName,
		user.Avatar, user.Role, user.IsActive, user.EmailVerified,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return apperrors.NewDuplicateKeyError(
				user.ID, fmt.Sprintf("User with ID '%s' already exists", user.ID), err)
		}
		return apperrors.NewDataAccessError("failed to insert user", err)
	}
	return nil
}
