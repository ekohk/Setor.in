// Package usecase orchestrates business logic for the order module.
package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/order/application/dto"
	"github.com/setorin/setorin/backend/internal/apps/order/application/ports"
	"github.com/setorin/setorin/backend/internal/apps/order/domain/model"
	"github.com/setorin/setorin/backend/internal/apps/order/domain/services"
	catalogports "github.com/setorin/setorin/backend/internal/apps/catalog/application/ports"
	userports "github.com/setorin/setorin/backend/internal/apps/user/application/ports"
	emailsvc "github.com/setorin/setorin/backend/internal/services/email"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
)

// UserOrderUseCase covers user (seller) actions on orders.
type UserOrderUseCase struct {
	orderRepo    ports.OrderRepository
	materialRepo catalogports.MaterialRepository
	userRepo     userports.UserRepository
	mail         *emailsvc.Service
}

func NewUserOrderUseCase(
	orderRepo ports.OrderRepository,
	materialRepo catalogports.MaterialRepository,
	userRepo userports.UserRepository,
	mail *emailsvc.Service,
) *UserOrderUseCase {
	return &UserOrderUseCase{orderRepo: orderRepo, materialRepo: materialRepo, userRepo: userRepo, mail: mail}
}

// Create — POST /v1/orders.
//
// Snapshots material's current price at order time, so price changes
// mid-flow don't surprise the user.
func (uc *UserOrderUseCase) Create(ctx context.Context, sellerID uuid.UUID, req dto.CreateOrderRequest) (*model.Order, error) {
	// 1. Look up material by slug + current price (must be active).
	m, err := uc.materialRepo.FindBySlug(ctx, req.MaterialSlug)
	if err != nil {
		return nil, err
	}
	if !m.IsActive {
		return nil, apperr.New(apperr.CodeValidation, "material not available")
	}
	if m.CurrentPrice == nil {
		return nil, apperr.New(apperr.CodeValidation, "no price set for this material")
	}

	// 2. Parse weight (string → float for math, then format back string for storage).
	w, err := strconv.ParseFloat(strings.TrimSpace(req.EstimatedWeightKg), 64)
	if err != nil || w <= 0 {
		return nil, apperr.New(apperr.CodeValidation, "estimated_weight_kg must be a positive number")
	}

	// 3. Compute estimated payout (integer rupiah, truncated).
	estPayout := int64(float64(*m.CurrentPrice) * w)
	if estPayout <= 0 {
		return nil, apperr.New(apperr.CodeValidation, "computed payout is zero")
	}

	// 4. Reserve order code.
	code, err := uc.orderRepo.NextOrderCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("reserve order code: %w", err)
	}

	// 5. Insert.
	o := &model.Order{
		OrderCode:           code,
		UserID:              sellerID,
		MaterialID:          m.ID,
		EstimatedWeightKg:   formatWeight(w),
		UnitPriceAtOrder:    *m.CurrentPrice,
		EstimatedPayout:     estPayout,
		Method:              model.Method(req.Method),
		AddressText:         req.AddressText,
		Latitude:            req.Latitude,
		Longitude:           req.Longitude,
		Notes:               req.Notes,
	}
	created, err := uc.orderRepo.Create(ctx, o)
	if err != nil {
		return nil, err
	}

	// 6. Notify user (best-effort, non-blocking).
	if uc.mail != nil {
		if u, err := uc.userRepo.FindByID(ctx, sellerID); err == nil {
			uc.mail.OrderCreated(u.Email, u.FullName, created.OrderCode, m.Name, estPayout)
		}
	}
	return created, nil
}

