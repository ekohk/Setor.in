// Package http is the HTTP delivery layer for the user module.
//
// Handlers are thin: parse request → call usecase → write response.
// No business logic; no SQL; no Keycloak calls.
package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/user/application/dto"
	"github.com/setorin/setorin/backend/internal/apps/user/application/ports"
	"github.com/setorin/setorin/backend/internal/apps/user/application/usecase"
	"github.com/setorin/setorin/backend/internal/shared/middleware"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
	"github.com/setorin/setorin/backend/internal/shared/response"
)

// Handler bundles user-facing HTTP handlers backed by a single use case.
type Handler struct {
	uc *usecase.UserUseCase
}

func NewHandler(uc *usecase.UserUseCase) *Handler {
	return &Handler{uc: uc}
}

// Sync handles POST /v1/auth/sync — idempotent first-login mirror to local DB.
func (h *Handler) Sync(c *gin.Context) {
	claims := middleware.MustClaims(c)

	u, err := h.uc.SyncFromClaims(c.Request.Context(), claims)
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, dto.MeResponse{
		UserResponse: dto.FromModel(u),
		Roles:        claims.RealmAccess.Roles,
	})
}

// Me handles GET /v1/auth/me and GET /v1/users/me.
func (h *Handler) Me(c *gin.Context) {
	claims := middleware.MustClaims(c)

	kcID, err := uuid.Parse(claims.Subject)
	if err != nil {
		response.Err(c, apperr.Wrap(apperr.CodeTokenInvalid, "subject is not a valid uuid", err))
		return
	}

	u, err := h.uc.GetByKeycloakID(c.Request.Context(), kcID)
	if err != nil {
		// CodeNotFound here means /v1/auth/sync hasn't been called yet.
		// Hint the frontend with a clearer message.
		if ae, ok := apperr.As(err); ok && ae.Code == apperr.CodeNotFound {
			response.Err(c, apperr.New(apperr.CodeNotFound,
				"local user record not found; call POST /v1/auth/sync first"))
			return
		}
		response.Err(c, err)
		return
	}

	response.OK(c, dto.MeResponse{
		UserResponse: dto.FromModel(u),
		Roles:        claims.RealmAccess.Roles,
	})
}

// UpdateMe handles PATCH /v1/users/me.
func (h *Handler) UpdateMe(c *gin.Context) {
	claims := middleware.MustClaims(c)

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperr.Wrap(apperr.CodeValidation, "invalid request body", err))
		return
	}

	kcID, err := uuid.Parse(claims.Subject)
	if err != nil {
		response.Err(c, apperr.Wrap(apperr.CodeTokenInvalid, "subject is not a valid uuid", err))
		return
	}

	u, err := h.uc.UpdateProfile(c.Request.Context(), kcID, ports.ProfilePatch{
		FullName:  req.FullName,
		Phone:     req.Phone,
		AvatarURL: req.AvatarURL,
	})
	if err != nil {
		response.Err(c, err)
		return
	}

	response.OK(c, dto.FromModel(u))
}
