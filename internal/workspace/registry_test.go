package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegistry_SaveLoadRoundTrip(t *testing.T) {
	root := t.TempDir()
	reg := &Registry{
		Version: SchemaVersion, Name: "shop", GeneratedAt: "2026-06-17T00:00:00Z",
		Members: []Member{
			{Alias: "web", Path: "../frontend", MapHash: "h1"},
			{Alias: "api", Path: "../backend", MapHash: "h2"},
		},
	}
	if err := SaveRegistry(root, reg); err != nil {
		t.Fatal(err)
	}
	got, err := LoadRegistry(root)
	if err != nil {
		t.Fatal(err)
	}
	// Members come back sorted by alias (api before web).
	if len(got.Members) != 2 || got.Members[0].Alias != "api" {
		t.Fatalf("members not sorted by alias: %+v", got.Members)
	}
	if got.Name != "shop" {
		t.Fatalf("name = %q", got.Name)
	}
}

func TestAddRemoveMember(t *testing.T) {
	reg := &Registry{Version: SchemaVersion}
	AddMember(reg, Member{Alias: "web", Path: "../w"})
	AddMember(reg, Member{Alias: "api", Path: "../a"})
	AddMember(reg, Member{Alias: "web", Path: "../w2"}) // replace, not duplicate
	if len(reg.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(reg.Members))
	}
	if reg.Members[0].Alias != "api" { // sorted
		t.Fatalf("not sorted: %+v", reg.Members)
	}
	for _, m := range reg.Members {
		if m.Alias == "web" && m.Path != "../w2" {
			t.Fatalf("web not replaced: %+v", m)
		}
	}
	if !RemoveMember(reg, "web") || len(reg.Members) != 1 {
		t.Fatalf("remove failed: %+v", reg.Members)
	}
}

func TestMapHash(t *testing.T) {
	dir := t.TempDir()
	if _, err := MapHash(dir); err == nil {
		t.Fatal("expected error when map.json is absent")
	}
	if err := os.MkdirAll(filepath.Join(dir, ".graffiti"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".graffiti", "map.json"), []byte(`{"a":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	h, err := MapHash(dir)
	if err != nil || len(h) != 64 {
		t.Fatalf("hash=%q err=%v", h, err)
	}
}
