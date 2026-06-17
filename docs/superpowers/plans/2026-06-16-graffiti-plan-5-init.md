# graffiti Plan 5 — Claude Code integration (`graffiti init`)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `graffiti init` wires graffiti into Claude Code — a short auto-invoked Skill, an idempotent always-on `CLAUDE.md` block, and an optional non-blocking PreToolUse hook that nudges grep → `graffiti query` — so the assistant prefers the map automatically (spec §9).

**Architecture:** A new pure-Go package `internal/integrate` owns all generated content and the merge logic. Three artifacts: (1) `.claude/skills/graffiti/SKILL.md` (we own this path — overwrite), (2) `CLAUDE.md` block delimited by `<!-- graffiti:start -->`/`<!-- graffiti:end -->` (surgical replace-or-append, preserving everything outside the markers), (3) `.claude/settings.json` PreToolUse entry merged into existing JSON without clobbering other keys/hooks (only with `--hook`). The hook is the binary itself: a hidden `graffiti hook` subcommand reads the PreToolUse event on stdin and, when `.graffiti/map.json` exists, emits a non-blocking `additionalContext` nudge — and **never blocks the tool under any error**. All generated content is constant text → byte-deterministic; merges are idempotent (running `init` twice is a no-op). The CLI computes target paths from `--user`/`--hook` flags and hands explicit paths to `integrate.Install`, so every test runs against `t.TempDir()` and never touches the real `~/.claude`.

**Tech Stack:** Go 1.26, stdlib only (`encoding/json`, `bytes`, `os`, `path/filepath`, `io`). No new dependencies. Build tags unchanged (`grammar_subset grammar_subset_go grammar_subset_gomod`).

---

## Verified Claude Code contracts (research, 2026-06-16)

These were verified against the current Claude Code docs and are hard-coded into the generated artifacts and tests:

- **Skill path:** project `=.claude/skills/<name>/SKILL.md`; user `=~/.claude/skills/<name>/SKILL.md`. Frontmatter `name` + `description` (both optional; `description` drives auto-invocation; ≤1536 chars).
- **Hooks settings file:** project (committed) `=.claude/settings.json`; user `=~/.claude/settings.json`. Shape:
  `{"hooks":{"PreToolUse":[{"matcher":"<tool-name-regex>","hooks":[{"type":"command","command":"<cmd>"}]}]}}`. `matcher` filters by **tool name**; pipe-separated alternatives are literal (`"Grep|Glob"` = Grep OR Glob). Canonical tool names include `Grep`, `Glob`, `Bash`.
- **PreToolUse input (stdin JSON):** `{session_id, transcript_path, cwd, permission_mode, hook_event_name:"PreToolUse", tool_name, tool_input}`.
- **PreToolUse output — non-blocking nudge:** stdout `{"hookSpecificOutput":{"hookEventName":"PreToolUse","additionalContext":"<text>"}}` with **exit 0** and **no** `permissionDecision` → the tool proceeds normally and the text is added to the model's context. Exit 2 would *block* (we never do this). Emitting nothing defers to the normal flow. This means our hook degrades to harmless even if a Claude Code build ignores `additionalContext`.

## File structure

```
internal/integrate/
  target.go        Target struct + ProjectTarget(root)/UserTarget(home) path constructors
  skill.go         SkillContent() — the deterministic SKILL.md body
  claudemd.go      ClaudeBlock() + MergeClaudeMD(existing) — idempotent marker merge
  settings.go      HookCommand const + MergeHookSettings(existing) — idempotent JSON merge
  hook.go          RunHook(in, out, fallbackCwd) — non-blocking PreToolUse handler
  install.go       Options, Action, Result, Install(target, opts) — orchestration + disk writes
  target_test.go  skill_test.go  claudemd_test.go  settings_test.go  hook_test.go  install_test.go
cmd/graffiti/main.go    + `update`, `init`, `hook` cases; usage text
cmd/graffiti/main_test.go   + init/hook/update run() tests
testdata/golden/init/   SKILL.md  CLAUDE.md  settings.json   (from-scratch project install)
internal/render/... (unchanged)   README.md  docs/superpowers/specs/...  (notes)
```

**Constant content (single source of truth — reuse verbatim across tasks):**

`SkillContent()` returns exactly (note the trailing newline):

```text
---
name: graffiti
description: Use when exploring or answering questions about THIS codebase's structure — where something is defined, how components connect, what the architecture looks like. graffiti turns the repo into a queryable code map so you query the graph instead of grepping blind.
---

# graffiti — read the map, don't grep blind

`graffiti` is a CLI that turns this repository into a queryable code map (no API key, $0, offline).

## First time in a repo
1. Run `graffiti build .` — it writes `.graffiti/map.json`, `.graffiti/MAP.md`, and `.graffiti/map.html`.
2. Read `.graffiti/MAP.md`. Tell the user the **god nodes** (most-connected code) and the **surprising connections**, then offer to trace the single most interesting question the map suggests.

## Answering questions about the code
- Prefer `graffiti query "<question>"` over grep/read — it returns a scoped, ranked subgraph (relevant nodes + edges), not raw text matches.
- To locate a symbol, run `graffiti query "<symbol name>"`.

## After editing code
- Run `graffiti update` so later queries reflect the current code.

Keep this lightweight: the deterministic binary does the heavy lifting — build, read MAP.md, query.
```

