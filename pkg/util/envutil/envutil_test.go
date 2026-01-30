// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package envutil

import (
	"reflect"
	"testing"
)

func TestEnvKeysSorted(t *testing.T) {
	env := map[string]string{
		"Z_VAR": "z",
		"A_VAR": "a",
		"B_VAR": "b",
	}
	got := EnvKeys(env)
	want := []string{"A_VAR", "B_VAR", "Z_VAR"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("EnvKeys returned %v, want %v", got, want)
	}
}
