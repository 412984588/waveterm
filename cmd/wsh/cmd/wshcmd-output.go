// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wavetermdev/waveterm/pkg/wshrpc"
	"github.com/wavetermdev/waveterm/pkg/wshrpc/wshclient"
)

var outputCmd = &cobra.Command{
	Use:     "output <blockid>",
	Short:   "get output from a terminal block",
	Long:    `Get the scrollback output from a terminal block.`,
	Args:    cobra.ExactArgs(1),
	RunE:    outputRun,
	PreRunE: preRunSetupRpcClient,
}

var outputLines int
var outputLastCommand bool

func init() {
	outputCmd.Flags().IntVarP(&outputLines, "lines", "l", 100, "number of lines")
	outputCmd.Flags().BoolVar(&outputLastCommand, "last-command", false, "get last command output")
	rootCmd.AddCommand(outputCmd)
}

func outputRun(cmd *cobra.Command, args []string) error {
	blockId := args[0]

	data := wshrpc.CommandTermGetScrollbackLinesData{
		LineStart:   -outputLines,
		LineEnd:     0,
		LastCommand: outputLastCommand,
	}

	result, err := wshclient.TermGetScrollbackLinesCommand(RpcClient, data, &wshrpc.RpcOpts{
		Route: "block:" + blockId,
	})
	if err != nil {
		return fmt.Errorf("get output failed: %w", err)
	}

	output := strings.Join(result.Lines, "\n")
	WriteStdout("%s\n", output)
	return nil
}
