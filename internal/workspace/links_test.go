package workspace

import "testing"

func TestParseLinks(t *testing.T) {
	in := `# workspace links
web::cartclient.fetchcart -> api::handlers.get_cart calls

api::db.save -> web::types.cart
  web::a.b -> api::c.d references  # trailing comment
`
	links, err := ParseLinks([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 3 {
		t.Fatalf("expected 3 links, got %d: %+v", len(links), links)
	}
	if links[0].FromAlias != "web" || links[0].FromID != "cartclient.fetchcart" {
		t.Fatalf("bad from: %+v", links[0])
	}
	if links[0].ToAlias != "api" || links[0].ToID != "handlers.get_cart" {
		t.Fatalf("bad to: %+v", links[0])
	}
	if links[0].Relation != "calls" {
		t.Fatalf("relation = %q, want calls", links[0].Relation)
	}
	if links[1].Relation != "references" { // default
		t.Fatalf("default relation = %q, want references", links[1].Relation)
	}
}

func TestParseLinks_Errors(t *testing.T) {
	for _, bad := range []string{
		"web::a -> noalias",       // RHS missing alias::
		"noarrow line",            // no ->
		"web::a -> api::b badrel", // unknown relation
		"::a -> api::b",           // empty alias
		"web:: -> api::b",         // empty id
	} {
		if _, err := ParseLinks([]byte(bad)); err == nil {
			t.Errorf("expected error for %q", bad)
		}
	}
}
