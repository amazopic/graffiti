# graffiti Plan 2 — Clustering, Analysis & `MAP.md` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** After `graffiti build .` assembles the directed graph (Plan 1, merged), the pipeline now runs `cluster → analyze → render(MAP.md)`: a deterministic in-package Louvain assigns every node a contiguous `community` (0..K-1), communities are named and populated into `Document.Communities`, an `analyze` stage derives god nodes / surprising cross-community connections / import cycles / exactly 3 suggested questions, and a deterministic `MAP.md` is written next to `map.json`. The `map.json` golden is regenerated (nodes now carry real community ids; `communities[]` is populated), the success line gains the 3 questions, and everything stays byte-identical modulo `generated_at`/`root`.

**Architecture:** Pure functional stages, plain structs in/out, no shared mutable state — the same shape Plan 1 established. Plan 2 inserts two stages between `build` and the existing `render`:

```
scan → parse → build → cluster → analyze → render(map.json + MAP.md)
```

- `internal/cluster` reads a `*graph.Document` whose `Community` is `-1` and mutates each `Node.Community` to a contiguous `0..K-1` label via a deterministic single-level Louvain over an **undirected, edge-count-weighted projection** of the directed edges. It also names the communities (§8.3 hybrid heuristic) and populates `doc.Communities`.
- `internal/analyze` reads the clustered `*graph.Document` and returns an `Analysis` value (god nodes, surprising connections, import cycles, 3 questions) — read-only, no mutation.
- `internal/render` gains a second emitter `WriteMapMD(doc, an, root)` alongside the existing `WriteMapJSON`.
- `internal/app.Build` wires `cluster.Cluster` then `analyze.Analyze` between `build.Assemble` and the writers, threads the 3 questions out via `Stats`, and the CLI prints them.

Determinism is paramount (spec §14): no Go-map iteration ever feeds an emitted order. Every emitted list (communities, members, god nodes, surprising edges, cycles, questions, MAP.md sections) is produced by sorted iteration with explicit total-order tie-breaks. No `math/rand`. The single `generated_at` source from Plan 1 is preserved (stamped once in `build.Assemble`, read off `doc.GeneratedAt`).

**Tech Stack:** Go 1.26; module `github.com/amazopic/graffiti`; **no new third-party dependencies** (cluster/analyze/render-md use only the stdlib: `sort`, `strings`, `path`, `fmt`). `CGO_ENABLED=0`. The grammar-subset build tags from Plan 1 still apply to any target that links `internal/parse` (i.e. `internal/app` and `cmd/graffiti`), so all `go test`/`go build` commands that touch those packages keep `-tags "grammar_subset grammar_subset_go grammar_subset_gomod"`; the pure `internal/cluster` and `internal/analyze` packages do not need the tags but always passing them is harmless and keeps commands uniform.

---

## Validated Prototype Facts (load-bearing; confirmed by compiling & running real Go in a scratch module on Go 1.26.2)

Before writing this plan the Louvain + naming + analyze + MAP.md were prototyped end-to-end in a throwaway `package main` against a copy of the real `graph` types, and the following were **observed**, not assumed. Treat them as ground truth; the task code below is the corrected-for-`graph.*` version of that validated prototype.

- **Two 4-node cliques joined by one bridge edge separate into exactly two communities.** With nodes `a,b,c,d` (left clique) and `e,f,g,h` (right clique) plus a single `d→e` bridge, the algorithm puts `{a,b,c,d}` in one community and `{e,f,g,h}` in the other (`K=2`).
- **Renumbering by smallest member id works:** node `a` (globally smallest id) lands in community `0`; the right clique becomes community `1`.
- **Clustered modularity beats the singleton partition:** measured singleton modularity `Q=-0.1272`, clustered `Q=+0.4231` on the two-cliques graph. The `Q_clustered >= Q_singleton` invariant holds.
- **Fully deterministic:** two consecutive runs produced byte-identical node→community maps, community labels/members, god-node lists, surprising edges, questions, and MAP.md text — repeated 5× with zero variance.
- **God-node cap is real:** with 8 nodes all of degree ≥ 2, the cap of 7 dropped exactly one node (`h`, the largest id among the lowest-degree tie), confirming the degree-desc / id-asc / cap ordering.
- **Naming heuristic:** the left clique (all files under `left/`) was labeled `"Left"` and the right `"Right"` via the dominant-directory rule (strict-majority directory → title-cased base segment); the most-central-member fallback fires when no directory holds a strict majority.
- **Empty / isolated graphs are safe:** an empty document returns `K=0`; three isolated edge-less nodes each become their own community (`K=3`), every node `Community >= 0`, contiguous.

**Why single-level Louvain (no recursive aggregation) for v1:** the local-moving pass alone optimizes modularity well enough to satisfy the spec's "city map / Districts" goal on the v1 fixture and is dramatically simpler to make provably deterministic (no graph-contraction renumbering ambiguity). Recursive multi-level aggregation is a documented later-plan refinement (Open Issue #2) — it would only ever *merge* communities further, so adding it later cannot break the invariants this plan's tests assert.

**Determinism mechanics (the exact rules the code enforces):**
- Nodes are processed in **ascending node-id order** (a sorted index, independent of slice order or map iteration).
- The undirected projection weight of pair `{u,v}` is the **count of directed edges** between them (any relation), self-loops ignored; parallel edges accumulate.
- Candidate target communities for a node are `{current} ∪ {neighbor communities}`, **sorted ascending**, and the best modularity gain wins; **ties keep the smallest community id** (first maximizer in the sorted candidate list). Gain is `w_to_c - deg_i * commTot_c / m2` (the constant terms equal across candidates are dropped, so no precision is lost to large offsets).
- Final communities are **renumbered by smallest member id** so labels are contiguous `0..K-1` and stable.
- All `float64` arithmetic is over small integer-valued weights (edge counts), so comparisons are stable; **no `math/rand`, no time, no map iteration** feeds any decision or any emitted order.

---

## File Structure

```
graffiti/
├── internal/
│   ├── cluster/
│   │   ├── cluster.go          # Cluster(doc) int — deterministic Louvain, mutates Node.Community
│   │   ├── naming.go           # NameCommunities(doc, deg) []graph.Community — §8.3 hybrid naming
│   │   ├── cluster_test.go     # invariant tests (two-cliques, contiguity, modularity, determinism, empty)
│   │   └── naming_test.go      # naming heuristic tests (dominant dir + central-member fallback)
│   ├── analyze/
│   │   ├── analyze.go          # Analysis, Analyze(doc, deg), Degrees(doc), god nodes/surprising/cycles/questions
│   │   └── analyze_test.go
│   ├── render/
│   │   ├── json.go             # (unchanged from Plan 1)
│   │   ├── mapmd.go            # NEW: WriteMapMD(doc, an, root) + RenderMapMD(doc, an) string
│   │   └── mapmd_test.go       # NEW
│   └── app/
│       ├── app.go              # MODIFIED: wire cluster→analyze→render; Stats.Questions
│       ├── golden_test.go      # MODIFIED: add MAP.md golden + cluster structural assertions
│       └── ...
├── cmd/graffiti/
│   └── main.go                 # MODIFIED: success line prints the 3 questions
└── testdata/golden/
    ├── gorepo.map.json         # REGENERATED (communities populated, node.community set)
    └── gorepo.MAP.md           # NEW golden (regenerated via UPDATE_GOLDEN)
```

**Package responsibilities (one job each):**
- `cluster` owns community detection **and** community naming/population of `doc.Communities` — no I/O, deterministic, stdlib-only. (Naming lives here because §5 lists district naming under the cluster/analyze stage and it needs only the clustered graph + degrees.)
- `analyze` owns derived read-only insights (god nodes, surprising connections, cycles, questions) — no mutation, no I/O.
- `render` gains MAP.md emission; the JSON writer is untouched.
- `app` wires the new stages in; `cmd/graffiti` prints the questions.

---

## Task 1: `internal/cluster` — deterministic Louvain (invariant-tested)

**Files:**
- Create: `internal/cluster/cluster.go`
- Create: `internal/cluster/cluster_test.go`

- [ ] **Step 1: Write the failing cluster invariant test**

