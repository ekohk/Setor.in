// Package ports defines the interfaces the application layer needs from
// outside (repositories, external clients). Implementations live in adapters/.
package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/user/domain/model"
)

// UserRepository abstracts persistence of User aggregates.
//
// Conventions:
//   - All methods take ctx for cancellation/timeout.
//   - "FindBy*" returns (nil, ErrNotFound) on miss — see errors.go.
//   - All write methods return updated User on success.
//   - Soft-delete: queries always filter deleted_at IS NULL unless suffix
//     "Including*" is used.
type UserRepository interface {
	// Create inserts a new user. Returns the inserted row (with generated id/timestamps).
	Create(ctx context.Context, u *model.User) (*model.User, error)

	// FindByKeycloakID returns the user mirrored from a Keycloak `sub`.
	FindByKeycloakID(ctx context.Context, kcID uuid.UUID) (*model.User, error)

	// FindByID returns the user by local id.
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)

	// FindByEmail returns the user by email (case-insensitive via citext column).
	FindByEmail(ctx context.Context, email string) (*model.User, error)

	// UpdateProfile patches mutable display fields. Pass nil to leave a field unchanged.
	UpdateProfile(ctx context.Context, id uuid.UUID, p ProfilePatch) (*model.User, error)

	// UpdateLastLogin sets last_login_at = NOW().
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error

	// UpdateStatusFromKeycloak refreshes email_verified + status. Used by /v1/auth/sync.
	UpdateAuthState(ctx context.Context, id uuid.UUID, emailVerified bool, status model.Status) (*model.User, error)
}

// ProfilePatch carries optional updates for the user profile.
// nil pointer = leave field unchanged; non-nil = overwrite.
type ProfilePatch struct {
	FullName  *string
	Phone     *string
	AvatarURL *string
}

// HasAny returns true if at least one field is set.
func (p ProfilePatch) HasAny() bool {
	return p.FullName != nil || p.Phone != nil || p.AvatarURL != nil
}
