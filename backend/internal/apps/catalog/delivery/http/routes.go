package http

import (
	"github.com/gin-gonic/gin"

	"github.com/setorin/setorin/backend/internal/infrastructure/keycloak"
	"github.com/setorin/setorin/backend/internal/shared/middleware"
)

// RegisterRoutes mounts catalog endpoints.
//
//   GET   /v1/materials                                   (public — no token needed)
//   GET   /v1/materials/:slug                             (public)
//   GET   /v1/admin/materials                             (admin)
//   PATCH /v1/admin/materials/:id                         (admin)
//   GET   /v1/admin/materials/:id/prices                  (admin)
//   POST  /v1/admin/materials/:id/prices                  (admin)
func RegisterRoutes(v1 *gin.RouterGroup, h *Handler, jwtV *keycloak.Validator) {
	// Public — no JWT middleware.
	v1.GET("/materials", h.ListPublic)
	v1.GET("/materials/:slug", h.GetBySlug)

	// Admin — require admin role (super_admin composite-includes admin).
	adm := v1.Group("/admin",
		middleware.JWT(jwtV),
		middleware.RequireRole(middleware.RoleAdmin),
	)
	{
		adm.GET("/materials", h.ListAdmin)
		adm.PATCH("/materials/:id", h.UpdateMaterial)
		adm.GET("/materials/:id/prices", h.PriceHistory)
		adm.POST("/materials/:id/prices", h.SetPrice)
	}
}
