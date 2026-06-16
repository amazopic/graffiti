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
	code := run([]string{"graffiti"}, bytes.NewReader(nil), &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(errOut.String(), "usage: graffiti") {
		t.Fatalf("expected usage on stderr, got %q", errOut.String())
	}
}

func TestRun_UnknownCommand_Errors(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "frobnicate"}, bytes.NewReader(nil), &out, &errOut)
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
	code := run([]string{"graffiti", "build", dir}, bytes.NewReader(nil), &out, &errOut)
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
	if code := run([]string{"graffiti", "build", dir}, bytes.NewReader(nil), &out, &errOut); code != 0 {
		t.Fatalf("build failed (%d): %s", code, errOut.String())
	}
	return dir
}

func TestRun_QueryPrintsSubgraph(t *testing.T) {
	dir := buildTempRepo(t)
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "query", "login handler", dir}, bytes.NewReader(nil), &out, &errOut)
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
	code := run([]string{"graffiti", "query", "x", t.TempDir()}, bytes.NewReader(nil), &out, &errOut)
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

func TestRun_ServeViaRun(t *testing.T) {
	dir := buildTempRepo(t)
	initLine := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}` + "\n"
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "serve", dir}, strings.NewReader(initLine), &out, &errOut)
	if code != 0 {
		t.Fatalf("serve via run exit code = %d (stderr=%q)", code, errOut.String())
	}
	if !strings.Contains(out.String(), `"protocolVersion":"2025-06-18"`) {
		t.Fatalf("serve via run initialize response missing version echo:\n%s", out.String())
	}
}

func TestRun_UpdateRebuilds(t *testing.T) {
	dir := buildTempRepo(t) // builds once; update should rebuild cleanly
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "update", dir}, bytes.NewReader(nil), &out, &errOut)
	if code != 0 {
		t.Fatalf("update exit code = %d (stderr=%q)", code, errOut.String())
	}
	if !strings.Contains(out.String(), "Done. 0 API calls, $0.") {
		t.Fatalf("update missing success line, got %q", out.String())
	}
}

func TestRun_QueryTooManyArgs_Errors(t *testing.T) {
	dir := buildTempRepo(t)
	var out, errOut bytes.Buffer
	// "login", "handler", and dir are three positional args after "query" — too many.
	code := run([]string{"graffiti", "query", "login", "handler", dir}, bytes.NewReader(nil), &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d (stderr=%q)", code, errOut.String())
	}
	if !strings.Contains(errOut.String(), "quote") {
		t.Fatalf("expected quoting hint in error, got %q", errOut.String())
	}
}

func TestRun_InitProject(t *testing.T) {
	dir := t.TempDir()
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "init", "--root", dir}, bytes.NewReader(nil), &out, &errOut)
	if code != 0 {
		t.Fatalf("init exit code = %d (stderr=%q)", code, errOut.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "skills", "graffiti", "SKILL.md")); err != nil {
		t.Fatalf("SKILL.md not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err != nil {
		t.Fatalf("CLAUDE.md not written: %v", err)
	}
	// no --hook → no settings.json
	if _, err := os.Stat(filepath.Join(dir, ".claude", "settings.json")); !os.IsNotExist(err) {
		t.Fatal("settings.json should not exist without --hook")
	}
	if !strings.Contains(out.String(), "graffiti wired into Claude Code") {
		t.Fatalf("missing success line:\n%s", out.String())
	}
}

func TestRun_InitWithHook(t *testing.T) {
	dir := t.TempDir()
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "init", "--hook", "--root", dir}, bytes.NewReader(nil), &out, &errOut)
	if code != 0 {
		t.Fatalf("init --hook exit code = %d (stderr=%q)", code, errOut.String())
	}
	b, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("settings.json not written: %v", err)
	}
	if !strings.Contains(string(b), "graffiti hook") {
		t.Fatalf("settings.json missing hook:\n%s", b)
	}
}

func TestRun_InitIdempotent(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 2; i++ {
		var out, errOut bytes.Buffer
		if code := run([]string{"graffiti", "init", "--hook", "--root", dir}, bytes.NewReader(nil), &out, &errOut); code != 0 {
			t.Fatalf("init run %d failed: %s", i, errOut.String())
		}
	}
	b, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if n := strings.Count(string(b), "graffiti:start"); n != 1 {
		t.Fatalf("CLAUDE.md should have exactly one block, got %d", n)
	}
}

func TestRun_HookNudgeWhenMapPresent(t *testing.T) {
	dir := buildTempRepo(t) // writes .graffiti/map.json
	event := `{"hook_event_name":"PreToolUse","tool_name":"Grep","cwd":"` + dir + `"}`
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "hook"}, strings.NewReader(event), &out, &errOut)
	if code != 0 {
		t.Fatalf("hook exit code = %d (stderr=%q)", code, errOut.String())
	}
	if !strings.Contains(out.String(), "additionalContext") || !strings.Contains(out.String(), "graffiti query") {
		t.Fatalf("expected a nudge, got %q", out.String())
	}
}

func TestRun_HookSilentWhenNoMap(t *testing.T) {
	dir := t.TempDir()
	event := `{"hook_event_name":"PreToolUse","tool_name":"Grep","cwd":"` + dir + `"}`
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "hook"}, strings.NewReader(event), &out, &errOut)
	if code != 0 {
		t.Fatalf("hook exit code = %d", code)
	}
	if out.Len() != 0 {
		t.Fatalf("hook should be silent without a map, got %q", out.String())
	}
}
