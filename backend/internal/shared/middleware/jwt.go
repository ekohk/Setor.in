// Package middleware contains cross-cutting Gin middleware used by all modules.
package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/setorin/setorin/backend/internal/infrastructure/keycloak"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
	"github.com/setorin/setorin/backend/internal/shared/response"
)

// Context keys for values set by the JWT middleware. Keep package-private and
// expose getters so call sites don't reach into the gin.Context with magic strings.
const (
	ctxKeyClaims = "auth_claims"
)

// JWT returns a Gin middleware that validates the Bearer token and attaches
// the parsed Claims to the request context. Aborts with 401 on any failure.
func JWT(v *keycloak.Validator) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := extractBearer(c.GetHeader("Authorization"))

		claims, err := v.Validate(raw)
		if err != nil {
			respondAuthError(c, err)
			return
		}

		c.Set(ctxKeyClaims, claims)
		c.Next()
	}
}

// MustClaims fetches the claims previously stored by JWT().
// Panics if missing — callers are protected handlers behind JWT().
func MustClaims(c *gin.Context) *keycloak.Claims {
	v, ok := c.Get(ctxKeyClaims)
	if !ok {
		panic("MustClaims called outside JWT middleware")
	}
	return v.(*keycloak.Claims)
}

// Claims is a non-panicking variant for places where presence is uncertain.
func Claims(c *gin.Context) (*keycloak.Claims, bool) {
	v, ok := c.Get(ctxKeyClaims)
	if !ok {
		return nil, false
	}
	cl, ok := v.(*keycloak.Claims)
	return cl, ok
}

func extractBearer(h string) string {
	const prefix = "Bearer "
	h = strings.TrimSpace(h)
	if len(h) <= len(prefix) {
		return ""
	}
	if !strings.EqualFold(h[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(h[len(prefix):])
}

func respondAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, keycloak.ErrTokenMissing):
		response.Err(c, apperr.New(apperr.CodeUnauthorized, "missing or malformed Authorization header"))
	case errors.Is(err, keycloak.ErrTokenExpired):
		response.Err(c, apperr.New(apperr.CodeTokenExpired, "token has expired"))
	case errors.Is(err, keycloak.ErrTokenInvalid):
		response.Err(c, apperr.New(apperr.CodeTokenInvalid, "token is invalid"))
	default:
		response.Err(c, apperr.Wrap(apperr.CodeUnauthorized, "authentication failed", err))
	}
	c.Abort()
}
