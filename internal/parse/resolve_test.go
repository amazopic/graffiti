package parse

import (
	"testing"

	"github.com/amazopic/graffiti/internal/graph"
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
