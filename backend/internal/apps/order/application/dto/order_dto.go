package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/order/domain/model"
)

// ─── Requests ────────────────────────────────────────────────────────────

// CreateOrderRequest — body for POST /v1/orders.
type CreateOrderRequest struct {
	MaterialSlug      string   `json:"material_slug"       binding:"required"`
	EstimatedWeightKg string   `json:"estimated_weight_kg" binding:"required"`   // e.g. "3.500"
	Method            string   `json:"method"              binding:"required,oneof=pickup dropoff"`
	AddressText       string   `json:"address_text"        binding:"required,min=10"`
	Latitude          *float64 `json:"latitude,omitempty"`
	Longitude         *float64 `json:"longitude,omitempty"`
	Notes             *string  `json:"notes,omitempty"`
}

// CancelOrderRequest — body for POST /v1/orders/:code/cancel.
type CancelOrderRequest struct {
	Reason string `json:"reason" binding:"required,min=3"`
}

// VerifyOTPRequest — body for POST /v1/collector/orders/:code/verify-otp.
type VerifyOTPRequest struct {
	OTPCode string `json:"otp_code" binding:"required,len=4"`
}

// WeighRequest — body for POST /v1/collector/orders/:code/weigh.
type WeighRequest struct {
	ActualWeightKg string `json:"actual_weight_kg" binding:"required"` // "3.700"
}

// QualityRequest — body for POST /v1/collector/orders/:code/quality.
type QualityRequest struct {
	Grade string  `json:"grade"            binding:"required,oneof=A B C"`
	Notes *string `json:"notes,omitempty"`
}

// ─── Responses ──────────────────────────────────────────────────────────

type OrderResponse struct {
	ID                  uuid.UUID  `json:"id"`
	OrderCode           string     `json:"order_code"`
	UserID              uuid.UUID  `json:"user_id"`
	CollectorID         *uuid.UUID `json:"collector_id,omitempty"`
	MaterialID          uuid.UUID  `json:"material_id"`

	EstimatedWeightKg   string     `json:"estimated_weight_kg"`
	ActualWeightKg      *string    `json:"actual_weight_kg,omitempty"`

	UnitPriceAtOrder    int64      `json:"unit_price_at_order"`
	EstimatedPayout     int64      `json:"estimated_payout"`
	FinalPayout         *int64     `json:"final_payout,omitempty"`

	QualityGrade        *string    `json:"quality_grade,omitempty"`
	QualityBonusPct     int        `json:"quality_bonus_pct"`

	Method              string     `json:"method"`
	PaymentMethod       string     `json:"payment_method"`
	PaymentStatus       string     `json:"payment_status"`
	PaidAt              *time.Time `json:"paid_at,omitempty"`

	Status              string     `json:"status"`
	OTPCode             *string    `json:"otp_code,omitempty"`           // returned only to seller / assigned collector
	OTPExpiresAt        *time.Time `json:"otp_expires_at,omitempty"`

	AddressText         string     `json:"address_text"`
	Latitude            *float64   `json:"latitude,omitempty"`
	Longitude           *float64   `json:"longitude,omitempty"`
	Notes               *string    `json:"notes,omitempty"`

	CreatedAt           time.Time  `json:"created_at"`
	AcceptedAt          *time.Time `json:"accepted_at,omitempty"`
	ArrivedAt           *time.Time `json:"arrived_at,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	CancelledAt         *time.Time `json:"cancelled_at,omitempty"`
	CancellationReason  *string    `json:"cancellation_reason,omitempty"`
}

// FromModel — viewerID controls which fields are exposed (OTP for seller +
// assigned collector only; admin always sees it).
func FromModel(o *model.Order, viewerID uuid.UUID, viewerIsAdmin bool) OrderResponse {
	r := OrderResponse{
		ID:                  o.ID,
		OrderCode:           o.OrderCode,
		UserID:              o.UserID,
		CollectorID:         o.CollectorID,
		MaterialID:          o.MaterialID,
		EstimatedWeightKg:   o.EstimatedWeightKg,
		ActualWeightKg:      o.ActualWeightKg,
		UnitPriceAtOrder:    o.UnitPriceAtOrder,
		EstimatedPayout:     o.EstimatedPayout,
		FinalPayout:         o.FinalPayout,
		QualityGrade:        o.QualityGrade,
		QualityBonusPct:     o.QualityBonusPct,
		Method:              string(o.Method),
		PaymentMethod:       o.PaymentMethod,
		PaymentStatus:       o.PaymentStatus,
		PaidAt:              o.PaidAt,
		Status:              string(o.Status),
		OTPExpiresAt:        o.OTPExpiresAt,
		AddressText:         o.AddressText,
		Latitude:            o.Latitude,
		Longitude:           o.Longitude,
		Notes:               o.Notes,
		CreatedAt:           o.CreatedAt,
		AcceptedAt:          o.AcceptedAt,
		ArrivedAt:           o.ArrivedAt,
		CompletedAt:         o.CompletedAt,
		CancelledAt:         o.CancelledAt,
		CancellationReason:  o.CancellationReason,
	}

	// OTP is sensitive: only seller, assigned collector, and admin see it.
	mayViewOTP := viewerIsAdmin ||
		viewerID == o.UserID ||
		(o.CollectorID != nil && viewerID == *o.CollectorID)
	if mayViewOTP {
		r.OTPCode = o.OTPCode
	}
	return r
}

type StatusHistoryResponse struct {
	FromStatus *string    `json:"from_status,omitempty"`
	ToStatus   string     `json:"to_status"`
	Notes      *string    `json:"notes,omitempty"`
	ChangedAt  time.Time  `json:"changed_at"`
}

func FromHistoryModel(h *model.StatusHistory) StatusHistoryResponse {
	var from *string
	if h.FromStatus != nil {
		s := string(*h.FromStatus)
		from = &s
	}
	return StatusHistoryResponse{
		FromStatus: from,
		ToStatus:   string(h.ToStatus),
		Notes:      h.Notes,
		ChangedAt:  h.ChangedAt,
	}
}
