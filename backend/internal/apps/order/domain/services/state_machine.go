// Package services holds domain-level pure logic (no I/O).
//
// state_machine.go encodes the allowed transitions for Order.Status. Any
// attempt to transition to an unlisted next-state is rejected at the
// application layer.
package services

import (
	"github.com/setorin/setorin/backend/internal/apps/order/domain/model"
)

// allowed encodes the transition graph from spec §4.
// Key   = current status
// Value = set of statuses reachable from here
var allowed = map[model.Status]map[model.Status]struct{}{
	model.StatusReceived: {
		model.StatusAccepted:  {},
		model.StatusCancelled: {},
		model.StatusDisputed:  {},
	},
	model.StatusAccepted: {
		model.StatusEnroute:   {},  // pickup case
		model.StatusArrived:   {},  // dropoff case (user comes directly)
		model.StatusCancelled: {},  // late cancel allowed (auto-refund cash basis: nothing)
		model.StatusDisputed:  {},
	},
	model.StatusEnroute: {
		model.StatusArrived:  {},
		model.StatusDisputed: {},
	},
	model.StatusArrived: {
		model.StatusWeighing: {},  // after OTP verified
		model.StatusDisputed: {},
	},
	model.StatusWeighing: {
		model.StatusQuality:  {},
		model.StatusDisputed: {},
	},
	model.StatusQuality: {
		model.StatusCashHandover: {},
		model.StatusDisputed:     {},
	},
	model.StatusCashHandover: {
		model.StatusDone:     {},  // user confirms received
		model.StatusDisputed: {},
	},
	// Terminal: done, cancelled, disputed — no outbound edges
}

// CanTransition returns true if the transition from → to is allowed.
func CanTransition(from, to model.Status) bool {
	nexts, ok := allowed[from]
	if !ok {
		return false
	}
	_, ok = nexts[to]
	return ok
}

// NextActor returns who is expected to drive the transition. Helps RBAC
// check at the handler layer.
func NextActor(currentStatus model.Status) string {
	switch currentStatus {
	case model.StatusReceived:
		return "collector" // collector accepts
	case model.StatusAccepted, model.StatusEnroute, model.StatusArrived,
		model.StatusWeighing, model.StatusQuality:
		return "collector"
	case model.StatusCashHandover:
		return "user" // user confirms cash
	}
	return ""
}
