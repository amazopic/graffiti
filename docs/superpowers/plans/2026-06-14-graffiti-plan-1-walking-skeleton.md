# graffiti Plan 1 — Walking Skeleton (Go-only `map.json`) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `graffiti .` (alias `graffiti build .`) run on a Go repository produces a deterministic, schema-valid `.graffiti/map.json` containing the directed knowledge graph of that Go code — functions/methods/types/files as nodes; `imports` (EXTRACTED), `contains`, and `calls` (EXTRACTED if import-backed else INFERRED) as edges — printing one success line.

**Architecture:** A pure functional pipeline `scan → parse → build → render(json-only)`. Each stage is a package under `internal/` that consumes plain structs and returns plain structs with no shared mutable state. Parsing uses the verified pure-Go tree-sitter runtime `github.com/odvcencio/gotreesitter` (no CGO, no WASM) behind graffiti's own swappable `parse.Parser` interface, so the backend can later fall back to wazero+WASM per-language without touching the rest of the pipeline. Only Go and Markdown are in scope for Plan 1 (Markdown files are discovered and emitted as `doc` nodes but not parsed); cluster/analyze/MAP.md/map.html/query/mcp/init/workspace are explicitly later plans.

> **SPEC-DEVIATION NOTICE (must be ratified before execution — see Open Issues #1).** Spec §10 mandates parsing via *"tree-sitter grammars compiled to WASM, executed via wazero, grammars embedded via `embed.FS`."* This plan instead adopts the **native pure-Go** tree-sitter runtime `github.com/odvcencio/gotreesitter` (its grammars are compiled-to-Go tables embedded by the library's own `go:embed`; there is **no** wazero and **no** graffiti-authored `embed.FS` of `.wasm`). This satisfies every *hard* constraint (pure-Go, no CGO, single static cross-compilable binary, offline, deterministic — all verified at v0.20.2) but **substitutes the specific technology §10 names**. The substitution is isolated behind the `parse.Parser` interface (`internal/parse/parser.go`), so the wazero+WASM path remains a drop-in per-language fallback affecting only `internal/parse/gotreesitter.go`. **Do not begin Task 6 until the owner records a decision:** (a) amend §10 to permit the gotreesitter backend (recommended), or (b) require the wazero+WASM `embed.FS` path, in which case Task 6 Step 5 is replaced (every other task is unaffected because they code only against the interface). Tasks 1–5 are independent of this decision and may proceed immediately.

**Tech Stack:** Go 1.26; `github.com/odvcencio/gotreesitter` v0.20.2 (pinned, pure-Go tree-sitter runtime, built with `grammar_subset` + `grammar_subset_go` + `grammar_subset_gomod` build tags for a lean ~8MB binary); `github.com/sabhiram/go-gitignore` for `.gitignore` matching; `golang.org/x/text` v0.21.0 for NFC normalization; Go stdlib `testing` for all tests; `encoding/json` for the writer; `crypto/sha256` for the content-hash cache. No CGO (`CGO_ENABLED=0` everywhere). No third-party JSON-schema validator — we ship `schema/map.schema.json` as the published contract and validate shape with a small in-repo structural validator so the binary stays dependency-light.

---

## Verified Spike Facts (load-bearing; confirmed by compiling & running real code against `github.com/odvcencio/gotreesitter@v0.20.2`)

Use these verbatim; do **not** trust generic spike snippets where they differ:

- Root package: `import ts "github.com/odvcencio/gotreesitter"`. Grammar registry: `import "github.com/odvcencio/gotreesitter/grammars"`.
- `grammars.GoLanguage() *ts.Language` returns the Go language directly. `grammars.DetectLanguage(filename string) *grammars.LangEntry` returns an entry whose `.Language` is a **field of type `func() *ts.Language`** (call it: `entry.Language()`), and `.Name` is the language name (e.g. `"go"`). Returns `nil` if unknown.
- `parser := ts.NewParser(lang)`; `tree, err := parser.Parse(src []byte) (*ts.Tree, error)`. Tree/Parser have **no `Close()`**; `*ts.Tree` has an optional `Release()` for arena pooling which we do **not** need to call (GC handles it). Do not call any teardown.
- `tree.RootNode() *ts.Node`.
- Node API: `n.Type(lang *ts.Language) string` (there is **no** `n.Kind()`); `n.StartByte() uint32`; `n.EndByte() uint32`; `n.StartPoint() ts.Point` where `ts.Point{Row, Column uint32}` is **0-based** (add 1 for the 1-based `Node.line`); `n.Text(source []byte) string`; `n.ChildByFieldName(name string, lang *ts.Language) *ts.Node` (may return `nil`); `n.NamedChild(i int) *ts.Node`; `n.NamedChildCount() int`.
- Go grammar field names (verified): `import_spec` has field `path` (text is a **quoted** string e.g. `"fmt"`); `function_declaration` has field `name`; `method_declaration` has fields `name` and `receiver`; `type_spec` has field `name`; `call_expression` has field `function` (text is either a bare ident `Hello` or a selector `fmt.Sprintf`). A pointer receiver `(g *Greeter)` still exposes a `type_identifier` child `Greeter`.
- Verified spike node lines for the sample in Task 6: `import_spec`=3, `function_declaration`=5, `type_spec`=9, `method_declaration`=11.
- Build tags: `-tags "grammar_subset grammar_subset_go grammar_subset_gomod"`. These tags are required only to hit the **~8MB binary-size goal (§10)** — *without* them the code still compiles and tests still pass, but the binary links the full multi-grammar set (~31MB). Always pass the tags via the `Makefile`. The `grammar_subset_gomod` tag keeps `grammars.DetectLanguage` total for `go.mod` files; it is harmless. With the subset tags the binary is ~8MB and cross-compiles with `CGO_ENABLED=0` to darwin/arm64, linux/amd64, windows/amd64 (verified).

---

## File Structure

```
graffiti/
├── go.mod                                  # module github.com/evgeniy-achin/graffiti ; go 1.26 ; pinned deps
├── go.sum                                  # checksums (generated)
├── .gitignore                              # ignore built binaries + /dist
├── Makefile                                # encodes the required build tags + xcompile
├── README.md
├── cmd/graffiti/
│   ├── main.go                             # CLI entry: `graffiti .` / `graffiti build .` / `graffiti build <path>`; calls app.Build; prints success line
│   └── main_test.go
├── internal/
│   ├── app/
│   │   ├── app.go                          # Build(root, generatedAt) orchestrates scan→parse→build→write; returns Stats
│   │   ├── app_test.go
│   │   └── golden_test.go                  # golden + determinism end-to-end tests
│   ├── scan/
│   │   ├── scan.go                         # Scan(root) []FileRef: walk, .gitignore, ext filter (.go/.md), deterministic sort
│   │   └── scan_test.go
│   ├── parse/
│   │   ├── parser.go                       # Parser/Tree/Node interface + Walk helper (backend-agnostic)
│   │   ├── gotreesitter.go                 # gotreesitter-backed implementation behind the interface
│   │   ├── golang.go                       # ParseGo: Pass-1 per-file extraction (defs + imports/contains + raw calls)
│   │   ├── resolve.go                      # Pass-2 cross-file call resolution (label index, EXTRACTED/INFERRED, drop-ambiguous)
│   │   ├── parser_test.go                  # the SPIKE test: one Go file → AST node kinds + lines
│   │   ├── golang_test.go                  # Pass-1 unit tests
│   │   └── resolve_test.go                 # Pass-2 unit tests
│   ├── graph/
│   │   ├── graph.go                        # Node, Edge, Document, Community types + enums + NewDocument
│   │   ├── id.go                           # NormalizeID (NFC + casefold + \w-collapse) + NodeID builder
│   │   ├── unicode_class.go                # localized unicode import
│   │   ├── merge.go                        # Merge(into, from, allowPrune) with merge-not-replace + anti-shrink guard
│   │   ├── graph_test.go
│   │   ├── id_test.go
│   │   └── merge_test.go
│   ├── build/
│   │   ├── build.go                        # Assemble(root, generatedAt, []Extraction) (*graph.Document, error)
│   │   └── build_test.go
│   ├── schemaval/
│   │   ├── schemaval.go                    # ValidateDocument(*graph.Document) error — structural validation
│   │   └── schemaval_test.go
│   ├── render/
│   │   ├── json.go                         # WriteMapJSON(doc, dir) — sorted keys/arrays, stable formatting, reads doc.GeneratedAt
│   │   └── json_test.go
│   └── cache/
│       ├── cache.go                        # HashBytes/HashFile + content-hash cache read/write under .graffiti/cache/
│       └── cache_test.go
├── schema/
│   ├── map.schema.json                     # published JSON Schema (draft 2020-12) for Document
│   └── schema_test.go
└── testdata/
    ├── fixtures/
    │   └── gorepo/                          # small committed Go fixture repo (golden source)
    │       ├── go.mod
    │       ├── main.go
    │       └── greet/greet.go
    │       └── greet/greet_helper.go
    └── golden/
        └── gorepo.map.json                 # golden expected map.json (modulo generated_at + root)
```

**Package responsibilities (one job each):**
- `graph` owns the data model and ID normalization — no I/O, no parsing.
- `scan` owns filesystem discovery only.
- `parse` owns AST → extraction (both passes) behind a swappable interface.
- `build` owns assembly + validation + determinism ordering, and stamps `generated_at`.
- `schemaval` owns structural validation against the published schema rules.
- `render` owns serialization to disk (reads `generated_at` off the document).
- `cache` owns content hashing.
- `app` wires them; `cmd/graffiti` is a thin CLI shell.

---

## Task 1: Repo bootstrap — Go module, CLI skeleton, first passing test

**Files:**
- Create: `go.mod`
- Create: `.gitignore`
- Create: `cmd/graffiti/main.go`
- Create: `cmd/graffiti/main_test.go`

> Note: the working directory `/Users/mylive/project/graffiti/graffiti` is ALREADY a git repo on branch `main` with the design docs committed. Do **not** run `git init`. Branch off `main`.

- [ ] **Step 1: Create the Go module**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti && go mod init github.com/evgeniy-achin/graffiti
```
Expected: prints `go: creating new go.mod: module github.com/evgeniy-achin/graffiti` and creates `go.mod`.

Then edit `go.mod` so it reads exactly:
```
module github.com/evgeniy-achin/graffiti

go 1.26
```

- [ ] **Step 2: Create `.gitignore`**

Create `/Users/mylive/project/graffiti/graffiti/.gitignore`:
```gitignore
# built binaries
/graffiti
/graffiti.exe
/dist/
*.test
# MCP scratch dir (already present in the working tree)
/.playwright-mcp/
```

- [ ] **Step 3: Write the failing CLI test**

Create `/Users/mylive/project/graffiti/graffiti/cmd/graffiti/main_test.go`:
```go
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
```

- [ ] **Step 4: Run the test to verify it fails**

Run: `go test ./cmd/graffiti/ -run TestRun -v`
Expected: FAIL — compilation error `undefined: run`.

- [ ] **Step 5: Write the minimal CLI implementation**

Create `/Users/mylive/project/graffiti/graffiti/cmd/graffiti/main.go`:
```go
// Command graffiti turns a code repository into a queryable directed knowledge graph.
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

// run is the testable entry point. It returns the process exit code.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		usage(stderr)
		return 2
	}

	cmd := args[1]
	switch cmd {
	case ".":
		return runBuild(".", stdout, stderr)
	case "build":
		root := "."
		if len(args) >= 3 {
			root = args[2]
		}
		return runBuild(root, stdout, stderr)
	default:
		// Treat an existing path as `build <path>` for the common `graffiti <path>` form.
		if info, err := os.Stat(cmd); err == nil && info.IsDir() {
			return runBuild(cmd, stdout, stderr)
		}
		fmt.Fprintf(stderr, "graffiti: unknown command %q\n", cmd)
		usage(stderr)
		return 2
	}
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "usage: graffiti <command>")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "commands:")
	fmt.Fprintln(w, "  .              build the map for the current repo")
	fmt.Fprintln(w, "  build <path>   build the map for <path> (default .)")
}

// runBuild is a stub so the CLI compiles; Task 13 Step 5 replaces this body verbatim.
func runBuild(root string, stdout, stderr io.Writer) int {
	fmt.Fprintln(stderr, "graffiti: build not yet wired")
	return 1
}
```

- [ ] **Step 6: Run the test to verify it passes**

Run: `go test ./cmd/graffiti/ -run TestRun -v`
Expected: PASS (both tests).

- [ ] **Step 7: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git checkout -b plan1-walking-skeleton
git add go.mod .gitignore cmd/graffiti/main.go cmd/graffiti/main_test.go
git commit -m "feat: bootstrap go module and graffiti CLI skeleton

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: Core graph types + published JSON Schema + shape tests

**Files:**
- Create: `internal/graph/graph.go`
- Create: `internal/graph/graph_test.go`
- Create: `schema/map.schema.json`
- Create: `schema/schema_test.go`

- [ ] **Step 1: Write the failing graph types test**

Create `/Users/mylive/project/graffiti/graffiti/internal/graph/graph_test.go`:
```go
package graph

import "testing"

func TestNewDocument_Defaults(t *testing.T) {
	d := NewDocument("myrepo")
	if d.Version != SchemaVersion {
		t.Fatalf("version = %q, want %q", d.Version, SchemaVersion)
	}
	if d.Root != "myrepo" {
		t.Fatalf("root = %q, want %q", d.Root, "myrepo")
	}
	if d.Nodes == nil || d.Edges == nil || d.Communities == nil {
		t.Fatalf("slices must be non-nil (got nodes=%v edges=%v comms=%v)", d.Nodes, d.Edges, d.Communities)
	}
	if len(d.Nodes) != 0 || len(d.Edges) != 0 {
		t.Fatalf("new document must start empty")
	}
}

func TestKindAndRelationConstants(t *testing.T) {
	kinds := []Kind{KindFunction, KindMethod, KindClass, KindModule, KindFile, KindDoc, KindConcept}
	want := []string{"function", "method", "class", "module", "file", "doc", "concept"}
	for i, k := range kinds {
		if string(k) != want[i] {
			t.Fatalf("kind[%d] = %q, want %q", i, k, want[i])
		}
	}
	rels := []Relation{RelCalls, RelImports, RelInherits, RelImplements, RelReferences, RelContains}
	wantRel := []string{"calls", "imports", "inherits", "implements", "references", "contains"}
	for i, r := range rels {
		if string(r) != wantRel[i] {
			t.Fatalf("relation[%d] = %q, want %q", i, r, wantRel[i])
		}
	}
	confs := []Confidence{ConfExtracted, ConfInferred, ConfAmbiguous}
	wantConf := []string{"EXTRACTED", "INFERRED", "AMBIGUOUS"}
	for i, c := range confs {
		if string(c) != wantConf[i] {
			t.Fatalf("confidence[%d] = %q, want %q", i, c, wantConf[i])
		}
	}
}