Create `/Users/mylive/project/graffiti/graffiti/internal/cluster/cluster_test.go`:
```go
package cluster

import (
	"fmt"
	"testing"

	"github.com/amazopic/graffiti/internal/graph"
)

// twoCliques builds two 4-node cliques joined by a single bridge edge (d->e).
// A correct community detector MUST place each clique in its own community.
func twoCliques() *graph.Document {
	doc := graph.NewDocument("repo")
	ids := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	for _, id := range ids {
		file := "left/" + id + ".go"
		if id >= "e" {
			file = "right/" + id + ".go"
		}
		doc.Nodes = append(doc.Nodes, graph.Node{
			ID: id, Label: id, Kind: graph.KindFunction, File: file, Line: 1,
			Community: graph.UnclusteredCommunity,
		})
	}
	add := func(a, b string) {
		doc.Edges = append(doc.Edges, graph.Edge{From: a, To: b, Relation: graph.RelCalls, Confidence: graph.ConfInferred})
	}
	for _, cl := range [][]string{{"a", "b", "c", "d"}, {"e", "f", "g", "h"}} {
		for i := 0; i < len(cl); i++ {
			for j := i + 1; j < len(cl); j++ {
				add(cl[i], cl[j])
			}
		}
	}
	add("d", "e") // bridge
	return doc
}

func commByID(doc *graph.Document) map[string]int {
	m := map[string]int{}
	for _, n := range doc.Nodes {
		m[n.ID] = n.Community
	}
	return m
}

func TestCluster_TwoCliquesSeparate(t *testing.T) {
	doc := twoCliques()
	k := Cluster(doc)

	comm := commByID(doc)
	for id, c := range comm {
		if c < 0 {
			t.Fatalf("node %s has negative community %d (every node must be >= 0)", id, c)
		}
	}
	for _, x := range []string{"b", "c", "d"} {
		if comm[x] != comm["a"] {
			t.Fatalf("left clique split: %s in %d, a in %d", x, comm[x], comm["a"])
		}
	}
	for _, x := range []string{"f", "g", "h"} {
		if comm[x] != comm["e"] {
			t.Fatalf("right clique split: %s in %d, e in %d", x, comm[x], comm["e"])
		}
	}
	if comm["a"] == comm["e"] {
		t.Fatalf("cliques not separated: both in community %d", comm["a"])
	}
	if k != 2 {
		t.Fatalf("K = %d, want 2", k)
	}
	if comm["a"] != 0 {
		t.Fatalf("smallest-id node 'a' must be community 0, got %d", comm["a"])
	}
}

func TestCluster_ContiguousLabels(t *testing.T) {
	doc := twoCliques()
	k := Cluster(doc)
	seen := map[int]bool{}
	for _, n := range doc.Nodes {
		seen[n.Community] = true
	}
	for i := 0; i < k; i++ {
		if !seen[i] {
			t.Fatalf("community %d missing — labels not contiguous 0..%d", i, k-1)
		}
	}
	if len(seen) != k {
		t.Fatalf("distinct communities = %d, want K = %d", len(seen), k)
	}
}

func TestCluster_ModularityBeatsSingleton(t *testing.T) {
	doc := twoCliques()
	for i := range doc.Nodes {
		doc.Nodes[i].Community = i
	}
	qSingleton := modularity(doc)
	for i := range doc.Nodes {
		doc.Nodes[i].Community = graph.UnclusteredCommunity
	}
	Cluster(doc)
	qClustered := modularity(doc)
	if qClustered < qSingleton {
		t.Fatalf("clustered modularity %.4f < singleton %.4f", qClustered, qSingleton)
	}
}

func TestCluster_Deterministic(t *testing.T) {
	snap := func() string {
		doc := twoCliques()
		Cluster(doc)
		out := ""
		for _, n := range doc.Nodes {
			out += fmt.Sprintf("%s=%d;", n.ID, n.Community)
		}
		return out
	}
	if a, b := snap(), snap(); a != b {
		t.Fatalf("non-deterministic:\nA=%s\nB=%s", a, b)
	}
}

func TestCluster_EmptyAndIsolated(t *testing.T) {
	if k := Cluster(graph.NewDocument("repo")); k != 0 {
		t.Fatalf("empty K = %d, want 0", k)
	}
	iso := graph.NewDocument("repo")
	for _, id := range []string{"x", "y", "z"} {
		iso.Nodes = append(iso.Nodes, graph.Node{ID: id, Label: id, Kind: graph.KindFunction, File: id + ".go", Line: 1, Community: graph.UnclusteredCommunity})
	}
	k := Cluster(iso)
	if k != 3 {
		t.Fatalf("isolated K = %d, want 3 (each its own community)", k)
	}
	for _, n := range iso.Nodes {
		if n.Community < 0 {
			t.Fatalf("isolated node %s community %d < 0", n.ID, n.Community)
		}
	}
}

// modularity scores the current Node.Community assignment over the undirected,
// edge-count-weighted projection. Test-only oracle (independent of Cluster's
// internals) so it genuinely cross-checks the algorithm.
func modularity(doc *graph.Document) float64 {
	idToComm := map[string]int{}
	for _, n := range doc.Nodes {
		idToComm[n.ID] = n.Community
	}
	pw := map[[2]string]float64{}
	deg := map[string]float64{}
	var m2 float64
	for _, e := range doc.Edges {
		if e.From == e.To {
			continue
		}
		a, b := e.From, e.To
		if a > b {
			a, b = b, a
		}
		pw[[2]string{a, b}]++
		deg[e.From]++
		deg[e.To]++
		m2 += 2
	}
	if m2 == 0 {
		return 0
	}
	ids := make([]string, 0, len(deg))
	for id := range deg {
		ids = append(ids, id)
	}
	var q float64
	for _, i := range ids {
		for _, j := range ids {
			if idToComm[i] != idToComm[j] {
				continue
			}
			a, b := i, j
			if a > b {
				a, b = b, a
			}
			var aij float64
			if i != j {
				aij = pw[[2]string{a, b}]
			}
			q += aij - deg[i]*deg[j]/m2
		}
	}
	return q / m2
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/cluster/ -run TestCluster -v`
Expected: FAIL — compilation error `undefined: Cluster`.

- [ ] **Step 3: Write the Louvain implementation**

Create `/Users/mylive/project/graffiti/graffiti/internal/cluster/cluster.go`:
```go
// Package cluster assigns each node a community via a deterministic, single-level
// Louvain over an undirected, edge-count-weighted projection of the directed graph
// (spec §5 cluster stage), and names the resulting communities (spec §8.3).
//
// Determinism (spec §14): nodes are processed in ascending node-id order;
// modularity-gain ties break by smallest community id (then implicitly smallest
// member id, since communities are seeded as singletons in id order); no
// math/rand; communities are renumbered by smallest member id so labels are a
// contiguous 0..K-1. No Go-map iteration feeds any decision or any emitted order.
package cluster

import (
	"sort"

	"github.com/amazopic/graffiti/internal/graph"
)

// Cluster assigns doc.Nodes[i].Community in place to a contiguous 0..K-1 label and
// returns K. Every node ends with Community >= 0 (edge-less nodes become their own
// singleton community). doc is otherwise unmodified.
func Cluster(doc *graph.Document) int {
	n := len(doc.Nodes)
	if n == 0 {
		return 0
	}

	// Rank nodes by ascending id; ranks are the canonical processing order.
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(a, b int) bool {
		return doc.Nodes[order[a]].ID < doc.Nodes[order[b]].ID
	})
	idToRank := make(map[string]int, n)
	for rank, ni := range order {
		idToRank[doc.Nodes[ni].ID] = rank
	}

	// Undirected, edge-count-weighted projection. A directed edge u->v (any
	// relation) adds weight 1 to the unordered pair {u,v}; parallel edges
	// accumulate; self-loops are ignored.
	type pair struct{ a, b int } // ranks, a < b
	pw := map[pair]float64{}
	for _, e := range doc.Edges {
		ui, ok1 := idToRank[e.From]
		vi, ok2 := idToRank[e.To]
		if !ok1 || !ok2 || ui == vi {
			continue
		}
		if ui > vi {
			ui, vi = vi, ui
		}
		pw[pair{ui, vi}]++
	}

	type wedge struct {
		to int
		w  float64
	}
	adj := make([][]wedge, n)
	deg := make([]float64, n)
	var m2 float64

	pairs := make([]pair, 0, len(pw))
	for p := range pw {
		pairs = append(pairs, p)
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].a != pairs[j].a {
			return pairs[i].a < pairs[j].a
		}
		return pairs[i].b < pairs[j].b
	})
	for _, p := range pairs {
		w := pw[p]
		adj[p.a] = append(adj[p.a], wedge{p.b, w})
		adj[p.b] = append(adj[p.b], wedge{p.a, w})
		deg[p.a] += w
		deg[p.b] += w
		m2 += 2 * w
	}

	// Seed singletons (community == rank).
	comm := make([]int, n)
	commTot := make([]float64, n) // total degree of community c
	for i := 0; i < n; i++ {
		comm[i] = i
		commTot[i] = deg[i]
	}

	if m2 == 0 {
		return finalize(doc, order, comm)
	}

	// Local-moving pass, repeated until no node changes community.
	improved := true
	for improved {
		improved = false
		for i := 0; i < n; i++ { // ascending rank == ascending id
			wTo := map[int]float64{}
			for _, e := range adj[i] {
				wTo[comm[e.to]] += e.w
			}

			cur := comm[i]
			commTot[cur] -= deg[i] // pull i out before scoring

			cands := []int{cur}
			seen := map[int]bool{cur: true}
			for c := range wTo {
				if !seen[c] {
					cands = append(cands, c)
					seen[c] = true
				}
			}
			sort.Ints(cands) // ascending: first maximizer wins ties (smallest comm id)

			bestC := cur
			bestGain := wTo[cur] - deg[i]*commTot[cur]/m2
			for _, c := range cands {
				if c == cur {
					continue
				}
				if gain := wTo[c] - deg[i]*commTot[c]/m2; gain > bestGain {
					bestGain, bestC = gain, c
				}
			}

			commTot[bestC] += deg[i]
			if bestC != cur {
				comm[i] = bestC
				improved = true
			}
		}
	}

	return finalize(doc, order, comm)
}

// finalize renumbers raw communities by smallest member rank (== smallest id) so
// labels are contiguous 0..K-1, writes Node.Community, and returns K.
func finalize(doc *graph.Document, order, comm []int) int {
	minRank := map[int]int{}
	for rank, c := range comm {
		if r, ok := minRank[c]; !ok || rank < r {
			minRank[c] = rank
		}
	}
	type cr struct{ raw, rank int }
	crs := make([]cr, 0, len(minRank))
	for raw, r := range minRank {
		crs = append(crs, cr{raw, r})
	}
	sort.Slice(crs, func(i, j int) bool { return crs[i].rank < crs[j].rank })
	remap := make(map[int]int, len(crs))
	for newID, c := range crs {
		remap[c.raw] = newID
	}
	for rank, c := range comm {
		doc.Nodes[order[rank]].Community = remap[c]
	}
	return len(crs)
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/cluster/ -run TestCluster -v`
Expected: PASS — `TestCluster_TwoCliquesSeparate`, `TestCluster_ContiguousLabels`, `TestCluster_ModularityBeatsSingleton`, `TestCluster_Deterministic`, `TestCluster_EmptyAndIsolated`.

