package integrate

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeMap creates a minimal .graffiti/map.json under dir.
func writeMap(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, ".graffiti"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".graffiti", "map.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunHook_NudgesWhenMapExists(t *testing.T) {
	dir := t.TempDir()
	writeMap(t, dir)
	event := `{"hook_event_name":"PreToolUse","tool_name":"Grep","cwd":` + jsonString(dir) + `}`
	var out bytes.Buffer
	RunHook(strings.NewReader(event), &out, "/nonexistent-fallback")
	if out.Len() == 0 {
		t.Fatal("expected a nudge when a map exists")
	}
	var resp struct {
		HookSpecificOutput struct {
			HookEventName     string `json:"hookEventName"`
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("nudge is not valid JSON: %v\n%s", err, out.String())
	}
	if resp.HookSpecificOutput.HookEventName != "PreToolUse" {
		t.Fatalf("hookEventName = %q", resp.HookSpecificOutput.HookEventName)
	}
	if !strings.Contains(resp.HookSpecificOutput.AdditionalContext, "graffiti query") {
		t.Fatalf("nudge should mention graffiti query, got %q", resp.HookSpecificOutput.AdditionalContext)
	}
	// Never blocks: no permissionDecision field anywhere in the output.
	if strings.Contains(out.String(), "permissionDecision") {
		t.Fatal("hook must never emit a permissionDecision (must not block)")
	}
}

func TestRunHook_SilentWhenNoMap(t *testing.T) {
	dir := t.TempDir() // no .graffiti/map.json
	event := `{"hook_event_name":"PreToolUse","tool_name":"Grep","cwd":` + jsonString(dir) + `}`
	var out bytes.Buffer
	RunHook(strings.NewReader(event), &out, "/nonexistent-fallback")
	if out.Len() != 0 {
		t.Fatalf("expected no output when no map exists, got %q", out.String())
	}
}

func TestRunHook_UsesFallbackCwdWhenEventOmitsIt(t *testing.T) {
	dir := t.TempDir()
	writeMap(t, dir)
	// Event without cwd → handler must fall back to the provided cwd.
	var out bytes.Buffer
	RunHook(strings.NewReader(`{"tool_name":"Glob"}`), &out, dir)
	if out.Len() == 0 {
		t.Fatal("expected a nudge via fallback cwd")
	}
}

func TestRunHook_NeverBreaksOnBadInput(t *testing.T) {
	var out bytes.Buffer
	// Garbage stdin + a fallback cwd with no map → emit nothing, do not panic.
	RunHook(strings.NewReader("not json at all"), &out, t.TempDir())
	if out.Len() != 0 {
		t.Fatalf("bad input should produce no output, got %q", out.String())
	}
}

// jsonString quotes s as a JSON string literal (handles Windows backslashes).
func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
