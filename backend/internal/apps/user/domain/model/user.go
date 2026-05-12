// Package model contains the User aggregate and value types.
//
// Domain layer rule: NO imports of database, http, gin, pgx, etc.
// Only stdlib + uuid + types from other domain models.
package model

import (
	"time"

	"github.com/google/uuid"
)

// User mirrors authenticated identity from Keycloak into the Setor.in DB.
// Email and primary credentials live in Keycloak; this row holds business
// profile + role denormalization for display.
type User struct {
	ID            uuid.UUID  `db:"id"`
	KeycloakID    uuid.UUID  `db:"keycloak_id"`
	Email         string     `db:"email"`
	Phone         *string    `db:"phone"`
	FullName      string     `db:"full_name"`
	AvatarURL     *string    `db:"avatar_url"`
	PrimaryRole   Role       `db:"primary_role"`
	Status        Status     `db:"status"`
	EmailVerified bool       `db:"email_verified"`
	LastLoginAt   *time.Time `db:"last_login_at"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at"`
}

// IsActive returns true if the user can log in and use the platform.
func (u *User) IsActive() bool {
	return u.Status == StatusActive && u.DeletedAt == nil
}

// IsDeleted returns true if soft-deleted.
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil || u.Status == StatusDeleted
}
