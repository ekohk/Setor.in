// Package ports defines the auth module's outbound interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/auth/domain/model"
	usermodel "github.com/setorin/setorin/backend/internal/apps/user/domain/model"
)

// AuthEventRepository writes audit rows. Append-only — no Update/Delete.
type AuthEventRepository interface {
	// Insert persists an event. Caller may pass nil UserID/KeycloakID/IP/UA/Metadata.
	Insert(ctx context.Context, e *model.AuthEvent) error
}

// CollectorApplicationRepository persists collector applications.
type CollectorApplicationRepository interface {
	Create(ctx context.Context, a *model.CollectorApplication) (*model.CollectorApplication, error)

	// FindByID returns the application or CodeNotFound.
	FindByID(ctx context.Context, id uuid.UUID) (*model.CollectorApplication, error)

	// FindPendingByUser returns the user's pending application (if any). Returns
	// (nil, nil) if no pending — pending is checked frequently and not-found is normal.
	FindPendingByUser(ctx context.Context, userID uuid.UUID) (*model.CollectorApplication, error)

	// List returns paginated applications, optionally filtered by status.
	List(ctx context.Context, statusFilter *model.ApplicationStatus, page, pageSize int) ([]model.CollectorApplication, int64, error)

	// Approve transitions pending → approved and stamps reviewer.
	Approve(ctx context.Context, id uuid.UUID, reviewerID uuid.UUID) (*model.CollectorApplication, error)

	// Reject transitions pending → rejected with a reason.
	Reject(ctx context.Context, id uuid.UUID, reviewerID uuid.UUID, reason string) (*model.CollectorApplication, error)
}

// AdminUserRepository is a query-side repo used by /v1/admin/users endpoints.
//
// We keep this separate from the `user` module's UserRepository to (a) avoid
// growing that interface with admin-specific concerns, and (b) make it explicit
// at call sites that these calls are admin-only and bypass normal scoping.
type AdminUserRepository interface {
	List(ctx context.Context, f UserListFilter) ([]usermodel.User, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status usermodel.Status) (*usermodel.User, error)
	UpdatePrimaryRole(ctx context.Context, id uuid.UUID, role usermodel.Role) (*usermodel.User, error)
}

// UserListFilter is the query-time filter for admin user listing.
// Zero values mean "no filter".
type UserListFilter struct {
	Page     int
	PageSize int
	Role     *usermodel.Role
	Status   *usermodel.Status
	Query    string // ILIKE on email or full_name
}
