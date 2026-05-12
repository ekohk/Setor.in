// Package model — auth module domain types (audit events, collector applications).
package model

import (
	"time"

	"github.com/google/uuid"
)

// AuthEventType is a stable string identifier for audit events.
// New types are append-only; never rename or repurpose existing values
// (downstream analytics depend on stability).
type AuthEventType string

const (
	EventLogin                          AuthEventType = "login"
	EventLogout                         AuthEventType = "logout"
	EventRegisterSynced                 AuthEventType = "register_synced"
	EventProfileUpdated                 AuthEventType = "profile_updated"
	EventRoleChanged                    AuthEventType = "role_changed"
	EventRolePromoted                   AuthEventType = "role_promoted"
	EventRoleDemoted                    AuthEventType = "role_demoted"
	EventAccountSuspended               AuthEventType = "account_suspended"
	EventAccountUnsuspended             AuthEventType = "account_unsuspended"
	EventCollectorApplicationSubmitted  AuthEventType = "collector_application_submitted"
	EventCollectorApplicationApproved   AuthEventType = "collector_application_approved"
	EventCollectorApplicationRejected   AuthEventType = "collector_application_rejected"
	EventFailedAuth                     AuthEventType = "failed_auth"
)

// AuthEvent is one row in the `auth_events` audit table.
// Append-only — never UPDATE or DELETE.
type AuthEvent struct {
	ID         int64         `db:"id"`
	UserID     *uuid.UUID    `db:"user_id"`     // nullable: user may not exist yet (e.g. failed_auth)
	KeycloakID *uuid.UUID    `db:"keycloak_id"` // nullable: useful before sync
	EventType  AuthEventType `db:"event_type"`
	IPAddress  *string       `db:"ip_address"`
	UserAgent  *string       `db:"user_agent"`
	Metadata   []byte        `db:"metadata"` // JSONB (raw)
	OccurredAt time.Time     `db:"occurred_at"`
}
