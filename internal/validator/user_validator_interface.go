package validator

import "github.com/fauzie/golang-sekeleton/internal/domain"

// UserValidatorInterface lets usecase.UserUseCase depend on validation
// behaviour without depending on go-playground/validator directly.
type UserValidatorInterface interface {
	Validate(user *domain.User) error
}
