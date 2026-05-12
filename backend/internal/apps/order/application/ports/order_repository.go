// Package ports defines outbound interfaces for the order module.
package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/order/domain/model"
)

// OrderRepository abstracts persistence of Order aggregates.
//
// Conventions:
//   - All mutating methods that touch order.status take changedBy + notes for
//     audit; repo writes both `orders` and `order_status_history` in one tx.
//   - Pagination methods take page (1-based) + pageSize.
//   - Lookups by order_code (human-friendly) AND by id (internal).
type OrderRepository interface {
	// NextOrderCode returns the next sequential order code in the form ECC-XXXXX.
	NextOrderCode(ctx context.Context) (string, error)

	// Create inserts a new order. Caller already snapshotted price/payout.
	Create(ctx context.Context, o *model.Order) (*model.Order, error)

	FindByID(ctx context.Context, id uuid.UUID) (*model.Order, error)
	FindByCode(ctx context.Context, code string) (*model.Order, error)

	// ListByUser returns orders where user_id = the seller.
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.Order, int64, error)

	// ListByCollector returns orders assigned to a specific collector.
	ListByCollector(ctx context.Context, collectorID uuid.UUID, page, pageSize int) ([]model.Order, int64, error)

	// ListIncoming returns pending orders not yet assigned to any collector.
	// MVP: no geo filter; phase 2 will add radius via PostGIS.
	ListIncoming(ctx context.Context, page, pageSize int) ([]model.Order, int64, error)

	// ListAdmin returns all orders, optionally filtered by status.
	ListAdmin(ctx context.Context, statusFilter *model.Status, page, pageSize int) ([]model.Order, int64, error)

	// Transition is the single atomic operation that:
	//   1) verifies current status matches `expectedStatus` (optimistic lock)
	//   2) updates status to newStatus
	//   3) appends a row to order_status_history
	//   4) optionally patches a set of fields atomically
	//
	// Returns ErrConflict (mapped from apperr.CodeConflict) if expectedStatus
	// doesn't match (another writer beat us, or invalid attempt).
	Transition(ctx context.Context, orderID uuid.UUID, p TransitionParams) (*model.Order, error)

	// StatusHistory returns the audit trail of a single order.
	StatusHistory(ctx context.Context, orderID uuid.UUID) ([]model.StatusHistory, error)
}

// TransitionParams bundles everything needed to advance an order's state.
type TransitionParams struct {
	ExpectedStatus model.Status   // current status we believe the order is in; if DB says otherwise → conflict
	NewStatus      model.Status
	ChangedBy      uuid.UUID      // local users.id of actor
	Notes          *string

	// Optional field patches applied in the same UPDATE statement. All nil = no extra updates.
	CollectorID         *uuid.UUID
	ActualWeightKg      *string
	FinalPayout         *int64
	QualityGrade        *string
	QualityBonusPct     *int
	OTPCode             *string  // pass empty string "" to clear
	OTPVerifiedAt       *bool    // true = set NOW(), nil = don't touch
	OTPExpiresAt        *string  // pass "clear" or "set:<duration>" — handled in repo
	PaymentStatus       *string
	PaidAt              *bool    // true = set NOW()
	CancellationReason  *string
	StampAcceptedAt     bool     // true = set accepted_at = NOW()
	StampArrivedAt      bool
	StampCompletedAt    bool
	StampCancelledAt    bool
}
