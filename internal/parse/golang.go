package parse

import (
	"strings"

	ts "github.com/odvcencio/gotreesitter"

	"github.com/amazopic/graffiti/internal/graph"
)

// RawCall is an unresolved call site stashed by Pass 1 for Pass-2 resolution.
type RawCall struct {
	FromID  string   // node id of the enclosing definition (function/method)
	Callee  string   // call target text: bare "Hello" or selector "fmt.Sprintf"
	Line    int      // 1-based line of the call site
	File    string   // repo-relative file the call occurs in
	Imports []string // full import paths visible in the file (for Pass-2 decision)
}

// Extraction is the per-file output of Pass 1 (spec §5).
type Extraction struct {
	File     string
	Nodes    []graph.Node
	Edges    []graph.Edge
	RawCalls []RawCall
}

// ParseGo runs Pass 1 over one Go file: it emits a file node, definition nodes
// (function/method/type), imports edges (EXTRACTED) to synthesized module nodes
// keyed by full import path, contains edges, and stashes raw call sites.
func ParseGo(p Parser, relPath string, src []byte) (*Extraction, error) {
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

	// 1) Imports: walk import_spec nodes; emit module node (keyed by full path) + imports edge.
	var importPaths []string
	Walk(root, func(n Node) {
		if n.Type(lang) != "import_spec" {
			return
		}
		pathNode := n.ChildByField("path")
		if pathNode == nil {
			return
		}
		imp := unquote(pathNode.Text())
		if imp == "" {
			return
		}
		importPaths = append(importPaths, imp)
		modID := graph.NodeID("module:"+imp, importBase(imp))
		ex.Nodes = append(ex.Nodes, graph.Node{
			ID: modID, Label: importBase(imp), Kind: graph.KindModule, File: relPath,
			Line: int(n.StartPoint().Row) + 1, Community: graph.UnclusteredCommunity,
		})
		ex.Edges = append(ex.Edges, graph.Edge{
			From: fileID, To: modID, Relation: graph.RelImports, Confidence: graph.ConfExtracted,
		})
	})

	// 2) Definitions: function_declaration, method_declaration, type_spec.
	Walk(root, func(n Node) {
		switch n.Type(lang) {
		case "function_declaration":
			name := fieldText(n, "name")
			if name == "" {
				return
			}
			defID := graph.NodeID(relPath, name)
			emitDef(ex, fileID, defID, name, graph.KindFunction, relPath, int(n.StartPoint().Row)+1)
			collectCalls(ex, n, defID, relPath, importPaths, lang)
		case "method_declaration":
			name := fieldText(n, "name")
			if name == "" {
				return
			}
			recv := receiverTypeName(n, lang)
			label := name
			if recv != "" {
				label = recv + "." + name
			}
			defID := graph.NodeID(relPath, label)
			emitDef(ex, fileID, defID, label, graph.KindMethod, relPath, int(n.StartPoint().Row)+1)
			collectCalls(ex, n, defID, relPath, importPaths, lang)
		case "type_spec":
			name := fieldText(n, "name")
			if name == "" {
				return
			}
			defID := graph.NodeID(relPath, name)
			emitDef(ex, fileID, defID, name, graph.KindClass, relPath, int(n.StartPoint().Row)+1)
		}
	})

	return ex, nil
}

func emitDef(ex *Extraction, fileID, defID, label string, kind graph.Kind, file string, line int) {
	ex.Nodes = append(ex.Nodes, graph.Node{
		ID: defID, Label: label, Kind: kind, File: file, Line: line,
		Community: graph.UnclusteredCommunity,
	})
	ex.Edges = append(ex.Edges, graph.Edge{
		From: fileID, To: defID, Relation: graph.RelContains, Confidence: graph.ConfExtracted,
	})
}

// collectCalls walks a definition subtree and stashes every call_expression's
// callee as a RawCall attributed to defID.
func collectCalls(ex *Extraction, defNode Node, defID, file string, importPaths []string, lang *ts.Language) {
	Walk(defNode, func(n Node) {
		if n.Type(lang) != "call_expression" {
			return
		}
		fn := n.ChildByField("function")
		if fn == nil {
			return
		}
		callee := strings.TrimSpace(fn.Text())
		if callee == "" {
			return
		}
		ex.RawCalls = append(ex.RawCalls, RawCall{
			FromID: defID, Callee: callee, Line: int(n.StartPoint().Row) + 1, File: file,
			Imports: importPaths,
		})
	})
}

func fieldText(n Node, field string) string {
	c := n.ChildByField(field)
	if c == nil {
		return ""
	}
	return strings.TrimSpace(c.Text())
}

// receiverTypeName extracts the receiver's bare type name from a method_declaration.
// e.g. "(g Greeter)" -> "Greeter"; "(g *Greeter)" -> "Greeter".
func receiverTypeName(method Node, lang *ts.Language) string {
	recv := method.ChildByField("receiver")
	if recv == nil {
		return ""
	}
	var typeName string
	Walk(recv, func(n Node) {
		if typeName != "" {
			return
		}
		if n.Type(lang) == "type_identifier" {
			typeName = strings.TrimSpace(n.Text())
		}
	})
	return typeName
}

// unquote strips a single layer of surrounding double quotes from a Go string
// literal (the import_spec path text is quoted, e.g. `"fmt"`).
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// importBase returns the last path segment of an import path for the module label.
func importBase(imp string) string {
	if i := strings.LastIndex(imp, "/"); i >= 0 {
		return imp[i+1:]
	}
	return imp
}
