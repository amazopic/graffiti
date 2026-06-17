package parse

import (
	"strings"

	ts "github.com/odvcencio/gotreesitter"

	"github.com/amazopic/graffiti/internal/graph"
)

// Extract runs Pass 1 over one file using a table-driven LangSpec (spec §5). It
// emits a file node, definition nodes (functions, classes/structs/interfaces/
// enums/traits, and methods labeled "Class.method"), contains edges, imports
// edges to synthesized module nodes, and stashes raw call sites for Pass 2.
//
// Compared with ParseGo it deliberately UNDER-extracts (no Go-style receiver
// typing, no nested-definition recursion past a function body): honesty-first,
// per the §16 doctrine. The shared helpers emitDef/fieldText/Walk/importBase are
// reused from the package.
func Extract(p Parser, relPath string, src []byte, spec LangSpec) (*Extraction, error) {
	tree, err := p.Parse(src)
	if err != nil {
		return nil, err
	}
	lang := tree.Lang()
	root := tree.Root()

	ex := &Extraction{File: relPath}
	fileID := graph.NodeID(relPath, relPath)
	ex.Nodes = append(ex.Nodes, graph.Node{
		ID: fileID, Label: relPath, Kind: graph.KindFile, File: relPath, Line: 1,
		Community: graph.UnclusteredCommunity,
	})

	classSet := toSet(spec.ClassKinds)
	methodSet := toSet(spec.MethodKinds)
	funcSet := toSet(spec.FuncKinds)
	implSet := toSet(spec.ImplKinds)
	importSet := toSet(spec.ImportKinds)

	// 1) Imports (flat walk): emit a module node (keyed by raw import text) + edge.
	var importPaths []string
	Walk(root, func(n Node) {
		if !importSet[n.Type(lang)] {
			return
		}
		imp := importText(n, spec)
		if imp == "" {
			return
		}
		importPaths = append(importPaths, imp)
		label := lastSegment(imp)
		modID := graph.NodeID("module:"+imp, label)
		ex.Nodes = append(ex.Nodes, graph.Node{
			ID: modID, Label: label, Kind: graph.KindModule, File: relPath,
			Line: int(n.StartPoint().Row) + 1, Community: graph.UnclusteredCommunity,
		})
		ex.Edges = append(ex.Edges, graph.Edge{
			From: fileID, To: modID, Relation: graph.RelImports, Confidence: graph.ConfExtracted,
		})
	})

	// 2) Definitions: structured recursion so methods carry their class qualifier.
	var visit func(n Node, qualifier string)
	visit = func(n Node, qualifier string) {
		k := n.Type(lang)
		switch {
		case classSet[k]:
			name := fieldText(n, "name")
			if name != "" {
				defID := graph.NodeID(relPath, name)
				emitDef(ex, fileID, defID, name, graph.KindClass, relPath, line(n))
				for _, c := range n.NamedChildren() {
					visit(c, name)
				}
				return
			}
		case implSet[k]:
			// Rust impl: qualifier is the implemented type's name.
			q := namedChildText(n, "type_identifier", lang)
			for _, c := range n.NamedChildren() {
				visit(c, q)
			}
			return
		case methodSet[k] && qualifier != "":
			name := fieldText(n, "name")
			if name != "" {
				label := qualifier + "." + name
				defID := graph.NodeID(relPath, label)
				emitDef(ex, fileID, defID, label, graph.KindMethod, relPath, line(n))
				collectCallsSpec(ex, n, defID, relPath, importPaths, lang, spec)
				return
			}
		case funcSet[k] && qualifier == "":
			name := fieldText(n, "name")
			if name != "" {
				defID := graph.NodeID(relPath, name)
				emitDef(ex, fileID, defID, name, graph.KindFunction, relPath, line(n))
				collectCallsSpec(ex, n, defID, relPath, importPaths, lang, spec)
				return
			}
		}
		for _, c := range n.NamedChildren() {
			visit(c, qualifier)
		}
	}
	visit(root, "")

	return ex, nil
}

// collectCallsSpec walks a definition subtree and stashes each call site's callee
// text (extracted from the grammar field named by spec.CallKinds[kind]).
func collectCallsSpec(ex *Extraction, defNode Node, defID, file string, importPaths []string, lang *ts.Language, spec LangSpec) {
	Walk(defNode, func(n Node) {
		field, ok := spec.CallKinds[n.Type(lang)]
		if !ok {
			return
		}
		c := n.ChildByField(field)
		if c == nil {
			return
		}
		callee := strings.TrimSpace(c.Text())
		if callee == "" {
			return
		}
		ex.RawCalls = append(ex.RawCalls, RawCall{
			FromID: defID, Callee: callee, Line: int(n.StartPoint().Row) + 1, File: file,
			Imports: importPaths,
		})
	})
}

func line(n Node) int { return int(n.StartPoint().Row) + 1 }

func toSet(xs []string) map[string]bool {
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}
	return m
}

// namedChildText returns the text of the first named child of the given kind.
func namedChildText(n Node, kind string, lang *ts.Language) string {
	for _, c := range n.NamedChildren() {
		if c.Type(lang) == kind {
			return strings.TrimSpace(c.Text())
		}
	}
	return ""
}

// importText extracts a best-effort module path from an import node. JS/TS keep
// the module string in a dedicated child (spec.ImportChild); everything else uses
// the first named child (a dotted/scoped/namespace name).
func importText(n Node, spec LangSpec) string {
	if spec.ImportChild != "" {
		// find a descendant of the requested kind (the module string)
		var found string
		Walk(n, func(m Node) {
			if found != "" {
				return
			}
			// note: Type needs lang; importText has no lang, so match by quotes instead.
			t := strings.TrimSpace(m.Text())
			if len(t) >= 2 && (t[0] == '\'' || t[0] == '"' || t[0] == '`') {
				found = stripQuotes(t)
			}
		})
		if found != "" {
			return found
		}
	}
	kids := n.NamedChildren()
	if len(kids) == 0 {
		return ""
	}
	return strings.TrimSpace(kids[0].Text())
}

func stripQuotes(s string) string {
	if len(s) >= 2 {
		switch s[0] {
		case '\'', '"', '`':
			if s[len(s)-1] == s[0] {
				return s[1 : len(s)-1]
			}
		}
	}
	return s
}

// codeExts are source-file extensions stripped from module labels so that a
// path-style import like "./auth/session.js" labels as "session", not "js".
var codeExts = []string{".jsx", ".tsx", ".mjs", ".cjs", ".js", ".ts", ".py", ".rs", ".java", ".php"}

// lastSegment returns a human-friendly module label: the final path/namespace
// component of an import. Path separators ("/", "\", "::") are handled before a
// trailing source extension is stripped and before the "." namespace separator,
// so "./auth/session.js" -> "session" (JS/TS) while "java.util.List" -> "List".
func lastSegment(imp string) string {
	imp = strings.TrimSpace(imp)
	for _, sep := range []string{"::", "\\", "/"} {
		if i := strings.LastIndex(imp, sep); i >= 0 {
			imp = imp[i+len(sep):]
		}
	}
	for _, ext := range codeExts {
		if strings.HasSuffix(imp, ext) {
			imp = strings.TrimSuffix(imp, ext)
			break
		}
	}
	if i := strings.LastIndex(imp, "."); i >= 0 {
		imp = imp[i+1:]
	}
	return imp
}
