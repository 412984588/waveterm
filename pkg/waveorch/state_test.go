// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package waveorch

import (
	"testing"
)

func TestStateMachine_Transition(t *testing.T) {
	sm := NewStateMachine()

	if sm.CurrentState() != StateIdle {
		t.Errorf("initial state = %v, want IDLE", sm.CurrentState())
	}

	// Valid transition
	if err := sm.Transition(StatePlanning); err != nil {
		t.Errorf("transition to PLANNING failed: %v", err)
	}

	// Invalid transition
	if err := sm.Transition(StateCompleted); err == nil {
		t.Error("should fail: PLANNING -> COMPLETED not allowed")
	}
}
