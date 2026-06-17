package parse

import (
	"testing"

	"github.com/amazopic/graffiti/internal/graph"
)

const pass1Src = `package greet

import (
	"fmt"
	"strings"
)

func Hello(name string) string {
	return fmt.Sprintf("%s", strings.ToUpper(name))
}

type Greeter struct{ Prefix string }

func (g Greeter) Greet(n string) string {
	return g.Prefix + Hello(n)
}
`

func TestParseGo_Pass1_Nodes(t *testing.T) {
	p, err := NewGoParser()
	if err != nil {
		t.Fatal(err)
	}
	ex, err := ParseGo(p, "greet/greet.go", []byte(pass1Src))
	if err != nil {
		t.Fatalf("ParseGo: %v", err)
	}

	byID := map[string]graph.Node{}
	for _, n := range ex.Nodes {
		byID[n.ID] = n
	}

	fileID := graph.NodeID("greet/greet.go", "greet/greet.go")
	if fn, ok := byID[fileID]; !ok || fn.Kind != graph.KindFile {
		t.Fatalf("missing file node %q: %+v", fileID, byID)
	}

	helloID := graph.NodeID("greet/greet.go", "Hello")
	if n, ok := byID[helloID]; !ok || n.Kind != graph.KindFunction || n.Line != 8 || n.Label != "Hello" {
		t.Fatalf("Hello node wrong: ok=%v %+v", ok, byID[helloID])
	}

	greeterID := graph.NodeID("greet/greet.go", "Greeter")
	if n, ok := byID[greeterID]; !ok || n.Kind != graph.KindClass || n.Line != 12 {
		t.Fatalf("Greeter node wrong: ok=%v %+v", ok, byID[greeterID])
	}

	greetID := graph.NodeID("greet/greet.go", "Greeter.Greet")
	if n, ok := byID[greetID]; !ok || n.Kind != graph.KindMethod || n.Line != 14 || n.Label != "Greeter.Greet" {
		t.Fatalf("Greet method node wrong: ok=%v %+v", ok, byID[greetID])
	}
}

func TestParseGo_Pass1_ImportsAndContains(t *testing.T) {
	p, err := NewGoParser()
	if err != nil {
		t.Fatal(err)
	}
	ex, err := ParseGo(p, "greet/greet.go", []byte(pass1Src))
	if err != nil {
		t.Fatalf("ParseGo: %v", err)
	}

	fileID := graph.NodeID("greet/greet.go", "greet/greet.go")
	fmtModID := graph.NodeID("module:fmt", "fmt")

	var sawImportFmt, sawContainsHello bool
	for _, e := range ex.Edges {
		if e.From == fileID && e.To == fmtModID && e.Relation == graph.RelImports && e.Confidence == graph.ConfExtracted {
			sawImportFmt = true
		}
		if e.From == fileID && e.To == graph.NodeID("greet/greet.go", "Hello") && e.Relation == graph.RelContains {
			sawContainsHello = true
		}
	}
	if !sawImportFmt {
		t.Fatalf("missing imports edge file->fmt (EXTRACTED); edges=%+v", ex.Edges)
	}
	if !sawContainsHello {
		t.Fatalf("missing contains edge file->Hello")
	}

	var sawFmtMod bool
	for _, n := range ex.Nodes {
		if n.ID == fmtModID && n.Kind == graph.KindModule && n.Label == "fmt" {
			sawFmtMod = true
		}
	}
	if !sawFmtMod {
		t.Fatalf("missing module node for fmt")
	}
}

func TestParseGo_Pass1_RawCalls(t *testing.T) {
	p, err := NewGoParser()
	if err != nil {
		t.Fatal(err)
	}
	ex, err := ParseGo(p, "greet/greet.go", []byte(pass1Src))
	if err != nil {
		t.Fatalf("ParseGo: %v", err)
	}

	names := map[string]bool{}
	for _, rc := range ex.RawCalls {
		names[rc.Callee] = true
	}
	for _, want := range []string{"fmt.Sprintf", "strings.ToUpper", "Hello"} {
		if !names[want] {
			t.Fatalf("missing raw call %q; got %v", want, names)
		}
	}
	for _, rc := range ex.RawCalls {
		if rc.FromID == "" {
			t.Fatalf("raw call %q has empty FromID", rc.Callee)
		}
		if len(rc.Imports) == 0 {
			t.Fatalf("raw call %q has no Imports recorded", rc.Callee)
		}
	}
}
