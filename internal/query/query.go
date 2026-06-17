// Package query is graffiti's LLM-free retrieval path (spec §7): tokenize the
// question, score nodes by IDF over their label+kind term-bags, seed by score
// (tie-break id asc), BFS-expand neighbors under a SOFT TOKEN BUDGET, and
// serialize the scoped subgraph to compact deterministic text. No inference call
// is made; the host model reasons over the returned text. Pure, sort-only, no I/O.
package query

import (
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/amazopic/graffiti/internal/graph"
	"github.com/amazopic/graffiti/internal/store"

	"golang.org/x/text/unicode/norm"
)

// DefaultTokenBudget mirrors spec §7 ("default ~2000"). The budget is a SOFT
// estimate over NODE text only (see estimateTokens / the budget note below).
const DefaultTokenBudget = 2000

// estimateTokens is a deterministic, zero-dependency token estimator (~4 chars
// per token). NOT a real tokenizer — chosen for determinism and zero deps per
// the project ethos (spec §3/§10). The budget bounds NODE selection only; edge
// text is emitted for free for every selected pair (see Query / Open Issue #1).
func estimateTokens(s string) int {
	n := len(s) / 4
	if n == 0 && len(s) > 0 {
		n = 1
	}
	return n
}

// tokenize NFC-normalizes (aligning with graph.NormalizeID so query terms fold
// the same way node-id slugs do), lowercases, splits on any non-word boundary,
// and ALSO splits camelCase/snake_case so "parseFile" matches "parse"/"file".
func tokenize(s string) []string {
	s = norm.NFC.String(s)
	var out []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, strings.ToLower(cur.String()))
			cur.Reset()
		}
	}
	var prev rune
	for _, r := range s {
		switch {
		case isLower(r), isDigit(r):
			cur.WriteRune(r)
		case isUpper(r):
			if isLower(prev) || isDigit(prev) { // camelCase boundary
				flush()
			}
			cur.WriteRune(r)
		case isWordRune(r): // non-ASCII letter/digit (mirrors graph.isWordRune)
			cur.WriteRune(r)
		default:
			flush()
		}
		prev = r
	}
	flush()
	return out
}

func isLower(r rune) bool { return r >= 'a' && r <= 'z' }
func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }
func isDigit(r rune) bool { return r >= '0' && r <= '9' }

// isWordRune mirrors graph.isWordRune: ASCII letters/digits plus any non-ASCII
// Unicode letter or digit (so accented/CJK identifiers tokenize like the slug).
func isWordRune(r rune) bool {
	return isLower(r) || isUpper(r) || isDigit(r) || (r > 0x7F && (unicode.IsLetter(r) || unicode.IsDigit(r)))
}

// nodeTerms returns a node's bag of terms: its label tokens plus its kind.
func nodeTerms(n graph.Node) []string {
	return append(tokenize(n.Label), string(n.Kind))
}

type scoredNode struct {
	id    string
	score float64
}

// scoreNodes computes IDF-weighted overlap of every node against the query terms
// and returns the scored ids sorted by (score desc, id asc).
func scoreNodes(idx *store.Index, question string) []scoredNode {
	qTerms := dedupe(tokenize(question))
	if len(qTerms) == 0 {
		return nil
	}
	ids := idx.IDs()
	df := make(map[string]int)
	bags := make(map[string]map[string]bool, len(ids))
	for _, id := range ids {
		n, _ := idx.Node(id)
		bag := make(map[string]bool)
		for _, t := range nodeTerms(n) {
			bag[t] = true
		}
		bags[id] = bag
		for t := range bag {
			df[t]++
		}
	}
	N := float64(idx.Len())
	idf := func(term string) float64 { return math.Log((N + 1) / (float64(df[term]) + 1)) }

	scored := make([]scoredNode, 0, len(ids))
	for _, id := range ids { // ids are already sorted; iteration is deterministic
		bag := bags[id]
		var s float64
		for _, qt := range qTerms {
			if bag[qt] {
				s += idf(qt)
			}
		}
		if s > 0 {
			scored = append(scored, scoredNode{id: id, score: s})
		}
	}
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		return scored[i].id < scored[j].id // tie-break id asc
	})
	return scored
}

