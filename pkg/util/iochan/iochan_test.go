// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package iochan_test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/wavetermdev/waveterm/pkg/util/iochan"
)

const (
	buflen = 1024
)

func TestIochan_Basic(t *testing.T) {
	srcPipeReader, srcPipeWriter := io.Pipe()
	destPipeReader, destPipeWriter := io.Pipe()
	ctx, cancel := context.WithCancelCause(context.Background())
	t.Cleanup(func() {
		cancel(nil)
		_ = srcPipeReader.Close()
		_ = srcPipeWriter.Close()
		_ = destPipeReader.Close()
		_ = destPipeWriter.Close()
	})

	// Write the packet to the source pipe from a goroutine
	packet := []byte("hello world")
	go func() {
		_, _ = srcPipeWriter.Write(packet)
		_ = srcPipeWriter.Close()
	}()

	// Initialize the reader channel
	readerChanCallbackCalled := make(chan struct{}, 1)
	readerChanCallback := func() {
		_ = srcPipeReader.Close()
		select {
		case readerChanCallbackCalled <- struct{}{}:
		default:
		}
	}
	ioch := iochan.ReaderChan(ctx, srcPipeReader, buflen, readerChanCallback)

	// Initialize the destination pipe and the writer channel
	writerChanCallbackCalled := make(chan struct{}, 1)
	writerChanCallback := func() {
		_ = destPipeReader.Close()
		_ = destPipeWriter.Close()
		select {
		case writerChanCallbackCalled <- struct{}{}:
		default:
		}
	}
	iochan.WriterChan(ctx, destPipeWriter, ioch, writerChanCallback, cancel)

	// Read the packet from the destination pipe and compare it to the original packet
	buf := make([]byte, buflen)
	n, err := destPipeReader.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n != len(packet) {
		t.Fatalf("Read length mismatch: %d != %d", n, len(packet))
	}
	if string(buf[:n]) != string(packet) {
		t.Fatalf("Read data mismatch: %s != %s", buf[:n], packet)
	}

	// Give the callbacks a chance to run before checking if they were called
	select {
	case <-readerChanCallbackCalled:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("ReaderChan callback not called")
	}
	select {
	case <-writerChanCallbackCalled:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("WriterChan callback not called")
	}
}
