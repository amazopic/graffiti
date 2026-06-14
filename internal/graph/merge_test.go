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
