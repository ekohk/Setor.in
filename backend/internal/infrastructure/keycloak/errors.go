package keycloak

import "errors"

// Sentinel errors returned by Validator and AdminClient. Callers compare
// with errors.Is() and translate to apperr.AppError at the handler boundary.
var (
	ErrTokenMissing  = errors.New("keycloak: token missing")
	ErrTokenInvalid  = errors.New("keycloak: token invalid")
	ErrTokenExpired  = errors.New("keycloak: token expired")
	ErrUserNotFound  = errors.New("keycloak: user not found")
	ErrUpstream      = errors.New("keycloak: upstream error")
)
