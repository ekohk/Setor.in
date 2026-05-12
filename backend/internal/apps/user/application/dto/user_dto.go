// Package dto holds request/response shapes that cross the application boundary.
package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/user/domain/model"
)

// UserResponse is the canonical user representation returned by the API.
type UserResponse struct {
	ID            uuid.UUID  `json:"id"`
	KeycloakID    uuid.UUID  `json:"keycloak_id"`
	Email         string     `json:"email"`
	Phone         *string    `json:"phone,omitempty"`
	FullName      string     `json:"full_name"`
	AvatarURL     *string    `json:"avatar_url,omitempty"`
	PrimaryRole   model.Role `json:"primary_role"`
	Status        model.Status `json:"status"`
	EmailVerified bool       `json:"email_verified"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// MeResponse extends UserResponse with the live JWT roles.
// Use for /v1/auth/me where the caller should see their effective permissions.
type MeResponse struct {
	UserResponse
	Roles []string `json:"roles"`
}

// FromModel converts a domain User to a wire-friendly UserResponse.
func FromModel(u *model.User) UserResponse {
	return UserResponse{
		ID:            u.ID,
		KeycloakID:    u.KeycloakID,
		Email:         u.Email,
		Phone:         u.Phone,
		FullName:      u.FullName,
		AvatarURL:     u.AvatarURL,
		PrimaryRole:   u.PrimaryRole,
		Status:        u.Status,
		EmailVerified: u.EmailVerified,
		LastLoginAt:   u.LastLoginAt,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

// UpdateProfileRequest is the body for PATCH /v1/users/me.
// All fields optional; only non-nil ones are updated.
type UpdateProfileRequest struct {
	FullName  *string `json:"full_name,omitempty"  binding:"omitempty,min=1,max=120"`
	Phone     *string `json:"phone,omitempty"      binding:"omitempty,e164"`
	AvatarURL *string `json:"avatar_url,omitempty" binding:"omitempty,url"`
}