`ClaudeBlock()` returns exactly (markers + trailing newline):

```text
<!-- graffiti:start -->
## graffiti code map

If `.graffiti/map.json` exists, this repo has a graffiti code map. For questions about the
codebase's structure (where something lives, how parts connect, the architecture), run
`graffiti query "<question>"` instead of grep/read — it returns a scoped subgraph. After
editing code, run `graffiti update` to refresh the map. If no map exists yet, run
`graffiti build .` first.
<!-- graffiti:end -->
```

Hook nudge text (single line, used by `hook.go`):

```text
This repo has a graffiti code map (.graffiti/map.json). For codebase-structure questions, prefer `graffiti query "<question>"` over raw grep/glob — it returns a scoped, ranked subgraph instead of raw matches.
```

---

## Task 1: `graffiti update` (full-rebuild alias)

The generated Skill and CLAUDE.md tell the model to run `graffiti update`. That command must exist and work, or the guidance produces "unknown command" errors. Implement it as a full rebuild now; the incremental AST-only optimization (spec §11) is deferred.

**Files:**
- Modify: `cmd/graffiti/main.go`
- Test: `cmd/graffiti/main_test.go`

- [ ] **Step 1: Write the failing test**

Add to `cmd/graffiti/main_test.go`:

```go
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
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./cmd/graffiti/ -run TestRun_UpdateRebuilds -v`
Expected: FAIL — `update` hits the `default` branch, `os.Stat("update")` fails, prints "unknown command", exit 2.

- [ ] **Step 3: Add the `update` case**

In `cmd/graffiti/main.go`, add a case to the `switch cmd` block, immediately after the `build` case:

```go
	case "update":
		// update is currently a full rebuild; the incremental AST-only rebuild
		// (spec §11) is a later optimization. Behaves exactly like `build`.
		root := "."
		if len(args) >= 3 {
			root = args[2]
		}
		return runBuild(root, stdout, stderr)
```

- [ ] **Step 4: Update usage text**

In `func usage`, add a line after the `build` line:

```go
	fmt.Fprintln(w, "  update [path]     rebuild the map for <path> (default .)")
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./cmd/graffiti/ -run TestRun_UpdateRebuilds -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add cmd/graffiti/main.go cmd/graffiti/main_test.go
git commit -m "feat(cli): add \`graffiti update\` as a full rebuild alias"
```

---

## Task 2: `integrate.Target` + path constructors

**Files:**
- Create: `internal/integrate/target.go`
- Test: `internal/integrate/target_test.go`

- [ ] **Step 1: Write the failing test**

`internal/integrate/target_test.go`:

```go
package integrate

import (
	"path/filepath"
	"testing"
)

func TestProjectTarget_Paths(t *testing.T) {
	tg := ProjectTarget("/repo")
	if tg.Scope != "project" {
		t.Fatalf("scope = %q", tg.Scope)
	}
	if tg.SkillPath != filepath.FromSlash("/repo/.claude/skills/graffiti/SKILL.md") {
		t.Fatalf("skill path = %q", tg.SkillPath)
	}
	if tg.ClaudeMDPath != filepath.FromSlash("/repo/CLAUDE.md") {
		t.Fatalf("claude md path = %q", tg.ClaudeMDPath)
	}
	if tg.SettingsPath != filepath.FromSlash("/repo/.claude/settings.json") {
		t.Fatalf("settings path = %q", tg.SettingsPath)
	}
}

func TestUserTarget_Paths(t *testing.T) {
	tg := UserTarget("/home/u")
	if tg.Scope != "user" {
		t.Fatalf("scope = %q", tg.Scope)
	}
	if tg.SkillPath != filepath.FromSlash("/home/u/.claude/skills/graffiti/SKILL.md") {
		t.Fatalf("skill path = %q", tg.SkillPath)
	}
	// User-scoped CLAUDE.md lives under ~/.claude, not the home root.
	if tg.ClaudeMDPath != filepath.FromSlash("/home/u/.claude/CLAUDE.md") {
		t.Fatalf("claude md path = %q", tg.ClaudeMDPath)
	}
	if tg.SettingsPath != filepath.FromSlash("/home/u/.claude/settings.json") {
		t.Fatalf("settings path = %q", tg.SettingsPath)
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./internal/integrate/ -run Target -v`
Expected: FAIL — package/identifiers do not exist (build error).

- [ ] **Step 3: Write `target.go`**

