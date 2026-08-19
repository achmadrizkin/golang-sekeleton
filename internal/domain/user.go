package domain

import "time"

// User represents a user entity. Tag `db` maps to SQL columns; `validate`
// is the input contract read by internal/validator. Nullable columns use
// pointers so an absent value can be told apart from an empty string.
type User struct {
	ID            string    `db:"id" json:"id" validate:"omitempty,uuid4"`
	Username      string    `db:"username" json:"username" validate:"required,min=3,max=50"`
	Email         string    `db:"email" json:"email" validate:"required,email,max=100"`
	Password      string    `db:"password" json:"password,omitempty" validate:"required,min=8,max=255"`
	FullName      string    `db:"full_name" json:"full_name" validate:"required,min=1,max=100"`
	Avatar        *string   `db:"avatar" json:"avatar,omitempty" validate:"omitempty,url,max=255"`
	Role          *string   `db:"role" json:"role,omitempty" validate:"omitempty,oneof=admin user"`
	IsActive      bool      `db:"is_active" json:"is_active"`
	EmailVerified bool      `db:"email_verified" json:"email_verified"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

// PaginatedUsers is a page of users plus pagination metadata.
type PaginatedUsers struct {
	Items      []*User         `json:"items"`
	Pagination *PaginationMeta `json:"pagination"`
}
