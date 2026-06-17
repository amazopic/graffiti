package schemaval

import (
	"strings"
	"testing"

	"github.com/amazopic/graffiti/internal/graph"
)

func validDoc() *graph.Document {
	d := graph.NewDocument("repo")
	d.GeneratedAt = "2026-06-14T00:00:00Z"
	d.Nodes = []graph.Node{
		{ID: "f", Label: "f.go", Kind: graph.KindFile, File: "f.go", Line: 1, Community: -1},
		{ID: "f:hello", Label: "Hello", Kind: graph.KindFunction, File: "f.go", Line: 2, Community: -1},
	}
	d.Edges = []graph.Edge{
		{From: "f", To: "f:hello", Relation: graph.RelContains, Confidence: graph.ConfExtracted},
	}
	return d
}

func TestValidateDocument_OK(t *testing.T) {
	if err := ValidateDocument(validDoc()); err != nil {
		t.Fatalf("valid doc rejected: %v", err)
	}
}

func TestValidateDocument_BadKind(t *testing.T) {
	d := validDoc()
	d.Nodes[1].Kind = "widget"
	err := ValidateDocument(d)
	if err == nil || !strings.Contains(err.Error(), "kind") {
		t.Fatalf("expected kind error, got %v", err)
	}
}

func TestValidateDocument_DanglingEdge(t *testing.T) {
	d := validDoc()
	d.Edges = append(d.Edges, graph.Edge{From: "f", To: "ghost", Relation: graph.RelCalls, Confidence: graph.ConfInferred})
	err := ValidateDocument(d)
	if err == nil || !strings.Contains(err.Error(), "ghost") {
		t.Fatalf("expected dangling-edge error mentioning ghost, got %v", err)
	}
}

func TestValidateDocument_DuplicateNodeID(t *testing.T) {
	d := validDoc()
	d.Nodes = append(d.Nodes, graph.Node{ID: "f", Label: "dup", Kind: graph.KindFile, File: "f.go", Line: 1, Community: -1})
	err := ValidateDocument(d)
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate node id error, got %v", err)
	}
}

func TestValidateDocument_BadConfidence(t *testing.T) {
	d := validDoc()
	d.Edges[0].Confidence = "MAYBE"
	err := ValidateDocument(d)
	if err == nil || !strings.Contains(err.Error(), "confidence") {
		t.Fatalf("expected confidence error, got %v", err)
	}
}

func TestValidateDocument_MissingVersion(t *testing.T) {
	d := validDoc()
	d.Version = ""
	err := ValidateDocument(d)
	if err == nil || !strings.Contains(err.Error(), "version") {
		t.Fatalf("expected version error, got %v", err)
	}
}

func TestValidateDocument_MissingGeneratedAt(t *testing.T) {
	d := validDoc()
	d.GeneratedAt = ""
	err := ValidateDocument(d)
	if err == nil || !strings.Contains(err.Error(), "generated_at") {
		t.Fatalf("expected generated_at error, got %v", err)
	}
}
