// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package waveorch

import (
	"testing"
)

func TestReportParser_ExtractJSON(t *testing.T) {
	parser := NewReportParser()

	input := `Some output text
<<<REPORT>>>
{"agent": "claude-code", "status": "SUCCESS"}
<<<END_REPORT>>>
More text`

	json := parser.ExtractJSON(input)
	expected := `{"agent": "claude-code", "status": "SUCCESS"}`
	if json != expected {
		t.Errorf("got %q, want %q", json, expected)
	}
}

func TestReportParser_Parse(t *testing.T) {
	parser := NewReportParser()

	input := `<<<REPORT>>>
{
  "agent": "claude-code",
  "status": "SUCCESS",
  "round": 1,
  "summary": "Task completed"
}
<<<END_REPORT>>>`

	report, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if report.Agent != "claude-code" {
		t.Errorf("agent = %q, want claude-code", report.Agent)
	}
	if report.Status != "SUCCESS" {
		t.Errorf("status = %q, want SUCCESS", report.Status)
	}
}

func TestReportParser_ContainsReport(t *testing.T) {
	parser := NewReportParser()

	if !parser.ContainsReport("<<<REPORT>>>...<<<END_REPORT>>>") {
		t.Error("should contain report")
	}
	if parser.ContainsReport("no markers here") {
		t.Error("should not contain report")
	}
}

func TestValidateReport(t *testing.T) {
	valid := &Report{Agent: "test", Status: "SUCCESS"}
	if !ValidateReport(valid) {
		t.Error("should be valid")
	}

	invalid := &Report{Agent: "", Status: "SUCCESS"}
	if ValidateReport(invalid) {
		t.Error("should be invalid - no agent")
	}
}
