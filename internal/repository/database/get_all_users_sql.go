package database

import (
	"context"
	"fmt"

	"github.com/fauzie/golang-sekeleton/internal/domain"
	apperrors "github.com/fauzie/golang-sekeleton/pkg/errors"
)

// GetAllUsers retrieves every user, ordered newest first.
func (rw *dbReadWriter) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, rw.queryTimeout)
	defer cancel()

	query := fmt.Sprintf(`SELECT %s FROM %s ORDER BY created_at DESC`, UserFields, UserTableName)

	rows, err := rw.db.QueryContext(ctx, query)
	if err != nil {
		return nil, apperrors.NewDataAccessError("failed to query users", err)
	}
	defer func() { _ = rows.Close() }()

	users := make([]*domain.User, 0)
	for rows.Next() {
		user := &domain.User{}
		if err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Password, &user.FullName,
			&user.Avatar, &user.Role, &user.IsActive, &user.EmailVerified,
			&user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, apperrors.NewDataAccessError("failed to scan user row", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.NewDataAccessError("rows error", err)
	}
	return users, nil
}
