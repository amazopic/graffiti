# graffiti Plan 4 — LLM-free `query` + hand-rolled MCP `serve` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** After `graffiti build .` writes `.graffiti/map.json` (Plans 1–3, merged), graffiti gains the two read-side capabilities the whole product positions around (spec §5/§7): an **LLM-free `graffiti query "<question>"`** that loads `map.json`, scores nodes by IDF over their label+kind term-bags, seeds the most relevant nodes, BFS-expands neighbors under a **soft token budget** (default 2000), and prints a compact, byte-deterministic text subgraph; and a **`graffiti serve`** MCP server over stdio exposing four tools — `query_graph`, `get_node`, `get_neighbors`, `shortest_path`. The MCP server is **hand-rolled minimal JSON-RPC 2.0 over newline-delimited stdio (stdlib only)** — no MCP SDK — formalized as a §10 spec amendment below (same isolate-and-pivot pattern as Plan 1's parser substitution). Zero API calls, zero keys, fully offline, no LLM; the host model (Claude Code) does all natural-language reasoning over the returned text.

**Architecture:** Same pure-functional read path the spec draws (§5): `query` and `mcp` are a **separate read path** that consumes the artifact written by the build pipeline; they never touch `scan/parse/build/cluster/analyze/layout/render`. Plan 4 adds three packages plus two CLI commands:

```
build pipeline ──writes──▶ .graffiti/map.json
                                  │
                  store.Load ─────┤ (*graph.Document + store.Index adjacency/term-index)
                                  │
                   ┌──────────────┴───────────────┐
              query.Query                      mcp.Server (stdio JSON-RPC 2.0)
            (graffiti query)                   query_graph / get_node /
                                               get_neighbors / shortest_path
                                                    (graffiti serve)
```

