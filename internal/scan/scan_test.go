package scan

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScan_FiltersExtensionsAndSortsDeterministically(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "zebra.go", "package main")
	writeFile(t, dir, "alpha.go", "package main")
	writeFile(t, dir, "README.md", "# hi")
	writeFile(t, dir, "notes.txt", "ignored ext")
	writeFile(t, dir, "img.png", "binary")
	writeFile(t, dir, "sub/deep.go", "package sub")

	refs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	var rels []string
	for _, r := range refs {
		rels = append(rels, r.RelPath)
	}
	want := []string{"README.md", "alpha.go", "sub/deep.go", "zebra.go"}
	if !reflect.DeepEqual(rels, want) {
		t.Fatalf("scan order = %v, want %v", rels, want)
	}
}

func TestScan_HonorsGitignore(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".gitignore", "ignored/\n*.gen.go\n")
	writeFile(t, dir, "keep.go", "package main")
	writeFile(t, dir, "thing.gen.go", "package main")
	writeFile(t, dir, "ignored/secret.go", "package secret")

	refs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	var rels []string
	for _, r := range refs {
		rels = append(rels, r.RelPath)
	}
	want := []string{"keep.go"}
	if !reflect.DeepEqual(rels, want) {
		t.Fatalf("gitignore not honored: got %v, want %v", rels, want)
	}
}

func TestScan_AlwaysSkipsVendorDirs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, ".git/config.go", "package x")           // .git must never be scanned
	writeFile(t, dir, ".graffiti/cache/x.go", "package x")      // our own output dir
	writeFile(t, dir, "node_modules/dep/index.go", "package x") // always skipped
	writeFile(t, dir, "vendor/v/v.go", "package x")             // always skipped
	writeFile(t, dir, "real.go", "package main")

	refs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(refs) != 1 || refs[0].RelPath != "real.go" {
		t.Fatalf("must skip .git/.graffiti/node_modules/vendor; got %+v", refs)
	}
}

func TestScan_ClassifiesNewLanguages(t *testing.T) {
	dir := t.TempDir()
	files := map[string]Lang{
		"a.py":   LangPython,
		"b.js":   LangJavaScript,
		"c.jsx":  LangJavaScript,
		"d.mjs":  LangJavaScript,
		"e.ts":   LangTypeScript,
		"f.tsx":  LangTypeScript,
		"g.rs":   LangRust,
		"h.java": LangJava,
		"i.php":  LangPHP,
	}
	for name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	refs, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]Lang{}
	for _, r := range refs {
		got[r.RelPath] = r.Lang
	}
	for name, want := range files {
		if got[name] != want {
			t.Errorf("%s: lang = %q, want %q", name, got[name], want)
		}
	}
}

func TestScan_LangClassification(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", "package main")
	writeFile(t, dir, "b.md", "# x")
	refs, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	got := map[string]Lang{}
	for _, r := range refs {
		got[r.RelPath] = r.Lang
	}
	if got["a.go"] != LangGo {
		t.Fatalf("a.go lang = %q, want go", got["a.go"])
	}
	if got["b.md"] != LangMarkdown {
		t.Fatalf("b.md lang = %q, want markdown", got["b.md"])
	}
}
