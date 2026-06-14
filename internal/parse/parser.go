// Package parse turns source files into graffiti graph extractions (spec §5 parse).
// The tree-sitter backend is hidden behind the Parser/Tree/Node interfaces so it
// can be swapped (e.g. a per-language wazero+WASM fallback) without changing the
// rest of the pipeline.
package parse

import ts "github.com/odvcencio/gotreesitter"

// Node is a backend-agnostic AST node.
type Node interface {
	Type(lang *ts.Language) string  // tree-sitter node type, e.g. "function_declaration"
	StartPoint() ts.Point           // 0-based start position; 1-based line = int(StartPoint().Row)+1
	StartByte() int                 // byte offset into source
	EndByte() int                   // byte offset into source
	Text() string                   // source text spanned by the node
	ChildByField(field string) Node // named child for a grammar field (or nil)
	NamedChildren() []Node          // named children in source order
}

// Tree is a parsed syntax tree.
type Tree interface {
	Root() Node
	Lang() *ts.Language
}

// Parser parses source bytes into a Tree.
type Parser interface {
	Parse(src []byte) (Tree, error)
}

// Walk performs a deterministic pre-order traversal over named nodes, invoking
// fn for each node (including the root). Children are visited in source order,
// which is stable, preserving the §14 byte-determinism guarantee downstream.
func Walk(root Node, fn func(Node)) {
	if root == nil {
		return
	}
	fn(root)
	for _, c := range root.NamedChildren() {
		Walk(c, fn)
	}
}