- `internal/store` owns the **loader** `Load(path) (*graph.Document, error)` (reads `.graffiti/map.json` back into a `*graph.Document`, mirroring `render.WriteMapJSON`'s `orderedDocument`) plus a deterministic in-memory `Index` (id→node map, sorted out/in adjacency, a term→nodes index aligned with `graph.NormalizeID`). Both `query` and `mcp` build their view from `store.Index`, so there is one place where the on-disk artifact becomes a queryable structure.
- `internal/query` owns the validated LLM-free retrieval: tokenize → IDF-score → seed-by-score (tie-break id asc) → BFS-expand under the token budget → compact deterministic serialize. Pure, `sort`-only, no I/O.
- `internal/mcp` owns the hand-rolled stdio JSON-RPC server: an injectable `Serve(r io.Reader, w io.Writer) error` read/dispatch/write loop handling `initialize` + `tools/list` + `tools/call`, ignoring notifications, exposing the four tools. No external dependency; `encoding/json` + `bufio` only.
- `cmd/graffiti/main.go` gains `query` and `serve` dispatch and updated usage; `internal/app` gains nothing on the build side (read path is independent), though `query`/`serve` link `internal/parse` transitively only via the `cmd` package's existing import graph, so the grammar-subset build tags continue to apply to every `cmd`/`app` command (see Tech Stack).

**Tech Stack:** Go 1.26; module `github.com/amazopic/graffiti`; **no new third-party dependencies** (`store` uses `encoding/json`, `os`, `path/filepath`, `sort`; `query` uses `math`, `sort`, `strings`, plus `golang.org/x/text/unicode/norm` **already in `go.mod`** via `internal/graph` — reused for NFC normalization, not a new dep; `mcp` uses `bufio`, `encoding/json`, `io`, `sort`). `CGO_ENABLED=0`. The grammar-subset build tags from Plan 1 still apply to any target that links `internal/parse` — i.e. `internal/app` and `cmd/graffiti` (which dispatches `build`, `query`, and `serve`) — so every `go test`/`go build` command that touches `cmd/graffiti` or `internal/app` keeps `-tags "grammar_subset grammar_subset_go grammar_subset_gomod"`. The pure `internal/store`, `internal/query`, and `internal/mcp` packages do **not** link `internal/parse` and need no tags, but always passing them is harmless and keeps commands uniform with the `Makefile`.

---

## Validated Prototype Facts (load-bearing; the two risky pieces were prototyped end-to-end in a scratch module and their tests pass)

Before writing this plan the two risky pieces — the **deterministic LLM-free query** and the **hand-rolled minimal MCP stdio server** — were prototyped in a throwaway `package p4proto` against a local mirror of the `graph.*` types, with passing tests. Treat these as ground truth; the task code below is the same validated prototype with the local mirror types replaced by the real `graph.*` types fed through `store.Index`. The prototypes live at `/tmp/p4proto/{query,mcp}.go` (+ `proto_test.go`).

**Query (validated — determinism ×5, budget, relevance, tie-breaks):**
- **IDF scoring over node term-bags.** Each node's term-bag = `tokenize(label)` ∪ `{string(kind)}`. Document frequency `df[t]` is counted over every node's bag; `idf(t) = log((N+1)/(df[t]+1))` (smoothed). A node's score is `Σ idf(qt)` over query terms present in its bag. Only nodes with score > 0 are seeds.
- **Tokenizer** lowercases, splits on any non-alphanumeric boundary, **and** splits camelCase/snake_case (`parseFile` → `parse`,`file`) so identifiers match natural-language terms. (Adaptation: aligned with `graph.NormalizeID` — see deltas below.)
- **Seed + BFS expansion under a token budget.** Seeds are added highest-score-first (ties already broken id-asc by the score sort) until the budget would be exceeded; then BFS walks the selection in order, pulling each node's neighbors (out-edges then in-edges, each group sorted by `(relation, other-id, confidence)`) into the set until the budget is hit. `estimateTokens(s) = len(s)/4` (min 1 for non-empty) — a **deterministic, zero-dependency soft estimate**, not a real tokenizer, chosen for determinism and zero deps per the project ethos.
- **Budget covers NODES only (validated decision).** The prototype's `add()` charges `estimateTokens(formatNode(n))` per selected node; edges are emitted for free for every pair whose endpoints are both selected. **Decision (carried into this plan):** the budget is a **soft node budget**; edge text is **not** counted. Rationale: edges are bounded by the selected node set (≤ selected² in the worst case but in practice sparse), the budget's job is to bound *which nodes* enter the context, and counting edges would make node selection depend on edge emission order (a determinism hazard). This is documented in code, in the CLI help, and in Open Issue #1.
- **Compact deterministic serialization.** Output is two blocks:
  - `NODES` — every selected node, **sorted by id**, one per line as `id [kind] label @ file:line`.
  - `EDGES` — every edge whose **both** endpoints are selected, de-duped on `(from,to,relation,confidence)`, **sorted by `(from,to,relation,confidence)`**, one per line as `from -relation-> to (CONFIDENCE)`.
- **Determinism proven:** the prototype runs the full query 5× on the same graph+question and asserts byte-identical output; it asserts the budget is respected; it asserts a relevant seed is selected for a targeted query; it asserts tie-broken seed order (equal-score nodes ordered by id asc); and it asserts **no Go-map iteration ever feeds emitted order** (every map is consumed through an explicitly sorted key list).

**MCP (validated — hand-rolled, no SDK):**
- **MCP's stdio transport is plain JSON-RPC 2.0 with NEWLINE-DELIMITED framing** (one JSON object per line) — **NOT** the LSP-style `Content-Length:` header framing. (Verified against the MCP stdio transport docs.) The server reads with a `bufio.Scanner` (line-delimited, buffer raised to 4 MiB for large `tools/call` results) and writes with a `json.Encoder` (which appends `\n` per `Encode`, giving newline framing for free).
- **Three JSON-RPC methods + notification tolerance.** `initialize`, `tools/list`, `tools/call`; any request with **no `id`** is a notification (e.g. `notifications/initialized`) and gets **no reply**; an unparseable line emits a single `-32700` parse-error response (id `null`); an unknown method emits `-32601`.
- **`Serve(r io.Reader, w io.Writer) error`** is injectable so tests drive it with `strings.Reader`/`bytes.Buffer` — no real stdin/stdout needed. The CLI wires `os.Stdin`/`os.Stdout`.
- The prototype wired only `query_graph`; this plan adds the other three tools (`get_node`, `get_neighbors`, `shortest_path`) using the same `handleCall` switch and the same `callToolResult{Content:[]textContent{...}}` shape.

**Adaptation deltas from prototype → real code (the only differences):**
1. Local mirror `Graph` (with `nodes`/`order`/`out`/`in` fields) → **`store.Index`** built by `store.Load`. The query and mcp code take a `*store.Index` (or the fields it exposes) instead of the mirror `*Graph`; field accessors (`g.nodes[id]`, `g.order`, `g.out[id]`, `g.in[id]`) become `Index` methods/fields with the same semantics and the same deterministic sorting.
2. Local mirror `Node`/`Edge`/`Kind`/`Relation`/`Confidence` → `graph.Node`/`graph.Edge` and the string-backed `graph.Kind`/`graph.Relation`/`graph.Confidence` (so `string(n.Kind)`, `string(e.Relation)`, `string(e.Confidence)` where the prototype used the mirror string types — the prototype already does exactly this).
3. **Tokenizer normalization aligned with `graph.NormalizeID` (the open issue called out in scope).** Node ids are NFC+casefold+`\w`-collapse slugs (`internal/graph/id.go`). To make query terms match consistently, the query tokenizer **NFC-normalizes** the input (via `golang.org/x/text/unicode/norm`, the same dep `graph` already uses) before lowercasing/splitting, and treats any Unicode letter/digit as a word rune (mirroring `graph.isWordRune`), so a query term and the label it should match fold to the same tokens. The camelCase/snake_case split is retained on top of that. Documented in `query.go` and Open Issue #2.
4. The prototype's single-tool catalog → a four-tool catalog; `query_graph` is byte-for-byte the validated handler, the three new handlers reuse `Index` lookups and the validated serializer helpers (`formatNode`/`formatEdge`).
5. `ProtocolVersion` is no longer a single hardcoded constant returned blindly: on `initialize` the server **echoes the client's requested `protocolVersion` if it is in a small allow-list, else returns the server's latest** (see Task 4). The prototype returned a constant; this is the one behavioral change, and it is unit-tested.

---

## File Structure

```
graffiti/
├── internal/
│   ├── store/
│   │   ├── store.go            # NEW: Load(path) (*graph.Document, error) + Index (id map, sorted adjacency, term index)
│   │   └── store_test.go       # NEW: round-trip load, adjacency sorted, term index, missing-file error
│   ├── query/
│   │   ├── query.go            # NEW: Query(idx, question, budget) string — IDF + seed + BFS + soft node budget + serialize
│   │   └── query_test.go       # NEW: determinism / budget / relevance / tie-break / serialize-order
│   ├── mcp/
│   │   ├── mcp.go              # NEW: hand-rolled JSON-RPC 2.0 stdio Server; 4 tools; initialize version-echo
│   │   └── mcp_test.go         # NEW: initialize(+version echo) / tools/list / each tools/call / notification / parse-error
│   └── app/                    # (unchanged — read path is independent of the build pipeline)
├── cmd/graffiti/
│   ├── main.go                 # MODIFIED: add `query` + `serve` dispatch; updated usage
│   └── main_test.go            # MODIFIED: query end-to-end + serve smoke (run loop over injected reader)
└── testdata/golden/
    └── gorepo.query.txt        # NEW (via UPDATE_GOLDEN): query output on the gorepo fixture (no generated_at — nothing to strip)
```

**Package responsibilities (one job each):**
- `store` owns the **read-side loader + index**: it is the single place that turns the on-disk `map.json` artifact into an in-memory, deterministically-sorted queryable structure. It mirrors `render.WriteMapJSON`'s shape (so a round-trip is lossless) and performs the only I/O on the read path.
- `query` owns the deterministic LLM-free retrieval (IDF + BFS + soft node budget + compact serialize) — pure, `sort`-only, consumes a `*store.Index`.
- `mcp` owns the hand-rolled stdio JSON-RPC server and the four tool handlers — `encoding/json`+`bufio`+`io` only, `Serve(r,w)` injectable.
- `cmd/graffiti` wires the two new commands; the build pipeline (`app` and everything below it) is untouched, so all Plan 1–3 goldens stay byte-identical.

---

## §10 SPEC AMENDMENT — MCP is hand-rolled minimal JSON-RPC 2.0 over stdio (stdlib only)

> **RATIFIED 2026-06-16 — §10 amended to adopt the hand-rolled stdlib MCP server (official SDK is a documented drop-in fallback behind `internal/mcp.Serve`). All tasks may proceed.** Spec §10 originally named the transport as *"MCP: Go MCP SDK over stdio."* This plan instead implements MCP as a **hand-rolled minimal JSON-RPC 2.0 server over newline-delimited stdio using only the Go standard library** (`encoding/json` + `bufio` + `io`), with **no MCP SDK dependency**. This is the **same isolate-and-pivot pattern Plan 1 applied to the parser** (`gotreesitter` substituted for the spec's `wazero`+WASM, isolated behind `parse.Parser`): here the substitution is isolated inside `internal/mcp`, which exposes a single `Serve(r io.Reader, w io.Writer) error` seam, so adopting the official SDK later (if its dependency footprint shrinks or graffiti needs richer MCP features) is a drop-in replacement affecting only `internal/mcp`.
>
> **Reasoning (the project ethos, §3/§10):** graffiti's three wedges rest on a *single static pure-Go binary with zero runtime dependencies, fully offline, byte-deterministic*. The official Go MCP SDK pulls **~10 transitive dependencies** (JWT/OAuth2/HTTP-auth machinery and friends) that graffiti's **stdio-only, keyless, offline** server needs none of — they grow the binary, widen the supply-chain surface, and add code paths (auth, HTTP transport) that contradict the "no auth, no HTTP MCP server" non-goal (§4). graffiti speaks exactly three JSON-RPC methods over one stdio stream; that is ~150 lines of stdlib JSON. Hand-rolling keeps the binary lean, the deps at zero, and the wire behavior fully under our control and trivially unit-testable via an injected reader/writer. The wire format used is the validated newline-delimited JSON-RPC 2.0 the MCP stdio transport specifies (NOT Content-Length framing).
>
> **Scope of the amendment:** transport/library only. The *capability* §5 mandates — exposing `query`/`get_node`/`neighbors`/`path` as MCP tools over stdio — is delivered in full (all four tools). **Proposed resolution: ratify option (a) — §10 amended to "hand-rolled minimal JSON-RPC 2.0 over stdio (stdlib only); official SDK is a documented drop-in fallback behind `internal/mcp.Serve`."** Tasks 1–3, 5–6 do not depend on this and may proceed regardless; Task 4 (the MCP server) is gated on this ratification.

---

## Task 1: `internal/store` — `map.json` loader + deterministic in-memory `Index`

**Files:**
- Create: `internal/store/store.go`
- Create: `internal/store/store_test.go`

> The loader is the mirror of `render.WriteMapJSON` (which marshals an `orderedDocument` with json tags `communities/edges/generated_at/nodes/root/version`). `graph.Document`'s own json tags (`internal/graph/graph.go`) already match those field names, so `json.Unmarshal` straight into a `*graph.Document` is lossless for the public fields (the unexported `nodeHighWaterMark` is not serialized and is irrelevant to the read path). `Index` builds the deterministic adjacency + term index the query/mcp code needs; **every slice it exposes is sorted with an explicit total order** so no map iteration ever feeds query output (spec §14).

- [ ] **Step 1: Write the failing loader + index test**

Create `/Users/mylive/project/graffiti/graffiti/internal/store/store_test.go`:
```go
package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/amazopic/graffiti/internal/graph"
)

// writeMap writes a minimal valid map.json (same key shape render.WriteMapJSON
// emits) into <dir>/.graffiti/map.json and returns dir.
func writeMap(t *testing.T, doc *graph.Document) string {
	t.Helper()
	dir := t.TempDir()
	gdir := filepath.Join(dir, ".graffiti")
	if err := os.MkdirAll(gdir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Marshal via the public json tags (mirror of render's orderedDocument).
	b, err := json.MarshalIndent(struct {
		Communities []graph.Community `json:"communities"`
		Edges       []graph.Edge      `json:"edges"`
		GeneratedAt string            `json:"generated_at"`
		Nodes       []graph.Node      `json:"nodes"`
		Root        string            `json:"root"`
		Version     string            `json:"version"`
	}{doc.Communities, doc.Edges, doc.GeneratedAt, doc.Nodes, doc.Root, doc.Version}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gdir, "map.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func sampleDoc() *graph.Document {
	doc := graph.NewDocument("demo")
	doc.GeneratedAt = "2026-06-16T00:00:00Z"
	doc.Nodes = []graph.Node{
		{ID: "f.go:parsefile", Label: "parseFile", Kind: graph.KindFunction, File: "f.go", Line: 10, Community: 0},
		{ID: "f.go:readall", Label: "readAll", Kind: graph.KindFunction, File: "f.go", Line: 20, Community: 0},
		{ID: "g.go:cache", Label: "Cache", Kind: graph.KindClass, File: "g.go", Line: 3, Community: 1},
	}
	doc.Edges = []graph.Edge{
		{From: "f.go:parsefile", To: "f.go:readall", Relation: graph.RelCalls, Confidence: graph.ConfExtracted},
		{From: "f.go:parsefile", To: "g.go:cache", Relation: graph.RelReferences, Confidence: graph.ConfInferred},
	}
	doc.Communities = []graph.Community{
		{ID: 0, Label: "Parser", Members: []string{"f.go:parsefile", "f.go:readall"}},
		{ID: 1, Label: "Cache", Members: []string{"g.go:cache"}},
	}
	return doc
}

func TestLoad_RoundTrip(t *testing.T) {
	dir := writeMap(t, sampleDoc())
	doc, err := Load(filepath.Join(dir, ".graffiti", "map.json"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if doc.Version != graph.SchemaVersion {
		t.Fatalf("version = %q, want %q", doc.Version, graph.SchemaVersion)
	}
	if len(doc.Nodes) != 3 || len(doc.Edges) != 2 || len(doc.Communities) != 2 {
		t.Fatalf("loaded counts wrong: %d nodes, %d edges, %d communities", len(doc.Nodes), len(doc.Edges), len(doc.Communities))
	}
	if doc.Nodes[0].Label != "parseFile" || doc.Nodes[0].Kind != graph.KindFunction {
		t.Fatalf("node[0] mismatch: %+v", doc.Nodes[0])
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), ".graffiti", "map.json"))
	if err == nil {
		t.Fatal("expected error for missing map.json, got nil")
	}
}

func TestIndex_AdjacencySortedAndDeterministic(t *testing.T) {
	idx := NewIndex(sampleDoc())

	// Node lookup.
	n, ok := idx.Node("f.go:parsefile")
	if !ok || n.Label != "parseFile" {
		t.Fatalf("Node(parsefile) = %+v, %v", n, ok)
	}
	if _, ok := idx.Node("nope"); ok {
		t.Fatal("Node(nope) should be absent")
	}

	// IDs are sorted ascending.
	want := []string{"f.go:parsefile", "f.go:readall", "g.go:cache"}
	if got := idx.IDs(); len(got) != 3 || got[0] != want[0] || got[2] != want[2] {
		t.Fatalf("IDs() = %v, want %v", got, want)
	}

	// Out-edges of parsefile are sorted by (relation, to, confidence): calls<references.
	out := idx.Out("f.go:parsefile")
	if len(out) != 2 || out[0].Relation != graph.RelCalls || out[1].Relation != graph.RelReferences {
		t.Fatalf("Out(parsefile) not sorted by relation: %+v", out)
	}

	// Building twice yields identical adjacency order (no map-iteration leakage).
	a, b := NewIndex(sampleDoc()), NewIndex(sampleDoc())
	for _, id := range a.IDs() {
		oa, ob := a.Out(id), b.Out(id)
		if len(oa) != len(ob) {
			t.Fatalf("out len mismatch for %s", id)
		}
		for i := range oa {
			if oa[i] != ob[i] {
				t.Fatalf("non-deterministic out-edge order for %s at %d", id, i)
			}
		}
	}
}
```

- [ ] **Step 2: Run the test, watch it fail to compile (no `store.go` yet)**

Run: `go test ./internal/store/ -run 'TestLoad|TestIndex' 2>&1 | head`
Expected: build failure — `undefined: Load`, `undefined: NewIndex`. This is the red of TDD; Step 3 makes it green.

- [ ] **Step 3: Implement `store.go` (loader + Index)**

Create `/Users/mylive/project/graffiti/graffiti/internal/store/store.go`:
```go
// Package store is graffiti's read-side: it loads the .graffiti/map.json artifact
// back into a *graph.Document (the mirror of render.WriteMapJSON) and builds a
// deterministic in-memory Index (id->node, sorted out/in adjacency, term index)
// that internal/query and internal/mcp consume. It performs the only I/O on the
// read path and never touches the build pipeline.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/amazopic/graffiti/internal/graph"
)

// Load reads a map.json file at path and unmarshals it into a *graph.Document.
// graph.Document's json tags (version/generated_at/root/nodes/edges/communities)
// match the keys render.WriteMapJSON emits, so this is a lossless round-trip of
// every public field (the unexported high-water mark is not serialized and is
// unused on the read path).
func Load(path string) (*graph.Document, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("store: read %s: %w", path, err)
	}
	var doc graph.Document
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("store: parse %s: %w", path, err)
	}
	return &doc, nil
}

// Index is a deterministic, read-only adjacency + term index over a Document.
// Every exposed slice is sorted with an explicit total order so no Go-map
// iteration ever feeds query/mcp output (spec §14).
type Index struct {
	nodes map[string]graph.Node
	ids   []string                 // all node ids, sorted ascending
	out   map[string][]graph.Edge  // out-edges per id, sorted (relation,to,confidence)
	in    map[string][]graph.Edge  // in-edges per id, sorted (relation,from,confidence)
	terms map[string][]string      // term -> node ids (sorted), term index for IDF
}

// NewIndex builds the Index from a loaded Document.
func NewIndex(doc *graph.Document) *Index {
	idx := &Index{
		nodes: make(map[string]graph.Node, len(doc.Nodes)),
		out:   make(map[string][]graph.Edge),
		in:    make(map[string][]graph.Edge),
		terms: make(map[string][]string),
	}
	for _, n := range doc.Nodes {
		idx.nodes[n.ID] = n
		idx.ids = append(idx.ids, n.ID)
	}
	sort.Strings(idx.ids)

	for _, e := range doc.Edges {
		idx.out[e.From] = append(idx.out[e.From], e)
		idx.in[e.To] = append(idx.in[e.To], e)
	}
	for id := range idx.out {
		sortEdges(idx.out[id], func(e graph.Edge) string { return e.To })
	}
	for id := range idx.in {
		sortEdges(idx.in[id], func(e graph.Edge) string { return e.From })
	}
	return idx
}

// sortEdges sorts edges by (relation, other-endpoint, confidence). `otherOf`
// selects the endpoint that is NOT the indexed node (To for out, From for in).
func sortEdges(es []graph.Edge, otherOf func(graph.Edge) string) {
	sort.SliceStable(es, func(i, j int) bool {
		a, b := es[i], es[j]
		if a.Relation != b.Relation {
			return a.Relation < b.Relation
		}
		oa, ob := otherOf(a), otherOf(b)
		if oa != ob {
			return oa < ob
		}
		return a.Confidence < b.Confidence
	})
}

// Node returns the node for id and whether it exists.
func (x *Index) Node(id string) (graph.Node, bool) { n, ok := x.nodes[id]; return n, ok }

// IDs returns all node ids, sorted ascending (a fresh copy is unnecessary —
// callers must not mutate; query only reads).
func (x *Index) IDs() []string { return x.ids }

// Out returns the sorted out-edges of id (nil if none).
func (x *Index) Out(id string) []graph.Edge { return x.out[id] }

// In returns the sorted in-edges of id (nil if none).
func (x *Index) In(id string) []graph.Edge { return x.in[id] }

// Len returns the node count (corpus size N for IDF).
func (x *Index) Len() int { return len(x.ids) }
```

> The `terms` field is declared here but populated lazily by `query` (Task 2) via the tokenizer that lives in `query` — keeping the NFC/casefold tokenizer in one package. `NewIndex` deliberately does **not** import `query` (that would be a cycle); instead `query` reads `Index.IDs()`/`Index.Node()` to build its own per-query IDF tables. (We keep the `terms` map declaration minimal; if it proves unused after Task 2, drop it — the determinism contract is in `out`/`in`/`ids`.)

- [ ] **Step 4: Run the store tests — green**

Run: `go test ./internal/store/ -run 'TestLoad|TestIndex' -v`
Expected: PASS — `TestLoad_RoundTrip`, `TestLoad_MissingFile`, `TestIndex_AdjacencySortedAndDeterministic`.

- [ ] **Step 5: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/store/store.go internal/store/store_test.go
git commit -m "feat(store): map.json loader + deterministic in-memory adjacency Index

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: `internal/query` — LLM-free IDF + BFS + soft node budget + compact serialize

**Files:**
- Create: `internal/query/query.go`
- Create: `internal/query/query_test.go`

> This is the validated `/tmp/p4proto/query.go` retargeted from the mirror `*Graph` to `*store.Index`, with the tokenizer NFC-aligned to `graph.NormalizeID` (delta #3). The budget is the **soft node budget** decided above: `estimateTokens(formatNode(n))` is charged per selected node; **edges are not budgeted**. Everything that feeds emitted order is sorted (spec §14).

- [ ] **Step 1: Write the failing query test (determinism / budget / relevance / tie-break / serialize-order)**

Create `/Users/mylive/project/graffiti/graffiti/internal/query/query_test.go`:
```go
package query

import (
	"strings"
	"testing"

	"github.com/amazopic/graffiti/internal/graph"
	"github.com/amazopic/graffiti/internal/store"
)

func idxFor(nodes []graph.Node, edges []graph.Edge) *store.Index {
	doc := graph.NewDocument("demo")
	doc.Nodes = nodes
	doc.Edges = edges
	return store.NewIndex(doc)
}

func sampleIndex() *store.Index {
	nodes := []graph.Node{
		{ID: "auth.go:login", Label: "loginHandler", Kind: graph.KindFunction, File: "auth.go", Line: 12, Community: 0},
		{ID: "auth.go:session", Label: "createSession", Kind: graph.KindFunction, File: "auth.go", Line: 30, Community: 0},
		{ID: "http.go:router", Label: "routeRequest", Kind: graph.KindFunction, File: "http.go", Line: 5, Community: 1},
		{ID: "cache.go:get", Label: "cacheGet", Kind: graph.KindFunction, File: "cache.go", Line: 8, Community: 2},
	}
	edges := []graph.Edge{
		{From: "http.go:router", To: "auth.go:login", Relation: graph.RelCalls, Confidence: graph.ConfExtracted},
		{From: "auth.go:login", To: "auth.go:session", Relation: graph.RelCalls, Confidence: graph.ConfExtracted},
		{From: "auth.go:login", To: "cache.go:get", Relation: graph.RelReferences, Confidence: graph.ConfInferred},
	}
	return idxFor(nodes, edges)
}

func TestQuery_Deterministic(t *testing.T) {
	idx := sampleIndex()
	first := Query(idx, "where is login session auth handled", 2000)
	for i := 0; i < 5; i++ {
		if got := Query(idx, "where is login session auth handled", 2000); got != first {
			t.Fatalf("run %d differs:\n%s\n---\n%s", i, first, got)
		}
	}
}

func TestQuery_Relevance(t *testing.T) {
	out := Query(sampleIndex(), "login session", 2000)
	if !strings.Contains(out, "auth.go:login") {
		t.Fatalf("expected login node selected, got:\n%s", out)
	}
	if !strings.Contains(out, "auth.go:session") {
		t.Fatalf("expected session node pulled in by expansion, got:\n%s", out)
	}
}

func TestQuery_BudgetRespected(t *testing.T) {
	idx := sampleIndex()
	// A tiny budget admits only the single highest-scoring seed (one node line).
	out := Query(idx, "login", 8)
	nodeLines := 0
	for _, l := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.Contains(l, " @ ") { // node lines have " @ file:line"
			nodeLines++
		}
	}
	if nodeLines == 0 {
		t.Fatalf("tiny budget should still admit one seed, got:\n%s", out)
	}
	if nodeLines > 1 {
		t.Fatalf("budget=8 admitted %d node lines (expected 1), got:\n%s", nodeLines, out)
	}
}

func TestQuery_SeedTieBreakByID(t *testing.T) {
	// Two nodes with the SAME single distinguishing term => equal score; the
	// one with the smaller id must seed first.
	nodes := []graph.Node{
		{ID: "b.go:widget", Label: "widget", Kind: graph.KindFunction, File: "b.go", Line: 1},
		{ID: "a.go:widget", Label: "widget", Kind: graph.KindFunction, File: "a.go", Line: 1},
	}
	out := Query(idxFor(nodes, nil), "widget", 2000)
	ia := strings.Index(out, "a.go:widget")
	ib := strings.Index(out, "b.go:widget")
	if ia < 0 || ib < 0 || ia > ib {
		t.Fatalf("expected a.go:widget before b.go:widget (id-asc), got:\n%s", out)
	}
}

func TestQuery_SerializeOrderAndEdges(t *testing.T) {
	out := Query(sampleIndex(), "router login session cache", 2000)
	if !strings.Contains(out, "NODES\n") || !strings.Contains(out, "EDGES\n") {
		t.Fatalf("missing NODES/EDGES blocks:\n%s", out)
	}
	// Edge lines look like "from -relation-> to (CONFIDENCE)".
	if !strings.Contains(out, "auth.go:login -calls-> auth.go:session (EXTRACTED)") {
		t.Fatalf("expected serialized calls edge, got:\n%s", out)
	}
}

func TestQuery_EmptyQuestion(t *testing.T) {
	if out := Query(sampleIndex(), "   ", 2000); strings.TrimSpace(out) != "NODES\nEDGES" && !strings.HasPrefix(out, "NODES") {
		t.Fatalf("empty question should yield empty NODES/EDGES, got:\n%q", out)
	}
}
```

- [ ] **Step 2: Run the test, watch it fail to compile (no `query.go` yet)**

Run: `go test ./internal/query/ -run TestQuery 2>&1 | head`
Expected: build failure — `undefined: Query`. Red.

- [ ] **Step 3: Implement `query.go` (validated prototype, retargeted to `store.Index`)**

Create `/Users/mylive/project/graffiti/graffiti/internal/query/query.go`:
```go
// Package query is graffiti's LLM-free retrieval path (spec §7): tokenize the
// question, score nodes by IDF over their label+kind term-bags, seed by score
// (tie-break id asc), BFS-expand neighbors under a SOFT TOKEN BUDGET, and
// serialize the scoped subgraph to compact deterministic text. No inference call
// is made; the host model reasons over the returned text. Pure, sort-only, no I/O.
package query

import (
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/amazopic/graffiti/internal/graph"
	"github.com/amazopic/graffiti/internal/store"

	"golang.org/x/text/unicode/norm"
)

// DefaultTokenBudget mirrors spec §7 ("default ~2000"). The budget is a SOFT
// estimate over NODE text only (see estimateTokens / the budget note below).
const DefaultTokenBudget = 2000

// estimateTokens is a deterministic, zero-dependency token estimator (~4 chars
// per token). NOT a real tokenizer — chosen for determinism and zero deps per
// the project ethos (spec §3/§10). The budget bounds NODE selection only; edge
// text is emitted for free for every selected pair (see Query / Open Issue #1).
func estimateTokens(s string) int {
	n := len(s) / 4
	if n == 0 && len(s) > 0 {
		n = 1
	}
	return n
}

// tokenize NFC-normalizes (aligning with graph.NormalizeID so query terms fold
// the same way node-id slugs do), lowercases, splits on any non-word boundary,
// and ALSO splits camelCase/snake_case so "parseFile" matches "parse"/"file".
func tokenize(s string) []string {
	s = norm.NFC.String(s)
	var out []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, strings.ToLower(cur.String()))
			cur.Reset()
		}
	}
	var prev rune
	for _, r := range s {
		switch {
		case isLower(r), isDigit(r):
			cur.WriteRune(r)
		case isUpper(r):
			if isLower(prev) || isDigit(prev) { // camelCase boundary
				flush()
			}
			cur.WriteRune(r)
		case isWordRune(r): // non-ASCII letter/digit (mirrors graph.isWordRune)
			cur.WriteRune(r)
		default:
			flush()
		}
		prev = r
	}
	flush()
	return out
}

func isLower(r rune) bool { return r >= 'a' && r <= 'z' }
func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }
func isDigit(r rune) bool { return r >= '0' && r <= '9' }

// isWordRune mirrors graph.isWordRune: ASCII letters/digits plus any non-ASCII
// Unicode letter or digit (so accented/CJK identifiers tokenize like the slug).
func isWordRune(r rune) bool {
	return isLower(r) || isUpper(r) || isDigit(r) || (r > 0x7F && (unicode.IsLetter(r) || unicode.IsDigit(r)))
}

// nodeTerms returns a node's bag of terms: its label tokens plus its kind.
func nodeTerms(n graph.Node) []string {
	return append(tokenize(n.Label), string(n.Kind))
}

type scoredNode struct {
	id    string
	score float64
}

// scoreNodes computes IDF-weighted overlap of every node against the query terms
// and returns the scored ids sorted by (score desc, id asc).
func scoreNodes(idx *store.Index, question string) []scoredNode {
	qTerms := dedupe(tokenize(question))
	if len(qTerms) == 0 {
		return nil
	}
	ids := idx.IDs()
	df := make(map[string]int)
	bags := make(map[string]map[string]bool, len(ids))
	for _, id := range ids {
		n, _ := idx.Node(id)
		bag := make(map[string]bool)
		for _, t := range nodeTerms(n) {
			bag[t] = true
		}
		bags[id] = bag
		for t := range bag {
			df[t]++
		}
	}
	N := float64(idx.Len())
	idf := func(term string) float64 { return math.Log((N + 1) / (float64(df[term]) + 1)) }

	scored := make([]scoredNode, 0, len(ids))
	for _, id := range ids { // ids are already sorted; iteration is deterministic
		bag := bags[id]
		var s float64
		for _, qt := range qTerms {
			if bag[qt] {
				s += idf(qt)
			}
		}
		if s > 0 {
			scored = append(scored, scoredNode{id: id, score: s})
		}
	}
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		return scored[i].id < scored[j].id // tie-break id asc
	})
	return scored
}

// Query runs the full LLM-free retrieval and returns the serialized subgraph.
func Query(idx *store.Index, question string, budget int) string {
	if budget <= 0 {
		budget = DefaultTokenBudget
	}
	scored := scoreNodes(idx, question)

	selected := make(map[string]bool)
	var orderedSel []string
	used := 0
	add := func(id string) bool {
		if selected[id] {
			return true
		}
		n, ok := idx.Node(id)
		if !ok {
			return true
		}
		cost := estimateTokens(formatNode(n)) // SOFT NODE BUDGET — edges are free
		if used+cost > budget {
			return false
		}
		selected[id] = true
		orderedSel = append(orderedSel, id)
		used += cost
		return true
	}

	for _, sn := range scored { // seeds, highest score first
		if !add(sn.id) {
			break
		}
	}
	for i := 0; i < len(orderedSel); i++ { // BFS expansion
		if used >= budget {
			break
		}
		for _, e := range neighbors(idx, orderedSel[i]) {
			other := e.To
			if other == orderedSel[i] {
				other = e.From
			}
			_ = add(other) // budget-exhausted neighbors are skipped, keep trying smaller ones
		}
	}
	return serialize(idx, selected)
}

// neighbors returns out-edges then in-edges of id, each group already sorted by
// (relation, other-id, confidence) by store.Index — no re-sort, no map iteration.
func neighbors(idx *store.Index, id string) []graph.Edge {
	out := idx.Out(id)
	in := idx.In(id)
	es := make([]graph.Edge, 0, len(out)+len(in))
	es = append(es, out...)
	es = append(es, in...)
	return es
}

// serialize renders the selected subgraph: a NODES block (sorted by id) and an
// EDGES block (edges with BOTH endpoints selected, de-duped, sorted).
func serialize(idx *store.Index, selected map[string]bool) string {
	ids := make([]string, 0, len(selected))
	for id := range selected {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var b strings.Builder
	b.WriteString("NODES\n")
	for _, id := range ids {
		n, _ := idx.Node(id)
		b.WriteString(formatNode(n))
		b.WriteByte('\n')
	}

	type ek struct {
		from, to string
		rel      graph.Relation
		conf     graph.Confidence
	}
	seen := make(map[ek]bool)
	var edges []graph.Edge
	for _, id := range ids { // ids sorted => deterministic collection
		for _, e := range idx.Out(id) {
			if !selected[e.To] {
				continue
			}
			k := ek{e.From, e.To, e.Relation, e.Confidence}
			if seen[k] {
				continue
			}
			seen[k] = true
			edges = append(edges, e)
		}
	}
	sort.SliceStable(edges, func(i, j int) bool {
		a, c := edges[i], edges[j]
		if a.From != c.From {
			return a.From < c.From
		}
		if a.To != c.To {
			return a.To < c.To
		}
		if a.Relation != c.Relation {
			return a.Relation < c.Relation
		}
		return a.Confidence < c.Confidence
	})
	b.WriteString("EDGES\n")
	for _, e := range edges {
		b.WriteString(formatEdge(e))
		b.WriteByte('\n')
	}
	return b.String()
}

func formatNode(n graph.Node) string {
	return n.ID + " [" + string(n.Kind) + "] " + n.Label + " @ " + n.File + ":" + itoa(n.Line)
}

func formatEdge(e graph.Edge) string {
	return e.From + " -" + string(e.Relation) + "-> " + e.To + " (" + string(e.Confidence) + ")"
}

func dedupe(ss []string) []string {
	seen := make(map[string]bool, len(ss))
	out := ss[:0]
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
```

> **`formatNode`/`formatEdge` are exported-by-need to `mcp`.** `get_node`/`get_neighbors`/`shortest_path` (Task 4) need to render single nodes/edges in the same format. To avoid duplicating the formatter, expose thin wrappers `FormatNode(graph.Node) string` and `FormatEdge(graph.Edge) string` and `Serialize(idx *store.Index, ids []string) string` from this package (one-line exported aliases over the internal funcs), and have `mcp` call those. Add them in Step 3 alongside the internal funcs (exported funcs are simple pass-throughs; keep the internal ones for the unexported tests). This keeps the serialization format single-sourced and deterministic.

- [ ] **Step 4: Run the query tests — green**

Run: `go test ./internal/query/ -run TestQuery -v`
Expected: PASS — `TestQuery_Deterministic`, `TestQuery_Relevance`, `TestQuery_BudgetRespected`, `TestQuery_SeedTieBreakByID`, `TestQuery_SerializeOrderAndEdges`, `TestQuery_EmptyQuestion`.

- [ ] **Step 5: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/query/query.go internal/query/query_test.go
git commit -m "feat(query): LLM-free IDF+BFS retrieval, soft node budget, deterministic serialize

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: `internal/mcp` — hand-rolled stdio JSON-RPC 2.0 server, four tools, version echo

> **GATE CLEARED (ratified 2026-06-16):** §10 is amended to adopt the hand-rolled stdlib JSON-RPC server (official SDK is a documented drop-in fallback behind `internal/mcp.Serve`) — proceed.

**Files:**
- Create: `internal/mcp/mcp.go`
- Create: `internal/mcp/mcp_test.go`

> This is the validated `/tmp/p4proto/mcp.go` with three changes: (1) the `Server` holds a `*store.Index` and exposes four tools; (2) `initialize` echoes the client's `protocolVersion` if allow-listed, else returns the server's latest; (3) `query_graph`/`get_node`/`get_neighbors`/`shortest_path` handlers reuse `query.FormatNode`/`FormatEdge`/`Serialize` (Task 2) for byte-identical formatting. Wire format stays newline-delimited JSON-RPC 2.0 (NOT Content-Length).

- [ ] **Step 1: Write the failing MCP test (initialize+version-echo / tools/list / each tools/call / notification / parse-error)**

Create `/Users/mylive/project/graffiti/graffiti/internal/mcp/mcp_test.go`:
```go
package mcp

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/amazopic/graffiti/internal/graph"
	"github.com/amazopic/graffiti/internal/store"
)

func testServer() *Server {
	doc := graph.NewDocument("demo")
	doc.Nodes = []graph.Node{
		{ID: "a.go:f", Label: "fooHandler", Kind: graph.KindFunction, File: "a.go", Line: 1},
		{ID: "a.go:g", Label: "barHandler", Kind: graph.KindFunction, File: "a.go", Line: 9},
		{ID: "b.go:h", Label: "baz", Kind: graph.KindFunction, File: "b.go", Line: 3},
	}
	doc.Edges = []graph.Edge{
		{From: "a.go:f", To: "a.go:g", Relation: graph.RelCalls, Confidence: graph.ConfExtracted},
		{From: "a.go:g", To: "b.go:h", Relation: graph.RelCalls, Confidence: graph.ConfExtracted},
	}
	return NewServer(store.NewIndex(doc))
}

// roundtrip feeds one request line through Serve and returns the decoded responses.
func roundtrip(t *testing.T, s *Server, lines ...string) []map[string]any {
	t.Helper()
	var out bytes.Buffer
	if err := s.Serve(strings.NewReader(strings.Join(lines, "\n")+"\n"), &out); err != nil {
		t.Fatalf("Serve: %v", err)
	}
	var resps []map[string]any
	for _, l := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if l == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Fatalf("decode response %q: %v", l, err)
		}
		resps = append(resps, m)
	}
	return resps
}

func TestMCP_InitializeEchoesAllowedVersion(t *testing.T) {
	r := roundtrip(t, testServer(), `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26"}}`)
	res := r[0]["result"].(map[string]any)
	if res["protocolVersion"] != "2025-03-26" {
		t.Fatalf("expected echoed allow-listed version 2025-03-26, got %v", res["protocolVersion"])
	}
}

func TestMCP_InitializeUnknownVersionFallsBackToLatest(t *testing.T) {
	r := roundtrip(t, testServer(), `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"1999-01-01"}}`)
	res := r[0]["result"].(map[string]any)
	if res["protocolVersion"] != LatestProtocolVersion {
		t.Fatalf("expected fallback to latest %q, got %v", LatestProtocolVersion, res["protocolVersion"])
	}
}

func TestMCP_ToolsListHasAllFour(t *testing.T) {
	r := roundtrip(t, testServer(), `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	res := r[0]["result"].(map[string]any)
	tools := res["tools"].([]any)
	got := map[string]bool{}
	for _, tr := range tools {
		got[tr.(map[string]any)["name"].(string)] = true
	}
	for _, want := range []string{"query_graph", "get_node", "get_neighbors", "shortest_path"} {
		if !got[want] {
			t.Fatalf("tools/list missing %q (got %v)", want, got)
		}
	}
}

func callText(t *testing.T, m map[string]any) string {
	t.Helper()
	res := m["result"].(map[string]any)
	content := res["content"].([]any)
	return content[0].(map[string]any)["text"].(string)
}

func TestMCP_CallQueryGraph(t *testing.T) {
	r := roundtrip(t, testServer(), `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"query_graph","arguments":{"question":"foo handler"}}}`)
	if !strings.Contains(callText(t, r[0]), "a.go:f") {
		t.Fatalf("query_graph did not return foo node:\n%s", callText(t, r[0]))
	}
}

func TestMCP_CallGetNode(t *testing.T) {
	r := roundtrip(t, testServer(), `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_node","arguments":{"id":"a.go:f"}}}`)
	if !strings.Contains(callText(t, r[0]), "a.go:f [function] fooHandler @ a.go:1") {
		t.Fatalf("get_node format wrong:\n%s", callText(t, r[0]))
	}
}

func TestMCP_CallGetNeighbors(t *testing.T) {
	r := roundtrip(t, testServer(), `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"get_neighbors","arguments":{"id":"a.go:g"}}}`)
	txt := callText(t, r[0])
	if !strings.Contains(txt, "a.go:f -calls-> a.go:g") || !strings.Contains(txt, "a.go:g -calls-> b.go:h") {
		t.Fatalf("get_neighbors missing in/out edges:\n%s", txt)
	}
}

func TestMCP_CallShortestPath(t *testing.T) {
	r := roundtrip(t, testServer(), `{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"shortest_path","arguments":{"from":"a.go:f","to":"b.go:h"}}}`)
	txt := callText(t, r[0])
	// BFS path f -> g -> h, ids one per line in path order.
	if !strings.Contains(txt, "a.go:f") || !strings.Contains(txt, "a.go:g") || !strings.Contains(txt, "b.go:h") {
		t.Fatalf("shortest_path missing path nodes:\n%s", txt)
	}
	if strings.Index(txt, "a.go:f") > strings.Index(txt, "b.go:h") {
		t.Fatalf("shortest_path order wrong (from must precede to):\n%s", txt)
	}
}

func TestMCP_NotificationGetsNoReply(t *testing.T) {
	var out bytes.Buffer
	s := testServer()
	if err := s.Serve(strings.NewReader(`{"jsonrpc":"2.0","method":"notifications/initialized"}`+"\n"), &out); err != nil {
		t.Fatalf("Serve: %v", err)
	}
	if strings.TrimSpace(out.String()) != "" {
		t.Fatalf("notification produced a reply: %q", out.String())
	}
}

func TestMCP_ParseErrorOnGarbage(t *testing.T) {
	r := roundtrip(t, testServer(), `{not json`)
	e := r[0]["error"].(map[string]any)
	if int(e["code"].(float64)) != -32700 {
		t.Fatalf("expected parse error -32700, got %v", e["code"])
	}
}
```

- [ ] **Step 2: Run the test, watch it fail to compile (no `mcp.go` yet)**

Run: `go test ./internal/mcp/ -run TestMCP 2>&1 | head`
Expected: build failure — `undefined: NewServer`, `undefined: LatestProtocolVersion`. Red.

- [ ] **Step 3: Implement `mcp.go` (validated prototype + four tools + version echo)**

Create `/Users/mylive/project/graffiti/graffiti/internal/mcp/mcp.go`:
```go
// Package mcp is graffiti's hand-rolled, dependency-free MCP server over stdio
// (see the Plan 4 §10 SPEC AMENDMENT). MCP's stdio transport is plain JSON-RPC
// 2.0 with NEWLINE-DELIMITED framing (one JSON object per line) — NOT the
// LSP-style Content-Length header framing. graffiti speaks three methods
// (initialize, tools/list, tools/call), tolerates notifications by ignoring
// them, and exposes four tools (query_graph, get_node, get_neighbors,
// shortest_path). encoding/json + bufio + io only: no deps, pure Go, offline.
package mcp

import (
	"bufio"
	"encoding/json"
	"io"
	"sort"

	"github.com/amazopic/graffiti/internal/graph"
	"github.com/amazopic/graffiti/internal/query"
	"github.com/amazopic/graffiti/internal/store"
)

// LatestProtocolVersion is the MCP revision graffiti prefers. On initialize the
// server ECHOES the client's requested version if it is in supportedVersions,
// otherwise it returns this latest (never blindly returns one hardcoded value).
const LatestProtocolVersion = "2025-06-18"

// supportedVersions is the small allow-list of MCP revisions graffiti will echo.
var supportedVersions = map[string]bool{
	"2025-06-18": true,
	"2025-03-26": true,
	"2024-11-05": true,
}

// --- JSON-RPC 2.0 wire types ---

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"` // absent => notification
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// --- MCP payload types ---

type initializeResult struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ServerInfo      serverInfo     `json:"serverInfo"`
}
type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
type toolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}
type listToolsResult struct {
	Tools []toolDef `json:"tools"`
}
type callToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}
type textContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
type callToolResult struct {
	Content []textContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// Server exposes the graph as MCP tools over stdio.
type Server struct {
	idx  *store.Index
	name string
}

// NewServer builds a Server over an in-memory Index.
func NewServer(idx *store.Index) *Server { return &Server{idx: idx, name: "graffiti"} }

func objSchema(props map[string]any, required ...string) map[string]any {
	req := make([]any, len(required))
	for i, r := range required {
		req[i] = r
	}
	return map[string]any{"type": "object", "properties": props, "required": req}
}

func strProp(desc string) map[string]any { return map[string]any{"type": "string", "description": desc} }

// tools is the static, deterministically-ordered tool catalog.
func (s *Server) tools() []toolDef {
	return []toolDef{
		{Name: "query_graph", Description: "LLM-free scoped subgraph retrieval over the code graph. Returns compact text.",
			InputSchema: objSchema(map[string]any{
				"question": strProp("natural-language question"),
				"budget":   map[string]any{"type": "integer", "description": "soft node-token budget (default 2000)"},
			}, "question")},
		{Name: "get_node", Description: "Return one node by id as 'id [kind] label @ file:line'.",
			InputSchema: objSchema(map[string]any{"id": strProp("node id")}, "id")},
		{Name: "get_neighbors", Description: "Return a node and its sorted in/out edges as compact text.",
			InputSchema: objSchema(map[string]any{"id": strProp("node id")}, "id")},
		{Name: "shortest_path", Description: "Deterministic BFS shortest path between two node ids (id-ordered frontier).",
			InputSchema: objSchema(map[string]any{"from": strProp("start node id"), "to": strProp("end node id")}, "from", "to")},
	}
}

// Serve runs the read/dispatch/write loop until r is exhausted. r and w are
// injectable so tests drive it without real stdin/stdout. Newline-framed.
func (s *Server) Serve(r io.Reader, w io.Writer) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024) // allow large tools/call results
	enc := json.NewEncoder(w)
	for sc.Scan() {
		line := sc.Bytes()
		if len(trimSpace(line)) == 0 {
			continue
		}
		var req rpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.writeErr(enc, nil, -32700, "parse error")
			continue
		}
		isNotification := len(req.ID) == 0
		resp := s.dispatch(req)
		if isNotification {
			continue // notifications get no reply
		}
		if err := enc.Encode(resp); err != nil {
			return err
		}
	}
	return sc.Err()
}

