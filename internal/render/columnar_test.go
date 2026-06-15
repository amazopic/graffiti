package render

import (
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/layout"
)

// tinyScene is a hand-built Scene with two boxes, one pin, one bundle, one arc —
// enough to exercise every column + the interned string table.
func tinyScene() layout.Scene {
	return layout.Scene{
		W: 1600, H: 1000,
		Boxes: []layout.Box{
			{CommID: 0, Label: "Auth", Count: 3, X: 10, Y: 20, W: 100, H: 80, Border: 2},
			{CommID: 1, Label: "Http", Count: 1, X: 120, Y: 20, W: 40, H: 80, Border: 1},
		},
		Pins: []layout.Pin{
			{NodeID: "auth-login", Label: "Login", CommID: 0, X: 24, Y: 36},
		},
		Bundles: []layout.Bundle{
			{FromComm: 0, ToComm: 1, Count: 2, Pts: [][2]int{{60, 60}, {140, 60}, {140, 60}}},
		},
		Arcs: []layout.Arc{
			{FromComm: 1, ToComm: 0, Confidence: "AMBIGUOUS", Pts: [][2]int{{140, 60}, {60, 60}, {60, 60}}},
		},
	}
}

func TestToColumnar_ParallelArraysAndStringTable(t *testing.T) {
	sc := tinyScene()
	cs := toColumnar(sc)

	if cs.W != sc.W || cs.H2 != sc.H {
		t.Fatalf("canvas dims = %dx%d, want %dx%d", cs.W, cs.H2, sc.W, sc.H)
	}
	if len(cs.Strings) == 0 || cs.Strings[0] != "" {
		t.Fatalf("string table must exist with index 0 == \"\", got %v", cs.Strings)
	}
	nb := len(sc.Boxes)
	for name, col := range map[string][]int{
		"BoxComm": cs.BoxComm, "BoxLabel": cs.BoxLabel, "BoxCount": cs.BoxCount,
		"BoxX": cs.BoxX, "BoxY": cs.BoxY, "BoxW": cs.BoxW, "BoxH": cs.BoxH, "BoxBorder": cs.BoxBorder,
	} {
		if len(col) != nb {
			t.Fatalf("box column %s len %d, want %d", name, len(col), nb)
		}
	}
	// label indices resolve back to the real labels.
	for i, b := range sc.Boxes {
		if cs.Strings[cs.BoxLabel[i]] != b.Label {
			t.Fatalf("box %d label via table = %q, want %q", i, cs.Strings[cs.BoxLabel[i]], b.Label)
		}
	}
	// offset arrays are prefix sums of length n+1; last*2 == flattened pts len.
	if len(cs.BundleOff) != len(sc.Bundles)+1 {
		t.Fatalf("bundleOff len %d, want %d", len(cs.BundleOff), len(sc.Bundles)+1)
	}
	if got := cs.BundleOff[len(cs.BundleOff)-1] * 2; got != len(cs.BundlePts) {
		t.Fatalf("bundlePts len %d, want %d", len(cs.BundlePts), got)
	}
	if got := cs.ArcOff[len(cs.ArcOff)-1] * 2; got != len(cs.ArcPts) {
		t.Fatalf("arcPts len %d, want %d", len(cs.ArcPts), got)
	}
}

func TestToColumnar_Deterministic(t *testing.T) {
	a, b := toColumnar(tinyScene()), toColumnar(tinyScene())
	if a.W != b.W || len(a.Strings) != len(b.Strings) || len(a.BoxX) != len(b.BoxX) {
		t.Fatalf("columnar encoding not deterministic")
	}
	for i := range a.Strings {
		if a.Strings[i] != b.Strings[i] {
			t.Fatalf("string table index %d differs: %q vs %q", i, a.Strings[i], b.Strings[i])
		}
	}
}