func TestNode_DefaultCommunityIsMinusOne(t *testing.T) {
	n := Node{ID: "a", Label: "A", Kind: KindFunction, File: "a.go", Line: 1, Community: UnclusteredCommunity}
	if n.Community != -1 {
		t.Fatalf("unclustered community sentinel = %d, want -1", n.Community)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/graph/ -run TestNewDocument -v`
Expected: FAIL — `undefined: NewDocument` etc.

- [ ] **Step 3: Write the graph types implementation**

Create `/Users/mylive/project/graffiti/graffiti/internal/graph/graph.go`:
```go
// Package graph defines graffiti's directed knowledge-graph data model (spec §6).
// It performs no I/O and no parsing.
package graph

// SchemaVersion is the version string stamped into every Document and matched by
// the published schema/map.schema.json.
const SchemaVersion = "1"

// UnclusteredCommunity is the Community value for a node before clustering (spec §6).
const UnclusteredCommunity = -1

// Kind enumerates the node kinds (spec §6).
type Kind string

const (
	KindFunction Kind = "function"
	KindMethod   Kind = "method"
	KindClass    Kind = "class"
	KindModule   Kind = "module"
	KindFile     Kind = "file"
	KindDoc      Kind = "doc"
	KindConcept  Kind = "concept"
)

// ValidKinds is the closed set of allowed kinds.
var ValidKinds = map[Kind]bool{
	KindFunction: true, KindMethod: true, KindClass: true, KindModule: true,
	KindFile: true, KindDoc: true, KindConcept: true,
}

// Relation enumerates the edge relations (spec §6).
type Relation string

const (
	RelCalls      Relation = "calls"
	RelImports    Relation = "imports"
	RelInherits   Relation = "inherits"
	RelImplements Relation = "implements"
	RelReferences Relation = "references"
	RelContains   Relation = "contains"
)

// ValidRelations is the closed set of allowed relations.
var ValidRelations = map[Relation]bool{
	RelCalls: true, RelImports: true, RelInherits: true, RelImplements: true,
	RelReferences: true, RelContains: true,
}

// Confidence enumerates the edge confidence ladder (spec §5/§6).
type Confidence string

const (
	ConfExtracted Confidence = "EXTRACTED"
	ConfInferred  Confidence = "INFERRED"
	ConfAmbiguous Confidence = "AMBIGUOUS"
)

// ValidConfidences is the closed set of allowed confidence values.
var ValidConfidences = map[Confidence]bool{
	ConfExtracted: true, ConfInferred: true, ConfAmbiguous: true,
}

// Node is a vertex in the directed graph (spec §6).
type Node struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Kind      Kind   `json:"kind"`
	File      string `json:"file"`
	Line      int    `json:"line"`      // 1-based
	Community int    `json:"community"` // -1 before clustering
}

// Edge is a directed edge in the graph (spec §6).
type Edge struct {
	From       string     `json:"from"`
	To         string     `json:"to"`
	Relation   Relation   `json:"relation"`
	Confidence Confidence `json:"confidence"`
}

// Community is a cluster of nodes (populated by a later plan; empty in Plan 1).
type Community struct {
	ID      int      `json:"id"`
	Label   string   `json:"label"`
	Members []string `json:"members"`
}

// Document is the on-disk shape of .graffiti/map.json (spec §6).
type Document struct {
	Version     string      `json:"version"`
	GeneratedAt string      `json:"generated_at"` // RFC3339, stamped by build.Assemble
	Root        string      `json:"root"`
	Nodes       []Node      `json:"nodes"`
	Edges       []Edge      `json:"edges"`
	Communities []Community `json:"communities"`
}

// NewDocument returns an empty Document with non-nil slices and the current schema version.
func NewDocument(root string) *Document {
	return &Document{
		Version:     SchemaVersion,
		Root:        root,
		Nodes:       []Node{},
		Edges:       []Edge{},
		Communities: []Community{},
	}
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/graph/ -run 'TestNewDocument|TestKind|TestNode_' -v`
Expected: PASS (3 tests).

- [ ] **Step 5: Write the failing schema test**

Create `/Users/mylive/project/graffiti/graffiti/schema/schema_test.go`:
```go
package schema_test

import (
	"encoding/json"
	"os"
	"testing"
)

func TestMapSchemaIsValidJSON(t *testing.T) {
	b, err := os.ReadFile("map.schema.json")
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}
	if doc["$schema"] != "https://json-schema.org/draft/2020-12/schema" {
		t.Fatalf("$schema = %v, want draft 2020-12", doc["$schema"])
	}
	props, ok := doc["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema has no properties object")
	}
	for _, key := range []string{"version", "generated_at", "root", "nodes", "edges", "communities"} {
		if _, ok := props[key]; !ok {
			t.Fatalf("schema missing top-level property %q", key)
		}
	}
}
```

- [ ] **Step 6: Run the schema test to verify it fails**

Run: `go test ./schema/ -run TestMapSchema -v`
Expected: FAIL — `read schema: open map.schema.json: no such file or directory`.

- [ ] **Step 7: Write the published JSON Schema**

Create `/Users/mylive/project/graffiti/graffiti/schema/map.schema.json`:
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://graffiti.dev/schema/map.schema.json",
  "title": "graffiti map.json",
  "description": "Directed knowledge graph emitted by graffiti for a code repository.",
  "type": "object",
  "additionalProperties": false,
  "required": ["version", "generated_at", "root", "nodes", "edges", "communities"],
  "properties": {
    "version": { "type": "string", "const": "1" },
    "generated_at": { "type": "string", "format": "date-time" },
    "root": { "type": "string" },
    "nodes": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["id", "label", "kind", "file", "line", "community"],
        "properties": {
          "id": { "type": "string", "minLength": 1 },
          "label": { "type": "string" },
          "kind": { "type": "string", "enum": ["function", "method", "class", "module", "file", "doc", "concept"] },
          "file": { "type": "string" },
          "line": { "type": "integer", "minimum": 0 },
          "community": { "type": "integer", "minimum": -1 }
        }
      }
    },
    "edges": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["from", "to", "relation", "confidence"],
        "properties": {
          "from": { "type": "string", "minLength": 1 },
          "to": { "type": "string", "minLength": 1 },
          "relation": { "type": "string", "enum": ["calls", "imports", "inherits", "implements", "references", "contains"] },
          "confidence": { "type": "string", "enum": ["EXTRACTED", "INFERRED", "AMBIGUOUS"] }
        }
      }
    },
    "communities": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["id", "label", "members"],
        "properties": {
          "id": { "type": "integer" },
          "label": { "type": "string" },
          "members": { "type": "array", "items": { "type": "string" } }
        }
      }
    }
  }
}
```

- [ ] **Step 8: Run the schema test to verify it passes**

Run: `go test ./schema/ -run TestMapSchema -v`
Expected: PASS.

- [ ] **Step 9: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/graph/graph.go internal/graph/graph_test.go schema/map.schema.json schema/schema_test.go
git commit -m "feat: core graph data model and published map.schema.json

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: Deterministic node ID normalization

**Files:**
- Create: `internal/graph/id.go`
- Create: `internal/graph/unicode_class.go`
- Create: `internal/graph/id_test.go`

- [ ] **Step 1: Write the failing ID test**

Create `/Users/mylive/project/graffiti/graffiti/internal/graph/id_test.go`:
```go
package graph

import "testing"

func TestNormalizeID(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Hello", "hello"},
		{"HTTPRouting", "httprouting"},
		{"Auth & Sessions", "auth-sessions"},
		{"foo__bar", "foo-bar"},
		{"  leading-trailing  ", "leading-trailing"},
		{"a/b/c.go", "a-b-c-go"},
		{"already-clean", "already-clean"},
		{"Multiple   Spaces", "multiple-spaces"},
		{"", ""},
		{"!!!", ""},
	}
	for _, c := range cases {
		if got := NormalizeID(c.in); got != c.want {
			t.Errorf("NormalizeID(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNodeID_StableAndQualified(t *testing.T) {
	a := NodeID("greet/greet.go", "Hello")
	b := NodeID("greet/greet.go", "Hello")
	if a != b {
		t.Fatalf("NodeID not stable: %q vs %q", a, b)
	}
	c := NodeID("main.go", "Hello")
	if a == c {
		t.Fatalf("NodeID should differ across files: both %q", a)
	}
	if a != "greet-greet-go:hello" {
		t.Fatalf("NodeID format = %q, want %q", a, "greet-greet-go:hello")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/graph/ -run 'TestNormalizeID|TestNodeID' -v`
Expected: FAIL — `undefined: NormalizeID`, `undefined: NodeID`.

- [ ] **Step 3: Write the implementation**

Create `/Users/mylive/project/graffiti/graffiti/internal/graph/id.go`:
```go
package graph

import (
	"strings"

	"golang.org/x/text/unicode/norm"
)

// NormalizeID produces a deterministic slug per spec §6:
// NFC normalize, casefold (lowercase), collapse every run of non-word
// characters to a single '-', and trim leading/trailing '-'.
func NormalizeID(s string) string {
	s = norm.NFC.String(s)
	s = strings.ToLower(s)

	var b strings.Builder
	b.Grow(len(s))
	prevDash := false
	for _, r := range s {
		if isWordRune(r) {
			b.WriteRune(r)
			prevDash = false
		} else {
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// isWordRune reports whether r is a letter or digit (kept verbatim) for ID purposes.
func isWordRune(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z':
		return true
	case r >= '0' && r <= '9':
		return true
	case r > 0x7F:
		return unicodeIsLetterOrDigit(r)
	default:
		return false
	}
}

// NodeID builds a deterministic, file-qualified node id: "<file>:<label>", both
// normalized. File qualification prevents collisions between identically named
// symbols in different files.
func NodeID(file, label string) string {
	return NormalizeID(file) + ":" + NormalizeID(label)
}
```

Create `/Users/mylive/project/graffiti/graffiti/internal/graph/unicode_class.go`:
```go
package graph

import "unicode"

func unicodeIsLetterOrDigit(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
```

- [ ] **Step 4: Add the `golang.org/x/text` dependency**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti && GOFLAGS=-mod=mod go get golang.org/x/text@v0.21.0
```
Expected: `go: added golang.org/x/text v0.21.0`. Small, pure-Go dependency for NFC normalization required by spec §6.

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/graph/ -run 'TestNormalizeID|TestNodeID' -v`
Expected: PASS (2 tests).

- [ ] **Step 6: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/graph/id.go internal/graph/unicode_class.go internal/graph/id_test.go go.mod go.sum
git commit -m "feat: deterministic NFC+casefold node ID normalization

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: Merge-not-replace with anti-shrink guard

**Files:**
- Create: `internal/graph/merge.go`
- Create: `internal/graph/merge_test.go`

- [ ] **Step 1: Write the failing merge test**

Create `/Users/mylive/project/graffiti/graffiti/internal/graph/merge_test.go`:
```go
package graph

import "testing"

func TestMerge_AddsNewNodesAndEdges(t *testing.T) {
	into := NewDocument("repo")
	into.Nodes = append(into.Nodes, Node{ID: "a", Label: "A", Kind: KindFile, File: "a.go", Line: 1, Community: -1})

	from := NewDocument("repo")
	from.Nodes = append(from.Nodes,
		Node{ID: "a", Label: "A", Kind: KindFile, File: "a.go", Line: 1, Community: -1}, // dup
		Node{ID: "b", Label: "B", Kind: KindFunction, File: "a.go", Line: 5, Community: -1},
	)
	from.Edges = append(from.Edges, Edge{From: "a", To: "b", Relation: RelContains, Confidence: ConfExtracted})

	if err := Merge(into, from, false); err != nil {
		t.Fatalf("merge: %v", err)
	}
	if len(into.Nodes) != 2 {
		t.Fatalf("nodes = %d, want 2 (dedup of a)", len(into.Nodes))
	}
	if len(into.Edges) != 1 {
		t.Fatalf("edges = %d, want 1", len(into.Edges))
	}
}

func TestMerge_AntiShrinkGuard(t *testing.T) {
	into := NewDocument("repo")
	into.Nodes = append(into.Nodes,
		Node{ID: "a", Label: "A", Kind: KindFile, File: "a.go", Line: 1, Community: -1},
		Node{ID: "b", Label: "B", Kind: KindFile, File: "b.go", Line: 1, Community: -1},
	)
	from := NewDocument("repo") // empty

	// allowPrune=false must refuse a result smaller than `into`.
	// (No node is ever removed by Merge; the guard fires only if the result count
	// is below the pre-merge count, which our additive Merge can never do — so we
	// emulate shrink by pre-pruning `into` and asserting the guard semantics.)
	into2 := NewDocument("repo")
	if err := Merge(into2, into, false); err != nil {
		t.Fatalf("seeding merge failed: %v", err)
	}
	// Now drop a node out-of-band, then merge an empty `from`: result (1) < before (2).
	into2.Nodes = into2.Nodes[:1]
	err := Merge(into2, from, false)
	if err == nil {
		t.Fatalf("expected anti-shrink error when result node count is below pre-merge count")
	}

	// With allowPrune=true the guard is bypassed (explicit prune).
	into3 := NewDocument("repo")
	if err := Merge(into3, into, true); err != nil {
		t.Fatalf("seeding merge failed: %v", err)
	}
	into3.Nodes = into3.Nodes[:1]
	if err := Merge(into3, from, true); err != nil {
		t.Fatalf("with allowPrune the merge must succeed: %v", err)
	}
}

func TestMerge_DedupEdges(t *testing.T) {
	into := NewDocument("repo")
	into.Edges = append(into.Edges, Edge{From: "a", To: "b", Relation: RelCalls, Confidence: ConfInferred})
	from := NewDocument("repo")
	from.Edges = append(from.Edges, Edge{From: "a", To: "b", Relation: RelCalls, Confidence: ConfInferred})
	if err := Merge(into, from, true); err != nil {
		t.Fatalf("merge: %v", err)
	}
	if len(into.Edges) != 1 {
		t.Fatalf("edges = %d, want 1 (dedup identical edge)", len(into.Edges))
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/graph/ -run TestMerge -v`
Expected: FAIL — `undefined: Merge`.

- [ ] **Step 3: Write the implementation**

