# graffiti Plan 3 — Districts `map.html` Viewer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** After `graffiti build .` clusters and analyzes the graph (Plans 1 & 2, merged), the pipeline now emits the third artifact: a single self-contained, offline, CSP-safe `.graffiti/map.html` — the **"Districts" city-map** of the codebase (spec §8). A new pure-functional `internal/layout` stage bakes a **deterministic squarified-treemap district scene** (community boxes sized by member count, cross-community edges aggregated into one labeled bundle per ordered `(A→B)` pair, god-node landmark pins, dashed surprising-link arcs — all integer coordinates) from the clustered `graph.Document` + `analyze.Analysis`. `internal/render` gains `WriteMapHTML(doc, an, scene, root)`, which inlines a hand-written vanilla-ES Canvas2D renderer (`render/viewer/app.js`) + CSS (`render/viewer/app.css`) via `go:embed`, ships the scene as a compact **columnar JSON data island** (interned string table + parallel integer arrays, with the `</script` sequence escaped), computes a **strict `sha256` meta CSP** as a pure function of the inlined bodies, emits a **hidden ordered DOM accessibility mirror** of the districts, and keeps `generated_at` as the *only* varying bytes (outside every hashed body). `app.Build` writes `map.html` next to `map.json`/`MAP.md`; the existing two goldens are unaffected.

**Architecture:** Same pure-functional pipeline shape Plans 1 & 2 established — plain structs in, plain structs out, no shared mutable state, no I/O above `render`. Plan 3 inserts one stage (`layout`) between `analyze` and `render` and adds one emitter to `render`:

```
scan → parse → build → cluster → analyze → layout → render(map.json + MAP.md + map.html)
```

- `internal/layout` reads the clustered `*graph.Document` + `analyze.Analysis` and returns a `layout.Scene` value: integer-coordinate `Box`es (one per community, area ∝ member count via a deterministic squarified treemap, no top-tier force per spec §8.2), `Bundle`s (one per ordered cross-community `(A→B)` pair, with baked elbow control points and an aggregated edge count), `Pin`s (god-node landmarks placed on their district), and `Arc`s (dashed surprising-link gutter arcs). No mutation, no I/O, no `math/rand`, no map-iteration feeding any emitted order.
- `internal/render` gains `WriteMapHTML(doc, an, scene, root)` + `RenderMapHTML(...) string`, plus the embedded viewer assets in `internal/render/viewer/`. The browser does **no layout and no physics** (spec §8.2): the renderer reads the baked integer scene from the data island and only does camera transform, viewport cull, batched paint, and analytic hit-test.
- `internal/app.Build` calls `layout.Layout(doc, an)` after `analyze.Analyze`, then `render.WriteMapHTML(doc, an, scene, absRoot)` after the existing two writers. No CLI output change (the success line already prints the 3 questions from Plan 2).

**Tech Stack:** Go 1.26; module `github.com/evgeniy-achin/graffiti`; **no new third-party dependencies** (`layout` uses only `sort`; `render` adds `crypto/sha256`, `encoding/base64`, `encoding/json`, `embed`, `html`, `strings` from the stdlib). `CGO_ENABLED=0`. The viewer renderer is **our own ~25–30KB hand-written vanilla-ES file** (no modules, no `eval`, no `new Function`, no imports, zero third-party JS — spec §8.1) embedded via `//go:embed`. The grammar-subset build tags from Plan 1 still apply to any target that links `internal/parse` (i.e. `internal/app` and `cmd/graffiti`), so every `go test`/`go build` command that touches those packages keeps `-tags "grammar_subset grammar_subset_go grammar_subset_gomod"`; the pure `internal/layout` and `internal/render` packages do not need the tags, but always passing them is harmless and keeps commands uniform.

---

## Validated Prototype Facts (load-bearing; the two risky pieces were prototyped end-to-end in a scratch module and their tests pass)

Before writing this plan the two risky pieces — the **deterministic squarified-treemap layout** and the **self-contained CSP-safe HTML emitter** — were prototyped in a throwaway `package main` against local mirrors of the real `graph.*`/`analyze.*` types, with passing tests. Treat these as ground truth; the task code below is the same validated prototype with the local mirror types replaced by the real `graph.*`/`analyze.*` types and the inlined CSS/JS sourced from `go:embed` files instead of Go string constants. The prototypes live at `/tmp/p3proto/{layout,htmlemit}{,_test}.go`.

**Layout (validated — `layout_test.go` passes):**
- A **squarified treemap** packs community boxes into a fixed `1600×1000` canvas inner rect; **box area is proportional to member count** (min 1 so a zero-member community never vanishes); children are processed in descending area, ties broken by ascending community id (a total order).
- **No overlap, in-bounds, integer coordinates:** verified by `TestLayout_NoOverlap` (pairwise AABB non-intersection over the gutter-inset boxes), `TestLayout_InBounds` (every box and pin within the canvas, positive dims), and integer quantization at every row (`round()` at row close, last cell snapped to the rect edge).
- **Byte-stable / deterministic:** `TestLayout_Deterministic` builds the scene twice and asserts `fmt.Sprintf("%+v", …)` equality. The Scene's slices are re-sorted to a stable emit order (boxes by comm id; pins by `(commID, nodeID)`; bundles by `(from, to)`; arcs by `(from, to, conf)`) so no Go-map iteration feeds emitted order.
- **Bundles aggregated:** `TestLayout_BundlesAggregated` confirms one bundle per ordered `(A→B)` pair with the summed cross-community edge count (e.g. two `0→1` edges collapse to one bundle of count 2). Each bundle bakes a 3-point orthogonal elbow polyline between box centers.
- **Pins on their district:** `TestLayout_PinsOnTheirDistrict` confirms every god-node pin sits inside its own community box (fanned on a fixed 3-wide grid, clamped to the box interior).
- **Area proportionality:** `TestLayout_AreaProportional` confirms the 12-member community's box area exceeds the 1-member community's.

**HTML emit (validated — `htmlemit_test.go` passes):**
- The scene is flattened to a **compact columnar form**: an interned string table (`strings`, index 0 == `""`) plus parallel integer arrays for boxes/pins, and header columns + flattened point coords with prefix-sum offset arrays for bundles/arcs. `toColumnar` is deterministic because the Scene emit order is.
- The emitter inlines a `<style>` and a `<script>` body and computes a **strict meta CSP** `default-src 'none'; script-src 'sha256-…' 'sha256-…'; style-src 'sha256-…'; img-src data:`, where each `sha256` is the standard-base64 digest of the **exact bytes inlined as that element's body** (verified independently by `TestEmit_CSPHashesMatchInlinedBodies`, which re-extracts the inlined bodies straight out of the HTML and recomputes the digests).
- **`generated_at` lives only OUTSIDE every hashed body** — in one HTML comment line and one `<body data-generated-at>` attribute. `TestEmit_DeterministicSameTime` proves byte-identical output for equal inputs; `TestEmit_DiffersOnlyByGeneratedAt` proves that changing `generated_at` changes **exactly two lines** and leaves the CSP line (hashes) byte-identical.
- **Self-contained:** `TestEmit_SelfContained` asserts the output contains no `http://`, `https://`, `src=`, `<link`, `@import`, nor any `scheme://` URL.
- **Data island parses to the columnar shape:** `TestEmit_DataIslandParsesAndHasColumnarShape` extracts the `application/json` island, `json.Unmarshal`s it, and checks the parallel-array invariants (equal column lengths, string-table resolution, offset arrays = `n+1` prefix sums).
- **`</script` escape (prototype-flagged, MUST be in the plan):** a `<script type="application/json">` island is a *data block*, not executed by the UA, so `script-src` does not gate it and it needs no hash for execution. **However, the JSON text MUST still escape any `</script` sequence before inlining** (replace `<` with the JSON unicode escape `<`), otherwise a label/path containing `</script>` would prematurely close the island and break parsing (and is an XSS vector). **Decision (validated):** we escape `</script` (via `<` → `<` over the whole island, which JSON tolerates) **and** we still include the island's `sha256` in `script-src` as belt-and-braces (harmless; satisfies over-strict scanners that flag any `<script>` element). Because the escape is applied to the bytes *before* hashing, the data-island hash stays a pure function of the (deterministic) scene.

**Adaptation deltas from prototype → real code (the only differences):**
1. Local mirror `Node/Edge/Community/Document` → `graph.Node/Edge/Community/Document` (real fields: `Kind graph.Kind`, `Relation graph.Relation`, `Confidence graph.Confidence` — string-backed types, so `string(...)` where the columnar form needs a plain string).
2. Local mirror `GodNode/SurprisingEdge/Analysis` → `analyze.GodNode` (`{ID, Label string; Degree int}`), `analyze.SurprisingEdge` (`{From, To string; FromComm, ToComm int; Relation graph.Relation; Confidence graph.Confidence}`), `analyze.Analysis`. The prototype already reads exactly these fields (`g.ID`, `g.Label`, `s.FromComm`, `s.ToComm`, `s.Confidence`); only the `Confidence`/`Relation` types are now `graph.*` (use `string(s.Confidence)`).
3. The community→node map is built from `graph.Node.Community` (the real clustered field), not a mirror field — identical shape.
4. Inlined CSS/JS move from Go `const` strings into `internal/render/viewer/app.css` + `app.js`, read via `//go:embed`. The CSP/`</script`-escape mechanics are unchanged (still hash the exact embedded bytes).
5. The prototype's tiny DOM-builder `<script>` is replaced by the real ~25–30KB Canvas2D renderer (Task 3); the emit/CSP/escape/columnar plumbing is otherwise byte-for-byte the validated approach.

