# graffiti Plan 9 — Force-graph viewer (replaces Districts)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the baked treemap "Districts" `map.html` with the **force-directed node-link graph** the user iterated on and approved: nodes (sphere-shaded, sized by degree, colored by sector/package), edges, an in-browser physics layout, a **2D/3D toggle**, code **categories** (client / tests / external libs), **sector zones**, a **hierarchical left panel** (project → directory → file tree with checkboxes + sizes, resizable), **fit-to-window**, and zoom/pan/drag/hover. Add a **workspace** view (`graffiti workspace render` → `workspace.html`) where the tree's top level is the federated projects and cross-project links are drawn (spec §16.7).

**Architecture:** The prototype already splits **data** (a JSON island) from **renderer** (vanilla Canvas2D JS). So the port keeps graffiti's proven `internal/render` machinery — single self-contained offline file, CSP `meta` with sha256 over the inlined bodies, `generated_at` outside every hashed body, go:embed'd viewer assets, `</script` escape, hidden a11y mirror — and only swaps **(a)** the data island shape (a compact columnar graph: per-node label/kind/file/line/degree/category + edge index pairs) and **(b)** the embedded `viewer/app.js` + `app.css` (the validated force-graph renderer). The browser computes layout (force sim), so the file stays **byte-deterministic in its data + assets** (golden-testable) while node *positions* are runtime-only. The treemap `internal/layout` is removed. A workspace render builds a **combined** island from all member `map.json` files (alias-prefixed file paths) plus the overlay's cross-edges — reusing the Plan-7 model.

**Editability (why this matters):** after the port the viewer remains plain, embedded source — edit `internal/render/viewer/app.js`/`app.css`, `make build`, reopen `map.html`. The only added step vs the prototype is refreshing the `map.html` strip-golden when the assets change (one command).

**Tech Stack:** Go 1.26 (stdlib: `encoding/json`, `regexp`, `crypto/sha256`, `embed`), vanilla Canvas2D JS. No new dependencies. The force-graph renderer is **pseudo-3D in Canvas2D** (sphere gradients + hover elevation), not WebGL — staying offline/single-file/CSP-safe (this is the optional "showcase" view §4 allowed).

