package integrate

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

// hookNudge is the additionalContext text injected when a map exists.
const hookNudge = "This repo has a graffiti code map (.graffiti/map.json). " +
	"For codebase-structure questions, prefer `graffiti query \"<question>\"` over raw grep/glob — " +
	"it returns a scoped, ranked subgraph instead of raw matches."

// hookResponse is the PreToolUse stdout contract for a non-blocking nudge.
// No permissionDecision field is set, so the tool always proceeds.
type hookResponse struct {
	HookSpecificOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	} `json:"hookSpecificOutput"`
}

// RunHook implements the `graffiti hook` PreToolUse handler. It is deliberately
// total and non-blocking: it NEVER returns an error and NEVER blocks the tool.
// When .graffiti/map.json exists (relative to the event's cwd, else fallbackCwd),
// it writes a JSON nudge to out; otherwise it writes nothing (defer to normal flow).
// Any malformed input simply yields no output.
func RunHook(in io.Reader, out io.Writer, fallbackCwd string) {
	cwd := fallbackCwd
	if b, err := io.ReadAll(io.LimitReader(in, 1<<20)); err == nil && len(bytes.TrimSpace(b)) > 0 {
		var ev struct {
			Cwd string `json:"cwd"`
		}
		if json.Unmarshal(b, &ev) == nil && ev.Cwd != "" {
			cwd = ev.Cwd
		}
	}
	if cwd == "" {
		return
	}
	if _, err := os.Stat(filepath.Join(cwd, ".graffiti", "map.json")); err != nil {
		return // no map → emit nothing
	}
	var resp hookResponse
	resp.HookSpecificOutput.HookEventName = "PreToolUse"
	resp.HookSpecificOutput.AdditionalContext = hookNudge
	b, err := json.Marshal(&resp)
	if err != nil {
		return
	}
	_, _ = out.Write(b)
}
