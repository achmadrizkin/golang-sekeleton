package database

import (
	"context"
	"fmt"

	apperrors "github.com/fauzie/golang-sekeleton/pkg/errors"
)

// DeleteUser removes a user row by ID.
func (rw *dbReadWriter) DeleteUser(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, rw.queryTimeout)
	defer cancel()

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = %s`, UserTableName, rw.dialect.arg(1))

	result, err := rw.db.ExecContext(ctx, query, id)
	if err != nil {
		return apperrors.NewDataAccessError("failed to delete user", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.NewDataAccessError("failed to read delete result", err)
	}
	if rows == 0 {
		return apperrors.NewNotFoundError("User", id)
	}
	return nil
}
