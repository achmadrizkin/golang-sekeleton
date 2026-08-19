package database

import (
	"context"
	"fmt"

	"github.com/fauzie/golang-sekeleton/internal/domain"
	apperrors "github.com/fauzie/golang-sekeleton/pkg/errors"
)

// SearchUsers matches query against username, email, and full_name.
func (rw *dbReadWriter) SearchUsers(ctx context.Context, query string) ([]*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, rw.queryTimeout)
	defer cancel()

	sqlQuery := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE username LIKE %s OR email LIKE %s OR full_name LIKE %s
		ORDER BY created_at DESC
	`, UserFields, UserTableName, rw.dialect.arg(1), rw.dialect.arg(2), rw.dialect.arg(3))

	like := "%" + query + "%"
	rows, err := rw.db.QueryContext(ctx, sqlQuery, like, like, like)
	if err != nil {
		return nil, apperrors.NewDataAccessError("failed to search users", err)
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
