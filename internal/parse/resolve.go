package parse

import (
	"sort"
	"strings"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

// ResolveCalls performs Pass 2 (spec §5): resolve raw call sites against a global
// label index over all definition/module nodes and emit `calls` edges.
//
// Resolution rules:
//   - Selector callee "pkg.Sym": match "pkg" to the file's imports by last path
//     segment; if exactly one import matches AND a module node for that import
//     path exists, emit an EXTRACTED edge to that module node (coarse, by design).
//     Zero or multiple matching imports -> drop (unresolved/ambiguous).
//   - Bare callee "Sym": look up definitions labeled "Sym".
//   - exactly one defining file -> INFERRED edge to it.
//   - defined in >= 2 distinct files -> DROP (ambiguous common name).
//   - zero definitions -> drop (external/builtin).
//
// Output edges are deduped by (from,to,relation,confidence) and sorted for
// determinism (spec §14).
func ResolveCalls(defs []graph.Node, calls []RawCall) []graph.Edge {
	labelIdx := map[string][]graph.Node{} // label -> definition nodes

	for _, n := range defs {
		switch n.Kind {
		case graph.KindFunction, graph.KindMethod, graph.KindClass:
			labelIdx[n.Label] = append(labelIdx[n.Label], n)
		}
	}

	seen := map[edgeDedupKey]bool{}
	var out []graph.Edge
	add := func(from, to string, conf graph.Confidence) {
		k := edgeDedupKey{from: from, to: to, conf: conf}
		if seen[k] {
			return
		}
		seen[k] = true
		out = append(out, graph.Edge{From: from, To: to, Relation: graph.RelCalls, Confidence: conf})
	}

	for _, rc := range calls {
		if pkg, _, isSel := splitSelector(rc.Callee); isSel {
			// Match pkg to imports by last segment; require a UNIQUE match.
			matches := matchingImports(rc.Imports, pkg)
			if len(matches) != 1 {
				continue // unresolved or ambiguous selector -> drop
			}
			// module node id for that import path == NodeID("module:"+imp, base)
			modID := graph.NodeID("module:"+matches[0], importBase(matches[0]))
			// confirm the module node actually exists (by label index fallback)
			if !moduleNodeExists(defs, modID) {
				continue
			}
			add(rc.FromID, modID, graph.ConfExtracted)
			continue
		}

		// bare call
		cands := labelIdx[rc.Callee]
		if len(cands) == 0 {
			continue
		}
		if distinctFiles(cands) >= 2 {
			continue // ambiguous common name -> drop
		}
		target := pickDeterministic(cands)
		add(rc.FromID, target.ID, graph.ConfInferred)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		if out[i].To != out[j].To {
			return out[i].To < out[j].To
		}
		if out[i].Relation != out[j].Relation {
			return out[i].Relation < out[j].Relation
		}
		return out[i].Confidence < out[j].Confidence
	})
	return out
}

type edgeDedupKey struct {
	from, to string
	conf     graph.Confidence
}

// splitSelector splits "pkg.Sym" into ("pkg","Sym",true). For a bare "Sym" returns
// isSel=false. For "a.b.C" uses the FIRST segment as pkg.
func splitSelector(callee string) (pkg, sym string, isSel bool) {
	i := strings.Index(callee, ".")
	if i < 0 {
		return "", callee, false
	}
	return callee[:i], callee[i+1:], true
}

// matchingImports returns import paths whose last segment equals pkg.
func matchingImports(imports []string, pkg string) []string {
	var out []string
	for _, imp := range imports {
		if importBase(imp) == pkg {
			out = append(out, imp)
		}
	}
	return out
}

func moduleNodeExists(defs []graph.Node, id string) bool {
	for _, n := range defs {
		if n.Kind == graph.KindModule && n.ID == id {
			return true
		}
	}
	return false
}

func distinctFiles(ns []graph.Node) int {
	files := map[string]bool{}
	for _, n := range ns {
		files[n.File] = true
	}
	return len(files)
}

// pickDeterministic chooses a stable target among candidates in a single file
// (lowest ID).
func pickDeterministic(ns []graph.Node) graph.Node {
	best := ns[0]
	for _, n := range ns[1:] {
		if n.ID < best.ID {
			best = n
		}
	}
	return best
}
