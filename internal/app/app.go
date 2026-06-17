// Package app wires the graffiti pipeline (scan → parse → build → render) for a
// single Go repository (Plan 1 scope).
package app

import (
	"os"
	"path/filepath"

	"github.com/amazopic/graffiti/internal/analyze"
	"github.com/amazopic/graffiti/internal/build"
	"github.com/amazopic/graffiti/internal/cache"
	"github.com/amazopic/graffiti/internal/cluster"
	"github.com/amazopic/graffiti/internal/graph"
	"github.com/amazopic/graffiti/internal/parse"
	"github.com/amazopic/graffiti/internal/render"
	"github.com/amazopic/graffiti/internal/scan"
)

// Stats summarizes a build for the CLI success line.
type Stats struct {
	Files       int
	Nodes       int
	Edges       int
	Communities int
	HasDocNode  bool     // whether a markdown doc node was emitted
	Questions   []string // the 3 suggested questions (spec §11), deterministic
}

// Build runs the full pipeline against root, stamping generatedAt into the
// document (via build.Assemble), and returns Stats. generatedAt should be RFC3339.
//
// The document root is set to filepath.Base(absRoot) so map.json is byte-identical
// for the same repo regardless of the absolute build directory (determinism, §14).
func Build(root, generatedAt string) (Stats, error) {
	var stats Stats

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return stats, err
	}
	docRoot := filepath.Base(absRoot) // stable, location-independent

	refs, err := scan.Scan(absRoot)
	if err != nil {
		return stats, err
	}
	stats.Files = len(refs)

	goParser, err := parse.NewGoParser()
	if err != nil {
		return stats, err
	}

	c := cache.New(absRoot)
	_ = c.Load() // loaded for forward-compat; Plan 1 does not skip on hash match

	parsers := map[scan.Lang]parse.Parser{}

	var extractions []*parse.Extraction
	for _, ref := range refs {
		src, readErr := os.ReadFile(ref.AbsPath)
		if readErr != nil {
			return stats, readErr
		}
		_ = c.Put(ref.RelPath, cache.HashBytes(src))

		switch ref.Lang {
		case scan.LangGo:
			ex, perr := parse.ParseGo(goParser, ref.RelPath, src)
			if perr != nil {
				return stats, perr
			}
			extractions = append(extractions, ex)
		case scan.LangMarkdown:
			extractions = append(extractions, markdownExtraction(ref.RelPath))
			stats.HasDocNode = true
		default:
			spec, ok := parse.SpecFor(ref.Lang)
			if !ok {
				continue // unsupported language reached scan but has no extractor; skip
			}
			p, perr := parserFor(parsers, ref.Lang)
			if perr != nil {
				return stats, perr
			}
			ex, perr := parse.Extract(p, ref.RelPath, src, spec)
			if perr != nil {
				return stats, perr
			}
			extractions = append(extractions, ex)
		}
	}

	doc, err := build.Assemble(docRoot, generatedAt, extractions)
	if err != nil {
		return stats, err
	}

	// Plan 2: cluster, name communities, analyze — all deterministic, no I/O above render.
	cluster.Cluster(doc)
	deg := analyze.Degrees(doc)
	doc.Communities = cluster.NameCommunities(doc, deg)
	an := analyze.Analyze(doc, deg)

	if err := render.WriteMapJSON(doc, absRoot); err != nil {
		return stats, err
	}
	if err := render.WriteMapMD(doc, an, absRoot); err != nil {
		return stats, err
	}
	// Plan 9: map.html is an in-browser force-directed graph; layout is client-side.
	if err := render.WriteMapHTML(doc, an, absRoot); err != nil {
		return stats, err
	}
	if err := c.Flush(); err != nil {
		return stats, err
	}

	stats.Nodes = len(doc.Nodes)
	stats.Edges = len(doc.Edges)
	stats.Communities = len(doc.Communities)
	stats.Questions = an.Questions
	return stats, nil
}

// markdownExtraction emits a single doc node for a Markdown file (no parsing).
func markdownExtraction(relPath string) *parse.Extraction {
	id := graph.NodeID(relPath, relPath)
	return &parse.Extraction{
		File: relPath,
		Nodes: []graph.Node{
			{ID: id, Label: relPath, Kind: graph.KindDoc, File: relPath, Line: 1, Community: graph.UnclusteredCommunity},
		},
	}
}

// parserFor lazily constructs and caches a parser per language so each grammar
// blob is loaded at most once per build.
func parserFor(cache map[scan.Lang]parse.Parser, l scan.Lang) (parse.Parser, error) {
	if p, ok := cache[l]; ok {
		return p, nil
	}
	p, err := parse.NewParser(l)
	if err != nil {
		return nil, err
	}
	cache[l] = p
	return p, nil
}
