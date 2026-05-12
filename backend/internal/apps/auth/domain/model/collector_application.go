package model

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ApplicationStatus mirrors the Postgres application_status enum.
type ApplicationStatus string

const (
	ApplicationPending  ApplicationStatus = "pending"
	ApplicationApproved ApplicationStatus = "approved"
	ApplicationRejected ApplicationStatus = "rejected"
)

func (s ApplicationStatus) IsValid() bool {
	switch s {
	case ApplicationPending, ApplicationApproved, ApplicationRejected:
		return true
	}
	return false
}

func (s *ApplicationStatus) Scan(src any) error {
	switch v := src.(type) {
	case string:
		*s = ApplicationStatus(v)
	case []byte:
		*s = ApplicationStatus(string(v))
	case nil:
		*s = ""
	default:
		return fmt.Errorf("scan ApplicationStatus: unsupported type %T", src)
	}
	return nil
}

func (s ApplicationStatus) Value() (driver.Value, error) {
	if s == "" {
		return nil, nil
	}
	return string(s), nil
}

// CollectorApplication is a request from a `user` to be promoted to `collector`.
// One pending application per user is enforced at the DB level (partial unique).
type CollectorApplication struct {
	ID              uuid.UUID         `db:"id"`
	UserID          uuid.UUID         `db:"user_id"`
	BusinessName    string            `db:"business_name"`
	LicenseNo       *string           `db:"license_no"`
	KTPURL          string            `db:"ktp_url"`
	SIUPURL         *string           `db:"siup_url"`
	Address         string            `db:"address"`
	Status          ApplicationStatus `db:"status"`
	RejectionReason *string           `db:"rejection_reason"`
	ReviewedBy      *uuid.UUID        `db:"reviewed_by"`
	SubmittedAt     time.Time         `db:"submitted_at"`
	ReviewedAt      *time.Time        `db:"reviewed_at"`
}