func (s *Server) dispatch(req rpcRequest) rpcResponse {
	switch req.Method {
	case "initialize":
		return s.ok(req.ID, initializeResult{
			ProtocolVersion: s.negotiateVersion(req.Params),
			Capabilities:    map[string]any{"tools": map[string]any{}},
			ServerInfo:      serverInfo{Name: s.name, Version: "0.4.0"},
		})
	case "tools/list":
		return s.ok(req.ID, listToolsResult{Tools: s.tools()})
	case "tools/call":
		return s.handleCall(req)
	default:
		return rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{Code: -32601, Message: "method not found: " + req.Method}}
	}
}

// negotiateVersion echoes the client's requested protocolVersion when it is in
// the allow-list, else returns the server's latest (never blindly hardcoded).
func (s *Server) negotiateVersion(params json.RawMessage) string {
	var p struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	_ = json.Unmarshal(params, &p)
	if supportedVersions[p.ProtocolVersion] {
		return p.ProtocolVersion
	}
	return LatestProtocolVersion
}

func (s *Server) handleCall(req rpcRequest) rpcResponse {
	var p callToolParams
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return s.toolErr(req.ID, "invalid params")
	}
	switch p.Name {
	case "query_graph":
		var a struct {
			Question string `json:"question"`
			Budget   int    `json:"budget"`
		}
		_ = json.Unmarshal(p.Arguments, &a)
		return s.toolText(req.ID, query.Query(s.idx, a.Question, a.Budget))
	case "get_node":
		var a struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(p.Arguments, &a)
		n, ok := s.idx.Node(a.ID)
		if !ok {
			return s.toolErr(req.ID, "node not found: "+a.ID)
		}
		return s.toolText(req.ID, query.FormatNode(n))
	case "get_neighbors":
		var a struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(p.Arguments, &a)
		n, ok := s.idx.Node(a.ID)
		if !ok {
			return s.toolErr(req.ID, "node not found: "+a.ID)
		}
		return s.toolText(req.ID, s.neighborsText(n))
	case "shortest_path":
		var a struct {
			From string `json:"from"`
			To   string `json:"to"`
		}
		_ = json.Unmarshal(p.Arguments, &a)
		path, ok := s.shortestPath(a.From, a.To)
		if !ok {
			return s.toolErr(req.ID, "no path from "+a.From+" to "+a.To)
		}
		out := ""
		for _, id := range path {
			out += id + "\n"
		}
		return s.toolText(req.ID, out)
	default:
		return s.toolErr(req.ID, "unknown tool: "+p.Name)
	}
}

