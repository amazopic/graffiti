package contract

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amazopic/graffiti/internal/graph"
)

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func has(eps []graph.Endpoint, kind graph.EndpointKind, key string) *graph.Endpoint {
	for i := range eps {
		if eps[i].Kind == kind && eps[i].Key == key {
			return &eps[i]
		}
	}
	return nil
}

func TestExtract_AllSources(t *testing.T) {
	dir := t.TempDir()

	// OpenAPI → provides http GET /carts/{}
	writeFile(t, dir, "openapi.json", `{"paths":{"/carts/{id}":{"get":{"summary":"x"},"parameters":[]}}}`)
	// proto → provides rpc Orders.Create
	writeFile(t, dir, "proto/orders.proto", "service Orders {\n  rpc Create (Req) returns (Res);\n}\n")
	// explicit contract → consumes rpc Orders.Create
	writeFile(t, dir, "graffiti.contract.json", `{"consumes":[{"kind":"rpc","name":"Orders.Create"}]}`)
	// source heuristics: route provide, literal-URL consume, queue publish consume
	writeFile(t, dir, "main.go", `package main
func main() {
	http.HandleFunc("/health", h)
	http.Get("http://orders:8080/orders/42")
	bus.Publish("orders.created", x)
}
`)

	doc := &graph.Document{Nodes: []graph.Node{
		{ID: "main.go:main", Label: "main", Kind: graph.KindFunction, File: "main.go", Line: 2},
	}}

	provides, consumes := Extract(dir, doc)

	if has(provides, graph.EPHTTP, "GET /carts/{}") == nil {
		t.Errorf("missing openapi provide GET /carts/{}; got %+v", provides)
	}
	if has(provides, graph.EPRPC, "Orders.Create") == nil {
		t.Errorf("missing proto provide Orders.Create; got %+v", provides)
	}
	if has(provides, graph.EPHTTP, "GET /health") == nil {
		t.Errorf("missing route provide GET /health; got %+v", provides)
	}
	if c := has(consumes, graph.EPRPC, "Orders.Create"); c == nil {
		t.Errorf("missing explicit consume Orders.Create; got %+v", consumes)
	} else if c.Confidence != graph.ConfExtracted || c.Source != "contract" {
		t.Errorf("contract consume should be EXTRACTED/contract, got %s/%s", c.Confidence, c.Source)
	}
	// literal URL "http://orders:8080/orders/42" → host stripped, 42 → {}
	if has(consumes, graph.EPHTTP, "GET /orders/{}") == nil {
		t.Errorf("missing literal consume GET /orders/{}; got %+v", consumes)
	}
	if q := has(consumes, graph.EPQueue, "orders.created"); q == nil {
		t.Errorf("missing queue consume orders.created; got %+v", consumes)
	}

	// node association: the route provide in main.go should bind to the nearest node.
	if r := has(provides, graph.EPHTTP, "GET /health"); r != nil && r.Node != "main.go:main" {
		t.Errorf("route provide node = %q, want main.go:main", r.Node)
	}
}

func TestNormPath(t *testing.T) {
	cases := map[string]string{
		"/carts/{id}":      "/carts/{}",
		"/carts/:id":       "/carts/{}",
		"/orders/42":       "/orders/{}",
		"/a/b/":            "/a/b",
		"/users/{id}/cart": "/users/{}/cart",
		"users":            "/users",
	}
	for in, want := range cases {
		if got := normPath(in); got != want {
			t.Errorf("normPath(%q) = %q, want %q", in, got, want)
		}
	}
}
