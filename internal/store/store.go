// Package store is graffiti's read-side: it loads the .graffiti/map.json artifact
// back into a *graph.Document (the mirror of render.WriteMapJSON) and builds a
// deterministic in-memory Index (id->node, sorted out/in adjacency) that
// internal/query and internal/mcp consume. It performs the only I/O on the read
// path and never touches the build pipeline.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/amazopic/graffiti/internal/graph"
)

// Load reads a map.json file at path and unmarshals it into a *graph.Document.
// graph.Document's json tags (version/generated_at/root/nodes/edges/communities)
// match the keys render.WriteMapJSON emits, so this is a lossless round-trip of
// every public field (the unexported high-water mark is not serialized and is
// unused on the read path).
func Load(path string) (*graph.Document, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("store: read %s: %w", path, err)
	}
	var doc graph.Document
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("store: parse %s: %w", path, err)
	}
	return &doc, nil
}

// Index is a deterministic, read-only adjacency index over a Document. Every
// exposed slice is sorted with an explicit total order so no Go-map iteration
// ever feeds query/mcp output (spec §14). IDF term scoring is computed by
// internal/query over the nodes exposed here (it builds its own df/term-bag
// tables per question), so the Index keeps no term index of its own.
type Index struct {
	nodes map[string]graph.Node
	ids   []string                // all node ids, sorted ascending
	out   map[string][]graph.Edge // out-edges per id, sorted (relation,to,confidence)
	in    map[string][]graph.Edge // in-edges per id, sorted (relation,from,confidence)
}

// NewIndex builds the Index from a loaded Document.
func NewIndex(doc *graph.Document) *Index {
	idx := &Index{
		nodes: make(map[string]graph.Node, len(doc.Nodes)),
		out:   make(map[string][]graph.Edge),
		in:    make(map[string][]graph.Edge),
	}
	for _, n := range doc.Nodes {
		if _, exists := idx.nodes[n.ID]; exists {
			continue // dedup: first writer wins; skip duplicate ids
		}
		idx.nodes[n.ID] = n
		idx.ids = append(idx.ids, n.ID)
	}
	sort.Strings(idx.ids)

	for _, e := range doc.Edges {
		idx.out[e.From] = append(idx.out[e.From], e)
		idx.in[e.To] = append(idx.in[e.To], e)
	}
	for id := range idx.out {
		sortEdges(idx.out[id], func(e graph.Edge) string { return e.To })
	}
	for id := range idx.in {
		sortEdges(idx.in[id], func(e graph.Edge) string { return e.From })
	}
	return idx
}

// sortEdges sorts edges by (relation, other-endpoint, confidence). `otherOf`
// selects the endpoint that is NOT the indexed node (To for out, From for in).
func sortEdges(es []graph.Edge, otherOf func(graph.Edge) string) {
	sort.SliceStable(es, func(i, j int) bool {
		a, b := es[i], es[j]
		if a.Relation != b.Relation {
			return a.Relation < b.Relation
		}
		oa, ob := otherOf(a), otherOf(b)
		if oa != ob {
			return oa < ob
		}
		return a.Confidence < b.Confidence
	})
}

// Node returns the node for id and whether it exists.
func (x *Index) Node(id string) (graph.Node, bool) { n, ok := x.nodes[id]; return n, ok }

// IDs returns all node ids, sorted ascending (a fresh copy is unnecessary —
// callers must not mutate; query only reads).
func (x *Index) IDs() []string { return x.ids }

// Out returns the sorted out-edges of id (nil if none).
func (x *Index) Out(id string) []graph.Edge { return x.out[id] }

// In returns the sorted in-edges of id (nil if none).
func (x *Index) In(id string) []graph.Edge { return x.in[id] }

// Len returns the node count (corpus size N for IDF).
func (x *Index) Len() int { return len(x.ids) }
