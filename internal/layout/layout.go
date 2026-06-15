// Package layout bakes a deterministic Tier-1 "Districts" scene (spec §8.2/§8.3)
// from a clustered graph.Document + analyze.Analysis. It performs no I/O and no
// mutation. Geometry is precomputed here as integer coordinates so the browser
// does NO layout and NO physics — the only way to satisfy the §8.8/§14
// byte-identical guarantee. Communities are packed by a squarified treemap (box
// area proportional to member count, no top-tier force); cross-community edges
// aggregate into one bundle per ordered (A->B) pair; god nodes become landmark
// pins; surprising links become dashed gutter arcs. No math/rand, no time, and
// no Go-map iteration ever feeds an emitted order (every emitted slice is
// re-sorted with an explicit total order).
package layout

import (
	"sort"

	"github.com/evgeniy-achin/graffiti/internal/analyze"
	"github.com/evgeniy-achin/graffiti/internal/graph"
)

// Canvas + spacing constants. The fixed canvas keeps coords byte-stable across
// runs; the browser scales it to the viewport.
const (
	CanvasW    = 1600
	CanvasH    = 1000
	Pad        = 24 // outer padding
	Gutter     = 16 // min gap baked between districts
	PinR       = 7
	titleBandH = 28
)

// Box is one community district (integer coords; gutter-inset).
type Box struct {
	CommID      int
	Label       string
	Count       int
	X, Y, W, H int
	Border      int // centrality-derived border weight (1..4)
}

// Pin is a god-node landmark placed inside its district box.
type Pin struct {
	NodeID, Label string
	CommID        int
	X, Y          int
}

// Bundle is one aggregated ordered (FromComm->ToComm) flow of Count edges, with
// a baked orthogonal/elbow polyline through the gutters.
type Bundle struct {
	FromComm, ToComm int
	Count            int
	Pts              [][2]int
}

// Arc is one dashed surprising cross-community link (baked elbow polyline).
type Arc struct {
	FromComm, ToComm int
	Confidence       string
	Pts              [][2]int
}

// Scene is the full deterministic Tier-1 district scene.
type Scene struct {
	W, H    int
	Boxes   []Box
	Pins    []Pin
	Bundles []Bundle
	Arcs    []Arc
}