- [ ] **Step 5: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git checkout -b plan2-clustering-analysis
git add internal/cluster/cluster.go internal/cluster/cluster_test.go
git commit -m "feat: deterministic single-level Louvain clustering

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: Community naming + populate `Document.Communities` (§8.3 hybrid)

**Files:**
- Create: `internal/cluster/naming.go`
- Create: `internal/cluster/naming_test.go`

- [ ] **Step 1: Write the failing naming test**

Create `/Users/mylive/project/graffiti/graffiti/internal/cluster/naming_test.go`:
```go
package cluster

import (
	"testing"

	"github.com/amazopic/graffiti/internal/graph"
)

func TestNameCommunities_DominantDirectory(t *testing.T) {
	doc := graph.NewDocument("repo")
	// community 0: three members all under internal/auth -> "Auth"
	doc.Nodes = []graph.Node{
		{ID: "a", Label: "Login", Kind: graph.KindFunction, File: "internal/auth/login.go", Line: 1, Community: 0},
		{ID: "b", Label: "Logout", Kind: graph.KindFunction, File: "internal/auth/logout.go", Line: 1, Community: 0},
		{ID: "c", Label: "Session", Kind: graph.KindClass, File: "internal/auth/session.go", Line: 1, Community: 0},
	}
	deg := map[string]int{"a": 5, "b": 1, "c": 2}
	comms := NameCommunities(doc, deg)
	if len(comms) != 1 {
		t.Fatalf("communities = %d, want 1", len(comms))
	}
	if comms[0].Label != "Auth" {
		t.Fatalf("label = %q, want %q", comms[0].Label, "Auth")
	}
	wantMembers := []string{"a", "b", "c"}
	if len(comms[0].Members) != 3 || comms[0].Members[0] != wantMembers[0] {
		t.Fatalf("members = %v, want sorted %v", comms[0].Members, wantMembers)
	}
}

func TestNameCommunities_FallbackMostCentral(t *testing.T) {
	doc := graph.NewDocument("repo")
	// No directory holds a strict majority (all distinct, root-level files), so
	// fall back to the most-central (highest-degree) member's label.
	doc.Nodes = []graph.Node{
		{ID: "a", Label: "Alpha", Kind: graph.KindFunction, File: "alpha.go", Line: 1, Community: 0},
		{ID: "b", Label: "Beta", Kind: graph.KindFunction, File: "beta.go", Line: 1, Community: 0},
		{ID: "c", Label: "Gamma", Kind: graph.KindFunction, File: "gamma.go", Line: 1, Community: 0},
	}
	deg := map[string]int{"a": 2, "b": 9, "c": 3} // Beta most central
	comms := NameCommunities(doc, deg)
	if comms[0].Label != "Beta" {
		t.Fatalf("fallback label = %q, want %q (most-central member)", comms[0].Label, "Beta")
	}
}

func TestNameCommunities_TitleCasesBaseSegment(t *testing.T) {
	doc := graph.NewDocument("repo")
	doc.Nodes = []graph.Node{
		{ID: "a", Label: "X", File: "internal/http_router/a.go", Community: 0},
		{ID: "b", Label: "Y", File: "internal/http_router/b.go", Community: 0},
	}
	comms := NameCommunities(doc, map[string]int{"a": 1, "b": 1})
	if comms[0].Label != "Http Router" {
		t.Fatalf("label = %q, want %q", comms[0].Label, "Http Router")
	}
}

func TestNameCommunities_SortedByID(t *testing.T) {
	doc := graph.NewDocument("repo")
	doc.Nodes = []graph.Node{
		{ID: "z", Label: "Z", File: "b/z.go", Community: 1},
		{ID: "a", Label: "A", File: "a/a.go", Community: 0},
	}
	comms := NameCommunities(doc, map[string]int{"a": 1, "z": 1})
	if len(comms) != 2 || comms[0].ID != 0 || comms[1].ID != 1 {
		t.Fatalf("communities must be sorted by id ascending, got %+v", comms)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/cluster/ -run TestNameCommunities -v`
Expected: FAIL — `undefined: NameCommunities`.

- [ ] **Step 3: Write the naming implementation**

Create `/Users/mylive/project/graffiti/graffiti/internal/cluster/naming.go`:
```go
package cluster

import (
	"path"
	"sort"
	"strings"

	"github.com/amazopic/graffiti/internal/graph"
)

// NameCommunities builds the doc.Communities slice from the clustered nodes
// (spec §8.3): each community's label is the human form of its members' dominant
// source directory (strict majority), falling back to the most-central member's
// label when no directory dominates or the dominant directory is generic/root.
// Members are sorted; communities are returned sorted by id ascending.
//
// deg maps node id -> total (in+out) degree (see analyze.Degrees); it is used
// only for the most-central tie-break, so callers may pass an empty map if they
// do not need the fallback to be centrality-aware (it then falls back to the
// smallest-id member, which is still deterministic).
func NameCommunities(doc *graph.Document, deg map[string]int) []graph.Community {
	members := map[int][]string{}
	maxComm := -1
	for _, nd := range doc.Nodes {
		if nd.Community < 0 {
			continue
		}
		members[nd.Community] = append(members[nd.Community], nd.ID)
		if nd.Community > maxComm {
			maxComm = nd.Community
		}
	}

	byID := make(map[string]graph.Node, len(doc.Nodes))
	for _, nd := range doc.Nodes {
		byID[nd.ID] = nd
	}

	out := make([]graph.Community, 0, maxComm+1)
	for c := 0; c <= maxComm; c++ {
		mem := members[c]
		if len(mem) == 0 {
			continue
		}
		sort.Strings(mem)
		out = append(out, graph.Community{
			ID:      c,
			Label:   labelFor(mem, byID, deg),
			Members: mem,
		})
	}
	return out
}

// labelFor implements the §8.3 hybrid heuristic. members must be pre-sorted.
func labelFor(members []string, byID map[string]graph.Node, deg map[string]int) string {
	// Count members per (non-generic) directory.
	dirCount := map[string]int{}
	for _, id := range members {
		if d := dirOf(byID[id].File); d != "" {
			dirCount[d]++
		}
	}
	// Pick the dominant directory deterministically: highest count, ties by the
	// lexicographically smallest directory.
	dirs := make([]string, 0, len(dirCount))
	for d := range dirCount {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	bestDir, bestN := "", 0
	for _, d := range dirs {
		if dirCount[d] > bestN {
			bestN, bestDir = dirCount[d], d
		}
	}
	// "Dominant" == strict majority of members share it.
	if bestDir != "" && bestN*2 > len(members) {
		return dirLabel(bestDir)
	}

	// Fallback: most-central member (highest degree; ties by smallest id, which
	// is members[0] since members is sorted ascending).
	best, bestDeg := members[0], deg[members[0]]
	for _, id := range members[1:] {
		if deg[id] > bestDeg {
			bestDeg, best = deg[id], id
		}
	}
	return byID[best].Label
}

// dirOf returns the directory of a repo-relative file path ("" for root files).
func dirOf(file string) string {
	d := path.Dir(file)
	if d == "." || d == "/" || d == "" {
		return ""
	}
	return d
}

// dirLabel turns a directory like "internal/auth" into a human label "Auth",
// title-casing the base segment (split on '-'/'_').
func dirLabel(dir string) string {
	base := path.Base(dir)
	if base == "" || base == "." || base == "/" {
		return dir
	}
	parts := strings.FieldsFunc(base, func(r rune) bool { return r == '-' || r == '_' })
	for i, p := range parts {
		if p != "" {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	if len(parts) == 0 {
		return base
	}
	return strings.Join(parts, " ")
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/cluster/ -run TestNameCommunities -v`
Expected: PASS (4 tests).

