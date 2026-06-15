package analyze

import (
	"strings"
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/graph"
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