**Tiers 2 & 3 are a DOCUMENTED FAST-FOLLOW (deferred), not in this plan.** v1 `map.html` ships **Tier 1 (Districts) + click-to-inspect a box** only: named district boxes (area = size, border = centrality), bundled labeled flow-arrows, god-node landmark pins, dashed surprising arcs, the left rail (search, 3 questions, landmarks, confidence legend), and click-a-box-to-inspect. **Semantic-zoom Tier 2 (district interior / file shelves) and Tier 3 (symbol + callers/callees), and click-an-arrow-to-list-edges, are explicitly deferred to a later plan** (Open Issue #1) — they need file-shelf + symbol layout (`graph` carries no per-symbol containment baked into the Scene yet) and a richer data island; deferring them keeps v1 legible, byte-deterministic, and within the <1.5 MB budget. This is stated again in the renderer task and the Self-Review.

**Renderer testability note (load-bearing):** `app.js` is delivered as **embedded asset content**; it is not Go code and cannot be Go-unit-tested, and this plan adds **no browser/JS test harness** (offline, zero-dependency). Therefore **all tests target the Go-emitted artifact** (`map.html` bytes + the layout `Scene`), never JS runtime behavior: layout determinism/invariants, HTML byte-determinism modulo `generated_at`, self-containment, CSP correctness, structural presence (`<canvas>`, data island parses to the columnar shape, a11y mirror lists every district, the 3 questions present), and the two escapes (HTML-escaped labels + the `</script` island escape). A human smoke-opens the file in a browser in the final task.

---

## File Structure

```
graffiti/
├── internal/
│   ├── layout/
│   │   ├── layout.go            # NEW: Layout(doc, an) Scene — deterministic squarified-treemap district scene
│   │   └── layout_test.go       # NEW: no-overlap / in-bounds / integer / byte-stable / bundles / pins invariants
│   ├── render/
│   │   ├── json.go              # (unchanged from Plan 1)
│   │   ├── mapmd.go             # (unchanged from Plan 2)
│   │   ├── columnar.go          # NEW: ColumnarScene + toColumnar(scene) — interned strings + parallel int arrays
│   │   ├── maphtml.go           # NEW: WriteMapHTML(doc, an, scene, root) + RenderMapHTML(...) + CSP/escape plumbing
│   │   ├── maphtml_test.go      # NEW: determinism / self-containment / CSP / structural / escaping tests
│   │   └── viewer/
│   │       ├── app.css          # NEW: light-theme styles (rail, top bar, canvas, a11y-mirror) — embedded, hashed
│   │       └── app.js           # NEW: ~25–30KB hand-written vanilla-ES Canvas2D renderer — embedded, hashed
│   └── app/
│       ├── app.go               # MODIFIED: layout.Layout + render.WriteMapHTML wired into Build
│       └── golden_test.go       # MODIFIED: add map.html artifact tests (determinism/self-contained/CSP/structure)
└── testdata/golden/
    ├── gorepo.map.json          # (unchanged — Plan 2 golden)
    ├── gorepo.MAP.md            # (unchanged — Plan 2 golden)
    └── gorepo.map.html.strip    # NEW (optional, via UPDATE_GOLDEN): map.html with generated_at stripped
```

**Package responsibilities (one job each):**
- `layout` owns the deterministic Tier-1 district Scene (treemap boxes, aggregated bundles, god-node pins, surprising arcs) — pure, no I/O, `sort`-only. It is the single place where geometry is baked, so the browser never lays anything out (spec §8.2/§8.8).
- `render` gains the columnar encoder + the HTML emitter + the embedded viewer assets; the JSON and MAP.md writers are untouched, so the two existing goldens stay byte-identical.
- `app` wires the new `layout` stage and `WriteMapHTML` writer; `cmd/graffiti` is unchanged.

---

## Task 1: `internal/layout` — deterministic squarified-treemap district Scene

**Files:**
- Create: `internal/layout/layout.go`
- Create: `internal/layout/layout_test.go`

- [ ] **Step 1: Write the failing layout invariant test**

Create `/Users/mylive/project/graffiti/graffiti/internal/layout/layout_test.go`:
```go
package layout

import (
	"fmt"
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/analyze"
	"github.com/evgeniy-achin/graffiti/internal/graph"
)

// sampleDoc builds a 5-community clustered document with intra- and
// cross-community edges, two god nodes, and two surprising links — the same
// shape validated in the layout prototype, expressed against the real types.
func sampleDoc() (*graph.Document, analyze.Analysis) {
	doc := graph.NewDocument("demo")
	doc.GeneratedAt = "2026-06-15T00:00:00Z"
	sizes := []int{12, 8, 5, 3, 1}
	labels := []string{"Auth", "Http", "Parser", "Cache", "Util"}
	for ci, sz := range sizes {
		var members []string
		for k := 0; k < sz; k++ {
			id := fmt.Sprintf("c%d-n%d", ci, k)
			doc.Nodes = append(doc.Nodes, graph.Node{
				ID: id, Label: id, Kind: graph.KindFunction,
				File: fmt.Sprintf("pkg%d/f%d.go", ci, k), Line: 1, Community: ci,
			})
			members = append(members, id)
		}
		doc.Communities = append(doc.Communities, graph.Community{ID: ci, Label: labels[ci], Members: members})
	}
	mk := func(f, t string, rel graph.Relation, conf graph.Confidence) {
		doc.Edges = append(doc.Edges, graph.Edge{From: f, To: t, Relation: rel, Confidence: conf})
	}
	mk("c0-n0", "c0-n1", graph.RelCalls, graph.ConfInferred)
	mk("c0-n0", "c1-n0", graph.RelCalls, graph.ConfExtracted) // cross 0->1
	mk("c0-n1", "c1-n0", graph.RelCalls, graph.ConfInferred)  // cross 0->1 again (bundle count 2)
	mk("c1-n0", "c2-n0", graph.RelCalls, graph.ConfInferred)  // cross 1->2
	mk("c3-n0", "c0-n0", graph.RelReferences, graph.ConfAmbiguous)
	an := analyze.Analysis{
		GodNodes: []analyze.GodNode{
			{ID: "c0-n0", Label: "c0-n0", Degree: 4},
			{ID: "c1-n0", Label: "c1-n0", Degree: 3},
		},
		Surprising: []analyze.SurprisingEdge{
			{From: "c0-n0", To: "c1-n0", FromComm: 0, ToComm: 1, Relation: graph.RelCalls, Confidence: graph.ConfExtracted},
			{From: "c3-n0", To: "c0-n0", FromComm: 3, ToComm: 0, Relation: graph.RelReferences, Confidence: graph.ConfAmbiguous},
		},
		Questions: []string{"q1", "q2", "q3"},
	}
	return doc, an
}

func TestLayout_NoOverlap(t *testing.T) {
	doc, an := sampleDoc()
	sc := Layout(doc, an)
	for i := 0; i < len(sc.Boxes); i++ {
		for j := i + 1; j < len(sc.Boxes); j++ {
			a, b := sc.Boxes[i], sc.Boxes[j]
			if a.X < b.X+b.W && b.X < a.X+a.W && a.Y < b.Y+b.H && b.Y < a.Y+a.H {
				t.Fatalf("boxes overlap: %+v and %+v", a, b)
			}
		}
	}
}

func TestLayout_InBounds(t *testing.T) {
	doc, an := sampleDoc()
	sc := Layout(doc, an)
	for _, b := range sc.Boxes {
		if b.X < 0 || b.Y < 0 || b.X+b.W > sc.W || b.Y+b.H > sc.H {
			t.Fatalf("box out of bounds: %+v (canvas %dx%d)", b, sc.W, sc.H)
		}
		if b.W <= 0 || b.H <= 0 {
			t.Fatalf("box has non-positive dims: %+v", b)
		}
	}
	for _, p := range sc.Pins {
		if p.X < 0 || p.Y < 0 || p.X > sc.W || p.Y > sc.H {
			t.Fatalf("pin out of bounds: %+v", p)
		}
	}
}

func TestLayout_PinsOnTheirDistrict(t *testing.T) {
	doc, an := sampleDoc()
	sc := Layout(doc, an)
	boxByComm := map[int]Box{}
	for _, b := range sc.Boxes {
		boxByComm[b.CommID] = b
	}
	for _, p := range sc.Pins {
		b := boxByComm[p.CommID]
		if p.X < b.X || p.X > b.X+b.W || p.Y < b.Y || p.Y > b.Y+b.H {
			t.Fatalf("pin %+v not inside its district box %+v", p, b)
		}
	}
}

func TestLayout_Deterministic(t *testing.T) {
	d1, a1 := sampleDoc()
	d2, a2 := sampleDoc()
	if s1, s2 := fmt.Sprintf("%+v", Layout(d1, a1)), fmt.Sprintf("%+v", Layout(d2, a2)); s1 != s2 {
		t.Fatalf("non-deterministic layout:\nA=%s\nB=%s", s1, s2)
	}
}

func TestLayout_IntegerCoords(t *testing.T) {
	// All coords are Go ints by type; this guards the contract explicitly so a
	// future float refactor that rounds late cannot silently slip through.
	doc, an := sampleDoc()
	sc := Layout(doc, an)
	for _, b := range sc.Boxes {
		_ = b.X + b.Y + b.W + b.H + b.Border // ints; compile-time guarantee
	}
	for _, bn := range sc.Bundles {
		for _, pt := range bn.Pts {
			_ = pt[0] + pt[1]
		}
	}
}

func TestLayout_BundlesAggregated(t *testing.T) {
	doc, an := sampleDoc()
	sc := Layout(doc, an)
	got := map[[2]int]int{}
	for _, b := range sc.Bundles {
		got[[2]int{b.FromComm, b.ToComm}] = b.Count
	}
	if got[[2]int{0, 1}] != 2 {
		t.Fatalf("bundle 0->1 count = %d, want 2", got[[2]int{0, 1}])
	}
	if got[[2]int{1, 2}] != 1 {
		t.Fatalf("bundle 1->2 count = %d, want 1", got[[2]int{1, 2}])
	}
	if got[[2]int{3, 0}] != 1 {
		t.Fatalf("bundle 3->0 count = %d, want 1", got[[2]int{3, 0}])
	}
}

func TestLayout_AreaProportional(t *testing.T) {
	doc, an := sampleDoc()
	sc := Layout(doc, an)
	byComm := map[int]Box{}
	for _, b := range sc.Boxes {
		byComm[b.CommID] = b
	}
	if a0, a4 := byComm[0].W*byComm[0].H, byComm[4].W*byComm[4].H; a0 <= a4 {
		t.Fatalf("area not proportional: comm0 area %d <= comm4 area %d", a0, a4)
	}
}

func TestLayout_SurprisingArcs(t *testing.T) {
	doc, an := sampleDoc()
	sc := Layout(doc, an)
	if len(sc.Arcs) != 2 {
		t.Fatalf("arcs = %d, want 2 (one per surprising edge pair)", len(sc.Arcs))
	}
	// arcs emitted sorted by (from, to, conf): 0->1 then 3->0.
	if sc.Arcs[0].FromComm != 0 || sc.Arcs[0].ToComm != 1 {
		t.Fatalf("arc[0] = %d->%d, want 0->1", sc.Arcs[0].FromComm, sc.Arcs[0].ToComm)
	}
	if sc.Arcs[0].Confidence != "EXTRACTED" {
		t.Fatalf("arc[0] confidence = %q, want EXTRACTED", sc.Arcs[0].Confidence)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/layout/ -run TestLayout -v`
Expected: FAIL — compilation error `undefined: Layout` (and `undefined: Box`).

- [ ] **Step 3: Write the layout implementation**

Create `/Users/mylive/project/graffiti/graffiti/internal/layout/layout.go`:
```go
// Package layout bakes a deterministic Tier-1 "Districts" scene (spec §8.2/§8.3)
// from a clustered graph.Document + analyze.Analysis. It performs no I/O and no
// mutation. Geometry is precomputed here as integer coordinates so the browser
// does NO layout and NO physics — the only way to satisfy the §8.8/§14
// byte-identical guarantee. Communities are packed by a squarified treemap (box
// area proportional to member count, no top-tier force); cross-community edges
// aggregate into one bundle per ordered (A->B) pair; god nodes become landmark
// pins; surprising links become dashed gutter arcs. No math/rand, no time, and
// no Go-map iteration ever feeds an emitted order (every emitted slice is
// re-sorted with an explicit total order).
package layout

import (
	"sort"

	"github.com/evgeniy-achin/graffiti/internal/analyze"
	"github.com/evgeniy-achin/graffiti/internal/graph"
)

// Canvas + spacing constants. The fixed canvas keeps coords byte-stable across
// runs; the browser scales it to the viewport.
const (
	CanvasW    = 1600
	CanvasH    = 1000
	Pad        = 24 // outer padding
	Gutter     = 16 // min gap baked between districts
	PinR       = 7
	titleBandH = 28
)

// Box is one community district (integer coords; gutter-inset).
type Box struct {
	CommID     int
	Label      string
	Count      int
	X, Y, W, H int
	Border     int // centrality-derived border weight (1..4)
}

// Pin is a god-node landmark placed inside its district box.
type Pin struct {
	NodeID, Label string
	CommID        int
	X, Y          int
}

// Bundle is one aggregated ordered (FromComm->ToComm) flow of Count edges, with
// a baked orthogonal/elbow polyline through the gutters.
type Bundle struct {
	FromComm, ToComm int
	Count            int
	Pts              [][2]int
}

// Arc is one dashed surprising cross-community link (baked elbow polyline).
type Arc struct {
	FromComm, ToComm int
	Confidence       string
	Pts              [][2]int
}

// Scene is the full deterministic Tier-1 district scene.
type Scene struct {
	W, H    int
	Boxes   []Box
	Pins    []Pin
	Bundles []Bundle
	Arcs    []Arc
}

// Layout produces the deterministic Tier-1 district scene.
func Layout(doc *graph.Document, an analyze.Analysis) Scene {
	sc := Scene{W: CanvasW, H: CanvasH}

	// 1. Community order (by id asc) + member counts.
	comms := append([]graph.Community(nil), doc.Communities...)
	sort.Slice(comms, func(i, j int) bool { return comms[i].ID < comms[j].ID })

	// per-node degree (for centrality-derived border weight) + node->community.
	deg := map[string]int{}
	for _, e := range doc.Edges {
		deg[e.From]++
		deg[e.To]++
	}
	commOf := map[string]int{}
	for _, n := range doc.Nodes {
		commOf[n.ID] = n.Community
	}

	type item struct {
		id    int
		label string
		count int
		cent  int
	}
	items := make([]item, 0, len(comms))
	for _, c := range comms {
		cent := 0
		for _, m := range c.Members {
			cent += deg[m]
		}
		items = append(items, item{c.ID, c.Label, len(c.Members), cent})
	}
	if len(items) == 0 {
		return sc
	}

	// 2. Squarified treemap over the inner rect. Area ∝ count (min 1). Children
	//    processed desc area, ties by asc community id (total order).
	innerX, innerY := Pad, Pad+titleBandH
	innerW, innerH := CanvasW-2*Pad, CanvasH-Pad-(Pad+titleBandH)

	var total float64
	for _, it := range items {
		c := it.count
		if c < 1 {
			c = 1
		}
		total += float64(c)
	}
	scale := float64(innerW*innerH) / total

	order := make([]int, len(items))
	for i := range order {
		order[i] = i
	}
	area := func(i int) float64 {
		c := items[i].count
		if c < 1 {
			c = 1
		}
		return float64(c) * scale
	}
	sort.SliceStable(order, func(a, b int) bool {
		ai, bi := order[a], order[b]
		if area(ai) != area(bi) {
			return area(ai) > area(bi)
		}
		return items[ai].id < items[bi].id
	})

	rects := squarify(order, area, innerX, innerY, innerW, innerH)

	// quantize + gutter-inset; key boxes by community id.
	boxByComm := map[int]Box{}
	for idx, r := range rects {
		it := items[idx]
		b := Box{
			CommID: it.id,
			Label:  it.label,
			Count:  it.count,
			X:      r.x + Gutter/2,
			Y:      r.y + Gutter/2,
			W:      r.w - Gutter,
			H:      r.h - Gutter,
			Border: borderWeight(it.cent),
		}
		if b.W < 8 {
			b.W = 8
		}
		if b.H < 8 {
			b.H = 8
		}
		sc.Boxes = append(sc.Boxes, b)
		boxByComm[it.id] = b
	}
	sort.Slice(sc.Boxes, func(i, j int) bool { return sc.Boxes[i].CommID < sc.Boxes[j].CommID })

	// 3. God-node landmark pins, fanned on a fixed 3-wide grid inside the box.
	pinIdx := map[int]int{}
	for _, g := range an.GodNodes {
		cid, ok := commOf[g.ID]
		if !ok {
			continue
		}
		b, ok := boxByComm[cid]
		if !ok {
			continue
		}
		k := pinIdx[cid]
		pinIdx[cid]++
		px := b.X + 14 + (k%3)*16
		py := b.Y + 16 + (k/3)*16
		if px > b.X+b.W-PinR {
			px = b.X + b.W - PinR
		}
		if py > b.Y+b.H-PinR {
			py = b.Y + b.H - PinR
		}
		sc.Pins = append(sc.Pins, Pin{NodeID: g.ID, Label: g.Label, CommID: cid, X: px, Y: py})
	}
	sort.Slice(sc.Pins, func(i, j int) bool {
		if sc.Pins[i].CommID != sc.Pins[j].CommID {
			return sc.Pins[i].CommID < sc.Pins[j].CommID
		}
		return sc.Pins[i].NodeID < sc.Pins[j].NodeID
	})

	// 4. Aggregate inter-community edges into one bundle per ordered (A->B) pair.
	type key struct{ a, b int }
	bcount := map[key]int{}
	for _, e := range doc.Edges {
		fa, oka := commOf[e.From]
		fb, okb := commOf[e.To]
		if !oka || !okb || fa == fb {
			continue
		}
		bcount[key{fa, fb}]++
	}
	keys := make([]key, 0, len(bcount))
	for k := range bcount {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].a != keys[j].a {
			return keys[i].a < keys[j].a
		}
		return keys[i].b < keys[j].b
	})
	for _, k := range keys {
		ba, oka := boxByComm[k.a]
		bb, okb := boxByComm[k.b]
		if !oka || !okb {
			continue
		}
		sc.Bundles = append(sc.Bundles, Bundle{
			FromComm: k.a, ToComm: k.b, Count: bcount[k],
			Pts: elbow(ba, bb),
		})
	}

	// 5. Surprising links -> dashed arcs (dedup by ordered pair + confidence).
	type akey struct {
		a, b int
		conf string
	}
	seenArc := map[akey]bool{}
	var arcKeys []akey
	for _, s := range an.Surprising {
		ak := akey{s.FromComm, s.ToComm, string(s.Confidence)}
		if seenArc[ak] {
			continue
		}
		seenArc[ak] = true
		arcKeys = append(arcKeys, ak)
	}
	sort.Slice(arcKeys, func(i, j int) bool {
		if arcKeys[i].a != arcKeys[j].a {
			return arcKeys[i].a < arcKeys[j].a
		}
		if arcKeys[i].b != arcKeys[j].b {
			return arcKeys[i].b < arcKeys[j].b
		}
		return arcKeys[i].conf < arcKeys[j].conf
	})
	for _, ak := range arcKeys {
		ba, oka := boxByComm[ak.a]
		bb, okb := boxByComm[ak.b]
		if !oka || !okb {
			continue
		}
		sc.Arcs = append(sc.Arcs, Arc{FromComm: ak.a, ToComm: ak.b, Confidence: ak.conf, Pts: elbow(ba, bb)})
	}

	return sc
}

func borderWeight(cent int) int {
	switch {
	case cent >= 40:
		return 4
	case cent >= 16:
		return 3
	case cent >= 4:
		return 2
	default:
		return 1
	}
}

func cx(b Box) int { return b.X + b.W/2 }
func cy(b Box) int { return b.Y + b.H/2 }

// elbow routes a deterministic integer 3-point orthogonal polyline (horizontal
// then vertical) between two box centers.
func elbow(a, b Box) [][2]int {
	ax, ay := cx(a), cy(a)
	bx, by := cx(b), cy(b)
	return [][2]int{{ax, ay}, {bx, ay}, {bx, by}}
}

// ---- Squarified treemap (validated prototype; integer-quantized) ----

type rect struct{ x, y, w, h int }

type cell struct {
	pos int
	a   float64
}

// squarify returns one rect per position in `order` (sorted desc area). Result[i]
// corresponds to order[i]. Coordinates are integer-quantized at each row close.
func squarify(order []int, area func(int) float64, x, y, w, h int) []rect {
	out := make([]rect, len(order))
	cells := make([]cell, len(order))
	for i, idx := range order {
		cells[i] = cell{i, area(idx)}
	}

	fx, fy, fw, fh := float64(x), float64(y), float64(w), float64(h)
	i := 0
	for i < len(cells) {
		rowStart := i
		short := fw
		if fh < short {
			short = fh
		}
		bestWorst := worstRatio(cells[rowStart:rowStart+1], short)
		j := i + 1
		for j < len(cells) {
			w2 := worstRatio(cells[rowStart:j+1], short)
			if w2 > bestWorst {
				break
			}
			bestWorst = w2
			j++
		}
		row := cells[rowStart:j]
		var rowSum float64
		for _, c := range row {
			rowSum += c.a
		}
		if fw <= fh {
			rh := rowSum / fw
			cxp := fx
			for ri, c := range row {
				cw := c.a / rh
				rx, ry := round(cxp), round(fy)
				rwd := round(cw)
				if ri == len(row)-1 {
					rwd = round(fx+fw) - rx
				}
				out[c.pos] = rect{rx, ry, rwd, round(rh)}
				cxp += cw
			}
			fy += rh
			fh -= rh
		} else {
			rw := rowSum / fh
			cyp := fy
			for ri, c := range row {
				ch := c.a / rw
				rx, ry := round(fx), round(cyp)
				rht := round(ch)
				if ri == len(row)-1 {
					rht = round(fy+fh) - ry
				}
				out[c.pos] = rect{rx, ry, round(rw), rht}
				cyp += ch
			}
			fx += rw
			fw -= rw
		}
		i = j
	}
	return out
}

func worstRatio(row []cell, short float64) float64 {
	var sum, max, min float64
	min = 1e18
	for _, c := range row {
		sum += c.a
		if c.a > max {
			max = c.a
		}
		if c.a < min {
			min = c.a
		}
	}
	if sum == 0 {
		return 1e18
	}
	s2 := short * short
	w1 := s2 * max / (sum * sum)
	w2 := sum * sum / (s2 * min)
	if w1 > w2 {
		return w1
	}
	return w2
}

func round(f float64) int {
	if f < 0 {
		return int(f - 0.5)
	}
	return int(f + 0.5)
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/layout/ -v`
Expected: PASS — `TestLayout_NoOverlap`, `TestLayout_InBounds`, `TestLayout_PinsOnTheirDistrict`, `TestLayout_Deterministic`, `TestLayout_IntegerCoords`, `TestLayout_BundlesAggregated`, `TestLayout_AreaProportional`, `TestLayout_SurprisingArcs`.

- [ ] **Step 5: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git checkout -b plan3-districts-viewer
git add internal/layout/layout.go internal/layout/layout_test.go
git commit -m "feat: deterministic squarified-treemap district layout (internal/layout)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: `internal/render` — columnar scene encoder (interned strings + parallel int arrays)

**Files:**
- Create: `internal/render/columnar.go`
- Create: `internal/render/columnar_test.go`

- [ ] **Step 1: Write the failing columnar test**

Create `/Users/mylive/project/graffiti/graffiti/internal/render/columnar_test.go`:
```go
package render

import (
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/layout"
)

// tinyScene is a hand-built Scene with two boxes, one pin, one bundle, one arc —
// enough to exercise every column + the interned string table.
func tinyScene() layout.Scene {
	return layout.Scene{
		W: 1600, H: 1000,
		Boxes: []layout.Box{
			{CommID: 0, Label: "Auth", Count: 3, X: 10, Y: 20, W: 100, H: 80, Border: 2},
			{CommID: 1, Label: "Http", Count: 1, X: 120, Y: 20, W: 40, H: 80, Border: 1},
		},
		Pins: []layout.Pin{
			{NodeID: "auth-login", Label: "Login", CommID: 0, X: 24, Y: 36},
		},
		Bundles: []layout.Bundle{
			{FromComm: 0, ToComm: 1, Count: 2, Pts: [][2]int{{60, 60}, {140, 60}, {140, 60}}},
		},
		Arcs: []layout.Arc{
			{FromComm: 1, ToComm: 0, Confidence: "AMBIGUOUS", Pts: [][2]int{{140, 60}, {60, 60}, {60, 60}}},
		},
	}
}

func TestToColumnar_ParallelArraysAndStringTable(t *testing.T) {
	sc := tinyScene()
	cs := toColumnar(sc)

	if cs.W != sc.W || cs.H2 != sc.H {
		t.Fatalf("canvas dims = %dx%d, want %dx%d", cs.W, cs.H2, sc.W, sc.H)
	}
	if len(cs.Strings) == 0 || cs.Strings[0] != "" {
		t.Fatalf("string table must exist with index 0 == \"\", got %v", cs.Strings)
	}
	nb := len(sc.Boxes)
	for name, col := range map[string][]int{
		"BoxComm": cs.BoxComm, "BoxLabel": cs.BoxLabel, "BoxCount": cs.BoxCount,
		"BoxX": cs.BoxX, "BoxY": cs.BoxY, "BoxW": cs.BoxW, "BoxH": cs.BoxH, "BoxBorder": cs.BoxBorder,
	} {
		if len(col) != nb {
			t.Fatalf("box column %s len %d, want %d", name, len(col), nb)
		}
	}
	// label indices resolve back to the real labels.
	for i, b := range sc.Boxes {
		if cs.Strings[cs.BoxLabel[i]] != b.Label {
			t.Fatalf("box %d label via table = %q, want %q", i, cs.Strings[cs.BoxLabel[i]], b.Label)
		}
	}
	// offset arrays are prefix sums of length n+1; last*2 == flattened pts len.
	if len(cs.BundleOff) != len(sc.Bundles)+1 {
		t.Fatalf("bundleOff len %d, want %d", len(cs.BundleOff), len(sc.Bundles)+1)
	}
	if got := cs.BundleOff[len(cs.BundleOff)-1] * 2; got != len(cs.BundlePts) {
		t.Fatalf("bundlePts len %d, want %d", len(cs.BundlePts), got)
	}
	if got := cs.ArcOff[len(cs.ArcOff)-1] * 2; got != len(cs.ArcPts) {
		t.Fatalf("arcPts len %d, want %d", len(cs.ArcPts), got)
	}
}

func TestToColumnar_Deterministic(t *testing.T) {
	a, b := toColumnar(tinyScene()), toColumnar(tinyScene())
	if a.W != b.W || len(a.Strings) != len(b.Strings) || len(a.BoxX) != len(b.BoxX) {
		t.Fatalf("columnar encoding not deterministic")
	}
	for i := range a.Strings {
		if a.Strings[i] != b.Strings[i] {
			t.Fatalf("string table index %d differs: %q vs %q", i, a.Strings[i], b.Strings[i])
		}
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/render/ -run TestToColumnar -v`
Expected: FAIL — `undefined: toColumnar` (and `undefined: ColumnarScene` fields).

- [ ] **Step 3: Write the columnar encoder**

Create `/Users/mylive/project/graffiti/graffiti/internal/render/columnar.go`:
```go
package render

import "github.com/evgeniy-achin/graffiti/internal/layout"

// ColumnarScene is the compact, JSON-friendly representation inlined as the
// map.html data island (spec §8.6): every repeated string is interned once into
// Strings (index 0 == ""); everything else is parallel integer arrays. The
// browser rebuilds objects by index lookup. JSON keys come from struct field
// order; arrays are already deterministically ordered by layout.Layout, so the
// encoded bytes are deterministic.
type ColumnarScene struct {
	W       int      `json:"w"`
	H2      int      `json:"h"`
	Strings []string `json:"strings"` // interned string table; index 0 == ""

	BoxComm   []int `json:"boxComm"`
	BoxLabel  []int `json:"boxLabel"`
	BoxCount  []int `json:"boxCount"`
	BoxX      []int `json:"boxX"`
	BoxY      []int `json:"boxY"`
	BoxW      []int `json:"boxW"`
	BoxH      []int `json:"boxH"`
	BoxBorder []int `json:"boxBorder"`

	PinNode  []int `json:"pinNode"`
	PinLabel []int `json:"pinLabel"`
	PinComm  []int `json:"pinComm"`
	PinX     []int `json:"pinX"`
	PinY     []int `json:"pinY"`

	// Points for item i live in Pts[Off[i]*2 : Off[i+1]*2] as [x0,y0,x1,y1,...].
	BundleFrom  []int `json:"bundleFrom"`
	BundleTo    []int `json:"bundleTo"`
	BundleCount []int `json:"bundleCount"`
	BundleOff   []int `json:"bundleOff"`
	BundlePts   []int `json:"bundlePts"`

	ArcFrom []int `json:"arcFrom"`
	ArcTo   []int `json:"arcTo"`
	ArcConf []int `json:"arcConf"`
	ArcOff  []int `json:"arcOff"`
	ArcPts  []int `json:"arcPts"`
}

// interner builds a deterministic interned string table. The empty string is
// always index 0. Strings are assigned ids in first-seen order, deterministic
// because the Scene emit order is.
type interner struct {
	idx map[string]int
	tab []string
}

func newInterner() *interner {
	in := &interner{idx: map[string]int{}}
	in.intern("") // reserve index 0
	return in
}

func (in *interner) intern(s string) int {
	if id, ok := in.idx[s]; ok {
		return id
	}
	id := len(in.tab)
	in.idx[s] = id
	in.tab = append(in.tab, s)
	return id
}

// toColumnar flattens a layout.Scene into the compact columnar form.
func toColumnar(sc layout.Scene) ColumnarScene {
	in := newInterner()
	cs := ColumnarScene{W: sc.W, H2: sc.H}

	for _, b := range sc.Boxes {
		cs.BoxComm = append(cs.BoxComm, b.CommID)
		cs.BoxLabel = append(cs.BoxLabel, in.intern(b.Label))
		cs.BoxCount = append(cs.BoxCount, b.Count)
		cs.BoxX = append(cs.BoxX, b.X)
		cs.BoxY = append(cs.BoxY, b.Y)
		cs.BoxW = append(cs.BoxW, b.W)
		cs.BoxH = append(cs.BoxH, b.H)
		cs.BoxBorder = append(cs.BoxBorder, b.Border)
	}
	for _, p := range sc.Pins {
		cs.PinNode = append(cs.PinNode, in.intern(p.NodeID))
		cs.PinLabel = append(cs.PinLabel, in.intern(p.Label))
		cs.PinComm = append(cs.PinComm, p.CommID)
		cs.PinX = append(cs.PinX, p.X)
		cs.PinY = append(cs.PinY, p.Y)
	}

	cs.BundleOff = append(cs.BundleOff, 0)
	for _, bn := range sc.Bundles {
		cs.BundleFrom = append(cs.BundleFrom, bn.FromComm)
		cs.BundleTo = append(cs.BundleTo, bn.ToComm)
		cs.BundleCount = append(cs.BundleCount, bn.Count)
		for _, pt := range bn.Pts {
			cs.BundlePts = append(cs.BundlePts, pt[0], pt[1])
		}
		cs.BundleOff = append(cs.BundleOff, len(cs.BundlePts)/2)
	}

	cs.ArcOff = append(cs.ArcOff, 0)
	for _, ar := range sc.Arcs {
		cs.ArcFrom = append(cs.ArcFrom, ar.FromComm)
		cs.ArcTo = append(cs.ArcTo, ar.ToComm)
		cs.ArcConf = append(cs.ArcConf, in.intern(ar.Confidence))
		for _, pt := range ar.Pts {
			cs.ArcPts = append(cs.ArcPts, pt[0], pt[1])
		}
		cs.ArcOff = append(cs.ArcOff, len(cs.ArcPts)/2)
	}

	cs.Strings = in.tab
	return cs
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/render/ -run TestToColumnar -v`
Expected: PASS (2 tests).

- [ ] **Step 5: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/render/columnar.go internal/render/columnar_test.go
git commit -m "feat: columnar scene encoder (interned strings + parallel int arrays)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: `internal/render/viewer` — embedded light-theme CSS + vanilla-ES Canvas2D renderer (Tier 1 + click-to-inspect)

**Files:**
- Create: `internal/render/viewer/app.css`
- Create: `internal/render/viewer/app.js`

> These two files are **embedded asset content** inlined verbatim into `map.html` (Task 4) and hashed for the CSP. They are NOT Go code and are NOT unit-tested directly (spec §8.1: zero third-party JS, no test harness for the offline file). All correctness is enforced against the emitted artifact in Task 4's tests. The renderer ships **Tier 1 (Districts) + click-to-inspect only**; semantic-zoom Tiers 2/3 and click-an-arrow-to-list-edges are a documented fast-follow (Open Issue #1).
>
> **Hard constraints the file MUST satisfy (enforced by Task 4 tests):** no `import`, no `export`, no `eval`, no `new Function`, no `http://`/`https://`/`src=`/`<link`/`@import`/`scheme://`, no external fonts (system stack only). The script reads the data island via `getElementById("graffiti-data").textContent` + `JSON.parse` (never a JS literal, never `eval`). All DOM text comes from `textContent` (never `innerHTML`) so labels can never inject markup at runtime.

- [ ] **Step 1: Write the embedded CSS**

Create `/Users/mylive/project/graffiti/graffiti/internal/render/viewer/app.css`:
```css
:root{color-scheme:light}
*{box-sizing:border-box}
html,body{margin:0;height:100%}
body{background:#f6f7f9;color:#1c2330;font:13px/1.5 system-ui,-apple-system,Segoe UI,Roboto,sans-serif}
#top{position:fixed;top:0;left:0;right:0;height:40px;display:flex;align-items:center;gap:12px;padding:0 14px;background:#fff;border-bottom:1px solid #dfe3ea;z-index:3}
#top .title{font-weight:700}
#top .cost{color:#1a7f4b;font-weight:600}
#wrap{position:fixed;top:40px;left:0;right:0;bottom:0;display:flex}
#rail{width:300px;flex:0 0 300px;overflow:auto;padding:12px;background:#fff;border-right:1px solid #dfe3ea}
#rail h2{font-size:11px;letter-spacing:.06em;text-transform:uppercase;color:#6b7480;margin:16px 0 6px}
#rail h2:first-child{margin-top:0}
#search{width:100%;padding:7px 9px;border:1px solid #cfd5de;border-radius:6px;font:inherit}
#rail ol,#rail ul{margin:0;padding-left:18px}
#rail li{margin:4px 0}
#rail .chip{display:block;width:100%;text-align:left;border:1px solid #cfd5de;background:#fbfcfe;border-radius:6px;padding:6px 8px;margin:4px 0;cursor:pointer;font:inherit;color:inherit}
#rail .chip:hover{background:#eef3fb}
#legend .row{display:flex;align-items:center;gap:8px;margin:3px 0}
#legend .swatch{width:22px;height:0;border-top-width:2px;border-top-style:solid;border-color:#1c2330}
#legend .extracted{border-top-style:solid}
#legend .inferred{border-top-style:solid;opacity:.55}
#legend .ambiguous{border-top-style:dashed;opacity:.55}
#stage{position:relative;flex:1 1 auto;overflow:hidden}
#canvas{display:block;width:100%;height:100%;cursor:grab}
#canvas.drag{cursor:grabbing}
#inspect{position:absolute;right:12px;top:12px;width:240px;max-height:60%;overflow:auto;background:#fff;border:1px solid #cfd5de;border-radius:8px;padding:10px 12px;box-shadow:0 4px 14px rgba(20,30,50,.12);display:none}
#inspect.show{display:block}
#inspect h3{margin:0 0 4px;font-size:14px}
#inspect .meta{color:#6b7480}
#a11y{position:absolute;width:1px;height:1px;overflow:hidden;clip:rect(0 0 0 0);white-space:nowrap}
```

- [ ] **Step 2: Write the embedded vanilla-ES Canvas2D renderer**

Create `/Users/mylive/project/graffiti/graffiti/internal/render/viewer/app.js`. The renderer: parses the data island, rebuilds the scene from the columnar arrays, draws the Tier-1 district city-map on a single HiDPI Canvas2D layer (batched paint, redraw only on camera change via `requestAnimationFrame` + dirty flag), supports camera pan/zoom, wires the left rail (search filter, 3 questions, landmarks, confidence legend), and click-to-inspect a box. No imports, no `eval`, no external refs.
```js
"use strict";
(function () {
  // ---- 1. Load + rebuild the scene from the columnar data island. ----
  function loadData() {
    var el = document.getElementById("graffiti-data");
    if (!el) return null;
    try { return JSON.parse(el.textContent); } catch (e) { return null; }
  }
  function str(d, i) { return (d.strings && d.strings[i]) || ""; }
  function pts(flat, off, i) {
    var a = off[i] * 2, b = off[i + 1] * 2, out = [];
    for (var k = a; k < b; k += 2) out.push([flat[k], flat[k + 1]]);
    return out;
  }
  function rebuild(d) {
    var i, boxes = [], pins = [], bundles = [], arcs = [];
    for (i = 0; i < d.boxComm.length; i++) {
      boxes.push({ comm: d.boxComm[i], label: str(d, d.boxLabel[i]), count: d.boxCount[i],
        x: d.boxX[i], y: d.boxY[i], w: d.boxW[i], h: d.boxH[i], border: d.boxBorder[i] });
    }
    for (i = 0; i < d.pinComm.length; i++) {
      pins.push({ node: str(d, d.pinNode[i]), label: str(d, d.pinLabel[i]),
        comm: d.pinComm[i], x: d.pinX[i], y: d.pinY[i] });
    }
    for (i = 0; i < d.bundleFrom.length; i++) {
      bundles.push({ from: d.bundleFrom[i], to: d.bundleTo[i], count: d.bundleCount[i],
        pts: pts(d.bundlePts, d.bundleOff, i) });
    }
    for (i = 0; i < d.arcFrom.length; i++) {
      arcs.push({ from: d.arcFrom[i], to: d.arcTo[i], conf: str(d, d.arcConf[i]),
        pts: pts(d.arcPts, d.arcOff, i) });
    }
    return { w: d.w, h: d.h, boxes: boxes, pins: pins, bundles: bundles, arcs: arcs };
  }

  var data = loadData();
  if (!data) return;
  var scene = rebuild(data);

  // ---- 2. Canvas + HiDPI + camera (pan/zoom). No layout, no physics. ----
  var canvas = document.getElementById("canvas");
  var stage = document.getElementById("stage");
  if (!canvas || !stage) return;
  var ctx = canvas.getContext("2d");
  var dpr = window.devicePixelRatio || 1;
  var cam = { x: 0, y: 0, scale: 1 }; // world->screen: screen = (world - cam)/?, see toScreen
  var dirty = true;

  function fit() {
    var vw = stage.clientWidth, vh = stage.clientHeight;
    canvas.width = Math.round(vw * dpr);
    canvas.height = Math.round(vh * dpr);
    var s = Math.min(vw / scene.w, vh / scene.h);
    cam.scale = s;
    cam.x = (scene.w * s - vw) / 2 / s;
    cam.y = (scene.h * s - vh) / 2 / s;
    dirty = true;
  }
  function toScreen(wx, wy) { return [(wx - cam.x) * cam.scale, (wy - cam.y) * cam.scale]; }
  function toWorld(sx, sy) { return [sx / cam.scale + cam.x, sy / cam.scale + cam.y]; }

  // colorblind-safe confidence stroke styles (line style + alpha) per spec §8.5.
  function strokeForConf(conf) {
    if (conf === "AMBIGUOUS") return { dash: [6, 5], alpha: 0.55, width: 1 };
    if (conf === "INFERRED") return { dash: [], alpha: 0.55, width: 1 };
    return { dash: [], alpha: 1, width: 2 }; // EXTRACTED
  }

  var hot = null; // currently inspected box

  function draw() {
    var vw = stage.clientWidth, vh = stage.clientHeight;
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx.clearRect(0, 0, vw, vh);

    // 2a. Bundled flow-arrows (batched stroke; thickness = bundled count).
    ctx.strokeStyle = "#8893a5";
    var i, j, p;
    for (i = 0; i < scene.bundles.length; i++) {
      var bn = scene.bundles[i];
      ctx.lineWidth = Math.max(1, Math.min(8, Math.log2(bn.count + 1) * 2));
      ctx.beginPath();
      for (j = 0; j < bn.pts.length; j++) {
        p = toScreen(bn.pts[j][0], bn.pts[j][1]);
        if (j === 0) ctx.moveTo(p[0], p[1]); else ctx.lineTo(p[0], p[1]);
      }
      ctx.stroke();
    }

    // 2b. Surprising arcs (dashed, brightly-tinted).
    for (i = 0; i < scene.arcs.length; i++) {
      var ar = scene.arcs[i], st = strokeForConf(ar.conf);
      ctx.save();
      ctx.globalAlpha = st.alpha;
      ctx.strokeStyle = "#c0392b";
      ctx.lineWidth = st.width;
      ctx.setLineDash([6, 5]);
      ctx.beginPath();
      for (j = 0; j < ar.pts.length; j++) {
        p = toScreen(ar.pts[j][0], ar.pts[j][1]);
        if (j === 0) ctx.moveTo(p[0], p[1]); else ctx.lineTo(p[0], p[1]);
      }
      ctx.stroke();
      ctx.restore();
    }

    // 2c. District boxes (area = size, border weight = centrality).
    ctx.textBaseline = "top";
    ctx.font = "12px system-ui,sans-serif";
    for (i = 0; i < scene.boxes.length; i++) {
      var b = scene.boxes[i];
      var tl = toScreen(b.x, b.y);
      var w = b.w * cam.scale, h = b.h * cam.scale;
      ctx.fillStyle = (hot && hot.comm === b.comm) ? "#dbe7fb" : "#eef2f8";
      ctx.fillRect(tl[0], tl[1], w, h);
      ctx.strokeStyle = "#5b6b86";
      ctx.lineWidth = b.border;
      ctx.setLineDash([]);
      ctx.strokeRect(tl[0], tl[1], w, h);
      if (w > 46 && h > 22) {
        ctx.fillStyle = "#1c2330";
        ctx.fillText(b.label + " (" + b.count + ")", tl[0] + 6, tl[1] + 5);
      }
    }

    // 2d. God-node landmark pins (starred halo).
    for (i = 0; i < scene.pins.length; i++) {
      var pn = scene.pins[i], c = toScreen(pn.x, pn.y);
      ctx.fillStyle = "rgba(226,179,64,.35)";
      ctx.beginPath(); ctx.arc(c[0], c[1], 9, 0, Math.PI * 2); ctx.fill();
      ctx.fillStyle = "#d99a13";
      ctx.beginPath(); ctx.arc(c[0], c[1], 5, 0, Math.PI * 2); ctx.fill();
    }
    dirty = false;
  }

  function frame() { if (dirty) draw(); requestAnimationFrame(frame); }

  // ---- 3. Interaction: pan, zoom, click-to-inspect. ----
  var dragging = false, lastX = 0, lastY = 0;
  canvas.addEventListener("mousedown", function (e) {
    dragging = true; lastX = e.clientX; lastY = e.clientY; canvas.classList.add("drag");
  });
  window.addEventListener("mouseup", function () { dragging = false; canvas.classList.remove("drag"); });
  window.addEventListener("mousemove", function (e) {
    if (!dragging) return;
    cam.x -= (e.clientX - lastX) / cam.scale;
    cam.y -= (e.clientY - lastY) / cam.scale;
    lastX = e.clientX; lastY = e.clientY; dirty = true;
  });
  canvas.addEventListener("wheel", function (e) {
    e.preventDefault();
    var r = canvas.getBoundingClientRect();
    var sx = e.clientX - r.left, sy = e.clientY - r.top;
    var before = toWorld(sx, sy);
    var f = e.deltaY < 0 ? 1.1 : 1 / 1.1;
    cam.scale = Math.max(0.2, Math.min(8, cam.scale * f));
    var after = toWorld(sx, sy);
    cam.x += before[0] - after[0]; cam.y += before[1] - after[1];
    dirty = true;
  }, { passive: false });

  function boxAt(sx, sy) {
    var w = toWorld(sx, sy);
    for (var i = 0; i < scene.boxes.length; i++) {
      var b = scene.boxes[i];
      if (w[0] >= b.x && w[0] <= b.x + b.w && w[1] >= b.y && w[1] <= b.y + b.h) return b;
    }
    return null;
  }
  var inspect = document.getElementById("inspect");
  function showInspect(b) {
    hot = b; dirty = true;
    if (!inspect) return;
    if (!b) { inspect.className = ""; return; }
    inspect.textContent = "";
    var h = document.createElement("h3"); h.textContent = b.label; inspect.appendChild(h);
    var m = document.createElement("div"); m.className = "meta";
    m.textContent = b.count + " things · centrality " + b.border; inspect.appendChild(m);
    inspect.className = "show";
  }
  canvas.addEventListener("click", function (e) {
    var r = canvas.getBoundingClientRect();
    showInspect(boxAt(e.clientX - r.left, e.clientY - r.top));
  });

  // ---- 4. Left rail: search filter + landmark fly-to. ----
  var search = document.getElementById("search");
  if (search) {
    search.addEventListener("input", function () {
      var q = search.value.toLowerCase();
      var hit = null;
      for (var i = 0; i < scene.boxes.length; i++) {
        if (q && scene.boxes[i].label.toLowerCase().indexOf(q) >= 0) { hit = scene.boxes[i]; break; }
      }
      if (hit) flyTo(hit);
    });
  }
  function flyTo(b) {
    var vw = stage.clientWidth, vh = stage.clientHeight;
    cam.scale = Math.max(0.5, Math.min(4, Math.min(vw / (b.w * 2), vh / (b.h * 2))));
    cam.x = b.x + b.w / 2 - vw / 2 / cam.scale;
    cam.y = b.y + b.h / 2 - vh / 2 / cam.scale;
    showInspect(b);
  }
  // landmark / district chips emitted by the Go side carry data-comm.
  var chips = document.querySelectorAll("[data-comm]");
  for (var ci = 0; ci < chips.length; ci++) {
    (function (el) {
      el.addEventListener("click", function () {
        var cid = parseInt(el.getAttribute("data-comm"), 10);
        for (var i = 0; i < scene.boxes.length; i++) {
          if (scene.boxes[i].comm === cid) { flyTo(scene.boxes[i]); break; }
        }
      });
    })(chips[ci]);
  }

  window.addEventListener("resize", fit);
  fit();
  requestAnimationFrame(frame);
})();
```

- [ ] **Step 3: Sanity-check the asset files are well-formed and constraint-clean**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
node --check internal/render/viewer/app.js 2>/dev/null && echo "js parses (if node present)" || echo "node not present; rely on Task-4 emitted-artifact tests"
grep -nE 'import |export |eval\(|new Function|https?://|src=|<link|@import' internal/render/viewer/app.js internal/render/viewer/app.css && echo "VIOLATION FOUND" || echo "no forbidden constructs"
wc -c internal/render/viewer/app.js
```
Expected: no forbidden constructs; `app.js` size roughly 8–16 KB (well under the ~25–30KB ceiling; Task-4 tests enforce CSP correctness regardless). `node --check` is optional (`node` may be absent — the emitted-artifact tests in Task 4 are the real gate).

- [ ] **Step 4: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/render/viewer/app.css internal/render/viewer/app.js
git commit -m "feat: embedded light-theme Canvas2D Tier-1 districts renderer (viewer assets)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: `internal/render` — `WriteMapHTML` self-contained CSP-safe emitter (validated htmlemit approach)

**Files:**
- Create: `internal/render/maphtml.go`
- Create: `internal/render/maphtml_test.go`

- [ ] **Step 1: Write the failing map.html test**

Create `/Users/mylive/project/graffiti/graffiti/internal/render/maphtml_test.go`:
```go
package render

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/analyze"
	"github.com/evgeniy-achin/graffiti/internal/graph"
	"github.com/evgeniy-achin/graffiti/internal/layout"
)

