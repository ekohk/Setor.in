package model

import (
	"database/sql/driver"
	"fmt"
)

// Status mirrors the order_status PG enum. See specs/0004-order/spec.md §4.
type Status string

const (
	StatusReceived      Status = "received"
	StatusAccepted      Status = "accepted"
	StatusEnroute       Status = "enroute"
	StatusArrived       Status = "arrived"
	StatusWeighing      Status = "weighing"
	StatusQuality       Status = "quality"
	StatusCashHandover  Status = "cash_handover"
	StatusDone          Status = "done"
	StatusCancelled     Status = "cancelled"
	StatusDisputed      Status = "disputed"
)

// IsTerminal returns true if no further transition is allowed.
func (s Status) IsTerminal() bool {
	switch s {
	case StatusDone, StatusCancelled, StatusDisputed:
		return true
	}
	return false
}

func (s Status) IsValid() bool {
	switch s {
	case StatusReceived, StatusAccepted, StatusEnroute, StatusArrived,
		StatusWeighing, StatusQuality, StatusCashHandover,
		StatusDone, StatusCancelled, StatusDisputed:
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

// Method is pickup vs dropoff.
type Method string

const (
	MethodPickup  Method = "pickup"
	MethodDropoff Method = "dropoff"
)

func (m Method) IsValid() bool {
	return m == MethodPickup || m == MethodDropoff
}

func (m *Method) Scan(src any) error {
	switch v := src.(type) {
	case string:
		*m = Method(v)
	case []byte:
		*m = Method(string(v))
	case nil:
		*m = ""
	default:
		return fmt.Errorf("scan Method: unsupported type %T", src)
	}
	return nil
}

func (m Method) Value() (driver.Value, error) {
	if m == "" {
		return nil, nil
	}
	return string(m), nil
}

// Quality grade impacts payout. Hardcoded MVP — admin-configurable in Phase 2.
type Grade string

const (
	GradeA Grade = "A" // clean, sorted     → harga penuh (potongan Rp 0/kg)
	GradeB Grade = "B" // mixed              → potongan Rp 1.000/kg
	GradeC Grade = "C" // contaminated       → potongan Rp 2.000/kg
)

// DeductionPerKg returns the flat price deduction in rupiah per kg.
// Grade A = Rp 0 (full price), B = Rp -1.000, C = Rp -2.000.
func DeductionPerKg(g Grade) int {
	switch g {
	case GradeB:
		return -1000
	case GradeC:
		return -2000
	}
	return 0
}
