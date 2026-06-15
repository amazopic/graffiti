package cluster

import (
	"path"
	"sort"
	"strings"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

// NameCommunities builds the doc.Communities slice from the clustered nodes
// (spec §8.3): each community's label is the human form of its members' dominant
// source directory (strict majority), falling back to the most-central member's
// label when no directory dominates or the dominant directory is generic/root.
// Members are sorted; communities are returned sorted by id ascending.
//
// deg maps node id -> total (in+out) degree (see analyze.Degrees); it is used
// only for the most-central tie-break, so callers may pass an empty map if they
// do not need the fallback to be centrality-aware (it then falls back to the
// smallest-id member, which is still deterministic).
func NameCommunities(doc *graph.Document, deg map[string]int) []graph.Community {
	members := map[int][]string{}
	maxComm := -1
	for _, nd := range doc.Nodes {
		if nd.Community < 0 {
			continue
		}
		members[nd.Community] = append(members[nd.Community], nd.ID)
		if nd.Community > maxComm {
			maxComm = nd.Community
		}
	}

	byID := make(map[string]graph.Node, len(doc.Nodes))
	for _, nd := range doc.Nodes {
		byID[nd.ID] = nd
	}

	out := make([]graph.Community, 0, maxComm+1)
	for c := 0; c <= maxComm; c++ {
		mem := members[c]
		if len(mem) == 0 {
			continue
		}
		sort.Strings(mem)
		out = append(out, graph.Community{
			ID:      c,
			Label:   labelFor(mem, byID, deg),
			Members: mem,
		})
	}
	return out
}

// labelFor implements the §8.3 hybrid heuristic. members must be pre-sorted.
func labelFor(members []string, byID map[string]graph.Node, deg map[string]int) string {
	// Count members per (non-generic) directory.
	dirCount := map[string]int{}
	for _, id := range members {
		if d := dirOf(byID[id].File); d != "" {
			dirCount[d]++
		}
	}
	// Pick the dominant directory deterministically: highest count, ties by the
	// lexicographically smallest directory.
	dirs := make([]string, 0, len(dirCount))
	for d := range dirCount {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)
	bestDir, bestN := "", 0
	for _, d := range dirs {
		if dirCount[d] > bestN {
			bestN, bestDir = dirCount[d], d
		}
	}
	// "Dominant" == strict majority of members share it.
	if bestDir != "" && bestN*2 > len(members) {
		return dirLabel(bestDir)
	}

	// Fallback: most-central member (highest degree; ties by smallest id, which
	// is members[0] since members is sorted ascending).
	best, bestDeg := members[0], deg[members[0]]
	for _, id := range members[1:] {
		if deg[id] > bestDeg {
			bestDeg, best = deg[id], id
		}
	}
	return byID[best].Label
}

// dirOf returns the directory of a repo-relative file path ("" for root files).
func dirOf(file string) string {
	d := path.Dir(file)
	if d == "." || d == "/" || d == "" {
		return ""
	}
	return d
}

// dirLabel turns a directory like "internal/auth" into a human label "Auth",
// title-casing the base segment (split on '-'/'_').
func dirLabel(dir string) string {
	base := path.Base(dir)
	if base == "" || base == "." || base == "/" {
		return dir
	}
	parts := strings.FieldsFunc(base, func(r rune) bool { return r == '-' || r == '_' })
	for i, p := range parts {
		if p != "" {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	if len(parts) == 0 {
		return base
	}
	return strings.Join(parts, " ")
}
