// Package mcp is graffiti's hand-rolled, dependency-free MCP server over stdio
// (see the Plan 4 §10 SPEC AMENDMENT). MCP's stdio transport is plain JSON-RPC
// 2.0 with NEWLINE-DELIMITED framing (one JSON object per line) — NOT the
// LSP-style Content-Length header framing. graffiti speaks three methods
// (initialize, tools/list, tools/call), tolerates notifications by ignoring
// them, and exposes four tools (query_graph, get_node, get_neighbors,
// shortest_path). encoding/json + bufio + io only: no deps, pure Go, offline.
package mcp

import (
	"bufio"
	"encoding/json"
	"io"
	"sort"

	"github.com/evgeniy-achin/graffiti/internal/graph"
	"github.com/evgeniy-achin/graffiti/internal/query"
	"github.com/evgeniy-achin/graffiti/internal/store"
)

// LatestProtocolVersion is the MCP revision graffiti prefers. On initialize the
// server ECHOES the client's requested version if it is in supportedVersions,
// otherwise it returns this latest (never blindly returns one hardcoded value).
const LatestProtocolVersion = "2025-06-18"

// supportedVersions is the small allow-list of MCP revisions graffiti will echo.
var supportedVersions = map[string]bool{
	"2025-06-18": true,
	"2025-03-26": true,
	"2024-11-05": true,
}

// --- JSON-RPC 2.0 wire types ---

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"` // absent => notification
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// --- MCP payload types ---

type initializeResult struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ServerInfo      serverInfo     `json:"serverInfo"`
}
type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
type toolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}
type listToolsResult struct {
	Tools []toolDef `json:"tools"`
}
type callToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}
type textContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
type callToolResult struct {
	Content []textContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// Server exposes the graph as MCP tools over stdio.
type Server struct {
	idx  *store.Index
	name string
}

// NewServer builds a Server over an in-memory Index.
func NewServer(idx *store.Index) *Server { return &Server{idx: idx, name: "graffiti"} }

func objSchema(props map[string]any, required ...string) map[string]any {
	req := make([]any, len(required))
	for i, r := range required {
		req[i] = r
	}
	return map[string]any{"type": "object", "properties": props, "required": req}
}

func strProp(desc string) map[string]any {
	return map[string]any{"type": "string", "description": desc}
}

// tools is the static, deterministically-ordered tool catalog.
func (s *Server) tools() []toolDef {
	return []toolDef{
		{Name: "query_graph", Description: "LLM-free scoped subgraph retrieval over the code graph. Returns compact text.",
			InputSchema: objSchema(map[string]any{
				"question": strProp("natural-language question"),
				"budget":   map[string]any{"type": "integer", "description": "soft node-token budget (default 2000)"},
			}, "question")},
		{Name: "get_node", Description: "Return one node by id as 'id [kind] label @ file:line'.",
			InputSchema: objSchema(map[string]any{"id": strProp("node id")}, "id")},
		{Name: "get_neighbors", Description: "Return a node and its sorted in/out edges as compact text.",
			InputSchema: objSchema(map[string]any{"id": strProp("node id")}, "id")},
		{Name: "shortest_path", Description: "Deterministic BFS shortest path between two node ids (id-ordered frontier).",
			InputSchema: objSchema(map[string]any{"from": strProp("start node id"), "to": strProp("end node id")}, "from", "to")},
	}
}

// Serve runs the read/dispatch/write loop until r is exhausted. r and w are
// injectable so tests drive it without real stdin/stdout. Newline-framed.
func (s *Server) Serve(r io.Reader, w io.Writer) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024) // allow large tools/call results
	enc := json.NewEncoder(w)
	for sc.Scan() {
		line := sc.Bytes()
		if len(trimSpace(line)) == 0 {
			continue
		}
		var req rpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.writeErr(enc, nil, -32700, "parse error")
			continue
		}
		isNotification := len(req.ID) == 0
		resp := s.dispatch(req)
		if isNotification {
			continue // notifications get no reply
		}
		if err := enc.Encode(resp); err != nil {
			return err
		}
	}
	return sc.Err()
}

func (s *Server) dispatch(req rpcRequest) rpcResponse {
	switch req.Method {
	case "initialize":
		return s.ok(req.ID, initializeResult{
			ProtocolVersion: s.negotiateVersion(req.Params),
			Capabilities:    map[string]any{"tools": map[string]any{}},
			ServerInfo:      serverInfo{Name: s.name, Version: "0.4.0"},
		})
	case "tools/list":
		return s.ok(req.ID, listToolsResult{Tools: s.tools()})
	case "tools/call":
		return s.handleCall(req)
	default:
		return rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{Code: -32601, Message: "method not found: " + req.Method}}
	}
}

// negotiateVersion echoes the client's requested protocolVersion when it is in
// the allow-list, else returns the server's latest (never blindly hardcoded).
func (s *Server) negotiateVersion(params json.RawMessage) string {
	var p struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	_ = json.Unmarshal(params, &p)
	if supportedVersions[p.ProtocolVersion] {
		return p.ProtocolVersion
	}
	return LatestProtocolVersion
}

