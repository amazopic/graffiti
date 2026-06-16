package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeGoRepo creates a tiny buildable Go repo under dir/sub and returns its path.
func writeGoRepo(t *testing.T, dir, sub, pkg, src string) string {
	t.Helper()
	p := filepath.Join(dir, sub)
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRun_LinkBuildsAndFederates(t *testing.T) {
	base := t.TempDir()
	web := writeGoRepo(t, base, "frontend", "web", "package web\n\nfunc FetchCart() {}\n")
	api := writeGoRepo(t, base, "backend", "api", "package api\n\nfunc GetCart() {}\n")

	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "link", "--name", "shop", web, api}, bytes.NewReader(nil), &out, &errOut)
	if code != 0 {
		t.Fatalf("link exit=%d stderr=%q", code, errOut.String())
	}
	// workspace.json + overlay.json written under the common ancestor (base)
	if _, err := os.Stat(filepath.Join(base, ".graffiti-workspace", "workspace.json")); err != nil {
		t.Fatalf("workspace.json missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(base, ".graffiti-workspace", "overlay.json")); err != nil {
		t.Fatalf("overlay.json missing: %v", err)
	}
	if !strings.Contains(out.String(), "Linked 2 projects") {
		t.Fatalf("missing success line:\n%s", out.String())
	}
}