- [ ] **Step 5: Run the whole cluster package**

Run: `go test ./internal/cluster/ -v`
Expected: PASS (all Task 1 + Task 2 tests).

- [ ] **Step 6: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/cluster/naming.go internal/cluster/naming_test.go
git commit -m "feat: deterministic community naming (dominant dir, central-member fallback)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: `internal/analyze` — god nodes, surprising connections, import cycles, 3 questions

**Files:**
- Create: `internal/analyze/analyze.go`
- Create: `internal/analyze/analyze_test.go`

- [ ] **Step 1: Write the failing analyze test**

Create `/Users/mylive/project/graffiti/graffiti/internal/analyze/analyze_test.go`:
```go
package analyze

import (
	"strings"
	"testing"

	"github.com/amazopic/graffiti/internal/graph"
)

// hub: one central node touched by many, plus a 2nd community with a cross edge.
func hubDoc() *graph.Document {
	doc := graph.NewDocument("repo")
	doc.Nodes = []graph.Node{
		{ID: "hub", Label: "Hub", Kind: graph.KindFunction, File: "core/hub.go", Line: 1, Community: 0},
		{ID: "a", Label: "A", Kind: graph.KindFunction, File: "core/a.go", Line: 1, Community: 0},
		{ID: "b", Label: "B", Kind: graph.KindFunction, File: "core/b.go", Line: 1, Community: 0},
		{ID: "c", Label: "C", Kind: graph.KindFunction, File: "core/c.go", Line: 1, Community: 0},
		{ID: "far", Label: "Far", Kind: graph.KindFunction, File: "ext/far.go", Line: 1, Community: 1},
	}
	mk := func(f, t string) graph.Edge {
		return graph.Edge{From: f, To: t, Relation: graph.RelCalls, Confidence: graph.ConfInferred}
	}
	doc.Edges = []graph.Edge{
		mk("a", "hub"), mk("b", "hub"), mk("c", "hub"),
		mk("hub", "far"), // cross-community (surprising)
	}
	return doc
}

func TestAnalyze_GodNodes(t *testing.T) {
	doc := hubDoc()
	an := Analyze(doc, Degrees(doc))
	if len(an.GodNodes) == 0 {
		t.Fatalf("expected at least one god node")
	}
	if an.GodNodes[0].ID != "hub" {
		t.Fatalf("top god node = %q, want hub", an.GodNodes[0].ID)
	}
	if an.GodNodes[0].Degree != 4 {
		t.Fatalf("hub degree = %d, want 4", an.GodNodes[0].Degree)
	}
}

func TestAnalyze_GodNodeCap(t *testing.T) {
	doc := graph.NewDocument("repo")
	// 10 nodes, each with degree >= 2, all distinct communities irrelevant here.
	for i := 0; i < 10; i++ {
		id := string(rune('a' + i))
		doc.Nodes = append(doc.Nodes, graph.Node{ID: id, Label: id, Kind: graph.KindFunction, File: id + ".go", Line: 1, Community: 0})
	}
	// chain a-b-c-...-j plus wrap, so every node has degree 2.
	ids := "abcdefghij"
	for i := 0; i < len(ids); i++ {
		from := string(ids[i])
		to := string(ids[(i+1)%len(ids)])
		doc.Edges = append(doc.Edges, graph.Edge{From: from, To: to, Relation: graph.RelCalls, Confidence: graph.ConfInferred})
	}
	an := Analyze(doc, Degrees(doc))
	if len(an.GodNodes) > 7 {
		t.Fatalf("god nodes = %d, want <= 7 (cap)", len(an.GodNodes))
	}
}

func TestAnalyze_SurprisingCrossCommunity(t *testing.T) {
	doc := hubDoc()
	an := Analyze(doc, Degrees(doc))
	if len(an.Surprising) != 1 {
		t.Fatalf("surprising = %d, want 1", len(an.Surprising))
	}
	s := an.Surprising[0]
	if s.From != "hub" || s.To != "far" {
		t.Fatalf("surprising edge = %s->%s, want hub->far", s.From, s.To)
	}
	if s.FromComm == s.ToComm {
		t.Fatalf("surprising edge must cross communities")
	}
}

func TestAnalyze_ImportCycle(t *testing.T) {
	doc := graph.NewDocument("repo")
	doc.Nodes = []graph.Node{
		{ID: "p", Label: "p", Kind: graph.KindModule, File: "p.go", Line: 1, Community: 0},
		{ID: "q", Label: "q", Kind: graph.KindModule, File: "q.go", Line: 1, Community: 0},
	}
	doc.Edges = []graph.Edge{
		{From: "p", To: "q", Relation: graph.RelImports, Confidence: graph.ConfExtracted},
		{From: "q", To: "p", Relation: graph.RelImports, Confidence: graph.ConfExtracted},
	}
	an := Analyze(doc, Degrees(doc))
	if len(an.Cycles) != 1 {
		t.Fatalf("cycles = %d, want 1", len(an.Cycles))
	}
	// canonical rotation starts at the smallest id "p".
	if an.Cycles[0][0] != "p" {
		t.Fatalf("cycle canonical start = %q, want p", an.Cycles[0][0])
	}
}

func TestAnalyze_ExactlyThreeQuestions(t *testing.T) {
	doc := hubDoc()
	doc.Communities = []graph.Community{{ID: 0, Label: "Core", Members: []string{"a", "b", "c", "hub"}}, {ID: 1, Label: "Ext", Members: []string{"far"}}}
	an := Analyze(doc, Degrees(doc))
	if len(an.Questions) != 3 {
		t.Fatalf("questions = %d, want exactly 3", len(an.Questions))
	}
	// the top god node's label must surface in a question.
	joined := strings.Join(an.Questions, " | ")
	if !strings.Contains(joined, "Hub") {
		t.Fatalf("expected the top god node 'Hub' in a question, got %q", joined)
	}
}

func TestAnalyze_Deterministic(t *testing.T) {
	gen := func() string {
		doc := hubDoc()
		doc.Communities = []graph.Community{{ID: 0, Label: "Core", Members: []string{"a", "b", "c", "hub"}}, {ID: 1, Label: "Ext", Members: []string{"far"}}}
		an := Analyze(doc, Degrees(doc))
		return strings.Join(an.Questions, "\n")
	}
	if a, b := gen(), gen(); a != b {
		t.Fatalf("questions non-deterministic:\nA=%s\nB=%s", a, b)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/analyze/ -run TestAnalyze -v`
Expected: FAIL — `undefined: Analyze`, `undefined: Degrees`.

- [ ] **Step 3: Write the analyze implementation**

