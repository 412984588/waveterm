// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/wavetermdev/waveterm/pkg/waveorch"
)

var waveOrchCmd = &cobra.Command{
	Use:   "wave-orch",
	Short: "Wave-Orch orchestration control",
	Long:  `Control the Wave-Orch multi-agent orchestration system.`,
}

var waveOrchStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "show orchestration status",
	RunE:  waveOrchStatusRun,
}

var waveOrchPauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "pause all automatic injection",
	RunE:  waveOrchPauseRun,
}

var waveOrchResumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "resume automatic injection",
	RunE:  waveOrchResumeRun,
}

var waveOrchCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "cleanup old logs",
	RunE:  waveOrchCleanupRun,
}

var waveOrchDiagnosticCmd = &cobra.Command{
	Use:   "diagnostic",
	Short: "generate environment diagnostic snapshot",
	RunE:  waveOrchDiagnosticRun,
}

var cleanupDays int
var diagnosticProject string

func init() {
	waveOrchCmd.AddCommand(waveOrchStatusCmd)
	waveOrchCmd.AddCommand(waveOrchPauseCmd)
	waveOrchCmd.AddCommand(waveOrchResumeCmd)
	waveOrchCleanupCmd.Flags().IntVar(&cleanupDays, "days", 7, "retention days")
	waveOrchCmd.AddCommand(waveOrchCleanupCmd)
	rootCmd.AddCommand(waveOrchCmd)
}

func getWaveOrchStateDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".wave-orch", "state")
}

func getPausedFilePath() string {
	return filepath.Join(getWaveOrchStateDir(), "paused")
}

func isPaused() bool {
	_, err := os.Stat(getPausedFilePath())
	return err == nil
}

func waveOrchStatusRun(cmd *cobra.Command, args []string) error {
	status := map[string]any{
		"paused": isPaused(),
	}
	data, _ := json.MarshalIndent(status, "", "  ")
	WriteStdout("%s\n", string(data))
	return nil
}

func waveOrchPauseRun(cmd *cobra.Command, args []string) error {
	stateDir := getWaveOrchStateDir()
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	f, err := os.Create(getPausedFilePath())
	if err != nil {
		return fmt.Errorf("create pause file: %w", err)
	}
	f.Close()
	WriteStdout("Wave-Orch paused\n")
	return nil
}

func waveOrchResumeRun(cmd *cobra.Command, args []string) error {
	err := os.Remove(getPausedFilePath())
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove pause file: %w", err)
	}
	WriteStdout("Wave-Orch resumed\n")
	return nil
}

func waveOrchCleanupRun(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	logsDir := filepath.Join(home, ".wave-orch", "logs")

	entries, err := os.ReadDir(logsDir)
	if err != nil {
		if os.IsNotExist(err) {
			WriteStdout("No logs to clean\n")
			return nil
		}
		return err
	}

	cleaned := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, _ := entry.Info()
		if info.ModTime().Before(cutoffTime(cleanupDays)) {
			os.RemoveAll(filepath.Join(logsDir, entry.Name()))
			cleaned++
		}
	}
	WriteStdout("Cleaned %d old log directories\n", cleaned)
	return nil
}

func cutoffTime(days int) time.Time {
	return time.Now().AddDate(0, 0, -days)
}
