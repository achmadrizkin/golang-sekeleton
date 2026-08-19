package database

import (
	"context"
	"fmt"

	"github.com/fauzie/golang-sekeleton/internal/domain"
	apperrors "github.com/fauzie/golang-sekeleton/pkg/errors"
)

// UpdateUser updates every mutable column of an existing user row.
func (rw *dbReadWriter) UpdateUser(ctx context.Context, user *domain.User) error {
	ctx, cancel := context.WithTimeout(ctx, rw.queryTimeout)
	defer cancel()

	query := fmt.Sprintf(`
		UPDATE %s SET
			username = %s, email = %s, full_name = %s,
			avatar = %s, role = %s, is_active = %s, email_verified = %s,
			updated_at = NOW()
		WHERE id = %s
	`, UserTableName,
		rw.dialect.arg(1), rw.dialect.arg(2), rw.dialect.arg(3),
		rw.dialect.arg(4), rw.dialect.arg(5), rw.dialect.arg(6), rw.dialect.arg(7),
		rw.dialect.arg(8))

	result, err := rw.db.ExecContext(ctx, query,
		user.Username, user.Email, user.FullName,
		user.Avatar, user.Role, user.IsActive, user.EmailVerified,
		user.ID,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return apperrors.NewDuplicateKeyError(user.ID, "email or username already in use", err)
		}
		return apperrors.NewDataAccessError("failed to update user", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewDataAccessError("failed to read update result", err)
	}
	if rows == 0 {
		return apperrors.NewNotFoundError("User", user.ID)
	}
	return nil
}