// neighborsText renders a node line followed by its sorted in+out edges.
func (s *Server) neighborsText(n graph.Node) string {
	out := query.FormatNode(n) + "\nEDGES\n"
	for _, e := range s.idx.In(n.ID) {
		out += query.FormatEdge(e) + "\n"
	}
	for _, e := range s.idx.Out(n.ID) {
		out += query.FormatEdge(e) + "\n"
	}
	return out
}

// shortestPath is a DETERMINISTIC BFS over a directed-as-undirected adjacency
// (follow both out and in edges) with an ID-ORDERED frontier: at each node the
// neighbor ids are gathered, de-duped, and SORTED before enqueueing, so the
// discovered path is byte-identical for the same graph (spec §14).
func (s *Server) shortestPath(from, to string) ([]string, bool) {
	if _, ok := s.idx.Node(from); !ok {
		return nil, false
	}
	if _, ok := s.idx.Node(to); !ok {
		return nil, false
	}
	if from == to {
		return []string{from}, true
	}
	prev := map[string]string{from: ""}
	queue := []string{from}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		// Gather neighbor ids (both directions), dedupe, sort => id-ordered frontier.
		seen := map[string]bool{}
		var nbrs []string
		for _, e := range s.idx.Out(cur) {
			if !seen[e.To] {
				seen[e.To] = true
				nbrs = append(nbrs, e.To)
			}
		}
		for _, e := range s.idx.In(cur) {
			if !seen[e.From] {
				seen[e.From] = true
				nbrs = append(nbrs, e.From)
			}
		}
		sort.Strings(nbrs)
		for _, nb := range nbrs {
			if _, ok := prev[nb]; ok {
				continue
			}
			prev[nb] = cur
			if nb == to {
				return reconstruct(prev, from, to), true
			}
			queue = append(queue, nb)
		}
	}
	return nil, false
}

