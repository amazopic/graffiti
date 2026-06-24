package system

import (
	"sort"
	"strings"

	"github.com/amazopic/graffiti/internal/graph"
)

// SystemLink is one discovered cross-service edge: a consumer in one service
// reaching a provider in another. Direction is consumer → provider (for queues,
// publisher → subscriber).
type SystemLink struct {
	FromService string
	ToService   string
	Kind        graph.EndpointKind
	Key         string
	FromNode    string // raw consumer node id ("" if unknown)
	ToNode      string // raw provider node id ("" if unknown)
	Confidence  graph.Confidence
	Via         string // provenance, e.g. "literal↔openapi" or "literal↔route+path-only"
}

// Dangling is a consume that matched no provider anywhere — a likely dead or
// renamed endpoint (audit signal).
type Dangling struct {
	Service string
	Kind    graph.EndpointKind
	Key     string
	Display string
}

// Orphan is a provide that no other service consumes.
type Orphan struct {
	Service string
	Kind    graph.EndpointKind
	Key     string
	Display string
}

// MatchResult is the discovered system overlay before serialization.
type MatchResult struct {
	Links     []SystemLink
	Ambiguous []SystemLink
	Dangling  []Dangling
	Orphans   []Orphan
}

type provRef struct {
	service string
	ep      graph.Endpoint
}

// keyPath returns the path portion of an http key "METHOD /path".
func keyPath(key string) string {
	if i := strings.IndexByte(key, ' '); i >= 0 {
		return key[i+1:]
	}
	return key
}

func combineConf(a, b graph.Confidence) graph.Confidence {
	if a == graph.ConfExtracted && b == graph.ConfExtracted {
		return graph.ConfExtracted
	}
	return graph.ConfInferred
}

// Match discovers cross-service links by matching every service's `consumes`
// against every other service's `provides`. docs is keyed by service name.
func Match(docs map[string]*graph.Document) MatchResult {
	// Stable service order.
	names := make([]string, 0, len(docs))
	for n := range docs {
		names = append(names, n)
	}
	sort.Strings(names)

	exact := map[string][]provRef{}    // (kind|key) → providers
	pathIdx := map[string][]provRef{}   // normalized http path → providers
	matched := map[string]bool{}        // service|kind|key marked consumed
	var allProv []provRef

	for _, s := range names {
		for _, ep := range docs[s].Provides {
			ref := provRef{service: s, ep: ep}
			allProv = append(allProv, ref)
			exact[string(ep.Kind)+"|"+ep.Key] = append(exact[string(ep.Kind)+"|"+ep.Key], ref)
			if ep.Kind == graph.EPHTTP {
				pathIdx[keyPath(ep.Key)] = append(pathIdx[keyPath(ep.Key)], ref)
			}
		}
	}

	var res MatchResult
	for _, s := range names {
		for _, con := range docs[s].Consumes {
			cands := others(exact[string(con.Kind)+"|"+con.Key], s)
			fallback := false
			if len(cands) == 0 && con.Kind == graph.EPHTTP {
				cands = others(pathIdx[keyPath(con.Key)], s)
				fallback = len(cands) > 0
			}
			switch {
			case len(cands) == 0:
				res.Dangling = append(res.Dangling, Dangling{s, con.Kind, con.Key, con.Display})
			case len(cands) == 1:
				p := cands[0]
				conf := combineConf(con.Confidence, p.ep.Confidence)
				via := con.Source + "↔" + p.ep.Source
				if fallback {
					conf = graph.ConfInferred
					via += "+path-only"
				}
				res.Links = append(res.Links, SystemLink{
					FromService: s, ToService: p.service, Kind: con.Kind, Key: p.ep.Key,
					FromNode: con.Node, ToNode: p.ep.Node, Confidence: conf, Via: via,
				})
				matched[p.service+"|"+string(p.ep.Kind)+"|"+p.ep.Key] = true
			default: // ambiguous: >1 provider
				for _, p := range cands {
					res.Ambiguous = append(res.Ambiguous, SystemLink{
						FromService: s, ToService: p.service, Kind: con.Kind, Key: p.ep.Key,
						FromNode: con.Node, ToNode: p.ep.Node, Confidence: graph.ConfAmbiguous,
						Via: con.Source + "↔" + p.ep.Source,
					})
					matched[p.service+"|"+string(p.ep.Kind)+"|"+p.ep.Key] = true
				}
			}
		}
	}

	// Orphans: provides never consumed (dedup by service|kind|key).
	seen := map[string]bool{}
	for _, p := range allProv {
		k := p.service + "|" + string(p.ep.Kind) + "|" + p.ep.Key
		if matched[k] || seen[k] {
			continue
		}
		seen[k] = true
		res.Orphans = append(res.Orphans, Orphan{p.service, p.ep.Kind, p.ep.Key, p.ep.Display})
	}

	sortResult(&res)
	return res
}

// others returns providers not owned by service s.
func others(refs []provRef, s string) []provRef {
	var out []provRef
	for _, r := range refs {
		if r.service != s {
			out = append(out, r)
		}
	}
	return out
}

func sortResult(res *MatchResult) {
	linkLess := func(a, b SystemLink) bool {
		switch {
		case a.FromService != b.FromService:
			return a.FromService < b.FromService
		case a.ToService != b.ToService:
			return a.ToService < b.ToService
		case a.Kind != b.Kind:
			return a.Kind < b.Kind
		default:
			return a.Key < b.Key
		}
	}
	sort.SliceStable(res.Links, func(i, j int) bool { return linkLess(res.Links[i], res.Links[j]) })
	sort.SliceStable(res.Ambiguous, func(i, j int) bool { return linkLess(res.Ambiguous[i], res.Ambiguous[j]) })
	sort.SliceStable(res.Dangling, func(i, j int) bool {
		a, b := res.Dangling[i], res.Dangling[j]
		if a.Service != b.Service {
			return a.Service < b.Service
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Key < b.Key
	})
	sort.SliceStable(res.Orphans, func(i, j int) bool {
		a, b := res.Orphans[i], res.Orphans[j]
		if a.Service != b.Service {
			return a.Service < b.Service
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Key < b.Key
	})
}
