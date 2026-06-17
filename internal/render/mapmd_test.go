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