// sampleClustered builds a small clustered doc + analysis + baked scene.
func sampleClustered(t *testing.T) (*graph.Document, analyze.Analysis, layout.Scene) {
	t.Helper()
	doc := graph.NewDocument("demo")
	doc.GeneratedAt = "2026-06-15T00:00:00Z"
	doc.Nodes = []graph.Node{
		{ID: "auth-login", Label: "Login", Kind: graph.KindFunction, File: "internal/auth/login.go", Line: 4, Community: 0},
		{ID: "auth-session", Label: "Session", Kind: graph.KindClass, File: "internal/auth/session.go", Line: 2, Community: 0},
		{ID: "http-route", Label: "Route", Kind: graph.KindFunction, File: "internal/http/route.go", Line: 9, Community: 1},
	}
	doc.Edges = []graph.Edge{
		{From: "http-route", To: "auth-login", Relation: graph.RelCalls, Confidence: graph.ConfInferred},
		{From: "auth-login", To: "auth-session", Relation: graph.RelCalls, Confidence: graph.ConfExtracted},
	}
	doc.Communities = []graph.Community{
		{ID: 0, Label: "Auth", Members: []string{"auth-login", "auth-session"}},
		{ID: 1, Label: "Http", Members: []string{"http-route"}},
	}
	an := analyze.Analyze(doc, analyze.Degrees(doc))
	return doc, an, layout.Layout(doc, an)
}

