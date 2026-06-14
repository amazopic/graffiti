package graph

import "fmt"

// Merge folds `from` into `into` (spec §6: merge-not-replace with anti-shrink
// guard). Nodes are deduped by ID (first writer wins on conflicting fields).
// Edges are deduped by the (from,to,relation,confidence) tuple.
//
// If allowPrune is false and the resulting node count would be LESS than the
// pre-merge node count of `into`, Merge returns an error rather than silently
// shrinking the graph. allowPrune=true is the explicit-prune escape hatch.
func Merge(into, from *Document, allowPrune bool) error {
	// before is the high-water mark: the maximum node count ever committed via
	// Merge. This lets the guard detect out-of-band pruning (e.g. Nodes[:n]).
	before := into.nodeHighWaterMark
	if len(into.Nodes) > before {
		before = len(into.Nodes)
	}

	nodeIdx := make(map[string]bool, len(into.Nodes))
	for _, n := range into.Nodes {
		nodeIdx[n.ID] = true
	}
	for _, n := range from.Nodes {
		if !nodeIdx[n.ID] {
			into.Nodes = append(into.Nodes, n)
			nodeIdx[n.ID] = true
		}
	}

	edgeIdx := make(map[edgeKey]bool, len(into.Edges))
	for _, e := range into.Edges {
		edgeIdx[keyOf(e)] = true
	}
	for _, e := range from.Edges {
		k := keyOf(e)
		if !edgeIdx[k] {
			into.Edges = append(into.Edges, e)
			edgeIdx[k] = true
		}
	}

	after := len(into.Nodes)
	if !allowPrune && after < before {
		return fmt.Errorf("merge would shrink node count from %d to %d without explicit prune", before, after)
	}
	// Update the high-water mark.
	if after > into.nodeHighWaterMark {
		into.nodeHighWaterMark = after
	}
	return nil
}

type edgeKey struct {
	from, to string
	rel      Relation
	conf     Confidence
}

func keyOf(e Edge) edgeKey {
	return edgeKey{from: e.From, to: e.To, rel: e.Relation, conf: e.Confidence}
}
