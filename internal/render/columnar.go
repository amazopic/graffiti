package render

import "github.com/evgeniy-achin/graffiti/internal/layout"

// ColumnarScene is the compact, JSON-friendly representation inlined as the
// map.html data island (spec §8.6): every repeated string is interned once into
// Strings (index 0 == ""); everything else is parallel integer arrays. The
// browser rebuilds objects by index lookup. JSON keys come from struct field
// order; arrays are already deterministically ordered by layout.Layout, so the
// encoded bytes are deterministic.
type ColumnarScene struct {
	W       int      `json:"w"`
	H2      int      `json:"h"`
	Strings []string `json:"strings"` // interned string table; index 0 == ""

	BoxComm   []int `json:"boxComm"`
	BoxLabel  []int `json:"boxLabel"`
	BoxCount  []int `json:"boxCount"`
	BoxX      []int `json:"boxX"`
	BoxY      []int `json:"boxY"`
	BoxW      []int `json:"boxW"`
	BoxH      []int `json:"boxH"`
	BoxBorder []int `json:"boxBorder"`

	PinNode  []int `json:"pinNode"`
	PinLabel []int `json:"pinLabel"`
	PinComm  []int `json:"pinComm"`
	PinX     []int `json:"pinX"`
	PinY     []int `json:"pinY"`

	// Points for item i live in Pts[Off[i]*2 : Off[i+1]*2] as [x0,y0,x1,y1,...].
	BundleFrom  []int `json:"bundleFrom"`
	BundleTo    []int `json:"bundleTo"`
	BundleCount []int `json:"bundleCount"`
	BundleOff   []int `json:"bundleOff"`
	BundlePts   []int `json:"bundlePts"`

	ArcFrom []int `json:"arcFrom"`
	ArcTo   []int `json:"arcTo"`
	ArcConf []int `json:"arcConf"`
	ArcOff  []int `json:"arcOff"`
	ArcPts  []int `json:"arcPts"`
}

// interner builds a deterministic interned string table. The empty string is
// always index 0. Strings are assigned ids in first-seen order, deterministic
// because the Scene emit order is.
type interner struct {
	idx map[string]int
	tab []string
}

func newInterner() *interner {
	in := &interner{idx: map[string]int{}}
	in.intern("") // reserve index 0
	return in
}

func (in *interner) intern(s string) int {
	if id, ok := in.idx[s]; ok {
		return id
	}
	id := len(in.tab)
	in.idx[s] = id
	in.tab = append(in.tab, s)
	return id
}

// toColumnar flattens a layout.Scene into the compact columnar form.
func toColumnar(sc layout.Scene) ColumnarScene {
	in := newInterner()
	cs := ColumnarScene{W: sc.W, H2: sc.H}

	for _, b := range sc.Boxes {
		cs.BoxComm = append(cs.BoxComm, b.CommID)
		cs.BoxLabel = append(cs.BoxLabel, in.intern(b.Label))
		cs.BoxCount = append(cs.BoxCount, b.Count)
		cs.BoxX = append(cs.BoxX, b.X)
		cs.BoxY = append(cs.BoxY, b.Y)
		cs.BoxW = append(cs.BoxW, b.W)
		cs.BoxH = append(cs.BoxH, b.H)
		cs.BoxBorder = append(cs.BoxBorder, b.Border)
	}
	for _, p := range sc.Pins {
		cs.PinNode = append(cs.PinNode, in.intern(p.NodeID))
		cs.PinLabel = append(cs.PinLabel, in.intern(p.Label))
		cs.PinComm = append(cs.PinComm, p.CommID)
		cs.PinX = append(cs.PinX, p.X)
		cs.PinY = append(cs.PinY, p.Y)
	}

	cs.BundleOff = append(cs.BundleOff, 0)
	for _, bn := range sc.Bundles {
		cs.BundleFrom = append(cs.BundleFrom, bn.FromComm)
		cs.BundleTo = append(cs.BundleTo, bn.ToComm)
		cs.BundleCount = append(cs.BundleCount, bn.Count)
		for _, pt := range bn.Pts {
			cs.BundlePts = append(cs.BundlePts, pt[0], pt[1])
		}
		cs.BundleOff = append(cs.BundleOff, len(cs.BundlePts)/2)
	}

	cs.ArcOff = append(cs.ArcOff, 0)
	for _, ar := range sc.Arcs {
		cs.ArcFrom = append(cs.ArcFrom, ar.FromComm)
		cs.ArcTo = append(cs.ArcTo, ar.ToComm)
		cs.ArcConf = append(cs.ArcConf, in.intern(ar.Confidence))
		for _, pt := range ar.Pts {
			cs.ArcPts = append(cs.ArcPts, pt[0], pt[1])
		}
		cs.ArcOff = append(cs.ArcOff, len(cs.ArcPts)/2)
	}

	cs.Strings = in.tab
	return cs
}
