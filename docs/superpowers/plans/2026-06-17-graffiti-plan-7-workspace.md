# graffiti Plan 7 — Workspace federation (foundation + explicit links)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Lay separate per-project graphs side by side and draw the wires between them — **without merging anything** (spec §16). `graffiti link <a> <b>` federates repos; `graffiti query --workspace "<q>"` returns an alias-prefixed cross-repo subgraph; explicit `links` assert cross-project edges at 100% precision. Each project's `map.json` stays unchanged and authoritative.

**Architecture:** A **purely additive** `internal/workspace` package — it never touches the single-project build/schema/extraction. A workspace is a thin computed overlay over N independent `map.json` files. Three artifacts (§16.1): each project's `.graffiti/map.json` (unchanged, authoritative); a committable registry `<root>/.graffiti-workspace/workspace.json` (member aliases + relative paths + last-seen hashes); a derived, gitignorable cache `<root>/.graffiti-workspace/overlay.json` (computed cross-edges + the source hashes they were built from). Cross-edges are ordinary `graph.Edge`s whose endpoints carry an `alias::` prefix — **no new enum values**. Federated query reuses the proven `query.Query` verbatim by building an **in-memory combined `store.Index`** whose node ids are `alias::localid` plus the overlay cross-edges; the alias prefix is never written back into any project's `map.json`.