func extractBetween(t *testing.T, html, open, close string) string {
	t.Helper()
	i := strings.Index(html, open)
	if i < 0 {
		t.Fatalf("open marker %q not found", open)
	}
	i += len(open)
	j := strings.Index(html[i:], close)
	if j < 0 {
		t.Fatalf("close marker %q not found", close)
	}
	return html[i : i+j]
}
func recompSha256B64(s string) string {
	sum := sha256.Sum256([]byte(s))
	return base64.StdEncoding.EncodeToString(sum[:])
}
func lineContaining(t *testing.T, lines []string, sub string) string {
	t.Helper()
	for _, l := range lines {
		if strings.Contains(l, sub) {
			return l
		}
	}
	t.Fatalf("no line containing %q", sub)
	return ""
}

// --- 1. Determinism: same generated_at => byte-identical. ---
func TestMapHTML_DeterministicSameTime(t *testing.T) {
	doc, an, sc := sampleClustered(t)
	a := RenderMapHTML(doc, an, sc, "2026-06-15T00:00:00Z")
	b := RenderMapHTML(doc, an, sc, "2026-06-15T00:00:00Z")
	if a != b {
		t.Fatalf("not byte-identical for identical inputs (len %d vs %d)", len(a), len(b))
	}
}

// --- 1b. Determinism: different generated_at => identical EXCEPT the two
// generated_at-bearing lines; CSP hashes unchanged. ---
func TestMapHTML_DiffersOnlyByGeneratedAt(t *testing.T) {
	doc, an, sc := sampleClustered(t)
	t1, t2 := "2026-06-15T00:00:00Z", "1999-12-31T23:59:59Z"
	a := RenderMapHTML(doc, an, sc, t1)
	b := RenderMapHTML(doc, an, sc, t2)
	if a == b {
		t.Fatal("expected difference when generated_at changes")
	}
	la, lb := strings.Split(a, "\n"), strings.Split(b, "\n")
	if len(la) != len(lb) {
		t.Fatalf("line count differs: %d vs %d", len(la), len(lb))
	}
	var diff []int
	for i := range la {
		if la[i] != lb[i] {
			diff = append(diff, i)
		}
	}
	if len(diff) != 2 {
		t.Fatalf("expected exactly 2 differing lines, got %d: %v", len(diff), diff)
	}
	for _, i := range diff {
		if !strings.Contains(la[i], "generated") {
			t.Fatalf("differing line %d is not a generated_at carrier: %q", i, la[i])
		}
	}
	if lineContaining(t, la, "Content-Security-Policy") != lineContaining(t, lb, "Content-Security-Policy") {
		t.Fatal("CSP line changed across generated_at")
	}
}