```go
// Package integrate generates and installs graffiti's Claude Code integration:
// a Skill, an always-on CLAUDE.md block, and an optional PreToolUse hook (spec §9).
// All generated content is constant (byte-deterministic); merges are idempotent.
package integrate

import "path/filepath"

// Target is the set of absolute destination paths for one install scope.
// The CLI computes these from --user/--hook flags and hands them to Install,
// which keeps Install free of any environment/home-dir coupling (fully testable).
type Target struct {
	Scope        string // "project" or "user" — for the success message only
	SkillPath    string // .../.claude/skills/graffiti/SKILL.md
	ClaudeMDPath string // CLAUDE.md (repo root for project, ~/.claude for user)
	SettingsPath string // .../.claude/settings.json
}

// ProjectTarget installs into a repository rooted at root. CLAUDE.md sits at the
// repo root (the conventional project memory file).
func ProjectTarget(root string) Target {
	claudeDir := filepath.Join(root, ".claude")
	return Target{
		Scope:        "project",
		SkillPath:    filepath.Join(claudeDir, "skills", "graffiti", "SKILL.md"),
		ClaudeMDPath: filepath.Join(root, "CLAUDE.md"),
		SettingsPath: filepath.Join(claudeDir, "settings.json"),
	}
}

// UserTarget installs into the user's home. CLAUDE.md sits under ~/.claude.
func UserTarget(home string) Target {
	claudeDir := filepath.Join(home, ".claude")
	return Target{
		Scope:        "user",
		SkillPath:    filepath.Join(claudeDir, "skills", "graffiti", "SKILL.md"),
		ClaudeMDPath: filepath.Join(claudeDir, "CLAUDE.md"),
		SettingsPath: filepath.Join(claudeDir, "settings.json"),
	}
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/integrate/ -run Target -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/integrate/target.go internal/integrate/target_test.go
git commit -m "feat(integrate): install Target with project/user path constructors"
```

---

## Task 3: Skill content

**Files:**
- Create: `internal/integrate/skill.go`
- Test: `internal/integrate/skill_test.go`

- [ ] **Step 1: Write the failing test**

`internal/integrate/skill_test.go`:

```go
package integrate

import (
	"strings"
	"testing"
)

func TestSkillContent_Shape(t *testing.T) {
	s := SkillContent()
	if !strings.HasPrefix(s, "---\nname: graffiti\n") {
		head := s
		if len(head) > 40 {
			head = head[:40]
		}
		t.Fatalf("skill must start with YAML frontmatter incl. name: graffiti\n%q", head)
	}
	if !strings.Contains(s, "description:") {
		t.Fatal("skill frontmatter missing description")
	}
	// Frontmatter closes before the body heading.
	if strings.Count(s, "\n---\n") < 1 {
		t.Fatal("skill frontmatter not closed with ---")
	}
	for _, must := range []string{
		"graffiti build .",
		"graffiti query",
		"graffiti update",
		".graffiti/MAP.md",
	} {
		if !strings.Contains(s, must) {
			t.Fatalf("skill missing %q", must)
		}
	}
	if !strings.HasSuffix(s, "\n") {
		t.Fatal("skill must end with a trailing newline")
	}
	// description stays within Claude Code's 1536-char cap.
	desc := s[strings.Index(s, "description:"):]
	desc = desc[:strings.Index(desc, "\n")]
	if len(desc) > 1536 {
		t.Fatalf("description too long: %d chars", len(desc))
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./internal/integrate/ -run Skill -v`
Expected: FAIL — `SkillContent` undefined.

- [ ] **Step 3: Write `skill.go`**

Use a raw string literal. The content must match the "Constant content" block above byte-for-byte.

```go
package integrate

// SkillContent returns the body of .claude/skills/graffiti/SKILL.md. It is a
// short, declarative Skill (spec §9): the deterministic binary does the heavy
// lifting, so this just tells the assistant to build, read MAP.md, and query.
func SkillContent() string {
	return `---
name: graffiti
description: Use when exploring or answering questions about THIS codebase's structure — where something is defined, how components connect, what the architecture looks like. graffiti turns the repo into a queryable code map so you query the graph instead of grepping blind.
---

# graffiti — read the map, don't grep blind

` + "`graffiti`" + ` is a CLI that turns this repository into a queryable code map (no API key, $0, offline).

## First time in a repo
1. Run ` + "`graffiti build .`" + ` — it writes ` + "`.graffiti/map.json`" + `, ` + "`.graffiti/MAP.md`" + `, and ` + "`.graffiti/map.html`" + `.
2. Read ` + "`.graffiti/MAP.md`" + `. Tell the user the **god nodes** (most-connected code) and the **surprising connections**, then offer to trace the single most interesting question the map suggests.

## Answering questions about the code
- Prefer ` + "`graffiti query \"<question>\"`" + ` over grep/read — it returns a scoped, ranked subgraph (relevant nodes + edges), not raw text matches.
- To locate a symbol, run ` + "`graffiti query \"<symbol name>\"`" + `.

## After editing code
- Run ` + "`graffiti update`" + ` so later queries reflect the current code.

Keep this lightweight: the deterministic binary does the heavy lifting — build, read MAP.md, query.
`
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/integrate/ -run Skill -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/integrate/skill.go internal/integrate/skill_test.go
git commit -m "feat(integrate): deterministic SKILL.md content"
```

---

## Task 4: `MergeClaudeMD` — idempotent marker merge

(Reference logic validated in a standalone prototype before this plan was written.)

**Files:**
- Create: `internal/integrate/claudemd.go`
- Test: `internal/integrate/claudemd_test.go`

- [ ] **Step 1: Write the failing test**

`internal/integrate/claudemd_test.go`:

```go
package integrate

import (
	"bytes"
	"testing"
)

func TestMergeClaudeMD_CreateFromEmpty(t *testing.T) {
	out := MergeClaudeMD(nil)
	if !bytes.Equal(out, []byte(ClaudeBlock())) {
		t.Fatalf("from empty should equal the block exactly, got:\n%s", out)
	}
}

