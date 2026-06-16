package workspace

import (
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/query"
)

func TestCombinedIndex_PrefixesAndLinks(t *testing.T) {
	root := t.TempDir()
	writeMember(t, root, "frontend", "cartclient.fetchcart")
	writeMember(t, root, "backend", "handlers.get_cart")
	reg := newRegistry(t, root)
	ov := &Overlay{
		Version: SchemaVersion,
		Links: []Link{{
			From: "web::cartclient.fetchcart", To: "api::handlers.get_cart",
			Relation: "calls", Confidence: "EXTRACTED", Via: "explicit",
		}},
	}
	idx, err := CombinedIndex(root, reg, ov)
	if err != nil {
		t.Fatal(err)
	}
	// alias-prefixed node ids exist
	if _, ok := idx.Node("web::cartclient.fetchcart"); !ok {
		t.Fatal("missing alias-prefixed web node")
	}
	if _, ok := idx.Node("api::handlers.get_cart"); !ok {
		t.Fatal("missing alias-prefixed api node")
	}
	// the cross-edge is present (out of the web node)
	found := false
	for _, e := range idx.Out("web::cartclient.fetchcart") {
		if e.To == "api::handlers.get_cart" && string(e.Relation) == "calls" {
			found = true
		}
	}
	if !found {
		t.Fatal("cross-edge not present in combined index")
	}
	// a federated query over the combined index returns alias-prefixed text
	out := query.Query(idx, "fetchcart", query.DefaultTokenBudget)
	if !contains(out, "web::cartclient.fetchcart") {
		t.Fatalf("federated query output not alias-prefixed:\n%s", out)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
