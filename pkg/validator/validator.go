// Package validator wraps go-playground/validator so entity-specific
// validators (internal/validator) only need to embed *Validator and expose a
// typed Validate method.
package validator

import (
	"fmt"
	"strings"

	playground "github.com/go-playground/validator/v10"

	apperrors "github.com/fauzie/golang-sekeleton/pkg/errors"
)

// Validator evaluates the `validate` struct tag on domain entities.
type Validator struct {
	v *playground.Validate
}

// New builds a Validator with struct-level (not just field-level) validation
// enabled.
func New() *Validator {
	return &Validator{v: playground.New()}
}

// Validate runs struct tag validation and, on failure, returns a single
// *apperrors.AppError (CodeValidation) describing every failing field so
// callers only need to handle one error type.
func (val *Validator) Validate(s interface{}) error {
	if err := val.v.Struct(s); err != nil {
		var invalid *playground.InvalidValidationError
		if ok := isInvalidValidationError(err, &invalid); ok {
			return apperrors.NewValidationError("invalid value passed to validator", err)
		}

		var msgs []string
		for _, fe := range err.(playground.ValidationErrors) {
			msgs = append(msgs, fmt.Sprintf("%s failed on '%s'", fe.Field(), fe.Tag()))
		}
		return apperrors.NewValidationError(strings.Join(msgs, "; "), err)
	}
	return nil
}

func isInvalidValidationError(err error, target **playground.InvalidValidationError) bool {
	if e, ok := err.(*playground.InvalidValidationError); ok {
		*target = e
		return true
	}
	return false
}