// Cancel — only allowed if status = received (collector hasn't accepted).
func (uc *UserOrderUseCase) Cancel(ctx context.Context, sellerID uuid.UUID, orderCode, reason string) (*model.Order, error) {
	o, err := uc.orderRepo.FindByCode(ctx, orderCode)
	if err != nil {
		return nil, err
	}
	if o.UserID != sellerID {
		return nil, apperr.New(apperr.CodeForbidden, "you can only cancel your own orders")
	}
	if o.Status != model.StatusReceived {
		return nil, apperr.New(apperr.CodeConflict,
			"can only cancel an order while still in 'received' state")
	}
	if !services.CanTransition(o.Status, model.StatusCancelled) {
		return nil, apperr.New(apperr.CodeConflict, "invalid state transition")
	}

	updated, err := uc.orderRepo.Transition(ctx, o.ID, ports.TransitionParams{
		ExpectedStatus:     model.StatusReceived,
		NewStatus:          model.StatusCancelled,
		ChangedBy:          sellerID,
		Notes:              ptr("cancelled by user"),
		CancellationReason: &reason,
		StampCancelledAt:   true,
	})
	if err != nil {
		return nil, err
	}

	if uc.mail != nil {
		if u, err := uc.userRepo.FindByID(ctx, sellerID); err == nil {
			uc.mail.OrderCancelled(u.Email, u.FullName, updated.OrderCode, reason)
		}
	}
	return updated, nil
}

// ConfirmCash — POST /v1/orders/:code/confirm-cash. User confirms cash received.
// Final transition cash_handover → done. Marks payment_status=paid.
func (uc *UserOrderUseCase) ConfirmCash(ctx context.Context, sellerID uuid.UUID, orderCode string) (*model.Order, error) {
	o, err := uc.orderRepo.FindByCode(ctx, orderCode)
	if err != nil {
		return nil, err
	}
	if o.UserID != sellerID {
		return nil, apperr.New(apperr.CodeForbidden, "only the seller can confirm cash receipt")
	}
	if o.Status != model.StatusCashHandover {
		return nil, apperr.New(apperr.CodeConflict,
			fmt.Sprintf("order is in state %q; can only confirm cash from 'cash_handover'", o.Status))
	}

	paid := "paid"
	stampPaid := true
	updated, err := uc.orderRepo.Transition(ctx, o.ID, ports.TransitionParams{
		ExpectedStatus:   model.StatusCashHandover,
		NewStatus:        model.StatusDone,
		ChangedBy:        sellerID,
		Notes:            ptr("user confirmed cash received"),
		PaymentStatus:    &paid,
		PaidAt:           &stampPaid,
		StampCompletedAt: true,
	})
	if err != nil {
		return nil, err
	}

	if uc.mail != nil {
		if u, err := uc.userRepo.FindByID(ctx, sellerID); err == nil {
			finalPayout := int64(0)
			if updated.FinalPayout != nil {
				finalPayout = *updated.FinalPayout
			}
			uc.mail.OrderCompleted(u.Email, u.FullName, updated.OrderCode, finalPayout)
		}
	}
	return updated, nil
}

// ─── Read-side ──────────────────────────────────────────────────────────

func (uc *UserOrderUseCase) ListMine(ctx context.Context, sellerID uuid.UUID, page, pageSize int) ([]model.Order, int64, error) {
	return uc.orderRepo.ListByUser(ctx, sellerID, page, pageSize)
}

func (uc *UserOrderUseCase) GetByCode(ctx context.Context, viewerID uuid.UUID, orderCode string, isAdmin bool) (*model.Order, error) {
	o, err := uc.orderRepo.FindByCode(ctx, orderCode)
	if err != nil {
		return nil, err
	}
	// Authorization: seller, assigned collector, or admin can view.
	if !isAdmin && viewerID != o.UserID && (o.CollectorID == nil || viewerID != *o.CollectorID) {
		return nil, apperr.New(apperr.CodeForbidden, "you don't have access to this order")
	}
	return o, nil
}

// ─── helpers ────────────────────────────────────────────────────────────

func ptr[T any](v T) *T { return &v }

// formatWeight → 3 decimal places, no trailing zeros stripped (DB NUMERIC handles).
func formatWeight(w float64) string {
	return strconv.FormatFloat(w, 'f', 3, 64)
}