func (s *Server) handleCall(req rpcRequest) rpcResponse {
	var p callToolParams
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return s.toolErr(req.ID, "invalid params")
	}
	switch p.Name {
	case "query_graph":
		var a struct {
			Question string `json:"question"`
			Budget   int    `json:"budget"`
		}
		if err := json.Unmarshal(p.Arguments, &a); err != nil {
			return s.toolErr(req.ID, "invalid arguments for query_graph")
		}
		return s.toolText(req.ID, query.Query(s.idx, a.Question, a.Budget))
	case "get_node":
		var a struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(p.Arguments, &a)
		n, ok := s.idx.Node(a.ID)
		if !ok {
			return s.toolErr(req.ID, "node not found: "+a.ID)
		}
		return s.toolText(req.ID, query.FormatNode(n))
	case "get_neighbors":
		var a struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(p.Arguments, &a)
		n, ok := s.idx.Node(a.ID)
		if !ok {
			return s.toolErr(req.ID, "node not found: "+a.ID)
		}
		return s.toolText(req.ID, s.neighborsText(n))
	case "shortest_path":
		var a struct {
			From string `json:"from"`
			To   string `json:"to"`
		}
		_ = json.Unmarshal(p.Arguments, &a)
		path, ok := s.shortestPath(a.From, a.To)
		if !ok {
			return s.toolErr(req.ID, "no path from "+a.From+" to "+a.To)
		}
		out := ""
		for _, id := range path {
			out += id + "\n"
		}
		return s.toolText(req.ID, out)
	default:
		return s.toolErr(req.ID, "unknown tool: "+p.Name)
	}
}

// neighborsText renders a node line followed by its sorted in+out edges.
func (s *Server) neighborsText(n graph.Node) string {
	out := query.FormatNode(n) + "\nEDGES\n"
	for _, e := range s.idx.In(n.ID) {
		out += query.FormatEdge(e) + "\n"
	}
	for _, e := range s.idx.Out(n.ID) {
		out += query.FormatEdge(e) + "\n"
	}
	return out
}

// shortestPath is a DETERMINISTIC BFS over a directed-as-undirected adjacency
// (follow both out and in edges) with an ID-ORDERED frontier: at each node the
// neighbor ids are gathered, de-duped, and SORTED before enqueueing, so the
// discovered path is byte-identical for the same graph (spec §14).
func (s *Server) shortestPath(from, to string) ([]string, bool) {
	if _, ok := s.idx.Node(from); !ok {
		return nil, false
	}
	if _, ok := s.idx.Node(to); !ok {
		return nil, false
	}
	if from == to {
		return []string{from}, true
	}
	prev := map[string]string{from: ""}
	queue := []string{from}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		// Gather neighbor ids (both directions), dedupe, sort => id-ordered frontier.
		seen := map[string]bool{}
		var nbrs []string
		for _, e := range s.idx.Out(cur) {
			if !seen[e.To] {
				seen[e.To] = true
				nbrs = append(nbrs, e.To)
			}
		}
		for _, e := range s.idx.In(cur) {
			if !seen[e.From] {
				seen[e.From] = true
				nbrs = append(nbrs, e.From)
			}
		}
		sort.Strings(nbrs)
		for _, nb := range nbrs {
			if _, ok := prev[nb]; ok {
				continue
			}
			prev[nb] = cur
			if nb == to {
				return reconstruct(prev, from, to), true
			}
			queue = append(queue, nb)
		}
	}
	return nil, false
}

func reconstruct(prev map[string]string, from, to string) []string {
	var rev []string
	for cur := to; cur != ""; cur = prev[cur] {
		rev = append(rev, cur)
		if cur == from {
			break
		}
	}
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	return rev
}

func (s *Server) ok(id json.RawMessage, result any) rpcResponse {
	return rpcResponse{JSONRPC: "2.0", ID: id, Result: result}
}
func (s *Server) toolText(id json.RawMessage, text string) rpcResponse {
	return s.ok(id, callToolResult{Content: []textContent{{Type: "text", Text: text}}})
}
func (s *Server) toolErr(id json.RawMessage, msg string) rpcResponse {
	return s.ok(id, callToolResult{Content: []textContent{{Type: "text", Text: msg}}, IsError: true})
}
func (s *Server) writeErr(enc *json.Encoder, id json.RawMessage, code int, msg string) {
	_ = enc.Encode(rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}})
}

func trimSpace(b []byte) []byte {
	i, j := 0, len(b)
	for i < j && (b[i] == ' ' || b[i] == '\t' || b[i] == '\r' || b[i] == '\n') {
		i++
	}
	for j > i && (b[j-1] == ' ' || b[j-1] == '\t' || b[j-1] == '\r' || b[j-1] == '\n') {
		j--
	}
	return b[i:j]
}
