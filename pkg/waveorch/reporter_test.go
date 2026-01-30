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

func TestReportParser_ExtractJSON_MultipleReports(t *testing.T) {
	parser := NewReportParser()

	input := `noise
<<<REPORT>>>
{"agent":"a","status":"SUCCESS"}
<<<END_REPORT>>>
more noise
<<<REPORT>>>
{"agent":"b","status":"FAIL"}
<<<END_REPORT>>>`

	json := parser.ExtractJSON(input)
	expected := `{"agent":"a","status":"SUCCESS"}`
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
	// 完整有效的 Report
	valid := &Report{
		ProjectID:    "test-project",
		Agent:        "test",
		Round:        1,
		Status:       "SUCCESS",
		Summary:      "Task completed",
		FilesChanged: []FileChange{},
		CommandsRun:  []string{},
	}
	if !ValidateReport(valid) {
		t.Error("should be valid")
	}

	// 旧测试：空 agent
	invalid := &Report{Agent: "", Status: "SUCCESS"}
	if ValidateReport(invalid) {
		t.Error("should be invalid - no agent")
	}
}

func TestValidateReportStrict(t *testing.T) {
	tests := []struct {
		name      string
		report    *Report
		wantField string
	}{
		{"nil report", nil, "report"},
		{"empty project_id", &Report{ProjectID: "", Agent: "a", Round: 1, Status: "SUCCESS", Summary: "s", FilesChanged: []FileChange{}, CommandsRun: []string{}}, "project_id"},
		{"empty agent", &Report{ProjectID: "p", Agent: "", Round: 1, Status: "SUCCESS", Summary: "s", FilesChanged: []FileChange{}, CommandsRun: []string{}}, "agent"},
		{"round zero", &Report{ProjectID: "p", Agent: "a", Round: 0, Status: "SUCCESS", Summary: "s", FilesChanged: []FileChange{}, CommandsRun: []string{}}, "round"},
		{"negative round", &Report{ProjectID: "p", Agent: "a", Round: -1, Status: "SUCCESS", Summary: "s", FilesChanged: []FileChange{}, CommandsRun: []string{}}, "round"},
		{"empty status", &Report{ProjectID: "p", Agent: "a", Round: 1, Status: "", Summary: "s", FilesChanged: []FileChange{}, CommandsRun: []string{}}, "status"},
		{"invalid status", &Report{ProjectID: "p", Agent: "a", Round: 1, Status: "INVALID", Summary: "s", FilesChanged: []FileChange{}, CommandsRun: []string{}}, "status"},
		{"empty summary", &Report{ProjectID: "p", Agent: "a", Round: 1, Status: "SUCCESS", Summary: "", FilesChanged: []FileChange{}, CommandsRun: []string{}}, "summary"},
		{"nil files_changed", &Report{ProjectID: "p", Agent: "a", Round: 1, Status: "SUCCESS", Summary: "s", FilesChanged: nil, CommandsRun: []string{}}, "files_changed"},
		{"nil commands_run", &Report{ProjectID: "p", Agent: "a", Round: 1, Status: "SUCCESS", Summary: "s", FilesChanged: []FileChange{}, CommandsRun: nil}, "commands_run"},
		{"needs_human without reason", &Report{ProjectID: "p", Agent: "a", Round: 1, Status: "SUCCESS", Summary: "s", FilesChanged: []FileChange{}, CommandsRun: []string{}, NeedsHuman: true, NeedsReason: ""}, "needs_human_reason"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateReportStrict(tt.report)
			if err == nil {
				t.Errorf("expected error for field %s", tt.wantField)
				return
			}
			if err.Field != tt.wantField {
				t.Errorf("got field %q, want %q", err.Field, tt.wantField)
			}
		})
	}

	// 测试有效 Report
	valid := &Report{
		ProjectID:    "test-project",
		Agent:        "claude",
		Round:        1,
		Status:       "SUCCESS",
		Summary:      "Done",
		FilesChanged: []FileChange{},
		CommandsRun:  []string{},
	}
	if err := ValidateReportStrict(valid); err != nil {
		t.Errorf("valid report failed: %v", err)
	}

	// 测试 needs_human=true 带 reason
	withReason := &Report{
		ProjectID:    "p",
		Agent:        "a",
		Round:        1,
		Status:       "BLOCKED",
		Summary:      "Need help",
		FilesChanged: []FileChange{},
		CommandsRun:  []string{},
		NeedsHuman:   true,
		NeedsReason:  "Cannot resolve conflict",
	}
	if err := ValidateReportStrict(withReason); err != nil {
		t.Errorf("report with reason failed: %v", err)
	}
}
