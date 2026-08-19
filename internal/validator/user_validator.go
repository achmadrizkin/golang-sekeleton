package validator

import (
	"github.com/fauzie/golang-sekeleton/internal/domain"
	pkgvalidator "github.com/fauzie/golang-sekeleton/pkg/validator"
)

// UserValidator validates domain.User against its `validate` struct tags.
type UserValidator struct {
	*pkgvalidator.Validator
}

// NewUserValidator builds a UserValidator.
func NewUserValidator() *UserValidator {
	return &UserValidator{Validator: pkgvalidator.New()}
}

// Validate validates a User entity.
func (v *UserValidator) Validate(user *domain.User) error {
	return v.Validator.Validate(user)
}