Create `/Users/mylive/project/graffiti/graffiti/internal/graph/merge.go`:
```go
package graph

import "fmt"

// Merge folds `from` into `into` (spec §6: merge-not-replace with anti-shrink
// guard). Nodes are deduped by ID (first writer wins on conflicting fields).
// Edges are deduped by the (from,to,relation,confidence) tuple.
//
// If allowPrune is false and the resulting node count would be LESS than the
// pre-merge node count of `into`, Merge returns an error rather than silently
// shrinking the graph. allowPrune=true is the explicit-prune escape hatch.
func Merge(into, from *Document, allowPrune bool) error {
	before := len(into.Nodes)

	nodeIdx := make(map[string]bool, len(into.Nodes))
	for _, n := range into.Nodes {
		nodeIdx[n.ID] = true
	}
	for _, n := range from.Nodes {
		if !nodeIdx[n.ID] {
			into.Nodes = append(into.Nodes, n)
			nodeIdx[n.ID] = true
		}
	}

	edgeIdx := make(map[edgeKey]bool, len(into.Edges))
	for _, e := range into.Edges {
		edgeIdx[keyOf(e)] = true
	}
	for _, e := range from.Edges {
		k := keyOf(e)
		if !edgeIdx[k] {
			into.Edges = append(into.Edges, e)
			edgeIdx[k] = true
		}
	}

	if !allowPrune && len(into.Nodes) < before {
		return fmt.Errorf("merge would shrink node count from %d to %d without explicit prune", before, len(into.Nodes))
	}
	return nil
}

type edgeKey struct {
	from, to string
	rel      Relation
	conf     Confidence
}

func keyOf(e Edge) edgeKey {
	return edgeKey{from: e.From, to: e.To, rel: e.Relation, conf: e.Confidence}
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/graph/ -run TestMerge -v`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/graph/merge.go internal/graph/merge_test.go
git commit -m "feat: merge-not-replace with anti-shrink guard

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: File discovery (`internal/scan`) — .gitignore, extension filter, deterministic order

**Files:**
- Create: `internal/scan/scan.go`
- Create: `internal/scan/scan_test.go`

- [ ] **Step 1: Add the gitignore dependency**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti && GOFLAGS=-mod=mod go get github.com/sabhiram/go-gitignore@v0.0.0-20210923224102-525f6e181f06
```
Expected: `go: added github.com/sabhiram/go-gitignore ...`. Pure-Go, parses `.gitignore` patterns (`CompileIgnoreFile` / `MatchesPath`, both verified to exist).

- [ ] **Step 2: Write the failing scan test**

Create `/Users/mylive/project/graffiti/graffiti/internal/scan/scan_test.go`:
```go
package scan

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScan_FiltersExtensionsAndSortsDeterministically(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "zebra.go", "package main")
	writeFile(t, dir, "alpha.go", "package main")
	writeFile(t, dir, "README.md", "# hi")
	writeFile(t, dir, "notes.txt", "ignored ext")
	writeFile(t, dir, "img.png", "binary")
	writeFile(t, dir, "sub/deep.go", "package sub")

	refs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	var rels []string
	for _, r := range refs {
		rels = append(rels, r.RelPath)
	}
	want := []string{"README.md", "alpha.go", "sub/deep.go", "zebra.go"}
	if !reflect.DeepEqual(rels, want) {
		t.Fatalf("scan order = %v, want %v", rels, want)
	}
}

func TestScan_HonorsGitignore(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".gitignore", "ignored/\n*.gen.go\n")
	writeFile(t, dir, "keep.go", "package main")
	writeFile(t, dir, "thing.gen.go", "package main")
	writeFile(t, dir, "ignored/secret.go", "package secret")

	refs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	var rels []string
	for _, r := range refs {
		rels = append(rels, r.RelPath)
	}
	want := []string{"keep.go"}
	if !reflect.DeepEqual(rels, want) {
		t.Fatalf("gitignore not honored: got %v, want %v", rels, want)
	}
}

func TestScan_AlwaysSkipsVendorDirs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".git/config.go", "package x")            // .git must never be scanned
	writeFile(t, dir, ".graffiti/cache/x.go", "package x")       // our own output dir
	writeFile(t, dir, "node_modules/dep/index.go", "package x")  // always skipped
	writeFile(t, dir, "vendor/v/v.go", "package x")              // always skipped
	writeFile(t, dir, "real.go", "package main")

	refs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(refs) != 1 || refs[0].RelPath != "real.go" {
		t.Fatalf("must skip .git/.graffiti/node_modules/vendor; got %+v", refs)
	}
}

func TestScan_LangClassification(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", "package main")
	writeFile(t, dir, "b.md", "# x")
	refs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	got := map[string]Lang{}
	for _, r := range refs {
		got[r.RelPath] = r.Lang
	}
	if got["a.go"] != LangGo {
		t.Fatalf("a.go lang = %q, want go", got["a.go"])
	}
	if got["b.md"] != LangMarkdown {
		t.Fatalf("b.md lang = %q, want markdown", got["b.md"])
	}
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./internal/scan/ -run TestScan -v`
Expected: FAIL — `undefined: Scan`, `undefined: Lang`, etc.

- [ ] **Step 4: Write the implementation**

Create `/Users/mylive/project/graffiti/graffiti/internal/scan/scan.go`:
```go
// Package scan discovers and classifies source files under a repository root,
// honoring .gitignore, filtering to supported extensions, and returning a
// deterministically ordered slice (spec §5 scan stage).
package scan

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

// Lang classifies a discovered file.
type Lang string

const (
	LangGo       Lang = "go"
	LangMarkdown Lang = "markdown"
)

// extLang maps supported file extensions to their language (Plan 1 scope:
// Go + Markdown only).
var extLang = map[string]Lang{
	".go": LangGo,
	".md": LangMarkdown,
}

// FileRef is a discovered, classified file.
type FileRef struct {
	AbsPath string // absolute path on disk
	RelPath string // path relative to root, slash-separated
	Lang    Lang
}

// alwaysSkipDirs are directory names never descended into, regardless of .gitignore.
var alwaysSkipDirs = map[string]bool{
	".git":         true,
	".graffiti":    true,
	"node_modules": true,
	"vendor":       true,
}

// Scan walks root and returns supported files in deterministic order
// (by RelPath, slash-separated, ascending).
func Scan(root string) ([]FileRef, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	ign := loadGitignore(absRoot)

	var refs []FileRef
	walkErr := filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == absRoot {
			return nil
		}
		rel, relErr := filepath.Rel(absRoot, path)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)

		if d.IsDir() {
			if alwaysSkipDirs[d.Name()] {
				return filepath.SkipDir
			}
			if ign != nil && ign.MatchesPath(rel+"/") {
				return filepath.SkipDir
			}
			return nil
		}

		if ign != nil && ign.MatchesPath(rel) {
			return nil
		}
		lang, ok := extLang[strings.ToLower(filepath.Ext(rel))]
		if !ok {
			return nil
		}
		refs = append(refs, FileRef{AbsPath: path, RelPath: rel, Lang: lang})
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	sort.Slice(refs, func(i, j int) bool { return refs[i].RelPath < refs[j].RelPath })
	return refs, nil
}

// loadGitignore reads <root>/.gitignore if present. Returns nil if there is none.
func loadGitignore(absRoot string) *gitignore.GitIgnore {
	p := filepath.Join(absRoot, ".gitignore")
	if _, err := os.Stat(p); err != nil {
		return nil
	}
	ign, err := gitignore.CompileIgnoreFile(p)
	if err != nil {
		return nil
	}
	return ign
}
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/scan/ -run TestScan -v`
Expected: PASS (4 tests).

- [ ] **Step 6: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/scan/scan.go internal/scan/scan_test.go go.mod go.sum
git commit -m "feat: deterministic file discovery with gitignore and ext filter

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: THE SPIKE — `internal/parse` interface + gotreesitter backend, one Go file to AST

> **GATE:** Do not start this task until the §10 spec-deviation decision (Open Issues #1) is recorded. If the owner requires the wazero+WASM path, only Step 5 (`gotreesitter.go`) changes — the interface in `parser.go` and all downstream tasks are unaffected.

**Files:**
- Create: `internal/parse/parser.go`
- Create: `internal/parse/gotreesitter.go`
- Create: `internal/parse/parser_test.go`

- [ ] **Step 1: Add and pin the gotreesitter dependency**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
GOFLAGS=-mod=mod go get github.com/odvcencio/gotreesitter@v0.20.2
```
Expected: `go: added github.com/odvcencio/gotreesitter v0.20.2` (the `grammars` subpackage resolves with the same module).

- [ ] **Step 2: Write the failing spike test**

Create `/Users/mylive/project/graffiti/graffiti/internal/parse/parser_test.go`:
```go
package parse

import (
	"testing"
)

const sampleGo = `package main

import "fmt"

func Hello(name string) string {
	return fmt.Sprintf("hi %s", name)
}

type Greeter struct{ Prefix string }

func (g Greeter) Greet(n string) string { return g.Prefix + Hello(n) }
`

// TestSpike_OneGoFileToAST is the literal "get ONE Go file to an AST" gate.
func TestSpike_OneGoFileToAST(t *testing.T) {
	p, err := NewGoParser()
	if err != nil {
		t.Fatalf("NewGoParser: %v", err)
	}
	tree, err := p.Parse([]byte(sampleGo))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	lang := tree.Lang()
	root := tree.Root()
	if root.Type(lang) != "source_file" {
		t.Fatalf("root kind = %q, want source_file", root.Type(lang))
	}

	type hit struct {
		kind string
		line int
	}
	var hits []hit
	Walk(root, func(n Node) {
		switch n.Type(lang) {
		case "import_spec", "function_declaration", "type_spec", "method_declaration":
			hits = append(hits, hit{n.Type(lang), int(n.StartPoint().Row) + 1})
		}
	})

	want := []hit{
		{"import_spec", 3},
		{"function_declaration", 5},
		{"type_spec", 9},
		{"method_declaration", 11},
	}
	if len(hits) != len(want) {
		t.Fatalf("hits = %+v, want %+v", hits, want)
	}
	for i := range want {
		if hits[i] != want[i] {
			t.Fatalf("hit[%d] = %+v, want %+v", i, hits[i], want[i])
		}
	}
}

func TestSpike_NodeTextAndFields(t *testing.T) {
	p, err := NewGoParser()
	if err != nil {
		t.Fatalf("NewGoParser: %v", err)
	}
	tree, err := p.Parse([]byte(sampleGo))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	lang := tree.Lang()
	var funcName, importPath string
	Walk(tree.Root(), func(n Node) {
		switch n.Type(lang) {
		case "function_declaration":
			if c := n.ChildByField("name"); c != nil {
				funcName = c.Text()
			}
		case "import_spec":
			if c := n.ChildByField("path"); c != nil {
				importPath = c.Text()
			}
		}
	})
	if funcName != "Hello" {
		t.Fatalf("func name = %q, want Hello", funcName)
	}
	if importPath != `"fmt"` {
		t.Fatalf("import path = %q, want %q", importPath, `"fmt"`)
	}
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/parse/ -run TestSpike -v`
Expected: FAIL — `undefined: NewGoParser`, `undefined: Walk`, `undefined: Node`.

- [ ] **Step 4: Write the backend-agnostic interface**

Create `/Users/mylive/project/graffiti/graffiti/internal/parse/parser.go`:
```go
// Package parse turns source files into graffiti graph extractions (spec §5 parse).
// The tree-sitter backend is hidden behind the Parser/Tree/Node interfaces so it
// can be swapped (e.g. a per-language wazero+WASM fallback) without changing the
// rest of the pipeline.
package parse

import ts "github.com/odvcencio/gotreesitter"

// Node is a backend-agnostic AST node.
type Node interface {
	Type(lang *ts.Language) string  // tree-sitter node type, e.g. "function_declaration"
	StartPoint() ts.Point           // 0-based start position; 1-based line = int(StartPoint().Row)+1
	StartByte() int                 // byte offset into source
	EndByte() int                   // byte offset into source
	Text() string                   // source text spanned by the node
	ChildByField(field string) Node // named child for a grammar field (or nil)
	NamedChildren() []Node          // named children in source order
}

// Tree is a parsed syntax tree.
type Tree interface {
	Root() Node
	Lang() *ts.Language
}

// Parser parses source bytes into a Tree.
type Parser interface {
	Parse(src []byte) (Tree, error)
}

// Walk performs a deterministic pre-order traversal over named nodes, invoking
// fn for each node (including the root). Children are visited in source order,
// which is stable, preserving the §14 byte-determinism guarantee downstream.
func Walk(root Node, fn func(Node)) {
	if root == nil {
		return
	}
	fn(root)
	for _, c := range root.NamedChildren() {
		Walk(c, fn)
	}
}
```

- [ ] **Step 5: Write the gotreesitter backend**

Create `/Users/mylive/project/graffiti/graffiti/internal/parse/gotreesitter.go`:
```go
package parse

import (
	ts "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
)

// gtsParser is the gotreesitter-backed Parser. The library is pure-Go (no CGO,
// no WASM) and shipped with the grammar_subset+grammar_subset_go build tags.
type gtsParser struct {
	lang *ts.Language
}

// NewGoParser returns a Parser for the Go language.
func NewGoParser() (Parser, error) {
	return &gtsParser{lang: grammars.GoLanguage()}, nil
}

func (p *gtsParser) Parse(src []byte) (Tree, error) {
	parser := ts.NewParser(p.lang)
	tree, err := parser.Parse(src)
	if err != nil {
		return nil, err
	}
	return &gtsTree{tree: tree, lang: p.lang, src: src}, nil
}

type gtsTree struct {
	tree *ts.Tree
	lang *ts.Language
	src  []byte
}

func (t *gtsTree) Root() Node {
	return wrap(t.tree.RootNode(), t.lang, t.src)
}

func (t *gtsTree) Lang() *ts.Language { return t.lang }

type gtsNode struct {
	n    *ts.Node
	lang *ts.Language
	src  []byte
}

func wrap(n *ts.Node, lang *ts.Language, src []byte) Node {
	if n == nil {
		return nil
	}
	return &gtsNode{n: n, lang: lang, src: src}
}

func (g *gtsNode) Type(lang *ts.Language) string { return g.n.Type(lang) }
func (g *gtsNode) StartPoint() ts.Point          { return g.n.StartPoint() }
func (g *gtsNode) StartByte() int                { return int(g.n.StartByte()) }
func (g *gtsNode) EndByte() int                  { return int(g.n.EndByte()) }
func (g *gtsNode) Text() string                  { return g.n.Text(g.src) }

func (g *gtsNode) ChildByField(field string) Node {
	c := g.n.ChildByFieldName(field, g.lang)
	if c == nil {
		return nil
	}
	return wrap(c, g.lang, g.src)
}

func (g *gtsNode) NamedChildren() []Node {
	cnt := g.n.NamedChildCount()
	out := make([]Node, 0, cnt)
	for i := 0; i < cnt; i++ {
		out = append(out, wrap(g.n.NamedChild(i), g.lang, g.src))
	}
	return out
}
```

- [ ] **Step 6: Run the spike test to verify it passes**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/parse/ -run TestSpike -v`
Expected: PASS (`TestSpike_OneGoFileToAST`, `TestSpike_NodeTextAndFields`).

Note: the subset tags are required only for the binary-size goal; the code compiles and these tests pass even without them (linking the full grammar set). The Makefile (Task 15) is the single source of truth for the tags.

- [ ] **Step 7: Cross-compile gate (verifies the §10 single-static-binary constraint on this backend)**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/parse/
GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/parse/
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/parse/
```
Expected: all three exit 0 with no output (CGO disabled, all three targets).

- [ ] **Step 8: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/parse/parser.go internal/parse/gotreesitter.go internal/parse/parser_test.go go.mod go.sum
git commit -m "feat: parse spike - one Go file to AST behind swappable Parser interface

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 7: Parse-Go Pass 1 — definition nodes + structural edges + raw calls

