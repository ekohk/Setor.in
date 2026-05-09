// Package errors defines the application's error taxonomy.
//
// Each AppError carries a stable Code (used for client-facing API responses
// and logs) plus a human-readable Message. HTTP status is derived from Code.
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type Code string

const (
	CodeUnauthorized      Code = "UNAUTHORIZED"
	CodeForbidden         Code = "FORBIDDEN"
	CodeInsufficientRole  Code = "INSUFFICIENT_ROLE"
	CodeNotFound          Code = "NOT_FOUND"
	CodeConflict          Code = "CONFLICT"
	CodeValidation        Code = "VALIDATION_ERROR"
	CodeBadRequest        Code = "BAD_REQUEST"
	CodeTokenExpired      Code = "TOKEN_EXPIRED"
	CodeTokenInvalid      Code = "TOKEN_INVALID"
	CodeAccountSuspended  Code = "ACCOUNT_SUSPENDED"
	CodeAccountDeleted    Code = "ACCOUNT_DELETED"
	CodeRateLimited       Code = "RATE_LIMITED"
	CodeInternal          Code = "INTERNAL_ERROR"
	CodeUpstreamUnavail   Code = "UPSTREAM_UNAVAILABLE"
)

// AppError is the canonical error type returned from application/usecase
// layers. Handlers convert it into an HTTP response via response.Envelope.
type AppError struct {
	Code    Code
	Message string
	Details map[string]any
	cause   error
}

func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.cause }

// HTTPStatus maps Code to HTTP status code.
func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case CodeUnauthorized, CodeTokenExpired, CodeTokenInvalid:
		return http.StatusUnauthorized
	case CodeForbidden, CodeInsufficientRole, CodeAccountSuspended, CodeAccountDeleted:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeValidation, CodeBadRequest:
		return http.StatusBadRequest
	case CodeRateLimited:
		return http.StatusTooManyRequests
	case CodeUpstreamUnavail:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

// New constructs an AppError with no cause.
func New(code Code, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}

// Wrap constructs an AppError preserving the underlying cause.
func Wrap(code Code, msg string, cause error) *AppError {
	return &AppError{Code: code, Message: msg, cause: cause}
}

// WithDetails adds a details map (returned in the API response).
func (e *AppError) WithDetails(d map[string]any) *AppError {
	e.Details = d
	return e
}

// As is a thin wrapper to extract an *AppError from an error chain.
func As(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
