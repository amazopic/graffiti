package integrate

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// HookCommand is the command Claude Code runs for the PreToolUse hook. It relies
// on `graffiti` being on PATH (the install script puts it there), so the entry is
// portable across machines and safe to commit to a repo's .claude/settings.json.
const HookCommand = "graffiti hook"

// hookEntry is graffiti's PreToolUse matcher entry. We match the Grep and Glob
// tools (the common "search the code" path); Bash-grep detection is deferred.
func hookEntry() map[string]any {
	return map[string]any{
		"matcher": "Grep|Glob",
		"hooks": []any{
			map[string]any{"type": "command", "command": HookCommand},
		},
	}
}

// entryHasGraffitiHook reports whether a PreToolUse entry already contains our hook
// (identified by its command), so re-runs don't duplicate it.
func entryHasGraffitiHook(entry any) bool {
	m, ok := entry.(map[string]any)
	if !ok {
		return false
	}
	hooks, ok := m["hooks"].([]any)
	if !ok {
		return false
	}
	for _, h := range hooks {
		hm, ok := h.(map[string]any)
		if !ok {
			continue
		}
		if cmd, _ := hm["command"].(string); cmd == HookCommand {
			return true
		}
	}
	return false
}

// MergeHookSettings inserts graffiti's PreToolUse hook into an existing
// settings.json body (which may be empty), preserving every other key and hook.
// Idempotent. Returns an error for malformed JSON so the caller never clobbers a
// file it could not parse.
func MergeHookSettings(existing []byte) ([]byte, error) {
	root := map[string]any{}
	if len(bytes.TrimSpace(existing)) > 0 {
		if err := json.Unmarshal(existing, &root); err != nil {
			return nil, fmt.Errorf("settings.json is not valid JSON: %w", err)
		}
	}
	hooks, ok := root["hooks"].(map[string]any)
	if !ok {
		hooks = map[string]any{}
		root["hooks"] = hooks
	}
	pre, _ := hooks["PreToolUse"].([]any)
	for _, e := range pre {
		if entryHasGraffitiHook(e) {
			return marshalSettings(root) // already installed → no-op (stable re-marshal)
		}
	}
	hooks["PreToolUse"] = append(pre, hookEntry())
	return marshalSettings(root)
}

// marshalSettings pretty-prints with 2-space indent and a trailing newline. Go's
// json.MarshalIndent emits object keys in sorted order → deterministic output.
func marshalSettings(v any) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}