Create `/Users/mylive/project/graffiti/graffiti/internal/analyze/analyze.go`:
```go
// Package analyze derives read-only insights from a clustered graph (spec §5
// analyze stage / §8.5): god nodes (high-degree hubs), surprising connections
// (cross-community edges), import cycles, and exactly 3 suggested questions.
// It mutates nothing and performs no I/O. All outputs are deterministic: sorted
// iteration with explicit total-order tie-breaks; no math/rand; no LLM.
package analyze

import (
	"fmt"
	"sort"

	"github.com/amazopic/graffiti/internal/graph"
)

// GodNodeCap is the documented limit on map-surfaced god nodes (spec §8.5: ~7).
const GodNodeCap = 7

// Analysis is the bundle of derived insights.
type Analysis struct {
	GodNodes   []GodNode
	Surprising []SurprisingEdge
	Cycles     [][]string // each cycle is a canonical rotation of node ids
	Questions  []string   // exactly 3
}

// GodNode is a high-degree hub.
type GodNode struct {
	ID     string
	Label  string
	Degree int
}

// SurprisingEdge is a directed edge whose endpoints are in different communities.
type SurprisingEdge struct {
	From, To         string
	FromComm, ToComm int
	Relation         graph.Relation
	Confidence       graph.Confidence
}

// Degrees returns total (in + out) degree per node id over all edges.
func Degrees(doc *graph.Document) map[string]int {
	deg := make(map[string]int, len(doc.Nodes))
	for _, e := range doc.Edges {
		deg[e.From]++
		deg[e.To]++
	}
	return deg
}

// Analyze computes the full Analysis for a clustered document. deg should come
// from Degrees(doc).
func Analyze(doc *graph.Document, deg map[string]int) Analysis {
	byID := make(map[string]graph.Node, len(doc.Nodes))
	for _, n := range doc.Nodes {
		byID[n.ID] = n
	}
	gods := godNodes(doc, byID, deg)
	surp := surprising(doc, byID)
	cycles := importCycles(doc)
	qs := questions(doc, gods, surp, byID)
	return Analysis{GodNodes: gods, Surprising: surp, Cycles: cycles, Questions: qs}
}

// godNodes returns up to GodNodeCap highest-degree nodes (degree >= 2), ordered
// by degree desc then id asc.
func godNodes(doc *graph.Document, byID map[string]graph.Node, deg map[string]int) []GodNode {
	type dn struct {
		id  string
		deg int
	}
	cand := make([]dn, 0, len(doc.Nodes))
	for _, n := range doc.Nodes {
		if d := deg[n.ID]; d >= 2 {
			cand = append(cand, dn{n.ID, d})
		}
	}
	sort.Slice(cand, func(i, j int) bool {
		if cand[i].deg != cand[j].deg {
			return cand[i].deg > cand[j].deg
		}
		return cand[i].id < cand[j].id
	})
	out := make([]GodNode, 0, GodNodeCap)
	for _, x := range cand {
		if len(out) >= GodNodeCap {
			break
		}
		out = append(out, GodNode{ID: x.id, Label: byID[x.id].Label, Degree: x.deg})
	}
	return out
}

// surprising returns cross-community edges in (from,to,relation,confidence) order.
func surprising(doc *graph.Document, byID map[string]graph.Node) []SurprisingEdge {
	var out []SurprisingEdge
	for _, e := range doc.Edges {
		fn, ok1 := byID[e.From]
		tn, ok2 := byID[e.To]
		if !ok1 || !ok2 {
			continue
		}
		if fn.Community >= 0 && tn.Community >= 0 && fn.Community != tn.Community {
			out = append(out, SurprisingEdge{
				From: e.From, To: e.To,
				FromComm: fn.Community, ToComm: tn.Community,
				Relation: e.Relation, Confidence: e.Confidence,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
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
	return out
}

// importCycles finds cycles among `imports` edges via a deterministic DFS.
// Each cycle is returned as a canonical rotation (rotated to start at its
// smallest id); the set is de-duplicated and sorted.
func importCycles(doc *graph.Document) [][]string {
	adj := map[string][]string{}
	nodeset := map[string]bool{}
	for _, e := range doc.Edges {
		if e.Relation != graph.RelImports {
			continue
		}
		adj[e.From] = append(adj[e.From], e.To)
		nodeset[e.From] = true
		nodeset[e.To] = true
	}
	for k := range adj {
		sort.Strings(adj[k])
	}
	nodes := make([]string, 0, len(nodeset))
	for n := range nodeset {
		nodes = append(nodes, n)
	}
	sort.Strings(nodes)

	var found [][]string
	seen := map[string]bool{}
	color := map[string]int{} // 0 white, 1 gray, 2 black
	var stack []string

	var dfs func(u string)
	dfs = func(u string) {
		color[u] = 1
		stack = append(stack, u)
		for _, v := range adj[u] {
			switch color[v] {
			case 0:
				dfs(v)
			case 1:
				idx := 0
				for i, s := range stack {
					if s == v {
						idx = i
						break
					}
				}
				canon := canonicalCycle(append([]string(nil), stack[idx:]...))
				key := fmt.Sprint(canon)
				if !seen[key] {
					seen[key] = true
					found = append(found, canon)
				}
			}
		}
		stack = stack[:len(stack)-1]
		color[u] = 2
	}
	for _, n := range nodes {
		if color[n] == 0 {
			dfs(n)
		}
	}
	sort.Slice(found, func(i, j int) bool { return fmt.Sprint(found[i]) < fmt.Sprint(found[j]) })
	return found
}

func canonicalCycle(c []string) []string {
	if len(c) == 0 {
		return c
	}
	min, at := c[0], 0
	for i, s := range c {
		if s < min {
			min, at = s, i
		}
	}
	out := make([]string, 0, len(c))
	for i := 0; i < len(c); i++ {
		out = append(out, c[(at+i)%len(c)])
	}
	return out
}

// questions builds exactly 3 deterministic template questions (spec §11), padding
// from a fixed fallback list when the graph is too small to fill all three.
func questions(doc *graph.Document, gods []GodNode, surp []SurprisingEdge, byID map[string]graph.Node) []string {
	var qs []string
	if len(gods) > 0 {
		qs = append(qs, fmt.Sprintf("What does %q do, and why is it touched by %d things?", gods[0].Label, gods[0].Degree))
	}
	if len(surp) > 0 {
		qs = append(qs, fmt.Sprintf("Why does %q connect to %q across subsystems?",
			byID[surp[0].From].Label, byID[surp[0].To].Label))
	}
	if len(doc.Communities) > 0 {
		biggest := doc.Communities[0]
		for _, c := range doc.Communities {
			if len(c.Members) > len(biggest.Members) {
				biggest = c
			}
		}
		qs = append(qs, fmt.Sprintf("How does the %q subsystem (%d things) fit together?", biggest.Label, len(biggest.Members)))
	}
	fallback := []string{
		"What are the most central pieces of this codebase?",
		"Which parts of the code are most interconnected?",
		"Where would a change ripple the furthest?",
	}
	for fi := 0; len(qs) < 3; fi++ {
		qs = append(qs, fallback[fi%len(fallback)])
	}
	return qs[:3]
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/analyze/ -v`
Expected: PASS (7 tests).

- [ ] **Step 5: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/analyze/analyze.go internal/analyze/analyze_test.go
git commit -m "feat: analyze stage (god nodes, surprising links, cycles, 3 questions)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: `internal/render` — deterministic `MAP.md` emitter

**Files:**
- Create: `internal/render/mapmd.go`
- Create: `internal/render/mapmd_test.go`

- [ ] **Step 1: Write the failing MAP.md test**

Create `/Users/mylive/project/graffiti/graffiti/internal/render/mapmd_test.go`:
```go
package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amazopic/graffiti/internal/analyze"
	"github.com/amazopic/graffiti/internal/graph"
)

func sampleClustered() (*graph.Document, analyze.Analysis) {
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
	return doc, an
}

func TestRenderMapMD_SectionsPresentAndOrdered(t *testing.T) {
	doc, an := sampleClustered()
	md := RenderMapMD(doc, an)

	wantInOrder := []string{
		"## Start here",
		"## Landmarks (god nodes)",
		"## Districts",
		"### Auth",
		"### Http",
		"## Surprising connections",
		"## Confidence legend",
	}
	last := -1
	for _, h := range wantInOrder {
		idx := strings.Index(md, h)
		if idx < 0 {
			t.Fatalf("MAP.md missing section %q\n---\n%s", h, md)
		}
		if idx < last {
			t.Fatalf("section %q out of order", h)
		}
		last = idx
	}
	// the cross-community http-route -> auth-login must be surfaced as surprising.
	if !strings.Contains(md, "Surprising connections") || !strings.Contains(md, "Route") {
		t.Fatalf("surprising connection not rendered:\n%s", md)
	}
	// confidence legend plain-English (spec §8.5).
	for _, term := range []string{"EXTRACTED", "INFERRED", "AMBIGUOUS", "verify"} {
		if !strings.Contains(md, term) {
			t.Fatalf("confidence legend missing %q", term)
		}
	}
}

func TestRenderMapMD_Deterministic(t *testing.T) {
	doc, an := sampleClustered()
	if RenderMapMD(doc, an) != RenderMapMD(doc, an) {
		t.Fatalf("MAP.md not deterministic")
	}
}

func TestWriteMapMD_WritesNextToJSON(t *testing.T) {
	doc, an := sampleClustered()
	dir := t.TempDir()
	if err := WriteMapMD(doc, an, dir); err != nil {
		t.Fatalf("WriteMapMD: %v", err)
	}
	p := filepath.Join(dir, ".graffiti", "MAP.md")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read MAP.md: %v", err)
	}
	if !strings.Contains(string(b), "## Districts") {
		t.Fatalf("written MAP.md missing Districts section")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/render/ -run 'TestRenderMapMD|TestWriteMapMD' -v`
Expected: FAIL — `undefined: RenderMapMD`, `undefined: WriteMapMD`.

- [ ] **Step 3: Write the MAP.md implementation**