**Files:**
- Create: `internal/parse/golang.go`
- Create: `internal/parse/golang_test.go`

Pass 1 (spec §5) walks one file's AST and emits: a `file` node; `function`/`method`/`type` (class) definition nodes; `imports` edges (EXTRACTED) from the file node to a synthesized module node per import (keyed by full import path); `contains` edges from the file node to each definition; and raw (unresolved) call sites stashed for Pass 2.

- [ ] **Step 1: Write the failing Pass-1 test**

Create `/Users/mylive/project/graffiti/graffiti/internal/parse/golang_test.go`:
```go
package parse

import (
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

const pass1Src = `package greet

import (
	"fmt"
	"strings"
)

func Hello(name string) string {
	return fmt.Sprintf("%s", strings.ToUpper(name))
}

type Greeter struct{ Prefix string }

func (g Greeter) Greet(n string) string {
	return g.Prefix + Hello(n)
}
`

func TestParseGo_Pass1_Nodes(t *testing.T) {
	p, err := NewGoParser()
	if err != nil {
		t.Fatal(err)
	}
	ex, err := ParseGo(p, "greet/greet.go", []byte(pass1Src))
	if err != nil {
		t.Fatalf("ParseGo: %v", err)
	}

	byID := map[string]graph.Node{}
	for _, n := range ex.Nodes {
		byID[n.ID] = n
	}

	fileID := graph.NodeID("greet/greet.go", "greet/greet.go")
	if fn, ok := byID[fileID]; !ok || fn.Kind != graph.KindFile {
		t.Fatalf("missing file node %q: %+v", fileID, byID)
	}

	helloID := graph.NodeID("greet/greet.go", "Hello")
	if n, ok := byID[helloID]; !ok || n.Kind != graph.KindFunction || n.Line != 8 || n.Label != "Hello" {
		t.Fatalf("Hello node wrong: ok=%v %+v", ok, byID[helloID])
	}

	greeterID := graph.NodeID("greet/greet.go", "Greeter")
	if n, ok := byID[greeterID]; !ok || n.Kind != graph.KindClass || n.Line != 12 {
		t.Fatalf("Greeter node wrong: ok=%v %+v", ok, byID[greeterID])
	}

	greetID := graph.NodeID("greet/greet.go", "Greeter.Greet")
	if n, ok := byID[greetID]; !ok || n.Kind != graph.KindMethod || n.Line != 14 || n.Label != "Greeter.Greet" {
		t.Fatalf("Greet method node wrong: ok=%v %+v", ok, byID[greetID])
	}
}

func TestParseGo_Pass1_ImportsAndContains(t *testing.T) {
	p, err := NewGoParser()
	if err != nil {
		t.Fatal(err)
	}
	ex, err := ParseGo(p, "greet/greet.go", []byte(pass1Src))
	if err != nil {
		t.Fatalf("ParseGo: %v", err)
	}

	fileID := graph.NodeID("greet/greet.go", "greet/greet.go")
	fmtModID := graph.NodeID("module:fmt", "fmt")

	var sawImportFmt, sawContainsHello bool
	for _, e := range ex.Edges {
		if e.From == fileID && e.To == fmtModID && e.Relation == graph.RelImports && e.Confidence == graph.ConfExtracted {
			sawImportFmt = true
		}
		if e.From == fileID && e.To == graph.NodeID("greet/greet.go", "Hello") && e.Relation == graph.RelContains {
			sawContainsHello = true
		}
	}
	if !sawImportFmt {
		t.Fatalf("missing imports edge file->fmt (EXTRACTED); edges=%+v", ex.Edges)
	}
	if !sawContainsHello {
		t.Fatalf("missing contains edge file->Hello")
	}

	var sawFmtMod bool
	for _, n := range ex.Nodes {
		if n.ID == fmtModID && n.Kind == graph.KindModule && n.Label == "fmt" {
			sawFmtMod = true
		}
	}
	if !sawFmtMod {
		t.Fatalf("missing module node for fmt")
	}
}

func TestParseGo_Pass1_RawCalls(t *testing.T) {
	p, err := NewGoParser()
	if err != nil {
		t.Fatal(err)
	}
	ex, err := ParseGo(p, "greet/greet.go", []byte(pass1Src))
	if err != nil {
		t.Fatalf("ParseGo: %v", err)
	}

	names := map[string]bool{}
	for _, rc := range ex.RawCalls {
		names[rc.Callee] = true
	}
	for _, want := range []string{"fmt.Sprintf", "strings.ToUpper", "Hello"} {
		if !names[want] {
			t.Fatalf("missing raw call %q; got %v", want, names)
		}
	}
	for _, rc := range ex.RawCalls {
		if rc.FromID == "" {
			t.Fatalf("raw call %q has empty FromID", rc.Callee)
		}
		if len(rc.Imports) == 0 {
			t.Fatalf("raw call %q has no Imports recorded", rc.Callee)
		}
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/parse/ -run TestParseGo_Pass1 -v`
Expected: FAIL — `undefined: ParseGo`, `undefined: Extraction`, `undefined: RawCall`.

- [ ] **Step 3: Write the Pass-1 implementation**

Create `/Users/mylive/project/graffiti/graffiti/internal/parse/golang.go`:
```go
package parse

import (
	"strings"

	ts "github.com/odvcencio/gotreesitter"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

// RawCall is an unresolved call site stashed by Pass 1 for Pass-2 resolution.
type RawCall struct {
	FromID  string   // node id of the enclosing definition (function/method)
	Callee  string   // call target text: bare "Hello" or selector "fmt.Sprintf"
	Line    int      // 1-based line of the call site
	File    string   // repo-relative file the call occurs in
	Imports []string // full import paths visible in the file (for Pass-2 decision)
}

// Extraction is the per-file output of Pass 1 (spec §5).
type Extraction struct {
	File     string
	Nodes    []graph.Node
	Edges    []graph.Edge
	RawCalls []RawCall
}

// ParseGo runs Pass 1 over one Go file: it emits a file node, definition nodes
// (function/method/type), imports edges (EXTRACTED) to synthesized module nodes
// keyed by full import path, contains edges, and stashes raw call sites.
func ParseGo(p Parser, relPath string, src []byte) (*Extraction, error) {
	tree, err := p.Parse(src)
	if err != nil {
		return nil, err
	}
	lang := tree.Lang()
	root := tree.Root()

	ex := &Extraction{File: relPath}

	fileID := graph.NodeID(relPath, relPath)
	ex.Nodes = append(ex.Nodes, graph.Node{
		ID: fileID, Label: relPath, Kind: graph.KindFile, File: relPath, Line: 1,
		Community: graph.UnclusteredCommunity,
	})

	// 1) Imports: walk import_spec nodes; emit module node (keyed by full path) + imports edge.
	var importPaths []string
	Walk(root, func(n Node) {
		if n.Type(lang) != "import_spec" {
			return
		}
		pathNode := n.ChildByField("path")
		if pathNode == nil {
			return
		}
		imp := unquote(pathNode.Text())
		if imp == "" {
			return
		}
		importPaths = append(importPaths, imp)
		modID := graph.NodeID("module:"+imp, importBase(imp))
		ex.Nodes = append(ex.Nodes, graph.Node{
			ID: modID, Label: importBase(imp), Kind: graph.KindModule, File: relPath,
			Line: int(n.StartPoint().Row) + 1, Community: graph.UnclusteredCommunity,
		})
		ex.Edges = append(ex.Edges, graph.Edge{
			From: fileID, To: modID, Relation: graph.RelImports, Confidence: graph.ConfExtracted,
		})
	})

	// 2) Definitions: function_declaration, method_declaration, type_spec.
	Walk(root, func(n Node) {
		switch n.Type(lang) {
		case "function_declaration":
			name := fieldText(n, "name")
			if name == "" {
				return
			}
			defID := graph.NodeID(relPath, name)
			emitDef(ex, fileID, defID, name, graph.KindFunction, relPath, int(n.StartPoint().Row)+1)
			collectCalls(ex, n, defID, relPath, importPaths, lang)
		case "method_declaration":
			name := fieldText(n, "name")
			if name == "" {
				return
			}
			recv := receiverTypeName(n, lang)
			label := name
			if recv != "" {
				label = recv + "." + name
			}
			defID := graph.NodeID(relPath, label)
			emitDef(ex, fileID, defID, label, graph.KindMethod, relPath, int(n.StartPoint().Row)+1)
			collectCalls(ex, n, defID, relPath, importPaths, lang)
		case "type_spec":
			name := fieldText(n, "name")
			if name == "" {
				return
			}
			defID := graph.NodeID(relPath, name)
			emitDef(ex, fileID, defID, name, graph.KindClass, relPath, int(n.StartPoint().Row)+1)
		}
	})

	return ex, nil
}

func emitDef(ex *Extraction, fileID, defID, label string, kind graph.Kind, file string, line int) {
	ex.Nodes = append(ex.Nodes, graph.Node{
		ID: defID, Label: label, Kind: kind, File: file, Line: line,
		Community: graph.UnclusteredCommunity,
	})
	ex.Edges = append(ex.Edges, graph.Edge{
		From: fileID, To: defID, Relation: graph.RelContains, Confidence: graph.ConfExtracted,
	})
}

// collectCalls walks a definition subtree and stashes every call_expression's
// callee as a RawCall attributed to defID.
func collectCalls(ex *Extraction, defNode Node, defID, file string, importPaths []string, lang *ts.Language) {
	Walk(defNode, func(n Node) {
		if n.Type(lang) != "call_expression" {
			return
		}
		fn := n.ChildByField("function")
		if fn == nil {
			return
		}
		callee := strings.TrimSpace(fn.Text())
		if callee == "" {
			return
		}
		ex.RawCalls = append(ex.RawCalls, RawCall{
			FromID: defID, Callee: callee, Line: int(n.StartPoint().Row) + 1, File: file,
			Imports: importPaths,
		})
	})
}

func fieldText(n Node, field string) string {
	c := n.ChildByField(field)
	if c == nil {
		return ""
	}
	return strings.TrimSpace(c.Text())
}

// receiverTypeName extracts the receiver's bare type name from a method_declaration.
// e.g. "(g Greeter)" -> "Greeter"; "(g *Greeter)" -> "Greeter".
func receiverTypeName(method Node, lang *ts.Language) string {
	recv := method.ChildByField("receiver")
	if recv == nil {
		return ""
	}
	var typeName string
	Walk(recv, func(n Node) {
		if typeName != "" {
			return
		}
		if n.Type(lang) == "type_identifier" {
			typeName = strings.TrimSpace(n.Text())
		}
	})
	return typeName
}

// unquote strips a single layer of surrounding double quotes from a Go string
// literal (the import_spec path text is quoted, e.g. `"fmt"`).
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// importBase returns the last path segment of an import path for the module label.
func importBase(imp string) string {
	if i := strings.LastIndex(imp, "/"); i >= 0 {
		return imp[i+1:]
	}
	return imp
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/parse/ -run TestParseGo_Pass1 -v`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/parse/golang.go internal/parse/golang_test.go
git commit -m "feat: parse-Go pass 1 - definitions, imports, contains, raw calls

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 8: Parse-Go Pass 2 — cross-file call resolution

**Files:**
- Create: `internal/parse/resolve.go`
- Create: `internal/parse/resolve_test.go`

Pass 2 (spec §5) builds a global `label → [node_id]` index over all definition nodes plus a `import-path → module-node-id` index, resolves each `RawCall`, promotes import-backed selector calls to EXTRACTED, marks bare resolved calls INFERRED, and **drops ambiguous common names** (a bare callee resolving to ≥2 files) to prevent god-node inflation.

> **DELIBERATE ASYMMETRY (documented, spec §5):** A selector call `pkg.Sym` whose package is imported resolves to the **module node** (coarse), even when `Sym` is itself defined in-repo, while a same-package bare call `Sym` resolves to the precise **definition node**. Precise cross-package symbol resolution (matching `pkg.Sym` to the def whose file lives under the imported package dir) is a later plan. This is intentional, not a bug, and is asserted by the golden in Task 14.
>
> **KNOWN LIMITATION (documented, deferred):** `moduleIdx` is keyed by full import path, and the selector package is matched to a specific import via the file's import set (last-segment match). If two imports in the same file share a last segment (e.g. `a/util` and `b/util`), the selector `util.Foo` is treated as **AMBIGUOUS** and dropped rather than guessed. The Plan-1 fixture has no such collision; broader alias-aware resolution is a later plan.

- [ ] **Step 1: Write the failing Pass-2 test**

Create `/Users/mylive/project/graffiti/graffiti/internal/parse/resolve_test.go`:
```go
package parse

import (
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

func mkNode(file, label string, kind graph.Kind) graph.Node {
	return graph.Node{ID: graph.NodeID(file, label), Label: label, Kind: kind, File: file, Line: 1, Community: -1}
}

func mkModule(importPath string) graph.Node {
	return graph.Node{ID: graph.NodeID("module:"+importPath, importBase(importPath)), Label: importBase(importPath), Kind: graph.KindModule, File: "x.go", Line: 1, Community: -1}
}

func TestResolveCalls_InferredBareCall(t *testing.T) {
	defs := []graph.Node{
		mkNode("greet/greet.go", "Hello", graph.KindFunction),
		mkNode("main.go", "main", graph.KindFunction),
	}
	calls := []RawCall{
		{FromID: graph.NodeID("main.go", "main"), Callee: "Hello", Line: 5, File: "main.go", Imports: nil},
	}
	edges := ResolveCalls(defs, calls)
	if len(edges) != 1 {
		t.Fatalf("edges = %d, want 1: %+v", len(edges), edges)
	}
	e := edges[0]
	if e.From != graph.NodeID("main.go", "main") || e.To != graph.NodeID("greet/greet.go", "Hello") {
		t.Fatalf("edge endpoints wrong: %+v", e)
	}
	if e.Relation != graph.RelCalls || e.Confidence != graph.ConfInferred {
		t.Fatalf("edge should be calls/INFERRED: %+v", e)
	}
}

func TestResolveCalls_ExtractedWhenImportBacked(t *testing.T) {
	defs := []graph.Node{
		mkNode("main.go", "main", graph.KindFunction),
		mkModule("fmt"),
	}
	calls := []RawCall{
		{FromID: graph.NodeID("main.go", "main"), Callee: "fmt.Sprintf", Line: 6, File: "main.go", Imports: []string{"fmt"}},
	}
	edges := ResolveCalls(defs, calls)
	if len(edges) != 1 {
		t.Fatalf("edges = %d, want 1: %+v", len(edges), edges)
	}
	e := edges[0]
	if e.To != graph.NodeID("module:fmt", "fmt") {
		t.Fatalf("selector call should target the imported module node, got %q", e.To)
	}
	if e.Confidence != graph.ConfExtracted {
		t.Fatalf("import-backed call must be EXTRACTED, got %q", e.Confidence)
	}
}

func TestResolveCalls_DropsAmbiguousCommonName(t *testing.T) {
	defs := []graph.Node{
		mkNode("a/a.go", "Run", graph.KindFunction),
		mkNode("b/b.go", "Run", graph.KindFunction),
		mkNode("main.go", "main", graph.KindFunction),
	}
	calls := []RawCall{
		{FromID: graph.NodeID("main.go", "main"), Callee: "Run", Line: 9, File: "main.go", Imports: nil},
	}
	edges := ResolveCalls(defs, calls)
	if len(edges) != 0 {
		t.Fatalf("ambiguous bare call must be dropped, got %+v", edges)
	}
}

func TestResolveCalls_UnresolvedSelectorDropped(t *testing.T) {
	defs := []graph.Node{mkNode("main.go", "main", graph.KindFunction)}
	calls := []RawCall{
		{FromID: graph.NodeID("main.go", "main"), Callee: "os.Exit", Line: 2, File: "main.go", Imports: nil},
	}
	edges := ResolveCalls(defs, calls)
	if len(edges) != 0 {
		t.Fatalf("unresolved selector must be dropped, got %+v", edges)
	}
}

func TestResolveCalls_Deterministic_NoDupEdges(t *testing.T) {
	defs := []graph.Node{
		mkNode("greet/greet.go", "Hello", graph.KindFunction),
		mkNode("main.go", "main", graph.KindFunction),
	}
	calls := []RawCall{
		{FromID: graph.NodeID("main.go", "main"), Callee: "Hello", Line: 5, File: "main.go"},
		{FromID: graph.NodeID("main.go", "main"), Callee: "Hello", Line: 7, File: "main.go"},
	}
	edges := ResolveCalls(defs, calls)
	if len(edges) != 1 {
		t.Fatalf("duplicate resolved calls must dedup to 1 edge, got %+v", edges)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/parse/ -run TestResolveCalls -v`
