package model

import (
	"database/sql/driver"
	"fmt"
)

// Role mirrors the user_role enum in Postgres and the Keycloak realm role.
type Role string

const (
	RoleUser       Role = "user"
	RoleCollector  Role = "collector"
	RoleAdmin      Role = "admin"
	RoleSuperAdmin Role = "super_admin"
)

// AllRoles in promotion order (lowest → highest privilege).
var AllRoles = []Role{RoleUser, RoleCollector, RoleAdmin, RoleSuperAdmin}

// IsValid returns true if r is one of AllRoles.
func (r Role) IsValid() bool {
	for _, v := range AllRoles {
		if v == r {
			return true
		}
	}
	return false
}

// Scan implements sql.Scanner for reading from Postgres enum.
func (r *Role) Scan(src any) error {
	switch v := src.(type) {
	case string:
		*r = Role(v)
	case []byte:
		*r = Role(string(v))
	case nil:
		*r = ""
	default:
		return fmt.Errorf("scan Role: unsupported type %T", src)
	}
	return nil
}

// Value implements driver.Valuer for writing to Postgres enum.
func (r Role) Value() (driver.Value, error) {
	if r == "" {
		return nil, nil
	}
	return string(r), nil
}

// Status mirrors the user_status enum.
type Status string

const (
	StatusPendingVerification Status = "pending_verification"
	StatusActive              Status = "active"
	StatusSuspended           Status = "suspended"
	StatusDeleted             Status = "deleted"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusPendingVerification, StatusActive, StatusSuspended, StatusDeleted:
		return true
	}
	return false
}

func (s *Status) Scan(src any) error {
	switch v := src.(type) {
	case string:
		*s = Status(v)
	case []byte:
		*s = Status(string(v))
	case nil:
		*s = ""
	default:
		return fmt.Errorf("scan Status: unsupported type %T", src)
	}
	return nil
}

func (s Status) Value() (driver.Value, error) {
	if s == "" {
		return nil, nil
	}
	return string(s), nil
}
