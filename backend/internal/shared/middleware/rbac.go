package middleware

import (
	"github.com/gin-gonic/gin"

	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
	"github.com/setorin/setorin/backend/internal/shared/response"
)

// Setor.in role constants — kept here so call sites avoid stringly-typed args.
const (
	RoleUser        = "user"
	RoleCollector   = "collector"
	RoleCV          = "cv"
	RoleAdmin       = "admin"
	RoleSuperAdmin  = "super_admin"
)

// RequireRole denies the request unless the token has *at least one* of roles.
// Use after JWT(). Order doesn't matter; composite role expansion is already
// done by Keycloak at token issue time.
//
// Example:
//   admin := r.Group("/v1/admin", middleware.JWT(v), middleware.RequireRole("admin"))
func RequireRole(roles ...string) gin.HandlerFunc {
	if len(roles) == 0 {
		// Misuse: a route protected by zero roles is more dangerous than no
		// protection (looks intentional but allows everyone through). Fail loudly.
		panic("RequireRole called with no roles")
	}
	return func(c *gin.Context) {
		claims := MustClaims(c)
		if !claims.HasAnyRole(roles...) {
			response.Err(c, apperr.New(
				apperr.CodeInsufficientRole,
				"insufficient role for this resource",
			).WithDetails(map[string]any{
				"required_any_of": roles,
				"granted":         claims.RealmAccess.Roles,
			}))
			c.Abort()
			return
		}
		c.Next()
	}
}