func TestMergeClaudeMD_AppendsAndPreserves(t *testing.T) {
	user := []byte("# My project\n\nSome rules here.\n")
	out := MergeClaudeMD(user)
	if !bytes.HasPrefix(out, user) {
		t.Fatal("user content must be preserved as a prefix")
	}
	if !bytes.Contains(out, []byte(claudeStart)) || !bytes.Contains(out, []byte("Some rules here.")) {
		t.Fatal("expected both the block and the original text")
	}
	// blank-line separator between user content and the block
	if !bytes.Contains(out, []byte("\n\n"+claudeStart)) {
		t.Fatal("expected a blank line before the inserted block")
	}
}

func TestMergeClaudeMD_Idempotent(t *testing.T) {
	user := []byte("# My project\n\nSome rules here.\n")
	once := MergeClaudeMD(user)
	twice := MergeClaudeMD(once)
	if !bytes.Equal(once, twice) {
		t.Fatalf("merge must be idempotent\nonce:\n%s\ntwice:\n%s", once, twice)
	}
	if n := bytes.Count(twice, []byte(claudeStart)); n != 1 {
		t.Fatalf("expected exactly one start marker, got %d", n)
	}
}

func TestMergeClaudeMD_RefreshesBetweenMarkers(t *testing.T) {
	stale := []byte("# Top\n\n" + claudeStart + "\nOLD STALE TEXT\n" + claudeEnd + "\n\n# Bottom\n")
	out := MergeClaudeMD(stale)
	if bytes.Contains(out, []byte("OLD STALE TEXT")) {
		t.Fatal("stale block content must be replaced")
	}
	if !bytes.Contains(out, []byte("# Top")) || !bytes.Contains(out, []byte("# Bottom")) {
		t.Fatal("content surrounding the markers must be preserved")
	}
	if n := bytes.Count(out, []byte(claudeStart)); n != 1 {
		t.Fatalf("expected one start marker, got %d", n)
	}
	if !bytes.Equal(out, MergeClaudeMD(out)) {
		t.Fatal("refreshed file must be idempotent")
	}
}

