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

// linkShop builds two members and federates them, returning the workspace root (base).
func linkShop(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	web := writeGoRepo(t, base, "frontend", "web", "package web\n\nfunc FetchCart() {}\n")
	api := writeGoRepo(t, base, "backend", "api", "package api\n\nfunc GetCart() {}\n")
	var out, errOut bytes.Buffer
	if code := run([]string{"graffiti", "link", web, api}, bytes.NewReader(nil), &out, &errOut); code != 0 {
		t.Fatalf("link failed: %s", errOut.String())
	}
	return base
}

func TestRun_WorkspaceList(t *testing.T) {
	base := linkShop(t)
	// run from inside the workspace (cwd discovery) by passing --root.
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "workspace", "list", "--root", base}, bytes.NewReader(nil), &out, &errOut)
	if code != 0 {
		t.Fatalf("list exit=%d stderr=%q", code, errOut.String())
	}
	if !strings.Contains(out.String(), "frontend") || !strings.Contains(out.String(), "backend") {
		t.Fatalf("list missing members:\n%s", out.String())
	}
}

func TestRun_LinksCheck(t *testing.T) {
	base := linkShop(t)
	// write a links file: one resolvable, one ghost.
	wsDir := filepath.Join(base, ".graffiti-workspace")
	// Verified node-id slugs: FetchCart -> "main-go:fetchcart", GetCart -> "main-go:getcart".
	// (alias::id splits on the FIRST "::", so the single colon inside the id is fine.)
	links := "frontend::main-go:fetchcart -> backend::main-go:getcart calls\nfrontend::main-go:ghost -> backend::main-go:getcart\n"
	if err := os.WriteFile(filepath.Join(wsDir, "links"), []byte(links), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "links", "check", "--root", base}, bytes.NewReader(nil), &out, &errOut)
	// non-zero exit because one link is unresolved
	if code == 0 {
		t.Fatalf("expected non-zero exit for an unresolved link:\n%s", out.String())
	}
	if !strings.Contains(out.String()+errOut.String(), "ghost") {
		t.Fatalf("expected the unresolved 'ghost' link to be reported")
	}
}

func TestRun_WorkspaceRender(t *testing.T) {
	base := linkShop(t)
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "workspace", "render", "--root", base}, bytes.NewReader(nil), &out, &errOut)
	if code != 0 {
		t.Fatalf("workspace render exit=%d stderr=%q", code, errOut.String())
	}
	b, err := os.ReadFile(filepath.Join(base, ".graffiti-workspace", "workspace.html"))
	if err != nil {
		t.Fatalf("workspace.html missing: %v", err)
	}
	h := string(b)
	if !strings.Contains(h, `<canvas id="c"`) {
		t.Fatal("workspace.html missing the force-graph canvas")
	}
	if !strings.Contains(h, "frontend/") || !strings.Contains(h, "backend/") {
		t.Fatalf("workspace.html island missing alias-prefixed project paths")
	}
	for _, banned := range []string{"http://", "https://", "<link", "@import"} {
		if strings.Contains(h, banned) {
			t.Fatalf("workspace.html not self-contained: found %q", banned)
		}
	}
}

func TestRun_QueryWorkspace(t *testing.T) {
	base := linkShop(t)
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "query", "--workspace", "--root", base, "cart"}, bytes.NewReader(nil), &out, &errOut)
	if code != 0 {
		t.Fatalf("query --workspace exit=%d stderr=%q", code, errOut.String())
	}
	s := out.String()
	if !strings.Contains(s, "NODES") {
		t.Fatalf("missing NODES block:\n%s", s)
	}
	// alias-prefixed ids appear (both members are searched)
	if !strings.Contains(s, "frontend::") && !strings.Contains(s, "backend::") {
		t.Fatalf("expected alias-prefixed federated output:\n%s", s)
	}
}

func TestRun_ServeWorkspace(t *testing.T) {
	base := linkShop(t)
	initLine := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}` + "\n"
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "serve", "--workspace", "--root", base}, strings.NewReader(initLine), &out, &errOut)
	if code != 0 {
		t.Fatalf("serve --workspace exit=%d stderr=%q", code, errOut.String())
	}
	if !strings.Contains(out.String(), `"protocolVersion":"2025-06-18"`) {
		t.Fatalf("serve --workspace missing initialize echo:\n%s", out.String())
	}
}

func TestRun_UpdateWorkspace(t *testing.T) {
	base := linkShop(t)
	var out, errOut bytes.Buffer
	code := run([]string{"graffiti", "update", "--workspace", "--root", base}, bytes.NewReader(nil), &out, &errOut)
	if code != 0 {
		t.Fatalf("update --workspace exit=%d stderr=%q", code, errOut.String())
	}
	if !strings.Contains(out.String(), "overlay") {
		t.Fatalf("update --workspace should mention overlay recompute:\n%s", out.String())
	}
}
