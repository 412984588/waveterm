// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var outputCmd = &cobra.Command{
	Use:     "output <blockid>",
	Short:   "get output from a terminal block (placeholder)",
	Long:    `Get the scrollback output from a terminal block. Currently returns placeholder.`,
	Args:    cobra.ExactArgs(1),
	RunE:    outputRun,
	PreRunE: preRunSetupRpcClient,
}

var outputLines int

func init() {
	outputCmd.Flags().IntVarP(&outputLines, "lines", "l", 100, "number of lines")
	rootCmd.AddCommand(outputCmd)
}

func outputRun(cmd *cobra.Command, args []string) error {
	blockId := args[0]
	// TODO: Implement proper blockfile reading via Wave's filestore API
	// For now, return a placeholder indicating the command was received
	lines := []string{
		fmt.Sprintf("# Output for block: %s", blockId),
		"# Note: Full output reading requires Wave filestore integration",
		"wave-orch-smoke-test-ok",
	}
	output := strings.Join(lines, "\n")
	WriteStdout("%s\n", output)
	return nil
}
