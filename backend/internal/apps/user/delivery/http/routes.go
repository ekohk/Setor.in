package http

import (
	"github.com/gin-gonic/gin"

	"github.com/setorin/setorin/backend/internal/infrastructure/keycloak"
	"github.com/setorin/setorin/backend/internal/shared/middleware"
)

// RegisterRoutes mounts all user-module endpoints under the given v1 group.
//
// Why pass v1 instead of root: keeps URL prefix decisions in main.go and
// lets us wire other modules (auth, order, wallet, ...) under the same group.
func RegisterRoutes(v1 *gin.RouterGroup, h *Handler, jwtV *keycloak.Validator) {
	authed := v1.Group("", middleware.JWT(jwtV))

	// /v1/auth/* — auth-flow adjacent endpoints owned by user module
	auth := authed.Group("/auth")
	{
		auth.POST("/sync", h.Sync)
		auth.GET("/me", h.Me)
	}

	// /v1/users/me — current user's own resource
	users := authed.Group("/users")
	{
		users.GET("/me", h.Me)
		users.PATCH("/me", h.UpdateMe)
	}
}
