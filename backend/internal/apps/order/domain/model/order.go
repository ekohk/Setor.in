// Package model holds the order aggregate.
//
// Domain layer rule: pure data + invariants, no I/O, no DB driver, no HTTP.
package model

import (
	"time"

	"github.com/google/uuid"
)

// Order represents one recycling transaction from creation to completion.
//
// Money fields are BIGINT rupiah (integer — no fractional rupiah).
// Weight is fractional in kg (e.g. 3.500). Stored as string in DB (NUMERIC)
// and as string here too for safety; usecase converts to/from float64 only
// at display boundaries.
type Order struct {
	ID                  uuid.UUID  `db:"id"`
	OrderCode           string     `db:"order_code"`
	UserID              uuid.UUID  `db:"user_id"`
	CollectorID         *uuid.UUID `db:"collector_id"`
	MaterialID          uuid.UUID  `db:"material_id"`

	EstimatedWeightKg   string     `db:"estimated_weight_kg"`   // NUMERIC stored as string for precision
	ActualWeightKg      *string    `db:"actual_weight_kg"`

	UnitPriceAtOrder    int64      `db:"unit_price_at_order"`
	EstimatedPayout     int64      `db:"estimated_payout"`
	FinalPayout         *int64     `db:"final_payout"`

	QualityGrade        *string    `db:"quality_grade"`
	QualityBonusPct     int        `db:"quality_bonus_pct"`

	Method              Method     `db:"method"`

	PaymentMethod       string     `db:"payment_method"`  // 'cash' for MVP
	PaymentStatus       string     `db:"payment_status"`  // 'pending' | 'paid' | 'disputed'
	PaidAt              *time.Time `db:"paid_at"`

	Status              Status     `db:"status"`

	OTPCode             *string    `db:"otp_code"`
	OTPVerifiedAt       *time.Time `db:"otp_verified_at"`
	OTPExpiresAt        *time.Time `db:"otp_expires_at"`

	AddressText         string     `db:"address_text"`
	Latitude            *float64   `db:"latitude"`
	Longitude           *float64   `db:"longitude"`
	Notes               *string    `db:"notes"`

	CreatedAt           time.Time  `db:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at"`
	AcceptedAt          *time.Time `db:"accepted_at"`
	ArrivedAt           *time.Time `db:"arrived_at"`
	CompletedAt         *time.Time `db:"completed_at"`
	CancelledAt         *time.Time `db:"cancelled_at"`
	CancelledBy         *uuid.UUID `db:"cancelled_by"`
	CancellationReason  *string    `db:"cancellation_reason"`
}

// StatusHistory is one row in `order_status_history`.
type StatusHistory struct {
	ID         int64      `db:"id"`
	OrderID    uuid.UUID  `db:"order_id"`
	FromStatus *Status    `db:"from_status"`
	ToStatus   Status     `db:"to_status"`
	ChangedBy  *uuid.UUID `db:"changed_by"`
	Notes      *string    `db:"notes"`
	ChangedAt  time.Time  `db:"changed_at"`
}
