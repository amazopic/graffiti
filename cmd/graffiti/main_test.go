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

func buildTempRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	src := "package auth\n\n// LoginHandler authenticates a request.\nfunc LoginHandler() {}\n\nfunc createSession() {}\n"
	if err := os.WriteFile(filepath.Join(dir, "auth.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := run([]string{"graffiti", "build", dir}, &out, &errOut); code != 0 {
		t.Fatalf("build failed (%d): %s", code, errOut.String())
	}
	return dir
}

func TestRun_QueryPrintsSubgraph(t *testing.T) {
	dir := buildTempRepo(t)
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "query", "login handler", dir}, &out, &errOut)
	if code != 0 {
		t.Fatalf("query exit code = %d (stderr=%q)", code, errOut.String())
	}
	if !strings.Contains(out.String(), "NODES") {
		t.Fatalf("query output missing NODES block:\n%s", out.String())
	}
	if !strings.Contains(strings.ToLower(out.String()), "login") {
		t.Fatalf("query output should mention login:\n%s", out.String())
	}
}

func TestRun_QueryMissingMap_Errors(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "query", "x", t.TempDir()}, &out, &errOut)
	if code == 0 {
		t.Fatal("expected non-zero exit when map.json is absent")
	}
	if !strings.Contains(errOut.String(), "graffiti:") {
		t.Fatalf("expected error on stderr, got %q", errOut.String())
	}
}

func TestRun_ServeHandlesInitialize(t *testing.T) {
	dir := buildTempRepo(t)
	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}` + "\n")
	var out, errOut bytes.Buffer
	code := serve(dir, in, &out, &errOut)
	if code != 0 {
		t.Fatalf("serve exit code = %d (stderr=%q)", code, errOut.String())
	}
	if !strings.Contains(out.String(), `"protocolVersion":"2025-06-18"`) {
		t.Fatalf("serve initialize response missing version echo:\n%s", out.String())
	}
}