// --- 2. CSP correctness: independently recompute the inlined bodies' hashes. ---
func TestMapHTML_CSPHashesMatchInlinedBodies(t *testing.T) {
	doc, an, sc := sampleClustered(t)
	html := RenderMapHTML(doc, an, sc, "2026-06-15T00:00:00Z")
	cspLine := lineContaining(t, strings.Split(html, "\n"), "Content-Security-Policy")

	styleBody := extractBetween(t, html, "<style>", "</style>")
	scriptBody := extractBetween(t, html, "<script>", "</script>")
	dataBody := extractBetween(t, html, `<script type="application/json" id="graffiti-data">`, "</script>")

	for _, want := range []string{
		"sha256-" + recompSha256B64(styleBody),
		"sha256-" + recompSha256B64(scriptBody),
		"sha256-" + recompSha256B64(dataBody),
	} {
		if !strings.Contains(cspLine, want) {
			t.Fatalf("hash %q not in CSP: %s", want, cspLine)
		}
	}
	for _, must := range []string{"default-src 'none'", "style-src 'sha256-", "img-src data:"} {
		if !strings.Contains(cspLine, must) {
			t.Fatalf("CSP missing directive %q: %s", must, cspLine)
		}
	}
}

// --- 3. Self-containment: no external refs anywhere. ---
func TestMapHTML_SelfContained(t *testing.T) {
	doc, an, sc := sampleClustered(t)
	html := RenderMapHTML(doc, an, sc, "2026-06-15T00:00:00Z")
	for _, b := range []string{"http://", "https://", "src=", "<link", "@import"} {
		if strings.Contains(html, b) {
			t.Fatalf("self-containment violated: found %q", b)
		}
	}
	if regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9+.-]*://`).MatchString(html) {
		t.Fatal("self-containment violated: found a scheme:// URL")
	}
}

