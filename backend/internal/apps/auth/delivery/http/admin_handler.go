// Package http is the HTTP delivery for admin + collector-application endpoints.
package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/auth/application/dto"
	"github.com/setorin/setorin/backend/internal/apps/auth/application/ports"
	"github.com/setorin/setorin/backend/internal/apps/auth/application/usecase"
	"github.com/setorin/setorin/backend/internal/apps/auth/domain/model"
	userports "github.com/setorin/setorin/backend/internal/apps/user/application/ports"
	usermodel "github.com/setorin/setorin/backend/internal/apps/user/domain/model"
	"github.com/setorin/setorin/backend/internal/shared/middleware"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
	"github.com/setorin/setorin/backend/internal/shared/response"
)

// AdminHandler covers /v1/admin/* endpoints.
type AdminHandler struct {
	users    *usecase.AdminUserUseCase
	apps     *usecase.CollectorApplicationUseCase
	userRepo userports.UserRepository
}

func NewAdminHandler(
	users *usecase.AdminUserUseCase,
	apps *usecase.CollectorApplicationUseCase,
	userRepo userports.UserRepository,
) *AdminHandler {
	return &AdminHandler{users: users, apps: apps, userRepo: userRepo}
}

// ───── Users management ─────

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	f := ports.UserListFilter{
		Page:     page,
		PageSize: pageSize,
		Query:    c.Query("q"),
	}

	if r := c.Query("role"); r != "" {
		role := usermodel.Role(r)
		if !role.IsValid() {
			response.Err(c, apperr.New(apperr.CodeValidation, "invalid role filter"))
			return
		}
		f.Role = &role
	}
	if s := c.Query("status"); s != "" {
		status := usermodel.Status(s)
		if !status.IsValid() {
			response.Err(c, apperr.New(apperr.CodeValidation, "invalid status filter"))
			return
		}
		f.Status = &status
	}

	list, total, err := h.users.List(c.Request.Context(), f)
	if err != nil {
		response.Err(c, err)
		return
	}

	out := make([]dto.AdminUserResponse, 0, len(list))
	for i := range list {
		out = append(out, dto.FromUserModel(&list[i]))
	}
	response.Paginated(c, out, response.Pagination{
		Page: f.Page, PageSize: f.PageSize, Total: total,
	})
}

func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, apperr.New(apperr.CodeValidation, "invalid user id"))
		return
	}
	var req dto.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperr.Wrap(apperr.CodeValidation, "invalid request body", err))
		return
	}

	actorID, err := h.actorIDFromCtx(c)
	if err != nil {
		response.Err(c, err)
		return
	}

	updated, err := h.users.UpdateStatus(c.Request.Context(), id, req.Status, req.Reason,
		actorID, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, dto.FromUserModel(updated))
}

func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, apperr.New(apperr.CodeValidation, "invalid user id"))
		return
	}
	var req dto.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperr.Wrap(apperr.CodeValidation, "invalid request body", err))
		return
	}

	actorID, err := h.actorIDFromCtx(c)
	if err != nil {
		response.Err(c, err)
		return
	}

	updated, err := h.users.UpdateRole(c.Request.Context(), id, req.Role, req.Reason,
		actorID, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, dto.FromUserModel(updated))
}

// ───── Collector applications ─────

func (h *AdminHandler) ListApplications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var statusF *model.ApplicationStatus
	if s := c.Query("status"); s != "" {
		st := model.ApplicationStatus(s)
		if !st.IsValid() {
			response.Err(c, apperr.New(apperr.CodeValidation, "invalid status filter"))
			return
		}
		statusF = &st
	}

	list, total, err := h.apps.List(c.Request.Context(), statusF, page, pageSize)
	if err != nil {
		response.Err(c, err)
		return
	}

	out := make([]dto.CollectorApplicationResponse, 0, len(list))
	for i := range list {
		out = append(out, dto.FromApplicationModel(&list[i]))
	}
	response.Paginated(c, out, response.Pagination{
		Page: page, PageSize: pageSize, Total: total,
	})
}

func (h *AdminHandler) ApproveApplication(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, apperr.New(apperr.CodeValidation, "invalid application id"))
		return
	}
	actorID, err := h.actorIDFromCtx(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	actorKCID, err := actorKCIDFromCtx(c)
	if err != nil {
		response.Err(c, err)
		return
	}

	app, err := h.apps.Approve(c.Request.Context(), id, actorID, actorKCID,
		c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, dto.FromApplicationModel(app))
}

func (h *AdminHandler) RejectApplication(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Err(c, apperr.New(apperr.CodeValidation, "invalid application id"))
		return
	}
	var req dto.RejectApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Err(c, apperr.Wrap(apperr.CodeValidation, "invalid request body", err))
		return
	}
	actorID, err := h.actorIDFromCtx(c)
	if err != nil {
		response.Err(c, err)
		return
	}
	actorKCID, err := actorKCIDFromCtx(c)
	if err != nil {
		response.Err(c, err)
		return
	}

	app, err := h.apps.Reject(c.Request.Context(), id, actorID, actorKCID, req.Reason,
		c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.Err(c, err)
		return
	}
	response.OK(c, dto.FromApplicationModel(app))
}

// actorIDFromCtx returns the admin's LOCAL users.id by looking up the JWT
// subject (keycloak_id) in the user repository.
//
// Why we can't just use keycloak_id: foreign keys like
// collector_applications.reviewed_by reference users.id, not keycloak_id.
//
// In the future we plan to add a `db_user_id` custom claim mapper in
// Keycloak so we can skip this lookup; for now this lookup is the
// authoritative way to get the admin's local id.
func (h *AdminHandler) actorIDFromCtx(c *gin.Context) (uuid.UUID, error) {
	kcID, err := actorKCIDFromCtx(c)
	if err != nil {
		return uuid.Nil, err
	}
	u, err := h.userRepo.FindByKeycloakID(c.Request.Context(), kcID)
	if err != nil {
		// User belum sync? Frontend should call /v1/auth/sync first; admin
		// can't perform admin actions without a local user row.
		if ae, ok := apperr.As(err); ok && ae.Code == apperr.CodeNotFound {
			return uuid.Nil, apperr.New(apperr.CodeForbidden,
				"admin local profile not synced; call POST /v1/auth/sync first")
		}
		return uuid.Nil, err
	}
	return u.ID, nil
}

func actorKCIDFromCtx(c *gin.Context) (uuid.UUID, error) {
	claims := middleware.MustClaims(c)
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, apperr.Wrap(apperr.CodeTokenInvalid, "subject not a uuid", err)
	}
	return id, nil
}
