package http

import (
	"github.com/gin-gonic/gin"

	"github.com/setorin/setorin/backend/internal/infrastructure/keycloak"
	"github.com/setorin/setorin/backend/internal/shared/middleware"
)

// RegisterRoutes mounts auth-module routes under v1.
//
//   POST  /v1/users/me/become-collector       (any authenticated user)
//   GET   /v1/admin/users                     (admin or super_admin)
//   PATCH /v1/admin/users/:id/status          (admin or super_admin)
//   PATCH /v1/admin/users/:id/roles           (super_admin only)
//   GET   /v1/admin/collector-applications    (admin)
//   POST  /v1/admin/collector-applications/:id/approve   (admin)
//   POST  /v1/admin/collector-applications/:id/reject    (admin)
func RegisterRoutes(
	v1 *gin.RouterGroup,
	admin *AdminHandler,
	userColl *UserCollectorHandler,
	jwtV *keycloak.Validator,
) {
	authed := v1.Group("", middleware.JWT(jwtV))

	// User-facing
	authed.POST("/users/me/become-collector", userColl.Submit)

	// Admin: read + status updates require admin (or super_admin via composite)
	adm := authed.Group("/admin", middleware.RequireRole(middleware.RoleAdmin))
	{
		adm.GET("/users", admin.ListUsers)
		adm.PATCH("/users/:id/status", admin.UpdateUserStatus)

		adm.GET("/collector-applications", admin.ListApplications)
		adm.POST("/collector-applications/:id/approve", admin.ApproveApplication)
		adm.POST("/collector-applications/:id/reject", admin.RejectApplication)
	}

	// super_admin-only: role mutations
	sa := authed.Group("/admin", middleware.RequireRole(middleware.RoleSuperAdmin))
	{
		sa.PATCH("/users/:id/roles", admin.UpdateUserRole)
	}
}