Expected: FAIL — `undefined: ResolveCalls`.

- [ ] **Step 3: Write the Pass-2 implementation**

Create `/Users/mylive/project/graffiti/graffiti/internal/parse/resolve.go`:
```go
package parse

import (
	"sort"
	"strings"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

// ResolveCalls performs Pass 2 (spec §5): resolve raw call sites against a global
// label index over all definition/module nodes and emit `calls` edges.
//
// Resolution rules:
//   - Selector callee "pkg.Sym": match "pkg" to the file's imports by last path
//     segment; if exactly one import matches AND a module node for that import
//     path exists, emit an EXTRACTED edge to that module node (coarse, by design).
//     Zero or multiple matching imports -> drop (unresolved/ambiguous).
//   - Bare callee "Sym": look up definitions labeled "Sym".
//       * exactly one defining file -> INFERRED edge to it.
//       * defined in >= 2 distinct files -> DROP (ambiguous common name).
//       * zero definitions -> drop (external/builtin).
//
// Output edges are deduped by (from,to,relation,confidence) and sorted for
// determinism (spec §14).
func ResolveCalls(defs []graph.Node, calls []RawCall) []graph.Edge {
	labelIdx := map[string][]graph.Node{}    // label -> definition nodes
	moduleByPath := map[string]string{}       // full import path -> module node id

	for _, n := range defs {
		switch n.Kind {
		case graph.KindFunction, graph.KindMethod, graph.KindClass:
			labelIdx[n.Label] = append(labelIdx[n.Label], n)
		case graph.KindModule:
			// Reconstruct the import path from the module node id (built as
			// NodeID("module:"+imp, base)); we instead index by recovering the
			// path from the node id's normalized form is lossy, so we index by a
			// per-call lookup below using the import set. To keep this exact, we
			// index module nodes by their *label* (last segment) AND remember the
			// id, then disambiguate per-call via the import set.
			moduleByPath[n.Label] = n.ID // keyed by last-segment label; see selector resolution
		}
	}

	seen := map[edgeDedupKey]bool{}
	var out []graph.Edge
	add := func(from, to string, conf graph.Confidence) {
		k := edgeDedupKey{from: from, to: to, conf: conf}
		if seen[k] {
			return
		}
		seen[k] = true
		out = append(out, graph.Edge{From: from, To: to, Relation: graph.RelCalls, Confidence: conf})
	}

	for _, rc := range calls {
		if pkg, _, isSel := splitSelector(rc.Callee); isSel {
			// Match pkg to imports by last segment; require a UNIQUE match.
			matches := matchingImports(rc.Imports, pkg)
			if len(matches) != 1 {
				continue // unresolved or ambiguous selector -> drop
			}
			// module node id for that import path == NodeID("module:"+imp, base)
			modID := graph.NodeID("module:"+matches[0], importBase(matches[0]))
			// confirm the module node actually exists (by label index fallback)
			if !moduleNodeExists(defs, modID) {
				continue
			}
			add(rc.FromID, modID, graph.ConfExtracted)
			continue
		}

		// bare call
		cands := labelIdx[rc.Callee]
		if len(cands) == 0 {
			continue
		}
		if distinctFiles(cands) >= 2 {
			continue // ambiguous common name -> drop
		}
		target := pickDeterministic(cands)
		add(rc.FromID, target.ID, graph.ConfInferred)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		if out[i].To != out[j].To {
			return out[i].To < out[j].To
		}
		if out[i].Relation != out[j].Relation {
			return out[i].Relation < out[j].Relation
		}
		return out[i].Confidence < out[j].Confidence
	})
	return out
}

type edgeDedupKey struct {
	from, to string
	conf     graph.Confidence
}

// splitSelector splits "pkg.Sym" into ("pkg","Sym",true). For a bare "Sym" returns
// isSel=false. For "a.b.C" uses the FIRST segment as pkg.
func splitSelector(callee string) (pkg, sym string, isSel bool) {
	i := strings.Index(callee, ".")
	if i < 0 {
		return "", callee, false
	}
	return callee[:i], callee[i+1:], true
}

// matchingImports returns import paths whose last segment equals pkg.
func matchingImports(imports []string, pkg string) []string {
	var out []string
	for _, imp := range imports {
		if importBase(imp) == pkg {
			out = append(out, imp)
		}
	}
	return out
}

func moduleNodeExists(defs []graph.Node, id string) bool {
	for _, n := range defs {
		if n.Kind == graph.KindModule && n.ID == id {
			return true
		}
	}
	return false
}

func distinctFiles(ns []graph.Node) int {
	files := map[string]bool{}
	for _, n := range ns {
		files[n.File] = true
	}
	return len(files)
}

// pickDeterministic chooses a stable target among candidates in a single file
// (lowest ID).
func pickDeterministic(ns []graph.Node) graph.Node {
	best := ns[0]
	for _, n := range ns[1:] {
		if n.ID < best.ID {
			best = n
		}
	}
	return best
}
```

> Note: `moduleByPath` is constructed for clarity but `ResolveCalls` resolves selectors deterministically by recomputing the module node id from the unique matching import path and confirming the node exists via `moduleNodeExists`. The `moduleByPath` map is intentionally minimal; keep it as written (it compiles and is unused by the resolution path, serving only as documentation of the label keying). If `go vet`/compiler flags `moduleByPath` as unused, delete its declaration and the loop branch that fills it — they are not load-bearing.

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/parse/ -run TestResolveCalls -v`
Expected: PASS (5 tests). If the compiler reports `moduleByPath declared and not used`, remove the `moduleByPath` declaration and its `case graph.KindModule:` assignment (they are non-load-bearing), then re-run.

- [ ] **Step 5: Run the whole parse package to confirm no regressions**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/parse/ -v`
Expected: PASS (all spike + Pass-1 + Pass-2 tests).

- [ ] **Step 6: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/parse/resolve.go internal/parse/resolve_test.go
git commit -m "feat: parse-Go pass 2 - cross-file call resolution with ambiguity drop

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 9: Structural schema validation (`internal/schemaval`)

**Files:**
- Create: `internal/schemaval/schemaval.go`
- Create: `internal/schemaval/schemaval_test.go`

A small in-repo structural validator (no third-party JSON-schema dependency) enforcing the same rules as `schema/map.schema.json`: required fields (including `generated_at`), closed enums, dangling-edge check, unique node IDs.

- [ ] **Step 1: Write the failing validator test**

Create `/Users/mylive/project/graffiti/graffiti/internal/schemaval/schemaval_test.go`:
```go
package schemaval

import (
	"strings"
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

func validDoc() *graph.Document {
	d := graph.NewDocument("repo")
	d.GeneratedAt = "2026-06-14T00:00:00Z"
	d.Nodes = []graph.Node{
		{ID: "f", Label: "f.go", Kind: graph.KindFile, File: "f.go", Line: 1, Community: -1},
		{ID: "f:hello", Label: "Hello", Kind: graph.KindFunction, File: "f.go", Line: 2, Community: -1},
	}
	d.Edges = []graph.Edge{
		{From: "f", To: "f:hello", Relation: graph.RelContains, Confidence: graph.ConfExtracted},
	}
	return d
}

func TestValidateDocument_OK(t *testing.T) {
	if err := ValidateDocument(validDoc()); err != nil {
		t.Fatalf("valid doc rejected: %v", err)
	}
}

func TestValidateDocument_BadKind(t *testing.T) {
	d := validDoc()
	d.Nodes[1].Kind = "widget"
	err := ValidateDocument(d)
	if err == nil || !strings.Contains(err.Error(), "kind") {
		t.Fatalf("expected kind error, got %v", err)
	}
}

func TestValidateDocument_DanglingEdge(t *testing.T) {
	d := validDoc()
	d.Edges = append(d.Edges, graph.Edge{From: "f", To: "ghost", Relation: graph.RelCalls, Confidence: graph.ConfInferred})
	err := ValidateDocument(d)
	if err == nil || !strings.Contains(err.Error(), "ghost") {
		t.Fatalf("expected dangling-edge error mentioning ghost, got %v", err)
	}
}

func TestValidateDocument_DuplicateNodeID(t *testing.T) {
	d := validDoc()
	d.Nodes = append(d.Nodes, graph.Node{ID: "f", Label: "dup", Kind: graph.KindFile, File: "f.go", Line: 1, Community: -1})
	err := ValidateDocument(d)
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate node id error, got %v", err)
	}
}

func TestValidateDocument_BadConfidence(t *testing.T) {
	d := validDoc()
	d.Edges[0].Confidence = "MAYBE"
	err := ValidateDocument(d)
	if err == nil || !strings.Contains(err.Error(), "confidence") {
		t.Fatalf("expected confidence error, got %v", err)
	}
}

func TestValidateDocument_MissingVersion(t *testing.T) {
	d := validDoc()
	d.Version = ""
	err := ValidateDocument(d)
	if err == nil || !strings.Contains(err.Error(), "version") {
		t.Fatalf("expected version error, got %v", err)
	}
}

func TestValidateDocument_MissingGeneratedAt(t *testing.T) {
	d := validDoc()
	d.GeneratedAt = ""
	err := ValidateDocument(d)
	if err == nil || !strings.Contains(err.Error(), "generated_at") {
		t.Fatalf("expected generated_at error, got %v", err)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/schemaval/ -run TestValidateDocument -v`
Expected: FAIL — `undefined: ValidateDocument`.

- [ ] **Step 3: Write the validator**

Create `/Users/mylive/project/graffiti/graffiti/internal/schemaval/schemaval.go`:
```go
// Package schemaval structurally validates a graph.Document against the rules
// published in schema/map.schema.json, without a third-party JSON-schema engine.
package schemaval

import (
	"fmt"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

// ValidateDocument checks required fields, closed enums, unique node IDs, and
// that every edge endpoint references an existing node. It returns the first
// error found (deterministic order: document fields, then nodes, then edges).
func ValidateDocument(d *graph.Document) error {
	if d == nil {
		return fmt.Errorf("nil document")
	}
	if d.Version != graph.SchemaVersion {
		return fmt.Errorf("version: got %q want %q", d.Version, graph.SchemaVersion)
	}
	if d.GeneratedAt == "" {
		return fmt.Errorf("generated_at: must be set")
	}
	if d.Root == "" {
		return fmt.Errorf("root: must be set")
	}

	ids := make(map[string]bool, len(d.Nodes))
	for i, n := range d.Nodes {
		if n.ID == "" {
			return fmt.Errorf("nodes[%d]: empty id", i)
		}
		if ids[n.ID] {
			return fmt.Errorf("nodes[%d]: duplicate id %q", i, n.ID)
		}
		ids[n.ID] = true
		if !graph.ValidKinds[n.Kind] {
			return fmt.Errorf("nodes[%d] (%q): invalid kind %q", i, n.ID, n.Kind)
		}
		if n.Line < 0 {
			return fmt.Errorf("nodes[%d] (%q): negative line %d", i, n.ID, n.Line)
		}
		if n.Community < -1 {
			return fmt.Errorf("nodes[%d] (%q): community < -1: %d", i, n.ID, n.Community)
		}
	}

	for i, e := range d.Edges {
		if e.From == "" || e.To == "" {
			return fmt.Errorf("edges[%d]: empty endpoint", i)
		}
		if !graph.ValidRelations[e.Relation] {
			return fmt.Errorf("edges[%d]: invalid relation %q", i, e.Relation)
		}
		if !graph.ValidConfidences[e.Confidence] {
			return fmt.Errorf("edges[%d]: invalid confidence %q", i, e.Confidence)
		}
		if !ids[e.From] {
			return fmt.Errorf("edges[%d]: dangling 'from' node %q", i, e.From)
		}
		if !ids[e.To] {
			return fmt.Errorf("edges[%d]: dangling 'to' node %q", i, e.To)
		}
	}
	return nil
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/schemaval/ -run TestValidateDocument -v`
Expected: PASS (7 tests).

- [ ] **Step 5: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/schemaval/schemaval.go internal/schemaval/schemaval_test.go
git commit -m "feat: structural schema validation for graph documents

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 10: Assemble the graph (`internal/build`)

**Files:**
- Create: `internal/build/build.go`
- Create: `internal/build/build_test.go`

