package app

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func write(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestBuild_ProducesMapAndStats(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "go.mod", "module example.com/demo\n\ngo 1.26\n")
	write(t, dir, "main.go", "package main\n\nfunc main() { Hello() }\n\nfunc Hello() {}\n")
	write(t, dir, "README.md", "# demo\n")

	stats, err := Build(dir, "2026-06-14T00:00:00Z")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if stats.Files < 2 {
		t.Fatalf("expected >=2 files scanned, got %d", stats.Files)
	}
	if stats.Nodes == 0 || stats.Edges == 0 {
		t.Fatalf("expected nonzero nodes/edges, got %d/%d", stats.Nodes, stats.Edges)
	}
	if _, err := os.Stat(filepath.Join(dir, ".graffiti", "map.json")); err != nil {
		t.Fatalf("map.json missing: %v", err)
	}
	// cache written with an entry per scanned file (forward-compat artifact)
	if _, err := os.Stat(filepath.Join(dir, ".graffiti", "cache", "hashes.json")); err != nil {
		t.Fatalf("cache hashes.json missing: %v", err)
	}
	if !stats.HasDocNode {
		t.Fatalf("expected a doc node for README.md")
	}
}

var reGenAtApp = regexp.MustCompile(`("generated_at":\s*")[^"]*(")`)
var reRootApp = regexp.MustCompile(`("root":\s*")[^"]*(")`)

func normApp(b []byte) string {
	b = reGenAtApp.ReplaceAll(b, []byte(`${1}X${2}`))
	b = reRootApp.ReplaceAll(b, []byte(`${1}X${2}`))
	return string(b)
}

func TestBuild_DeterministicModuloGeneratedAtAndRoot(t *testing.T) {
	src := "package main\n\nfunc main() { Hello() }\n\nfunc Hello() {}\n"

	dir1 := t.TempDir()
	write(t, dir1, "main.go", src)
	if _, err := Build(dir1, "2026-06-14T00:00:00Z"); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(filepath.Join(dir1, ".graffiti", "map.json"))

	dir2 := t.TempDir()
	write(t, dir2, "main.go", src)
	if _, err := Build(dir2, "2099-12-31T23:59:59Z"); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(filepath.Join(dir2, ".graffiti", "map.json"))

	if normApp(first) != normApp(second) {
		t.Fatalf("not deterministic modulo generated_at+root:\n%s\n---\n%s", normApp(first), normApp(second))
	}
}
