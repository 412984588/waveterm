// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package waveorch

import (
	"encoding/json"
	"regexp"
	"strings"
)

// Report Agent 输出的结构化报告
type Report struct {
	ProjectID    string       `json:"project_id"`
	Agent        string       `json:"agent"`
	Round        int          `json:"round"`
	Status       string       `json:"status"` // SUCCESS, FAIL, BLOCKED, PARTIAL
	Summary      string       `json:"summary"`
	Actions      []string     `json:"actions,omitempty"`
	FilesChanged []FileChange `json:"files_changed,omitempty"`
	CommandsRun  []string     `json:"commands_run,omitempty"`
	Tests        *TestResult  `json:"tests,omitempty"`
	Risks        []string     `json:"risks,omitempty"`
	NeedsHuman   bool         `json:"needs_human"`
	NeedsReason  string       `json:"needs_human_reason,omitempty"`
	NextActions  []string     `json:"next_actions,omitempty"`
}

// FileChange 文件变更记录
type FileChange struct {
	Path    string `json:"path"`
	Action  string `json:"action"` // create, modify, delete
	Summary string `json:"summary,omitempty"`
}

// TestResult 测试结果
type TestResult struct {
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

// ReportParser REPORT 解析器
type ReportParser struct {
	startMarker string
	endMarker   string
}

// NewReportParser 创建解析器
func NewReportParser() *ReportParser {
	return &ReportParser{
		startMarker: "<<<REPORT>>>",
		endMarker:   "<<<END_REPORT>>>",
	}
}

// Parse 从输出中解析 REPORT
func (rp *ReportParser) Parse(output string) (*Report, error) {
	jsonStr := rp.ExtractJSON(output)
	if jsonStr == "" {
		return nil, nil
	}
	var report Report
	if err := json.Unmarshal([]byte(jsonStr), &report); err != nil {
		return nil, err
	}
	return &report, nil
}

// ExtractJSON 从输出中提取 JSON 字符串
func (rp *ReportParser) ExtractJSON(output string) string {
	startIdx := strings.Index(output, rp.startMarker)
	if startIdx == -1 {
		return ""
	}
	searchStart := startIdx + len(rp.startMarker)
	endRel := strings.Index(output[searchStart:], rp.endMarker)
	if endRel == -1 {
		return ""
	}
	endIdx := searchStart + endRel
	if endIdx <= startIdx {
		return ""
	}
	jsonStr := output[searchStart:endIdx]
	return strings.TrimSpace(jsonStr)
}

// ContainsReport 检查输出是否包含 REPORT
func (rp *ReportParser) ContainsReport(output string) bool {
	return strings.Contains(output, rp.startMarker) && strings.Contains(output, rp.endMarker)
}

// ValidationError 验证错误详情
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// ValidateReportStrict 严格验证 Report，返回具体错误
func ValidateReportStrict(r *Report) *ValidationError {
	if r == nil {
		return &ValidationError{Field: "report", Message: "report is nil"}
	}
	if r.ProjectID == "" {
		return &ValidationError{Field: "project_id", Message: "must not be empty"}
	}
	if r.Agent == "" {
		return &ValidationError{Field: "agent", Message: "must not be empty"}
	}
	if r.Round <= 0 {
		return &ValidationError{Field: "round", Message: "must be > 0"}
	}
	if r.Status == "" {
		return &ValidationError{Field: "status", Message: "must not be empty"}
	}
	validStatus := map[string]bool{"SUCCESS": true, "FAIL": true, "BLOCKED": true, "PARTIAL": true}
	if !validStatus[r.Status] {
		return &ValidationError{Field: "status", Message: "must be SUCCESS/FAIL/BLOCKED/PARTIAL"}
	}
	if r.Summary == "" {
		return &ValidationError{Field: "summary", Message: "must not be empty"}
	}
	if r.FilesChanged == nil {
		return &ValidationError{Field: "files_changed", Message: "must be present (can be empty array)"}
	}
	if r.CommandsRun == nil {
		return &ValidationError{Field: "commands_run", Message: "must be present (can be empty array)"}
	}
	if r.NeedsHuman && r.NeedsReason == "" {
		return &ValidationError{Field: "needs_human_reason", Message: "required when needs_human=true"}
	}
	return nil
}

// ValidateReport 验证 Report 必填字段（兼容旧接口）
func ValidateReport(r *Report) bool {
	return ValidateReportStrict(r) == nil
}

// ReportStatusPattern 用于快速检测状态的正则
var ReportStatusPattern = regexp.MustCompile(`"status"\s*:\s*"(SUCCESS|FAIL|BLOCKED|PARTIAL)"`)
