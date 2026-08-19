// Package errors provides a small taxonomy of typed application errors.
//
// Every error created here carries a stable Code. That code is the single
// source of truth used by two independent decisions elsewhere in the code
// base: pkg/middleware maps it to a gRPC/HTTP status, and IsRetryable uses it
// to decide whether a message consumer should ACK or NACK. Wrap terminal
// failures with the right constructor below instead of fmt.Errorf, otherwise
// they will be treated as retryable by default.
package errors

import (
	"errors"
	"fmt"
)

// Code identifies the category of an AppError.
type Code string

const (
	CodeValidation   Code = "VALIDATION_ERROR"
	CodeNotFound     Code = "NOT_FOUND"
	CodeDuplicateKey Code = "DUPLICATE_KEY"
	CodeDataAccess   Code = "DATA_ACCESS_ERROR"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeForbidden    Code = "FORBIDDEN"
	CodeConflict     Code = "CONFLICT"
	CodeInternal     Code = "INTERNAL_ERROR"
)

// AppError is the concrete error type produced by every constructor in this
// package. It deliberately keeps Resource/ID/Message separate from the
// wrapped cause so callers (logging, mapping to gRPC status) can pick
// whichever level of detail they need.
type AppError struct {
	Code     Code
	Message  string
	Resource string
	Cause    error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.Cause }

func newErr(code Code, message string, cause error) *AppError {
	return &AppError{Code: code, Message: message, Cause: cause}
}

// NewValidationError reports invalid input. Not retryable.
func NewValidationError(message string, cause error) *AppError {
	return newErr(CodeValidation, message, cause)
}

// NewNotFoundError reports a missing resource (DB row or cache miss). Not retryable.
func NewNotFoundError(resource, id string) *AppError {
	return &AppError{
		Code:     CodeNotFound,
		Message:  fmt.Sprintf("%s with ID '%s' not found", resource, id),
		Resource: resource,
	}
}

// NewDuplicateKeyError reports a unique-constraint violation. Not retryable.
func NewDuplicateKeyError(id, message string, cause error) *AppError {
	return &AppError{Code: CodeDuplicateKey, Message: message, Resource: id, Cause: cause}
}

// NewDataAccessError reports a query/marshal/publish failure. Not retryable
// by design: a bad query or malformed payload will not succeed on retry.
func NewDataAccessError(message string, cause error) *AppError {
	return newErr(CodeDataAccess, message, cause)
}

// NewUnauthorizedError reports a missing/invalid credential. Not retryable.
func NewUnauthorizedError(message string) *AppError {
	return newErr(CodeUnauthorized, message, nil)
}

// NewForbiddenError reports insufficient permission. Not retryable.
func NewForbiddenError(message string) *AppError {
	return newErr(CodeForbidden, message, nil)
}

// NewConflictError reports a business-rule conflict. Not retryable.
func NewConflictError(message string) *AppError {
	return newErr(CodeConflict, message, nil)
}

// NewInternalError reports an unexpected failure. Retryable.
func NewInternalError(message string, cause error) *AppError {
	return newErr(CodeInternal, message, cause)
}

// IsRetryable decides whether a message consumer should NACK (retry) or ACK
// (drop) after a failure. It is a deny-list: everything except the
// terminal codes above is considered retryable, so an error that escaped
// this package (e.g. a bare fmt.Errorf, a network blip) is retried by
// default. Wrap errors that are genuinely final with the matching
// constructor so they are not retried forever.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case CodeValidation, CodeNotFound, CodeDuplicateKey, CodeDataAccess,
			CodeUnauthorized, CodeForbidden, CodeConflict:
			return false
		case CodeInternal:
			return true
		}
	}
	return true
}

// CodeOf extracts the Code carried by err, or "" if err is not an AppError.
func CodeOf(err error) Code {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ""
}
