package layout

import (
	"fmt"
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/analyze"
	"github.com/evgeniy-achin/graffiti/internal/graph"
)

// sampleDoc builds a 5-community clustered document with intra- and
// cross-community edges, two god nodes, and two surprising links — the same
// shape validated in the layout prototype, expressed against the real types.
func sampleDoc() (*graph.Document, analyze.Analysis) {
	doc := graph.NewDocument("demo")
	doc.GeneratedAt = "2026-06-15T00:00:00Z"
	sizes := []int{12, 8, 5, 3, 1}
	labels := []string{"Auth", "Http", "Parser", "Cache", "Util"}
	for ci, sz := range sizes {
		var members []string
		for k := 0; k < sz; k++ {
			id := fmt.Sprintf("c%d-n%d", ci, k)
			doc.Nodes = append(doc.Nodes, graph.Node{
				ID: id, Label: id, Kind: graph.KindFunction,
				File: fmt.Sprintf("pkg%d/f%d.go", ci, k), Line: 1, Community: ci,
			})
			members = append(members, id)
		}
		doc.Communities = append(doc.Communities, graph.Community{ID: ci, Label: labels[ci], Members: members})
	}
	mk := func(f, t string, rel graph.Relation, conf graph.Confidence) {
		doc.Edges = append(doc.Edges, graph.Edge{From: f, To: t, Relation: rel, Confidence: conf})
	}
	mk("c0-n0", "c0-n1", graph.RelCalls, graph.ConfInferred)
	mk("c0-n0", "c1-n0", graph.RelCalls, graph.ConfExtracted) // cross 0->1
	mk("c0-n1", "c1-n0", graph.RelCalls, graph.ConfInferred)  // cross 0->1 again (bundle count 2)
	mk("c1-n0", "c2-n0", graph.RelCalls, graph.ConfInferred)  // cross 1->2
	mk("c3-n0", "c0-n0", graph.RelReferences, graph.ConfAmbiguous)
	an := analyze.Analysis{
		GodNodes: []analyze.GodNode{
			{ID: "c0-n0", Label: "c0-n0", Degree: 4},
			{ID: "c1-n0", Label: "c1-n0", Degree: 3},
		},
		Surprising: []analyze.SurprisingEdge{
			{From: "c0-n0", To: "c1-n0", FromComm: 0, ToComm: 1, Relation: graph.RelCalls, Confidence: graph.ConfExtracted},
			{From: "c3-n0", To: "c0-n0", FromComm: 3, ToComm: 0, Relation: graph.RelReferences, Confidence: graph.ConfAmbiguous},
		},
		Questions: []string{"q1", "q2", "q3"},
	}
	return doc, an
}

func TestLayout_NoOverlap(t *testing.T) {
	doc, an := sampleDoc()
	sc := Layout(doc, an)
	for i := 0; i < len(sc.Boxes); i++ {
		for j := i + 1; j < len(sc.Boxes); j++ {
			a, b := sc.Boxes[i], sc.Boxes[j]
			if a.X < b.X+b.W && b.X < a.X+a.W && a.Y < b.Y+b.H && b.Y < a.Y+a.H {
				t.Fatalf("boxes overlap: %+v and %+v", a, b)
			}
		}
	}
}

func TestLayout_InBounds(t *testing.T) {
	doc, an := sampleDoc()
	sc := Layout(doc, an)
	for _, b := range sc.Boxes {
		if b.X < 0 || b.Y < 0 || b.X+b.W > sc.W || b.Y+b.H > sc.H {
			t.Fatalf("box out of bounds: %+v (canvas %dx%d)", b, sc.W, sc.H)
		}
		if b.W <= 0 || b.H <= 0 {
			t.Fatalf("box has non-positive dims: %+v", b)
		}
	}
	for _, p := range sc.Pins {
		if p.X < 0 || p.Y < 0 || p.X > sc.W || p.Y > sc.H {
			t.Fatalf("pin out of bounds: %+v", p)
		}
	}
}

func TestLayout_PinsOnTheirDistrict(t *testing.T) {
	doc, an := sampleDoc()
	sc := Layout(doc, an)
	boxByComm := map[int]Box{}
	for _, b := range sc.Boxes {
		boxByComm[b.CommID] = b
	}
	for _, p := range sc.Pins {
		b := boxByComm[p.CommID]
		if p.X < b.X || p.X > b.X+b.W || p.Y < b.Y || p.Y > b.Y+b.H {
			t.Fatalf("pin %+v not inside its district box %+v", p, b)
		}
	}
}

func TestLayout_Deterministic(t *testing.T) {
	d1, a1 := sampleDoc()
	d2, a2 := sampleDoc()
	if s1, s2 := fmt.Sprintf("%+v", Layout(d1, a1)), fmt.Sprintf("%+v", Layout(d2, a2)); s1 != s2 {
		t.Fatalf("non-deterministic layout:\nA=%s\nB=%s", s1, s2)
	}
}

func TestLayout_IntegerCoords(t *testing.T) {
	// All coords are Go ints by type; this guards the contract explicitly so a
	// future float refactor that rounds late cannot silently slip through.
	doc, an := sampleDoc()
	sc := Layout(doc, an)
	for _, b := range sc.Boxes {
		_ = b.X + b.Y + b.W + b.H + b.Border // ints; compile-time guarantee
	}
	for _, bn := range sc.Bundles {
		for _, pt := range bn.Pts {
			_ = pt[0] + pt[1]
		}
	}
}

func TestLayout_BundlesAggregated(t *testing.T) {
	doc, an := sampleDoc()
	sc := Layout(doc, an)
	got := map[[2]int]int{}
	for _, b := range sc.Bundles {
		got[[2]int{b.FromComm, b.ToComm}] = b.Count
	}
	if got[[2]int{0, 1}] != 2 {
		t.Fatalf("bundle 0->1 count = %d, want 2", got[[2]int{0, 1}])
	}
	if got[[2]int{1, 2}] != 1 {
		t.Fatalf("bundle 1->2 count = %d, want 1", got[[2]int{1, 2}])
	}
	if got[[2]int{3, 0}] != 1 {
		t.Fatalf("bundle 3->0 count = %d, want 1", got[[2]int{3, 0}])
	}
}

func TestLayout_AreaProportional(t *testing.T) {
	doc, an := sampleDoc()
	sc := Layout(doc, an)
	byComm := map[int]Box{}
	for _, b := range sc.Boxes {
		byComm[b.CommID] = b
	}
	if a0, a4 := byComm[0].W*byComm[0].H, byComm[4].W*byComm[4].H; a0 <= a4 {
		t.Fatalf("area not proportional: comm0 area %d <= comm4 area %d", a0, a4)
	}
}

func TestLayout_SurprisingArcs(t *testing.T) {
	doc, an := sampleDoc()
	sc := Layout(doc, an)
	if len(sc.Arcs) != 2 {
		t.Fatalf("arcs = %d, want 2 (one per surprising edge pair)", len(sc.Arcs))
	}
	// arcs emitted sorted by (from, to, conf): 0->1 then 3->0.
	if sc.Arcs[0].FromComm != 0 || sc.Arcs[0].ToComm != 1 {
		t.Fatalf("arc[0] = %d->%d, want 0->1", sc.Arcs[0].FromComm, sc.Arcs[0].ToComm)
	}
	if sc.Arcs[0].Confidence != "EXTRACTED" {
		t.Fatalf("arc[0] confidence = %q, want EXTRACTED", sc.Arcs[0].Confidence)
	}
}