func reconstruct(prev map[string]string, from, to string) []string {
	var rev []string
	for cur := to; cur != ""; cur = prev[cur] {
		rev = append(rev, cur)
		if cur == from {
			break
		}
	}
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	return rev
}

func (s *Server) ok(id json.RawMessage, result any) rpcResponse {
	return rpcResponse{JSONRPC: "2.0", ID: id, Result: result}
}
func (s *Server) toolText(id json.RawMessage, text string) rpcResponse {
	return s.ok(id, callToolResult{Content: []textContent{{Type: "text", Text: text}}})
}
func (s *Server) toolErr(id json.RawMessage, msg string) rpcResponse {
	return s.ok(id, callToolResult{Content: []textContent{{Type: "text", Text: msg}}, IsError: true})
}
func (s *Server) writeErr(enc *json.Encoder, id json.RawMessage, code int, msg string) {
	_ = enc.Encode(rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}})
}

func trimSpace(b []byte) []byte {
	i, j := 0, len(b)
	for i < j && (b[i] == ' ' || b[i] == '\t' || b[i] == '\r' || b[i] == '\n') {
		i++
	}
	for j > i && (b[j-1] == ' ' || b[j-1] == '\t' || b[j-1] == '\r' || b[j-1] == '\n') {
		j--
	}
	return b[i:j]
}
```

> `query.FormatNode`/`query.FormatEdge`/`query.Query` are the exported entry points added in Task 2 Step 3; `mcp` imports `internal/query` and `internal/store`. The `shortest_path` BFS follows edges in **both** directions (the code graph is directed, but "is there a path between these two symbols" is a reachability question over the undirected projection); the **id-ordered frontier** (sort neighbor ids before enqueue) makes the discovered shortest path deterministic when multiple equal-length paths exist.

- [ ] **Step 4: Run the MCP tests — green**

Run: `go test ./internal/mcp/ -run TestMCP -v`
Expected: PASS — `TestMCP_InitializeEchoesAllowedVersion`, `TestMCP_InitializeUnknownVersionFallsBackToLatest`, `TestMCP_ToolsListHasAllFour`, `TestMCP_CallQueryGraph`, `TestMCP_CallGetNode`, `TestMCP_CallGetNeighbors`, `TestMCP_CallShortestPath`, `TestMCP_NotificationGetsNoReply`, `TestMCP_ParseErrorOnGarbage`.

- [ ] **Step 5: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/mcp/mcp.go internal/mcp/mcp_test.go
git commit -m "feat(mcp): hand-rolled stdio JSON-RPC 2.0 server, 4 tools, protocolVersion echo

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: CLI — `graffiti query "<q>"` + `graffiti serve`, updated usage

**Files:**
- Modify: `cmd/graffiti/main.go`
- Modify: `cmd/graffiti/main_test.go`

> `query` loads `.graffiti/map.json` from the given root (default `.`), builds the `store.Index`, runs `query.Query`, and prints the subgraph text to stdout. `serve` builds the same `Index` and runs `mcp.Server.Serve(os.Stdin, os.Stdout)`. Both share a small `loadIndex(root)` helper. `run` stays the testable entry point (returns the exit code); `serve` takes an explicit `io.Reader`/`io.Writer` so the smoke test drives it without real stdin/stdout.

- [ ] **Step 1: Add failing CLI tests for `query` and `serve`**

In `/Users/mylive/project/graffiti/graffiti/cmd/graffiti/main_test.go`, append:
```go
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
```

> `bytes`, `os`, `path/filepath`, `strings`, `testing` are already imported by `main_test.go`; no import changes needed.

- [ ] **Step 2: Run the CLI tests, watch them fail (no `query`/`serve` dispatch yet)**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./cmd/graffiti/ -run 'TestRun_Query|TestRun_Serve' 2>&1 | head`
Expected: build failure / FAIL — `undefined: serve`, and `run` returns `2` (unknown command) for `query`. Red.

