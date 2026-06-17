// Package schemaval structurally validates a graph.Document against the rules
// published in schema/map.schema.json, without a third-party JSON-schema engine.
package schemaval

import (
	"fmt"

	"github.com/amazopic/graffiti/internal/graph"
)

// ValidateDocument checks required fields, closed enums, unique node IDs, and
// that every edge endpoint references an existing node. It returns the first
// error found (deterministic order: document fields, then nodes, then edges).
func ValidateDocument(d *graph.Document) error {
	if d == nil {
		return fmt.Errorf("nil document")
	}
	if d.Version != graph.SchemaVersion {
		return fmt.Errorf("version: got %q want %q", d.Version, graph.SchemaVersion)
	}
	if d.GeneratedAt == "" {
		return fmt.Errorf("generated_at: must be set")
	}
	if d.Root == "" {
		return fmt.Errorf("root: must be set")
	}

	ids := make(map[string]bool, len(d.Nodes))
	for i, n := range d.Nodes {
		if n.ID == "" {
			return fmt.Errorf("nodes[%d]: empty id", i)
		}
		if ids[n.ID] {
			return fmt.Errorf("nodes[%d]: duplicate id %q", i, n.ID)
		}
		ids[n.ID] = true
		if !graph.ValidKinds[n.Kind] {
			return fmt.Errorf("nodes[%d] (%q): invalid kind %q", i, n.ID, n.Kind)
		}
		if n.Line < 0 {
			return fmt.Errorf("nodes[%d] (%q): negative line %d", i, n.ID, n.Line)
		}
		if n.Community < -1 {
			return fmt.Errorf("nodes[%d] (%q): community < -1: %d", i, n.ID, n.Community)
		}
	}

	for i, e := range d.Edges {
		if e.From == "" || e.To == "" {
			return fmt.Errorf("edges[%d]: empty endpoint", i)
		}
		if !graph.ValidRelations[e.Relation] {
			return fmt.Errorf("edges[%d]: invalid relation %q", i, e.Relation)
		}
		if !graph.ValidConfidences[e.Confidence] {
			return fmt.Errorf("edges[%d]: invalid confidence %q", i, e.Confidence)
		}
		if !ids[e.From] {
			return fmt.Errorf("edges[%d]: dangling 'from' node %q", i, e.From)
		}
		if !ids[e.To] {
			return fmt.Errorf("edges[%d]: dangling 'to' node %q", i, e.To)
		}
	}
	return nil
}
