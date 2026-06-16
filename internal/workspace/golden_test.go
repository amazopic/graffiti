package workspace

import (
	"encoding/json"
	"testing"
)

func normalizeOverlay(t *testing.T, ov *Overlay) string {
	t.Helper()
	ov.GeneratedAt = "FIXED"
	b, err := json.MarshalIndent(ov, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestOverlay_DeterministicAndGolden(t *testing.T) {
	root := t.TempDir()
	writeMember(t, root, "frontend", "cartclient.fetchcart", "ui.render")
	writeMember(t, root, "backend", "handlers.get_cart", "db.save")
	reg := newRegistry(t, root)
	links := []ParsedLink{
		{"web", "cartclient.fetchcart", "api", "handlers.get_cart", "calls"},
		{"api", "db.save", "web", "ui.render", "references"},
	}
	a, err := ComputeOverlay(root, reg, links)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ComputeOverlay(root, reg, links)
	if err != nil {
		t.Fatal(err)
	}
	if normalizeOverlay(t, a) != normalizeOverlay(t, b) {
		t.Fatal("overlay computation is non-deterministic")
	}
	if len(a.Links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(a.Links))
	}
	// links sorted by (from,to,relation): api::db.save before web::cartclient...
	if a.Links[0].From != "api::db.save" {
		t.Fatalf("links not sorted: %+v", a.Links)
	}
}
