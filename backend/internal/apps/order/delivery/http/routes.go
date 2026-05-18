package http

import (
	"github.com/gin-gonic/gin"

	"github.com/setorin/setorin/backend/internal/infrastructure/keycloak"
	"github.com/setorin/setorin/backend/internal/shared/middleware"
)

// RegisterRoutes mounts all order endpoints under the /v1 group.
//
// User-side (any authenticated user):
//   POST /v1/orders                        — create
//   GET  /v1/orders/me                     — list my orders
//   GET  /v1/orders/:code                  — detail (RBAC inside)
//   POST /v1/orders/:code/cancel           — cancel (only if status=received)
//   POST /v1/orders/:code/confirm-cash     — confirm cash → done
//
// Collector-side (require `collector` role):
//   GET  /v1/collector/orders/incoming
//   GET  /v1/collector/orders/me
//   GET  /v1/collector/orders/:code
//   POST /v1/collector/orders/:code/accept
//   POST /v1/collector/orders/:code/start-pickup
//   POST /v1/collector/orders/:code/arrive
//   POST /v1/collector/orders/:code/verify-otp
//   POST /v1/collector/orders/:code/weigh
//   POST /v1/collector/orders/:code/quality
func RegisterRoutes(
	v1 *gin.RouterGroup,
	userH *UserHandler,
	collectorH *CollectorHandler,
	jwtV *keycloak.Validator,
) {
	authed := v1.Group("", middleware.JWT(jwtV))

	// Seller side
	authed.POST("/orders", userH.Create)
	authed.GET("/orders/me", userH.ListMine)
	authed.GET("/orders/:code", userH.GetByCode)
	authed.POST("/orders/:code/cancel", userH.Cancel)
	authed.POST("/orders/:code/confirm-cash", userH.ConfirmCash)

	// Collector side — additional role check
	col := authed.Group("/collector", middleware.RequireRole(middleware.RoleCollector))
	{
		col.GET("/orders/incoming", collectorH.ListIncoming)
		col.GET("/orders/me", collectorH.ListMine)
		col.GET("/orders/:code", collectorH.GetByCode)
		col.POST("/orders/:code/accept", collectorH.Accept)
		col.POST("/orders/:code/start-pickup", collectorH.StartPickup)
		col.POST("/orders/:code/arrive", collectorH.Arrive)
		col.POST("/orders/:code/verify-otp", collectorH.VerifyOTP)
		col.POST("/orders/:code/weigh", collectorH.Weigh)
		col.POST("/orders/:code/quality", collectorH.Quality)
	}
}
