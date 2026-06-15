package cluster

import (
	"fmt"
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

// twoCliques builds two 4-node cliques joined by a single bridge edge (d->e).
// A correct community detector MUST place each clique in its own community.
func twoCliques() *graph.Document {
	doc := graph.NewDocument("repo")
	ids := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	for _, id := range ids {
		file := "left/" + id + ".go"
		if id >= "e" {
			file = "right/" + id + ".go"
		}
		doc.Nodes = append(doc.Nodes, graph.Node{
			ID: id, Label: id, Kind: graph.KindFunction, File: file, Line: 1,
			Community: graph.UnclusteredCommunity,
		})
	}
	add := func(a, b string) {
		doc.Edges = append(doc.Edges, graph.Edge{From: a, To: b, Relation: graph.RelCalls, Confidence: graph.ConfInferred})
	}
	for _, cl := range [][]string{{"a", "b", "c", "d"}, {"e", "f", "g", "h"}} {
		for i := 0; i < len(cl); i++ {
			for j := i + 1; j < len(cl); j++ {
				add(cl[i], cl[j])
			}
		}
	}
	add("d", "e") // bridge
	return doc
}

func commByID(doc *graph.Document) map[string]int {
	m := map[string]int{}
	for _, n := range doc.Nodes {
		m[n.ID] = n.Community
	}
	return m
}

func TestCluster_TwoCliquesSeparate(t *testing.T) {
	doc := twoCliques()
	k := Cluster(doc)

	comm := commByID(doc)
	for id, c := range comm {
		if c < 0 {
			t.Fatalf("node %s has negative community %d (every node must be >= 0)", id, c)
		}
	}
	for _, x := range []string{"b", "c", "d"} {
		if comm[x] != comm["a"] {
			t.Fatalf("left clique split: %s in %d, a in %d", x, comm[x], comm["a"])
		}
	}
	for _, x := range []string{"f", "g", "h"} {
		if comm[x] != comm["e"] {
			t.Fatalf("right clique split: %s in %d, e in %d", x, comm[x], comm["e"])
		}
	}
	if comm["a"] == comm["e"] {
		t.Fatalf("cliques not separated: both in community %d", comm["a"])
	}
	if k != 2 {
		t.Fatalf("K = %d, want 2", k)
	}
	if comm["a"] != 0 {
		t.Fatalf("smallest-id node 'a' must be community 0, got %d", comm["a"])
	}
}

func TestCluster_ContiguousLabels(t *testing.T) {
	doc := twoCliques()
	k := Cluster(doc)
	seen := map[int]bool{}
	for _, n := range doc.Nodes {
		seen[n.Community] = true
	}
	for i := 0; i < k; i++ {
		if !seen[i] {
			t.Fatalf("community %d missing — labels not contiguous 0..%d", i, k-1)
		}
	}
	if len(seen) != k {
		t.Fatalf("distinct communities = %d, want K = %d", len(seen), k)
	}
}

func TestCluster_ModularityBeatsSingleton(t *testing.T) {
	doc := twoCliques()
	for i := range doc.Nodes {
		doc.Nodes[i].Community = i
	}
	qSingleton := modularity(doc)
	for i := range doc.Nodes {
		doc.Nodes[i].Community = graph.UnclusteredCommunity
	}
	Cluster(doc)
	qClustered := modularity(doc)
	if qClustered < qSingleton {
		t.Fatalf("clustered modularity %.4f < singleton %.4f", qClustered, qSingleton)
	}
}

func TestCluster_Deterministic(t *testing.T) {
	snap := func() string {
		doc := twoCliques()
		Cluster(doc)
		out := ""
		for _, n := range doc.Nodes {
			out += fmt.Sprintf("%s=%d;", n.ID, n.Community)
		}
		return out
	}
	if a, b := snap(), snap(); a != b {
		t.Fatalf("non-deterministic:\nA=%s\nB=%s", a, b)
	}
}

func TestCluster_EmptyAndIsolated(t *testing.T) {
	if k := Cluster(graph.NewDocument("repo")); k != 0 {
		t.Fatalf("empty K = %d, want 0", k)
	}
	iso := graph.NewDocument("repo")
	for _, id := range []string{"x", "y", "z"} {
		iso.Nodes = append(iso.Nodes, graph.Node{ID: id, Label: id, Kind: graph.KindFunction, File: id + ".go", Line: 1, Community: graph.UnclusteredCommunity})
	}
	k := Cluster(iso)
	if k != 3 {
		t.Fatalf("isolated K = %d, want 3 (each its own community)", k)
	}
	for _, n := range iso.Nodes {
		if n.Community < 0 {
			t.Fatalf("isolated node %s community %d < 0", n.ID, n.Community)
		}
	}
}

// modularity scores the current Node.Community assignment over the undirected,
// edge-count-weighted projection. Test-only oracle (independent of Cluster's
// internals) so it genuinely cross-checks the algorithm.
func modularity(doc *graph.Document) float64 {
	idToComm := map[string]int{}
	for _, n := range doc.Nodes {
		idToComm[n.ID] = n.Community
	}
	pw := map[[2]string]float64{}
	deg := map[string]float64{}
	var m2 float64
	for _, e := range doc.Edges {
		if e.From == e.To {
			continue
		}
		a, b := e.From, e.To
		if a > b {
			a, b = b, a
		}
		pw[[2]string{a, b}]++
		deg[e.From]++
		deg[e.To]++
		m2 += 2
	}
	if m2 == 0 {
		return 0
	}
	ids := make([]string, 0, len(deg))
	for id := range deg {
		ids = append(ids, id)
	}
	var q float64
	for _, i := range ids {
		for _, j := range ids {
			if idToComm[i] != idToComm[j] {
				continue
			}
			a, b := i, j
			if a > b {
				a, b = b, a
			}
			var aij float64
			if i != j {
				aij = pw[[2]string{a, b}]
			}
			q += aij - deg[i]*deg[j]/m2
		}
	}
	return q / m2
}
