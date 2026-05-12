// Package http exposes catalog endpoints under /v1/materials and /v1/admin/materials.
package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/catalog/application/dto"
	"github.com/setorin/setorin/backend/internal/apps/catalog/application/ports"
	"github.com/setorin/setorin/backend/internal/apps/catalog/application/usecase"
	userports "github.com/setorin/setorin/backend/internal/apps/user/application/ports"
	"github.com/setorin/setorin/backend/internal/shared/middleware"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
	"github.com/setorin/setorin/backend/internal/shared/response"
)

type Handler struct {
	uc       *usecase.MaterialUseCase
	userRepo userports.UserRepository // for actor_id lookup on POST /prices
}

func NewHandler(uc *usecase.MaterialUseCase, userRepo userports.UserRepository) *Handler {
	return &Handler{uc: uc, userRepo: userRepo}
}

// ─── Public ─────

// GET /v1/materials — list active materials with current price.
func (h *Handler) ListPublic(c *gin.Context) {
	rows, err := h.uc.ListPublic(c.Request.Context())
	if err != nil {
		response.Err(c, err)
		return
	}
	out := make([]dto.MaterialResponse, 0, len(rows))
	for i := range rows {
		out = append(out, dto.FromMaterialWithPrice(&rows[i]))
	}
	response.OK(c, out)
}

// GET /v1/materials/:slug
func (h *Handler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	m, err := h.uc.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, dto.FromMaterialWithPrice(m))
}

// ─── Admin ─────

// GET /v1/admin/materials — list all including inactive.
func (h *Handler) ListAdmin(c *gin.Context) {
	rows, err := h.uc.ListAdmin(c.Request.Context())
	if err != nil {
		response.Err(c, err)
		return
	}
	out := make([]dto.MaterialResponse, 0, len(rows))
	for i := range rows {
		out = append(out, dto.FromMaterialWithPrice(&rows[i]))
	}
	response.OK(c, out)
}

// PATCH /v1/admin/materials/:id
func (h *Handler) UpdateMaterial(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, apperr.New(apperr.CodeValidation, "invalid material id"))
		return
	}
	var req dto.UpdateMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperr.Wrap(apperr.CodeValidation, "invalid request body", err))
		return
	}
	m, err := h.uc.UpdateMaterial(c.Request.Context(), id, ports.MaterialPatch{
		Name:        req.Name,
		Icon:        req.Icon,
		Description: req.Description,
		IsActive:    req.IsActive,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		response.Err(c, err)
		return
	}
	// Re-fetch to include current price in response.
	withPrice, err := h.uc.GetByID(c.Request.Context(), m.ID)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, dto.FromMaterialWithPrice(withPrice))
}

// POST /v1/admin/materials/:id/prices
func (h *Handler) SetPrice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, apperr.New(apperr.CodeValidation, "invalid material id"))
		return
	}
	var req dto.SetPriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperr.Wrap(apperr.CodeValidation, "invalid request body", err))
		return
	}

	// Resolve actor's users.id from JWT subject so created_by FK is satisfied.
	actorID, err := h.actorIDFromCtx(c)
	if err != nil {
		response.Err(c, err)
		return
	}

	p, err := h.uc.SetPrice(c.Request.Context(), id, req.PricePerKg, &actorID, req.Note)
	if err != nil {
		response.Err(c, err)
		return
	}
	c.JSON(201, response.Envelope{Data: dto.FromMaterialPrice(p)})
}

// GET /v1/admin/materials/:id/prices?limit=50
func (h *Handler) PriceHistory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, apperr.New(apperr.CodeValidation, "invalid material id"))
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	list, err := h.uc.PriceHistory(c.Request.Context(), id, limit)
	if err != nil {
		response.Err(c, err)
		return
	}
	out := make([]dto.MaterialPriceResponse, 0, len(list))
	for i := range list {
		out = append(out, dto.FromMaterialPrice(&list[i]))
	}
	response.OK(c, out)
}

// actorIDFromCtx resolves the local users.id of the admin making the call.
// Mirrors the pattern from auth module's admin_handler.
func (h *Handler) actorIDFromCtx(c *gin.Context) (uuid.UUID, error) {
	claims := middleware.MustClaims(c)
	kcID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, apperr.Wrap(apperr.CodeTokenInvalid, "subject not a uuid", err)
	}
	u, err := h.userRepo.FindByKeycloakID(c.Request.Context(), kcID)
	if err != nil {
		if ae, ok := apperr.As(err); ok && ae.Code == apperr.CodeNotFound {
			return uuid.Nil, apperr.New(apperr.CodeForbidden,
				"admin local profile not synced; call POST /v1/auth/sync first")
		}
		return uuid.Nil, err
	}
	return u.ID, nil
}
