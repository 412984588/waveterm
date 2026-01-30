// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package waveorch

import (
	"testing"
)

func TestEngine_SubmitTask(t *testing.T) {
	registry := NewAgentRegistry()
	engine := NewEngine(registry, 3)

	task, err := engine.SubmitTask("/tmp/test", "test prompt", "claude-code")
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	if task.Status != "pending" {
		t.Errorf("status = %q, want pending", task.Status)
	}
}

func TestEngine_PauseResume(t *testing.T) {
	registry := NewAgentRegistry()
	engine := NewEngine(registry, 3)

	engine.Pause()
	if !engine.IsPaused() {
		t.Error("should be paused")
	}

	_, err := engine.SubmitTask("/tmp/test", "test", "claude-code")
	if err == nil {
		t.Error("should fail when paused")
	}

	engine.Resume()
	if engine.IsPaused() {
		t.Error("should not be paused")
	}
}
