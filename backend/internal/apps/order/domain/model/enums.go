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
	StatusInspection    Status = "inspection"
	StatusFinalOffer    Status = "final_offer"
	StatusCashHandover  Status = "cash_handover"
	StatusDone          Status = "done"
	StatusCancelled     Status = "cancelled"
	StatusRejectedByUser Status = "rejected_by_user"
	StatusDisputed      Status = "disputed"
)

// IsTerminal returns true if no further transition is allowed.
func (s Status) IsTerminal() bool {
	switch s {
	case StatusDone, StatusCancelled, StatusRejectedByUser, StatusDisputed:
		return true
	}
	return false
}

func (s Status) IsValid() bool {
	switch s {
	case StatusReceived, StatusAccepted, StatusEnroute, StatusArrived,
		StatusWeighing, StatusInspection, StatusFinalOffer, StatusCashHandover,
		StatusDone, StatusCancelled, StatusRejectedByUser, StatusDisputed:
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
