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
