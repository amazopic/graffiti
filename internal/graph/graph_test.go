package graph

import "testing"

func TestNewDocument_Defaults(t *testing.T) {
	d := NewDocument("myrepo")
	if d.Version != SchemaVersion {
		t.Fatalf("version = %q, want %q", d.Version, SchemaVersion)
	}
	if d.Root != "myrepo" {
		t.Fatalf("root = %q, want %q", d.Root, "myrepo")
	}
	if d.Nodes == nil || d.Edges == nil || d.Communities == nil {
		t.Fatalf("slices must be non-nil (got nodes=%v edges=%v comms=%v)", d.Nodes, d.Edges, d.Communities)
	}
	if len(d.Nodes) != 0 || len(d.Edges) != 0 {
		t.Fatalf("new document must start empty")
	}
}

func TestKindAndRelationConstants(t *testing.T) {
	kinds := []Kind{KindFunction, KindMethod, KindClass, KindModule, KindFile, KindDoc, KindConcept}
	want := []string{"function", "method", "class", "module", "file", "doc", "concept"}
	for i, k := range kinds {
		if string(k) != want[i] {
			t.Fatalf("kind[%d] = %q, want %q", i, k, want[i])
		}
	}
	rels := []Relation{RelCalls, RelImports, RelInherits, RelImplements, RelReferences, RelContains}
	wantRel := []string{"calls", "imports", "inherits", "implements", "references", "contains"}
	for i, r := range rels {
		if string(r) != wantRel[i] {
			t.Fatalf("relation[%d] = %q, want %q", i, r, wantRel[i])
		}
	}
	confs := []Confidence{ConfExtracted, ConfInferred, ConfAmbiguous}
	wantConf := []string{"EXTRACTED", "INFERRED", "AMBIGUOUS"}
	for i, c := range confs {
		if string(c) != wantConf[i] {
			t.Fatalf("confidence[%d] = %q, want %q", i, c, wantConf[i])
		}
	}
}

func TestNode_DefaultCommunityIsMinusOne(t *testing.T) {
	n := Node{ID: "a", Label: "A", Kind: KindFunction, File: "a.go", Line: 1, Community: UnclusteredCommunity}
	if n.Community != -1 {
		t.Fatalf("unclustered community sentinel = %d, want -1", n.Community)
	}
}
