// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package waveorch

import (
	"testing"
)

func TestRedact_OpenAIKey(t *testing.T) {
	input := "my key is sk-1234567890abcdefghijklmnopqrstuvwxyz123456"
	result := Redact(input)
	if result == input {
		t.Error("OpenAI key should be redacted")
	}
	if !contains(result, "sk-***REDACTED***") {
		t.Errorf("Expected redacted format, got: %s", result)
	}
}

func TestRedact_AnthropicKey(t *testing.T) {
	input := "key: sk-ant-api03-abcdefghijklmnopqrstuvwxyz1234567890"
	result := Redact(input)
	if !contains(result, "sk-ant-***REDACTED***") {
		t.Errorf("Anthropic key not redacted properly: %s", result)
	}
}

func TestRedact_Email(t *testing.T) {
	input := "contact me at test@example.com"
	result := Redact(input)
	if contains(result, "test@example.com") {
		t.Error("Email should be redacted")
	}
	if !contains(result, "***@***.***") {
		t.Errorf("Email not redacted properly: %s", result)
	}
}

func TestRedact_Phone(t *testing.T) {
	input := "call me at 13812345678"
	result := Redact(input)
	if contains(result, "13812345678") {
		t.Error("Phone should be redacted")
	}
}

func TestRedact_PhoneIntl(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"+1-555-123-4567", "+1-***REDACTED***"},
		{"+86 138 1234 5678", "+86***REDACTED***"},
		{"+44.20.7946.0958", "+44***REDACTED***"},
	}
	for _, tt := range tests {
		result := Redact(tt.input)
		if contains(result, tt.input) {
			t.Errorf("Intl phone %s should be redacted", tt.input)
		}
	}
}

func TestRedact_JWT(t *testing.T) {
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
	result := Redact(jwt)
	if contains(result, "eyJzdWIi") {
		t.Errorf("JWT should be redacted, got: %s", result)
	}
	if !contains(result, "eyJ***REDACTED_JWT***") {
		t.Errorf("JWT not redacted properly: %s", result)
	}
}

func TestRedact_Bearer(t *testing.T) {
	input := `Authorization: Bearer eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.rTCH8cLoGxAm_xw68z-zXVKi9ie6xJn9tnVWjd_9ftE`
	result := Redact(input)
	if contains(result, "eyJhbGci") {
		t.Errorf("Bearer token should be redacted: %s", result)
	}
}

func TestRedact_GitLab(t *testing.T) {
	tests := []string{
		"glpat-xxxxxxxxxxxxxxxxxxxx",
		"glptt-abcdefghijklmnopqrstuvwxyz",
	}
	for _, token := range tests {
		result := Redact(token)
		if result == token {
			t.Errorf("GitLab token should be redacted: %s", token)
		}
	}
}

func TestRedact_Slack(t *testing.T) {
	tests := []string{
		"xoxb-REDACTED",
		"xoxp-123456789012-1234567890123",
		"xoxa-2-1234567890",
	}
	for _, token := range tests {
		result := Redact(token)
		if result == token {
			t.Errorf("Slack token should be redacted: %s", token)
		}
		if !contains(result, "xox") {
			t.Errorf("Slack prefix should be preserved: %s", result)
		}
	}
}

func TestRedactMap(t *testing.T) {
	m := map[string]string{
		"api_key":  "sk-1234567890abcdef",
		"username": "testuser",
	}
	result := RedactMap(m)
	if result["api_key"] != "***REDACTED***" {
		t.Error("api_key should be fully redacted")
	}
	if result["username"] != "testuser" {
		t.Error("username should not be redacted")
	}
}

func TestRedactAny_Map(t *testing.T) {
	input := map[string]any{
		"api_key": "sk-1234567890abcdef",
		"note":    "hello",
		"nested": map[string]any{
			"email": "test@example.com",
		},
	}
	out, ok := RedactAny(input).(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", out)
	}
	if out["api_key"] != "***REDACTED***" {
		t.Errorf("api_key should be fully redacted, got: %v", out["api_key"])
	}
	if out["note"] != "hello" {
		t.Errorf("note should remain unchanged, got: %v", out["note"])
	}
	nested, ok := out["nested"].(map[string]any)
	if !ok {
		t.Fatalf("nested should be map, got %T", out["nested"])
	}
	if nested["email"] == "test@example.com" {
		t.Error("email should be redacted")
	}
}

func TestRedactAny_Struct(t *testing.T) {
	type payload struct {
		Token string `json:"token"`
		Note  string `json:"note"`
	}
	input := payload{Token: "sk-1234567890abcdef", Note: "ok"}
	out, ok := RedactAny(input).(map[string]any)
	if !ok {
		t.Fatalf("expected map result, got %T", out)
	}
	if out["token"] == "sk-1234567890abcdef" {
		t.Error("token should be redacted")
	}
	if out["note"] != "ok" {
		t.Errorf("note should remain ok, got: %v", out["note"])
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
