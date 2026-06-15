// Package cluster assigns each node a community via a deterministic, single-level
// Louvain over an undirected, edge-count-weighted projection of the directed graph
// (spec §5 cluster stage), and names the resulting communities (spec §8.3).
//
// Determinism (spec §14): nodes are processed in ascending node-id order;
// modularity-gain ties break by smallest community id (then implicitly smallest
// member id, since communities are seeded as singletons in id order); no
// math/rand; communities are renumbered by smallest member id so labels are a
// contiguous 0..K-1. No Go-map iteration feeds any decision or any emitted order.
package cluster

import (
	"sort"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

// Cluster assigns doc.Nodes[i].Community in place to a contiguous 0..K-1 label and
// returns K. Every node ends with Community >= 0 (edge-less nodes become their own
// singleton community). doc is otherwise unmodified.
func Cluster(doc *graph.Document) int {
	n := len(doc.Nodes)
	if n == 0 {
		return 0
	}

	// Rank nodes by ascending id; ranks are the canonical processing order.
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(a, b int) bool {
		return doc.Nodes[order[a]].ID < doc.Nodes[order[b]].ID
	})
	idToRank := make(map[string]int, n)
	for rank, ni := range order {
		idToRank[doc.Nodes[ni].ID] = rank
	}

	// Undirected, edge-count-weighted projection. A directed edge u->v (any
	// relation) adds weight 1 to the unordered pair {u,v}; parallel edges
	// accumulate; self-loops are ignored.
	type pair struct{ a, b int } // ranks, a < b
	pw := map[pair]float64{}
	for _, e := range doc.Edges {
		ui, ok1 := idToRank[e.From]
		vi, ok2 := idToRank[e.To]
		if !ok1 || !ok2 || ui == vi {
			continue
		}
		if ui > vi {
			ui, vi = vi, ui
		}
		pw[pair{ui, vi}]++
	}

	type wedge struct {
		to int
		w  float64
	}
	adj := make([][]wedge, n)
	deg := make([]float64, n)
	var m2 float64

	pairs := make([]pair, 0, len(pw))
	for p := range pw {
		pairs = append(pairs, p)
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].a != pairs[j].a {
			return pairs[i].a < pairs[j].a
		}
		return pairs[i].b < pairs[j].b
	})
	for _, p := range pairs {
		w := pw[p]
		adj[p.a] = append(adj[p.a], wedge{p.b, w})
		adj[p.b] = append(adj[p.b], wedge{p.a, w})
		deg[p.a] += w
		deg[p.b] += w
		m2 += 2 * w
	}

	// Seed singletons (community == rank).
	comm := make([]int, n)
	commTot := make([]float64, n) // total degree of community c
	for i := 0; i < n; i++ {
		comm[i] = i
		commTot[i] = deg[i]
	}

	if m2 == 0 {
		return finalize(doc, order, comm)
	}

	// Local-moving pass, repeated until no node changes community.
	improved := true
	for improved {
		improved = false
		for i := 0; i < n; i++ { // ascending rank == ascending id
			wTo := map[int]float64{}
			for _, e := range adj[i] {
				wTo[comm[e.to]] += e.w
			}

			cur := comm[i]
			commTot[cur] -= deg[i] // pull i out before scoring

			cands := []int{cur}
			seen := map[int]bool{cur: true}
			for c := range wTo {
				if !seen[c] {
					cands = append(cands, c)
					seen[c] = true
				}
			}
			sort.Ints(cands) // ascending: first maximizer wins ties (smallest comm id)

			bestC := cur
			bestGain := wTo[cur] - deg[i]*commTot[cur]/m2
			for _, c := range cands {
				if c == cur {
					continue
				}
				if gain := wTo[c] - deg[i]*commTot[c]/m2; gain > bestGain {
					bestGain, bestC = gain, c
				}
			}

			commTot[bestC] += deg[i]
			if bestC != cur {
				comm[i] = bestC
				improved = true
			}
		}
	}

	return finalize(doc, order, comm)
}

// finalize renumbers raw communities by smallest member rank (== smallest id) so
// labels are contiguous 0..K-1, writes Node.Community, and returns K.
func finalize(doc *graph.Document, order, comm []int) int {
	minRank := map[int]int{}
	for rank, c := range comm {
		if r, ok := minRank[c]; !ok || rank < r {
			minRank[c] = rank
		}
	}
	type cr struct{ raw, rank int }
	crs := make([]cr, 0, len(minRank))
	for raw, r := range minRank {
		crs = append(crs, cr{raw, r})
	}
	sort.Slice(crs, func(i, j int) bool { return crs[i].rank < crs[j].rank })
	remap := make(map[int]int, len(crs))
	for newID, c := range crs {
		remap[c.raw] = newID
	}
	for rank, c := range comm {
		doc.Nodes[order[rank]].Community = remap[c]
	}
	return len(crs)
}
