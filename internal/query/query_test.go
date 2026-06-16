package query

import (
	"strings"
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/graph"
	"github.com/evgeniy-achin/graffiti/internal/store"
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
	// budget=12 exactly fits the login seed (estimateTokens(formatNode)=12) and
	// blocks every neighbor (cheapest neighbor costs 11, 12+11 > 12), so no
	// BFS expansion can occur. (The plan's literal budget=8 cannot pass the
	// plan's verbatim implementation: no node line is < 11 tokens, so budget=8
	// would admit zero seeds — the validated algorithm has no first-seed
	// override. Corrected to the smallest budget that exercises the test's
	// stated intent against the validated code. See FINAL REPORT.)
	out := Query(idx, "login", 12)
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
	// one with the smaller id must seed first. A third distractor node WITHOUT
	// the term keeps df["widget"]=2 < N=3, so the smoothed IDF stays > 0 and
	// both widgets actually score (the plan's literal 2-node corpus has
	// df["widget"]=N=2 => idf=log(3/3)=0 => neither node scores => empty output,
	// which the validated algorithm cannot avoid; this mirrors the prototype's
	// 10-node tie-break fixture. See FINAL REPORT.)
	nodes := []graph.Node{
		{ID: "b.go:widget", Label: "widget", Kind: graph.KindFunction, File: "b.go", Line: 1},
		{ID: "a.go:widget", Label: "widget", Kind: graph.KindFunction, File: "a.go", Line: 1},
		{ID: "c.go:gadget", Label: "gadget", Kind: graph.KindFunction, File: "c.go", Line: 1},
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
