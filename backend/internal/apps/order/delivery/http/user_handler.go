// Package http is the HTTP delivery for the order module.
package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/order/application/dto"
	"github.com/setorin/setorin/backend/internal/apps/order/application/usecase"
	userports "github.com/setorin/setorin/backend/internal/apps/user/application/ports"
	"github.com/setorin/setorin/backend/internal/shared/middleware"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
	"github.com/setorin/setorin/backend/internal/shared/response"
)

// UserHandler — endpoints for the seller side of an order.
type UserHandler struct {
	uc       *usecase.UserOrderUseCase
	userRepo userports.UserRepository // resolves keycloak_id → users.id
}

func NewUserHandler(uc *usecase.UserOrderUseCase, userRepo userports.UserRepository) *UserHandler {
	return &UserHandler{uc: uc, userRepo: userRepo}
}

// POST /v1/orders
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperr.Wrap(apperr.CodeValidation, "invalid request body", err))
		return
	}
	sellerID, err := h.localUserID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	o, err := h.uc.Create(c.Request.Context(), sellerID, req)
	if err != nil {
		response.Err(c, err)
		return
	}
	c.JSON(201, response.Envelope{Data: dto.FromModel(o, sellerID, false)})
}

// GET /v1/orders/me
func (h *UserHandler) ListMine(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	sellerID, err := h.localUserID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	list, total, err := h.uc.ListMine(c.Request.Context(), sellerID, page, pageSize)
	if err != nil {
		response.Err(c, err)
		return
	}
	out := make([]dto.OrderResponse, 0, len(list))
	for i := range list {
		out = append(out, dto.FromModel(&list[i], sellerID, false))
	}
	response.Paginated(c, out, response.Pagination{Page: page, PageSize: pageSize, Total: total})
}

// GET /v1/orders/:code
func (h *UserHandler) GetByCode(c *gin.Context) {
	code := c.Param("code")
	viewerID, err := h.localUserID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	isAdmin := claimsHasRole(c, middleware.RoleAdmin)
	o, err := h.uc.GetByCode(c.Request.Context(), viewerID, code, isAdmin)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, dto.FromModel(o, viewerID, isAdmin))
}

// POST /v1/orders/:code/cancel
func (h *UserHandler) Cancel(c *gin.Context) {
	code := c.Param("code")
	var req dto.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperr.Wrap(apperr.CodeValidation, "invalid request body", err))
		return
	}
	sellerID, err := h.localUserID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	o, err := h.uc.Cancel(c.Request.Context(), sellerID, code, req.Reason)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, dto.FromModel(o, sellerID, false))
}

// POST /v1/orders/:code/confirm-cash
func (h *UserHandler) ConfirmCash(c *gin.Context) {
	code := c.Param("code")
	sellerID, err := h.localUserID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	o, err := h.uc.ConfirmCash(c.Request.Context(), sellerID, code)
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, dto.FromModel(o, sellerID, false))
}

// ─── helpers ────────────────────────────────────────────────────────────

func (h *UserHandler) localUserID(c *gin.Context) (uuid.UUID, error) {
	claims := middleware.MustClaims(c)
	kcID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, apperr.Wrap(apperr.CodeTokenInvalid, "subject not a uuid", err)
	}
	u, err := h.userRepo.FindByKeycloakID(c.Request.Context(), kcID)
	if err != nil {
		if ae, ok := apperr.As(err); ok && ae.Code == apperr.CodeNotFound {
			return uuid.Nil, apperr.New(apperr.CodeForbidden,
				"local profile not synced; call POST /v1/auth/sync first")
		}
		return uuid.Nil, err
	}
	return u.ID, nil
}

func claimsHasRole(c *gin.Context, role string) bool {
	claims := middleware.MustClaims(c)
	for _, r := range claims.RealmAccess.Roles {
		if r == role || r == middleware.RoleSuperAdmin {
			return true
		}
	}
	return false
}
