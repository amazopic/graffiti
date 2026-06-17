package build

import (
	"testing"

	"github.com/amazopic/graffiti/internal/graph"
	"github.com/amazopic/graffiti/internal/parse"
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
