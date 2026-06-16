package workspace

import (
	"path/filepath"

	"github.com/evgeniy-achin/graffiti/internal/graph"
	"github.com/evgeniy-achin/graffiti/internal/store"
)

// CombinedIndex builds an in-memory store.Index over all members with every node
// id and edge endpoint prefixed by "alias::", plus the overlay's confident
// cross-edges. The prefix lives ONLY here and in overlay.json — never written
// back into any project's map.json (§16.1). Reusing store.NewIndex + query.Query
// over this index gives federated retrieval with no changes to the query engine.
func CombinedIndex(root string, reg *Registry, ov *Overlay) (*store.Index, error) {
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
	return store.NewIndex(combined), nil
}