Create `/Users/mylive/project/graffiti/graffiti/internal/render/mapmd.go`:
```go
package render

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/amazopic/graffiti/internal/analyze"
	"github.com/amazopic/graffiti/internal/graph"
)

// WriteMapMD renders MAP.md and writes it to <root>/.graffiti/MAP.md, next to
// map.json. Deterministic modulo nothing (MAP.md carries no timestamp).
func WriteMapMD(doc *graph.Document, an analyze.Analysis, root string) error {
	dir := filepath.Join(root, ".graffiti")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "MAP.md"), []byte(RenderMapMD(doc, an)), 0o644)
}

// RenderMapMD renders the deterministic MAP.md text (spec §8.3/§8.5). It is a
// pure function of the clustered Document + Analysis; no Go-map iteration feeds
// any emitted order. Sections, in fixed order: title, summary, Start here,
// Landmarks (god nodes), Districts (per community), Surprising connections,
// Confidence legend.
func RenderMapMD(doc *graph.Document, an analyze.Analysis) string {
	byID := make(map[string]graph.Node, len(doc.Nodes))
	for _, n := range doc.Nodes {
		byID[n.ID] = n
	}

	var b strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&b, format, args...) }

	w("# %s — Map\n\n", doc.Root)
	w("_%d nodes, %d edges, %d communities. 0 API calls, $0._\n\n",
		len(doc.Nodes), len(doc.Edges), len(doc.Communities))

	w("## Start here\n\n")
	for i, q := range an.Questions {
		w("%d. %s\n", i+1, q)
	}
	w("\n")

	w("## Landmarks (god nodes)\n\n")
	if len(an.GodNodes) == 0 {
		w("_None._\n\n")
	} else {
		for _, g := range an.GodNodes {
			w("- **%s** — touched by %d things — change carefully.\n", g.Label, g.Degree)
		}
		w("\n")
	}

	w("## Districts\n\n")
	comms := append([]graph.Community(nil), doc.Communities...)
	sort.Slice(comms, func(i, j int) bool { return comms[i].ID < comms[j].ID })
	for _, c := range comms {
		w("### %s (%d things)\n\n", c.Label, len(c.Members))
		for _, id := range c.Members { // Members are pre-sorted by cluster.NameCommunities
			n := byID[id]
			w("- `%s` (%s, %s:%d)\n", n.Label, n.Kind, n.File, n.Line)
		}
		w("\n")
	}

	w("## Surprising connections\n\n")
	if len(an.Surprising) == 0 {
		w("_None._\n\n")
	} else {
		for _, s := range an.Surprising {
			w("- `%s` → `%s` (%s, %s)\n", byID[s.From].Label, byID[s.To].Label, s.Relation, s.Confidence)
		}
		w("\n")
	}

	w("## Confidence legend\n\n")
	w("- **EXTRACTED** — definite (verified from imports/syntax).\n")
	w("- **INFERRED** — inferred (same-package name match).\n")
	w("- **AMBIGUOUS** — guessed — verify.\n")

	return b.String()
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/render/ -run 'TestRenderMapMD|TestWriteMapMD' -v`
Expected: PASS (3 tests).

- [ ] **Step 5: Run the whole render package (json + mapmd)**

Run: `go test ./internal/render/ -v`
Expected: PASS (existing Plan-1 json tests + new mapmd tests).

- [ ] **Step 6: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/render/mapmd.go internal/render/mapmd_test.go
git commit -m "feat: deterministic MAP.md renderer

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: Wire `cluster → analyze → render(MAP.md)` into `app.Build` + thread the 3 questions

**Files:**
- Modify: `internal/app/app.go`
- Modify: `cmd/graffiti/main.go`
- Modify: `cmd/graffiti/main_test.go`

- [ ] **Step 1: Update `Stats` and `Build` in `internal/app/app.go`**

In `/Users/mylive/project/graffiti/graffiti/internal/app/app.go`, add the two new imports to the existing import block:
```go
	"github.com/amazopic/graffiti/internal/analyze"
	"github.com/amazopic/graffiti/internal/cluster"
```
so the block reads:
```go
import (
	"os"
	"path/filepath"

	"github.com/amazopic/graffiti/internal/analyze"
	"github.com/amazopic/graffiti/internal/build"
	"github.com/amazopic/graffiti/internal/cache"
	"github.com/amazopic/graffiti/internal/cluster"
	"github.com/amazopic/graffiti/internal/graph"
	"github.com/amazopic/graffiti/internal/parse"
	"github.com/amazopic/graffiti/internal/render"
	"github.com/amazopic/graffiti/internal/scan"
)
```

Add a `Questions` field to `Stats`:
```go
// Stats summarizes a build for the CLI success line.
type Stats struct {
	Files       int
	Nodes       int
	Edges       int
	Communities int
	HasDocNode  bool     // whether a markdown doc node was emitted
	Questions   []string // the 3 suggested questions (spec §11), deterministic
}
```

Replace the post-`Assemble` tail of `Build` (the block from `doc, err := build.Assemble(...)` through `return stats, nil`) with the wired version:
```go
	doc, err := build.Assemble(docRoot, generatedAt, extractions)
	if err != nil {
		return stats, err
	}

	// Plan 2: cluster, name communities, analyze — all deterministic, no I/O above render.
	cluster.Cluster(doc)
	deg := analyze.Degrees(doc)
	doc.Communities = cluster.NameCommunities(doc, deg)
	an := analyze.Analyze(doc, deg)

	if err := render.WriteMapJSON(doc, absRoot); err != nil {
		return stats, err
	}
	if err := render.WriteMapMD(doc, an, absRoot); err != nil {
		return stats, err
	}
	if err := c.Flush(); err != nil {
		return stats, err
	}

	stats.Nodes = len(doc.Nodes)
	stats.Edges = len(doc.Edges)
	stats.Communities = len(doc.Communities)
	stats.Questions = an.Questions
	return stats, nil
```

> Note: `cluster.Cluster` and `cluster.NameCommunities` run *after* `build.Assemble` (which sorts nodes by id and edges deterministically) and *before* the writers, so the JSON node order is unchanged (still sorted by id) — only each node's `community` value and the `communities[]` array are newly populated. `generated_at` remains single-sourced (stamped in `Assemble`, read off `doc.GeneratedAt` by the JSON writer); `time.Now()` is still called exactly once, in `cmd/graffiti/main.go`.

- [ ] **Step 2: Run the app build test to verify it still compiles + passes**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/app/ -run TestBuild -v`
Expected: PASS (the Plan-1 `TestBuild*` tests; communities are now > 0 but those tests assert counts/shape, not exact community ids — if any Plan-1 app test hard-codes `Communities == 0`, update it to `>= 1` here).

- [ ] **Step 3: Update the CLI success line to print the 3 questions**

In `/Users/mylive/project/graffiti/graffiti/cmd/graffiti/main.go`, replace the body of `runBuild` (keep the signature and the `time.Now()` single source):
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
	fmt.Fprintln(stdout, "  The 3 most interesting questions your map can answer:")
	for i, q := range stats.Questions {
		fmt.Fprintf(stdout, "    %d) %s\n", i+1, q)
	}
	return 0
}
```

- [ ] **Step 4: Update the end-to-end CLI test to assert the questions print**

In `/Users/mylive/project/graffiti/graffiti/cmd/graffiti/main_test.go`, replace the body of `TestRun_BuildPrintsSuccessLine` (keep its name) with:
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
	s := out.String()
	if !strings.Contains(s, "Done. 0 API calls, $0.") {
		t.Fatalf("missing success line, got %q", s)
	}
	if !strings.Contains(s, "The 3 most interesting questions your map can answer:") {
		t.Fatalf("missing questions header, got %q", s)
	}
	if !strings.Contains(s, "1) ") || !strings.Contains(s, "2) ") || !strings.Contains(s, "3) ") {
		t.Fatalf("expected 3 numbered questions, got %q", s)
	}
}
```

- [ ] **Step 5: Run the CLI tests to verify they pass**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./cmd/graffiti/ -v`
Expected: PASS (`TestRun_NoArgs_PrintsUsage`, `TestRun_UnknownCommand_Errors`, `TestRun_BuildPrintsSuccessLine`).

- [ ] **Step 6: Commit**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/app/app.go cmd/graffiti/main.go cmd/graffiti/main_test.go
git commit -m "feat: wire cluster->analyze->MAP.md into pipeline; print 3 questions

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: Regenerate the `map.json` golden + new `MAP.md` golden + clustering structural assertions

**Files:**
- Modify: `internal/app/golden_test.go`
- Regenerate: `testdata/golden/gorepo.map.json`
- Create (via UPDATE_GOLDEN): `testdata/golden/gorepo.MAP.md`

> The Plan-1 golden `testdata/golden/gorepo.map.json` has `"community": -1` on every node and `"communities": []`. After Task 5 the build clusters, so this golden is now stale by design. It is regenerated from REAL output here (never hand-edited), and a new code-asserted structural test locks the *clustering invariants* (not blind community numbers) so a broken clustering can't be silently frozen.

- [ ] **Step 1: Add clustering + MAP.md golden tests to `internal/app/golden_test.go`**

