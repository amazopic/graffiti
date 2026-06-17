package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amazopic/graffiti/internal/graph"
	"github.com/amazopic/graffiti/internal/render"
)

// writeMember builds a member dir with a map.json containing the given node ids.
func writeMember(t *testing.T, root, rel string, ids ...string) {
	t.Helper()
	dir := filepath.Join(root, rel)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := graph.NewDocument(rel)
	doc.GeneratedAt = "2026-06-17T00:00:00Z"
	for _, id := range ids {
		doc.Nodes = append(doc.Nodes, graph.Node{ID: id, Label: id, Kind: graph.KindFunction, File: "f", Line: 1, Community: -1})
	}
	if err := render.WriteMapJSON(doc, dir); err != nil {
		t.Fatal(err)
	}
}

func newRegistry(t *testing.T, root string) *Registry {
	t.Helper()
	reg := &Registry{Version: SchemaVersion, Name: "ws", GeneratedAt: "2026-06-17T00:00:00Z"}
	AddMember(reg, Member{Alias: "web", Path: "frontend"})
	AddMember(reg, Member{Alias: "api", Path: "backend"})
	return reg
}

func TestComputeOverlay_ResolvesExplicitLinks(t *testing.T) {
	root := t.TempDir()
	writeMember(t, root, "frontend", "cartclient.fetchcart")
	writeMember(t, root, "backend", "handlers.get_cart")
	reg := newRegistry(t, root)
	links := []ParsedLink{{"web", "cartclient.fetchcart", "api", "handlers.get_cart", "calls"}}

	ov, err := ComputeOverlay(root, reg, links)
	if err != nil {
		t.Fatal(err)
	}
	if len(ov.Links) != 1 {
		t.Fatalf("expected 1 link, got %d (%+v)", len(ov.Links), ov)
	}
	l := ov.Links[0]
	if l.From != "web::cartclient.fetchcart" || l.To != "api::handlers.get_cart" {
		t.Fatalf("bad endpoints: %+v", l)
	}
	if l.Confidence != string(graph.ConfExtracted) || l.Via != "explicit" || l.Relation != "calls" {
		t.Fatalf("bad link metadata: %+v", l)
	}
	if ov.SourceHashes["web"] == "" || ov.SourceHashes["api"] == "" {
		t.Fatalf("missing source hashes: %+v", ov.SourceHashes)
	}
}

func TestComputeOverlay_DropsUnresolvable(t *testing.T) {
	root := t.TempDir()
	writeMember(t, root, "frontend", "real.node")
	writeMember(t, root, "backend", "other.node")
	reg := newRegistry(t, root)
	links := []ParsedLink{
		{"web", "real.node", "api", "ghost.node", "calls"}, // To ghost → unresolved
		{"web", "missing", "api", "other.node", "calls"},   // From ghost → unresolved
		{"nope", "x", "api", "other.node", "calls"},        // bad alias → unresolved
	}
	ov, err := ComputeOverlay(root, reg, links)
	if err != nil {
		t.Fatal(err)
	}
	if len(ov.Links) != 0 {
		t.Fatalf("expected 0 confident links, got %+v", ov.Links)
	}
	if len(ov.Unresolved) != 3 {
		t.Fatalf("expected 3 unresolved, got %d", len(ov.Unresolved))
	}
}

func TestOverlay_SaveLoad_DeterministicSort(t *testing.T) {
	root := t.TempDir()
	ov := &Overlay{
		Version: SchemaVersion, GeneratedAt: "T",
		SourceHashes: map[string]string{"api": "h"},
		Links: []Link{
			{From: "web::b", To: "api::z", Relation: "calls", Confidence: "EXTRACTED", Via: "explicit"},
			{From: "web::a", To: "api::y", Relation: "calls", Confidence: "EXTRACTED", Via: "explicit"},
		},
	}
	if err := SaveOverlay(root, ov); err != nil {
		t.Fatal(err)
	}
	got, err := LoadOverlay(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Links[0].From != "web::a" { // sorted by (from,to,relation)
		t.Fatalf("links not sorted: %+v", got.Links)
	}
}
