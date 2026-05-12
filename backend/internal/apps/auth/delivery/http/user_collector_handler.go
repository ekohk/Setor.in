package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/auth/application/dto"
	"github.com/setorin/setorin/backend/internal/apps/auth/application/usecase"
	"github.com/setorin/setorin/backend/internal/shared/middleware"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
	"github.com/setorin/setorin/backend/internal/shared/response"
)

// UserCollectorHandler holds the user-facing endpoint for submitting an
// application. Lives in auth module because the action gates a role change.
type UserCollectorHandler struct {
	uc *usecase.CollectorApplicationUseCase
}

func NewUserCollectorHandler(uc *usecase.CollectorApplicationUseCase) *UserCollectorHandler {
	return &UserCollectorHandler{uc: uc}
}

// Submit handles POST /v1/users/me/become-collector.
func (h *UserCollectorHandler) Submit(c *gin.Context) {
	claims := middleware.MustClaims(c)

	kcID, err := uuid.Parse(claims.Subject)
	if err != nil {
		response.Err(c, apperr.Wrap(apperr.CodeTokenInvalid, "subject not a uuid", err))
		return
	}

	var req dto.BecomeCollectorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperr.Wrap(apperr.CodeValidation, "invalid request body", err))
		return
	}

	app, err := h.uc.Submit(c.Request.Context(),
		kcID,
		claims.RealmAccess.Roles,
		req.BusinessName, req.LicenseNo, req.KTPURL, req.SIUPURL, req.Address,
		c.ClientIP(), c.Request.UserAgent(),
	)
	if err != nil {
		response.Err(c, err)
		return
	}

	c.JSON(201, response.Envelope{Data: dto.FromApplicationModel(app)})
}
