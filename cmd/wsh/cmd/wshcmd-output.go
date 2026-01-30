// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/wavetermdev/waveterm/pkg/wshrpc"
	"github.com/wavetermdev/waveterm/pkg/wshrpc/wshclient"
	"github.com/wavetermdev/waveterm/pkg/wshutil"
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

func init() {
	outputCmd.Flags().IntVarP(&outputLines, "lines", "l", 100, "number of lines")
	rootCmd.AddCommand(outputCmd)
}

func outputRun(cmd *cobra.Command, args []string) error {
	blockId := args[0]

	// Use TermGetScrollbackLinesCommand via RPC
	data := wshrpc.CommandTermGetScrollbackLinesData{
		LineStart: 0,
		LineEnd:   outputLines,
	}

	route := wshutil.MakeFeBlockRouteId(blockId)
	result, err := wshclient.TermGetScrollbackLinesCommand(
		RpcClient,
		data,
		&wshrpc.RpcOpts{Route: route},
	)
	if err != nil {
		return err
	}

	// Join lines into output
	content := strings.Join(result.Lines, "\n")
	WriteStdout("%s\n", content)
	return nil
}