**Reference (source of the JS/CSS):** the validated prototype generator at `/tmp/gen_graph.py` — its `<style>…</style>` body is the new `app.css`, and its `<script>…</script>` body (the renderer, NOT the `JSON.parse` data line's source) is the new `app.js`. The prototype's data object `D = {label,kind,file,line,deg,cat,edges,root}` is exactly the island Go will emit.

## File structure

```
internal/render/columnar.go     → rewrite: graphIsland(doc) building {label,kind,file,line,deg,cat,edges,root}
internal/render/island_test.go  island encoder tests (deterministic/parses/category/escaping)   [new name ok]
internal/render/viewer/app.css  ← prototype <style> body (force-graph theme)
internal/render/viewer/app.js   ← prototype <script> body (reads #graffiti-data)
internal/render/maphtml.go      WriteMapHTML(doc, an, root); new skeleton + island + CSP + a11y-by-sector
internal/render/maphtml_test.go updated CSP/self-contained/island/escape/determinism tests
internal/render/workspace_html.go  RenderWorkspaceHTML(combined island) + WriteWorkspaceHTML            [new]
internal/app/app.go             drop layout.Layout/scene; call render.WriteMapHTML(doc, an, absRoot)
internal/layout/                DELETE (treemap obsolete; layout is in-browser now)
cmd/graffiti/main.go            add `workspace render` subcommand
testdata/golden/gorepo.map.html.strip   regenerated
README.md, docs/.../2026-06-14-graffiti-design.md  §8/§16.7 amendments
```

---

## Task 1: graph data-island encoder

Replace the treemap columnar encoder with a graph island matching the prototype's `D` object.

**Files:** rewrite `internal/render/columnar.go`; test `internal/render/columnar_test.go`.

- [ ] **Step 1: Write the failing tests** (`internal/render/columnar_test.go`, replacing the old scene-columnar tests):

```go
package render

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

func sampleDoc() *graph.Document {
	d := graph.NewDocument("demo")
	d.GeneratedAt = "2026-06-17T00:00:00Z"
	d.Nodes = []graph.Node{
		{ID: "a.go:a.go", Label: "a.go", Kind: graph.KindFile, File: "a.go", Line: 1, Community: 0},
		{ID: "a.go:foo", Label: "Foo", Kind: graph.KindFunction, File: "a.go", Line: 3, Community: 0},
		{ID: "a_test.go:t", Label: "TestFoo", Kind: graph.KindFunction, File: "a_test.go", Line: 5, Community: 0},
		{ID: "module:fmt", Label: "fmt", Kind: graph.KindModule, File: "a.go", Line: 1, Community: 0},
	}
	d.Edges = []graph.Edge{
		{From: "a.go:a.go", To: "a.go:foo", Relation: graph.RelContains, Confidence: graph.ConfExtracted},
		{From: "a.go:foo", To: "module:fmt", Relation: graph.RelImports, Confidence: graph.ConfExtracted},
	}
	return d
}

func TestGraphIsland_ShapeAndCategories(t *testing.T) {
	is := graphIsland(sampleDoc())
	if len(is.Label) != 4 || len(is.Kind) != 4 || len(is.File) != 4 || len(is.Deg) != 4 || len(is.Cat) != 4 {
		t.Fatalf("columnar arrays must all be len 4")
	}
	// nodes are emitted in sorted-id order (doc already sorted). Find by label.
	idxOf := map[string]int{}
	for i, l := range is.Label {
		idxOf[l] = i
	}
	if is.Cat[idxOf["fmt"]] != 2 {
		t.Errorf("module fmt must be external (2)")
	}
	if is.Cat[idxOf["TestFoo"]] != 1 {
		t.Errorf("TestFoo must be test (1)")
	}
	if is.Cat[idxOf["Foo"]] != 0 {
		t.Errorf("Foo must be client (0)")
	}
	// degree: foo touches file + fmt = 2
	if is.Deg[idxOf["Foo"]] != 2 {
		t.Errorf("Foo degree = %d, want 2", is.Deg[idxOf["Foo"]])
	}
	// edges are index pairs into the node arrays
	if len(is.Edges) != 2 {
		t.Fatalf("want 2 edges, got %d", len(is.Edges))
	}
	for _, e := range is.Edges {
		if e[0] < 0 || e[0] >= 4 || e[1] < 0 || e[1] >= 4 {
			t.Fatalf("edge index out of range: %v", e)
		}
	}
	if is.Root != "demo" {
		t.Errorf("root = %q", is.Root)
	}
}

func TestGraphIsland_Deterministic(t *testing.T) {
	a, _ := json.Marshal(graphIsland(sampleDoc()))
	b, _ := json.Marshal(graphIsland(sampleDoc()))
	if string(a) != string(b) {
		t.Fatal("island marshal not deterministic")
	}
}

func TestCategoryHeuristics(t *testing.T) {
	cases := []struct {
		kind graph.Kind
		file, label string
		want int
	}{
		{graph.KindModule, "a.go", "fmt", 2},
		{graph.KindFunction, "a_test.go", "TestX", 1},
		{graph.KindFunction, "pkg/foo_test.go", "helper", 1},
		{graph.KindFunction, "src/app.test.ts", "x", 1},
		{graph.KindFunction, "tests/conftest.py", "fix", 1},
		{graph.KindFunction, "main.go", "main", 0},
		{graph.KindClass, "model.ts", "User", 0},
	}
	for _, c := range cases {
		if got := categoryOf(graph.Node{Kind: c.kind, File: c.file, Label: c.label}); got != c.want {
			t.Errorf("categoryOf(%q,%q,%v)=%d want %d", c.file, c.label, c.kind, got, c.want)
		}
	}
}
```

- [ ] **Step 2: Run to verify they fail** — `go test ./internal/render/ -run 'Island|Category' -v` → FAIL (undefined).

- [ ] **Step 3: Rewrite `internal/render/columnar.go`**:

```go
package render

import (
	"regexp"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

// graphData is the compact columnar graph island consumed by viewer/app.js.
// Parallel arrays index the same node; edges are [fromIdx,toIdx] pairs into them.
// (Sector color + layout are derived in the browser from file paths, so they are
// intentionally absent here.)
type graphData struct {
	Label []string `json:"label"`
	Kind  []string `json:"kind"`
	File  []string `json:"file"`
	Line  []int    `json:"line"`
	Deg   []int    `json:"deg"`
	Cat   []int    `json:"cat"`   // 0=client 1=test 2=external
	Edges [][2]int `json:"edges"`
	Root  string   `json:"root"`
}

// testFile matches common test-file conventions across graffiti's languages.
var testFile = regexp.MustCompile(`(^|/)(tests?)(/|_)|_test\.|\.test\.|\.spec\.|(^|/)test_|Tests?\.(java|kt|php|rs|py)$`)

// categoryOf classifies a node: 2=external (imported module), 1=test, 0=client.
func categoryOf(n graph.Node) int {
	if n.Kind == graph.KindModule {
		return 2
	}
	if testFile.MatchString(n.File) || (len(n.Label) >= 4 && n.Label[:4] == "Test") {
		return 1
	}
	return 0
}

// graphIsland builds the columnar island from a (sorted) Document. doc.Nodes is
// already id-sorted by build.Assemble, so node order — and thus edge indices —
// are deterministic.
func graphIsland(doc *graph.Document) graphData {
	idx := make(map[string]int, len(doc.Nodes))
	d := graphData{Root: doc.Root}
	for i, n := range doc.Nodes {
		idx[n.ID] = i
		d.Label = append(d.Label, n.Label)
		d.Kind = append(d.Kind, string(n.Kind))
		d.File = append(d.File, n.File)
		d.Line = append(d.Line, n.Line)
		d.Cat = append(d.Cat, categoryOf(n))
	}
	d.Deg = make([]int, len(doc.Nodes))
	for _, e := range doc.Edges {
		a, ok1 := idx[e.From]
		b, ok2 := idx[e.To]
		if !ok1 || !ok2 || a == b {
			continue
		}
		d.Deg[a]++
		d.Deg[b]++
		d.Edges = append(d.Edges, [2]int{a, b})
	}
	return d
}
```

(Delete the old `toColumnar`/`ColumnarScene` types and their tests. If `internal/render/columnar_test.go` had scene tests, they are replaced by the block above.)

- [ ] **Step 4: Run to verify pass** — `go test ./internal/render/ -run 'Island|Category|Deterministic' -v` → PASS.

- [ ] **Step 5: Commit** — `git commit -m "feat(render): graph data-island encoder (label/kind/file/deg/cat/edges)"`

---

## Task 2: viewer assets (force-graph app.js / app.css)

**Files:** overwrite `internal/render/viewer/app.css` and `internal/render/viewer/app.js`.

- [ ] **Step 1: app.css** — copy the prototype's `<style>…</style>` body (from `/tmp/gen_graph.py`) into `internal/render/viewer/app.css` verbatim (drop the surrounding `<style>` tags).

- [ ] **Step 2: app.js** — copy the prototype's `<script>…</script>` renderer body (from `/tmp/gen_graph.py`) into `internal/render/viewer/app.js`. One adaptation: the data-island element id. The prototype reads `document.getElementById('data')`; graffiti's island id is `graffiti-data`. Change the first line to:

```js
const D=JSON.parse(document.getElementById('graffiti-data').textContent);
```

Everything else (force sim, `fitVisible`, 2D/3D `threeD`, categories `catOn`, zones, dir/project tree, resizer, zoom/pan/drag/hover, sphere cache) is copied unchanged. The panel/canvas DOM ids it references (`panel,c,tip,fit,zones,d3,cats,tree,exp,col,nc,ec,sc,resizer`) must match the skeleton emitted in Task 3.

- [ ] **Step 3: Sanity** — `node --check internal/render/viewer/app.js` if `node` is available, else `grep -c 'function ' internal/render/viewer/app.js` to confirm it copied. (No Go test here; the renderer is validated by the browser smoke in Task 4.)

- [ ] **Step 4: Commit** — `git commit -m "feat(render): force-graph viewer assets (app.js/app.css)"`

---

## Task 3: maphtml.go rewrite + drop layout

**Files:** `internal/render/maphtml.go`, `internal/render/maphtml_test.go`, `internal/app/app.go`; DELETE `internal/layout/`.

- [ ] **Step 1: Rewrite `RenderMapHTML`/`WriteMapHTML`** to drop `scene layout.Scene`, emit the graph island, the force-graph skeleton, and the a11y mirror. Keep the CSP/escape/generated_at machinery verbatim. New signatures:

```go
func WriteMapHTML(doc *graph.Document, an analyze.Analysis, root string) error
func RenderMapHTML(doc *graph.Document, an analyze.Analysis, generatedAt string) string
```

In `RenderMapHTML`:
- `island := escapeScriptClose(mustJSON(graphIsland(doc)))` (add a tiny `mustJSON` helper or reuse json.Marshal+panic as before).
- CSP unchanged: `script-src 'sha256-<js>' 'sha256-<data>'; style-src 'sha256-<css>'; default-src 'none'; img-src data:`.
- Body skeleton = the prototype's `<body>…</body>` inner DOM (panel `#panel` with `#cats`/`#zones`/`#d3`/Structure header/`#tree`/`#resizer`, `<canvas id="c">`, `#tip`, `#fit` button, `.hint`) — emitted by Go string-building, with `doc.Root` HTML-escaped into the `<h1>`. The `generated_at` stays only in the HTML comment + `<body data-generated-at>`.
- Data island: `<script type="application/json" id="graffiti-data">island</script>`.
- Assets: `<style>viewerCSS</style>` in head, `<script>viewerJS</script>` before `</body>`.
- Replace `renderA11yMirror` to group nodes **by sector (directory)** instead of communities (the viewer no longer uses communities):

```go
func renderA11yMirror(b *strings.Builder, doc *graph.Document) {
	bySector := map[string][]graph.Node{}
	for _, n := range doc.Nodes {
		dir := n.File
		if i := strings.LastIndex(dir, "/"); i >= 0 {
			dir = dir[:i]
		} else {
			dir = "."
		}
		bySector[dir] = append(bySector[dir], n)
	}
	dirs := make([]string, 0, len(bySector))
	for d := range bySector {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	b.WriteString("<nav id=\"a11y\" aria-label=\"Code graph (accessibility mirror)\">\n")
	for _, d := range dirs {
		fmt.Fprintf(b, "<section><h3>%s (%d)</h3>\n<ul>\n", htmlEscape(d), len(bySector[d]))
		for _, n := range bySector[d] { // already id-sorted within doc.Nodes order
			fmt.Fprintf(b, "<li>%s (%s, %s:%d)</li>\n", htmlEscape(n.Label), htmlEscape(string(n.Kind)), htmlEscape(n.File), n.Line)
		}
		b.WriteString("</ul>\n</section>\n")
	}
	b.WriteString("</nav>\n")
}
```

Drop `renderRail` (the force-graph panel replaces it) and the `layout` import. Remove the old `Districts` title/top-bar (replaced by the in-panel `<h1>`). Keep the `<nav id="a11y">` hidden via app.css.

- [ ] **Step 2: Update `internal/app/app.go`** — remove `layout` import and the `scene := layout.Layout(...)` line; change the call to `render.WriteMapHTML(doc, an, absRoot)`.

- [ ] **Step 3: Delete `internal/layout/`** — `git rm internal/layout/layout.go internal/layout/layout_test.go`.

- [ ] **Step 4: Rewrite `internal/render/maphtml_test.go`** to the new shape (the existing tests reference `layout.Scene`/`ColumnarScene`/Districts text). Keep these invariants (adapt helpers to `RenderMapHTML(doc, an, generatedAt)`):
  - Determinism: same generated_at → byte-identical; different generated_at → differ only on the 2 generated_at-bearing lines, CSP line unchanged.
  - CSP hashes independently recomputed over the inlined `<style>`, `<script>` (renderer), and `<script id="graffiti-data">` bodies match the CSP.
  - Self-contained: no `http(s)://`, `src=`, `<link`, `@import`.
  - Island parses as JSON into the columnar shape (label/kind/file/deg/cat/edges) with consistent lengths.
  - Structural presence: `<canvas id="c"`, `id="a11y"`, `id="graffiti-data"`, `id="tree"`, `3D depth`, `fit to window`.
  - Escaping: a node label containing `</script><img onerror=…>` does not break the island (escaped to `<`) and is HTML-escaped in the a11y mirror body.
  - Writer writes `.graffiti/map.html`.

- [ ] **Step 5: Run** — `go test ./internal/render/ ./internal/app/ -v` → PASS (both tag-configs once wired).

- [ ] **Step 6: Commit** — `git commit -m "feat(render): force-graph map.html; drop treemap layout; a11y-by-sector"`

---

## Task 4: single-project golden + browser smoke

**Files:** `testdata/golden/gorepo.map.html.strip` (regenerate); browser smoke.

- [ ] **Step 1: Regenerate the strip golden.** The golden test strips `generated_at` and compares. Build the gorepo fixture, copy the produced `map.html` through the existing strip transform into `testdata/golden/gorepo.map.html.strip` (follow whatever the current golden test's "strip" definition is; reuse it). Then run the golden test → PASS.

- [ ] **Step 2: Both build configs green** — `go test ./...` and `make test` → all green.

- [ ] **Step 3: Browser smoke (manual, executor opens it).**

```bash
make build
S=$(mktemp -d); cp -r cmd internal "$S"/; ./graffiti build "$S" >/dev/null
open "$S/.graffiti/map.html"
```
Confirm: force-graph renders; 2D/3D toggle, categories, zones, project/dir tree, fit, resize, hover all work; self-contained (no network). Note: in-browser render is the one surface unit tests can't cover (as in Plan 3) — this manual smoke is the gate.

- [ ] **Step 4: Commit** — `git commit -m "test(render): regenerate map.html golden for the force-graph viewer"`

---

## Task 5: workspace render (`workspace.html`)

**Files:** `internal/render/workspace_html.go` (+ a combined-island builder), `cmd/graffiti/main.go`, tests.

- [ ] **Step 1: Combined island builder.** Add to `internal/workspace` a function returning the data the renderer needs, OR build it in render from the registry+overlay+member docs. Port the prototype's workspace `load()`: for each member load its `map.json`, prefix node ids `alias::` and file paths `alias/`, remap edges; append overlay confident links as edges. Produce a `*graph.Document`-like combined doc (alias-prefixed) and feed the SAME `graphIsland` + `RenderMapHTML` path — the file-path prefix makes the tree's top level the projects automatically. Tests: combined island has the project prefixes; cross-edges present; deterministic.

- [ ] **Step 2: `graffiti workspace render`.** Extend the `workspace` CLI case: `render` subcommand resolves the workspace root, loads registry + overlay, builds the combined doc, and writes `<root>/.graffiti-workspace/workspace.html` via the renderer. Reuse `resolveWorkspaceRoot`.

```
graffiti workspace render [--root dir]   # write .graffiti-workspace/workspace.html
```

- [ ] **Step 3: Test** (cmd-level): link the two-member fixture, `workspace render`, assert `workspace.html` exists, is self-contained, and its island top-level files are alias-prefixed (`backend/…`, `frontend/…`).

- [ ] **Step 4: Browser smoke.**

```bash
WS=$(mktemp -d); mkdir -p "$WS/graffiti" "$WS/polyglot"
cp -r cmd internal "$WS/graffiti"/; cp testdata/fixtures/polyglot/* "$WS/polyglot"/
./graffiti link --name demo "$WS/graffiti" "$WS/polyglot" >/dev/null
./graffiti workspace render --root "$WS"
open "$WS/.graffiti-workspace/workspace.html"
```
Confirm projects appear as top-level tree nodes; cross-edges drawn when present.

- [ ] **Step 5: Commit** — `git commit -m "feat(render,cli): graffiti workspace render → force-graph workspace.html"`

---

## Task 6: docs + full verification

- [ ] **Step 1: README** — update the viewer/Workspaces sections: `map.html` is now an interactive force-graph (2D/3D, categories, sector zones, project/dir tree, fit, zoom/pan); `graffiti workspace render` produces a federated `workspace.html`.

- [ ] **Step 2: Spec amendments** — §8: note the Districts treemap is **superseded** by the in-browser force-directed graph (pseudo-3D Canvas2D, layout client-side; file stays byte-deterministic in data+assets, positions runtime-only); §16.7: `workspace.html` **implemented** (projects as top-level tree, cross-project links drawn). Same isolate-and-pivot framing as prior §8/§10 amendments.

- [ ] **Step 3: Full verification** — `make vet`; `make test`; `go test ./...`; `go mod tidy && git diff --exit-code go.mod go.sum`; `make build && make xcompile` (size guard — assets are larger than the treemap but well within 16MB).

- [ ] **Step 4: Both browser smokes** (single + workspace) re-confirmed.

- [ ] **Step 5: Commit** — `git commit -m "docs: force-graph viewer + workspace render (spec §8/§16.7)"`

---

## Self-review checklist (before merge)

1. **Replaces Districts cleanly:** treemap `internal/layout` deleted; `app.go`/`maphtml.go` no longer reference it; no dead `ColumnarScene`/`toColumnar`.
2. **Render machinery preserved:** single offline file; CSP sha256 over the exact inlined bodies; `generated_at` only outside hashed bodies; `</script` escaped in the island; hidden a11y mirror present.
3. **Determinism:** island is a pure function of the sorted Document; `map.html` byte-identical modulo generated_at (golden). Force positions are runtime-only and intentionally not golden'd.
4. **Editability:** viewer is `viewer/app.js`/`app.css` (embedded) — edit + `make build` + refresh golden.
5. **No new deps:** stdlib + existing; `go mod tidy` no-op; binary within size guard.
6. **Workspace:** combined island reuses Plan-7 members+overlay; project prefix → projects as tree roots; cross-edges drawn.

## Deferred follow-ups (record in memory)

- Search box / click-node→open-file (the prototype has hover; wire a search + a `file://`-less "copy path" or host-callback later).
- Per-session layout cache (optional) so positions are stable across reloads (currently runtime-only).
- `graffiti init` could mention `graffiti workspace render` for federated repos.
- Larger graphs (>3–5k nodes): Barnes-Hut for the O(n²) repulsion if force settling gets slow.
