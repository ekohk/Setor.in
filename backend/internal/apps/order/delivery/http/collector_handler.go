package http

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/order/application/dto"
	"github.com/setorin/setorin/backend/internal/apps/order/application/usecase"
	"github.com/setorin/setorin/backend/internal/apps/order/domain/model"
	userports "github.com/setorin/setorin/backend/internal/apps/user/application/ports"
	"github.com/setorin/setorin/backend/internal/shared/middleware"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
	"github.com/setorin/setorin/backend/internal/shared/response"
)

// CollectorHandler — endpoints for the collector side.
type CollectorHandler struct {
	uc       *usecase.CollectorOrderUseCase
	userRepo userports.UserRepository
}

func NewCollectorHandler(uc *usecase.CollectorOrderUseCase, userRepo userports.UserRepository) *CollectorHandler {
	return &CollectorHandler{uc: uc, userRepo: userRepo}
}

// GET /v1/collector/orders/incoming
func (h *CollectorHandler) ListIncoming(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	collectorID, err := h.localUserID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	list, total, err := h.uc.ListIncoming(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Err(c, err)
		return
	}
	out := make([]dto.OrderResponse, 0, len(list))
	for i := range list {
		out = append(out, dto.FromModel(&list[i], collectorID, false))
	}
	response.Paginated(c, out, response.Pagination{Page: page, PageSize: pageSize, Total: total})
}

// GET /v1/collector/orders/me
func (h *CollectorHandler) ListMine(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	collectorID, err := h.localUserID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	list, total, err := h.uc.ListMine(c.Request.Context(), collectorID, page, pageSize)
	if err != nil {
		response.Err(c, err)
		return
	}
	out := make([]dto.OrderResponse, 0, len(list))
	for i := range list {
		out = append(out, dto.FromModel(&list[i], collectorID, false))
	}
	response.Paginated(c, out, response.Pagination{Page: page, PageSize: pageSize, Total: total})
}

// GET /v1/collector/orders/:code
func (h *CollectorHandler) GetByCode(c *gin.Context) {
	cid, err := h.localUserID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	o, err := h.uc.GetByCode(c.Request.Context(), cid, c.Param("code"))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, dto.FromModel(o, cid, false))
}

// POST /v1/collector/orders/:code/accept
func (h *CollectorHandler) Accept(c *gin.Context) {
	h.runMutate(c, func(ctx context.Context, cid uuid.UUID, code string) (*model.Order, error) {
		return h.uc.Accept(ctx, cid, code)
	})
}

// POST /v1/collector/orders/:code/start-pickup
func (h *CollectorHandler) StartPickup(c *gin.Context) {
	h.runMutate(c, func(ctx context.Context, cid uuid.UUID, code string) (*model.Order, error) {
		return h.uc.StartPickup(ctx, cid, code)
	})
}

// POST /v1/collector/orders/:code/arrive
func (h *CollectorHandler) Arrive(c *gin.Context) {
	h.runMutate(c, func(ctx context.Context, cid uuid.UUID, code string) (*model.Order, error) {
		return h.uc.Arrive(ctx, cid, code)
	})
}

// POST /v1/collector/orders/:code/verify-otp
func (h *CollectorHandler) VerifyOTP(c *gin.Context) {
	var req dto.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperr.Wrap(apperr.CodeValidation, "invalid request body", err))
		return
	}
	h.runMutate(c, func(ctx context.Context, cid uuid.UUID, code string) (*model.Order, error) {
		return h.uc.VerifyOTP(ctx, cid, code, req)
	})
}

// POST /v1/collector/orders/:code/weigh
func (h *CollectorHandler) Weigh(c *gin.Context) {
	var req dto.WeighRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperr.Wrap(apperr.CodeValidation, "invalid request body", err))
		return
	}
	h.runMutate(c, func(ctx context.Context, cid uuid.UUID, code string) (*model.Order, error) {
		return h.uc.Weigh(ctx, cid, code, req)
	})
}

// POST /v1/collector/orders/:code/quality
func (h *CollectorHandler) Quality(c *gin.Context) {
	var req dto.QualityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperr.Wrap(apperr.CodeValidation, "invalid request body", err))
		return
	}
	h.runMutate(c, func(ctx context.Context, cid uuid.UUID, code string) (*model.Order, error) {
		return h.uc.Quality(ctx, cid, code, req)
	})
}

// ─── helpers ────────────────────────────────────────────────────────────

// runMutate is shared boilerplate for collector POST endpoints:
// 1) resolve collector's local id, 2) call usecase fn, 3) write response.
func (h *CollectorHandler) runMutate(
	c *gin.Context,
	fn func(ctx context.Context, collectorID uuid.UUID, code string) (*model.Order, error),
) {
	cid, err := h.localUserID(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	o, err := fn(c.Request.Context(), cid, c.Param("code"))
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, dto.FromModel(o, cid, false))
}

func (h *CollectorHandler) localUserID(c *gin.Context) (uuid.UUID, error) {
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
