// Package build assembles per-file extractions into a single validated, directed,
// deterministically ordered graph.Document (spec §5 build stage).
package build

import (
	"sort"

	"github.com/amazopic/graffiti/internal/graph"
	"github.com/amazopic/graffiti/internal/parse"
	"github.com/amazopic/graffiti/internal/schemaval"
)

// Assemble folds extractions into one Document: stamp generatedAt, dedup nodes
// (first writer wins), dedup structural edges, run Pass-2 call resolution across
// all files, sort nodes/edges deterministically, then validate against the schema.
// Community is left at -1 (clustering is a later plan).
//
// generatedAt is threaded in here (not only at write time) so the schema's
// required generated_at field is satisfied before validation.
func Assemble(root, generatedAt string, exs []*parse.Extraction) (*graph.Document, error) {
	doc := graph.NewDocument(root)
	doc.GeneratedAt = generatedAt

	nodeIdx := map[string]bool{}
	allDefs := []graph.Node{}
	for _, ex := range exs {
		for _, n := range ex.Nodes {
			if !nodeIdx[n.ID] {
				nodeIdx[n.ID] = true
				doc.Nodes = append(doc.Nodes, n)
			}
			allDefs = append(allDefs, n)
		}
	}

	edgeIdx := map[edgeKey]bool{}
	addEdge := func(e graph.Edge) {
		k := edgeKey{e.From, e.To, e.Relation, e.Confidence}
		if edgeIdx[k] {
			return
		}
		edgeIdx[k] = true
		doc.Edges = append(doc.Edges, e)
	}

	for _, ex := range exs {
		for _, e := range ex.Edges {
			addEdge(e)
		}
	}

	var allCalls []parse.RawCall
	for _, ex := range exs {
		allCalls = append(allCalls, ex.RawCalls...)
	}
	for _, e := range parse.ResolveCalls(allDefs, allCalls) {
		addEdge(e)
	}

	sortDocument(doc)

	if err := schemaval.ValidateDocument(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

type edgeKey struct {
	from, to string
	rel      graph.Relation
	conf     graph.Confidence
}

// sortDocument imposes the canonical deterministic order (spec §8.8/§14):
// nodes by ID; edges by (from, to, relation, confidence).
func sortDocument(doc *graph.Document) {
	sort.Slice(doc.Nodes, func(i, j int) bool { return doc.Nodes[i].ID < doc.Nodes[j].ID })
	sort.Slice(doc.Edges, func(i, j int) bool {
		a, b := doc.Edges[i], doc.Edges[j]
		if a.From != b.From {
			return a.From < b.From
		}
		if a.To != b.To {
			return a.To < b.To
		}
		if a.Relation != b.Relation {
			return a.Relation < b.Relation
		}
		return a.Confidence < b.Confidence
	})
}