// --- 4. Data island parses to the columnar shape. ---
func TestMapHTML_DataIslandParses(t *testing.T) {
	doc, an, sc := sampleClustered(t)
	html := RenderMapHTML(doc, an, sc, "2026-06-15T00:00:00Z")
	island := extractBetween(t, html, `<script type="application/json" id="graffiti-data">`, "</script>")
	var cs ColumnarScene
	if err := json.Unmarshal([]byte(island), &cs); err != nil {
		t.Fatalf("data island did not parse as JSON: %v", err)
	}
	if cs.W != sc.W || cs.H2 != sc.H {
		t.Fatalf("canvas dims mismatch: %dx%d want %dx%d", cs.W, cs.H2, sc.W, sc.H)
	}
	if len(cs.BoxComm) != len(sc.Boxes) {
		t.Fatalf("boxes in island = %d, want %d", len(cs.BoxComm), len(sc.Boxes))
	}
	if len(cs.Strings) == 0 || cs.Strings[0] != "" {
		t.Fatalf("string table must start with \"\"")
	}
}

// --- 5. Structural presence: <canvas>, a11y mirror lists every district, 3 Qs. ---
func TestMapHTML_StructuralPresence(t *testing.T) {
	doc, an, sc := sampleClustered(t)
	html := RenderMapHTML(doc, an, sc, "2026-06-15T00:00:00Z")
	for _, must := range []string{`<canvas id="canvas"`, `id="a11y"`, `id="graffiti-data"`, "Start here", "Landmarks", "Confidence"} {
		if !strings.Contains(html, must) {
			t.Fatalf("missing structural element %q", must)
		}
	}
	// a11y mirror lists every district label.
	mirror := extractBetween(t, html, `<nav id="a11y"`, "</nav>")
	for _, c := range doc.Communities {
		if !strings.Contains(mirror, c.Label) {
			t.Fatalf("a11y mirror missing district %q", c.Label)
		}
	}
	// exactly the 3 questions appear (as <li> items in the Start here block).
	for _, q := range an.Questions {
		if !strings.Contains(html, htmlEscape(q)) {
			t.Fatalf("question missing from map.html: %q", q)
		}
	}
}

// --- 6. XSS escaping of labels AND the </script island escape. ---
func TestMapHTML_EscapesLabelsAndScriptClose(t *testing.T) {
	doc := graph.NewDocument("demo")
	doc.GeneratedAt = "2026-06-15T00:00:00Z"
	// A malicious label containing both an HTML tag and a script-close sequence.
	evil := `</script><img src=x onerror=alert(1)>`
	doc.Nodes = []graph.Node{
		{ID: "n0", Label: evil, Kind: graph.KindFunction, File: "a.go", Line: 1, Community: 0},
	}
	doc.Communities = []graph.Community{{ID: 0, Label: evil, Members: []string{"n0"}}}
	an := analyze.Analyze(doc, analyze.Degrees(doc))
	sc := layout.Layout(doc, an)
	html := RenderMapHTML(doc, an, sc, "2026-06-15T00:00:00Z")

	// The raw closing-script sequence must not appear inside the JSON island.
	island := extractBetween(t, html, `<script type="application/json" id="graffiti-data">`, "</script>")
	if strings.Contains(island, "</script") {
		t.Fatalf("data island contains an unescaped </script sequence:\n%s", island)
	}
	if !strings.Contains(island, `<`) {
		t.Fatalf("expected '<' to be escaped to \\u003c in the island")
	}
	// The island still parses (escape is JSON-legal) and round-trips the label.
	var cs ColumnarScene
	if err := json.Unmarshal([]byte(island), &cs); err != nil {
		t.Fatalf("escaped island did not parse: %v", err)
	}
	found := false
	for _, s := range cs.Strings {
		if s == evil {
			found = true
		}
	}
	if !found {
		t.Fatalf("evil label did not round-trip through the island")
	}
	// In the visible HTML body (rail/a11y mirror) the label must be HTML-escaped:
	// no raw "<img" and no raw onerror attribute should survive.
	body := html[strings.Index(html, "<body"):]
	if strings.Contains(body, "<img src=x onerror") {
		t.Fatalf("label was not HTML-escaped in the body (XSS):\n%s", body)
	}
}