- [ ] **Step 3: Add `query` + `serve` dispatch and the `loadIndex` helper to `main.go`**

In `/Users/mylive/project/graffiti/graffiti/cmd/graffiti/main.go`, extend the import block and `switch`, and add the helpers. The file becomes:
```go
// Command graffiti turns a code repository into a queryable directed knowledge graph.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/amazopic/graffiti/internal/app"
	"github.com/amazopic/graffiti/internal/mcp"
	"github.com/amazopic/graffiti/internal/query"
	"github.com/amazopic/graffiti/internal/store"
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
	case "query":
		if len(args) < 3 {
			fmt.Fprintln(stderr, "graffiti: query requires a question")
			usage(stderr)
			return 2
		}
		question := args[2]
		root := "."
		if len(args) >= 4 {
			root = args[3]
		}
		return runQuery(root, question, stdout, stderr)
	case "serve":
		root := "."
		if len(args) >= 3 {
			root = args[2]
		}
		return serve(root, os.Stdin, stdout, stderr)
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
	fmt.Fprintln(w, "  .                 build the map for the current repo")
	fmt.Fprintln(w, "  build <path>      build the map for <path> (default .)")
	fmt.Fprintln(w, "  query \"<q>\" [path]  LLM-free scoped subgraph retrieval (soft 2000-token node budget)")
	fmt.Fprintln(w, "  serve [path]      MCP server over stdio (JSON-RPC 2.0)")
}

func runBuild(root string, stdout, stderr io.Writer) int {
	generatedAt := time.Now().UTC().Format(time.RFC3339)
	stats, err := app.Build(root, generatedAt)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: build failed: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "✓ Done. 0 API calls, $0.  %d files → %d nodes, %d edges, %d communities.\n",
		stats.Files, stats.Nodes, stats.Edges, stats.Communities)
	fmt.Fprintln(stdout, "  The 3 most interesting questions your map can answer:")
	for i, q := range stats.Questions {
		fmt.Fprintf(stdout, "    %d) %s\n", i+1, q)
	}
	return 0
}

// loadIndex loads <root>/.graffiti/map.json and builds the read-side Index.
func loadIndex(root string) (*store.Index, error) {
	path := mapPath(root)
	doc, err := store.Load(path)
	if err != nil {
		return nil, err
	}
	return store.NewIndex(doc), nil
}

func mapPath(root string) string {
	return filepath.Join(root, ".graffiti", "map.json")
}

func runQuery(root, question string, stdout, stderr io.Writer) int {
	idx, err := loadIndex(root)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v (run `graffiti build %s` first)\n", err, root)
		return 1
	}
	fmt.Fprint(stdout, query.Query(idx, question, query.DefaultTokenBudget))
	return 0
}

// serve runs the MCP stdio server. r/w/errW are injectable for tests; main wires
// os.Stdin/os.Stdout. Returns the exit code.
func serve(root string, r io.Reader, w, errW io.Writer) int {
	idx, err := loadIndex(root)
	if err != nil {
		fmt.Fprintf(errW, "graffiti: %v (run `graffiti build %s` first)\n", err, root)
		return 1
	}
	if err := mcp.NewServer(idx).Serve(r, w); err != nil {
		fmt.Fprintf(errW, "graffiti: serve: %v\n", err)
		return 1
	}
	return 0
}
```

