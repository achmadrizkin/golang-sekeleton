package util

import "github.com/google/uuid"

// NewUUID generates a random (v4) UUID string. Centralized here so every
// layer that needs an ID generates it the same way.
func NewUUID() string {
	return uuid.NewString()
}