func TestMergeClaudeMD_NoTrailingNewline(t *testing.T) {
	in := []byte("line without newline")
	out := MergeClaudeMD(in)
	if !bytes.HasPrefix(out, in) {
		t.Fatal("must preserve content lacking a trailing newline")
	}
	if !bytes.Contains(out, []byte("\n\n"+claudeStart)) {
		t.Fatal("must insert a blank-line separator even without a trailing newline")
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./internal/integrate/ -run ClaudeMD -v`
Expected: FAIL — `MergeClaudeMD`/`ClaudeBlock`/`claudeStart` undefined.

- [ ] **Step 3: Write `claudemd.go`**

```go
package integrate

import "bytes"

const (
	claudeStart = "<!-- graffiti:start -->"
	claudeEnd   = "<!-- graffiti:end -->"
)

// ClaudeBlock is the always-on CLAUDE.md block (spec §9), wrapped in HTML-comment
// markers so it can be refreshed in place on re-run.
func ClaudeBlock() string {
	return claudeStart + `
## graffiti code map

If ` + "`.graffiti/map.json`" + ` exists, this repo has a graffiti code map. For questions about the
codebase's structure (where something lives, how parts connect, the architecture), run
` + "`graffiti query \"<question>\"`" + ` instead of grep/read — it returns a scoped subgraph. After
editing code, run ` + "`graffiti update`" + ` to refresh the map. If no map exists yet, run
` + "`graffiti build .`" + ` first.
` + claudeEnd + "\n"
}

// MergeClaudeMD returns the new CLAUDE.md content after inserting or refreshing
// the graffiti block. If both markers are present, the content between them is
// replaced (allowing content upgrades); otherwise the block is appended after a
// blank-line separator. Everything outside the markers is preserved byte-for-byte.
// The result is idempotent: MergeClaudeMD(MergeClaudeMD(x)) == MergeClaudeMD(x).
func MergeClaudeMD(existing []byte) []byte {
	block := ClaudeBlock()

	if i := bytes.Index(existing, []byte(claudeStart)); i >= 0 {
		if j := bytes.Index(existing[i:], []byte(claudeEnd)); j >= 0 {
			end := i + j + len(claudeEnd)
			// Absorb a single newline right after the end marker so the block's own
			// trailing newline doesn't accumulate blank lines across re-runs.
			if end < len(existing) && existing[end] == '\n' {
				end++
			}
			var out bytes.Buffer
			out.Write(existing[:i])
			out.WriteString(block)
			out.Write(existing[end:])
			return out.Bytes()
		}
		// start marker without a matching end marker: malformed; fall through to append.
	}

	if len(existing) == 0 {
		return []byte(block)
	}
	var out bytes.Buffer
	out.Write(existing)
	if !bytes.HasSuffix(existing, []byte("\n")) {
		out.WriteByte('\n')
	}
	out.WriteByte('\n')
	out.WriteString(block)
	return out.Bytes()
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/integrate/ -run ClaudeMD -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/integrate/claudemd.go internal/integrate/claudemd_test.go
git commit -m "feat(integrate): idempotent CLAUDE.md marker merge"
```

---

## Task 5: `MergeHookSettings` — idempotent JSON merge

(Reference logic validated in the same prototype.)

**Files:**
- Create: `internal/integrate/settings.go`
- Test: `internal/integrate/settings_test.go`

- [ ] **Step 1: Write the failing test**

`internal/integrate/settings_test.go`:

```go
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
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./internal/integrate/ -run HookSettings -v`
Expected: FAIL — `MergeHookSettings`/`HookCommand` undefined.

- [ ] **Step 3: Write `settings.go`**

```go
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
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/integrate/ -run HookSettings -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/integrate/settings.go internal/integrate/settings_test.go
git commit -m "feat(integrate): idempotent settings.json PreToolUse hook merge"
```

---

## Task 6: `RunHook` — non-blocking PreToolUse handler

The hook handler must be impossible to make break a Claude Code session: on any error it emits nothing and the tool proceeds. It nudges only when a map exists.

**Files:**
- Create: `internal/integrate/hook.go`
- Test: `internal/integrate/hook_test.go`

- [ ] **Step 1: Write the failing test**

`internal/integrate/hook_test.go`:

```go
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
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./internal/integrate/ -run RunHook -v`
Expected: FAIL — `RunHook` undefined.

- [ ] **Step 3: Write `hook.go`**

```go
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
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/integrate/ -run RunHook -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/integrate/hook.go internal/integrate/hook_test.go
git commit -m "feat(integrate): non-blocking PreToolUse hook handler"
```

---

## Task 7: `integrate.Install` — orchestration + disk writes

**Files:**
- Create: `internal/integrate/install.go`
- Test: `internal/integrate/install_test.go`

- [ ] **Step 1: Write the failing test**

`internal/integrate/install_test.go`:

```go
package integrate

import (
	"os"
	"testing"
)

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	return string(b)
}

func TestInstall_ProjectNoHook(t *testing.T) {
	root := t.TempDir()
	tg := ProjectTarget(root)
	res, err := Install(tg, Options{InstallHook: false})
	if err != nil {
		t.Fatal(err)
	}
	if res.Skill != ActionCreated {
		t.Fatalf("skill action = %v, want Created", res.Skill)
	}
	if res.ClaudeMD != ActionCreated {
		t.Fatalf("claudemd action = %v, want Created", res.ClaudeMD)
	}
	if res.HookInstalled {
		t.Fatal("hook must not be installed without InstallHook")
	}
	// Files exist with expected content.
	if got := readFile(t, tg.SkillPath); got != SkillContent() {
		t.Fatal("SKILL.md content mismatch")
	}
	if got := readFile(t, tg.ClaudeMDPath); got != ClaudeBlock() {
		t.Fatal("CLAUDE.md content mismatch")
	}
	if _, err := os.Stat(tg.SettingsPath); !os.IsNotExist(err) {
		t.Fatal("settings.json must not be written without --hook")
	}
}

func TestInstall_ProjectWithHook(t *testing.T) {
	root := t.TempDir()
	tg := ProjectTarget(root)
	res, err := Install(tg, Options{InstallHook: true})
	if err != nil {
		t.Fatal(err)
	}
	if !res.HookInstalled || res.Hook != ActionCreated {
		t.Fatalf("hook action = %v installed=%v", res.Hook, res.HookInstalled)
	}
	got := readFile(t, tg.SettingsPath)
	if !containsAll(got, "PreToolUse", HookCommand) {
		t.Fatalf("settings.json missing hook:\n%s", got)
	}
}

func TestInstall_IdempotentSecondRunUnchanged(t *testing.T) {
	root := t.TempDir()
	tg := ProjectTarget(root)
	if _, err := Install(tg, Options{InstallHook: true}); err != nil {
		t.Fatal(err)
	}
	res, err := Install(tg, Options{InstallHook: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Skill != ActionUnchanged || res.ClaudeMD != ActionUnchanged || res.Hook != ActionUnchanged {
		t.Fatalf("second run should be all-Unchanged, got skill=%v claude=%v hook=%v",
			res.Skill, res.ClaudeMD, res.Hook)
	}
}

func TestInstall_PreservesExistingClaudeMD(t *testing.T) {
	root := t.TempDir()
	tg := ProjectTarget(root)
	original := "# House rules\n\nAlways write tests.\n"
	if err := os.WriteFile(tg.ClaudeMDPath, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Install(tg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if res.ClaudeMD != ActionUpdated {
		t.Fatalf("claudemd action = %v, want Updated", res.ClaudeMD)
	}
	got := readFile(t, tg.ClaudeMDPath)
	if !containsAll(got, "House rules", "Always write tests.", claudeStart) {
		t.Fatalf("existing CLAUDE.md content not preserved:\n%s", got)
	}
}

func TestInstall_UserScopeIntoTempHome(t *testing.T) {
	home := t.TempDir()
	tg := UserTarget(home)
	if _, err := Install(tg, Options{InstallHook: true}); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{tg.SkillPath, tg.ClaudeMDPath, tg.SettingsPath} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected %s to exist: %v", p, err)
		}
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		found := false
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./internal/integrate/ -run Install -v`
Expected: FAIL — `Install`/`Options`/`Result`/`Action*` undefined.

- [ ] **Step 3: Write `install.go`**

```go
package integrate

import (
	"os"
	"path/filepath"
)

// Options controls what Install writes.
type Options struct {
	InstallHook bool // also merge the optional PreToolUse hook into settings.json
}

// Action records what happened to one artifact.
type Action int

const (
	ActionUnchanged Action = iota // already up to date
	ActionCreated                 // file did not exist; created
	ActionUpdated                 // file existed; content changed
)

func (a Action) String() string {
	switch a {
	case ActionCreated:
		return "created"
	case ActionUpdated:
		return "updated"
	default:
		return "unchanged"
	}
}

// Result summarizes one Install run for the CLI success message.
type Result struct {
	Target        Target
	Skill         Action
	ClaudeMD      Action
	Hook          Action
	HookInstalled bool // whether the hook was part of this run (Options.InstallHook)
}

// Install writes the three integration artifacts to the target paths. It creates
// parent directories as needed and is idempotent — a second run reports Unchanged.
func Install(t Target, opts Options) (Result, error) {
	res := Result{Target: t, HookInstalled: opts.InstallHook}

	// 1. Skill — we own this path; desired content is constant.
	skillAct, err := writeIfChanged(t.SkillPath, []byte(SkillContent()))
	if err != nil {
		return res, err
	}
	res.Skill = skillAct

	// 2. CLAUDE.md — surgical merge preserving user content.
	existingMD, err := readMaybe(t.ClaudeMDPath)
	if err != nil {
		return res, err
	}
	mergedMD := MergeClaudeMD(existingMD)
	mdAct, err := writeIfChanged(t.ClaudeMDPath, mergedMD)
	if err != nil {
		return res, err
	}
	res.ClaudeMD = mdAct

	// 3. settings.json hook — only with --hook.
	if opts.InstallHook {
		existingSettings, err := readMaybe(t.SettingsPath)
		if err != nil {
			return res, err
		}
		mergedSettings, err := MergeHookSettings(existingSettings)
		if err != nil {
			return res, err
		}
		hookAct, err := writeIfChanged(t.SettingsPath, mergedSettings)
		if err != nil {
			return res, err
		}
		res.Hook = hookAct
	}

	return res, nil
}

// readMaybe reads a file, treating "not found" as empty (no error).
func readMaybe(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return b, nil
}

// writeIfChanged writes content to path only if it differs from what's there,
// creating parent dirs. It returns Created (new file), Updated (changed), or
// Unchanged (byte-identical).
func writeIfChanged(path string, content []byte) (Action, error) {
	existing, err := os.ReadFile(path)
	switch {
	case err == nil:
		if string(existing) == string(content) {
			return ActionUnchanged, nil
		}
	case !os.IsNotExist(err):
		return ActionUnchanged, err
	}
	created := os.IsNotExist(err)
	if mkErr := os.MkdirAll(filepath.Dir(path), 0o755); mkErr != nil {
		return ActionUnchanged, mkErr
	}
	if wErr := os.WriteFile(path, content, 0o644); wErr != nil {
		return ActionUnchanged, wErr
	}
	if created {
		return ActionCreated, nil
	}
	return ActionUpdated, nil
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/integrate/ -run Install -v`
Expected: PASS

- [ ] **Step 5: Run the whole package**

Run: `go test ./internal/integrate/ -v`
Expected: PASS (all tasks 2–7)

- [ ] **Step 6: Commit**

```bash
git add internal/integrate/install.go internal/integrate/install_test.go
git commit -m "feat(integrate): Install orchestration with per-artifact actions"
```

---

## Task 8: CLI wiring — `graffiti init` and `graffiti hook`

**Files:**
- Modify: `cmd/graffiti/main.go`
- Test: `cmd/graffiti/main_test.go`

- [ ] **Step 1: Write the failing tests**

Add to `cmd/graffiti/main_test.go`:

```go
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
```

Note: the `--root` flag is a test seam so the project-scope path is injectable without `os.Chdir`. It defaults to `.` in real use.

- [ ] **Step 2: Run to verify they fail**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./cmd/graffiti/ -run 'Init|Hook' -v`
Expected: FAIL — unknown command `init`/`hook`.

- [ ] **Step 3: Add the import**

In `cmd/graffiti/main.go` imports, add:

```go
	"github.com/amazopic/graffiti/internal/integrate"
```

- [ ] **Step 4: Add `init` and `hook` cases**

In the `switch cmd` block, add before `default:`:

```go
	case "init":
		return runInit(args[2:], stdout, stderr)
	case "hook":
		// Internal: PreToolUse handler. Always exits 0; never blocks a tool.
		cwd, _ := os.Getwd()
		integrate.RunHook(stdin, stdout, cwd)
		return 0
```

- [ ] **Step 5: Implement `runInit`**

Add to `cmd/graffiti/main.go`:

```go
// runInit installs the Claude Code integration. Flags: --user (install into the
// home dir instead of the project), --hook (also install the PreToolUse hook),
// --root <dir> (project root; defaults to "." — primarily a test seam).
func runInit(args []string, stdout, stderr io.Writer) int {
	var user, hook bool
	root := "."
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--user":
			user = true
		case "--hook":
			hook = true
		case "--root":
			if i+1 >= len(args) {
				fmt.Fprintln(stderr, "graffiti: --root requires a directory")
				return 2
			}
			i++
			root = args[i]
		default:
			fmt.Fprintf(stderr, "graffiti: unknown init flag %q\n", args[i])
			return 2
		}
	}

	var target integrate.Target
	if user {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(stderr, "graffiti: cannot resolve home dir: %v\n", err)
			return 1
		}
		target = integrate.UserTarget(home)
	} else {
		target = integrate.ProjectTarget(root)
	}

	res, err := integrate.Install(target, integrate.Options{InstallHook: hook})
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: init failed: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "✓ graffiti wired into Claude Code (%s).\n", res.Target.Scope)
	fmt.Fprintf(stdout, "  • skill:     %s (%s)\n", res.Target.SkillPath, res.Skill)
	fmt.Fprintf(stdout, "  • CLAUDE.md: %s (%s)\n", res.Target.ClaudeMDPath, res.ClaudeMD)
	if res.HookInstalled {
		fmt.Fprintf(stdout, "  • hook:      %s (%s) — PreToolUse nudge grep → graffiti query\n", res.Target.SettingsPath, res.Hook)
	} else {
		fmt.Fprintln(stdout, "  • hook:      skipped (pass --hook to nudge grep → graffiti query)")
	}
	fmt.Fprintln(stdout, "  Re-run `graffiti init` any time; it is idempotent.")
	return 0
}
```

- [ ] **Step 6: Update usage text**

In `func usage`, add lines after the `serve` line:

```go
	fmt.Fprintln(w, "  init [--user] [--hook]  install Claude Code integration (skill + CLAUDE.md [+ hook])")
```

(Do not advertise the internal `hook` subcommand in usage.)

- [ ] **Step 7: Run the tests**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./cmd/graffiti/ -run 'Init|Hook|Update' -v`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add cmd/graffiti/main.go cmd/graffiti/main_test.go
git commit -m "feat(cli): graffiti init (--user/--hook) and internal graffiti hook"
```

---

## Task 9: Golden test, docs, and full verification

**Files:**
- Create: `testdata/golden/init/SKILL.md`, `testdata/golden/init/CLAUDE.md`, `testdata/golden/init/settings.json`
- Create: `internal/integrate/golden_test.go`
- Modify: `README.md`, `docs/superpowers/specs/2026-06-14-graffiti-design.md`

- [ ] **Step 1: Write the golden test**

`internal/integrate/golden_test.go`:

```go
package integrate

import (
	"os"
	"path/filepath"
	"testing"
)

// goldenDir resolves testdata/golden/init relative to the repo root (two levels
// up from internal/integrate).
func goldenDir(t *testing.T) string {
	t.Helper()
	return filepath.FromSlash("../../testdata/golden/init")
}

func TestInstall_MatchesGolden(t *testing.T) {
	root := t.TempDir()
	tg := ProjectTarget(root)
	if _, err := Install(tg, Options{InstallHook: true}); err != nil {
		t.Fatal(err)
	}
	cases := []struct{ got, golden string }{
		{tg.SkillPath, "SKILL.md"},
		{tg.ClaudeMDPath, "CLAUDE.md"},
		{tg.SettingsPath, "settings.json"},
	}
	for _, c := range cases {
		gotB, err := os.ReadFile(c.got)
		if err != nil {
			t.Fatalf("read produced %s: %v", c.golden, err)
		}
		wantB, err := os.ReadFile(filepath.Join(goldenDir(t), c.golden))
		if err != nil {
			t.Fatalf("read golden %s: %v", c.golden, err)
		}
		if string(gotB) != string(wantB) {
			t.Fatalf("%s differs from golden.\n--- got ---\n%s\n--- want ---\n%s", c.golden, gotB, wantB)
		}
	}
}
```

- [ ] **Step 2: Generate the golden files from the tool**

Build, run a throwaway install into a temp dir, and copy the three artifacts into the golden dir (the honest golden workflow — content is deterministic):

```bash
make build
TMP=$(mktemp -d)
./graffiti init --hook --root "$TMP" >/dev/null
mkdir -p testdata/golden/init
cp "$TMP/.claude/skills/graffiti/SKILL.md" testdata/golden/init/SKILL.md
cp "$TMP/CLAUDE.md" testdata/golden/init/CLAUDE.md
cp "$TMP/.claude/settings.json" testdata/golden/init/settings.json
rm -rf "$TMP"
```

- [ ] **Step 3: Run the golden test**

Run: `go test ./internal/integrate/ -run Golden -v`
Expected: PASS

- [ ] **Step 4: Confirm the golden settings.json is exactly what we expect**

Run: `cat testdata/golden/init/settings.json`
Expected (sorted keys, 2-space indent):

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "hooks": [
          {
            "command": "graffiti hook",
            "type": "command"
          }
        ],
        "matcher": "Grep|Glob"
      }
    ]
  }
}
```

- [ ] **Step 5: Update README**

Add a section to `README.md` documenting the integration:

```markdown
## Claude Code integration

```
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` writes:
- `.claude/skills/graffiti/SKILL.md` — a short skill so Claude Code knows to build/read/query the map.
- a `CLAUDE.md` block (between `<!-- graffiti:start -->` / `<!-- graffiti:end -->`) telling the assistant to prefer `graffiti query` over grep when a map exists.
- with `--hook`, a `.claude/settings.json` PreToolUse entry running `graffiti hook`, which adds a one-line nudge before Grep/Glob when `.graffiti/map.json` is present. The hook never blocks a tool.

It is idempotent — re-run any time; existing `CLAUDE.md`/`settings.json` content is preserved.
```

- [ ] **Step 6: Note the build-out in the spec**

Append this exact note to `docs/superpowers/specs/2026-06-14-graffiti-design.md`, immediately after the §9 bullet list:

```markdown
**Implemented (Plan 5, 2026-06-16):** `graffiti init [--user] [--hook]` installs three artifacts via `internal/integrate`: (1) `.claude/skills/graffiti/SKILL.md` (overwritten — namespaced path we own); (2) a `CLAUDE.md` block delimited by `<!-- graffiti:start -->`/`<!-- graffiti:end -->`, merged surgically (replace-between-markers or append) so all other content is preserved; (3) with `--hook`, a `.claude/settings.json` PreToolUse entry (`matcher: "Grep|Glob"`, command `graffiti hook`) merged idempotently into any existing JSON. The hook is the binary itself — a hidden `graffiti hook` subcommand reads the PreToolUse event on stdin and, when `.graffiti/map.json` exists, emits the verified non-blocking `{"hookSpecificOutput":{"hookEventName":"PreToolUse","additionalContext":"…"}}` (exit 0, no `permissionDecision`); it never blocks a tool and degrades to harmless if a future build ignores `additionalContext`. All content is byte-deterministic and golden-locked; every merge is idempotent. `graffiti update` (CLI surface §11) currently maps to a full rebuild; the incremental AST-only rebuild remains a later optimization.
```

- [ ] **Step 7: Full verification**

Run each and confirm:

```bash
go vet -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./...
go test ./...                                                            # no-tags config
go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./...   # subset config
go mod tidy && git diff --exit-code go.mod go.sum                       # zero new deps
make build && make xcompile                                             # size guard
```

Expected: vet clean; both test configs green (all packages incl. `internal/integrate`); `go mod tidy` a no-op (no new deps); cross-compile under the 16MB guard.

- [ ] **Step 8: End-to-end smoke**

```bash
SMOKE=$(mktemp -d)
cp -r testdata/fixtures/gorepo/* "$SMOKE"/
./graffiti init --hook --root "$SMOKE"
./graffiti build "$SMOKE"
# map now exists; the hook should nudge:
echo "{\"hook_event_name\":\"PreToolUse\",\"tool_name\":\"Grep\",\"cwd\":\"$SMOKE\"}" | ./graffiti hook
rm -rf "$SMOKE"
```

Expected: init prints the success block; build writes the map; the final `graffiti hook` prints a JSON object containing `additionalContext` + `graffiti query`.

- [ ] **Step 9: Commit**

```bash
git add testdata/golden/init internal/integrate/golden_test.go README.md docs/superpowers/specs/2026-06-14-graffiti-design.md
git commit -m "test(integrate): golden from-scratch install; docs for graffiti init"
```

---

## Self-review checklist (run before merge)

1. **Spec §9 coverage:** short skill ✓ (Task 3), always-on CLAUDE.md block ✓ (Task 4), optional PreToolUse hook ✓ (Tasks 5–6, `--hook`), "skill is short and declarative" ✓. CLI surface §11: `graffiti init` ✓ (Task 8), `graffiti update` ✓ (Task 1).
2. **Determinism (§14):** SKILL.md/CLAUDE.md constant; settings.json via `json.MarshalIndent` (sorted keys) → golden-locked (Task 9). No `generated_at` in any init artifact.
3. **Idempotency:** every artifact merge re-runs to Unchanged (Tasks 4, 5, 7, 8).
4. **Safety:** `MergeHookSettings` errors (never clobbers) on malformed JSON; `RunHook` never blocks/errs; CLAUDE.md merge preserves all out-of-marker content; SKILL.md is the only file we overwrite wholesale (namespaced path we own).
5. **No new deps:** stdlib only; `go mod tidy` no-op (Task 9 step 7).
6. **Type consistency:** `Action`/`Result`/`Options`/`Target` names match across `install.go` and tests; `HookCommand`/`hookNudge`/`claudeStart`/`claudeEnd` constants single-sourced.

## Deferred follow-ups (record in memory, non-blocking)

- Hook matcher is `Grep|Glob` only; Bash-grep/`rg`/`find` detection (parse `tool_input.command`) is deferred.
- Hook nudges on every matching tool call; a once-per-session dedupe (keyed by `session_id`, best-effort marker under `.graffiti/`) would cut noise.
- `graffiti update` is a full rebuild; the incremental AST-only rebuild (spec §11) is still open.
- `additionalContext` on PreToolUse is the verified mechanism; if a future Claude Code build changes the contract, the hook still degrades to harmless (never blocks).
- User-scope install resolves `~` via `os.UserHomeDir()`; a `--root`-style seam exists for project scope but user scope is only smoke-tested against a temp home in unit tests via `UserTarget`.