// Layout produces the deterministic Tier-1 district scene.
func Layout(doc *graph.Document, an analyze.Analysis) Scene {
	sc := Scene{W: CanvasW, H: CanvasH}

	// 1. Community order (by id asc) + member counts.
	comms := append([]graph.Community(nil), doc.Communities...)
	sort.Slice(comms, func(i, j int) bool { return comms[i].ID < comms[j].ID })

	// per-node degree (for centrality-derived border weight) + node->community.
	deg := map[string]int{}
	for _, e := range doc.Edges {
		deg[e.From]++
		deg[e.To]++
	}
	commOf := map[string]int{}
	for _, n := range doc.Nodes {
		commOf[n.ID] = n.Community
	}

	type item struct {
		id    int
		label string
		count int
		cent  int
	}
	items := make([]item, 0, len(comms))
	for _, c := range comms {
		cent := 0
		for _, m := range c.Members {
			cent += deg[m]
		}
		items = append(items, item{c.ID, c.Label, len(c.Members), cent})
	}
	if len(items) == 0 {
		return sc
	}

	// 2. Squarified treemap over the inner rect. Area ∝ count (min 1). Children
	//    processed desc area, ties by asc community id (total order).
	innerX, innerY := Pad, Pad+titleBandH
	innerW, innerH := CanvasW-2*Pad, CanvasH-Pad-(Pad+titleBandH)

	var total float64
	for _, it := range items {
		c := it.count
		if c < 1 {
			c = 1
		}
		total += float64(c)
	}
	scale := float64(innerW*innerH) / total

	order := make([]int, len(items))
	for i := range order {
		order[i] = i
	}
	area := func(i int) float64 {
		c := items[i].count
		if c < 1 {
			c = 1
		}
		return float64(c) * scale
	}
	sort.SliceStable(order, func(a, b int) bool {
		ai, bi := order[a], order[b]
		if area(ai) != area(bi) {
			return area(ai) > area(bi)
		}
		return items[ai].id < items[bi].id
	})

	rects := squarify(order, area, innerX, innerY, innerW, innerH)

	// quantize + gutter-inset; key boxes by community id.
	boxByComm := map[int]Box{}
	for idx, r := range rects {
		it := items[idx]
		b := Box{
			CommID: it.id,
			Label:  it.label,
			Count:  it.count,
			X:      r.x + Gutter/2,
			Y:      r.y + Gutter/2,
			W:      r.w - Gutter,
			H:      r.h - Gutter,
			Border: borderWeight(it.cent),
		}
		if b.W < 8 {
			b.W = 8
		}
		if b.H < 8 {
			b.H = 8
		}
		sc.Boxes = append(sc.Boxes, b)
		boxByComm[it.id] = b
	}
	sort.Slice(sc.Boxes, func(i, j int) bool { return sc.Boxes[i].CommID < sc.Boxes[j].CommID })

	// 3. God-node landmark pins, fanned on a fixed 3-wide grid inside the box.
	pinIdx := map[int]int{}
	for _, g := range an.GodNodes {
		cid, ok := commOf[g.ID]
		if !ok {
			continue
		}
		b, ok := boxByComm[cid]
		if !ok {
			continue
		}
		k := pinIdx[cid]
		pinIdx[cid]++
		px := b.X + 14 + (k%3)*16
		py := b.Y + 16 + (k/3)*16
		if px > b.X+b.W-PinR {
			px = b.X + b.W - PinR
		}
		if py > b.Y+b.H-PinR {
			py = b.Y + b.H - PinR
		}
		sc.Pins = append(sc.Pins, Pin{NodeID: g.ID, Label: g.Label, CommID: cid, X: px, Y: py})
	}
	sort.Slice(sc.Pins, func(i, j int) bool {
		if sc.Pins[i].CommID != sc.Pins[j].CommID {
			return sc.Pins[i].CommID < sc.Pins[j].CommID
		}
		return sc.Pins[i].NodeID < sc.Pins[j].NodeID
	})

	// 4. Aggregate inter-community edges into one bundle per ordered (A->B) pair.
	type key struct{ a, b int }
	bcount := map[key]int{}
	for _, e := range doc.Edges {
		fa, oka := commOf[e.From]
		fb, okb := commOf[e.To]
		if !oka || !okb || fa == fb {
			continue
		}
		bcount[key{fa, fb}]++
	}
	keys := make([]key, 0, len(bcount))
	for k := range bcount {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].a != keys[j].a {
			return keys[i].a < keys[j].a
		}
		return keys[i].b < keys[j].b
	})
	for _, k := range keys {
		ba, oka := boxByComm[k.a]
		bb, okb := boxByComm[k.b]
		if !oka || !okb {
			continue
		}
		sc.Bundles = append(sc.Bundles, Bundle{
			FromComm: k.a, ToComm: k.b, Count: bcount[k],
			Pts: elbow(ba, bb),
		})
	}

	// 5. Surprising links -> dashed arcs (dedup by ordered pair + confidence).
	type akey struct {
		a, b int
		conf string
	}
	seenArc := map[akey]bool{}
	var arcKeys []akey
	for _, s := range an.Surprising {
		ak := akey{s.FromComm, s.ToComm, string(s.Confidence)}
		if seenArc[ak] {
			continue
		}
		seenArc[ak] = true
		arcKeys = append(arcKeys, ak)
	}
	sort.Slice(arcKeys, func(i, j int) bool {
		if arcKeys[i].a != arcKeys[j].a {
			return arcKeys[i].a < arcKeys[j].a
		}
		if arcKeys[i].b != arcKeys[j].b {
			return arcKeys[i].b < arcKeys[j].b
		}
		return arcKeys[i].conf < arcKeys[j].conf
	})
	for _, ak := range arcKeys {
		ba, oka := boxByComm[ak.a]
		bb, okb := boxByComm[ak.b]
		if !oka || !okb {
			continue
		}
		sc.Arcs = append(sc.Arcs, Arc{FromComm: ak.a, ToComm: ak.b, Confidence: ak.conf, Pts: elbow(ba, bb)})
	}

	return sc
}