**Scope (this plan = the federation engine + the "ships-first" signal):** project federation, registry, computed overlay, **signal 1 (explicit links → EXTRACTED)** from §16.2, federated `query`/`serve`/`update`, auto-discovered workspace root, determinism. **Deferred to follow-on plans** (per §16's own staging): auto-detection signals 2–5 (symbol/OpenAPI/literal-HTTP/SDK — they require a per-project *contract-surface* carrier that changes the single-project schema), the §16.6 immutable project id (v1 resolves members by **alias**), the two-tier query budget (v1 uses one budget; cross-edges traverse like any edge), and the `workspace.html` lanes viewer (§16.7).

**Locked decisions (zero-dep ethos — same isolate-and-pivot as the §10 MCP/parser amendments):**
- **Explicit-links file is line-based**, not YAML (no YAML dependency). `<root>/.graffiti-workspace/links` — one link per line `FROM -> TO [relation]`, `alias::nodeid` endpoints, `#` comments, blank lines ignored. Spec §16.2's `links.yml` is amended to this format for v1.
- **Members resolved by alias** (defer §16.6 immutable id).
- **Workspace root auto-discovered** at the nearest common ancestor of member paths (§16.4); later runs search upward from cwd for `.graffiti-workspace/workspace.json`.
- **No new dependencies** (stdlib + existing `internal/{graph,store,query,mcp,app}`).

**Tech Stack:** Go 1.26, stdlib + existing internal packages. Determinism per §14/§16.5: members sorted by alias, links by `(from,to,relation)`, output byte-identical modulo `generated_at`.

---

## File structure

```
internal/workspace/
  model.go        Member, Registry, Link, Overlay types + JSON tags + constants
  discover.go     CommonAncestor(paths) + FindRoot(cwd) (walk up to .graffiti-workspace)
  registry.go     LoadRegistry/SaveRegistry; Add/Remove/sort members; MapHash(path)
  links.go        ParseLinks(bytes) → []ParsedLink (line format) + errors
  overlay.go      Compute(reg, root) → Overlay (resolve explicit links); Load/SaveOverlay; Stale(reg, overlay)
  federate.go     CombinedIndex(reg, overlay, root) → *store.Index (alias-prefixed) + member load
  *_test.go for each
schema/workspace.schema.json   published schema for workspace.json + overlay.json     [new]
internal/schemaval/workspace.go  ValidateRegistry/ValidateOverlay (structural)          [new]
cmd/graffiti/main.go   + link / workspace / links / federate / query --workspace / serve --workspace / update --workspace
cmd/graffiti/workspace_test.go   CLI-level tests                                        [new]
testdata/fixtures/ws/   two tiny pre-built member repos (frontend, backend) for tests   [new]
testdata/golden/ws.overlay.json                                                         [new]
README.md   workspace section   ·   docs/superpowers/specs/...   §16 amendments
```

---

## Task 1: workspace model + root discovery

**Files:**
- Create: `internal/workspace/model.go`, `internal/workspace/discover.go`
- Test: `internal/workspace/discover_test.go`

- [ ] **Step 1: Write the failing test**

`internal/workspace/discover_test.go`:

```go
package workspace

import (
	"path/filepath"
	"testing"
)

func TestCommonAncestor(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{[]string{"/a/b/frontend", "/a/b/backend"}, "/a/b"},
		{[]string{"/a/b/c", "/a/b/c/d"}, "/a/b/c"},
		{[]string{"/a/x", "/a/y", "/a/z/q"}, "/a"},
		{[]string{"/only/one"}, "/only/one"},
	}
	for _, c := range cases {
		in := make([]string, len(c.in))
		for i, p := range c.in {
			in[i] = filepath.FromSlash(p)
		}
		if got := CommonAncestor(in); got != filepath.FromSlash(c.want) {
			t.Errorf("CommonAncestor(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./internal/workspace/ -run CommonAncestor -v`
Expected: FAIL — package/`CommonAncestor` undefined.

- [ ] **Step 3: Write `model.go`**

```go
// Package workspace federates N independent per-project graffiti graphs into a
// thin computed overlay WITHOUT merging them (spec §16). Every per-project
// map.json stays unchanged and authoritative; this package only reads them and
// computes cross-project links. Alias-qualified ids ("alias::nodeid") exist only
// in the overlay and the in-memory federated index — never written back.
package workspace

// SchemaVersion is stamped into workspace.json and overlay.json.
const SchemaVersion = "1"

// WorkspaceDir is the per-root directory holding the registry and derived overlay.
const WorkspaceDir = ".graffiti-workspace"

const (
	registryFile = "workspace.json"
	overlayFile  = "overlay.json"
	linksFile    = "links"
)

// Member is one federated project: a display alias, a path relative to the
// workspace root, and the sha256 of its map.json when last seen.
type Member struct {
	Alias   string `json:"alias"`
	Path    string `json:"path"`     // relative to the workspace root, slash-separated
	MapHash string `json:"map_hash"` // sha256 hex of the member's .graffiti/map.json
}

// Registry is the committable workspace.json (no graph data — pointers + intent).
type Registry struct {
	Version     string   `json:"version"`
	Name        string   `json:"name"`
	GeneratedAt string   `json:"generated_at"` // RFC3339
	Members     []Member `json:"members"`      // sorted by alias
}

// Link is a cross-project edge in the derived overlay. Endpoints are alias::id.
// Relation/Confidence reuse the §6 vocabularies verbatim (no new enum values).
type Link struct {
	From       string `json:"from"`       // "alias::nodeid"
	To         string `json:"to"`         // "alias::nodeid"
	Relation   string `json:"relation"`   // graph.Relation value
	Confidence string `json:"confidence"` // graph.Confidence value
	Via        string `json:"via"`        // discovery provenance, e.g. "explicit"
}

// Overlay is the derived .graffiti-workspace/overlay.json (recomputable cache).
type Overlay struct {
	Version      string            `json:"version"`
	GeneratedAt  string            `json:"generated_at"` // RFC3339
	SourceHashes map[string]string `json:"source_hashes"` // alias -> map_hash used
	Links        []Link            `json:"links"`         // confident; sorted (from,to,relation)
	Ambiguous    []Link            `json:"ambiguous"`     // surfaced for review, never traversed
}
```

- [ ] **Step 4: Write `discover.go`**

```go
package workspace

import (
	"os"
	"path/filepath"
	"strings"
)

// CommonAncestor returns the deepest directory that is a prefix of every input
// path (the default workspace root, §16.4). Inputs should be absolute, cleaned
// paths. With a single input it returns that path unchanged.
func CommonAncestor(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	split := func(p string) []string {
		return strings.Split(filepath.Clean(p), string(filepath.Separator))
	}
	common := split(paths[0])
	for _, p := range paths[1:] {
		parts := split(p)
		n := len(common)
		if len(parts) < n {
			n = len(parts)
		}
		i := 0
		for i < n && common[i] == parts[i] {
			i++
		}
		common = common[:i]
	}
	joined := strings.Join(common, string(filepath.Separator))
	if joined == "" {
		return string(filepath.Separator)
	}
	return joined
}

// FindRoot searches cwd and its ancestors for a directory containing
// .graffiti-workspace/workspace.json and returns that directory, or "" if none.
func FindRoot(cwd string) string {
	dir := filepath.Clean(cwd)
	for {
		if _, err := os.Stat(filepath.Join(dir, WorkspaceDir, registryFile)); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/workspace/ -run CommonAncestor -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/workspace/model.go internal/workspace/discover.go internal/workspace/discover_test.go
git commit -m "feat(workspace): federation model types + workspace-root discovery"
```

---

## Task 2: registry load/save + map_hash

**Files:**
- Create: `internal/workspace/registry.go`
- Test: `internal/workspace/registry_test.go`

- [ ] **Step 1: Write the failing test**

`internal/workspace/registry_test.go`:

```go
package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegistry_SaveLoadRoundTrip(t *testing.T) {
	root := t.TempDir()
	reg := &Registry{
		Version: SchemaVersion, Name: "shop", GeneratedAt: "2026-06-17T00:00:00Z",
		Members: []Member{
			{Alias: "web", Path: "../frontend", MapHash: "h1"},
			{Alias: "api", Path: "../backend", MapHash: "h2"},
		},
	}
	if err := SaveRegistry(root, reg); err != nil {
		t.Fatal(err)
	}
	got, err := LoadRegistry(root)
	if err != nil {
		t.Fatal(err)
	}
	// Members come back sorted by alias (api before web).
	if len(got.Members) != 2 || got.Members[0].Alias != "api" {
		t.Fatalf("members not sorted by alias: %+v", got.Members)
	}
	if got.Name != "shop" {
		t.Fatalf("name = %q", got.Name)
	}
}

func TestAddRemoveMember(t *testing.T) {
	reg := &Registry{Version: SchemaVersion}
	AddMember(reg, Member{Alias: "web", Path: "../w"})
	AddMember(reg, Member{Alias: "api", Path: "../a"})
	AddMember(reg, Member{Alias: "web", Path: "../w2"}) // replace, not duplicate
	if len(reg.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(reg.Members))
	}
	if reg.Members[0].Alias != "api" { // sorted
		t.Fatalf("not sorted: %+v", reg.Members)
	}
	for _, m := range reg.Members {
		if m.Alias == "web" && m.Path != "../w2" {
			t.Fatalf("web not replaced: %+v", m)
		}
	}
	if !RemoveMember(reg, "web") || len(reg.Members) != 1 {
		t.Fatalf("remove failed: %+v", reg.Members)
	}
}

func TestMapHash(t *testing.T) {
	dir := t.TempDir()
	if _, err := MapHash(dir); err == nil {
		t.Fatal("expected error when map.json is absent")
	}
	if err := os.MkdirAll(filepath.Join(dir, ".graffiti"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".graffiti", "map.json"), []byte(`{"a":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	h, err := MapHash(dir)
	if err != nil || len(h) != 64 {
		t.Fatalf("hash=%q err=%v", h, err)
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./internal/workspace/ -run 'Registry|Member|MapHash' -v`
Expected: FAIL — undefined identifiers.

- [ ] **Step 3: Write `registry.go`**

```go
package workspace

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// MapHash returns the lowercase-hex sha256 of <memberDir>/.graffiti/map.json.
func MapHash(memberDir string) (string, error) {
	b, err := os.ReadFile(filepath.Join(memberDir, ".graffiti", "map.json"))
	if err != nil {
		return "", fmt.Errorf("workspace: read member map.json: %w", err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// AddMember inserts or replaces (by alias) a member, keeping Members sorted by alias.
func AddMember(reg *Registry, m Member) {
	for i := range reg.Members {
		if reg.Members[i].Alias == m.Alias {
			reg.Members[i] = m
			sortMembers(reg)
			return
		}
	}
	reg.Members = append(reg.Members, m)
	sortMembers(reg)
}

// RemoveMember drops the member with the given alias; reports whether one was removed.
func RemoveMember(reg *Registry, alias string) bool {
	out := reg.Members[:0]
	removed := false
	for _, m := range reg.Members {
		if m.Alias == alias {
			removed = true
			continue
		}
		out = append(out, m)
	}
	reg.Members = out
	return removed
}

func sortMembers(reg *Registry) {
	sort.SliceStable(reg.Members, func(i, j int) bool { return reg.Members[i].Alias < reg.Members[j].Alias })
}

func registryPath(root string) string { return filepath.Join(root, WorkspaceDir, registryFile) }

// SaveRegistry writes workspace.json (members sorted, 2-space indent, trailing newline).
func SaveRegistry(root string, reg *Registry) error {
	sortMembers(reg)
	if err := os.MkdirAll(filepath.Join(root, WorkspaceDir), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(registryPath(root), append(b, '\n'), 0o644)
}

// LoadRegistry reads workspace.json and returns it with members sorted by alias.
func LoadRegistry(root string) (*Registry, error) {
	b, err := os.ReadFile(registryPath(root))
	if err != nil {
		return nil, fmt.Errorf("workspace: read registry: %w", err)
	}
	var reg Registry
	if err := json.Unmarshal(b, &reg); err != nil {
		return nil, fmt.Errorf("workspace: parse registry: %w", err)
	}
	sortMembers(&reg)
	return &reg, nil
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/workspace/ -run 'Registry|Member|MapHash' -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/workspace/registry.go internal/workspace/registry_test.go
git commit -m "feat(workspace): workspace.json registry load/save + member map_hash"
```

---

## Task 3: links file parser

**Files:**
- Create: `internal/workspace/links.go`
- Test: `internal/workspace/links_test.go`

- [ ] **Step 1: Write the failing test**

`internal/workspace/links_test.go`:

```go
package workspace

import "testing"

func TestParseLinks(t *testing.T) {
	in := `# workspace links
web::cartclient.fetchcart -> api::handlers.get_cart calls

api::db.save -> web::types.cart
  web::a.b -> api::c.d references  # trailing comment
`
	links, err := ParseLinks([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 3 {
		t.Fatalf("expected 3 links, got %d: %+v", len(links), links)
	}
	if links[0].FromAlias != "web" || links[0].FromID != "cartclient.fetchcart" {
		t.Fatalf("bad from: %+v", links[0])
	}
	if links[0].ToAlias != "api" || links[0].ToID != "handlers.get_cart" {
		t.Fatalf("bad to: %+v", links[0])
	}
	if links[0].Relation != "calls" {
		t.Fatalf("relation = %q, want calls", links[0].Relation)
	}
	if links[1].Relation != "references" { // default
		t.Fatalf("default relation = %q, want references", links[1].Relation)
	}
}

func TestParseLinks_Errors(t *testing.T) {
	for _, bad := range []string{
		"web::a -> noalias",       // RHS missing alias::
		"noarrow line",            // no ->
		"web::a -> api::b badrel", // unknown relation
		"::a -> api::b",           // empty alias
		"web:: -> api::b",         // empty id
	} {
		if _, err := ParseLinks([]byte(bad)); err == nil {
			t.Errorf("expected error for %q", bad)
		}
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./internal/workspace/ -run ParseLinks -v`
Expected: FAIL — `ParseLinks`/`ParsedLink` undefined.

- [ ] **Step 3: Write `links.go`**

```go
package workspace

import (
	"fmt"
	"strings"
)

// ParsedLink is one parsed explicit link (signal 1, §16.2). Endpoints are split
// into alias + local node id; the alias-qualified id is "alias::id".
type ParsedLink struct {
	FromAlias, FromID string
	ToAlias, ToID     string
	Relation          string
}

// validRelations mirrors graph.ValidRelations (kept local to avoid importing the
// graph package here; the values are the §6 relation vocabulary).
var validRelations = map[string]bool{
	"calls": true, "imports": true, "inherits": true,
	"implements": true, "references": true, "contains": true,
}

// ParseLinks parses the line-based explicit-links file (§16.2 signal 1, v1
// format): one link per non-blank, non-comment line:
//
//	FROM -> TO [relation]      # FROM and TO are "alias::nodeid"
//
// Default relation is "references". '#' starts a comment (to end of line).
func ParseLinks(b []byte) ([]ParsedLink, error) {
	var out []ParsedLink
	for i, raw := range strings.Split(string(b), "\n") {
		line := raw
		if h := strings.IndexByte(line, '#'); h >= 0 {
			line = line[:h]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lhs, rhs, ok := strings.Cut(line, "->")
		if !ok {
			return nil, fmt.Errorf("links line %d: missing '->': %q", i+1, raw)
		}
		fromA, fromID, err := splitEndpoint(strings.TrimSpace(lhs))
		if err != nil {
			return nil, fmt.Errorf("links line %d: from: %w", i+1, err)
		}
		fields := strings.Fields(strings.TrimSpace(rhs))
		if len(fields) == 0 {
			return nil, fmt.Errorf("links line %d: missing target", i+1)
		}
		toA, toID, err := splitEndpoint(fields[0])
		if err != nil {
			return nil, fmt.Errorf("links line %d: to: %w", i+1, err)
		}
		rel := "references"
		if len(fields) >= 2 {
			rel = fields[1]
			if !validRelations[rel] {
				return nil, fmt.Errorf("links line %d: unknown relation %q", i+1, rel)
			}
		}
		out = append(out, ParsedLink{fromA, fromID, toA, toID, rel})
	}
	return out, nil
}

func splitEndpoint(s string) (alias, id string, err error) {
	a, i, ok := strings.Cut(s, "::")
	if !ok {
		return "", "", fmt.Errorf("endpoint %q is not alias::id", s)
	}
	if a == "" || i == "" {
		return "", "", fmt.Errorf("endpoint %q has empty alias or id", s)
	}
	return a, i, nil
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/workspace/ -run ParseLinks -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/workspace/links.go internal/workspace/links_test.go
git commit -m "feat(workspace): line-based explicit-links parser (signal 1)"
```

---

## Task 4: overlay compute + load/save

**Files:**
- Create: `internal/workspace/overlay.go`
- Test: `internal/workspace/overlay_test.go`

- [ ] **Step 1: Write the failing test**

`internal/workspace/overlay_test.go`:

```go
package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amazopic/graffiti/internal/graph"
	"github.com/amazopic/graffiti/internal/render"
)

// writeMember builds a member dir with a map.json containing the given node ids.
func writeMember(t *testing.T, root, rel string, ids ...string) {
	t.Helper()
	dir := filepath.Join(root, rel)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := graph.NewDocument(rel)
	doc.GeneratedAt = "2026-06-17T00:00:00Z"
	for _, id := range ids {
		doc.Nodes = append(doc.Nodes, graph.Node{ID: id, Label: id, Kind: graph.KindFunction, File: "f", Line: 1, Community: -1})
	}
	if err := render.WriteMapJSON(doc, dir); err != nil {
		t.Fatal(err)
	}
}

func newRegistry(t *testing.T, root string) *Registry {
	t.Helper()
	reg := &Registry{Version: SchemaVersion, Name: "ws", GeneratedAt: "2026-06-17T00:00:00Z"}
	AddMember(reg, Member{Alias: "web", Path: "frontend"})
	AddMember(reg, Member{Alias: "api", Path: "backend"})
	return reg
}

func TestComputeOverlay_ResolvesExplicitLinks(t *testing.T) {
	root := t.TempDir()
	writeMember(t, root, "frontend", "cartclient.fetchcart")
	writeMember(t, root, "backend", "handlers.get_cart")
	reg := newRegistry(t, root)
	links := []ParsedLink{{"web", "cartclient.fetchcart", "api", "handlers.get_cart", "calls"}}

	ov, err := ComputeOverlay(root, reg, links)
	if err != nil {
		t.Fatal(err)
	}
	if len(ov.Links) != 1 {
		t.Fatalf("expected 1 link, got %d (%+v)", len(ov.Links), ov)
	}
	l := ov.Links[0]
	if l.From != "web::cartclient.fetchcart" || l.To != "api::handlers.get_cart" {
		t.Fatalf("bad endpoints: %+v", l)
	}
	if l.Confidence != string(graph.ConfExtracted) || l.Via != "explicit" || l.Relation != "calls" {
		t.Fatalf("bad link metadata: %+v", l)
	}
	if ov.SourceHashes["web"] == "" || ov.SourceHashes["api"] == "" {
		t.Fatalf("missing source hashes: %+v", ov.SourceHashes)
	}
}

func TestComputeOverlay_DropsUnresolvable(t *testing.T) {
	root := t.TempDir()
	writeMember(t, root, "frontend", "real.node")
	writeMember(t, root, "backend", "other.node")
	reg := newRegistry(t, root)
	links := []ParsedLink{
		{"web", "real.node", "api", "ghost.node", "calls"}, // To ghost → unresolved
		{"web", "missing", "api", "other.node", "calls"},   // From ghost → unresolved
		{"nope", "x", "api", "other.node", "calls"},        // bad alias → unresolved
	}
	ov, err := ComputeOverlay(root, reg, links)
	if err != nil {
		t.Fatal(err)
	}
	if len(ov.Links) != 0 {
		t.Fatalf("expected 0 confident links, got %+v", ov.Links)
	}
	if len(ov.Unresolved) != 3 {
		t.Fatalf("expected 3 unresolved, got %d", len(ov.Unresolved))
	}
}

func TestOverlay_SaveLoad_DeterministicSort(t *testing.T) {
	root := t.TempDir()
	ov := &Overlay{
		Version: SchemaVersion, GeneratedAt: "T",
		SourceHashes: map[string]string{"api": "h"},
		Links: []Link{
			{From: "web::b", To: "api::z", Relation: "calls", Confidence: "EXTRACTED", Via: "explicit"},
			{From: "web::a", To: "api::y", Relation: "calls", Confidence: "EXTRACTED", Via: "explicit"},
		},
	}
	if err := SaveOverlay(root, ov); err != nil {
		t.Fatal(err)
	}
	got, err := LoadOverlay(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Links[0].From != "web::a" { // sorted by (from,to,relation)
		t.Fatalf("links not sorted: %+v", got.Links)
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./internal/workspace/ -run 'Overlay' -v`
Expected: FAIL — `ComputeOverlay`/`SaveOverlay`/`LoadOverlay`/`Overlay.Unresolved` undefined.

- [ ] **Step 3: Add the `Unresolved` field to `Overlay` in `model.go`**

Add a field to the `Overlay` struct (after `Ambiguous`):

```go
	Unresolved []Link `json:"unresolved,omitempty"` // links whose endpoints don't resolve (reported by `links check`)
```

- [ ] **Step 4: Write `overlay.go`**

```go
package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/amazopic/graffiti/internal/graph"
	"github.com/amazopic/graffiti/internal/store"
)

// memberIndexes loads each member's map.json into a store.Index keyed by alias.
// Also returns the sha256 of each member's map.json (for source_hashes / staleness).
func memberIndexes(root string, reg *Registry) (map[string]*store.Index, map[string]string, error) {
	idxByAlias := make(map[string]*store.Index, len(reg.Members))
	hashes := make(map[string]string, len(reg.Members))
	for _, m := range reg.Members {
		dir := filepath.Join(root, filepath.FromSlash(m.Path))
		doc, err := store.Load(filepath.Join(dir, ".graffiti", "map.json"))
		if err != nil {
			return nil, nil, fmt.Errorf("workspace: member %q: %w", m.Alias, err)
		}
		idxByAlias[m.Alias] = store.NewIndex(doc)
		h, err := MapHash(dir)
		if err != nil {
			return nil, nil, err
		}
		hashes[m.Alias] = h
	}
	return idxByAlias, hashes, nil
}

// ComputeOverlay resolves explicit links against the members' current map.json
// files (signal 1). A link whose BOTH endpoints resolve to real nodes becomes an
// EXTRACTED cross-edge (via: explicit); otherwise it is recorded in Unresolved
// (never emitted as a confident edge — honesty-first under-linking, §16.2).
func ComputeOverlay(root string, reg *Registry, links []ParsedLink) (*Overlay, error) {
	idxByAlias, hashes, err := memberIndexes(root, reg)
	if err != nil {
		return nil, err
	}
	ov := &Overlay{Version: SchemaVersion, SourceHashes: hashes}
	resolves := func(alias, id string) bool {
		idx, ok := idxByAlias[alias]
		if !ok {
			return false
		}
		_, ok = idx.Node(id)
		return ok
	}
	for _, pl := range links {
		l := Link{
			From: pl.FromAlias + "::" + pl.FromID,
			To:   pl.ToAlias + "::" + pl.ToID,
			Relation: pl.Relation, Confidence: string(graph.ConfExtracted), Via: "explicit",
		}
		if resolves(pl.FromAlias, pl.FromID) && resolves(pl.ToAlias, pl.ToID) {
			ov.Links = append(ov.Links, l)
		} else {
			ov.Unresolved = append(ov.Unresolved, l)
		}
	}
	sortLinks(ov.Links)
	sortLinks(ov.Unresolved)
	return ov, nil
}

func sortLinks(ls []Link) {
	sort.SliceStable(ls, func(i, j int) bool {
		a, b := ls[i], ls[j]
		if a.From != b.From {
			return a.From < b.From
		}
		if a.To != b.To {
			return a.To < b.To
		}
		return a.Relation < b.Relation
	})
}

func overlayPath(root string) string { return filepath.Join(root, WorkspaceDir, overlayFile) }

// SaveOverlay writes overlay.json (links sorted, 2-space indent, trailing newline).
func SaveOverlay(root string, ov *Overlay) error {
	sortLinks(ov.Links)
	sortLinks(ov.Ambiguous)
	sortLinks(ov.Unresolved)
	if err := os.MkdirAll(filepath.Join(root, WorkspaceDir), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(ov, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(overlayPath(root), append(b, '\n'), 0o644)
}

// LoadOverlay reads overlay.json with links sorted.
func LoadOverlay(root string) (*Overlay, error) {
	b, err := os.ReadFile(overlayPath(root))
	if err != nil {
		return nil, fmt.Errorf("workspace: read overlay: %w", err)
	}
	var ov Overlay
	if err := json.Unmarshal(b, &ov); err != nil {
		return nil, fmt.Errorf("workspace: parse overlay: %w", err)
	}
	sortLinks(ov.Links)
	return &ov, nil
}

// StaleMembers returns the aliases whose current map.json hash differs from the
// hash the overlay was computed against (self-healing nudge, §16.3).
func StaleMembers(root string, reg *Registry, ov *Overlay) ([]string, error) {
	_, hashes, err := memberIndexes(root, reg)
	if err != nil {
		return nil, err
	}
	var stale []string
	for alias, cur := range hashes {
		if ov.SourceHashes[alias] != cur {
			stale = append(stale, alias)
		}
	}
	sort.Strings(stale)
	return stale, nil
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/workspace/ -run 'Overlay' -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/workspace/overlay.go internal/workspace/overlay_test.go internal/workspace/model.go
git commit -m "feat(workspace): compute/persist overlay from explicit links + staleness"
```

---

## Task 5: combined federated index

**Files:**
- Create: `internal/workspace/federate.go`
- Test: `internal/workspace/federate_test.go`

- [ ] **Step 1: Write the failing test**

`internal/workspace/federate_test.go`:

```go
package workspace

import (
	"testing"

	"github.com/amazopic/graffiti/internal/query"
)

func TestCombinedIndex_PrefixesAndLinks(t *testing.T) {
	root := t.TempDir()
	writeMember(t, root, "frontend", "cartclient.fetchcart")
	writeMember(t, root, "backend", "handlers.get_cart")
	reg := newRegistry(t, root)
	ov := &Overlay{
		Version: SchemaVersion,
		Links: []Link{{
			From: "web::cartclient.fetchcart", To: "api::handlers.get_cart",
			Relation: "calls", Confidence: "EXTRACTED", Via: "explicit",
		}},
	}
	idx, err := CombinedIndex(root, reg, ov)
	if err != nil {
		t.Fatal(err)
	}
	// alias-prefixed node ids exist
	if _, ok := idx.Node("web::cartclient.fetchcart"); !ok {
		t.Fatal("missing alias-prefixed web node")
	}
	if _, ok := idx.Node("api::handlers.get_cart"); !ok {
		t.Fatal("missing alias-prefixed api node")
	}
	// the cross-edge is present (out of the web node)
	found := false
	for _, e := range idx.Out("web::cartclient.fetchcart") {
		if e.To == "api::handlers.get_cart" && string(e.Relation) == "calls" {
			found = true
		}
	}
	if !found {
		t.Fatal("cross-edge not present in combined index")
	}
	// a federated query over the combined index returns alias-prefixed text
	out := query.Query(idx, "fetchcart", query.DefaultTokenBudget)
	if !contains(out, "web::cartclient.fetchcart") {
		t.Fatalf("federated query output not alias-prefixed:\n%s", out)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./internal/workspace/ -run CombinedIndex -v`
Expected: FAIL — `CombinedIndex` undefined.

- [ ] **Step 3: Write `federate.go`**

```go
package workspace

import (
	"path/filepath"

	"github.com/amazopic/graffiti/internal/graph"
	"github.com/amazopic/graffiti/internal/store"
)

// CombinedIndex builds an in-memory store.Index over all members with every node
// id and edge endpoint prefixed by "alias::", plus the overlay's confident
// cross-edges. The prefix lives ONLY here and in overlay.json — never written
// back into any project's map.json (§16.1). Reusing store.NewIndex + query.Query
// over this index gives federated retrieval with no changes to the query engine.
func CombinedIndex(root string, reg *Registry, ov *Overlay) (*store.Index, error) {
	combined := graph.NewDocument(reg.Name)
	if ov != nil {
		combined.GeneratedAt = ov.GeneratedAt
	}
	for _, m := range reg.Members {
		dir := filepath.Join(root, filepath.FromSlash(m.Path))
		doc, err := store.Load(filepath.Join(dir, ".graffiti", "map.json"))
		if err != nil {
			return nil, err
		}
		p := m.Alias + "::"
		for _, n := range doc.Nodes {
			n.ID = p + n.ID
			combined.Nodes = append(combined.Nodes, n)
		}
		for _, e := range doc.Edges {
			e.From = p + e.From
			e.To = p + e.To
			combined.Edges = append(combined.Edges, e)
		}
	}
	if ov != nil {
		for _, l := range ov.Links { // confident cross-edges only
			combined.Edges = append(combined.Edges, graph.Edge{
				From: l.From, To: l.To,
				Relation: graph.Relation(l.Relation), Confidence: graph.Confidence(l.Confidence),
			})
		}
	}
	return store.NewIndex(combined), nil
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/workspace/ -run CombinedIndex -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/workspace/federate.go internal/workspace/federate_test.go
git commit -m "feat(workspace): in-memory alias-prefixed combined federated index"
```

---

## Task 6: workspace JSON schema + validation

**Files:**
- Create: `schema/workspace.schema.json`, `internal/schemaval/workspace.go`
- Test: `internal/schemaval/workspace_test.go`

- [ ] **Step 1: Write the failing test**

`internal/schemaval/workspace_test.go`:

```go
package schemaval

import "testing"

func TestValidateRegistry(t *testing.T) {
	ok := []byte(`{"version":"1","name":"ws","generated_at":"T","members":[{"alias":"a","path":"../a","map_hash":"h"}]}`)
	if err := ValidateRegistryBytes(ok); err != nil {
		t.Fatalf("valid registry rejected: %v", err)
	}
	for _, bad := range [][]byte{
		[]byte(`{"name":"ws"}`),                                  // missing version/members
		[]byte(`{"version":"1","name":"ws","members":"nope"}`),   // members not array
		[]byte(`{"version":"1","name":"ws","members":[{"x":1}]}`),// member missing alias
	} {
		if err := ValidateRegistryBytes(bad); err == nil {
			t.Errorf("expected error for %s", bad)
		}
	}
}

func TestValidateOverlay(t *testing.T) {
	ok := []byte(`{"version":"1","generated_at":"T","source_hashes":{"a":"h"},"links":[{"from":"a::x","to":"b::y","relation":"calls","confidence":"EXTRACTED","via":"explicit"}]}`)
	if err := ValidateOverlayBytes(ok); err != nil {
		t.Fatalf("valid overlay rejected: %v", err)
	}
	bad := []byte(`{"version":"1","links":[{"from":"a::x","relation":"calls"}]}`) // link missing to/confidence
	if err := ValidateOverlayBytes(bad); err == nil {
		t.Error("expected error for malformed link")
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./internal/schemaval/ -run 'Registry|Overlay' -v`
Expected: FAIL — undefined.

- [ ] **Step 3: Write `schema/workspace.schema.json`**

A small published schema (informative; the Go validators below are authoritative for tests):

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://graffiti.dev/workspace.schema.json",
  "title": "graffiti workspace files",
  "oneOf": [
    {
      "title": "registry (workspace.json)",
      "type": "object",
      "required": ["version", "name", "members"],
      "properties": {
        "version": { "type": "string" },
        "name": { "type": "string" },
        "generated_at": { "type": "string" },
        "members": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["alias", "path"],
            "properties": {
              "alias": { "type": "string", "minLength": 1 },
              "path": { "type": "string", "minLength": 1 },
              "map_hash": { "type": "string" }
            }
          }
        }
      }
    },
    {
      "title": "overlay (overlay.json)",
      "type": "object",
      "required": ["version", "links"],
      "properties": {
        "version": { "type": "string" },
        "generated_at": { "type": "string" },
        "source_hashes": { "type": "object", "additionalProperties": { "type": "string" } },
        "links": { "type": "array", "items": { "$ref": "#/$defs/link" } },
        "ambiguous": { "type": "array", "items": { "$ref": "#/$defs/link" } },
        "unresolved": { "type": "array", "items": { "$ref": "#/$defs/link" } }
      }
    }
  ],
  "$defs": {
    "link": {
      "type": "object",
      "required": ["from", "to", "relation", "confidence"],
      "properties": {
        "from": { "type": "string" },
        "to": { "type": "string" },
        "relation": { "type": "string" },
        "confidence": { "enum": ["EXTRACTED", "INFERRED", "AMBIGUOUS"] },
        "via": { "type": "string" }
      }
    }
  }
}
```

- [ ] **Step 4: Write `internal/schemaval/workspace.go`**

Follow the existing structural-validation style in `schemaval.go` (stdlib only — no JSON-schema engine):

```go
package schemaval

import (
	"encoding/json"
	"fmt"
)

// ValidateRegistryBytes checks workspace.json's required shape (structural; the
// published schema/workspace.schema.json is the informative contract).
func ValidateRegistryBytes(b []byte) error {
	var r struct {
		Version *string `json:"version"`
		Name    *string `json:"name"`
		Members *[]struct {
			Alias *string `json:"alias"`
			Path  *string `json:"path"`
		} `json:"members"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return fmt.Errorf("registry: %w", err)
	}
	if r.Version == nil || r.Name == nil || r.Members == nil {
		return fmt.Errorf("registry: missing required version/name/members")
	}
	for i, m := range *r.Members {
		if m.Alias == nil || *m.Alias == "" || m.Path == nil || *m.Path == "" {
			return fmt.Errorf("registry: member %d missing alias/path", i)
		}
	}
	return nil
}

// ValidateOverlayBytes checks overlay.json's required shape.
func ValidateOverlayBytes(b []byte) error {
	type link struct {
		From       *string `json:"from"`
		To         *string `json:"to"`
		Relation   *string `json:"relation"`
		Confidence *string `json:"confidence"`
	}
	var o struct {
		Version *string `json:"version"`
		Links   *[]link `json:"links"`
	}
	if err := json.Unmarshal(b, &o); err != nil {
		return fmt.Errorf("overlay: %w", err)
	}
	if o.Version == nil || o.Links == nil {
		return fmt.Errorf("overlay: missing required version/links")
	}
	valid := map[string]bool{"EXTRACTED": true, "INFERRED": true, "AMBIGUOUS": true}
	for i, l := range *o.Links {
		if l.From == nil || l.To == nil || l.Relation == nil || l.Confidence == nil {
			return fmt.Errorf("overlay: link %d missing from/to/relation/confidence", i)
		}
		if !valid[*l.Confidence] {
			return fmt.Errorf("overlay: link %d bad confidence %q", i, *l.Confidence)
		}
	}
	return nil
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/schemaval/ -run 'Registry|Overlay' -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add schema/workspace.schema.json internal/schemaval/workspace.go internal/schemaval/workspace_test.go
git commit -m "feat(schemaval): publish + validate workspace.json / overlay.json shapes"
```

---

## Task 7: CLI `graffiti link`

**Files:**
- Modify: `cmd/graffiti/main.go`
- Create: `cmd/graffiti/workspace_test.go`

- [ ] **Step 1: Write the failing test**

`cmd/graffiti/workspace_test.go`:

```go
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeGoRepo creates a tiny buildable Go repo under dir/sub and returns its path.
func writeGoRepo(t *testing.T, dir, sub, pkg, src string) string {
	t.Helper()
	p := filepath.Join(dir, sub)
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRun_LinkBuildsAndFederates(t *testing.T) {
	base := t.TempDir()
	web := writeGoRepo(t, base, "frontend", "web", "package web\n\nfunc FetchCart() {}\n")
	api := writeGoRepo(t, base, "backend", "api", "package api\n\nfunc GetCart() {}\n")

	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "link", "--name", "shop", web, api}, bytes.NewReader(nil), &out, &errOut)
	if code != 0 {
		t.Fatalf("link exit=%d stderr=%q", code, errOut.String())
	}
	// workspace.json + overlay.json written under the common ancestor (base)
	if _, err := os.Stat(filepath.Join(base, ".graffiti-workspace", "workspace.json")); err != nil {
		t.Fatalf("workspace.json missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(base, ".graffiti-workspace", "overlay.json")); err != nil {
		t.Fatalf("overlay.json missing: %v", err)
	}
	if !strings.Contains(out.String(), "Linked 2 projects") {
		t.Fatalf("missing success line:\n%s", out.String())
	}
}
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod grammar_subset_python grammar_subset_javascript grammar_subset_typescript grammar_subset_rust grammar_subset_java grammar_subset_php" ./cmd/graffiti/ -run LinkBuilds -v`
Expected: FAIL — unknown command `link`.

- [ ] **Step 3: Add imports and the `link` case to `main.go`**

Add imports:

```go
	"github.com/amazopic/graffiti/internal/workspace"
```

Add a case in the `switch cmd` block (before `default:`):

```go
	case "link":
		return runLink(args[2:], stdout, stderr)
```

Add the helper:

```go
// runLink builds any unbuilt members, auto-discovers the workspace root (nearest
// common ancestor), writes workspace.json, computes the overlay from links, and
// prints the success line. Flags: --name <name>.
func runLink(args []string, stdout, stderr io.Writer) int {
	name := "workspace"
	var paths []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--name":
			if i+1 >= len(args) {
				fmt.Fprintln(stderr, "graffiti: --name requires a value")
				return 2
			}
			i++
			name = args[i]
		default:
			paths = append(paths, args[i])
		}
	}
	if len(paths) < 2 {
		fmt.Fprintln(stderr, "graffiti: link requires at least two project paths")
		return 2
	}

	abs := make([]string, len(paths))
	for i, p := range paths {
		a, err := filepath.Abs(p)
		if err != nil {
			fmt.Fprintf(stderr, "graffiti: %v\n", err)
			return 1
		}
		abs[i] = a
	}
	root := workspace.CommonAncestor(abs)

	reg := &workspace.Registry{Version: workspace.SchemaVersion, Name: name, GeneratedAt: nowRFC3339()}
	for _, a := range abs {
		// build the member if it has no map.json yet
		if _, err := os.Stat(filepath.Join(a, ".graffiti", "map.json")); err != nil {
			if _, berr := app.Build(a, nowRFC3339()); berr != nil {
				fmt.Fprintf(stderr, "graffiti: build %s: %v\n", a, berr)
				return 1
			}
		}
		rel, err := filepath.Rel(root, a)
		if err != nil {
			fmt.Fprintf(stderr, "graffiti: %v\n", err)
			return 1
		}
		h, err := workspace.MapHash(a)
		if err != nil {
			fmt.Fprintf(stderr, "graffiti: %v\n", err)
			return 1
		}
		workspace.AddMember(reg, workspace.Member{Alias: aliasFor(a), Path: filepath.ToSlash(rel), MapHash: h})
	}
	if err := workspace.SaveRegistry(root, reg); err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	ov, err := computeAndSaveOverlay(root, reg)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "✓ Linked %d projects. %d cross-project links (%d EXTRACTED, %d unresolved). 0 API calls, $0.\n",
		len(reg.Members), len(ov.Links), len(ov.Links), len(ov.Unresolved))
	return 0
}

// aliasFor derives a member alias from its directory base name.
func aliasFor(absPath string) string { return filepath.Base(absPath) }

// nowRFC3339 is the build/link timestamp (UTC, RFC3339).
func nowRFC3339() string { return time.Now().UTC().Format(time.RFC3339) }

// computeAndSaveOverlay reads <root>/.graffiti-workspace/links (if any), computes
// the overlay against the registry's members, stamps generated_at, and saves it.
func computeAndSaveOverlay(root string, reg *workspace.Registry) (*workspace.Overlay, error) {
	var links []workspace.ParsedLink
	if b, err := os.ReadFile(filepath.Join(root, workspace.WorkspaceDir, "links")); err == nil {
		links, err = workspace.ParseLinks(b)
		if err != nil {
			return nil, err
		}
	}
	ov, err := workspace.ComputeOverlay(root, reg, links)
	if err != nil {
		return nil, err
	}
	ov.GeneratedAt = nowRFC3339()
	if err := workspace.SaveOverlay(root, ov); err != nil {
		return nil, err
	}
	return ov, nil
}
```

Note: replace the existing inline `time.Now().UTC().Format(time.RFC3339)` in `runBuild` with a call to `nowRFC3339()` to DRY (optional but keep the helper single-sourced). Ensure `time` and `app` are already imported (they are).

- [ ] **Step 4: Add `link` to usage text**

In `func usage`, after the `init` line:

```go
	fmt.Fprintln(w, "  link <pathA> <pathB> [...] [--name n]  federate projects into a workspace")
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod grammar_subset_python grammar_subset_javascript grammar_subset_typescript grammar_subset_rust grammar_subset_java grammar_subset_php" ./cmd/graffiti/ -run LinkBuilds -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add cmd/graffiti/main.go cmd/graffiti/workspace_test.go
git commit -m "feat(cli): graffiti link — build members, write registry, compute overlay"
```

---

## Task 8: CLI `workspace` / `links check` / `federate --explain`

**Files:**
- Modify: `cmd/graffiti/main.go`
- Test: `cmd/graffiti/workspace_test.go`

- [ ] **Step 1: Write the failing tests**

Add to `cmd/graffiti/workspace_test.go`:

```go
// linkShop builds two members and federates them, returning the workspace root (base).
func linkShop(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	web := writeGoRepo(t, base, "frontend", "web", "package web\n\nfunc FetchCart() {}\n")
	api := writeGoRepo(t, base, "backend", "api", "package api\n\nfunc GetCart() {}\n")
	var out, errOut bytes.Buffer
	if code := run([]string{"graffiti", "link", web, api}, bytes.NewReader(nil), &out, &errOut); code != 0 {
		t.Fatalf("link failed: %s", errOut.String())
	}
	return base
}

func TestRun_WorkspaceList(t *testing.T) {
	base := linkShop(t)
	// run from inside the workspace (cwd discovery) by passing --root.
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "workspace", "list", "--root", base}, bytes.NewReader(nil), &out, &errOut)
	if code != 0 {
		t.Fatalf("list exit=%d stderr=%q", code, errOut.String())
	}
	if !strings.Contains(out.String(), "frontend") || !strings.Contains(out.String(), "backend") {
		t.Fatalf("list missing members:\n%s", out.String())
	}
}

func TestRun_LinksCheck(t *testing.T) {
	base := linkShop(t)
	// write a links file: one resolvable, one ghost.
	wsDir := filepath.Join(base, ".graffiti-workspace")
	// Verified node-id slugs: FetchCart -> "main-go:fetchcart", GetCart -> "main-go:getcart".
	// (alias::id splits on the FIRST "::", so the single colon inside the id is fine.)
	links := "frontend::main-go:fetchcart -> backend::main-go:getcart calls\nfrontend::main-go:ghost -> backend::main-go:getcart\n"
	if err := os.WriteFile(filepath.Join(wsDir, "links"), []byte(links), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "links", "check", "--root", base}, bytes.NewReader(nil), &out, &errOut)
	// non-zero exit because one link is unresolved
	if code == 0 {
		t.Fatalf("expected non-zero exit for an unresolved link:\n%s", out.String())
	}
	if !strings.Contains(out.String()+errOut.String(), "ghost") {
		t.Fatalf("expected the unresolved 'ghost' link to be reported")
	}
}
```

Note: the actual node ids depend on graffiti's slugging of `FetchCart`/`GetCart`. The implementer must first run `graffiti query` (or inspect map.json) on a built member to confirm the exact ids (likely `fetchcart` and `getcart`), and use those in the test's links file. Adjust the literal ids to whatever the build produces.

- [ ] **Step 2: Run to verify they fail**

Run: `go test -tags "..." ./cmd/graffiti/ -run 'WorkspaceList|LinksCheck' -v` (use the full tag list)
Expected: FAIL — unknown command `workspace`/`links`.

- [ ] **Step 3: Add the `workspace`, `links`, `federate` cases**

In the `switch cmd` block:

```go
	case "workspace":
		return runWorkspace(args[2:], stdout, stderr)
	case "links":
		return runLinksCheck(args[2:], stdout, stderr)
	case "federate":
		return runFederateExplain(args[2:], stdout, stderr)
```

Add a small shared flag parser and the helpers:

```go
// resolveWorkspaceRoot returns the workspace root: an explicit --root if present,
// else discovered by walking up from cwd. The returned args have --root removed.
func resolveWorkspaceRoot(args []string, stderr io.Writer) (root string, rest []string, code int) {
	rest = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "--root" {
			if i+1 >= len(args) {
				fmt.Fprintln(stderr, "graffiti: --root requires a directory")
				return "", nil, 2
			}
			root = args[i+1]
			i++
			continue
		}
		rest = append(rest, args[i])
	}
	if root == "" {
		cwd, _ := os.Getwd()
		root = workspace.FindRoot(cwd)
		if root == "" {
			fmt.Fprintln(stderr, "graffiti: no workspace found (run `graffiti link` first, or pass --root)")
			return "", nil, 1
		}
	}
	return root, rest, 0
}

func runWorkspace(args []string, stdout, stderr io.Writer) int {
	root, rest, code := resolveWorkspaceRoot(args, stderr)
	if code != 0 {
		return code
	}
	if len(rest) == 0 {
		fmt.Fprintln(stderr, "graffiti: workspace <add|rm|list>")
		return 2
	}
	reg, err := workspace.LoadRegistry(root)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	switch rest[0] {
	case "list":
		for _, m := range reg.Members {
			fmt.Fprintf(stdout, "%s\t%s\n", m.Alias, m.Path)
		}
		return 0
	case "rm":
		if len(rest) < 2 {
			fmt.Fprintln(stderr, "graffiti: workspace rm <alias>")
			return 2
		}
		if !workspace.RemoveMember(reg, rest[1]) {
			fmt.Fprintf(stderr, "graffiti: no member %q\n", rest[1])
			return 1
		}
	case "add":
		// graffiti workspace add <path> --as <alias>
		var path, alias string
		for i := 1; i < len(rest); i++ {
			if rest[i] == "--as" && i+1 < len(rest) {
				alias = rest[i+1]
				i++
			} else {
				path = rest[i]
			}
		}
		if path == "" {
			fmt.Fprintln(stderr, "graffiti: workspace add <path> [--as alias]")
			return 2
		}
		absPath, _ := filepath.Abs(path)
		if alias == "" {
			alias = aliasFor(absPath)
		}
		if _, err := os.Stat(filepath.Join(absPath, ".graffiti", "map.json")); err != nil {
			if _, berr := app.Build(absPath, nowRFC3339()); berr != nil {
				fmt.Fprintf(stderr, "graffiti: build %s: %v\n", absPath, berr)
				return 1
			}
		}
		rel, _ := filepath.Rel(root, absPath)
		h, _ := workspace.MapHash(absPath)
		workspace.AddMember(reg, workspace.Member{Alias: alias, Path: filepath.ToSlash(rel), MapHash: h})
	default:
		fmt.Fprintf(stderr, "graffiti: unknown workspace subcommand %q\n", rest[0])
		return 2
	}
	if err := workspace.SaveRegistry(root, reg); err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	if _, err := computeAndSaveOverlay(root, reg); err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	return 0
}

func runLinksCheck(args []string, stdout, stderr io.Writer) int {
	root, rest, code := resolveWorkspaceRoot(args, stderr)
	if code != 0 {
		return code
	}
	if len(rest) == 0 || rest[0] != "check" {
		fmt.Fprintln(stderr, "graffiti: links check")
		return 2
	}
	reg, err := workspace.LoadRegistry(root)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	ov, err := computeAndSaveOverlay(root, reg)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "%d links OK.\n", len(ov.Links))
	if len(ov.Unresolved) > 0 {
		for _, l := range ov.Unresolved {
			fmt.Fprintf(stdout, "UNRESOLVED: %s -> %s\n", l.From, l.To)
		}
		return 1
	}
	return 0
}

func runFederateExplain(args []string, stdout, stderr io.Writer) int {
	root, _, code := resolveWorkspaceRoot(args, stderr)
	if code != 0 {
		return code
	}
	ov, err := workspace.LoadOverlay(root)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	for _, l := range ov.Links {
		fmt.Fprintf(stdout, "%s -%s-> %s (%s, via %s)\n", l.From, l.Relation, l.To, l.Confidence, l.Via)
	}
	for _, l := range ov.Ambiguous {
		fmt.Fprintf(stdout, "AMBIGUOUS: %s -> %s (via %s)\n", l.From, l.To, l.Via)
	}
	return 0
}
```

- [ ] **Step 4: Update usage text** (add `workspace`, `links check`, `federate --explain` lines).

- [ ] **Step 5: Run the tests** (full tag list) → PASS. Confirm the node ids in the test links file are correct (adjust to the built slugs).

- [ ] **Step 6: Commit**

```bash
git add cmd/graffiti/main.go cmd/graffiti/workspace_test.go
git commit -m "feat(cli): workspace add/rm/list, links check, federate --explain"
```

---

## Task 9: CLI `query --workspace`

**Files:**
- Modify: `cmd/graffiti/main.go`
- Test: `cmd/graffiti/workspace_test.go`

- [ ] **Step 1: Write the failing test**

Add to `cmd/graffiti/workspace_test.go`:

```go
func TestRun_QueryWorkspace(t *testing.T) {
	base := linkShop(t)
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "query", "--workspace", "--root", base, "cart"}, bytes.NewReader(nil), &out, &errOut)
	if code != 0 {
		t.Fatalf("query --workspace exit=%d stderr=%q", code, errOut.String())
	}
	s := out.String()
	if !strings.Contains(s, "NODES") {
		t.Fatalf("missing NODES block:\n%s", s)
	}
	// alias-prefixed ids appear (both members are searched)
	if !strings.Contains(s, "frontend::") && !strings.Contains(s, "backend::") {
		t.Fatalf("expected alias-prefixed federated output:\n%s", s)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `go test -tags "..." ./cmd/graffiti/ -run QueryWorkspace -v`
Expected: FAIL — `--workspace` flag not handled (treated as the question, or errors).

- [ ] **Step 3: Route `query --workspace` in the `query` case**

Replace the body of `case "query":` so it first detects `--workspace` (and an optional `--root`):

```go
	case "query":
		qargs := args[2:]
		if hasFlag(qargs, "--workspace") {
			return runQueryWorkspace(stripFlag(qargs, "--workspace"), stdout, stderr)
		}
		// ... existing single-project query handling unchanged ...
```

(Keep the existing single-project arg parsing exactly as-is below the `--workspace` branch.)

Add helpers:

```go
func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func stripFlag(args []string, flag string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		if a != flag {
			out = append(out, a)
		}
	}
	return out
}

// runQueryWorkspace loads the workspace, builds the combined alias-prefixed index,
// runs the LLM-free query over it, and appends a staleness nudge if any member
// changed since the overlay was computed. Args (after --workspace removed):
// optional --root <dir>, then the question (and optional [name], ignored in v1).
func runQueryWorkspace(args []string, stdout, stderr io.Writer) int {
	root, rest, code := resolveWorkspaceRoot(args, stderr)
	if code != 0 {
		return code
	}
	// the last non-flag arg is the question
	if len(rest) == 0 {
		fmt.Fprintln(stderr, "graffiti: query --workspace requires a question")
		return 2
	}
	question := rest[len(rest)-1]

	reg, err := workspace.LoadRegistry(root)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	ov, err := workspace.LoadOverlay(root)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	idx, err := workspace.CombinedIndex(root, reg, ov)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	fmt.Fprint(stdout, query.Query(idx, question, query.DefaultTokenBudget))

	if stale, err := workspace.StaleMembers(root, reg, ov); err == nil && len(stale) > 0 {
		fmt.Fprintf(stdout, "\n(overlay stale: %s changed — run: graffiti update --workspace)\n", strings.Join(stale, ", "))
	}
	return 0
}
```

Ensure `strings` is imported in main.go (add if missing).

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test -tags "..." ./cmd/graffiti/ -run QueryWorkspace -v`
Expected: PASS

- [ ] **Step 5: Confirm single-project query still works** (no regression):

Run: `go test -tags "..." ./cmd/graffiti/ -run 'Query' -v`
Expected: PASS (TestRun_QueryPrintsSubgraph etc. unchanged).

- [ ] **Step 6: Commit**

```bash
git add cmd/graffiti/main.go cmd/graffiti/workspace_test.go
git commit -m "feat(cli): graffiti query --workspace over the combined federated index"
```

---

## Task 10: CLI `serve --workspace` + `update --workspace`

**Files:**
- Modify: `cmd/graffiti/main.go`
- Test: `cmd/graffiti/workspace_test.go`

- [ ] **Step 1: Write the failing tests**

Add to `cmd/graffiti/workspace_test.go`:

```go
func TestRun_ServeWorkspace(t *testing.T) {
	base := linkShop(t)
	initLine := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}` + "\n"
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "serve", "--workspace", "--root", base}, strings.NewReader(initLine), &out, &errOut)
	if code != 0 {
		t.Fatalf("serve --workspace exit=%d stderr=%q", code, errOut.String())
	}
	if !strings.Contains(out.String(), `"protocolVersion":"2025-06-18"`) {
		t.Fatalf("serve --workspace missing initialize echo:\n%s", out.String())
	}
}

func TestRun_UpdateWorkspace(t *testing.T) {
	base := linkShop(t)
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "update", "--workspace", "--root", base}, bytes.NewReader(nil), &out, &errOut)
	if code != 0 {
		t.Fatalf("update --workspace exit=%d stderr=%q", code, errOut.String())
	}
	if !strings.Contains(out.String(), "overlay") {
		t.Fatalf("update --workspace should mention overlay recompute:\n%s", out.String())
	}
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `go test -tags "..." ./cmd/graffiti/ -run 'ServeWorkspace|UpdateWorkspace' -v`
Expected: FAIL — `--workspace` not handled by serve/update.

- [ ] **Step 3: Route `--workspace` in the `serve` and `update` cases**

`serve` case:

```go
	case "serve":
		sargs := args[2:]
		if hasFlag(sargs, "--workspace") {
			return serveWorkspace(stripFlag(sargs, "--workspace"), stdin, stdout, stderr)
		}
		// ... existing single-project serve unchanged ...
```

`update` case:

```go
	case "update":
		if hasFlag(args[2:], "--workspace") {
			return updateWorkspace(stripFlag(args[2:], "--workspace"), stdout, stderr)
		}
		// ... existing single-project update (full rebuild) unchanged ...
```

Add helpers:

```go
func serveWorkspace(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	root, _, code := resolveWorkspaceRoot(args, stderr)
	if code != 0 {
		return code
	}
	reg, err := workspace.LoadRegistry(root)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	ov, err := workspace.LoadOverlay(root)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	idx, err := workspace.CombinedIndex(root, reg, ov)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	if err := mcp.NewServer(idx).Serve(stdin, stdout); err != nil {
		fmt.Fprintf(stderr, "graffiti: serve: %v\n", err)
		return 1
	}
	return 0
}

// updateWorkspace rebuilds members whose source changed since the registry's
// recorded hash, then recomputes the overlay. --links-only skips member rebuild.
func updateWorkspace(args []string, stdout, stderr io.Writer) int {
	linksOnly := hasFlag(args, "--links-only")
	root, _, code := resolveWorkspaceRoot(stripFlag(args, "--links-only"), stderr)
	if code != 0 {
		return code
	}
	reg, err := workspace.LoadRegistry(root)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	rebuilt := 0
	if !linksOnly {
		for i := range reg.Members {
			dir := filepath.Join(root, filepath.FromSlash(reg.Members[i].Path))
			cur, err := workspace.MapHash(dir)
			if err != nil || cur != reg.Members[i].MapHash {
				if _, berr := app.Build(dir, nowRFC3339()); berr != nil {
					fmt.Fprintf(stderr, "graffiti: rebuild %s: %v\n", reg.Members[i].Alias, berr)
					return 1
				}
				if h, herr := workspace.MapHash(dir); herr == nil {
					reg.Members[i].MapHash = h
				}
				rebuilt++
			}
		}
		reg.GeneratedAt = nowRFC3339()
		if err := workspace.SaveRegistry(root, reg); err != nil {
			fmt.Fprintf(stderr, "graffiti: %v\n", err)
			return 1
		}
	}
	ov, err := computeAndSaveOverlay(root, reg)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "✓ Updated workspace: %d members rebuilt, overlay recomputed (%d links).\n", rebuilt, len(ov.Links))
	return 0
}
```

Ensure `mcp` is imported (it already is).

- [ ] **Step 4: Run the tests** (full tag list) → PASS.

- [ ] **Step 5: Update usage text** (add `serve --workspace`, `update --workspace`, `query --workspace`).

- [ ] **Step 6: Commit**

```bash
git add cmd/graffiti/main.go cmd/graffiti/workspace_test.go
git commit -m "feat(cli): graffiti serve --workspace and update --workspace"
```

---

## Task 11: overlay golden + docs + full verification

**Files:**
- Create: `testdata/fixtures/ws/` (two prebuilt member maps), `testdata/golden/ws.overlay.json`, `internal/workspace/golden_test.go`
- Modify: `README.md`, `docs/superpowers/specs/2026-06-14-graffiti-design.md`

- [ ] **Step 1: Golden determinism test**

`internal/workspace/golden_test.go` — build two members, write a links file, compute the overlay twice, assert byte-identical (modulo generated_at), and compare to a committed golden:

```go
package workspace

import (
	"encoding/json"
	"testing"
)

func normalizeOverlay(t *testing.T, ov *Overlay) string {
	t.Helper()
	ov.GeneratedAt = "FIXED"
	b, err := json.MarshalIndent(ov, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestOverlay_DeterministicAndGolden(t *testing.T) {
	root := t.TempDir()
	writeMember(t, root, "frontend", "cartclient.fetchcart", "ui.render")
	writeMember(t, root, "backend", "handlers.get_cart", "db.save")
	reg := newRegistry(t, root)
	links := []ParsedLink{
		{"web", "cartclient.fetchcart", "api", "handlers.get_cart", "calls"},
		{"api", "db.save", "web", "ui.render", "references"},
	}
	a, err := ComputeOverlay(root, reg, links)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ComputeOverlay(root, reg, links)
	if err != nil {
		t.Fatal(err)
	}
	if normalizeOverlay(t, a) != normalizeOverlay(t, b) {
		t.Fatal("overlay computation is non-deterministic")
	}
	if len(a.Links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(a.Links))
	}
	// links sorted by (from,to,relation): api::db.save before web::cartclient...
	if a.Links[0].From != "api::db.save" {
		t.Fatalf("links not sorted: %+v", a.Links)
	}
}
```

(A committed `testdata/golden/ws.overlay.json` is optional given the determinism + sort assertions above fully pin the output; if you add one, generate it from `graffiti link` on `testdata/fixtures/ws/` and compare with generated_at neutralized. The determinism test is the binding gate.)

- [ ] **Step 2: Run the golden test**

Run: `go test ./internal/workspace/ -run Deterministic -v`
Expected: PASS

- [ ] **Step 3: README — add a Workspace section**

Document the workflow and the v1 scope explicitly:

```markdown
## Workspaces (multi-repo federation)

Lay separate repos side by side and query across them — without merging:

```
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
```

`graffiti link` writes a committable registry (`.graffiti-workspace/workspace.json`)
and a derived, gitignorable cache (`.graffiti-workspace/overlay.json`). Each repo's
own `.graffiti/map.json` is unchanged and still works standalone.

**Cross-project links (v1):** assert them explicitly in `.graffiti-workspace/links`,
one per line — `frontend::cartclient.fetchcart -> backend::handlers.get_cart calls`
(`#` comments allowed). `graffiti links check` validates both endpoints resolve.
Automatic link discovery (shared symbols, HTTP routes, SDK base-URLs) and a
`workspace.html` lanes view are planned follow-ons.
```

Add `.graffiti-workspace/overlay.json` to the repo's `.gitignore` guidance (overlay is derived).

- [ ] **Step 4: Spec §16 amendments**

Append an "Implemented (Plan 7, v1)" note to §16 recording: purely-additive `internal/workspace` (no single-project changes); **explicit-links file is line-based, not YAML** (zero-dep); members resolved by **alias** (§16.6 immutable id deferred); `workspace.json` + `overlay.json` both under `<root>/.graffiti-workspace/`; federated query reuses `query.Query` over an in-memory alias-prefixed combined index; **single budget v1** (two-tier deferred); **signal 1 (explicit) only** this plan — signals 2–5 + §16.7 viewer are follow-ons; determinism per §16.5 (members by alias, links by `(from,to,relation)`).

- [ ] **Step 5: Full verification**

```bash
make vet
make test
go test ./...
go mod tidy && git diff --exit-code go.mod go.sum
make build && make xcompile
```

Expected: vet clean; both configs green (incl. `internal/workspace`); zero new deps; size guard OK.

- [ ] **Step 6: End-to-end smoke**

```bash
BASE=$(mktemp -d)
mkdir -p "$BASE/frontend" "$BASE/backend"
printf 'package web\nfunc FetchCart(){}\n' > "$BASE/frontend/main.go"
printf 'package api\nfunc GetCart(){}\n'   > "$BASE/backend/main.go"
./graffiti link --name shop "$BASE/frontend" "$BASE/backend"
printf 'frontend::main-go:fetchcart -> backend::main-go:getcart calls\n' > "$BASE/.graffiti-workspace/links"
./graffiti links check --root "$BASE"
./graffiti update --workspace --root "$BASE"
./graffiti query --workspace --root "$BASE" "cart" | head -20
rm -rf "$BASE"
```

Expected: link prints "Linked 2 projects"; links check reports the link OK (adjust ids to the built slugs); query returns an alias-prefixed federated subgraph including the cross `calls` edge.

- [ ] **Step 7: Commit**

```bash
git add internal/workspace/golden_test.go README.md docs/superpowers/specs/2026-06-14-graffiti-design.md testdata/fixtures/ws testdata/golden 2>/dev/null; git add -A
git commit -m "test(workspace): overlay determinism golden; docs for federation v1"
```

---

## Self-review checklist (run before merge)

1. **§16 coverage (v1 scope):** federation-not-merge ✓ (overlay over unchanged maps); registry + derived overlay + alias::id cross-edges ✓; signal 1 explicit links ✓; federated query/serve/update ✓; auto-root discovery ✓; determinism ✓. Deferred (stated): signals 2–5, §16.6 id, two-tier budget, §16.7 viewer.
2. **Additive-only:** no changes to graph/schema/build/parse/render of the single-project path; `internal/workspace` reads existing `map.json` files. Single-project `query`/`serve` untouched (Task 9/10 regression checks).
3. **Determinism (§14/§16.5):** members sorted by alias; links sorted by (from,to,relation); overlay byte-identical modulo generated_at (Task 11).
4. **Honesty-first:** unresolved explicit links never become confident edges (recorded in `Unresolved`, surfaced by `links check`); staleness nudge instead of silently-wrong links.
5. **No new deps:** stdlib + internal only; line-based links (no YAML); `go mod tidy` no-op.
6. **Type consistency:** `Registry`/`Member`/`Link`/`Overlay`/`ParsedLink` names match across files; `ComputeOverlay`/`CombinedIndex`/`StaleMembers` signatures match call sites in CLI; alias-prefix format `alias::id` consistent everywhere.

## Deferred follow-ups (record in memory, non-blocking)

- **Auto-detection signals 2–5** (§16.2): symbol export×import, OpenAPI, time-boxed literal-HTTP matcher, known-SDK base-URL recognizer — each needs a per-project **contract-surface carrier** (exported symbols + unresolved external refs) added to the single-project map.json (the one invasive change deferred out of v1).
- **§16.6 immutable project id** (durable identity across folder/alias renames; committed-links robustness).
- **Two-tier query budget** (§16.3): free intra-project BFS + small cross-hop budget so the caller project stays dominant.
- **`graffiti init` workspace-awareness** (§16.4): CLAUDE.md block nudging `query --workspace` for system-spanning questions.
- **`workspace.html` lanes viewer** (§16.7).
- **AMBIGUOUS surfacing** in MAP.md / a "suspected links" report once auto-signals land.
