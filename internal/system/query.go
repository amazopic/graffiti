package system

import (
	"sort"
	"strings"

	"github.com/amazopic/graffiti/internal/graph"
)

// Federate is the read path shared by `system build/render/impact/audit`: load
// the registry, collect every published map, and discover cross-service links.
func Federate(root string) (*Registry, map[string]*graph.Document, MatchResult, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, nil, MatchResult{}, err
	}
	docs, err := Collect(root, reg)
	if err != nil {
		return nil, nil, MatchResult{}, err
	}
	return reg, docs, Match(docs), nil
}

// ImpactReport answers "if this changes, who breaks?".
type ImpactReport struct {
	Target   string
	Direct   []SystemLink // links pointing INTO the target (its immediate consumers)
	Affected []string     // all transitively-dependent services (sorted, excludes target)
}

// Impact reverse-traverses the discovered links from a target service (or a
// specific "service::KEY" endpoint) to every service that depends on it,
// directly or transitively.
func Impact(res MatchResult, target string) ImpactReport {
	service, key := target, ""
	if i := strings.Index(target, "::"); i >= 0 {
		service, key = target[:i], target[i+2:]
	}

	rep := ImpactReport{Target: target}
	for _, l := range res.Links {
		if l.ToService == service && (key == "" || l.Key == key) {
			rep.Direct = append(rep.Direct, l)
		}
	}

	// Reverse adjacency: provider → consumers.
	rev := map[string][]string{}
	for _, l := range res.Links {
		rev[l.ToService] = append(rev[l.ToService], l.FromService)
	}
	visited := map[string]bool{service: true}
	queue := []string{service}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, c := range rev[p] {
			if !visited[c] {
				visited[c] = true
				rep.Affected = append(rep.Affected, c)
				queue = append(queue, c)
			}
		}
	}
	sort.Strings(rep.Affected)
	return rep
}