// --- 7. Writer writes next to map.json/MAP.md. ---
func TestWriteMapHTML_WritesNextToJSON(t *testing.T) {
	doc, an, sc := sampleClustered(t)
	dir := t.TempDir()
	if err := WriteMapHTML(doc, an, sc, dir); err != nil {
		t.Fatalf("WriteMapHTML: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, ".graffiti", "map.html"))
	if err != nil {
		t.Fatalf("read map.html: %v", err)
	}
	if !strings.Contains(string(b), `<canvas id="canvas"`) {
		t.Fatalf("written map.html missing <canvas>")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/render/ -run 'TestMapHTML|TestWriteMapHTML' -v`
Expected: FAIL — `undefined: RenderMapHTML`, `undefined: WriteMapHTML`, `undefined: htmlEscape`.

- [ ] **Step 3: Write the HTML emitter (validated htmlemit approach, embedded assets + `</script` escape)**

Create `/Users/mylive/project/graffiti/graffiti/internal/render/maphtml.go`:
```go
package render

import (
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/evgeniy-achin/graffiti/internal/analyze"
	"github.com/evgeniy-achin/graffiti/internal/graph"
	"github.com/evgeniy-achin/graffiti/internal/layout"
)

// Viewer assets are inlined verbatim into map.html and hashed for the CSP. They
// are pure data here so the emitted sha256 hashes are a function only of these
// (deterministic) bytes — never of generated_at or scene content.
//
//go:embed viewer/app.css
var viewerCSS string

//go:embed viewer/app.js
var viewerJS string

// Keep embed referenced even if a future refactor drops one of the strings.
var _ embed.FS

// htmlEscape escapes interpolated labels/paths for safe inlining into HTML body
// text (XSS-safe self-contained file, spec §8.1/§5/§6).
func htmlEscape(s string) string { return html.EscapeString(s) }

// escapeScriptClose makes a JSON byte string safe to inline inside a
// <script type="application/json"> island: it escapes every '<' to the JSON
// unicode escape <, which both (a) prevents a "</script" sequence in any
// label/path from prematurely closing the island and (b) is fully JSON-legal so
// the island still round-trips via JSON.parse. Applied to the bytes BEFORE
// hashing, so the data-island sha256 stays a pure function of the scene.
func escapeScriptClose(jsonBytes []byte) string {
	return strings.ReplaceAll(string(jsonBytes), "<", `<`)
}

func sha256b64(s string) string {
	sum := sha256.Sum256([]byte(s))
	return base64.StdEncoding.EncodeToString(sum[:])
}

// WriteMapHTML renders map.html and writes it to <root>/.graffiti/map.html, next
// to map.json and MAP.md. generatedAt is read off doc.GeneratedAt (single source
// of truth, stamped by build.Assemble).
func WriteMapHTML(doc *graph.Document, an analyze.Analysis, scene layout.Scene, root string) error {
	dir := filepath.Join(root, ".graffiti")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	out := RenderMapHTML(doc, an, scene, doc.GeneratedAt)
	return os.WriteFile(filepath.Join(dir, "map.html"), []byte(out), 0o644)
}

// RenderMapHTML builds the single self-contained, offline, CSP-safe map.html
// (spec §8.1–§8.8). It is a pure function of (doc, an, scene, generatedAt):
// generatedAt appears ONLY in an HTML comment + a body data-attribute, both
// OUTSIDE every hashed body, so the CSP hashes never vary with time or scene.
// The inlined CSS/JS come from go:embed; the scene ships as a compact columnar
// JSON island with the </script sequence escaped. The hidden ordered <nav id=a11y>
// mirrors the districts for screen readers (spec §8.7).
func RenderMapHTML(doc *graph.Document, an analyze.Analysis, scene layout.Scene, generatedAt string) string {
	cs := toColumnar(scene)
	dataJSON, err := json.Marshal(cs) // deterministic for a fixed struct type
	if err != nil {
		panic(fmt.Sprintf("marshal columnar scene: %v", err)) // ints/strings only
	}
	island := escapeScriptClose(dataJSON) // escape </script BEFORE hashing

	// CSP hashes over the EXACT bytes inlined as each element body.
	scriptHash := sha256b64(viewerJS)
	styleHash := sha256b64(viewerCSS)
	dataHash := sha256b64(island) // belt-and-braces (island isn't executed)

	csp := strings.Join([]string{
		"default-src 'none'",
		fmt.Sprintf("script-src 'sha256-%s' 'sha256-%s'", scriptHash, dataHash),
		fmt.Sprintf("style-src 'sha256-%s'", styleHash),
		"img-src data:",
	}, "; ")

	var b strings.Builder
	w := func(s string) { b.WriteString(s) }

	w("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	w("<meta charset=\"utf-8\">\n")
	w("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	w("<meta http-equiv=\"Content-Security-Policy\" content=\"" + csp + "\">\n")
	w("<title>Graffiti Districts — " + htmlEscape(doc.Root) + "</title>\n")
	w("<!-- generated_at: " + htmlEscape(generatedAt) + " -->\n") // OUTSIDE hashed bodies
	w("<style>" + viewerCSS + "</style>\n")
	w("</head>\n")
	w("<body data-generated-at=\"" + htmlEscape(generatedAt) + "\">\n") // OUTSIDE hashed bodies

	// Top bar (the 0 API calls · $0 promise, spec §8.3).
	w("<div id=\"top\"><span class=\"title\">Districts — " + htmlEscape(doc.Root) + "</span>")
	w("<span class=\"cost\">0 API calls · $0</span></div>\n")

	w("<div id=\"wrap\">\n")
	renderRail(&b, doc, an)
	w("<div id=\"stage\"><canvas id=\"canvas\"></canvas>")
	w("<div id=\"inspect\"></div>")
	renderA11yMirror(&b, doc)
	w("</div>\n") // #stage
	w("</div>\n") // #wrap

	// Data island (data block; not executed). </script escaped above.
	w("<script type=\"application/json\" id=\"graffiti-data\">")
	w(island)
	w("</script>\n")
	w("<script>" + viewerJS + "</script>\n")
	w("</body>\n</html>\n")
	return b.String()
}

// renderRail emits the left rail: search, Start here (3 questions), Landmarks
// (god nodes), Confidence legend. Deterministic; labels HTML-escaped.
func renderRail(b *strings.Builder, doc *graph.Document, an analyze.Analysis) {
	byID := make(map[string]graph.Node, len(doc.Nodes))
	for _, n := range doc.Nodes {
		byID[n.ID] = n
	}
	commOf := func(id string) int { return byID[id].Community }

	b.WriteString("<aside id=\"rail\">\n")
	b.WriteString("<input id=\"search\" type=\"text\" placeholder=\"Search districts…\" aria-label=\"Search districts\">\n")

	b.WriteString("<h2>Start here</h2>\n<ol>\n")
	for _, q := range an.Questions {
		b.WriteString("<li>" + htmlEscape(q) + "</li>\n")
	}
	b.WriteString("</ol>\n")

	b.WriteString("<h2>Landmarks</h2>\n")
	if len(an.GodNodes) == 0 {
		b.WriteString("<p>None.</p>\n")
	} else {
		for _, g := range an.GodNodes {
			fmt.Fprintf(b, "<button class=\"chip\" data-comm=\"%d\">%s — touched by %d things</button>\n",
				commOf(g.ID), htmlEscape(g.Label), g.Degree)
		}
	}

	b.WriteString("<h2>Confidence</h2>\n<div id=\"legend\">\n")
	b.WriteString("<div class=\"row\"><span class=\"swatch extracted\"></span>EXTRACTED — definite</div>\n")
	b.WriteString("<div class=\"row\"><span class=\"swatch inferred\"></span>INFERRED — inferred</div>\n")
	b.WriteString("<div class=\"row\"><span class=\"swatch ambiguous\"></span>AMBIGUOUS — guessed; verify</div>\n")
	b.WriteString("</div>\n")
	b.WriteString("</aside>\n")
}

// renderA11yMirror emits the hidden, ordered DOM mirror of districts (→ members)
// so screen readers and find-in-page work over the Canvas (spec §8.7). Ordered
// by community id; members are already sorted by cluster.NameCommunities.
func renderA11yMirror(b *strings.Builder, doc *graph.Document) {
	byID := make(map[string]graph.Node, len(doc.Nodes))
	for _, n := range doc.Nodes {
		byID[n.ID] = n
	}
	comms := append([]graph.Community(nil), doc.Communities...)
	sort.Slice(comms, func(i, j int) bool { return comms[i].ID < comms[j].ID })

	b.WriteString("<nav id=\"a11y\" aria-label=\"Districts (accessibility mirror)\">\n")
	for _, c := range comms {
		fmt.Fprintf(b, "<section><h3>%s (%d things)</h3>\n<ul>\n", htmlEscape(c.Label), len(c.Members))
		for _, id := range c.Members {
			n := byID[id]
			fmt.Fprintf(b, "<li>%s (%s, %s:%d)</li>\n",
				htmlEscape(n.Label), htmlEscape(string(n.Kind)), htmlEscape(n.File), n.Line)
		}
		b.WriteString("</ul>\n</section>\n")
	}
	b.WriteString("</nav>\n")
}
```

> **Note on the `</script>` extractor in tests vs. the escape:** the data-island body now contains `</script` (escaped), never a literal `</script`, so `extractBetween(..., "</script>")` correctly stops at the *real* closing tag of the island element. The HTML-body `</script>` for the renderer is the next one. The `TestMapHTML_EscapesLabelsAndScriptClose` test asserts the island has no literal `</script` and that `<` is present.

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/render/ -run 'TestMapHTML|TestWriteMapHTML' -v`
Expected: PASS (7 tests: DeterministicSameTime, DiffersOnlyByGeneratedAt, CSPHashesMatchInlinedBodies, SelfContained, DataIslandParses, StructuralPresence, EscapesLabelsAndScriptClose, plus WritesNextToJSON).

- [ ] **Step 5: Run the whole render package (json + mapmd + columnar + maphtml)**

Run: `go test ./internal/render/ -v`
Expected: PASS — existing Plan-1 json tests, Plan-2 mapmd tests, Task-2 columnar tests, Task-4 maphtml tests.

- [ ] **Step 6: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/render/maphtml.go internal/render/maphtml_test.go
git commit -m "feat: self-contained CSP-safe map.html emitter (embedded viewer, columnar island, </script escape)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: Wire `layout → render.WriteMapHTML` into `app.Build`

**Files:**
- Modify: `internal/app/app.go`

- [ ] **Step 1: Add the `layout` import**

In `/Users/mylive/project/graffiti/graffiti/internal/app/app.go`, add the layout import to the existing block so it reads:
```go
import (
	"os"
	"path/filepath"

	"github.com/evgeniy-achin/graffiti/internal/analyze"
	"github.com/evgeniy-achin/graffiti/internal/build"
	"github.com/evgeniy-achin/graffiti/internal/cache"
	"github.com/evgeniy-achin/graffiti/internal/cluster"
	"github.com/evgeniy-achin/graffiti/internal/graph"
	"github.com/evgeniy-achin/graffiti/internal/layout"
	"github.com/evgeniy-achin/graffiti/internal/parse"
	"github.com/evgeniy-achin/graffiti/internal/render"
	"github.com/evgeniy-achin/graffiti/internal/scan"
)
```

- [ ] **Step 2: Bake the scene and write map.html**

In `Build`, after the existing `an := analyze.Analyze(doc, deg)` line and immediately before `render.WriteMapJSON`, add the layout call; then add the `WriteMapHTML` writer after `WriteMapMD`. The post-`Analyze` block becomes:
```go
	// Plan 2: cluster, name communities, analyze — all deterministic, no I/O above render.
	cluster.Cluster(doc)
	deg := analyze.Degrees(doc)
	doc.Communities = cluster.NameCommunities(doc, deg)
	an := analyze.Analyze(doc, deg)

	// Plan 3: bake the deterministic Tier-1 district scene (integer coords).
	scene := layout.Layout(doc, an)

	if err := render.WriteMapJSON(doc, absRoot); err != nil {
		return stats, err
	}
	if err := render.WriteMapMD(doc, an, absRoot); err != nil {
		return stats, err
	}
	if err := render.WriteMapHTML(doc, an, scene, absRoot); err != nil {
		return stats, err
	}
	if err := c.Flush(); err != nil {
		return stats, err
	}
```

> `layout.Layout` runs *after* `analyze.Analyze` and *before* the writers; it mutates nothing (reads `doc` + `an`, returns a `Scene`). `map.json`/`MAP.md` emission is unchanged, so both existing goldens stay byte-identical. `generated_at` remains single-sourced: `WriteMapHTML` reads `doc.GeneratedAt` (stamped once in `build.Assemble`); `time.Now()` is still called exactly once, in `cmd/graffiti/main.go`. No `Stats` change (the success line already prints the 3 questions from Plan 2).

- [ ] **Step 3: Verify the app package compiles + existing app tests pass**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/app/ -run 'TestBuild|TestGolden|TestDeterminism|TestClustering|TestSuggestedQuestions' -v`
Expected: PASS — the Plan-1/Plan-2 app tests are unaffected (map.json + MAP.md goldens unchanged; map.html is an additional artifact not yet asserted here — Task 6 adds those assertions).

- [ ] **Step 4: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/app/app.go
git commit -m "feat: wire layout + WriteMapHTML into app.Build (map.html alongside map.json/MAP.md)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: `map.html` artifact tests in `app` (determinism / self-contained / CSP / structure) + optional strip-golden

**Files:**
- Modify: `internal/app/golden_test.go`
- Create (optional, via UPDATE_GOLDEN): `testdata/golden/gorepo.map.html.strip`

> These tests build the real `gorepo` fixture end-to-end and assert the produced `map.html` against the same invariants Task 4 enforces on synthetic input — proving the *wired* pipeline emits a correct artifact. The byte-determinism check uses the existing two-build pattern with a `generated_at`-strip (mirroring `TestDeterminism_TwoBuildsByteIdentical` for map.json). An optional stripped golden locks byte-exactness.

- [ ] **Step 1: Add map.html helpers + tests to `internal/app/golden_test.go`**

In `/Users/mylive/project/graffiti/graffiti/internal/app/golden_test.go`, the import block already has `os`, `path/filepath`, `regexp`, `strings`, `testing`. Add `crypto/sha256` and `encoding/base64` to it, then append:
```go
// buildFixtureMapHTML builds the fixture into a temp dir and returns the produced
// map.html bytes (companion to buildFixtureIntoTemp / buildFixtureMapMD).
func buildFixtureMapHTML(t *testing.T) []byte {
	t.Helper()
	src := filepath.Join("..", "..", "testdata", "fixtures", "gorepo")
	dst := t.TempDir()
	copyTree(t, src, dst)
	if _, err := Build(dst, fixtureGenAt); err != nil {
		t.Fatalf("Build fixture: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dst, ".graffiti", "map.html"))
	if err != nil {
		t.Fatalf("read produced map.html: %v", err)
	}
	return b
}

// stripGeneratedAt blanks the two generated_at-bearing carriers (HTML comment +
// body data-attribute) so two builds compare byte-equal modulo the timestamp.
var reHTMLComment = regexp.MustCompile(`<!-- generated_at: [^>]*-->`)
var reHTMLDataAttr = regexp.MustCompile(`data-generated-at="[^"]*"`)

func stripHTML(b []byte) []byte {
	b = reHTMLComment.ReplaceAll(b, []byte(`<!-- generated_at: X -->`))
	b = reHTMLDataAttr.ReplaceAll(b, []byte(`data-generated-at="X"`))
	return b
}

func TestMapHTML_TwoBuildsByteIdenticalModuloGeneratedAt(t *testing.T) {
	a := stripHTML(buildFixtureMapHTML(t))
	b := stripHTML(buildFixtureMapHTML(t))
	if string(a) != string(b) {
		t.Fatalf("two map.html builds not byte-identical modulo generated_at")
	}
}

func TestMapHTML_SelfContainedAndCSP(t *testing.T) {
	html := string(buildFixtureMapHTML(t))

	for _, banned := range []string{"http://", "https://", "src=", "<link", "@import"} {
		if strings.Contains(html, banned) {
			t.Fatalf("self-containment violated: found %q", banned)
		}
	}
	if regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9+.-]*://`).MatchString(html) {
		t.Fatalf("self-containment violated: scheme:// URL present")
	}

	// Recompute the inlined style/script hashes and assert they're in the CSP.
	extract := func(open, close string) string {
		i := strings.Index(html, open)
		if i < 0 {
			t.Fatalf("marker %q not found", open)
		}
		i += len(open)
		j := strings.Index(html[i:], close)
		if j < 0 {
			t.Fatalf("close %q not found", close)
		}
		return html[i : i+j]
	}
	h := func(s string) string {
		sum := sha256.Sum256([]byte(s))
		return "sha256-" + base64.StdEncoding.EncodeToString(sum[:])
	}
	var csp string
	for _, l := range strings.Split(html, "\n") {
		if strings.Contains(l, "Content-Security-Policy") {
			csp = l
		}
	}
	if csp == "" {
		t.Fatal("no CSP meta line")
	}
	for _, want := range []string{
		h(extract("<style>", "</style>")),
		h(extract("<script>", "</script>")),
		h(extract(`<script type="application/json" id="graffiti-data">`, "</script>")),
	} {
		if !strings.Contains(csp, want) {
			t.Fatalf("CSP missing hash %q\nCSP=%s", want, csp)
		}
	}
}

func TestMapHTML_StructuralAndA11yMirror(t *testing.T) {
	html := string(buildFixtureMapHTML(t))
	for _, must := range []string{`<canvas id="canvas"`, `<nav id="a11y"`, `id="graffiti-data"`, "Start here", "Landmarks", "Confidence"} {
		if !strings.Contains(html, must) {
			t.Fatalf("map.html missing %q", must)
		}
	}
	// Every clustered community label from the fixture appears in the a11y mirror.
	var doc graph.Document
	if err := json.Unmarshal(buildFixtureIntoTemp(t), &doc); err != nil {
		t.Fatalf("unmarshal map.json: %v", err)
	}
	mi := strings.Index(html, `<nav id="a11y"`)
	mirror := html[mi : mi+strings.Index(html[mi:], "</nav>")]
	for _, c := range doc.Communities {
		if !strings.Contains(mirror, c.Label) {
			t.Fatalf("a11y mirror missing district %q", c.Label)
		}
	}
}

func mapHTMLGoldenPath() string {
	return filepath.Join("..", "..", "testdata", "golden", "gorepo.map.html.strip")
}

// TestGolden_MapHTMLStrip locks the byte-exact map.html modulo generated_at. The
// golden stores the stripped bytes; regenerate via UPDATE_GOLDEN=1.
func TestGolden_MapHTMLStrip(t *testing.T) {
	got := stripHTML(buildFixtureMapHTML(t))
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(mapHTMLGoldenPath(), got, 0o644); err != nil {
			t.Fatalf("write map.html golden: %v", err)
		}
		t.Log("map.html strip golden updated")
		return
	}
	want, err := os.ReadFile(mapHTMLGoldenPath())
	if err != nil {
		t.Fatalf("read map.html golden (run UPDATE_GOLDEN=1 to create): %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("map.html differs from golden (modulo generated_at)")
	}
}
```

> `encoding/json` and `graph` are already imported by `golden_test.go` (used by `TestGolden_StructuralShape` / `TestClustering_StructuralInvariants`), so no extra import beyond `crypto/sha256` + `encoding/base64` is needed.

- [ ] **Step 2: Run the invariant tests (the real gate) before freezing any golden**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/app/ -run 'TestMapHTML' -v
```
Expected: PASS — `TestMapHTML_TwoBuildsByteIdenticalModuloGeneratedAt`, `TestMapHTML_SelfContainedAndCSP`, `TestMapHTML_StructuralAndA11yMirror`. Do **not** regenerate the strip golden until these are green.

- [ ] **Step 3: Generate the strip golden from real output**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
UPDATE_GOLDEN=1 go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/app/ -run TestGolden_MapHTMLStrip -v
```
Expected: PASS with log `map.html strip golden updated`; `testdata/golden/gorepo.map.html.strip` now exists. Open it once to sanity-check: it begins `<!DOCTYPE html>`, has the CSP meta line, a `<canvas id="canvas"`, the data island, the inlined `<style>`/`<script>`, and the `<nav id="a11y">` mirror listing the fixture's districts; the only `generated_at` carriers read `X` (stripped).

- [ ] **Step 4: Run the full app + render + layout suite**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/app/ ./internal/render/ ./internal/layout/ -v
```
Expected: PASS — Plan-1/Plan-2 goldens unchanged (`TestGolden_GoRepoMapJSON`, `TestGolden_MapMD`), plus all Plan-3 layout/render/app tests.

- [ ] **Step 5: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/app/golden_test.go testdata/golden/gorepo.map.html.strip
git commit -m "test: map.html artifact tests (determinism/self-contained/CSP/a11y) + strip golden

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 7: Full-suite verification, vet, size guard, and manual browser smoke

**Files:** none (verification only)

- [ ] **Step 1: Run the full test suite via the Makefile**

Run: `make test`
Expected: all packages PASS (`ok` for cmd/graffiti, internal/analyze, internal/app, internal/build, internal/cache, internal/cluster, internal/graph, internal/layout, internal/parse, internal/render, internal/scan, schemaval). The Makefile already passes the grammar-subset tags.

- [ ] **Step 2: Run vet**

Run: `make vet`
Expected: no output (clean). Note `embed` requires the `//go:embed` directives to reference existing files (`viewer/app.css`, `viewer/app.js`) — vet/build fails loudly if either asset is missing or misnamed.

- [ ] **Step 3: Confirm the binary size guard still holds (embedded assets are small)**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
make xcompile
```
Expected: `size-guard: all binaries within limit OK` — the embedded CSS+JS add only ~10–20 KB, nowhere near the 16 MB limit.

- [ ] **Step 4: Build + smoke-test map.html end-to-end on the fixture**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
make build
rm -rf /tmp/graffiti-p3-smoke && cp -r testdata/fixtures/gorepo /tmp/graffiti-p3-smoke
./graffiti build /tmp/graffiti-p3-smoke
test -f /tmp/graffiti-p3-smoke/.graffiti/map.html && echo "map.html OK"
grep -q '<canvas id="canvas"' /tmp/graffiti-p3-smoke/.graffiti/map.html && echo "canvas OK"
grep -q 'Content-Security-Policy' /tmp/graffiti-p3-smoke/.graffiti/map.html && echo "CSP OK"
grep -q '<nav id="a11y"' /tmp/graffiti-p3-smoke/.graffiti/map.html && echo "a11y mirror OK"
grep -Eq 'https?://|src=|<link|@import' /tmp/graffiti-p3-smoke/.graffiti/map.html && echo "EXTERNAL REF FOUND (FAIL)" || echo "self-contained OK"
```
Expected: the Plan-2 success line (`✓ Done. 0 API calls, $0.` + 3 questions), then `map.html OK`, `canvas OK`, `CSP OK`, `a11y mirror OK`, `self-contained OK`. Exit code 0.

- [ ] **Step 5: Confirm determinism at the binary level (two builds, identical map.html modulo generated_at)**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
rm -rf /tmp/p3a /tmp/p3b
cp -r testdata/fixtures/gorepo /tmp/p3a && cp -r testdata/fixtures/gorepo /tmp/p3b
./graffiti build /tmp/p3a >/dev/null && ./graffiti build /tmp/p3b >/dev/null
strip() { sed -E -e 's/<!-- generated_at: [^>]*-->/<!-- generated_at: X -->/' -e 's/data-generated-at="[^"]*"/data-generated-at="X"/'; }
diff <(strip < /tmp/p3a/.graffiti/map.html) <(strip < /tmp/p3b/.graffiti/map.html) && echo "map.html deterministic"
```
Expected: `map.html deterministic` (empty diff). (Note: `root` is the same basename `gorepo` for both builds, so only `generated_at` can differ.)

- [ ] **Step 6: Manual browser smoke (human-in-the-loop; the only non-automatable check)**

Open `/tmp/graffiti-p3-smoke/.graffiti/map.html` from `file://` in a browser with devtools open. Confirm: it opens offline with **no network requests** and **no CSP violations** in the console; the canvas shows named district boxes (area ∝ size), bundled flow-arrows, god-node pins, and any dashed surprising arc; pan (drag) and zoom (wheel) work; clicking a box shows the inspect panel; the left rail shows the 3 questions, landmarks, and the confidence legend; browser find-in-page (Cmd/Ctrl-F) finds a district label (proving the a11y mirror). This is a smoke check; the byte-level guarantees are already enforced by the Go tests.

- [ ] **Step 7: Final tidy + finish the branch**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
go mod tidy   # no-op: Plan 3 adds no dependencies
make test
git diff --quiet go.mod go.sum || (git add go.mod go.sum && git commit -m "build: go mod tidy (no-op for Plan 3)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>")
```
Expected: `make test` prints `ok` for every package; `go mod tidy` changes nothing (Plan 3 is stdlib-only).

Then use superpowers:finishing-a-development-branch to merge `plan3-districts-viewer` into `main` (or open a PR), per the flow Plans 1 & 2 used.

---

## Self-Review

**1. Spec coverage (Plan 3 scope only):**
- §8.1 self-contained / offline / CSP-safe: one `map.html`, inlined `<style>`/`<script>` from `go:embed`, `application/json` data island read via `getElementById`+`JSON.parse` (never `eval`), `</script` escaped (`<` → `<`), strict `sha256` meta CSP (`default-src 'none'; script-src …; style-src …; img-src data:`), system-font stack, all interpolated labels HTML-escaped → Tasks 3, 4. ✓
- §8.2 Canvas2D, layout baked in Go as integer coords (browser does no layout/physics): `internal/layout` bakes the integer Scene; `app.js` only transforms/culls/paints/hit-tests → Tasks 1, 3. ✓
- §8.3 default view: 8–40 named tinted district boxes (area = size, border = centrality), bundled labeled flow-arrows (thickness = count), god-node landmark pins, dashed surprising arcs, light theme, left rail (search, 3 questions, landmarks, confidence legend), `0 API calls · $0` top bar; district names from the §8.3 heuristic already populated by `cluster.NameCommunities` (Plan 2) → Tasks 1, 3, 4. ✓
- §8.5 encoding: god-node pins (capped ~7 by `analyze`), surprising dashed arcs, confidence line-style legend → Tasks 1, 3, 4. ✓
- §8.6 compact columnar encoding (interned string table + integer arrays + flattened edge/point arrays, sorted) → Task 2. ✓
- §8.7 accessibility: hidden ordered DOM mirror (`<nav id="a11y">`, districts → members) shipped in v1 → Task 4. ✓
- §8.8 determinism: byte-identical `map.html` modulo the single `generated_at`; integer-quantized baked layout; sorted JSON keys/arrays; CSP hashes pure functions of inlined content → Tasks 1, 4, 6. ✓
- Pipeline wiring `analyze → layout → render(map.html)` next to `map.json`/`MAP.md`; both existing goldens unaffected; single `generated_at`/single `time.Now()` preserved → Tasks 5, 6. ✓
- **Explicitly deferred (documented fast-follow, correctly absent):** semantic-zoom Tier 2 (district interior / file shelves, §8.4) and Tier 3 (symbol + callers/callees, §8.4), click-an-arrow-to-list-underlying-edges, and the §8.9 workspace "lanes" layer. v1 ships Tier 1 + click-to-inspect only (Open Issue #1).

**2. Determinism guarantees (spec §8.8/§14), enforced and tested:**
- Layout: integer coords quantized at each treemap row; every emitted Scene slice re-sorted with an explicit total order (boxes by comm id; pins by `(commID, nodeID)`; bundles by `(from, to)`; arcs by `(from, to, conf)`); no `math/rand`, no time, no map-iteration feeds emitted order. Verified by `TestLayout_Deterministic` + the prototype.
- Columnar: string ids assigned in first-seen order over the already-deterministic Scene; `json.Marshal` of a fixed struct type yields stable key order. Verified by `TestToColumnar_Deterministic`.
- HTML: `generated_at` appears only in one HTML comment + one body data-attribute, both outside every hashed body; CSP hashes are digests of the constant embedded CSS/JS and of the `</script`-escaped island (escape applied before hashing). Verified by `TestMapHTML_DeterministicSameTime`, `TestMapHTML_DiffersOnlyByGeneratedAt` (exactly 2 lines differ, CSP identical), and the binary-level diff in Task 7.

**3. Type consistency with the merged Plan-1/Plan-2 code (checked against the real sources):**
- `graph.Node{ID,Label,Kind,File,Line,Community}`, `graph.Edge{From,To,Relation,Confidence}`, `graph.Community{ID,Label,Members}`, `graph.Document{...GeneratedAt,Root,Communities}`, `graph.Kind`/`Relation`/`Confidence` (string-backed; `string(...)` where a plain string is needed) — used exactly as defined in `internal/graph/graph.go`. ✓
- `analyze.Analysis{GodNodes,Surprising,Cycles,Questions}`, `analyze.GodNode{ID,Label,Degree}`, `analyze.SurprisingEdge{From,To,FromComm,ToComm,Relation,Confidence}`, `analyze.Degrees`, `analyze.Analyze` — consumed read-only by `layout` and `render`. ✓
- `layout.Layout(doc *graph.Document, an analyze.Analysis) layout.Scene`; `layout.Scene{W,H,Boxes,Pins,Bundles,Arcs}` with integer-coord `Box/Pin/Bundle/Arc` — Task 1; consumed by `render`. ✓
- `render.toColumnar(layout.Scene) ColumnarScene` (internal); `render.RenderMapHTML(doc, an, scene, generatedAt) string`; `render.WriteMapHTML(doc, an, scene, root) error` reading `doc.GeneratedAt` — Tasks 2, 4. The existing `render.WriteMapJSON`/`WriteMapMD`/`orderedDocument` are untouched, so the two prior goldens stay byte-identical. ✓
- `app.Build(root, generatedAt string) (Stats, error)` signature unchanged; `Stats` unchanged (questions already threaded in Plan 2); pipeline `Assemble → Cluster → NameCommunities → Analyze → Layout → WriteMapJSON → WriteMapMD → WriteMapHTML → Flush`. `cmd/graffiti/main.go` is unchanged. ✓
- Module path `github.com/evgeniy-achin/graffiti` in every import; build tags applied to every command touching `internal/parse` (app, cmd). ✓

**4. Prototype evidence (load-bearing):** the squarified-treemap layout and the self-contained CSP-safe HTML emitter were prototyped end-to-end in a scratch module with passing tests (`/tmp/p3proto/{layout,htmlemit}{,_test}.go`): no-overlap / in-bounds / integer / byte-stable / area-proportional / bundles-aggregated / pins-on-district for layout; determinism / differs-only-by-generated_at / CSP-hashes-match-inlined-bodies / self-contained / data-island-parses for emit. The task code is that validated prototype with the local mirror types replaced by the real `graph.*`/`analyze.*` types and the inlined CSS/JS sourced from `go:embed`. The `</script` escape (`<` → `<`) the prototype flagged is included in `escapeScriptClose` and is covered by `TestMapHTML_EscapesLabelsAndScriptClose`. The renderer JS is embedded asset content (not Go-unit-testable); every test targets the Go-emitted artifact, never JS runtime behavior.

---

## Open Issues (require owner attention)

1. **Semantic-zoom Tiers 2 & 3 are deferred (documented fast-follow).** v1 `map.html` ships Tier 1 (Districts) + click-to-inspect-a-box only. Tier 2 (district interior: file shelves with top symbols, sibling boundary stubs) and Tier 3 (symbol with `file:line`, kind icon, direct callers/callees), plus click-an-arrow-to-list-the-real-underlying-edges, are a later plan: they require baking file-shelf + symbol layout into the Scene (the current Scene carries only community-level geometry) and a richer columnar data island (per-symbol coords, per-bundle underlying edge lists). Deferring keeps v1 legible, byte-deterministic, and within the <1.5 MB budget. **Owner to confirm Tier 1 + click-to-inspect is acceptable for v1** (it satisfies the §8.3 "name what the program does before touching a control" goal).

2. **Fixed `1600×1000` canvas vs. adaptive sizing.** The baked layout uses a fixed canvas so coordinates are byte-stable; the browser scales it to the viewport. For a repo with 40 districts this can make the smallest district boxes label-tight at default zoom (the renderer hides labels below a pixel threshold, and zoom reveals them). An adaptive canvas size (e.g. scaling with district count) would improve small-district legibility but must stay a deterministic pure function of the graph to preserve §8.8. **Owner may want a follow-up to make canvas dimensions a deterministic function of community count** (a one-line change in `layout`, fully covered by the existing determinism test).

3. **Bundle/arc routing is a simple 3-point elbow, not gutter-routed transit-map polylines.** Spec §8.2 describes orthogonal/elbow polylines "routed in the gutters." v1 bakes a center-to-center horizontal-then-vertical elbow (deterministic, integer, never invents a false endpoint — it always connects the true `(A→B)` district centers). True gutter-aware routing (avoiding crossing boxes) is a layout refinement for a later plan; it only changes `Bundle.Pts`/`Arc.Pts` (geometry), never the aggregation count or endpoints, so it cannot violate any invariant this plan's tests assert. **Owner to confirm the elbow look is acceptable for v1.**

4. **Optional strip-golden (`gorepo.map.html.strip`) maintenance.** Task 6 adds a byte-exact stripped golden for `map.html`. Because the golden embeds the full inlined `app.js`/`app.css`, **any edit to the viewer assets requires `UPDATE_GOLDEN=1` to regenerate it** (the structural + CSP + determinism tests still pass without regeneration and are the real correctness gate; the strip golden only locks byte-exactness). Owner may prefer to drop the strip golden and rely solely on the structural/determinism tests if frequent viewer iteration is expected — flagged so the choice is explicit.

---

**Plan complete and save location:** this document lives at `/Users/mylive/project/graffiti/graffiti/docs/superpowers/plans/2026-06-15-graffiti-plan-3-districts-viewer.md`. Execution options: (1) Subagent-Driven (recommended) — superpowers:subagent-driven-development; (2) Inline — superpowers:executing-plans.
