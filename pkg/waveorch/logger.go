// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package waveorch

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Component string    `json:"component"`
	Message   string    `json:"message"`
	Data      any       `json:"data,omitempty"`
}

// Logger 日志记录器
type Logger struct {
	baseDir   string
	retention int // 保留天数
}

// NewLogger 创建日志记录器
func NewLogger(retention int) *Logger {
	if retention <= 0 {
		retention = 7
	}
	home, _ := os.UserHomeDir()
	return &Logger{
		baseDir:   filepath.Join(home, ".wave-orch", "logs"),
		retention: retention,
	}
}

// Log 记录日志
func (l *Logger) Log(level, component, message string, data any) error {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Component: component,
		Message:   Redact(message),
		Data:      RedactAny(data),
	}

	today := time.Now().Format("2006-01-02")
	logDir := filepath.Join(l.baseDir, today)
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return err
	}

	logFile := filepath.Join(logDir, "wave-orch.jsonl")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	jsonData, _ := json.Marshal(entry)
	_, err = f.WriteString(string(jsonData) + "\n")
	return err
}

// CleanOldLogs 清理过期日志
func (l *Logger) CleanOldLogs() error {
	entries, err := os.ReadDir(l.baseDir)
	if err != nil {
		return err
	}

	cutoff := time.Now().AddDate(0, 0, -l.retention)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirDate, err := time.Parse("2006-01-02", entry.Name())
		if err != nil {
			continue
		}
		if dirDate.Before(cutoff) {
			os.RemoveAll(filepath.Join(l.baseDir, entry.Name()))
		}
	}
	return nil
}

// Info 记录信息日志
func (l *Logger) Info(component, message string) error {
	return l.Log("INFO", component, message, nil)
}

// Error 记录错误日志
func (l *Logger) Error(component, message string, err error) error {
	return l.Log("ERROR", component, message, map[string]string{"error": err.Error()})
}
