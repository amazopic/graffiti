package render

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/amazopic/graffiti/internal/analyze"
	"github.com/amazopic/graffiti/internal/graph"
)

// WriteMapMD renders MAP.md and writes it to <root>/.graffiti/MAP.md, next to
// map.json. Deterministic modulo nothing (MAP.md carries no timestamp).
func WriteMapMD(doc *graph.Document, an analyze.Analysis, root string) error {
	dir := filepath.Join(root, ".graffiti")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "MAP.md"), []byte(RenderMapMD(doc, an)), 0o644)
}

// RenderMapMD renders the deterministic MAP.md text (spec §8.3/§8.5). It is a
// pure function of the clustered Document + Analysis; no Go-map iteration feeds
// any emitted order. Sections, in fixed order: title, summary, Start here,
// Landmarks (god nodes), Districts (per community), Surprising connections,
// Confidence legend.
func RenderMapMD(doc *graph.Document, an analyze.Analysis) string {
	byID := make(map[string]graph.Node, len(doc.Nodes))
	for _, n := range doc.Nodes {
		byID[n.ID] = n
	}

	var b strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&b, format, args...) }

	w("# %s — Map\n\n", doc.Root)
	w("_%d nodes, %d edges, %d communities. 0 API calls, $0._\n\n",
		len(doc.Nodes), len(doc.Edges), len(doc.Communities))

	w("## Start here\n\n")
	for i, q := range an.Questions {
		w("%d. %s\n", i+1, q)
	}
	w("\n")

	w("## Landmarks (god nodes)\n\n")
	if len(an.GodNodes) == 0 {
		w("_None._\n\n")
	} else {
		for _, g := range an.GodNodes {
			w("- **%s** — touched by %d things — change carefully.\n", g.Label, g.Degree)
		}
		w("\n")
	}

	w("## Districts\n\n")
	comms := append([]graph.Community(nil), doc.Communities...)
	sort.Slice(comms, func(i, j int) bool { return comms[i].ID < comms[j].ID })
	for _, c := range comms {
		w("### %s (%d things)\n\n", c.Label, len(c.Members))
		for _, id := range c.Members { // Members are pre-sorted by cluster.NameCommunities
			n := byID[id]
			w("- `%s` (%s, %s:%d)\n", n.Label, n.Kind, n.File, n.Line)
		}
		w("\n")
	}

	w("## Surprising connections\n\n")
	if len(an.Surprising) == 0 {
		w("_None._\n\n")
	} else {
		for _, s := range an.Surprising {
			w("- `%s` → `%s` (%s, %s)\n", byID[s.From].Label, byID[s.To].Label, s.Relation, s.Confidence)
		}
		w("\n")
	}

	w("## Confidence legend\n\n")
	w("- **EXTRACTED** — definite (verified from imports/syntax).\n")
	w("- **INFERRED** — inferred (same-package name match).\n")
	w("- **AMBIGUOUS** — guessed — verify.\n")

	return b.String()
}
