// Package render serializes a graph.Document to disk artifacts. Plan 1 emits
// only map.json (MAP.md and map.html are later plans).
package render

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/amazopic/graffiti/internal/graph"
)

// orderedDocument mirrors graph.Document but with struct fields ordered so the
// emitted JSON object keys are alphabetical (communities, edges, generated_at,
// nodes, root, version) for byte-determinism (spec §8.8/§14).
type orderedDocument struct {
	Communities []graph.Community `json:"communities"`
	Consumes    []graph.Endpoint  `json:"consumes"`
	Edges       []graph.Edge      `json:"edges"`
	GeneratedAt string            `json:"generated_at"`
	Nodes       []graph.Node      `json:"nodes"`
	Provides    []graph.Endpoint  `json:"provides"`
	Root        string            `json:"root"`
	Version     string            `json:"version"`
}

// WriteMapJSON writes doc to <root>/.graffiti/map.json. The generated_at value is
// read directly off doc.GeneratedAt (stamped by build.Assemble) — single source
// of truth. Output is deterministic modulo generated_at and root: top-level keys
// are alphabetical and the node/edge arrays are assumed already sorted by Assemble.
func WriteMapJSON(doc *graph.Document, root string) error {
	od := orderedDocument{
		Communities: nonNilCommunities(doc.Communities),
		Consumes:    nonNilEndpoints(doc.Consumes),
		Edges:       nonNilEdges(doc.Edges),
		GeneratedAt: doc.GeneratedAt,
		Nodes:       nonNilNodes(doc.Nodes),
		Provides:    nonNilEndpoints(doc.Provides),
		Root:        doc.Root,
		Version:     doc.Version,
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(od); err != nil { // Encode appends a trailing '\n'
		return err
	}

	dir := filepath.Join(root, ".graffiti")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "map.json"), buf.Bytes(), 0o644)
}

func nonNilNodes(n []graph.Node) []graph.Node {
	if n == nil {
		return []graph.Node{}
	}
	return n
}
func nonNilEdges(e []graph.Edge) []graph.Edge {
	if e == nil {
		return []graph.Edge{}
	}
	return e
}
func nonNilCommunities(c []graph.Community) []graph.Community {
	if c == nil {
		return []graph.Community{}
	}
	return c
}
func nonNilEndpoints(e []graph.Endpoint) []graph.Endpoint {
	if e == nil {
		return []graph.Endpoint{}
	}
	return e
}
