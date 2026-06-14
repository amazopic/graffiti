package parse

import (
	ts "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
)

// gtsParser is the gotreesitter-backed Parser. The library is pure-Go (no CGO,
// no WASM) and shipped with the grammar_subset+grammar_subset_go build tags.
type gtsParser struct {
	lang *ts.Language
}

// NewGoParser returns a Parser for the Go language.
func NewGoParser() (Parser, error) {
	return &gtsParser{lang: grammars.GoLanguage()}, nil
}

func (p *gtsParser) Parse(src []byte) (Tree, error) {
	parser := ts.NewParser(p.lang)
	tree, err := parser.Parse(src)
	if err != nil {
		return nil, err
	}
	return &gtsTree{tree: tree, lang: p.lang, src: src}, nil
}

type gtsTree struct {
	tree *ts.Tree
	lang *ts.Language
	src  []byte
}

func (t *gtsTree) Root() Node {
	return wrap(t.tree.RootNode(), t.lang, t.src)
}

func (t *gtsTree) Lang() *ts.Language { return t.lang }

type gtsNode struct {
	n    *ts.Node
	lang *ts.Language
	src  []byte
}

func wrap(n *ts.Node, lang *ts.Language, src []byte) Node {
	if n == nil {
		return nil
	}
	return &gtsNode{n: n, lang: lang, src: src}
}

func (g *gtsNode) Type(lang *ts.Language) string { return g.n.Type(lang) }
func (g *gtsNode) StartPoint() ts.Point          { return g.n.StartPoint() }
func (g *gtsNode) StartByte() int                { return int(g.n.StartByte()) }
func (g *gtsNode) EndByte() int                  { return int(g.n.EndByte()) }
func (g *gtsNode) Text() string                  { return g.n.Text(g.src) }

func (g *gtsNode) ChildByField(field string) Node {
	c := g.n.ChildByFieldName(field, g.lang)
	if c == nil {
		return nil
	}
	return wrap(c, g.lang, g.src)
}

func (g *gtsNode) NamedChildren() []Node {
	cnt := g.n.NamedChildCount()
	out := make([]Node, 0, cnt)
	for i := 0; i < cnt; i++ {
		out = append(out, wrap(g.n.NamedChild(i), g.lang, g.src))
	}
	return out
}