> `mapPath` uses `filepath.Join(root, ".graffiti", "map.json")` so the read path mirrors `render.WriteMapJSON`'s write path (which writes to `filepath.Join(root, ".graffiti", "map.json")`) and stays correct on Windows. The import block above adds `path/filepath` accordingly; `cmd/graffiti` no longer needs `strings`.

- [ ] **Step 4: Run the CLI tests — green**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./cmd/graffiti/ -run 'TestRun' -v`
Expected: PASS — the existing `TestRun_NoArgs_PrintsUsage`, `TestRun_UnknownCommand_Errors`, `TestRun_BuildPrintsSuccessLine`, plus the new `TestRun_QueryPrintsSubgraph`, `TestRun_QueryMissingMap_Errors`, `TestRun_ServeHandlesInitialize`.

- [ ] **Step 5: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add cmd/graffiti/main.go cmd/graffiti/main_test.go
git commit -m "feat(cli): add `graffiti query` and `graffiti serve` (MCP stdio); update usage

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: Query golden on the `gorepo` fixture (end-to-end, build → query)

**Files:**
- Create: `internal/query/golden_test.go`
- Create (via UPDATE_GOLDEN): `testdata/golden/gorepo.query.txt`

> This builds the real `gorepo` fixture end-to-end (the same fixture Plans 1–3 golden against), loads its `map.json`, runs a fixed query, and locks the output byte-for-byte. **Query output has no `generated_at` and no `root`-dependent bytes** (node ids/labels/files/lines and edge relations only) — so unlike the map.json/MAP.md/map.html goldens there is **nothing time-dependent to strip**. The golden test lives in `internal/query` (not `internal/app`) to keep the read path's tests with the read path; it calls `app.Build` to materialize the fixture artifact, then `store.Load` + `query.Query`.

- [ ] **Step 1: Add the golden test**

Create `/Users/mylive/project/graffiti/graffiti/internal/query/golden_test.go`:
```go
package query_test // external test package: imports app (which links parse) — needs the build tags

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amazopic/graffiti/internal/app"
	"github.com/amazopic/graffiti/internal/query"
	"github.com/amazopic/graffiti/internal/store"
)

const fixtureGenAt = "2026-06-16T00:00:00Z"
const goldenQuestion = "where is the graph built and rendered"

