package system

import (
	"sort"

	"github.com/amazopic/graffiti/internal/graph"
	"github.com/amazopic/graffiti/internal/store"
)

// Collect loads every service's published map.json into a name→Document map.
func Collect(root string, reg *Registry) (map[string]*graph.Document, error) {
	docs := make(map[string]*graph.Document, len(reg.Services))
	for _, s := range reg.Services {
		doc, err := store.Load(artifactMapPath(root, s.Name))
		if err != nil {
			return nil, err
		}
		docs[s.Name] = doc
	}
	return docs, nil
}

// Combine federates the per-service documents into one system graph: every node
// id and file path is prefixed by its service name, each service gets a synthetic
// "@service" module node, and the matcher's confident cross-service links are
// added as edges (consumer → provider). Deterministic.
func Combine(name string, docs map[string]*graph.Document, res MatchResult) *graph.Document {
	combined := graph.NewDocument(name)

	names := make([]string, 0, len(docs))
	for n := range docs {
		names = append(names, n)
	}
	sort.Strings(names)

	exists := map[string]bool{}
	for _, s := range names {
		sid := s + "::@service"
		combined.Nodes = append(combined.Nodes, graph.Node{
			ID: sid, Label: s, Kind: graph.KindModule, File: s, Line: 0,
			Community: graph.UnclusteredCommunity,
		})
		exists[sid] = true
		for _, n := range docs[s].Nodes {
			id := s + "::" + n.ID
			combined.Nodes = append(combined.Nodes, graph.Node{
				ID: id, Label: n.Label, Kind: n.Kind, File: s + "/" + n.File,
				Line: n.Line, Community: graph.UnclusteredCommunity,
			})
			exists[id] = true
		}
		for _, e := range docs[s].Edges {
			combined.Edges = append(combined.Edges, graph.Edge{
				From: s + "::" + e.From, To: s + "::" + e.To,
				Relation: e.Relation, Confidence: e.Confidence,
			})
		}
	}

	resolve := func(service, raw string) string {
		if raw != "" {
			if id := service + "::" + raw; exists[id] {
				return id
			}
		}
		return service + "::@service"
	}
	for _, l := range res.Links {
		rel := graph.RelCalls
		if l.Kind == graph.EPQueue {
			rel = graph.RelReferences
		}
		combined.Edges = append(combined.Edges, graph.Edge{
			From: resolve(l.FromService, l.FromNode), To: resolve(l.ToService, l.ToNode),
			Relation: rel, Confidence: l.Confidence,
		})
	}

	sortDoc(combined)
	return combined
}

func sortDoc(doc *graph.Document) {
	sort.Slice(doc.Nodes, func(i, j int) bool { return doc.Nodes[i].ID < doc.Nodes[j].ID })
	sort.Slice(doc.Edges, func(i, j int) bool {
		a, b := doc.Edges[i], doc.Edges[j]
		switch {
		case a.From != b.From:
			return a.From < b.From
		case a.To != b.To:
			return a.To < b.To
		case a.Relation != b.Relation:
			return a.Relation < b.Relation
		default:
			return a.Confidence < b.Confidence
		}
	})
}
