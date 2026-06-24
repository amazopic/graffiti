// Package contract extracts a repository's "contract surface" — the cross-service
// wire contracts it PROVIDES (HTTP routes, gRPC methods, queue topics it serves)
// and CONSUMES (outbound HTTP calls, queue topics it publishes). It is the input
// to graffiti's system-level cross-service link discovery (internal/system).
//
// It is decoupled from the language parsers: it walks the repo and reads declared
// specs (openapi.json, *.proto, graffiti.contract.json) plus a set of conservative,
// language-agnostic source heuristics (router DSLs, literal URLs, queue calls).
// Everything is deterministic and sorted; confidence is always recorded and
// nothing low-confidence is asserted as fact.
package contract

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/amazopic/graffiti/internal/graph"
)

// skipDirs are never walked for contract sources.
var skipDirs = map[string]bool{
	".graffiti": true, ".graffiti-workspace": true, ".graffiti-system": true,
	".git": true, "node_modules": true, "vendor": true, "dist": true,
	"build": true, "target": true, ".venv": true, "__pycache__": true,
}

// codeExts are scanned by the source heuristics (router/URL/queue patterns).
var codeExts = map[string]bool{
	".go": true, ".js": true, ".mjs": true, ".ts": true, ".tsx": true, ".jsx": true,
	".py": true, ".rs": true, ".java": true, ".kt": true, ".php": true, ".rb": true,
	".cs": true, ".vue": true, ".svelte": true,
}

const maxFileBytes = 1 << 20 // 1 MiB — skip larger files for the heuristics

type srcFile struct {
	rel  string
	data []byte
}

// Extract walks root and returns the sorted, de-duplicated contract surface.
// doc is the already-assembled document, used to associate each endpoint with the
// nearest enclosing graph node (handler / call site) by file:line.
func Extract(root string, doc *graph.Document) (provides, consumes []graph.Endpoint) {
	var p, c []graph.Endpoint
	for _, f := range walk(root) {
		base := strings.ToLower(filepath.Base(f.rel))
		ext := strings.ToLower(filepath.Ext(f.rel))
		switch {
		case base == "graffiti.contract.json":
			pp, cc := parseContractJSON(f.rel, f.data)
			p, c = append(p, pp...), append(c, cc...)
		case base == "openapi.json" || base == "swagger.json":
			p = append(p, parseOpenAPI(f.rel, f.data)...)
		case ext == ".proto":
			p = append(p, parseProto(f.rel, f.data)...)
		}
		if codeExts[ext] {
			pp, cc := scanSource(f.rel, f.data)
			p, c = append(p, pp...), append(c, cc...)
		}
	}
	associate(doc, p)
	associate(doc, c)
	return finalize(p), finalize(c)
}

// walk returns the relevant files (specs + code), eagerly read (size-capped),
// in deterministic path order.
func walk(root string) []srcFile {
	var out []srcFile
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != root && skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		base := strings.ToLower(d.Name())
		ext := strings.ToLower(filepath.Ext(base))
		relevant := codeExts[ext] || ext == ".proto" ||
			base == "openapi.json" || base == "swagger.json" || base == "graffiti.contract.json"
		if !relevant {
			return nil
		}
		if fi, e := d.Info(); e == nil && fi.Size() > maxFileBytes {
			return nil
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return nil
		}
		rel, e := filepath.Rel(root, path)
		if e != nil {
			return nil
		}
		out = append(out, srcFile{rel: filepath.ToSlash(rel), data: b})
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].rel < out[j].rel })
	return out
}

// associate fills ep.Node with the nearest node in the same file at/above ep.Line.
func associate(doc *graph.Document, eps []graph.Endpoint) {
	if doc == nil {
		return
	}
	byFile := map[string][]graph.Node{}
	for _, n := range doc.Nodes {
		byFile[n.File] = append(byFile[n.File], n)
	}
	for f := range byFile {
		ns := byFile[f]
		sort.Slice(ns, func(i, j int) bool { return ns[i].Line < ns[j].Line })
		byFile[f] = ns
	}
	for i := range eps {
		if eps[i].Node != "" {
			continue
		}
		best := ""
		for _, n := range byFile[eps[i].File] {
			if n.Line <= eps[i].Line {
				best = n.ID
			} else {
				break
			}
		}
		eps[i].Node = best
	}
}

// finalize sorts and de-duplicates an endpoint slice deterministically.
func finalize(eps []graph.Endpoint) []graph.Endpoint {
	sort.Slice(eps, func(i, j int) bool {
		a, b := eps[i], eps[j]
		switch {
		case a.Kind != b.Kind:
			return a.Kind < b.Kind
		case a.Key != b.Key:
			return a.Key < b.Key
		case a.File != b.File:
			return a.File < b.File
		case a.Line != b.Line:
			return a.Line < b.Line
		default:
			return a.Source < b.Source
		}
	})
	type dk struct {
		kind graph.EndpointKind
		key  string
		file string
		line int
	}
	seen := map[dk]bool{}
	out := make([]graph.Endpoint, 0, len(eps))
	for _, e := range eps {
		k := dk{e.Kind, e.Key, e.File, e.Line}
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, e)
	}
	return out
}
