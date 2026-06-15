package main

import (
	"bytes"
	"os"
	"path/filepath"
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

func TestRun_BuildPrintsSuccessLine(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "build", dir}, &out, &errOut)
	if code != 0 {
		t.Fatalf("exit code = %d (stderr=%q)", code, errOut.String())
	}
	s := out.String()
	if !strings.Contains(s, "Done. 0 API calls, $0.") {
		t.Fatalf("missing success line, got %q", s)
	}
	if !strings.Contains(s, "The 3 most interesting questions your map can answer:") {
		t.Fatalf("missing questions header, got %q", s)
	}
	if !strings.Contains(s, "1) ") || !strings.Contains(s, "2) ") || !strings.Contains(s, "3) ") {
		t.Fatalf("expected 3 numbered questions, got %q", s)
	}
}
