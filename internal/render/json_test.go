package render

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

func sampleDoc(at string) *graph.Document {
	d := graph.NewDocument("repo")
	d.GeneratedAt = at
	d.Nodes = []graph.Node{
		{ID: "a", Label: "A", Kind: graph.KindFile, File: "a.go", Line: 1, Community: -1},
		{ID: "a:hello", Label: "Hello", Kind: graph.KindFunction, File: "a.go", Line: 2, Community: -1},
	}
	d.Edges = []graph.Edge{
		{From: "a", To: "a:hello", Relation: graph.RelContains, Confidence: graph.ConfExtracted},
	}
	return d
}

func TestWriteMapJSON_KeepsGeneratedAtAndIsValidJSON(t *testing.T) {
	dir := t.TempDir()
	at := "2026-06-14T12:00:00Z"
	if err := WriteMapJSON(sampleDoc(at), dir); err != nil {
		t.Fatalf("write: %v", err)
	}
	p := filepath.Join(dir, ".graffiti", "map.json")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var back graph.Document
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if back.GeneratedAt != at {
		t.Fatalf("generated_at = %q, want %q", back.GeneratedAt, at)
	}
	if !strings.HasSuffix(string(b), "}\n") {
		t.Fatalf("output must end with newline")
	}
}

func TestWriteMapJSON_DeterministicModuloGeneratedAt(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	if err := WriteMapJSON(sampleDoc("2026-06-14T00:00:00Z"), dir1); err != nil {
		t.Fatal(err)
	}
	if err := WriteMapJSON(sampleDoc("2099-01-01T00:00:00Z"), dir2); err != nil {
		t.Fatal(err)
	}
	b1, _ := os.ReadFile(filepath.Join(dir1, ".graffiti", "map.json"))
	b2, _ := os.ReadFile(filepath.Join(dir2, ".graffiti", "map.json"))

	reAt := regexp.MustCompile(`"generated_at":\s*"[^"]*"`)
	n1 := reAt.ReplaceAll(b1, []byte(`"generated_at":"X"`))
	n2 := reAt.ReplaceAll(b2, []byte(`"generated_at":"X"`))
	if string(n1) != string(n2) {
		t.Fatalf("output not byte-identical modulo generated_at:\n%s\n---\n%s", n1, n2)
	}
}

func TestWriteMapJSON_SortedTopLevelKeys(t *testing.T) {
	dir := t.TempDir()
	if err := WriteMapJSON(sampleDoc("2026-06-14T00:00:00Z"), dir); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".graffiti", "map.json"))
	s := string(b)
	order := []string{`"communities"`, `"edges"`, `"generated_at"`, `"nodes"`, `"root"`, `"version"`}
	last := -1
	for _, k := range order {
		idx := strings.Index(s, k)
		if idx < 0 {
			t.Fatalf("missing key %s", k)
		}
		if idx < last {
			t.Fatalf("key %s out of sorted order", k)
		}
		last = idx
	}
}
