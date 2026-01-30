// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/wavetermdev/waveterm/pkg/wshrpc"
	"github.com/wavetermdev/waveterm/pkg/wshrpc/wshclient"
)

var injectCmd = &cobra.Command{
	Use:     "inject <blockid> <text>",
	Short:   "inject text into a terminal block",
	Long:    `Inject text into a terminal block and optionally send Enter.`,
	Args:    cobra.MinimumNArgs(1),
	RunE:    injectRun,
	PreRunE: preRunSetupRpcClient,
}

var injectNoNewline bool
var injectFile string
var injectWait bool
var injectRetries int

func init() {
	injectCmd.Flags().BoolVarP(&injectNoNewline, "no-newline", "n", false, "do not append newline")
	injectCmd.Flags().StringVarP(&injectFile, "file", "f", "", "read input from file")
	injectCmd.Flags().BoolVarP(&injectWait, "wait", "w", false, "wait for shell ready (retry on failure)")
	injectCmd.Flags().IntVar(&injectRetries, "retries", 5, "max retries when --wait is set")
	rootCmd.AddCommand(injectCmd)
}

func injectRun(cmd *cobra.Command, args []string) error {
	blockId := args[0]

	var text string
	if injectFile != "" {
		data, err := os.ReadFile(injectFile)
		if err != nil {
			return fmt.Errorf("reading file: %w", err)
		}
		text = string(data)
	} else if len(args) > 1 {
		text = args[1]
	} else {
		return fmt.Errorf("no text provided")
	}

	if !injectNoNewline {
		text += "\n"
	}

	inputData := wshrpc.CommandBlockInputData{
		BlockId:     blockId,
		InputData64: base64.StdEncoding.EncodeToString([]byte(text)),
	}

	// Retry logic for --wait flag
	var err error
	maxAttempts := 1
	if injectWait {
		maxAttempts = injectRetries
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = wshclient.ControllerInputCommand(RpcClient, inputData, nil)
		if err == nil {
			WriteStdout("injected %d bytes\n", len(text))
			return nil
		}

		// Check if error is "no shell input chan" - retry if --wait
		if injectWait && strings.Contains(err.Error(), "no shell input chan") {
			if attempt < maxAttempts {
				// Exponential backoff: 200ms, 400ms, 800ms...
				delay := time.Duration(200<<(attempt-1)) * time.Millisecond
				time.Sleep(delay)
				continue
			}
		}
		break
	}

	return fmt.Errorf("inject failed: %w", err)
}
