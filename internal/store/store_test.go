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