// copyTree mirrors the helper Plans 1–3 use in app tests.
func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		s, d := filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())
		if e.IsDir() {
			copyTree(t, s, d)
			continue
		}
		b, err := os.ReadFile(s)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(d, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func runFixtureQuery(t *testing.T) string {
	t.Helper()
	src := filepath.Join("..", "..", "testdata", "fixtures", "gorepo")
	dst := t.TempDir()
	copyTree(t, src, dst)
	if _, err := app.Build(dst, fixtureGenAt); err != nil {
		t.Fatalf("build fixture: %v", err)
	}
	doc, err := store.Load(filepath.Join(dst, ".graffiti", "map.json"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return query.Query(store.NewIndex(doc), goldenQuestion, query.DefaultTokenBudget)
}

func TestGolden_GoRepoQuery(t *testing.T) {
	got := runFixtureQuery(t)
	path := filepath.Join("..", "..", "testdata", "golden", "gorepo.query.txt")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Log("query golden updated")
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden (run UPDATE_GOLDEN=1 to create): %v", err)
	}
	if got != string(want) {
		t.Fatalf("query output differs from golden:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestGolden_GoRepoQuery_Deterministic(t *testing.T) {
	if runFixtureQuery(t) != runFixtureQuery(t) {
		t.Fatal("two fixture queries not byte-identical")
	}
}
```

- [ ] **Step 2: Run the determinism check first (the real gate) before freezing the golden**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/query/ -run TestGolden_GoRepoQuery_Deterministic -v
```
Expected: PASS. Do **not** freeze the golden until this is green.

- [ ] **Step 3: Generate the golden from real output**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
UPDATE_GOLDEN=1 go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/query/ -run TestGolden_GoRepoQuery -v
```
Expected: PASS with log `query golden updated`; `testdata/golden/gorepo.query.txt` now exists. Open it once: it starts with `NODES`, lists `id [kind] label @ file:line` lines sorted by id, then an `EDGES` block of `from -relation-> to (CONFIDENCE)` lines sorted by `(from,to,relation,confidence)`. No timestamp anywhere.

- [ ] **Step 4: Re-run the golden test against the frozen file (no UPDATE_GOLDEN)**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/query/ -run TestGolden -v
```
Expected: PASS — `TestGolden_GoRepoQuery`, `TestGolden_GoRepoQuery_Deterministic`.

- [ ] **Step 5: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/query/golden_test.go testdata/golden/gorepo.query.txt
git commit -m "test(query): byte-exact query golden on gorepo fixture (no generated_at to strip)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: Full-suite verification, vet, size guard, dependency check, and end-to-end smoke

**Files:** none (verification only)

- [ ] **Step 1: Run the full test suite via the Makefile**

Run: `make test`
Expected: all packages PASS (`ok` for cmd/graffiti, internal/analyze, internal/app, internal/build, internal/cache, internal/cluster, internal/graph, internal/layout, internal/mcp, internal/parse, internal/query, internal/render, internal/scan, internal/store, schemaval). Plan 1–3 goldens (`gorepo.map.json`, `gorepo.MAP.md`, `gorepo.map.html.strip`) are unaffected — the read path adds no build-pipeline change.

- [ ] **Step 2: Run vet**

Run: `make vet`
Expected: no output (clean).

- [ ] **Step 3: Confirm NO new dependency was added (the §10 amendment's whole point)**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
go mod tidy
git diff --stat go.mod go.sum
go list -deps ./internal/mcp ./internal/query ./internal/store | grep -vE '^(internal/|github.com/amazopic/graffiti)' | grep -E '^(github.com|golang.org)' | sort -u
```
Expected: `go.mod`/`go.sum` unchanged (no diff). The dep list shows only `golang.org/x/text/...` (already present via `internal/graph`'s `NormalizeID`) and stdlib — **no MCP SDK, no jwt/oauth2/http-auth packages**. This is the binary-proof of the §10 hand-rolled-MCP amendment.

- [ ] **Step 4: Build the binary and smoke-test `query` end-to-end**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
make build
rm -rf /tmp/graffiti-p4-smoke && cp -r testdata/fixtures/gorepo /tmp/graffiti-p4-smoke
./graffiti build /tmp/graffiti-p4-smoke >/dev/null
./graffiti query "where is the graph built" /tmp/graffiti-p4-smoke | head -20
```
Expected: a `NODES` block of relevant `id [kind] label @ file:line` lines followed by an `EDGES` block — no spinner, no network, exit 0.

- [ ] **Step 5: Smoke-test `serve` (one initialize + tools/list round-trip over a pipe)**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
printf '%s\n%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
  | ./graffiti serve /tmp/graffiti-p4-smoke
```
Expected: two newline-delimited JSON objects. The first has `"result":{"protocolVersion":"2025-06-18",...,"serverInfo":{"name":"graffiti",...}}` (version echoed). The second lists four tools: `query_graph`, `get_node`, `get_neighbors`, `shortest_path`. No `Content-Length` headers (newline framing). Exit 0 (EOF on stdin ends the loop).

- [ ] **Step 6: Confirm query determinism at the binary level (two queries byte-identical)**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
./graffiti query "where is auth handled" /tmp/graffiti-p4-smoke > /tmp/q1.txt
./graffiti query "where is auth handled" /tmp/graffiti-p4-smoke > /tmp/q2.txt
diff /tmp/q1.txt /tmp/q2.txt && echo "query deterministic"
```
Expected: `query deterministic` (empty diff). Query output has no `generated_at`, so it is fully byte-stable for a fixed graph+question.

- [ ] **Step 7: Cross-compile + size guard (the embedded read path adds no weight)**

Run: `make xcompile`
Expected: `size-guard: all binaries within limit OK` — `query`/`serve` add only stdlib code (no SDK), so the binary stays ~8MB, well under the 16MB limit.

- [ ] **Step 8: Final tidy + finish the branch**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
go mod tidy   # no-op: Plan 4 adds no dependencies
make test
git diff --quiet go.mod go.sum || (git add go.mod go.sum && git commit -m "build: go mod tidy (no-op for Plan 4)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>")
```
Expected: `make test` prints `ok` for every package; `go mod tidy` changes nothing (Plan 4 is stdlib + the already-present `golang.org/x/text`).

Then use superpowers:finishing-a-development-branch to merge `plan4-query-mcp` into `main` (or open a PR), per the flow Plans 1–3 used.

---

## Self-Review

**1. Spec coverage (Plan 4 scope only):**
- §5 pipeline read path (`query` separate from build; `mcp` exposes `query`/`get_node`/`neighbors`/`path` over stdio): `internal/query` + `internal/mcp` consume `store.Index`, never the build pipeline; all four tools shipped → Tasks 2, 3, 4. ✓
- §7 LLM-free query path: (1) load `map.json`, rebuild the graph (`store.Load` + `NewIndex`); (2) tokenize + IDF score over labels/kinds; (3) seed + BFS-expand under a token budget (default 2000); (4) serialize to compact text (nodes with `file:line`, edges with relation+confidence). No inference call — the host model reasons over the text → Tasks 1, 2. ✓
- §11 CLI: `graffiti query "<q>"` and `graffiti serve` added; usage updated; the build success line is unchanged → Task 4. ✓
- §10 transport: **amended** — hand-rolled minimal JSON-RPC 2.0 over stdio (stdlib only), isolated behind `internal/mcp.Serve`, official SDK a documented drop-in fallback (§10 SPEC AMENDMENT + Open Issue #3). ✓ (capability) / ⚠ (mechanism — requires owner ratification, mirroring Plan 1).
- §14 determinism: byte-identical query output for the same graph+question (no `generated_at`); no Go-map iteration feeds emitted order (every map consumed via sorted key lists; `store.Index` adjacency pre-sorted; `shortest_path` id-ordered frontier); golden + determinism tests → Tasks 1, 2, 3, 5. ✓
- **Explicitly out of scope (correctly absent):** `graffiti update` (incremental rebuild, §11 — a build-side plan), `graffiti init` (§9 integration), and the entire §16 workspace/`--workspace` federation. Plan 4 is single-project query + MCP only. The `query` and `serve` commands accept a `[path]` so they are workspace-ready without `--workspace` semantics.

**2. Determinism guarantees (spec §14), enforced and tested:**
- `store.Index`: `ids` sorted ascending; every `out`/`in` adjacency slice sorted by `(relation, other-id, confidence)` with a total order; `NewIndex` builds identical structure across runs (verified by `TestIndex_AdjacencySortedAndDeterministic`).
- `query`: IDF tables built by iterating `idx.IDs()` (sorted), scores sorted `(score desc, id asc)`, seeds added in that order, neighbors pulled from the pre-sorted adjacency, NODES sorted by id, EDGES de-duped and sorted `(from,to,relation,confidence)`. Soft node budget is a pure function of node text. Verified by `TestQuery_Deterministic` (×5), `TestQuery_SeedTieBreakByID`, `TestQuery_SerializeOrderAndEdges`, and the binary-level diff in Task 6.
- `mcp`: tool catalog is a static ordered slice; `shortest_path` BFS sorts the neighbor frontier by id before enqueue, so equal-length paths resolve deterministically. Verified by `TestMCP_ToolsListHasAllFour`, `TestMCP_CallShortestPath`.
- Query output has **no `generated_at` and no `root`-relative bytes** (ids/labels/files/lines/relations only), so the golden is locked byte-exact with nothing to strip (Task 5).

**3. Type consistency with the merged Plan-1/Plan-2/Plan-3 code (checked against the real sources):**
- `graph.Node{ID,Label,Kind,File,Line,Community}`, `graph.Edge{From,To,Relation,Confidence}`, `graph.Community{ID,Label,Members}`, `graph.Document{Version,GeneratedAt,Root,Nodes,Edges,Communities}` with their json tags (`internal/graph/graph.go`) — `store.Load` unmarshals straight into `graph.Document`, lossless mirror of `render.WriteMapJSON`'s `orderedDocument`. ✓
- `graph.Kind`/`Relation`/`Confidence` are string-backed; `string(...)` used where plain strings are emitted. `graph.NormalizeID`/`graph.NewDocument`/`graph.SchemaVersion` referenced exactly as defined. The query tokenizer's NFC normalization reuses `golang.org/x/text/unicode/norm` (already a `graph` dep). ✓
- `store.Load(path string) (*graph.Document, error)`, `store.NewIndex(*graph.Document) *Index`, `Index.Node/IDs/Out/In/Len` — Task 1; consumed by `query` and `mcp`. ✓
- `query.Query(*store.Index, string, int) string`, `query.DefaultTokenBudget`, exported `query.FormatNode`/`FormatEdge`/`Serialize` — Task 2; reused by `mcp`. ✓
- `mcp.NewServer(*store.Index) *Server`, `Server.Serve(io.Reader, io.Writer) error`, `mcp.LatestProtocolVersion` — Task 3; wired by the CLI. ✓
- `cmd/graffiti` `run([]string, io.Writer, io.Writer) int` unchanged signature; new `runQuery`/`serve`/`loadIndex` helpers; `serve(root, io.Reader, io.Writer, io.Writer) int` injectable for tests; `app.Build` and the build pipeline untouched (all prior goldens byte-identical). Module path `github.com/amazopic/graffiti` in every import; grammar-subset tags applied to every `cmd`/`app` command. ✓

**4. Prototype evidence (load-bearing):** the deterministic LLM-free query (IDF + seed + BFS + soft node budget + compact serialize) and the hand-rolled minimal MCP stdio server were prototyped end-to-end in a scratch module with passing tests (`/tmp/p4proto/{query,mcp}.go` + `proto_test.go`): determinism ×5, budget respected, relevance, deterministic tie-breaks, no-map-iteration-in-output for query; newline-delimited JSON-RPC 2.0 `initialize`/`tools/list`/`tools/call`, notification-no-reply, parse-error for MCP. The task code is that validated prototype with the mirror `*Graph` replaced by `*store.Index`, the tokenizer NFC-aligned to `graph.NormalizeID`, three more tools added, and `initialize` upgraded from a hardcoded constant to an allow-list version echo. The budget-covers-nodes-only decision and the newline-not-Content-Length framing were both validated in the prototype and are restated in code comments.

---

## Open Issues (require owner attention)

1. **Token budget is a SOFT NODE budget; edge text is not counted.** The validated query charges `estimateTokens(formatNode(n)) = len/4` per *selected node* and emits all edges between selected nodes for free. This keeps node selection independent of edge-emission order (a determinism requirement) and matches the prototype, but means the serialized text can exceed the nominal budget when a dense subgraph is selected (many edges among few nodes). The `len/4` estimator is a deliberate zero-dependency rule of thumb, not a real tokenizer, so the budget is approximate either way. **Owner to confirm soft-node-budget semantics are acceptable for v1** (alternatives: also charge per edge line — at the cost of order-dependence — or switch to a tokenizer dep, which violates the zero-dep ethos). The behavior is documented in `query.go`, the CLI `usage`, and the `query_graph` tool description.

2. **Query tokenizer normalization is aligned with `graph.NormalizeID` but not identical.** `NormalizeID` NFC-normalizes, casefolds, collapses non-word runs to `-`, and trims; the query tokenizer NFC-normalizes, lowercases, splits on non-word boundaries, **and additionally splits camelCase/snake_case** (so `parseFile` → `parse`,`file` to match natural-language questions). This intentional asymmetry (the slug never camel-splits; the query tokenizer does) is what makes a question like "where is the parser" match a `parseFile` node. The shared NFC step ensures Unicode folding agrees. **Owner to confirm the camelCase-split-on-query-side-only policy.** If exact symmetry is later wanted, factor a shared `graph.tokenize` and have `NormalizeID` join its tokens with `-` — a refactor that cannot change existing node ids (the slug output is unchanged) but should be done deliberately and re-golden the build artifacts.

3. **§10 hand-rolled-MCP amendment requires owner ratification (Task 3 gate).** The §10 SPEC AMENDMENT substitutes a hand-rolled stdlib JSON-RPC server for the spec's "Go MCP SDK over stdio," isolated behind `internal/mcp.Serve` (same isolate-and-pivot pattern Plan 1 used for the parser). Reasoning: the official SDK pulls ~10 transitive deps (jwt/oauth2/http-auth) that a stdio-only, keyless, offline server needs none of, against the single-static-pure-Go/minimal-deps ethos. **Owner to ratify option (a) — adopt the hand-rolled server; SDK is a documented drop-in fallback.** Tasks 1, 2, 4 (query CLI), 5 do not depend on this; only Task 3 (the MCP server package) and the `serve` command's MCP behavior do. (Mirrors Plan 1's Open Issue #1 gate.)

4. **`get_neighbors`/`get_node` return plain compact text, not a structured tool result.** v1 returns the same `id [kind] label @ file:line` / `from -relation-> to (CONFIDENCE)` lines the query path emits, as a single `text` content block — deliberately, so the host model reads one consistent format across all tools and so the formatting is single-sourced through `query.FormatNode`/`FormatEdge`. A future plan could add a structured JSON content block (machine-parseable node/edge objects) alongside the text for programmatic MCP clients. **Owner to confirm text-only tool results are acceptable for v1** (the spec only requires the four tools be exposed, not their result schema).

---
