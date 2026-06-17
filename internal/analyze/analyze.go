// Package analyze derives read-only insights from a clustered graph (spec §5
// analyze stage / §8.5): god nodes (high-degree hubs), surprising connections
// (cross-community edges), import cycles, and exactly 3 suggested questions.
// It mutates nothing and performs no I/O. All outputs are deterministic: sorted
// iteration with explicit total-order tie-breaks; no math/rand; no LLM.
package analyze

import (
	"fmt"
	"sort"

	"github.com/amazopic/graffiti/internal/graph"
)

// GodNodeCap is the documented limit on map-surfaced god nodes (spec §8.5: ~7).
const GodNodeCap = 7

// Analysis is the bundle of derived insights.
type Analysis struct {
	GodNodes   []GodNode
	Surprising []SurprisingEdge
	Cycles     [][]string // each cycle is a canonical rotation of node ids
	Questions  []string   // exactly 3
}

// GodNode is a high-degree hub.
type GodNode struct {
	ID     string
	Label  string
	Degree int
}

// SurprisingEdge is a directed edge whose endpoints are in different communities.
type SurprisingEdge struct {
	From, To         string
	FromComm, ToComm int
	Relation         graph.Relation
	Confidence       graph.Confidence
}

// Degrees returns total (in + out) degree per node id over all edges.
func Degrees(doc *graph.Document) map[string]int {
	deg := make(map[string]int, len(doc.Nodes))
	for _, e := range doc.Edges {
		deg[e.From]++
		deg[e.To]++
	}
	return deg
}

// Analyze computes the full Analysis for a clustered document. deg should come
// from Degrees(doc).
func Analyze(doc *graph.Document, deg map[string]int) Analysis {
	byID := make(map[string]graph.Node, len(doc.Nodes))
	for _, n := range doc.Nodes {
		byID[n.ID] = n
	}
	gods := godNodes(doc, byID, deg)
	surp := surprising(doc, byID)
	cycles := importCycles(doc)
	qs := questions(doc, gods, surp, byID)
	return Analysis{GodNodes: gods, Surprising: surp, Cycles: cycles, Questions: qs}
}

// godNodes returns up to GodNodeCap highest-degree nodes (degree >= 2), ordered
// by degree desc then id asc.
func godNodes(doc *graph.Document, byID map[string]graph.Node, deg map[string]int) []GodNode {
	type dn struct {
		id  string
		deg int
	}
	cand := make([]dn, 0, len(doc.Nodes))
	for _, n := range doc.Nodes {
		if d := deg[n.ID]; d >= 2 {
			cand = append(cand, dn{n.ID, d})
		}
	}
	sort.Slice(cand, func(i, j int) bool {
		if cand[i].deg != cand[j].deg {
			return cand[i].deg > cand[j].deg
		}
		return cand[i].id < cand[j].id
	})
	out := make([]GodNode, 0, GodNodeCap)
	for _, x := range cand {
		if len(out) >= GodNodeCap {
			break
		}
		out = append(out, GodNode{ID: x.id, Label: byID[x.id].Label, Degree: x.deg})
	}
	return out
}

// surprising returns cross-community edges in (from,to,relation,confidence) order.
func surprising(doc *graph.Document, byID map[string]graph.Node) []SurprisingEdge {
	var out []SurprisingEdge
	for _, e := range doc.Edges {
		fn, ok1 := byID[e.From]
		tn, ok2 := byID[e.To]
		if !ok1 || !ok2 {
			continue
		}
		if fn.Community >= 0 && tn.Community >= 0 && fn.Community != tn.Community {
			out = append(out, SurprisingEdge{
				From: e.From, To: e.To,
				FromComm: fn.Community, ToComm: tn.Community,
				Relation: e.Relation, Confidence: e.Confidence,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
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
	return out
}

// importCycles finds cycles among `imports` edges via a deterministic DFS.
// Each cycle is returned as a canonical rotation (rotated to start at its
// smallest id); the set is de-duplicated and sorted.
func importCycles(doc *graph.Document) [][]string {
	adj := map[string][]string{}
	nodeset := map[string]bool{}
	for _, e := range doc.Edges {
		if e.Relation != graph.RelImports {
			continue
		}
		adj[e.From] = append(adj[e.From], e.To)
		nodeset[e.From] = true
		nodeset[e.To] = true
	}
	for k := range adj {
		sort.Strings(adj[k])
	}
	nodes := make([]string, 0, len(nodeset))
	for n := range nodeset {
		nodes = append(nodes, n)
	}
	sort.Strings(nodes)

	var found [][]string
	seen := map[string]bool{}
	color := map[string]int{} // 0 white, 1 gray, 2 black
	var stack []string

	var dfs func(u string)
	dfs = func(u string) {
		color[u] = 1
		stack = append(stack, u)
		for _, v := range adj[u] {
			switch color[v] {
			case 0:
				dfs(v)
			case 1:
				idx := 0
				for i, s := range stack {
					if s == v {
						idx = i
						break
					}
				}
				canon := canonicalCycle(append([]string(nil), stack[idx:]...))
				key := fmt.Sprint(canon)
				if !seen[key] {
					seen[key] = true
					found = append(found, canon)
				}
			}
		}
		stack = stack[:len(stack)-1]
		color[u] = 2
	}
	for _, n := range nodes {
		if color[n] == 0 {
			dfs(n)
		}
	}
	sort.Slice(found, func(i, j int) bool { return fmt.Sprint(found[i]) < fmt.Sprint(found[j]) })
	return found
}

func canonicalCycle(c []string) []string {
	if len(c) == 0 {
		return c
	}
	min, at := c[0], 0
	for i, s := range c {
		if s < min {
			min, at = s, i
		}
	}
	out := make([]string, 0, len(c))
	for i := 0; i < len(c); i++ {
		out = append(out, c[(at+i)%len(c)])
	}
	return out
}

// questions builds exactly 3 deterministic template questions (spec §11), padding
// from a fixed fallback list when the graph is too small to fill all three.
func questions(doc *graph.Document, gods []GodNode, surp []SurprisingEdge, byID map[string]graph.Node) []string {
	var qs []string
	if len(gods) > 0 {
		qs = append(qs, fmt.Sprintf("What does %q do, and why is it touched by %d things?", gods[0].Label, gods[0].Degree))
	}
	if len(surp) > 0 {
		qs = append(qs, fmt.Sprintf("Why does %q connect to %q across subsystems?",
			byID[surp[0].From].Label, byID[surp[0].To].Label))
	}
	if len(doc.Communities) > 0 {
		biggest := doc.Communities[0]
		for _, c := range doc.Communities {
			if len(c.Members) > len(biggest.Members) {
				biggest = c
			}
		}
		qs = append(qs, fmt.Sprintf("How does the %q subsystem (%d things) fit together?", biggest.Label, len(biggest.Members)))
	}
	fallback := []string{
		"What are the most central pieces of this codebase?",
		"Which parts of the code are most interconnected?",
		"Where would a change ripple the furthest?",
	}
	for fi := 0; len(qs) < 3; fi++ {
		qs = append(qs, fallback[fi%len(fallback)])
	}
	return qs[:3]
}
