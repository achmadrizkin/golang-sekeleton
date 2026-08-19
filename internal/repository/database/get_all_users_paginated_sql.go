package database

import (
	"context"
	"fmt"

	"github.com/fauzie/golang-sekeleton/internal/domain"
	apperrors "github.com/fauzie/golang-sekeleton/pkg/errors"
)

// GetAllUsersPaginated retrieves one page of users plus the total row count.
func (rw *dbReadWriter) GetAllUsersPaginated(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, rw.queryTimeout)
	defer cancel()

	offset := (page - 1) * pageSize

	var totalCount int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, UserTableName)
	if err := rw.db.QueryRowContext(ctx, countQuery).Scan(&totalCount); err != nil {
		return nil, 0, apperrors.NewDataAccessError("failed to count users", err)
	}

	query := fmt.Sprintf(`
		SELECT %s FROM %s
		ORDER BY created_at DESC
		LIMIT %s OFFSET %s
	`, UserFields, UserTableName, rw.dialect.arg(1), rw.dialect.arg(2))

	rows, err := rw.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, apperrors.NewDataAccessError("failed to query users", err)
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
			return nil, 0, apperrors.NewDataAccessError("failed to scan user row", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperrors.NewDataAccessError("rows error", err)
	}
	return users, totalCount, nil
}
