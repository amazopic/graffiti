package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRun_NoArgs_PrintsUsage(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "usage: graffiti") {
		t.Fatalf("expected usage on stderr, got %q", errOut.String())
	}
}

func TestRun_UnknownCommand_Errors(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "frobnicate"}, &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "unknown command") {
		t.Fatalf("expected unknown command error, got %q", errOut.String())
	}
}

// Keep os import referenced even before Task 13 adds the build E2E test.
var _ = os.Args