func borderWeight(cent int) int {
	switch {
	case cent >= 40:
		return 4
	case cent >= 16:
		return 3
	case cent >= 4:
		return 2
	default:
		return 1
	}
}

func cx(b Box) int { return b.X + b.W/2 }
func cy(b Box) int { return b.Y + b.H/2 }

// elbow routes a deterministic integer 3-point orthogonal polyline (horizontal
// then vertical) between two box centers.
func elbow(a, b Box) [][2]int {
	ax, ay := cx(a), cy(a)
	bx, by := cx(b), cy(b)
	return [][2]int{{ax, ay}, {bx, ay}, {bx, by}}
}

// ---- Squarified treemap (validated prototype; integer-quantized) ----

type rect struct{ x, y, w, h int }

type cell struct {
	pos int
	a   float64
}

// squarify returns one rect per position in `order` (sorted desc area). Result[i]
// corresponds to order[i]. Coordinates are integer-quantized at each row close.
func squarify(order []int, area func(int) float64, x, y, w, h int) []rect {
	out := make([]rect, len(order))
	cells := make([]cell, len(order))
	for i, idx := range order {
		cells[i] = cell{i, area(idx)}
	}

	fx, fy, fw, fh := float64(x), float64(y), float64(w), float64(h)
	i := 0
	for i < len(cells) {
		rowStart := i
		short := fw
		if fh < short {
			short = fh
		}
		bestWorst := worstRatio(cells[rowStart:rowStart+1], short)
		j := i + 1
		for j < len(cells) {
			w2 := worstRatio(cells[rowStart:j+1], short)
			if w2 > bestWorst {
				break
			}
			bestWorst = w2
			j++
		}
		row := cells[rowStart:j]
		var rowSum float64
		for _, c := range row {
			rowSum += c.a
		}
		if fw <= fh {
			rh := rowSum / fw
			cxp := fx
			for ri, c := range row {
				cw := c.a / rh
				rx, ry := round(cxp), round(fy)
				rwd := round(cw)
				if ri == len(row)-1 {
					rwd = round(fx+fw) - rx
				}
				out[c.pos] = rect{rx, ry, rwd, round(rh)}
				cxp += cw
			}
			fy += rh
			fh -= rh
		} else {
			rw := rowSum / fh
			cyp := fy
			for ri, c := range row {
				ch := c.a / rw
				rx, ry := round(fx), round(cyp)
				rht := round(ch)
				if ri == len(row)-1 {
					rht = round(fy+fh) - ry
				}
				out[c.pos] = rect{rx, ry, round(rw), rht}
				cyp += ch
			}
			fx += rw
			fw -= rw
		}
		i = j
	}
	return out
}

func worstRatio(row []cell, short float64) float64 {
	var sum, max, min float64
	min = 1e18
	for _, c := range row {
		sum += c.a
		if c.a > max {
			max = c.a
		}
		if c.a < min {
			min = c.a
		}
	}
	if sum == 0 {
		return 1e18
	}
	s2 := short * short
	w1 := s2 * max / (sum * sum)
	w2 := sum * sum / (s2 * min)
	if w1 > w2 {
		return w1
	}
	return w2
}

func round(f float64) int {
	if f < 0 {
		return int(f - 0.5)
	}
	return int(f + 0.5)
}