`Assemble(root, generatedAt, exs)` stamps `generated_at` (so the validator's required-field check passes — this is BLOCKER FIX #1), dedups nodes (first writer wins) and edges, runs Pass-2 resolution to produce `calls` edges, validates the result against the schema, and sorts nodes/edges deterministically.

- [ ] **Step 1: Write the failing build test**

Create `/Users/mylive/project/graffiti/graffiti/internal/build/build_test.go`:
```go
package build

import (
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/graph"
	"github.com/evgeniy-achin/graffiti/internal/parse"
)

const genAt = "2026-06-14T00:00:00Z"

func TestAssemble_DedupNodesAndSorts(t *testing.T) {
	exMain := &parse.Extraction{
		File: "main.go",
		Nodes: []graph.Node{
			{ID: graph.NodeID("main.go", "main.go"), Label: "main.go", Kind: graph.KindFile, File: "main.go", Line: 1, Community: -1},
			{ID: graph.NodeID("main.go", "main"), Label: "main", Kind: graph.KindFunction, File: "main.go", Line: 5, Community: -1},
			{ID: graph.NodeID("module:example.com/greet", "greet"), Label: "greet", Kind: graph.KindModule, File: "main.go", Line: 3, Community: -1},
		},
		Edges: []graph.Edge{
			{From: graph.NodeID("main.go", "main.go"), To: graph.NodeID("module:example.com/greet", "greet"), Relation: graph.RelImports, Confidence: graph.ConfExtracted},
			{From: graph.NodeID("main.go", "main.go"), To: graph.NodeID("main.go", "main"), Relation: graph.RelContains, Confidence: graph.ConfExtracted},
		},
		RawCalls: []parse.RawCall{
			{FromID: graph.NodeID("main.go", "main"), Callee: "Hello", Line: 6, File: "main.go", Imports: []string{"example.com/greet"}},
		},
	}
	exGreet := &parse.Extraction{
		File: "greet/greet.go",
		Nodes: []graph.Node{
			{ID: graph.NodeID("greet/greet.go", "greet/greet.go"), Label: "greet/greet.go", Kind: graph.KindFile, File: "greet/greet.go", Line: 1, Community: -1},
			{ID: graph.NodeID("greet/greet.go", "Hello"), Label: "Hello", Kind: graph.KindFunction, File: "greet/greet.go", Line: 3, Community: -1},
		},
		Edges: []graph.Edge{
			{From: graph.NodeID("greet/greet.go", "greet/greet.go"), To: graph.NodeID("greet/greet.go", "Hello"), Relation: graph.RelContains, Confidence: graph.ConfExtracted},
		},
	}

	doc, err := Assemble("example-repo", genAt, []*parse.Extraction{exMain, exGreet})
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	if doc.GeneratedAt != genAt {
		t.Fatalf("generated_at = %q, want %q", doc.GeneratedAt, genAt)
	}

	for i := 1; i < len(doc.Nodes); i++ {
		if doc.Nodes[i-1].ID > doc.Nodes[i].ID {
			t.Fatalf("nodes not sorted at %d: %q > %q", i, doc.Nodes[i-1].ID, doc.Nodes[i].ID)
		}
	}

	// bare call Hello (defined once) resolves to an INFERRED calls edge
	var sawCall bool
	for _, e := range doc.Edges {
		if e.Relation == graph.RelCalls &&
			e.From == graph.NodeID("main.go", "main") &&
			e.To == graph.NodeID("greet/greet.go", "Hello") &&
			e.Confidence == graph.ConfInferred {
			sawCall = true
		}
	}
	if !sawCall {
		t.Fatalf("expected INFERRED calls edge main->Hello; edges=%+v", doc.Edges)
	}

	for _, n := range doc.Nodes {
		if n.Community != -1 {
			t.Fatalf("node %q community = %d, want -1 (pre-cluster)", n.ID, n.Community)
		}
	}
}

func TestAssemble_ValidatesAndRejectsBadGraph(t *testing.T) {
	ex := &parse.Extraction{
		File: "x.go",
		Nodes: []graph.Node{
			{ID: graph.NodeID("x.go", "x.go"), Label: "x.go", Kind: graph.KindFile, File: "x.go", Line: 1, Community: -1},
		},
		Edges: []graph.Edge{
			{From: graph.NodeID("x.go", "x.go"), To: "ghost-node", Relation: graph.RelContains, Confidence: graph.ConfExtracted},
		},
	}
	_, err := Assemble("repo", genAt, []*parse.Extraction{ex})
	if err == nil {
		t.Fatalf("expected validation failure for dangling edge")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/build/ -run TestAssemble -v`
Expected: FAIL — `undefined: Assemble`.

- [ ] **Step 3: Write the build implementation**

Create `/Users/mylive/project/graffiti/graffiti/internal/build/build.go`:
```go
// Package build assembles per-file extractions into a single validated, directed,
// deterministically ordered graph.Document (spec §5 build stage).
package build

import (
	"sort"

	"github.com/evgeniy-achin/graffiti/internal/graph"
	"github.com/evgeniy-achin/graffiti/internal/parse"
	"github.com/evgeniy-achin/graffiti/internal/schemaval"
)

// Assemble folds extractions into one Document: stamp generatedAt, dedup nodes
// (first writer wins), dedup structural edges, run Pass-2 call resolution across
// all files, sort nodes/edges deterministically, then validate against the schema.
// Community is left at -1 (clustering is a later plan).
//
// generatedAt is threaded in here (not only at write time) so the schema's
// required generated_at field is satisfied before validation.
func Assemble(root, generatedAt string, exs []*parse.Extraction) (*graph.Document, error) {
	doc := graph.NewDocument(root)
	doc.GeneratedAt = generatedAt

	nodeIdx := map[string]bool{}
	allDefs := []graph.Node{}
	for _, ex := range exs {
		for _, n := range ex.Nodes {
			if !nodeIdx[n.ID] {
				nodeIdx[n.ID] = true
				doc.Nodes = append(doc.Nodes, n)
			}
			allDefs = append(allDefs, n)
		}
	}

	edgeIdx := map[edgeKey]bool{}
	addEdge := func(e graph.Edge) {
		k := edgeKey{e.From, e.To, e.Relation, e.Confidence}
		if edgeIdx[k] {
			return
		}
		edgeIdx[k] = true
		doc.Edges = append(doc.Edges, e)
	}

	for _, ex := range exs {
		for _, e := range ex.Edges {
			addEdge(e)
		}
	}

	var allCalls []parse.RawCall
	for _, ex := range exs {
		allCalls = append(allCalls, ex.RawCalls...)
	}
	for _, e := range parse.ResolveCalls(allDefs, allCalls) {
		addEdge(e)
	}

	sortDocument(doc)

	if err := schemaval.ValidateDocument(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

type edgeKey struct {
	from, to string
	rel      graph.Relation
	conf     graph.Confidence
}

// sortDocument imposes the canonical deterministic order (spec §8.8/§14):
// nodes by ID; edges by (from, to, relation, confidence).
func sortDocument(doc *graph.Document) {
	sort.Slice(doc.Nodes, func(i, j int) bool { return doc.Nodes[i].ID < doc.Nodes[j].ID })
	sort.Slice(doc.Edges, func(i, j int) bool {
		a, b := doc.Edges[i], doc.Edges[j]
		if a.From != b.From {
			return a.From < b.From
		}
		if a.To != b.To {
			return a.To < b.To
		}
		if a.Relation != b.Relation {
			return a.Relation < b.Relation
		}
		return a.Confidence < b.Confidence
	})
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/build/ -run TestAssemble -v`
Expected: PASS (2 tests).

- [ ] **Step 5: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/build/build.go internal/build/build_test.go
git commit -m "feat: assemble validated deterministic directed graph with generated_at

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 11: Content-hash cache (`internal/cache`)

**Files:**
- Create: `internal/cache/cache.go`
- Create: `internal/cache/cache_test.go`

SHA256 content hash per file written under `.graffiti/cache/` (spec §6 incremental groundwork). **Plan 1 writes hashes for FORWARD-COMPATIBILITY only and performs no skip** — the incremental `graffiti update` skip logic is a later plan. The cache files must still be written deterministically now.

- [ ] **Step 1: Write the failing cache test**

Create `/Users/mylive/project/graffiti/graffiti/internal/cache/cache_test.go`:
```go
package cache

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHashBytes_Stable(t *testing.T) {
	a := HashBytes([]byte("hello"))
	b := HashBytes([]byte("hello"))
	if a != b {
		t.Fatalf("hash not stable: %q vs %q", a, b)
	}
	const want = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if a != want {
		t.Fatalf("hash = %q, want %q", a, want)
	}
}

func TestHashFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x.go")
	if err := os.WriteFile(p, []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	h, err := HashFile(p)
	if err != nil {
		t.Fatalf("HashFile: %v", err)
	}
	if h != HashBytes([]byte("package main")) {
		t.Fatalf("HashFile mismatch")
	}
}

func TestCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)
	if err := c.Put("main.go", "abc123"); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if err := c.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".graffiti", "cache", "hashes.json")); err != nil {
		t.Fatalf("cache file missing: %v", err)
	}
	c2 := New(dir)
	if err := c2.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got, ok := c2.Get("main.go"); !ok || got != "abc123" {
		t.Fatalf("reload mismatch: ok=%v got=%q", ok, got)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/cache/ -run 'TestHash|TestCache' -v`
Expected: FAIL — `undefined: HashBytes`, `undefined: New`.

- [ ] **Step 3: Write the cache implementation**

Create `/Users/mylive/project/graffiti/graffiti/internal/cache/cache.go`:
```go
// Package cache stores per-file SHA256 content hashes under .graffiti/cache/
// to support incremental rebuilds (spec §6). Plan 1 only writes/reads hashes;
// it performs no skip-on-match (that is a later plan).
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// HashBytes returns the lowercase hex SHA256 of b.
func HashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// HashFile returns the SHA256 of the file at path.
func HashFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return HashBytes(b), nil
}

// Cache holds repo-relative path -> content hash, persisted to
// <root>/.graffiti/cache/hashes.json.
type Cache struct {
	root    string
	entries map[string]string
}

// New returns an empty cache rooted at repo root.
func New(root string) *Cache {
	return &Cache{root: root, entries: map[string]string{}}
}

func (c *Cache) path() string {
	return filepath.Join(c.root, ".graffiti", "cache", "hashes.json")
}

// Load reads existing cache entries; a missing file is not an error.
func (c *Cache) Load() error {
	b, err := os.ReadFile(c.path())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(b, &c.entries)
}

// Get returns the stored hash for relPath.
func (c *Cache) Get(relPath string) (string, bool) {
	h, ok := c.entries[relPath]
	return h, ok
}

// Put records a hash for relPath.
func (c *Cache) Put(relPath, hash string) error {
	c.entries[relPath] = hash
	return nil
}

// Flush writes the cache deterministically (sorted keys) to disk.
func (c *Cache) Flush() error {
	if err := os.MkdirAll(filepath.Dir(c.path()), 0o755); err != nil {
		return err
	}
	keys := make([]string, 0, len(c.entries))
	for k := range c.entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	ordered := make([][2]string, 0, len(keys))
	for _, k := range keys {
		ordered = append(ordered, [2]string{k, c.entries[k]})
	}
	b, err := marshalSorted(ordered)
	if err != nil {
		return err
	}
	return os.WriteFile(c.path(), b, 0o644)
}

// marshalSorted encodes [(key,value)...] as a JSON object in the given order.
func marshalSorted(pairs [][2]string) ([]byte, error) {
	out := []byte("{")
	for i, p := range pairs {
		if i > 0 {
			out = append(out, ',')
		}
		kb, err := json.Marshal(p[0])
		if err != nil {
			return nil, err
		}
		vb, err := json.Marshal(p[1])
		if err != nil {
			return nil, err
		}
		out = append(out, kb...)
		out = append(out, ':')
		out = append(out, vb...)
	}
	out = append(out, '}', '\n')
	return out, nil
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/cache/ -run 'TestHash|TestCache' -v`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/cache/cache.go internal/cache/cache_test.go
git commit -m "feat: per-file SHA256 content-hash cache under .graffiti/cache

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 12: Deterministic map.json writer (`internal/render`)

**Files:**
- Create: `internal/render/json.go`
- Create: `internal/render/json_test.go`

Writes `.graffiti/map.json` with sorted keys, a trailing newline, and the `generated_at` already stamped on the document (read from `doc.GeneratedAt` — single source of truth, BLOCKER FIX #1). Determinism (modulo `generated_at` and `root`) is the headline guarantee (§8.8/§14).

- [ ] **Step 1: Write the failing writer test**

Create `/Users/mylive/project/graffiti/graffiti/internal/render/json_test.go`:
```go
package render

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

func sampleDoc(at string) *graph.Document {
	d := graph.NewDocument("repo")
	d.GeneratedAt = at
	d.Nodes = []graph.Node{
		{ID: "a", Label: "A", Kind: graph.KindFile, File: "a.go", Line: 1, Community: -1},
		{ID: "a:hello", Label: "Hello", Kind: graph.KindFunction, File: "a.go", Line: 2, Community: -1},
	}
	d.Edges = []graph.Edge{
		{From: "a", To: "a:hello", Relation: graph.RelContains, Confidence: graph.ConfExtracted},
	}
	return d
}

func TestWriteMapJSON_KeepsGeneratedAtAndIsValidJSON(t *testing.T) {
	dir := t.TempDir()
	at := "2026-06-14T12:00:00Z"
	if err := WriteMapJSON(sampleDoc(at), dir); err != nil {
		t.Fatalf("write: %v", err)
	}
	p := filepath.Join(dir, ".graffiti", "map.json")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var back graph.Document
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if back.GeneratedAt != at {
		t.Fatalf("generated_at = %q, want %q", back.GeneratedAt, at)
	}
	if !strings.HasSuffix(string(b), "}\n") {
		t.Fatalf("output must end with newline")
	}
}

func TestWriteMapJSON_DeterministicModuloGeneratedAt(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	if err := WriteMapJSON(sampleDoc("2026-06-14T00:00:00Z"), dir1); err != nil {
		t.Fatal(err)
	}
	if err := WriteMapJSON(sampleDoc("2099-01-01T00:00:00Z"), dir2); err != nil {
		t.Fatal(err)
	}
	b1, _ := os.ReadFile(filepath.Join(dir1, ".graffiti", "map.json"))
	b2, _ := os.ReadFile(filepath.Join(dir2, ".graffiti", "map.json"))

	reAt := regexp.MustCompile(`"generated_at":\s*"[^"]*"`)
	n1 := reAt.ReplaceAll(b1, []byte(`"generated_at":"X"`))
	n2 := reAt.ReplaceAll(b2, []byte(`"generated_at":"X"`))
	if string(n1) != string(n2) {
		t.Fatalf("output not byte-identical modulo generated_at:\n%s\n---\n%s", n1, n2)
	}
}

func TestWriteMapJSON_SortedTopLevelKeys(t *testing.T) {
	dir := t.TempDir()
	if err := WriteMapJSON(sampleDoc("2026-06-14T00:00:00Z"), dir); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".graffiti", "map.json"))
	s := string(b)
	order := []string{`"communities"`, `"edges"`, `"generated_at"`, `"nodes"`, `"root"`, `"version"`}
	last := -1
	for _, k := range order {
		idx := strings.Index(s, k)
		if idx < 0 {
			t.Fatalf("missing key %s", k)
		}
		if idx < last {
			t.Fatalf("key %s out of sorted order", k)
		}
		last = idx
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/render/ -run TestWriteMapJSON -v`
Expected: FAIL — `undefined: WriteMapJSON`.

- [ ] **Step 3: Write the writer**

Create `/Users/mylive/project/graffiti/graffiti/internal/render/json.go`:
```go
// Package render serializes a graph.Document to disk artifacts. Plan 1 emits
// only map.json (MAP.md and map.html are later plans).
package render

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

// orderedDocument mirrors graph.Document but with struct fields ordered so the
// emitted JSON object keys are alphabetical (communities, edges, generated_at,
// nodes, root, version) for byte-determinism (spec §8.8/§14).
type orderedDocument struct {
	Communities []graph.Community `json:"communities"`
	Edges       []graph.Edge      `json:"edges"`
	GeneratedAt string            `json:"generated_at"`
	Nodes       []graph.Node      `json:"nodes"`
	Root        string            `json:"root"`
	Version     string            `json:"version"`
}

// WriteMapJSON writes doc to <root>/.graffiti/map.json. The generated_at value is
// read directly off doc.GeneratedAt (stamped by build.Assemble) — single source
// of truth. Output is deterministic modulo generated_at and root: top-level keys
// are alphabetical and the node/edge arrays are assumed already sorted by Assemble.
func WriteMapJSON(doc *graph.Document, root string) error {
	od := orderedDocument{
		Communities: nonNilCommunities(doc.Communities),
		Edges:       nonNilEdges(doc.Edges),
		GeneratedAt: doc.GeneratedAt,
		Nodes:       nonNilNodes(doc.Nodes),
		Root:        doc.Root,
		Version:     doc.Version,
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(od); err != nil { // Encode appends a trailing '\n'
		return err
	}

	dir := filepath.Join(root, ".graffiti")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "map.json"), buf.Bytes(), 0o644)
}

