package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// normalizeRoot neutralizes the two location/time-dependent fields (§14: output
// is byte-identical "modulo generated_at and root") — the temp-dir base and the
// wall-clock timestamp baked into the golden — then re-marshals canonically (Go
// sorts object keys) for a stable comparison of everything else.
func normalizeRoot(t *testing.T, b []byte) string {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal map.json: %v", err)
	}
	m["root"] = "polyglot"
	m["generated_at"] = "FIXED"
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestPolyglot_Golden(t *testing.T) {
	srcDir := filepath.FromSlash("../../testdata/fixtures/polyglot")
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		t.Fatal(err)
	}
	copyInto := func(dir string) {
		for _, e := range entries {
			b, err := os.ReadFile(filepath.Join(srcDir, e.Name()))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, e.Name()), b, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}

	dir := t.TempDir()
	copyInto(dir)
	if _, err := Build(dir, "2026-06-16T00:00:00Z"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, ".graffiti", "map.json"))
	if err != nil {
		t.Fatal(err)
	}

	want, err := os.ReadFile(filepath.FromSlash("../../testdata/golden/polyglot.map.json"))
	if err != nil {
		t.Fatalf("read golden (generate it per the plan's step): %v", err)
	}
	if normalizeRoot(t, got) != normalizeRoot(t, want) {
		t.Fatalf("polyglot map.json differs from golden.\n--- got ---\n%s", normalizeRoot(t, got))
	}

	// Determinism: a second build of the same inputs is byte-identical (modulo root).
	dir2 := t.TempDir()
	copyInto(dir2)
	if _, err := Build(dir2, "2026-06-16T00:00:00Z"); err != nil {
		t.Fatal(err)
	}
	got2, err := os.ReadFile(filepath.Join(dir2, ".graffiti", "map.json"))
	if err != nil {
		t.Fatal(err)
	}
	if normalizeRoot(t, got) != normalizeRoot(t, got2) {
		t.Fatal("non-deterministic: two builds of the same inputs differ")
	}
}
