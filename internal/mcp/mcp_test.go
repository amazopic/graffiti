package mcp

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/graph"
	"github.com/evgeniy-achin/graffiti/internal/store"
)

func testServer() *Server {
	doc := graph.NewDocument("demo")
	doc.Nodes = []graph.Node{
		{ID: "a.go:f", Label: "fooHandler", Kind: graph.KindFunction, File: "a.go", Line: 1},
		{ID: "a.go:g", Label: "barHandler", Kind: graph.KindFunction, File: "a.go", Line: 9},
		{ID: "b.go:h", Label: "baz", Kind: graph.KindFunction, File: "b.go", Line: 3},
	}
	doc.Edges = []graph.Edge{
		{From: "a.go:f", To: "a.go:g", Relation: graph.RelCalls, Confidence: graph.ConfExtracted},
		{From: "a.go:g", To: "b.go:h", Relation: graph.RelCalls, Confidence: graph.ConfExtracted},
	}
	return NewServer(store.NewIndex(doc))
}

// roundtrip feeds one request line through Serve and returns the decoded responses.
func roundtrip(t *testing.T, s *Server, lines ...string) []map[string]any {
	t.Helper()
	var out bytes.Buffer
	if err := s.Serve(strings.NewReader(strings.Join(lines, "\n")+"\n"), &out); err != nil {
		t.Fatalf("Serve: %v", err)
	}
	var resps []map[string]any
	for _, l := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if l == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Fatalf("decode response %q: %v", l, err)
		}
		resps = append(resps, m)
	}
	return resps
}

func TestMCP_InitializeEchoesAllowedVersion(t *testing.T) {
	r := roundtrip(t, testServer(), `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26"}}`)
	res := r[0]["result"].(map[string]any)
	if res["protocolVersion"] != "2025-03-26" {
		t.Fatalf("expected echoed allow-listed version 2025-03-26, got %v", res["protocolVersion"])
	}
}

func TestMCP_InitializeUnknownVersionFallsBackToLatest(t *testing.T) {
	r := roundtrip(t, testServer(), `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"1999-01-01"}}`)
	res := r[0]["result"].(map[string]any)
	if res["protocolVersion"] != LatestProtocolVersion {
		t.Fatalf("expected fallback to latest %q, got %v", LatestProtocolVersion, res["protocolVersion"])
	}
}

func TestMCP_ToolsListHasAllFour(t *testing.T) {
	r := roundtrip(t, testServer(), `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	res := r[0]["result"].(map[string]any)
	tools := res["tools"].([]any)
	got := map[string]bool{}
	for _, tr := range tools {
		got[tr.(map[string]any)["name"].(string)] = true
	}
	for _, want := range []string{"query_graph", "get_node", "get_neighbors", "shortest_path"} {
		if !got[want] {
			t.Fatalf("tools/list missing %q (got %v)", want, got)
		}
	}
}

func callText(t *testing.T, m map[string]any) string {
	t.Helper()
	res := m["result"].(map[string]any)
	content := res["content"].([]any)
	return content[0].(map[string]any)["text"].(string)
}

func TestMCP_CallQueryGraph(t *testing.T) {
	r := roundtrip(t, testServer(), `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"query_graph","arguments":{"question":"foo handler"}}}`)
	if !strings.Contains(callText(t, r[0]), "a.go:f") {
		t.Fatalf("query_graph did not return foo node:\n%s", callText(t, r[0]))
	}
}

func TestMCP_CallGetNode(t *testing.T) {
	r := roundtrip(t, testServer(), `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_node","arguments":{"id":"a.go:f"}}}`)
	if !strings.Contains(callText(t, r[0]), "a.go:f [function] fooHandler @ a.go:1") {
		t.Fatalf("get_node format wrong:\n%s", callText(t, r[0]))
	}
}

func TestMCP_CallGetNeighbors(t *testing.T) {
	r := roundtrip(t, testServer(), `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"get_neighbors","arguments":{"id":"a.go:g"}}}`)
	txt := callText(t, r[0])
	if !strings.Contains(txt, "a.go:f -calls-> a.go:g") || !strings.Contains(txt, "a.go:g -calls-> b.go:h") {
		t.Fatalf("get_neighbors missing in/out edges:\n%s", txt)
	}
}

func TestMCP_CallShortestPath(t *testing.T) {
	r := roundtrip(t, testServer(), `{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"shortest_path","arguments":{"from":"a.go:f","to":"b.go:h"}}}`)
	txt := callText(t, r[0])
	// BFS path f -> g -> h, ids one per line in path order.
	if !strings.Contains(txt, "a.go:f") || !strings.Contains(txt, "a.go:g") || !strings.Contains(txt, "b.go:h") {
		t.Fatalf("shortest_path missing path nodes:\n%s", txt)
	}
	if strings.Index(txt, "a.go:f") > strings.Index(txt, "b.go:h") {
		t.Fatalf("shortest_path order wrong (from must precede to):\n%s", txt)
	}
}

func TestMCP_NotificationGetsNoReply(t *testing.T) {
	var out bytes.Buffer
	s := testServer()
	if err := s.Serve(strings.NewReader(`{"jsonrpc":"2.0","method":"notifications/initialized"}`+"\n"), &out); err != nil {
		t.Fatalf("Serve: %v", err)
	}
	if strings.TrimSpace(out.String()) != "" {
		t.Fatalf("notification produced a reply: %q", out.String())
	}
}

func TestMCP_ParseErrorOnGarbage(t *testing.T) {
	r := roundtrip(t, testServer(), `{not json`)
	e := r[0]["error"].(map[string]any)
	if int(e["code"].(float64)) != -32700 {
		t.Fatalf("expected parse error -32700, got %v", e["code"])
	}
}

func TestMCP_CallQueryGraph_MalformedArgs(t *testing.T) {
	// Pass a JSON array instead of an object — unmarshal into a struct must fail.
	r := roundtrip(t, testServer(), `{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"query_graph","arguments":[1,2,3]}}`)
	res := r[0]["result"].(map[string]any)
	if res["isError"] != true {
		t.Fatalf("expected isError=true for malformed query_graph arguments, got result=%v", res)
	}
	content := res["content"].([]any)
	txt := content[0].(map[string]any)["text"].(string)
	if !strings.Contains(txt, "invalid arguments for query_graph") {
		t.Fatalf("expected error message about invalid arguments, got %q", txt)
	}
}
