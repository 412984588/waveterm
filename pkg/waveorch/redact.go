// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

// Package waveorch implements the Wave-Orch multi-agent orchestration system
package waveorch

import (
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
	phonePattern = regexp.MustCompile(`1[3-9]\d{9}`)

	// AWS keys
	awsKeyPattern = regexp.MustCompile(`AKIA[0-9A-Z]{16}`)

	// GitHub tokens
	githubTokenPattern = regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{36,}`)
)

// Redact 对字符串进行脱敏处理
func Redact(input string) string {
	result := input

	// OpenAI keys: 保留前缀
	result = openAIKeyPattern.ReplaceAllStringFunc(result, func(s string) string {
		return "sk-***REDACTED***"
	})

	// Anthropic keys: 保留前缀
	result = anthropicKeyPattern.ReplaceAllStringFunc(result, func(s string) string {
		return "sk-ant-***REDACTED***"
	})

	// AWS keys
	result = awsKeyPattern.ReplaceAllString(result, "AKIA***REDACTED***")

	// GitHub tokens
	result = githubTokenPattern.ReplaceAllStringFunc(result, func(s string) string {
		if len(s) > 4 {
			return s[:4] + "_***REDACTED***"
		}
		return "***REDACTED***"
	})

	// Generic keys (处理 key=value 格式)
	result = genericKeyPattern.ReplaceAllString(result, "${1}=***REDACTED***")

	// Emails: 保留域名结构
	result = emailPattern.ReplaceAllString(result, "***@***.***")

	// Phone numbers
	result = phonePattern.ReplaceAllStringFunc(result, func(s string) string {
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
