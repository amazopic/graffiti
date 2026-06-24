package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeSvc creates a service dir under root with the given files.
func writeSvc(t *testing.T, root, name string, files map[string]string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	for rel, content := range files {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// runCLI invokes the CLI entry point and returns (exitCode, stdout).
func runCLI(t *testing.T, args ...string) (int, string) {
	t.Helper()
	var out, errb bytes.Buffer
	code := run(append([]string{"graffiti"}, args...), strings.NewReader(""), &out, &errb)
	if errb.Len() > 0 {
		t.Logf("stderr: %s", errb.String())
	}
	return code, out.String()
}

// TestSystemOrchestration_EndToEnd publishes three services, federates them, and
// asserts that cross-service links are auto-discovered from contracts.
func TestSystemOrchestration_EndToEnd(t *testing.T) {
	root := t.TempDir()
	sys := filepath.Join(root, "store")
	if err := os.MkdirAll(sys, 0o755); err != nil {
		t.Fatal(err)
	}

	carts := writeSvc(t, root, "carts", map[string]string{
		"openapi.json": `{"paths":{"/carts/{id}":{"get":{"summary":"x"}}}}`,
		"main.go":      "package main\nimport \"net/http\"\nfunc main(){ http.HandleFunc(\"/carts/\", h) }\nfunc h(){}\n",
	})
	orders := writeSvc(t, root, "orders", map[string]string{
		"main.go": "package main\nfunc main(){ bus.Subscribe(\"orders.created\", handle) }\nfunc handle(){}\n",
	})
	gateway := writeSvc(t, root, "gateway", map[string]string{
		"main.go": "package main\nimport \"net/http\"\nfunc main(){\n http.Get(\"http://carts:8080/carts/42\")\n bus.Publish(\"orders.created\", nil)\n}\n",
	})

	for name, dir := range map[string]string{"carts": carts, "orders": orders, "gateway": gateway} {
		if code, _ := runCLI(t, "publish", dir, "--to", sys, "--as", name); code != 0 {
			t.Fatalf("publish %s exit %d", name, code)
		}
	}

	code, out := runCLI(t, "system", "build", "--root", sys)
	if code != 0 {
		t.Fatalf("system build exit %d", code)
	}
	if !strings.Contains(out, "2 cross-service links") {
		t.Errorf("expected 2 cross-service links; got: %s", out)
	}

	// audit: the carts route is an orphan; no dangling → exit 0.
	code, out = runCLI(t, "system", "audit", "--root", sys)
	if code != 0 {
		t.Errorf("audit exit %d (no dangling expected), out: %s", code, out)
	}
	if !strings.Contains(out, "ORPHAN") {
		t.Errorf("expected an ORPHAN line; got: %s", out)
	}

	// impact: changing carts affects gateway.
	_, out = runCLI(t, "system", "impact", "carts", "--root", sys)
	if !strings.Contains(out, "gateway") {
		t.Errorf("impact(carts) should list gateway; got: %s", out)
	}

	// render: system.html is written and self-contained.
	if code, _ := runCLI(t, "system", "render", "--root", sys); code != 0 {
		t.Fatalf("system render exit %d", code)
	}
	html, err := os.ReadFile(filepath.Join(sys, ".graffiti-system", "system.html"))
	if err != nil {
		t.Fatalf("read system.html: %v", err)
	}
	if !bytes.Contains(html, []byte(`id="graffiti-data"`)) {
		t.Error("system.html missing data island")
	}
}
