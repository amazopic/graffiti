package parse

import (
	"testing"
)

const sampleGo = `package main

import "fmt"

func Hello(name string) string {
	return fmt.Sprintf("hi %s", name)
}

type Greeter struct{ Prefix string }

func (g Greeter) Greet(n string) string { return g.Prefix + Hello(n) }
`

// TestSpike_OneGoFileToAST is the literal "get ONE Go file to an AST" gate.
func TestSpike_OneGoFileToAST(t *testing.T) {
	p, err := NewGoParser()
	if err != nil {
		t.Fatalf("NewGoParser: %v", err)
	}
	tree, err := p.Parse([]byte(sampleGo))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	lang := tree.Lang()
	root := tree.Root()
	if root.Type(lang) != "source_file" {
		t.Fatalf("root kind = %q, want source_file", root.Type(lang))
	}

	type hit struct {
		kind string
		line int
	}
	var hits []hit
	Walk(root, func(n Node) {
		switch n.Type(lang) {
		case "import_spec", "function_declaration", "type_spec", "method_declaration":
			hits = append(hits, hit{n.Type(lang), int(n.StartPoint().Row) + 1})
		}
	})

	want := []hit{
		{"import_spec", 3},
		{"function_declaration", 5},
		{"type_spec", 9},
		{"method_declaration", 11},
	}
	if len(hits) != len(want) {
		t.Fatalf("hits = %+v, want %+v", hits, want)
	}
	for i := range want {
		if hits[i] != want[i] {
			t.Fatalf("hit[%d] = %+v, want %+v", i, hits[i], want[i])
		}
	}
}

func TestSpike_NodeTextAndFields(t *testing.T) {
	p, err := NewGoParser()
	if err != nil {
		t.Fatalf("NewGoParser: %v", err)
	}
	tree, err := p.Parse([]byte(sampleGo))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	lang := tree.Lang()
	var funcName, importPath string
	Walk(tree.Root(), func(n Node) {
		switch n.Type(lang) {
		case "function_declaration":
			if c := n.ChildByField("name"); c != nil {
				funcName = c.Text()
			}
		case "import_spec":
			if c := n.ChildByField("path"); c != nil {
				importPath = c.Text()
			}
		}
	})
	if funcName != "Hello" {
		t.Fatalf("func name = %q, want Hello", funcName)
	}
	if importPath != `"fmt"` {
		t.Fatalf("import path = %q, want %q", importPath, `"fmt"`)
	}
}
