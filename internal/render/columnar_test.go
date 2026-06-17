package render

import (
	"encoding/json"
	"testing"

	"github.com/amazopic/graffiti/internal/graph"
)

func islandDoc() *graph.Document {
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
	is := graphIsland(islandDoc())
	if len(is.Label) != 4 || len(is.Kind) != 4 || len(is.File) != 4 || len(is.Deg) != 4 || len(is.Cat) != 4 {
		t.Fatalf("columnar arrays must all be len 4")
	}
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
	if is.Deg[idxOf["Foo"]] != 2 {
		t.Errorf("Foo degree = %d, want 2", is.Deg[idxOf["Foo"]])
	}
	if len(is.Edges) != 2 {
		t.Fatalf("want 2 edges, got %d", len(is.Edges))
	}
	for _, e := range is.Edges {
		if e[0] < 0 || e[0] >= 4 || e[1] < 0 || e[1] >= 4 {
			t.Fatalf("edge index out of range: %v", e)
		}
	}
}

func TestGraphIsland_Deterministic(t *testing.T) {
	a, _ := json.Marshal(graphIsland(islandDoc()))
	b, _ := json.Marshal(graphIsland(islandDoc()))
	if string(a) != string(b) {
		t.Fatal("island marshal not deterministic")
	}
}

func TestCategoryHeuristics(t *testing.T) {
	cases := []struct {
		kind        graph.Kind
		file, label string
		want        int
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
