package domain

import "time"

// UserCreatedPayload is the event body published to the "user_created"
// topic after a user is successfully added.
type UserCreatedPayload struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      *string   `json:"role,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// UserDeletedPayload is the event body published to the "user_deleted"
// topic after a user is successfully deleted.
type UserDeletedPayload struct {
	ID        string    `json:"id"`
	DeletedAt time.Time `json:"deleted_at"`
}