func nonNilNodes(n []graph.Node) []graph.Node {
	if n == nil {
		return []graph.Node{}
	}
	return n
}
func nonNilEdges(e []graph.Edge) []graph.Edge {
	if e == nil {
		return []graph.Edge{}
	}
	return e
}
func nonNilCommunities(c []graph.Community) []graph.Community {
	if c == nil {
		return []graph.Community{}
	}
	return c
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/render/ -run TestWriteMapJSON -v`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/render/json.go internal/render/json_test.go
git commit -m "feat: deterministic map.json writer reading generated_at off the document

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 13: Wire the pipeline (`internal/app`) + CLI success line

**Files:**
- Create: `internal/app/app.go`
- Create: `internal/app/app_test.go`
- Modify: `cmd/graffiti/main.go` (replace the `runBuild` stub)
- Modify: `cmd/graffiti/main_test.go` (add the E2E success-line test)

`app.Build` orchestrates scan → parse (per Go file) → build → write, also emitting `doc` nodes for Markdown files (discovered but not parsed in Plan 1), and writes the content-hash cache. **BLOCKER FIX #2:** the document `root` is set to a stable `filepath.Base(absRoot)` (not the absolute scan path) so two builds of the same repo into different directories are byte-identical. It returns `Stats` for the success line.

- [ ] **Step 1: Write the failing app test**

Create `/Users/mylive/project/graffiti/graffiti/internal/app/app_test.go`:
```go
package app

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func write(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestBuild_ProducesMapAndStats(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "go.mod", "module example.com/demo\n\ngo 1.26\n")
	write(t, dir, "main.go", "package main\n\nfunc main() { Hello() }\n\nfunc Hello() {}\n")
	write(t, dir, "README.md", "# demo\n")

	stats, err := Build(dir, "2026-06-14T00:00:00Z")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if stats.Files < 2 {
		t.Fatalf("expected >=2 files scanned, got %d", stats.Files)
	}
	if stats.Nodes == 0 || stats.Edges == 0 {
		t.Fatalf("expected nonzero nodes/edges, got %d/%d", stats.Nodes, stats.Edges)
	}
	if _, err := os.Stat(filepath.Join(dir, ".graffiti", "map.json")); err != nil {
		t.Fatalf("map.json missing: %v", err)
	}
	// cache written with an entry per scanned file (forward-compat artifact)
	if _, err := os.Stat(filepath.Join(dir, ".graffiti", "cache", "hashes.json")); err != nil {
		t.Fatalf("cache hashes.json missing: %v", err)
	}
	if !stats.HasDocNode {
		t.Fatalf("expected a doc node for README.md")
	}
}

var reGenAtApp = regexp.MustCompile(`("generated_at":\s*")[^"]*(")`)
var reRootApp = regexp.MustCompile(`("root":\s*")[^"]*(")`)

func normApp(b []byte) string {
	b = reGenAtApp.ReplaceAll(b, []byte(`${1}X${2}`))
	b = reRootApp.ReplaceAll(b, []byte(`${1}X${2}`))
	return string(b)
}

func TestBuild_DeterministicModuloGeneratedAtAndRoot(t *testing.T) {
	src := "package main\n\nfunc main() { Hello() }\n\nfunc Hello() {}\n"

	dir1 := t.TempDir()
	write(t, dir1, "main.go", src)
	if _, err := Build(dir1, "2026-06-14T00:00:00Z"); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(filepath.Join(dir1, ".graffiti", "map.json"))

	dir2 := t.TempDir()
	write(t, dir2, "main.go", src)
	if _, err := Build(dir2, "2099-12-31T23:59:59Z"); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(filepath.Join(dir2, ".graffiti", "map.json"))

	if normApp(first) != normApp(second) {
		t.Fatalf("not deterministic modulo generated_at+root:\n%s\n---\n%s", normApp(first), normApp(second))
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/app/ -run TestBuild -v`
Expected: FAIL — `undefined: Build`, `undefined: Stats`.

- [ ] **Step 3: Write the app orchestration**

Create `/Users/mylive/project/graffiti/graffiti/internal/app/app.go`:
```go
// Package app wires the graffiti pipeline (scan → parse → build → render) for a
// single Go repository (Plan 1 scope).
package app

import (
	"os"
	"path/filepath"

	"github.com/evgeniy-achin/graffiti/internal/build"
	"github.com/evgeniy-achin/graffiti/internal/cache"
	"github.com/evgeniy-achin/graffiti/internal/graph"
	"github.com/evgeniy-achin/graffiti/internal/parse"
	"github.com/evgeniy-achin/graffiti/internal/render"
	"github.com/evgeniy-achin/graffiti/internal/scan"
)

// Stats summarizes a build for the CLI success line.
type Stats struct {
	Files       int
	Nodes       int
	Edges       int
	Communities int
	HasDocNode  bool // whether a markdown doc node was emitted
}

// Build runs the full pipeline against root, stamping generatedAt into the
// document (via build.Assemble), and returns Stats. generatedAt should be RFC3339.
//
// The document root is set to filepath.Base(absRoot) so map.json is byte-identical
// for the same repo regardless of the absolute build directory (determinism, §14).
func Build(root, generatedAt string) (Stats, error) {
	var stats Stats

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return stats, err
	}
	docRoot := filepath.Base(absRoot) // stable, location-independent

	refs, err := scan.Scan(absRoot)
	if err != nil {
		return stats, err
	}
	stats.Files = len(refs)

	goParser, err := parse.NewGoParser()
	if err != nil {
		return stats, err
	}

	c := cache.New(absRoot)
	_ = c.Load() // loaded for forward-compat; Plan 1 does not skip on hash match

	var extractions []*parse.Extraction
	for _, ref := range refs {
		src, readErr := os.ReadFile(ref.AbsPath)
		if readErr != nil {
			return stats, readErr
		}
		_ = c.Put(ref.RelPath, cache.HashBytes(src))

		switch ref.Lang {
		case scan.LangGo:
			ex, perr := parse.ParseGo(goParser, ref.RelPath, src)
			if perr != nil {
				return stats, perr
			}
			extractions = append(extractions, ex)
		case scan.LangMarkdown:
			extractions = append(extractions, markdownExtraction(ref.RelPath))
			stats.HasDocNode = true
		}
	}

	doc, err := build.Assemble(docRoot, generatedAt, extractions)
	if err != nil {
		return stats, err
	}

	if err := render.WriteMapJSON(doc, absRoot); err != nil {
		return stats, err
	}
	if err := c.Flush(); err != nil {
		return stats, err
	}

	stats.Nodes = len(doc.Nodes)
	stats.Edges = len(doc.Edges)
	stats.Communities = len(doc.Communities)
	return stats, nil
}

// markdownExtraction emits a single doc node for a Markdown file (no parsing).
func markdownExtraction(relPath string) *parse.Extraction {
	id := graph.NodeID(relPath, relPath)
	return &parse.Extraction{
		File: relPath,
		Nodes: []graph.Node{
			{ID: id, Label: relPath, Kind: graph.KindDoc, File: relPath, Line: 1, Community: graph.UnclusteredCommunity},
		},
	}
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/app/ -run TestBuild -v`
Expected: PASS (2 tests).

- [ ] **Step 5: Replace the CLI `runBuild` stub**

In `/Users/mylive/project/graffiti/graffiti/cmd/graffiti/main.go`, replace the entire stub function:

```go
// runBuild is a stub so the CLI compiles; Task 13 Step 5 replaces this body verbatim.
func runBuild(root string, stdout, stderr io.Writer) int {
	fmt.Fprintln(stderr, "graffiti: build not yet wired")
	return 1
}
```

with the wired version:

```go
func runBuild(root string, stdout, stderr io.Writer) int {
	generatedAt := time.Now().UTC().Format(time.RFC3339)
	stats, err := app.Build(root, generatedAt)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: build failed: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "✓ Done. 0 API calls, $0.  %d files → %d nodes, %d edges, %d communities.\n",
		stats.Files, stats.Nodes, stats.Edges, stats.Communities)
	return 0
}
```

And update the imports block in the same file from:
```go
import (
	"fmt"
	"io"
	"os"
)
```
to:
```go
import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/evgeniy-achin/graffiti/internal/app"
)
```

- [ ] **Step 6: Add the end-to-end CLI test**

In `/Users/mylive/project/graffiti/graffiti/cmd/graffiti/main_test.go`, remove the placeholder line `var _ = os.Args` (no longer needed) and append:

```go
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
	if !strings.Contains(out.String(), "Done. 0 API calls, $0.") {
		t.Fatalf("missing success line, got %q", out.String())
	}
}
```

Update the test import block at the top of `cmd/graffiti/main_test.go` to:
```go
import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)
```

- [ ] **Step 7: Run the CLI tests to verify they pass**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./cmd/graffiti/ -v`
Expected: PASS (`TestRun_NoArgs_PrintsUsage`, `TestRun_UnknownCommand_Errors`, `TestRun_BuildPrintsSuccessLine`).

- [ ] **Step 8: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/app/app.go internal/app/app_test.go cmd/graffiti/main.go cmd/graffiti/main_test.go
git commit -m "feat: wire scan->parse->build->render pipeline with stable root and success line

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 14: Fixture repo + golden-file test + determinism test (end-to-end)

**Files:**
- Create: `testdata/fixtures/gorepo/go.mod`
- Create: `testdata/fixtures/gorepo/main.go`
- Create: `testdata/fixtures/gorepo/greet/greet.go`
- Create: `testdata/fixtures/gorepo/greet/greet_helper.go`
- Create: `testdata/golden/gorepo.map.json` (generated, inspected, then committed)
- Create: `internal/app/golden_test.go`

- [ ] **Step 1: Create the fixture Go repo**

Create `/Users/mylive/project/graffiti/graffiti/testdata/fixtures/gorepo/go.mod`:
```
module example.com/gorepo

go 1.26
```

Create `/Users/mylive/project/graffiti/graffiti/testdata/fixtures/gorepo/main.go`:
```go
package main

import (
	"fmt"

	"example.com/gorepo/greet"
)

func main() {
	fmt.Println(greet.Hello("world"))
}
```

Create `/Users/mylive/project/graffiti/graffiti/testdata/fixtures/gorepo/greet/greet.go`:
```go
package greet

import "strings"

func Hello(name string) string {
	return "hi " + upper(name)
}

func upper(s string) string {
	return strings.ToUpper(s)
}
```

Create `/Users/mylive/project/graffiti/graffiti/testdata/fixtures/gorepo/greet/greet_helper.go`:
```go
package greet

type Formatter struct {
	Prefix string
}

func (f Formatter) Format(s string) string {
	return f.Prefix + Hello(s)
}
```

- [ ] **Step 2: Write the golden + determinism + structural-assertion test**