// Query runs the full LLM-free retrieval and returns the serialized subgraph.
func Query(idx *store.Index, question string, budget int) string {
	if budget <= 0 {
		budget = DefaultTokenBudget
	}
	scored := scoreNodes(idx, question)

	selected := make(map[string]bool)
	var orderedSel []string
	used := 0
	add := func(id string) bool {
		if selected[id] {
			return true
		}
		n, ok := idx.Node(id)
		if !ok {
			return true
		}
		cost := estimateTokens(formatNode(n)) // SOFT NODE BUDGET — edges are free
		if used+cost > budget {
			return false
		}
		selected[id] = true
		orderedSel = append(orderedSel, id)
		used += cost
		return true
	}

	for _, sn := range scored { // seeds, highest score first
		if !add(sn.id) {
			break
		}
	}
	for i := 0; i < len(orderedSel); i++ { // BFS expansion
		if used >= budget {
			break
		}
		for _, e := range neighbors(idx, orderedSel[i]) {
			other := e.To
			if other == orderedSel[i] {
				other = e.From
			}
			_ = add(other) // budget-exhausted neighbors are skipped, keep trying smaller ones
		}
	}
	return serialize(idx, selected)
}

// neighbors returns out-edges then in-edges of id, each group already sorted by
// (relation, other-id, confidence) by store.Index — no re-sort, no map iteration.
func neighbors(idx *store.Index, id string) []graph.Edge {
	out := idx.Out(id)
	in := idx.In(id)
	es := make([]graph.Edge, 0, len(out)+len(in))
	es = append(es, out...)
	es = append(es, in...)
	return es
}

// serialize renders the selected subgraph: a NODES block (sorted by id) and an
// EDGES block (edges with BOTH endpoints selected, de-duped, sorted).
func serialize(idx *store.Index, selected map[string]bool) string {
	ids := make([]string, 0, len(selected))
	for id := range selected {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var b strings.Builder
	b.WriteString("NODES\n")
	for _, id := range ids {
		n, ok := idx.Node(id)
		if !ok {
			continue // skip unknown ids: no garbage line emitted
		}
		b.WriteString(formatNode(n))
		b.WriteByte('\n')
	}

	type ek struct {
		from, to string
		rel      graph.Relation
		conf     graph.Confidence
	}
	seen := make(map[ek]bool)
	var edges []graph.Edge
	for _, id := range ids { // ids sorted => deterministic collection
		for _, e := range idx.Out(id) {
			if !selected[e.To] {
				continue
			}
			k := ek{e.From, e.To, e.Relation, e.Confidence}
			if seen[k] {
				continue
			}
			seen[k] = true
			edges = append(edges, e)
		}
	}
	sort.SliceStable(edges, func(i, j int) bool {
		a, c := edges[i], edges[j]
		if a.From != c.From {
			return a.From < c.From
		}
		if a.To != c.To {
			return a.To < c.To
		}
		if a.Relation != c.Relation {
			return a.Relation < c.Relation
		}
		return a.Confidence < c.Confidence
	})
	b.WriteString("EDGES\n")
	for _, e := range edges {
		b.WriteString(formatEdge(e))
		b.WriteByte('\n')
	}
	return b.String()
}

func formatNode(n graph.Node) string {
	return n.ID + " [" + string(n.Kind) + "] " + n.Label + " @ " + n.File + ":" + itoa(n.Line)
}

func formatEdge(e graph.Edge) string {
	return e.From + " -" + string(e.Relation) + "-> " + e.To + " (" + string(e.Confidence) + ")"
}

// FormatNode renders a single node in the canonical compact format. Exported so
// internal/mcp's get_node/get_neighbors/shortest_path render in the same form
// without duplicating the formatter (single-sourced, deterministic).
func FormatNode(n graph.Node) string { return formatNode(n) }

// FormatEdge renders a single edge in the canonical compact format. Exported for
// internal/mcp (see FormatNode).
func FormatEdge(e graph.Edge) string { return formatEdge(e) }

// Serialize renders the subgraph induced by the given ids in the canonical
// NODES/EDGES format. Exported for internal/mcp to reuse the validated
// serializer over an explicit id set.
func Serialize(idx *store.Index, ids []string) string {
	selected := make(map[string]bool, len(ids))
	for _, id := range ids {
		selected[id] = true
	}
	return serialize(idx, selected)
}

func dedupe(ss []string) []string {
	seen := make(map[string]bool, len(ss))
	out := ss[:0]
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
