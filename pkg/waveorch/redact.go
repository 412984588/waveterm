// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

// Package waveorch implements the Wave-Orch multi-agent orchestration system
package waveorch

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// 敏感信息正则模式
var (
	// OpenAI API Key: sk-xxx (48+ chars)
	openAIKeyPattern = regexp.MustCompile(`sk-[a-zA-Z0-9]{20,}`)

	// Anthropic API Key: sk-ant-xxx
	anthropicKeyPattern = regexp.MustCompile(`sk-ant-[a-zA-Z0-9\-]{20,}`)

	// Generic API key patterns
	genericKeyPattern = regexp.MustCompile(`(?i)(api[_-]?key|token|secret|password|credential)["\s:=]+["']?([a-zA-Z0-9_\-]{16,})["']?`)

	// Email pattern
	emailPattern = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)

	// Chinese phone number
	phonePatternCN = regexp.MustCompile(`1[3-9]\d{9}`)

	// International phone: +1-xxx, +86-xxx, +44-xxx etc (various separators)
	phonePatternIntl = regexp.MustCompile(`\+\d{1,3}[-.\s]?\d{2,4}[-.\s]?\d{2,4}[-.\s]?\d{2,4}[-.\s]?\d{0,4}`)

	// AWS Access Key ID
	awsKeyPattern = regexp.MustCompile(`AKIA[0-9A-Z]{16}`)

	// AWS Secret Access Key (40 chars base64-like)
	awsSecretPattern = regexp.MustCompile(`(?i)(aws_secret_access_key|aws_secret)["\s:=]+["']?([A-Za-z0-9/+=]{40})["']?`)

	// GitHub tokens
	githubTokenPattern = regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{36,}`)

	// GitLab tokens: glpat-xxx, glptt-xxx
	gitlabTokenPattern = regexp.MustCompile(`gl(pat|ptt)-[A-Za-z0-9\-_]{20,}`)

	// Slack tokens: xoxb-, xoxp-, xoxa-, xoxs-
	slackTokenPattern = regexp.MustCompile(`xox[bpas]-[A-Za-z0-9\-]{10,}`)

	// JWT/Bearer tokens in Authorization header
	bearerPattern = regexp.MustCompile(`(?i)(Bearer|Authorization)["\s:=]+["']?([A-Za-z0-9\-_]+\.[A-Za-z0-9\-_]+\.[A-Za-z0-9\-_]+)["']?`)

	// Generic JWT pattern (three base64 parts separated by dots)
	jwtPattern = regexp.MustCompile(`eyJ[A-Za-z0-9\-_]+\.eyJ[A-Za-z0-9\-_]+\.[A-Za-z0-9\-_]+`)
)

// Redact 对字符串进行脱敏处理
func Redact(input string) string {
	result := input

	// JWT tokens (must be before Bearer to avoid double processing)
	result = jwtPattern.ReplaceAllString(result, "eyJ***REDACTED_JWT***")

	// Bearer/Authorization header with JWT
	result = bearerPattern.ReplaceAllString(result, "${1}=***REDACTED_TOKEN***")

	// OpenAI keys: 保留前缀
	result = openAIKeyPattern.ReplaceAllStringFunc(result, func(s string) string {
		return "sk-***REDACTED***"
	})

	// Anthropic keys: 保留前缀
	result = anthropicKeyPattern.ReplaceAllStringFunc(result, func(s string) string {
		return "sk-ant-***REDACTED***"
	})

	// AWS Access Key ID
	result = awsKeyPattern.ReplaceAllString(result, "AKIA***REDACTED***")

	// AWS Secret Access Key
	result = awsSecretPattern.ReplaceAllString(result, "${1}=***REDACTED***")

	// GitHub tokens
	result = githubTokenPattern.ReplaceAllStringFunc(result, func(s string) string {
		if len(s) > 4 {
			return s[:4] + "_***REDACTED***"
		}
		return "***REDACTED***"
	})

	// GitLab tokens
	result = gitlabTokenPattern.ReplaceAllStringFunc(result, func(s string) string {
		if len(s) > 6 {
			return s[:6] + "-***REDACTED***"
		}
		return "***REDACTED***"
	})

	// Slack tokens
	result = slackTokenPattern.ReplaceAllStringFunc(result, func(s string) string {
		if len(s) > 5 {
			return s[:5] + "-***REDACTED***"
		}
		return "***REDACTED***"
	})

	// Generic keys (处理 key=value 格式)
	result = genericKeyPattern.ReplaceAllString(result, "${1}=***REDACTED***")

	// Emails: 保留域名结构
	result = emailPattern.ReplaceAllString(result, "***@***.***")

	// International phone numbers (before CN to avoid partial match)
	result = phonePatternIntl.ReplaceAllStringFunc(result, func(s string) string {
		if len(s) >= 3 {
			return s[:3] + "***REDACTED***"
		}
		return "***REDACTED***"
	})

	// Chinese phone numbers
	result = phonePatternCN.ReplaceAllStringFunc(result, func(s string) string {
		if len(s) >= 3 {
			return s[:3] + "********"
		}
		return "***REDACTED***"
	})

	return result
}

// RedactMap 对 map 中的值进行脱敏
func RedactMap(m map[string]string) map[string]string {
	result := make(map[string]string, len(m))
	for k, v := range m {
		// 对敏感 key 名称的值进行脱敏
		lowerKey := strings.ToLower(k)
		if containsSensitiveKeyword(lowerKey) {
			result[k] = "***REDACTED***"
		} else {
			result[k] = Redact(v)
		}
	}
	return result
}

// RedactAny 对任意类型进行脱敏处理，返回 JSON 友好的结构
func RedactAny(input any) any {
	if input == nil {
		return nil
	}
	switch v := input.(type) {
	case string:
		return Redact(v)
	case []byte:
		return Redact(string(v))
	case error:
		return Redact(v.Error())
	case map[string]string:
		return RedactMap(v)
	case map[string]any:
		return redactInterface(v)
	case []string:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, Redact(item))
		}
		return out
	case []any:
		return redactInterface(v)
	}

	// Fallback: marshal/unmarshal to JSON-friendly types, then redact recursively.
	b, err := json.Marshal(input)
	if err != nil {
		return Redact(fmt.Sprintf("%v", input))
	}
	var decoded any
	if err := json.Unmarshal(b, &decoded); err != nil {
		return Redact(fmt.Sprintf("%v", input))
	}
	return redactInterface(decoded)
}

func redactInterface(input any) any {
	switch v := input.(type) {
	case string:
		return Redact(v)
	case []any:
		out := make([]any, 0, len(v))
		for _, item := range v {
			out = append(out, redactInterface(item))
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, val := range v {
			if containsSensitiveKeyword(strings.ToLower(k)) {
				out[k] = "***REDACTED***"
				continue
			}
			out[k] = redactInterface(val)
		}
		return out
	default:
		return v
	}
}

// containsSensitiveKeyword 检查 key 名称是否包含敏感关键词
func containsSensitiveKeyword(key string) bool {
	sensitiveKeywords := []string{
		"key", "token", "secret", "password", "credential",
		"auth", "api_key", "apikey", "access_token",
	}
	for _, kw := range sensitiveKeywords {
		if strings.Contains(key, kw) {
			return true
		}
	}
	return false
}