In `/Users/mylive/project/graffiti/graffiti/internal/app/golden_test.go`, add a helper to read the produced MAP.md and three new tests. Append the following to the file (the existing imports already include `encoding/json`, `os`, `path/filepath`, `regexp`, `testing`, and `graph`; add `strings` to the import block):
```go
// buildFixtureMapMD builds the fixture into a temp dir and returns the produced
// MAP.md bytes (companion to buildFixtureIntoTemp, which returns map.json).
func buildFixtureMapMD(t *testing.T) []byte {
	t.Helper()
	src := filepath.Join("..", "..", "testdata", "fixtures", "gorepo")
	dst := t.TempDir()
	copyTree(t, src, dst)
	if _, err := Build(dst, fixtureGenAt); err != nil {
		t.Fatalf("Build fixture: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dst, ".graffiti", "MAP.md"))
	if err != nil {
		t.Fatalf("read produced MAP.md: %v", err)
	}
	return b
}

func mapMDGoldenPath() string {
	return filepath.Join("..", "..", "testdata", "golden", "gorepo.MAP.md")
}

func TestGolden_MapMD(t *testing.T) {
	got := buildFixtureMapMD(t)
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(mapMDGoldenPath(), got, 0o644); err != nil {
			t.Fatalf("write MAP.md golden: %v", err)
		}
		t.Log("MAP.md golden updated")
		return
	}
	want, err := os.ReadFile(mapMDGoldenPath())
	if err != nil {
		t.Fatalf("read MAP.md golden (run UPDATE_GOLDEN=1 to create): %v", err)
	}
	// MAP.md prints the root in its title; normalize it like map.json's root.
	norm := func(b []byte) string {
		s := string(b)
		if i := strings.IndexByte(s, '\n'); i >= 0 {
			s = "# X — Map" + s[i:] // replace the title line (root-dependent)
		}
		return s
	}
	if norm(got) != norm(want) {
		t.Fatalf("MAP.md differs from golden.\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// TestClustering_StructuralInvariants asserts the clustering CONTRACT on the real
// fixture (not blind community numbers): every node has a community >= 0; labels
// are contiguous 0..K-1; communities[] matches node assignments; members sorted;
// the tightly-coupled greet package symbols land together; and a known god node /
// suggested-question substring is present. A broken clustering fails here loudly.
func TestClustering_StructuralInvariants(t *testing.T) {
	var doc graph.Document
	if err := json.Unmarshal(buildFixtureIntoTemp(t), &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// 1. Every node clustered (>= 0) and labels contiguous 0..K-1.
	maxC := -1
	seen := map[int]bool{}
	commOf := map[string]int{}
	for _, n := range doc.Nodes {
		if n.Community < 0 {
			t.Fatalf("node %q still unclustered (community=%d)", n.ID, n.Community)
		}
		seen[n.Community] = true
		commOf[n.ID] = n.Community
		if n.Community > maxC {
			maxC = n.Community
		}
	}
	for i := 0; i <= maxC; i++ {
		if !seen[i] {
			t.Fatalf("community %d missing — labels not contiguous", i)
		}
	}

	// 2. communities[] is consistent with node assignments, ids contiguous,
	//    members sorted and complete.
	if len(doc.Communities) != maxC+1 {
		t.Fatalf("communities len = %d, want %d", len(doc.Communities), maxC+1)
	}
	for i, c := range doc.Communities {
		if c.ID != i {
			t.Fatalf("communities[%d].id = %d, want %d (contiguous, sorted)", i, c.ID, i)
		}
		if c.Label == "" {
			t.Fatalf("community %d has empty label", c.ID)
		}
		if !sortedStrings(c.Members) {
			t.Fatalf("community %d members not sorted: %v", c.ID, c.Members)
		}
		for _, m := range c.Members {
			if commOf[m] != c.ID {
				t.Fatalf("member %q listed in community %d but node says %d", m, c.ID, commOf[m])
			}
		}
	}

	// 3. The greet package's own symbols (Hello, upper, Formatter.Format) are
	//    tightly coupled (Hello->upper, Formatter.Format->Hello) so they must
	//    share a community.
	hello := graph.NodeID("greet/greet.go", "Hello")
	upper := graph.NodeID("greet/greet.go", "upper")
	format := graph.NodeID("greet/greet_helper.go", "Formatter.Format")
	if commOf[hello] != commOf[upper] {
		t.Fatalf("Hello (c=%d) and upper (c=%d) should share a community", commOf[hello], commOf[upper])
	}
	if commOf[hello] != commOf[format] {
		t.Fatalf("Hello (c=%d) and Formatter.Format (c=%d) should share a community", commOf[hello], commOf[format])
	}
}

func sortedStrings(s []string) bool {
	for i := 1; i < len(s); i++ {
		if s[i-1] > s[i] {
			return false
		}
	}
	return true
}

// TestSuggestedQuestions_Shape asserts exactly 3 deterministic questions, with the
// most-connected greet symbol (Hello, the fixture's hub) surfaced in question 1.
func TestSuggestedQuestions_Shape(t *testing.T) {
	src := filepath.Join("..", "..", "testdata", "fixtures", "gorepo")
	dst := t.TempDir()
	copyTree(t, src, dst)
	stats, err := Build(dst, fixtureGenAt)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(stats.Questions) != 3 {
		t.Fatalf("questions = %d, want exactly 3", len(stats.Questions))
	}
	if !strings.Contains(stats.Questions[0], "Hello") {
		t.Fatalf("question 1 should name the hub symbol Hello, got %q", stats.Questions[0])
	}
}
```

> The `Hello`-as-hub assertion is grounded in the fixture: `Hello` is called by both `upper` (via `Hello->upper`) and `Formatter.Format` (`Formatter.Format->Hello`) and itself calls `upper`, giving it the highest degree among real defs. If, when you run Step 3, the actual top god node differs, **read the produced MAP.md / `stats.Questions[0]` and update this substring to the real hub** — the golden is the byte-source of truth; this test pins the *shape*, so adjust it to the validated real output, never the reverse.

- [ ] **Step 2: Run the clustering structural test first (the real gate) — it must pass before regenerating any golden**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/app/ -run 'TestClustering_StructuralInvariants|TestSuggestedQuestions_Shape' -v
```
Expected: PASS. If `TestSuggestedQuestions_Shape` fails only on the `"Hello"` substring, inspect the real top question and correct the substring (per the note above), then re-run until both pass. Do **not** regenerate the golden until this is green — never freeze a wrong clustering.

- [ ] **Step 3: Regenerate both goldens from real output**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
UPDATE_GOLDEN=1 go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/app/ -run 'TestGolden_GoRepoMapJSON|TestGolden_MapMD' -v
```
Expected: PASS with log lines `golden updated` and `MAP.md golden updated`. Now `testdata/golden/gorepo.map.json` has populated `community` values + a `communities[]` array, and `testdata/golden/gorepo.MAP.md` exists.

Read both regenerated files once to sanity-check: every node `community >= 0`; `communities[]` has contiguous ids with sorted members and human labels; MAP.md has the 7 sections (`Start here`, `Landmarks`, `Districts`, per-community `###`, `Surprising connections`, `Confidence legend`) and 3 numbered questions. The structural tests are the correctness source of truth; the goldens lock byte-exactness.

- [ ] **Step 4: Run the full golden + determinism + structural suite**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./internal/app/ -run 'TestGolden|TestDeterminism|TestClustering|TestSuggestedQuestions' -v
```
Expected: PASS — `TestGolden_GoRepoMapJSON`, `TestGolden_MapMD`, `TestGolden_StructuralShape` (Plan-1, still valid: node/edge shape is unchanged by clustering), `TestDeterminism_TwoBuildsByteIdentical` (now also covers community values), `TestClustering_StructuralInvariants`, `TestSuggestedQuestions_Shape`.

> The Plan-1 `TestDeterminism_TwoBuildsByteIdentical` already strips `generated_at` + `root` and compares full bytes; since clustering is deterministic, two builds remain byte-identical including community values — no change needed to that test. The Plan-1 `TestGolden_StructuralShape` asserts node kinds and `calls` edges only (no community field), so it is unaffected.

- [ ] **Step 5: Run the entire suite with build tags**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod" ./...`
Expected: all packages PASS (`ok` for cmd/graffiti, internal/analyze, internal/app, internal/build, internal/cache, internal/cluster, internal/graph, internal/parse, internal/render, internal/scan, schema).

- [ ] **Step 6: Commit the regenerated goldens + tests**

```bash
cd /Users/mylive/project/graffiti/graffiti
git add internal/app/golden_test.go testdata/golden/gorepo.map.json testdata/golden/gorepo.MAP.md
git commit -m "test: regenerate clustered map.json golden + MAP.md golden + clustering invariants

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 7: Full-suite verification, vet, and manual smoke

**Files:** none (verification only)

- [ ] **Step 1: Run the full test suite via the Makefile**

Run: `make test`
Expected: all packages PASS (the Makefile already passes the grammar-subset tags).

- [ ] **Step 2: Run vet**

Run: `make vet`
Expected: no output (clean).

- [ ] **Step 3: Build the binary and smoke-test cluster/analyze/MAP.md end-to-end on the fixture**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
make build
rm -rf /tmp/graffiti-p2-smoke && cp -r testdata/fixtures/gorepo /tmp/graffiti-p2-smoke
./graffiti build /tmp/graffiti-p2-smoke
test -f /tmp/graffiti-p2-smoke/.graffiti/map.json && echo "map.json OK"
test -f /tmp/graffiti-p2-smoke/.graffiti/MAP.md && echo "MAP.md OK"
grep -q '## Districts' /tmp/graffiti-p2-smoke/.graffiti/MAP.md && echo "Districts OK"
```
Expected stdout: a success line `✓ Done. 0 API calls, $0.  N files → M nodes, K edges, C communities.` where `C >= 1`, followed by `The 3 most interesting questions your map can answer:` and three numbered lines `1) … 2) … 3) …`; then `map.json OK`, `MAP.md OK`, `Districts OK`. Exit code 0.

