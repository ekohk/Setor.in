package usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/order/application/dto"
	"github.com/setorin/setorin/backend/internal/apps/order/application/ports"
	"github.com/setorin/setorin/backend/internal/apps/order/domain/model"
	userports "github.com/setorin/setorin/backend/internal/apps/user/application/ports"
	emailsvc "github.com/setorin/setorin/backend/internal/services/email"
	apperr "github.com/setorin/setorin/backend/internal/shared/errors"
)

// CollectorOrderUseCase covers collector (driver) actions.
type CollectorOrderUseCase struct {
	orderRepo ports.OrderRepository
	userRepo  userports.UserRepository
	mail      *emailsvc.Service
}

func NewCollectorOrderUseCase(
	orderRepo ports.OrderRepository,
	userRepo userports.UserRepository,
	mail *emailsvc.Service,
) *CollectorOrderUseCase {
	return &CollectorOrderUseCase{orderRepo: orderRepo, userRepo: userRepo, mail: mail}
}

// ListIncoming — pending orders, no collector assigned yet.
func (uc *CollectorOrderUseCase) ListIncoming(ctx context.Context, page, pageSize int) ([]model.Order, int64, error) {
	return uc.orderRepo.ListIncoming(ctx, page, pageSize)
}

// ListMine — orders assigned to this collector.
func (uc *CollectorOrderUseCase) ListMine(ctx context.Context, collectorID uuid.UUID, page, pageSize int) ([]model.Order, int64, error) {
	return uc.orderRepo.ListByCollector(ctx, collectorID, page, pageSize)
}

// Accept — collector picks up an order. Generates 4-digit OTP, sets
// expires_at = NOW() + 30min, transitions received → accepted.
func (uc *CollectorOrderUseCase) Accept(ctx context.Context, collectorID uuid.UUID, orderCode string) (*model.Order, error) {
	o, err := uc.orderRepo.FindByCode(ctx, orderCode)
	if err != nil {
		return nil, err
	}
	if o.CollectorID != nil {
		return nil, apperr.New(apperr.CodeConflict, "order already accepted by another collector")
	}

	otp, err := genOTP4()
	if err != nil {
		return nil, fmt.Errorf("generate OTP: %w", err)
	}
	expSpec := "set:30 minutes"
	updated, err := uc.orderRepo.Transition(ctx, o.ID, ports.TransitionParams{
		ExpectedStatus:  model.StatusReceived,
		NewStatus:       model.StatusAccepted,
		ChangedBy:       collectorID,
		Notes:           ptr("collector accepted; OTP generated"),
		CollectorID:     &collectorID,
		OTPCode:         &otp,
		OTPExpiresAt:    &expSpec,
		StampAcceptedAt: true,
	})
	if err != nil {
		return nil, err
	}

	if uc.mail != nil {
		if u, err := uc.userRepo.FindByID(ctx, o.UserID); err == nil {
			uc.mail.OrderAccepted(u.Email, u.FullName, updated.OrderCode, otp)
		}
	}
	return updated, nil
}

// StartPickup — driver memulai perjalanan (pickup case).
func (uc *CollectorOrderUseCase) StartPickup(ctx context.Context, collectorID uuid.UUID, orderCode string) (*model.Order, error) {
	o, err := uc.assertCollectorOwns(ctx, collectorID, orderCode)
	if err != nil {
		return nil, err
	}
	if o.Method != model.MethodPickup {
		return nil, apperr.New(apperr.CodeConflict, "start-pickup only valid for pickup orders")
	}
	updated, err := uc.orderRepo.Transition(ctx, o.ID, ports.TransitionParams{
		ExpectedStatus: model.StatusAccepted,
		NewStatus:      model.StatusEnroute,
		ChangedBy:      collectorID,
		Notes:          ptr("driver en route"),
	})
	if err != nil {
		return nil, err
	}
	if uc.mail != nil {
		if u, err := uc.userRepo.FindByID(ctx, o.UserID); err == nil {
			uc.mail.OrderEnroute(u.Email, u.FullName, updated.OrderCode)
		}
	}
	return updated, nil
}

// Arrive — driver tiba di lokasi. From either accepted (dropoff) or enroute (pickup).
func (uc *CollectorOrderUseCase) Arrive(ctx context.Context, collectorID uuid.UUID, orderCode string) (*model.Order, error) {
	o, err := uc.assertCollectorOwns(ctx, collectorID, orderCode)
	if err != nil {
		return nil, err
	}

	expected := model.StatusEnroute
	if o.Method == model.MethodDropoff {
		expected = model.StatusAccepted
	}
	updated, err := uc.orderRepo.Transition(ctx, o.ID, ports.TransitionParams{
		ExpectedStatus: expected,
		NewStatus:      model.StatusArrived,
		ChangedBy:      collectorID,
		Notes:          ptr("driver arrived"),
		StampArrivedAt: true,
	})
	if err != nil {
		return nil, err
	}
	if uc.mail != nil {
		if u, err := uc.userRepo.FindByID(ctx, o.UserID); err == nil {
			uc.mail.OrderArrived(u.Email, u.FullName, updated.OrderCode)
		}
	}
	return updated, nil
}

