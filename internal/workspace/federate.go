package workspace

import (
	"path/filepath"

	"github.com/amazopic/graffiti/internal/graph"
	"github.com/amazopic/graffiti/internal/store"
)

// CombinedDocument builds an in-memory *graph.Document over all members with every
// node id, edge endpoint, AND file path prefixed by the member alias ("alias::id",
// "alias/file"), plus the overlay's confident cross-edges. The alias prefix lives
// ONLY here and in overlay.json — never written back into any project's map.json
// (§16.1). The file-path prefix makes the project the top level of the viewer's
// directory tree. Document order follows the (alias-sorted) member order, so the
// result is deterministic.
func CombinedDocument(root string, reg *Registry, ov *Overlay) (*graph.Document, error) {
	combined := graph.NewDocument(reg.Name)
	if ov != nil {
		combined.GeneratedAt = ov.GeneratedAt
	}
	for _, m := range reg.Members {
		dir := filepath.Join(root, filepath.FromSlash(m.Path))
		doc, err := store.Load(filepath.Join(dir, ".graffiti", "map.json"))
		if err != nil {
			return nil, err
		}
		p := m.Alias + "::"
		for _, n := range doc.Nodes {
			n.ID = p + n.ID
			n.File = m.Alias + "/" + n.File
			combined.Nodes = append(combined.Nodes, n)
		}
		for _, e := range doc.Edges {
			e.From = p + e.From
			e.To = p + e.To
			combined.Edges = append(combined.Edges, e)
		}
	}
	if ov != nil {
		for _, l := range ov.Links { // confident cross-edges only
			combined.Edges = append(combined.Edges, graph.Edge{
				From: l.From, To: l.To,
				Relation: graph.Relation(l.Relation), Confidence: graph.Confidence(l.Confidence),
			})
		}
	}
	return combined, nil
}

// CombinedIndex builds the federated read-side index over CombinedDocument so
// query.Query traverses all members + cross-edges with no changes to the engine.
func CombinedIndex(root string, reg *Registry, ov *Overlay) (*store.Index, error) {
	doc, err := CombinedDocument(root, reg, ov)
	if err != nil {
		return nil, err
	}
	return store.NewIndex(doc), nil
}
