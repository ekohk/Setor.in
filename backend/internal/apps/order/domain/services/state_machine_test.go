package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/setorin/setorin/backend/internal/apps/order/domain/model"
)

// TestCanTransition_AllowedEdges verifies every explicitly listed transition
// in the spec §4 state machine.
func TestCanTransition_AllowedEdges(t *testing.T) {
	allowed := []struct {
		from model.Status
		to   model.Status
	}{
		// received
		{model.StatusReceived, model.StatusAccepted},
		{model.StatusReceived, model.StatusCancelled},
		{model.StatusReceived, model.StatusDisputed},
		// accepted
		{model.StatusAccepted, model.StatusEnroute},
		{model.StatusAccepted, model.StatusArrived}, // dropoff fast-path
		{model.StatusAccepted, model.StatusCancelled},
		{model.StatusAccepted, model.StatusDisputed},
		// enroute
		{model.StatusEnroute, model.StatusArrived},
		{model.StatusEnroute, model.StatusDisputed},
		// arrived
		{model.StatusArrived, model.StatusWeighing},
		{model.StatusArrived, model.StatusDisputed},
		// weighing
		{model.StatusWeighing, model.StatusQuality},
		{model.StatusWeighing, model.StatusDisputed},
		// quality
		{model.StatusQuality, model.StatusCashHandover},
		{model.StatusQuality, model.StatusDisputed},
		// cash_handover
		{model.StatusCashHandover, model.StatusDone},
		{model.StatusCashHandover, model.StatusDisputed},
	}

	for _, tc := range allowed {
		assert.True(t, CanTransition(tc.from, tc.to),
			"expected allowed: %s → %s", tc.from, tc.to)
	}
}

// TestCanTransition_ForbiddenEdges verifies that illegal transitions are rejected.
func TestCanTransition_ForbiddenEdges(t *testing.T) {
	forbidden := []struct {
		from model.Status
		to   model.Status
	}{
		// Skip a stage
		{model.StatusReceived, model.StatusEnroute},
		{model.StatusReceived, model.StatusDone},
		{model.StatusAccepted, model.StatusWeighing},
		{model.StatusEnroute, model.StatusWeighing},
		{model.StatusArrived, model.StatusCashHandover},
		// Backwards
		{model.StatusAccepted, model.StatusReceived},
		{model.StatusDone, model.StatusReceived},
		// Terminal → anything
		{model.StatusDone, model.StatusAccepted},
		{model.StatusCancelled, model.StatusReceived},
		{model.StatusDisputed, model.StatusDone},
		// received → done (skip everything)
		{model.StatusReceived, model.StatusWeighing},
		{model.StatusReceived, model.StatusQuality},
		{model.StatusReceived, model.StatusCashHandover},
	}

	for _, tc := range forbidden {
		assert.False(t, CanTransition(tc.from, tc.to),
			"expected forbidden: %s → %s", tc.from, tc.to)
	}
}

// TestCanTransition_TerminalStatesHaveNoOutbound verifies done/cancelled/disputed
// are true terminals — no outbound edges.
func TestCanTransition_TerminalStatesHaveNoOutbound(t *testing.T) {
	allStatuses := []model.Status{
		model.StatusReceived, model.StatusAccepted, model.StatusEnroute,
		model.StatusArrived, model.StatusWeighing, model.StatusQuality,
		model.StatusCashHandover, model.StatusDone,
		model.StatusCancelled, model.StatusDisputed,
	}
	terminals := []model.Status{model.StatusDone, model.StatusCancelled, model.StatusDisputed}

	for _, term := range terminals {
		for _, next := range allStatuses {
			assert.False(t, CanTransition(term, next),
				"terminal %s must not transition to %s", term, next)
		}
	}
}

// TestIsTerminal covers the Status.IsTerminal() helper.
func TestIsTerminal(t *testing.T) {
	assert.True(t, model.StatusDone.IsTerminal())
	assert.True(t, model.StatusCancelled.IsTerminal())
	assert.True(t, model.StatusDisputed.IsTerminal())

	assert.False(t, model.StatusReceived.IsTerminal())
	assert.False(t, model.StatusAccepted.IsTerminal())
	assert.False(t, model.StatusEnroute.IsTerminal())
	assert.False(t, model.StatusArrived.IsTerminal())
	assert.False(t, model.StatusWeighing.IsTerminal())
	assert.False(t, model.StatusQuality.IsTerminal())
	assert.False(t, model.StatusCashHandover.IsTerminal())
}

// TestNextActor verifies that the correct actor drives each transition stage.
func TestNextActor(t *testing.T) {
	cases := []struct {
		status model.Status
		actor  string
	}{
		{model.StatusReceived, "collector"},
		{model.StatusAccepted, "collector"},
		{model.StatusEnroute, "collector"},
		{model.StatusArrived, "collector"},
		{model.StatusWeighing, "collector"},
		{model.StatusQuality, "collector"},
		{model.StatusCashHandover, "user"},
		{model.StatusDone, ""},
		{model.StatusCancelled, ""},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.actor, NextActor(tc.status), "status=%s", tc.status)
	}
}
