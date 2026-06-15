package cluster

import (
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

func TestNameCommunities_DominantDirectory(t *testing.T) {
	doc := graph.NewDocument("repo")
	// community 0: three members all under internal/auth -> "Auth"
	doc.Nodes = []graph.Node{
		{ID: "a", Label: "Login", Kind: graph.KindFunction, File: "internal/auth/login.go", Line: 1, Community: 0},
		{ID: "b", Label: "Logout", Kind: graph.KindFunction, File: "internal/auth/logout.go", Line: 1, Community: 0},
		{ID: "c", Label: "Session", Kind: graph.KindClass, File: "internal/auth/session.go", Line: 1, Community: 0},
	}
	deg := map[string]int{"a": 5, "b": 1, "c": 2}
	comms := NameCommunities(doc, deg)
	if len(comms) != 1 {
		t.Fatalf("communities = %d, want 1", len(comms))
	}
	if comms[0].Label != "Auth" {
		t.Fatalf("label = %q, want %q", comms[0].Label, "Auth")
	}
	wantMembers := []string{"a", "b", "c"}
	if len(comms[0].Members) != 3 || comms[0].Members[0] != wantMembers[0] {
		t.Fatalf("members = %v, want sorted %v", comms[0].Members, wantMembers)
	}
}

func TestNameCommunities_FallbackMostCentral(t *testing.T) {
	doc := graph.NewDocument("repo")
	// No directory holds a strict majority (all distinct, root-level files), so
	// fall back to the most-central (highest-degree) member's label.
	doc.Nodes = []graph.Node{
		{ID: "a", Label: "Alpha", Kind: graph.KindFunction, File: "alpha.go", Line: 1, Community: 0},
		{ID: "b", Label: "Beta", Kind: graph.KindFunction, File: "beta.go", Line: 1, Community: 0},
		{ID: "c", Label: "Gamma", Kind: graph.KindFunction, File: "gamma.go", Line: 1, Community: 0},
	}
	deg := map[string]int{"a": 2, "b": 9, "c": 3} // Beta most central
	comms := NameCommunities(doc, deg)
	if comms[0].Label != "Beta" {
		t.Fatalf("fallback label = %q, want %q (most-central member)", comms[0].Label, "Beta")
	}
}

func TestNameCommunities_TitleCasesBaseSegment(t *testing.T) {
	doc := graph.NewDocument("repo")
	doc.Nodes = []graph.Node{
		{ID: "a", Label: "X", File: "internal/http_router/a.go", Community: 0},
		{ID: "b", Label: "Y", File: "internal/http_router/b.go", Community: 0},
	}
	comms := NameCommunities(doc, map[string]int{"a": 1, "b": 1})
	if comms[0].Label != "Http Router" {
		t.Fatalf("label = %q, want %q", comms[0].Label, "Http Router")
	}
}

func TestNameCommunities_SortedByID(t *testing.T) {
	doc := graph.NewDocument("repo")
	doc.Nodes = []graph.Node{
		{ID: "z", Label: "Z", File: "b/z.go", Community: 1},
		{ID: "a", Label: "A", File: "a/a.go", Community: 0},
	}
	comms := NameCommunities(doc, map[string]int{"a": 1, "z": 1})
	if len(comms) != 2 || comms[0].ID != 0 || comms[1].ID != 1 {
		t.Fatalf("communities must be sorted by id ascending, got %+v", comms)
	}
}