Create `/Users/mylive/project/graffiti/graffiti/internal/app/golden_test.go`:
```go
package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

const fixtureGenAt = "2026-06-14T00:00:00Z"

func goldenPath() string {
	return filepath.Join("..", "..", "testdata", "golden", "gorepo.map.json")
}

// buildFixtureIntoTemp copies the committed fixture repo into a temp dir, builds
// it there (so we never write .graffiti into testdata), and returns the produced
// map.json bytes.
func buildFixtureIntoTemp(t *testing.T) []byte {
	t.Helper()
	src := filepath.Join("..", "..", "testdata", "fixtures", "gorepo")
	dst := t.TempDir()
	copyTree(t, src, dst)

	if _, err := Build(dst, fixtureGenAt); err != nil {
		t.Fatalf("Build fixture: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dst, ".graffiti", "map.json"))
	if err != nil {
		t.Fatalf("read produced map.json: %v", err)
	}
	return b
}

func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("copyTree: %v", err)
	}
}

var reGenAt = regexp.MustCompile(`("generated_at":\s*")[^"]*(")`)
var reRoot = regexp.MustCompile(`("root":\s*")[^"]*(")`)

func strip(b []byte) []byte {
	b = reGenAt.ReplaceAll(b, []byte(`${1}X${2}`))
	b = reRoot.ReplaceAll(b, []byte(`${1}X${2}`))
	return b
}

func TestGolden_GoRepoMapJSON(t *testing.T) {
	got := buildFixtureIntoTemp(t)

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(goldenPath(), got, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Log("golden updated")
		return
	}

	want, err := os.ReadFile(goldenPath())
	if err != nil {
		t.Fatalf("read golden (run UPDATE_GOLDEN=1 to create): %v", err)
	}
	if string(strip(got)) != string(strip(want)) {
		t.Fatalf("map.json differs from golden (modulo generated_at+root).\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestDeterminism_TwoBuildsByteIdentical(t *testing.T) {
	a := buildFixtureIntoTemp(t)
	b := buildFixtureIntoTemp(t)
	if string(strip(a)) != string(strip(b)) {
		t.Fatalf("two builds not byte-identical modulo generated_at+root")
	}
}

// TestGolden_StructuralShape enforces the EXPECTED GRAPH by code (not only by a
// frozen blob), so a wrong golden cannot silently pass.
func TestGolden_StructuralShape(t *testing.T) {
	var doc graph.Document
	if err := json.Unmarshal(buildFixtureIntoTemp(t), &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	byID := map[string]graph.Node{}
	for _, n := range doc.Nodes {
		byID[n.ID] = n
	}
	mustNode := func(file, label string, kind graph.Kind) string {
		id := graph.NodeID(file, label)
		n, ok := byID[id]
		if !ok {
			t.Fatalf("missing node %q (%s in %s)", id, label, file)
		}
		if n.Kind != kind {
			t.Fatalf("node %q kind = %q, want %q", id, n.Kind, kind)
		}
		return id
	}

	// file nodes
	mustNode("main.go", "main.go", graph.KindFile)
	mustNode("greet/greet.go", "greet/greet.go", graph.KindFile)
	mustNode("greet/greet_helper.go", "greet/greet_helper.go", graph.KindFile)

	// definitions
	mainID := mustNode("main.go", "main", graph.KindFunction)
	helloID := mustNode("greet/greet.go", "Hello", graph.KindFunction)
	mustNode("greet/greet.go", "upper", graph.KindFunction)
	mustNode("greet/greet_helper.go", "Formatter", graph.KindClass)
	fmtFormatID := mustNode("greet/greet_helper.go", "Formatter.Format", graph.KindMethod)

	// module nodes (keyed by full import path)
	fmtModID := graph.NodeID("module:fmt", "fmt")
	stringsModID := graph.NodeID("module:strings", "strings")
	greetModID := graph.NodeID("module:example.com/gorepo/greet", "greet")
	for _, id := range []string{fmtModID, stringsModID, greetModID} {
		if n, ok := byID[id]; !ok || n.Kind != graph.KindModule {
			t.Fatalf("missing module node %q", id)
		}
	}

	// expected calls edges (deliberate asymmetry, spec §5):
	//   main -> greet (module, EXTRACTED via import)   [greet.Hello selector]
	//   main -> fmt   (module, EXTRACTED via import)    [fmt.Println selector]
	//   Hello -> upper (function, INFERRED, same package)
	//   upper -> strings (module, EXTRACTED via import) [strings.ToUpper selector]
	//   Formatter.Format -> Hello (function, INFERRED)
	wantCalls := map[[2]string]graph.Confidence{
		{mainID, greetModID}:        graph.ConfExtracted,
		{mainID, fmtModID}:          graph.ConfExtracted,
		{helloID, graph.NodeID("greet/greet.go", "upper")}:    graph.ConfInferred,
		{graph.NodeID("greet/greet.go", "upper"), stringsModID}: graph.ConfExtracted,
		{fmtFormatID, helloID}:      graph.ConfInferred,
	}
	gotCalls := map[[2]string]graph.Confidence{}
	for _, e := range doc.Edges {
		if e.Relation == graph.RelCalls {
			gotCalls[[2]string{e.From, e.To}] = e.Confidence
		}
	}
	for k, conf := range wantCalls {
		if got, ok := gotCalls[k]; !ok || got != conf {
			t.Fatalf("calls edge %v: got conf=%q ok=%v, want %q", k, got, ok, conf)
		}
	}
}
```

- [ ] **Step 3: Run the test to verify it fails (no golden file yet)**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/app/ -run TestGolden -v`
Expected: FAIL — `read golden (run UPDATE_GOLDEN=1 to create): open ...gorepo.map.json: no such file or directory`. (`TestGolden_StructuralShape` should already PASS if Pass-1/Pass-2 are correct; if it FAILS, fix Tasks 7/8 before generating the golden — never freeze a wrong tree.)

- [ ] **Step 4: Verify the structural shape, THEN generate the golden**

First confirm the code-asserted shape is correct (this is the real correctness gate, per the feasibility spike's warning):
```bash
cd /Users/mylive/project/graffiti/graffiti
go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/app/ -run TestGolden_StructuralShape -v
```
Expected: PASS. If it fails, fix Pass-1/Pass-2 (Tasks 7-8) and re-run; do not proceed until green.

Then generate the golden:
```bash
cd /Users/mylive/project/graffiti/graffiti
mkdir -p testdata/golden
UPDATE_GOLDEN=1 go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/app/ -run TestGolden_GoRepoMapJSON -v
```
Expected: PASS with log line `golden updated`, and `testdata/golden/gorepo.map.json` now exists. Read it once to sanity-check it matches the structural assertions above (file/function/method/class/module nodes and the five `calls` edges with their EXTRACTED/INFERRED confidences). The structural test is the source of truth for shape; the golden locks byte-exactness.

- [ ] **Step 5: Run the golden + determinism + structural tests to verify they pass**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/app/ -run 'TestGolden|TestDeterminism' -v`
Expected: PASS (`TestGolden_GoRepoMapJSON`, `TestGolden_StructuralShape`, `TestDeterminism_TwoBuildsByteIdentical`).

- [ ] **Step 6: Run the entire test suite with build tags**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./...`
Expected: all packages PASS (`ok` for cmd/graffiti, internal/app, internal/build, internal/cache, internal/graph, internal/parse, internal/render, internal/scan, schema).

- [ ] **Step 7: Commit fixture, golden, and tests**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add testdata/fixtures/gorepo testdata/golden/gorepo.map.json internal/app/golden_test.go
git commit -m "test: golden map.json + structural-shape + determinism tests on Go fixture

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 15: Build helper + cross-compile + end-to-end manual smoke (single static binary)

**Files:**
- Create: `Makefile`
- Create: `README.md`

- [ ] **Step 1: Create the Makefile encoding the required build tags**

Create `/Users/mylive/project/graffiti/graffiti/Makefile`:
```makefile
# graffiti build helpers. The grammar_subset tags ship only the Go grammar,
# keeping the binary small (~8MB) and CGO-free. Without them the code still
# builds, but links the full grammar set (~31MB).
TAGS := grammar_subset grammar_subset_go grammar_subset_gomod
PKG  := ./cmd/graffiti

.PHONY: build test vet xcompile

build:
	CGO_ENABLED=0 go build -tags "$(TAGS)" -o graffiti $(PKG)

test:
	go test -tags "$(TAGS)" ./...

vet:
	go vet -tags "$(TAGS)" ./...

# Cross-compile the static binary for all v1 targets (spec §10).
xcompile:
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -tags "$(TAGS)" -o dist/graffiti-darwin-arm64 $(PKG)
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -tags "$(TAGS)" -o dist/graffiti-darwin-amd64 $(PKG)
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -tags "$(TAGS)" -o dist/graffiti-linux-amd64  $(PKG)
	GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -tags "$(TAGS)" -o dist/graffiti-linux-arm64  $(PKG)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags "$(TAGS)" -o dist/graffiti-windows-amd64.exe $(PKG)
```

- [ ] **Step 2: Run the full test suite via the Makefile**

Run: `make test`
Expected: all packages PASS.

- [ ] **Step 3: Run vet**

Run: `make vet`
Expected: no output (clean).

- [ ] **Step 4: Build the local binary and smoke-test it on the fixture**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
make build
rm -rf /tmp/graffiti-smoke && cp -r testdata/fixtures/gorepo /tmp/graffiti-smoke
./graffiti build /tmp/graffiti-smoke
test -f /tmp/graffiti-smoke/.graffiti/map.json && echo OK
```
Expected stdout: a success line of the form `✓ Done. 0 API calls, $0.  N files → M nodes, K edges, 0 communities.` (the exact counts are whatever the golden encodes — the golden file is the source of truth; do not hardcode counts here), then `OK`. Exit code 0.

- [ ] **Step 5: Cross-compile all targets and confirm static single binaries + size guard**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
make xcompile
ls -la dist/
file dist/graffiti-darwin-arm64
# size guard: the subset-tagged binary must be well under the full-grammar ~31MB.
# A missing-tag regression would balloon it; assert < 16MB.
SIZE=$(wc -c < dist/graffiti-darwin-arm64)
echo "darwin-arm64 size bytes: $SIZE"
test "$SIZE" -lt 16000000 && echo "SIZE OK (subset tags applied)" || (echo "SIZE TOO LARGE — subset tags likely missing" && exit 1)
```
Expected: five binaries in `dist/`, each ~8MB; `file` reports `Mach-O 64-bit executable arm64` for the darwin one; `SIZE OK`. Building for all five GOOS/GOARCH pairs with `CGO_ENABLED=0` is the §10 single-static-binary proof.

- [ ] **Step 6: Write the README**

Create `/Users/mylive/project/graffiti/graffiti/README.md`:
```markdown
# graffiti

One command turns your repository into a directed knowledge graph your AI coding
assistant reads instead of blindly grepping.

> **Status:** Plan 1 (walking skeleton). `graffiti .` builds a deterministic,
> schema-valid `.graffiti/map.json` for a Go repository. Clustering, MAP.md,
> map.html, query, MCP, init, more languages, and workspace federation are later
> plans.

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~8MB, Go grammar only)
make test       # runs the full test suite with required build tags
make xcompile   # cross-compiles static binaries for all v1 targets into dist/
```

The build tags `grammar_subset grammar_subset_go grammar_subset_gomod` ship only
the Go tree-sitter grammar (pure-Go runtime via `github.com/odvcencio/gotreesitter`,
no CGO, no WASM). They are required for the ~8MB size target; without them the code
still compiles but links the full grammar set (~31MB). Always pass them (the
Makefile does this for you).

## Usage

```bash
graffiti .              # build the map for the current repo
graffiti build <path>   # build the map for <path>
graffiti <path>         # shorthand for `build <path>` when <path> is a directory
```

Output: `<path>/.graffiti/map.json` (see `schema/map.schema.json` for the contract)
and a per-file content-hash cache under `<path>/.graffiti/cache/`.

## Guarantees (Plan 1)

- 0 API calls, $0, fully offline.
- Deterministic: same repo → byte-identical `map.json` modulo the single
  `generated_at` timestamp and the `root` basename.
- Single static binary, no runtime dependencies, no C toolchain.
```

- [ ] **Step 7: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add Makefile README.md
git commit -m "build: Makefile with grammar-subset tags, cross-compile, size guard, README

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

- [ ] **Step 8: Final verification — full suite + tidy**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
go mod tidy
make test
git diff --quiet go.mod go.sum || (git add go.mod go.sum && git commit -m "build: go mod tidy

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>")
```
Expected: `make test` prints `ok` for every package. `go mod tidy` leaves a clean, minimal `go.mod`/`go.sum`.

---

## Self-Review

**1. Spec coverage (Plan 1 scope only — later-plan items intentionally excluded):**
- §5 scan → Task 5. ✓
- §5 parse (two-pass, tree-sitter via the verified backend) → Tasks 6 (spike), 7 (Pass 1), 8 (Pass 2). Confidence ladder EXTRACTED/INFERRED with ambiguity-drop matches §5. ✓
- §5 build (validate against schema, directed graph, deterministic IDs, merge-not-replace + anti-shrink) → Tasks 3, 4, 9, 10. ✓
- §6 data model (Node/Edge/Document, directed, community -1 pre-cluster, 1-based line) → Task 2. ✓
- §6 published `schema/map.schema.json` → Task 2. ✓
- §6 content-hash cache under `.graffiti/cache/` (forward-compat, no skip in Plan 1) → Task 11. ✓
- §8.8/§14 determinism (sorted keys/arrays, single generated_at, stable root, byte-identical) → Tasks 10, 12, 13, 14. ✓
- §10 single static CGO-free cross-compilable binary → Tasks 6, 15. **§10's specific WASM+wazero+embed.FS mechanism is substituted (see SPEC-DEVIATION NOTICE / Open Issue #1) — requires owner ratification.** ✓ (constraints) / ⚠ (mechanism)
- §11 CLI `graffiti .` / `graffiti build .` + one success line → Tasks 1, 13. ✓
- §14 golden test + per-language parse fixtures + determinism test + structural assertions → Tasks 6/7/8, 14. ✓
- Markdown in scope as `doc` nodes (discovered, not parsed) → Task 13. ✓
- Explicitly deferred: cluster, analyze, MAP.md, map.html, query, mcp, init, languages beyond Go, workspace, distribution/install.sh — correctly absent.

**2. Blocker/major critic fixes applied:**
- BLOCKER #1 (validator requires generated_at, never set before validation): `build.Assemble(root, generatedAt, exs)` now stamps `doc.GeneratedAt` before validation; `render.WriteMapJSON(doc, root)` reads it off the document. Tasks 9, 10, 12, 13 updated and cross-consistent. ✓
- BLOCKER #2 (absolute temp `root` breaks determinism & golden): `app.Build` sets `docRoot = filepath.Base(absRoot)`; golden/determinism/app tests strip both `generated_at` and `root`. Tasks 13, 14 updated. ✓
- MAJOR (spec §10 mechanism substitution): surfaced as a binding SPEC-DEVIATION NOTICE + Task 6 GATE + Open Issue #1 requiring explicit owner sign-off; isolated behind `parse.Parser`. ✓
- MINOR fixes folded in: cache wiring documented as forward-compat + cache existence asserted (Tasks 11, 13); build-tag claim corrected to "size-only, code still compiles" (Verified Spike Facts, Tasks 6/15, README); fabricated node/edge counts removed from Task 15 + replaced with code-asserted structural test in Task 14; selector→module asymmetry + import-base-collision limitation documented in Task 8 + asserted in Task 14; node_modules/vendor skip now tested (Task 5); `graffiti <path>` shorthand handled and dead `if cmd=="."` branch removed (Task 1); Tree `Release()` note corrected (Verified Spike Facts); `RawCall.ImportSet` renamed to `RawCall.Imports` carrying full import paths and consistently used Tasks 7/8/10. ✓

**3. Cross-task type consistency (checked):**
- `graph.NodeID(file, label)`, `graph.NormalizeID` — Task 3; used Tasks 7, 8, 10, 13, 14. ✓
- `graph.Node/Edge/Community/Document`, `graph.Kind*`/`Rel*`/`Conf*`, `UnclusteredCommunity`, `SchemaVersion`, `NewDocument` — Task 2; used everywhere. ✓
- `parse.Node`/`Tree`/`Parser`, `parse.Walk`, `parse.NewGoParser()` — Task 6; used Tasks 7, 8, 13. ✓
- `parse.ParseGo(p, relPath, src)`, `parse.Extraction{File,Nodes,Edges,RawCalls}`, `parse.RawCall{FromID,Callee,Line,File,Imports}`, `parse.ResolveCalls(defs, calls)` — Tasks 7-8; used Tasks 10, 13, 14. ✓
- `scan.Scan`, `scan.FileRef{AbsPath,RelPath,Lang}`, `scan.LangGo/LangMarkdown` — Task 5; used Task 13. ✓
- `schemaval.ValidateDocument` — Task 9; used Task 10. ✓
- `build.Assemble(root, generatedAt string, []*parse.Extraction)` — Task 10; used Task 13. ✓
- `render.WriteMapJSON(doc, root)` — Task 12; used Task 13. ✓
- `cache.New/HashBytes/HashFile/Put/Get/Load/Flush` — Task 11; used Task 13. ✓
- `app.Build(root, generatedAt) (Stats, error)`, `Stats{Files,Nodes,Edges,Communities,HasDocNode}` — Task 13; used by CLI + golden/app tests. ✓
- Module path `github.com/evgeniy-achin/graffiti` consistent in every import. ✓
- Build tags applied consistently and codified in the Makefile. ✓

All load-bearing `gotreesitter` API calls were verified by compiling and running real code against pinned `v0.20.2`; the Go grammar field names and spike line numbers were confirmed empirically. The plan does not depend on any unverified API.

---

**Plan complete and save location:** save this document to `/Users/mylive/project/graffiti/graffiti/docs/superpowers/plans/2026-06-14-graffiti-plan1-walking-skeleton.md`. Execution options: (1) Subagent-Driven (recommended) — superpowers:subagent-driven-development; (2) Inline — superpowers:executing-plans.