// VerifyOTP — driver inputs OTP shown by user. On match, transition
// arrived → weighing and CLEAR the OTP (one-time use).
func (uc *CollectorOrderUseCase) VerifyOTP(ctx context.Context, collectorID uuid.UUID, orderCode string, req dto.VerifyOTPRequest) (*model.Order, error) {
	o, err := uc.assertCollectorOwns(ctx, collectorID, orderCode)
	if err != nil {
		return nil, err
	}
	if o.OTPCode == nil || *o.OTPCode == "" {
		return nil, apperr.New(apperr.CodeConflict, "OTP not set or already verified")
	}
	if *o.OTPCode != req.OTPCode {
		return nil, apperr.New(apperr.CodeValidation, "incorrect OTP")
	}

	verified := true
	clearOTP := ""
	clearExpiry := "clear"
	updated, err := uc.orderRepo.Transition(ctx, o.ID, ports.TransitionParams{
		ExpectedStatus: model.StatusArrived,
		NewStatus:      model.StatusWeighing,
		ChangedBy:      collectorID,
		Notes:          ptr("OTP verified"),
		OTPCode:        &clearOTP,
		OTPVerifiedAt:  &verified,
		OTPExpiresAt:   &clearExpiry,
	})
	return updated, err
}

// Weigh — driver inputs actual weight after weighing on certified scale.
func (uc *CollectorOrderUseCase) Weigh(ctx context.Context, collectorID uuid.UUID, orderCode string, req dto.WeighRequest) (*model.Order, error) {
	o, err := uc.assertCollectorOwns(ctx, collectorID, orderCode)
	if err != nil {
		return nil, err
	}
	w, err := strconv.ParseFloat(strings.TrimSpace(req.ActualWeightKg), 64)
	if err != nil || w <= 0 {
		return nil, apperr.New(apperr.CodeValidation, "actual_weight_kg must be a positive number")
	}
	weightStr := formatWeight(w)
	updated, err := uc.orderRepo.Transition(ctx, o.ID, ports.TransitionParams{
		ExpectedStatus: model.StatusWeighing,
		NewStatus:      model.StatusQuality,
		ChangedBy:      collectorID,
		Notes:          ptr(fmt.Sprintf("weighed: %s kg", weightStr)),
		ActualWeightKg: &weightStr,
	})
	return updated, err
}

// Quality — driver assigns grade and final payout is computed.
//   final_payout = actual_weight * unit_price * (1 + bonus_pct/100)
func (uc *CollectorOrderUseCase) Quality(ctx context.Context, collectorID uuid.UUID, orderCode string, req dto.QualityRequest) (*model.Order, error) {
	o, err := uc.assertCollectorOwns(ctx, collectorID, orderCode)
	if err != nil {
		return nil, err
	}
	if o.ActualWeightKg == nil {
		return nil, apperr.New(apperr.CodeConflict, "weight not yet recorded")
	}
	w, err := strconv.ParseFloat(*o.ActualWeightKg, 64)
	if err != nil {
		return nil, fmt.Errorf("parse stored weight: %w", err)
	}

	bonus := model.BonusForGrade(model.Grade(req.Grade))
	finalPayout := int64(float64(o.UnitPriceAtOrder) * w * (1 + float64(bonus)/100))

	updated, err := uc.orderRepo.Transition(ctx, o.ID, ports.TransitionParams{
		ExpectedStatus:  model.StatusQuality,
		NewStatus:       model.StatusCashHandover,
		ChangedBy:       collectorID,
		Notes:           req.Notes,
		QualityGrade:    &req.Grade,
		QualityBonusPct: &bonus,
		FinalPayout:     &finalPayout,
	})
	return updated, err
}

// ─── helpers ────────────────────────────────────────────────────────────

func (uc *CollectorOrderUseCase) assertCollectorOwns(ctx context.Context, collectorID uuid.UUID, orderCode string) (*model.Order, error) {
	o, err := uc.orderRepo.FindByCode(ctx, orderCode)
	if err != nil {
		return nil, err
	}
	if o.CollectorID == nil || *o.CollectorID != collectorID {
		return nil, apperr.New(apperr.CodeForbidden, "this order is not assigned to you")
	}
	return o, nil
}

// genOTP4 returns a 4-digit string using crypto/rand (not math/rand which is
// predictable). Range 0000-9999 with leading zeros preserved.
func genOTP4() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d", n.Int64()), nil
}
