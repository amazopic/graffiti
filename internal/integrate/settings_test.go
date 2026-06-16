package integrate

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestMergeHookSettings_CreateFromEmpty(t *testing.T) {
	out, err := MergeHookSettings(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out, []byte("PreToolUse")) || !bytes.Contains(out, []byte(HookCommand)) {
		t.Fatalf("missing PreToolUse/hook command:\n%s", out)
	}
	var v map[string]any
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}
}

func TestMergeHookSettings_Idempotent(t *testing.T) {
	once, err := MergeHookSettings(nil)
	if err != nil {
		t.Fatal(err)
	}
	twice, err := MergeHookSettings(once)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(once, twice) {
		t.Fatalf("merge must be idempotent\nonce:\n%s\ntwice:\n%s", once, twice)
	}
	if n := bytes.Count(twice, []byte(`"`+HookCommand+`"`)); n != 1 {
		t.Fatalf("expected exactly one hook command, got %d", n)
	}
}

func TestMergeHookSettings_PreservesOtherKeysAndHooks(t *testing.T) {
	in := []byte(`{
  "model": "opus",
  "hooks": {
    "PreToolUse": [
      {"matcher": "Bash", "hooks": [{"type": "command", "command": "echo hi"}]}
    ],
    "PostToolUse": [
      {"matcher": "Edit", "hooks": [{"type": "command", "command": "gofmt -w"}]}
    ]
  }
}`)
	out, err := MergeHookSettings(in)
	if err != nil {
		t.Fatal(err)
	}
	for _, must := range []string{`"model"`, "echo hi", "PostToolUse", "gofmt -w", HookCommand} {
		if !bytes.Contains(out, []byte(must)) {
			t.Fatalf("merge dropped %q:\n%s", must, out)
		}
	}
	var v map[string]any
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatal(err)
	}
	arr := v["hooks"].(map[string]any)["PreToolUse"].([]any)
	if len(arr) != 2 {
		t.Fatalf("PreToolUse should have 2 entries, got %d", len(arr))
	}
	// idempotent on an already-merged file
	out2, _ := MergeHookSettings(out)
	if !bytes.Equal(out, out2) {
		t.Fatal("merge not idempotent over a pre-populated file")
	}
}

func TestMergeHookSettings_MalformedErrors(t *testing.T) {
	if _, err := MergeHookSettings([]byte("{not json")); err == nil {
		t.Fatal("expected an error for malformed JSON (must not clobber the file)")
	}
}