- [ ] **Step 4: Confirm determinism at the binary level (two builds, identical artifacts modulo generated_at)**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
rm -rf /tmp/p2a /tmp/p2b
cp -r testdata/fixtures/gorepo /tmp/p2a && cp -r testdata/fixtures/gorepo /tmp/p2b
./graffiti build /tmp/p2a >/dev/null && ./graffiti build /tmp/p2b >/dev/null
# MAP.md carries no timestamp and root is identical (same basename differs: p2a vs p2b)
# so compare MAP.md modulo the title line, and map.json modulo generated_at+root.
diff <(sed '1s/.*/TITLE/' /tmp/p2a/.graffiti/MAP.md) <(sed '1s/.*/TITLE/' /tmp/p2b/.graffiti/MAP.md) && echo "MAP.md deterministic"
diff \
  <(sed -E 's/"(generated_at|root)": *"[^"]*"/"\1":"X"/' /tmp/p2a/.graffiti/map.json) \
  <(sed -E 's/"(generated_at|root)": *"[^"]*"/"\1":"X"/' /tmp/p2b/.graffiti/map.json) \
  && echo "map.json deterministic"
```
Expected: `MAP.md deterministic` and `map.json deterministic` (both diffs empty).

- [ ] **Step 5: Final tidy + finish the branch**

Run:
```bash
cd /Users/mylive/project/graffiti/graffiti
go mod tidy   # should be a no-op: Plan 2 adds no dependencies
make test
git diff --quiet go.mod go.sum || (git add go.mod go.sum && git commit -m "build: go mod tidy (no-op for Plan 2)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>")
```
Expected: `make test` prints `ok` for every package; `go mod tidy` changes nothing (Plan 2 is stdlib-only).

Then use superpowers:finishing-a-development-branch to merge `plan2-clustering-analysis` into `main` (or open a PR), per the same flow Plan 1 used.

---

## Self-Review

**1. Spec coverage (Plan 2 scope only):**
- §5 cluster (Louvain, community per node) → Task 1. Deterministic single-level Louvain over the undirected edge-count-weighted projection; `Node.Community` set to contiguous `0..K-1`. ✓
- §8.3 district naming (dominant source dir → most-central-member fallback, deterministic, covered by determinism test) → Task 2. ✓
- §5/§8.5 analyze (god nodes capped ~7, surprising cross-community connections, import cycles) → Task 3. ✓
- §11 exactly 3 suggested questions, deterministic templates, no LLM, printed in the success line → Tasks 3 (generation), 5 (CLI). ✓
- §5/§8.5 MAP.md (god nodes, surprising connections, districts, questions, confidence legend) written next to map.json → Task 4. ✓
- Pipeline wiring `build → cluster → analyze → render(map.json + MAP.md)`; communities populate `Document.Communities`; `map.json` golden regenerated; success line extended; single `generated_at` / single `time.Now()` preserved → Task 5, 6. ✓
- §14 determinism (no map-iteration in emitted order; sorted + total-order tie-breaks; byte-identical) + golden tests for both `map.json` and `MAP.md` + concrete structural assertions → Tasks 1–6. ✓
- Explicitly deferred (later plans, correctly absent): Canvas `map.html`, query/mcp, init, more languages, workspace, distribution; recursive multi-level Louvain aggregation (Open Issue #2).

**2. Determinism guarantees (spec §14), enforced and tested:**
- Node processing in ascending-id order; gain ties → smallest community id; communities renumbered by smallest member id; no `math/rand`/time/map-iteration in any decision (Task 1). Verified in the prototype: 5× byte-identical.
- Every emitted list is sorted with an explicit total order: members (id), communities (id), god nodes (degree desc, id asc), surprising edges (from,to,relation,confidence), cycles (canonical rotation then lexical), questions (fixed template order + fixed fallback). MAP.md section order is hard-coded.
- `generated_at` stamped once in `build.Assemble`; `time.Now()` called once in `cmd/graffiti/main.go`; MAP.md carries no timestamp at all. Tested by `TestDeterminism_TwoBuildsByteIdentical` (map.json) and `TestRenderMapMD_Deterministic` + the binary-level diff in Task 7.

**3. Type consistency with the merged Plan-1 code (checked against the real sources):**
- `graph.Node{ID,Label,Kind,File,Line,Community}`, `graph.Edge{From,To,Relation,Confidence}`, `graph.Community{ID,Label,Members}`, `graph.Document{...Communities}`, `graph.UnclusteredCommunity` (-1), `graph.NodeID`, `graph.Kind*/Rel*/Conf*`, `graph.NewDocument` — used exactly as defined in `internal/graph/graph.go` / `id.go`. ✓
- `cluster.Cluster(doc *graph.Document) int` (mutates `Node.Community`), `cluster.NameCommunities(doc, deg) []graph.Community` — Tasks 1–2; called from `app.Build`. ✓
- `analyze.Degrees(doc) map[string]int`, `analyze.Analyze(doc, deg) Analysis`, `Analysis{GodNodes,Surprising,Cycles,Questions}`, `GodNode{ID,Label,Degree}`, `SurprisingEdge{...}` — Task 3; consumed by render + app. ✓
- `render.RenderMapMD(doc, an) string`, `render.WriteMapMD(doc, an, root) error` — Task 4; the existing `render.WriteMapJSON(doc, absRoot)` and `orderedDocument` (alphabetical keys) are untouched, so node/edge/community key ordering stays deterministic. ✓
- `app.Build(root, generatedAt string) (Stats, error)` unchanged signature; `Stats` gains `Questions []string`; pipeline order `Assemble → Cluster → NameCommunities → Analyze → WriteMapJSON → WriteMapMD → Flush`. ✓
- `cmd/graffiti/main.go` keeps the single `time.Now()` and the exact Plan-1 success-line prefix, appending the 3-questions block. ✓
- Module path `github.com/amazopic/graffiti` used in every import; build tags applied to every command touching `internal/parse` (app, cmd). ✓

**4. Prototype evidence (load-bearing):** the Louvain, naming, analyze, and MAP.md renderer were compiled and run on Go 1.26.2 against a copy of the real `graph` types; the two-cliques separation, contiguity, `Q_clustered (0.4231) >= Q_singleton (-0.1272)`, god-node cap (8→7), naming heuristic, empty/isolated handling, and 5× byte-identical determinism were all observed. The task code is that validated prototype with `graph.*`-qualified types. No load-bearing claim is unverified.

---

## Open Issues (require owner attention)

1. **Fixture hub identity in `TestSuggestedQuestions_Shape`.** The test asserts the top suggested question names `Hello` (the fixture's highest-degree real symbol). This is grounded in the fixture's edges, but the *exact* top god node depends on the final degree tally; Task 6 Step 2 instructs the implementer to read the real `stats.Questions[0]` and correct the substring if it differs, *then* freeze the golden. This is a deliberate "validate-then-pin" gate, not a guess — flagged so the owner knows the substring is allowed to be adjusted to real output during execution.

2. **Single-level vs. multi-level Louvain.** This plan ships single-level local-moving (sufficient and provably deterministic for v1). Recursive multi-level aggregation (graph contraction + repeat) yields coarser districts on very large repos and is a candidate refinement for a later plan; it can only *merge* communities further, so it will not violate any invariant this plan's tests assert. Owner to confirm single-level is acceptable for v1's target repos (the "8–40 districts" §8.3 goal is met for small/medium repos; revisit if a 10k-node repo produces too many tiny districts).

3. **Naming "strict majority" threshold.** §8.3 says "dominant source directory"; this plan operationalizes "dominant" as a strict majority (`count*2 > len(members)`) of members sharing a directory, else the most-central-member fallback. This is deterministic and matches the spec's intent, but the exact threshold is a judgment call — owner may prefer plurality (largest bucket even without majority). Changing it is a one-line edit in `cluster.labelFor` and only affects labels, never membership or determinism.

---

**Plan complete and save location:** this document lives at `/Users/mylive/project/graffiti/graffiti/docs/superpowers/plans/2026-06-15-graffiti-plan-2-clustering-analysis-mapmd.md`. Execution options: (1) Subagent-Driven (recommended) — superpowers:subagent-driven-development; (2) Inline — superpowers:executing-plans.
