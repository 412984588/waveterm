// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/wavetermdev/waveterm/pkg/wshrpc"
	"github.com/wavetermdev/waveterm/pkg/wshrpc/wshclient"
)

var waitCmd = &cobra.Command{
	Use:     "wait <blockid>",
	Short:   "wait for pattern in terminal output",
	Long:    `Wait until a pattern appears in terminal output.`,
	Args:    cobra.ExactArgs(1),
	RunE:    waitRun,
	PreRunE: preRunSetupRpcClient,
}

var waitPattern string
var waitTimeout int
var waitInterval int

func init() {
	waitCmd.Flags().StringVarP(&waitPattern, "pattern", "p", "", "pattern to wait for")
	waitCmd.Flags().IntVarP(&waitTimeout, "timeout", "t", 420000, "timeout in ms")
	waitCmd.Flags().IntVar(&waitInterval, "interval", 1000, "poll interval in ms")
	rootCmd.AddCommand(waitCmd)
}

func waitRun(cmd *cobra.Command, args []string) error {
	if waitPattern == "" {
		return fmt.Errorf("--pattern is required")
	}

	blockId := args[0]
	re, err := regexp.Compile(waitPattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	deadline := time.Now().Add(time.Duration(waitTimeout) * time.Millisecond)
	ticker := time.NewTicker(time.Duration(waitInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for pattern")
			}

			data := wshrpc.CommandTermGetScrollbackLinesData{
				LineStart: -500,
				LineEnd:   0,
			}
			result, err := wshclient.TermGetScrollbackLinesCommand(RpcClient, data, &wshrpc.RpcOpts{
				Route: "block:" + blockId,
			})
			if err != nil {
				continue
			}

			output := strings.Join(result.Lines, "\n")
			if re.MatchString(output) {
				WriteStdout("pattern found\n")
				return nil
			}
		}
	}
}
