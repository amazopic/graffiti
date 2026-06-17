package parse

import (
	"testing"

	"github.com/amazopic/graffiti/internal/graph"
	"github.com/amazopic/graffiti/internal/scan"
)

// extractLang is a small test helper: parse src in language l and extract.
func extractLang(t *testing.T, l scan.Lang, relPath, src string) *Extraction {
	t.Helper()
	p, err := NewParser(l)
	if err != nil {
		t.Fatal(err)
	}
	spec, ok := SpecFor(l)
	if !ok {
		t.Fatalf("no spec for %s", l)
	}
	ex, err := Extract(p, relPath, []byte(src), spec)
	if err != nil {
		t.Fatal(err)
	}
	return ex
}

func labels(ex *Extraction, kind graph.Kind) []string {
	var out []string
	for _, n := range ex.Nodes {
		if n.Kind == kind {
			out = append(out, n.Label)
		}
	}
	return out
}

func has(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

func TestExtract_PythonClassMethodAndFunction(t *testing.T) {
	src := "import os\nfrom a.b import c\n\nclass Greeter:\n    def greet(self, name):\n        return say_hello(name)\n\ndef say_hello(n):\n    return n\n"
	ex := extractLang(t, scan.LangPython, "g.py", src)
	if !has(labels(ex, graph.KindClass), "Greeter") {
		t.Errorf("missing class Greeter: %v", labels(ex, graph.KindClass))
	}
	if !has(labels(ex, graph.KindMethod), "Greeter.greet") {
		t.Errorf("missing method Greeter.greet: %v", labels(ex, graph.KindMethod))
	}
	if !has(labels(ex, graph.KindFunction), "say_hello") {
		t.Errorf("missing function say_hello: %v", labels(ex, graph.KindFunction))
	}
	// imports → module nodes
	if len(labels(ex, graph.KindModule)) == 0 {
		t.Error("expected at least one imported module node")
	}
	// the call inside greet was collected
	if len(ex.RawCalls) == 0 {
		t.Error("expected raw calls collected from greet()")
	}
}

func TestExtract_RustImplMethods(t *testing.T) {
	src := "use std::fmt;\n\npub struct Greeter { n: i32 }\n\nimpl Greeter {\n    pub fn greet(&self) { helper(); }\n}\n\nfn helper() {}\n"
	ex := extractLang(t, scan.LangRust, "g.rs", src)
	if !has(labels(ex, graph.KindClass), "Greeter") {
		t.Errorf("missing struct Greeter: %v", labels(ex, graph.KindClass))
	}
	if !has(labels(ex, graph.KindMethod), "Greeter.greet") {
		t.Errorf("missing impl method Greeter.greet: %v", labels(ex, graph.KindMethod))
	}
	if !has(labels(ex, graph.KindFunction), "helper") {
		t.Errorf("missing top-level fn helper: %v", labels(ex, graph.KindFunction))
	}
}

func TestExtract_JavaMethodsNoTopLevelFunc(t *testing.T) {
	src := "package p;\nimport java.util.List;\n\npublic class Greeter {\n    public String greet(String n) { return sayHi(n); }\n    private String sayHi(String n) { return n; }\n}\n"
	ex := extractLang(t, scan.LangJava, "G.java", src)
	if !has(labels(ex, graph.KindClass), "Greeter") {
		t.Errorf("missing class Greeter: %v", labels(ex, graph.KindClass))
	}
	if !has(labels(ex, graph.KindMethod), "Greeter.greet") || !has(labels(ex, graph.KindMethod), "Greeter.sayHi") {
		t.Errorf("missing Java methods: %v", labels(ex, graph.KindMethod))
	}
	if len(labels(ex, graph.KindFunction)) != 0 {
		t.Errorf("Java should have no top-level functions: %v", labels(ex, graph.KindFunction))
	}
}

func TestExtract_EveryDefHasContainsEdge(t *testing.T) {
	ex := extractLang(t, scan.LangPython, "g.py", "class A:\n    def m(self): pass\ndef f(): pass\n")
	fileID := graph.NodeID("g.py", "g.py")
	defs := 0
	for _, n := range ex.Nodes {
		if n.Kind == graph.KindClass || n.Kind == graph.KindMethod || n.Kind == graph.KindFunction {
			defs++
		}
	}
	contains := 0
	for _, e := range ex.Edges {
		if e.Relation == graph.RelContains && e.From == fileID {
			contains++
		}
	}
	if contains != defs {
		t.Errorf("contains edges = %d, defs = %d (each def needs a contains edge)", contains, defs)
	}
}
