// Package dto holds request/response shapes for the auth module.
package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/auth/domain/model"
	usermodel "github.com/setorin/setorin/backend/internal/apps/user/domain/model"
)

// ─── Collector application ───────────────────────────────────────────────────

type BecomeCollectorRequest struct {
	BusinessName string  `json:"business_name" binding:"required,min=3,max=120"`
	LicenseNo    *string `json:"license_no,omitempty"`
	KTPURL       string  `json:"ktp_url"       binding:"required,url"`
	SIUPURL      *string `json:"siup_url,omitempty" binding:"omitempty,url"`
	Address      string  `json:"address"       binding:"required,min=10"`
}

type RejectApplicationRequest struct {
	Reason string `json:"reason" binding:"required,min=5"`
}

type CollectorApplicationResponse struct {
	ID              uuid.UUID               `json:"id"`
	UserID          uuid.UUID               `json:"user_id"`
	BusinessName    string                  `json:"business_name"`
	LicenseNo       *string                 `json:"license_no,omitempty"`
	KTPURL          string                  `json:"ktp_url"`
	SIUPURL         *string                 `json:"siup_url,omitempty"`
	Address         string                  `json:"address"`
	Status          model.ApplicationStatus `json:"status"`
	RejectionReason *string                 `json:"rejection_reason,omitempty"`
	ReviewedBy      *uuid.UUID              `json:"reviewed_by,omitempty"`
	SubmittedAt     time.Time               `json:"submitted_at"`
	ReviewedAt      *time.Time              `json:"reviewed_at,omitempty"`
}

func FromApplicationModel(a *model.CollectorApplication) CollectorApplicationResponse {
	return CollectorApplicationResponse{
		ID:              a.ID,
		UserID:          a.UserID,
		BusinessName:    a.BusinessName,
		LicenseNo:       a.LicenseNo,
		KTPURL:          a.KTPURL,
		SIUPURL:         a.SIUPURL,
		Address:         a.Address,
		Status:          a.Status,
		RejectionReason: a.RejectionReason,
		ReviewedBy:      a.ReviewedBy,
		SubmittedAt:     a.SubmittedAt,
		ReviewedAt:      a.ReviewedAt,
	}
}

// ─── Admin user management ───────────────────────────────────────────────────

type UpdateUserStatusRequest struct {
	Status usermodel.Status `json:"status" binding:"required,oneof=active suspended"`
	Reason *string          `json:"reason,omitempty"`
}

type UpdateUserRoleRequest struct {
	// New primary_role to set (also reflected in Keycloak realm roles).
	Role   usermodel.Role `json:"role" binding:"required,oneof=user collector admin super_admin"`
	Reason *string        `json:"reason,omitempty"`
}

// AdminUserResponse mirrors the user's view but with no roles array
// (admin lists by primary_role denormalization).
type AdminUserResponse struct {
	ID            uuid.UUID         `json:"id"`
	KeycloakID    uuid.UUID         `json:"keycloak_id"`
	Email         string            `json:"email"`
	Phone         *string           `json:"phone,omitempty"`
	FullName      string            `json:"full_name"`
	PrimaryRole   usermodel.Role    `json:"primary_role"`
	Status        usermodel.Status  `json:"status"`
	EmailVerified bool              `json:"email_verified"`
	LastLoginAt   *time.Time        `json:"last_login_at,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
}

func FromUserModel(u *usermodel.User) AdminUserResponse {
	return AdminUserResponse{
		ID:            u.ID,
		KeycloakID:    u.KeycloakID,
		Email:         u.Email,
		Phone:         u.Phone,
		FullName:      u.FullName,
		PrimaryRole:   u.PrimaryRole,
		Status:        u.Status,
		EmailVerified: u.EmailVerified,
		LastLoginAt:   u.LastLoginAt,
		CreatedAt:     u.CreatedAt,
	}
}
