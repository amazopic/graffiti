package render

import (
	"regexp"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

// graphData is the compact columnar graph island consumed by viewer/app.js.
// Parallel arrays index the same node; edges are [fromIdx,toIdx] pairs into them.
// Sector color and node layout are derived in the browser (from file paths and a
// force simulation), so they are intentionally absent here — keeping the island a
// pure, deterministic function of the Document.
type graphData struct {
	Label []string `json:"label"`
	Kind  []string `json:"kind"`
	File  []string `json:"file"`
	Line  []int    `json:"line"`
	Deg   []int    `json:"deg"`
	Cat   []int    `json:"cat"` // 0=client 1=test 2=external
	Edges [][2]int `json:"edges"`
}

// testFile matches common test-file conventions across graffiti's languages
// (Go _test.go, JS/TS .test./.spec., Python test_/tests/, *Test.{java,kt,php,rs,py}).
var testFile = regexp.MustCompile(`(^|/)(tests?)(/|_)|_test\.|\.test\.|\.spec\.|(^|/)test_|Tests?\.(java|kt|php|rs|py)$`)

// categoryOf classifies a node: 2=external (imported module), 1=test, 0=client.
func categoryOf(n graph.Node) int {
	if n.Kind == graph.KindModule {
		return 2
	}
	if testFile.MatchString(n.File) || (len(n.Label) >= 4 && n.Label[:4] == "Test") {
		return 1
	}
	return 0
}

// graphIsland builds the columnar island from a (sorted) Document. doc.Nodes is
// already id-sorted by build.Assemble, so node order — and thus edge indices —
// are deterministic.
func graphIsland(doc *graph.Document) graphData {
	idx := make(map[string]int, len(doc.Nodes))
	d := graphData{Edges: [][2]int{}}
	for i, n := range doc.Nodes {
		idx[n.ID] = i
		d.Label = append(d.Label, n.Label)
		d.Kind = append(d.Kind, string(n.Kind))
		d.File = append(d.File, n.File)
		d.Line = append(d.Line, n.Line)
		d.Cat = append(d.Cat, categoryOf(n))
	}
	d.Deg = make([]int, len(doc.Nodes))
	for _, e := range doc.Edges {
		a, ok1 := idx[e.From]
		b, ok2 := idx[e.To]
		if !ok1 || !ok2 || a == b {
			continue
		}
		d.Deg[a]++
		d.Deg[b]++
		d.Edges = append(d.Edges, [2]int{a, b})
	}
	return d
}
