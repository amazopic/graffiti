package parse

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amazopic/graffiti/internal/graph"
	"github.com/amazopic/graffiti/internal/scan"
)

// countErrorNodes returns how many ERROR nodes the parse produced (a fidelity proxy).
func countErrorNodes(t *testing.T, l scan.Lang, src []byte) int {
	t.Helper()
	p, err := NewParser(l)
	if err != nil {
		t.Fatal(err)
	}
	tree, err := p.Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	lang := tree.Lang()
	var errs int
	Walk(tree.Root(), func(n Node) {
		if n.Type(lang) == "ERROR" {
			errs++
		}
	})
	return errs
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.FromSlash("../../testdata/fixtures/polyglot/" + name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestFidelity_NoErrorNodesAndExpectedDefs(t *testing.T) {
	cases := []struct {
		lang    scan.Lang
		file    string
		classes []string
		methods []string
		funcs   []string
	}{
		{scan.LangPython, "svc.py", []string{"Service"}, []string{"Service.handle"}, []string{"validate"}},
		{scan.LangJavaScript, "app.js", []string{"App"}, []string{"App.run"}, []string{"validate"}},
		{scan.LangTypeScript, "model.ts", []string{"User", "Model"}, []string{"Model.load"}, []string{"fetchUser"}},
		{scan.LangRust, "lib.rs", []string{"Store"}, []string{"Store.get"}, []string{"normalize"}},
		{scan.LangJava, "Greeter.java", []string{"Greeter"}, []string{"Greeter.greet", "Greeter.sanitize"}, nil},
		{scan.LangPHP, "helper.php", []string{"Helper"}, []string{"Helper.clean"}, []string{"normalize"}},
	}
	for _, c := range cases {
		t.Run(string(c.lang), func(t *testing.T) {
			src := readFixture(t, c.file)
			if n := countErrorNodes(t, c.lang, src); n != 0 {
				t.Fatalf("%s: %d ERROR nodes (fidelity gate: must be 0)", c.lang, n)
			}
			ex := extractLang(t, c.lang, c.file, string(src))
			for _, w := range c.classes {
				if !has(labels(ex, graph.KindClass), w) {
					t.Errorf("%s: missing class %q (got %v)", c.lang, w, labels(ex, graph.KindClass))
				}
			}
			for _, w := range c.methods {
				if !has(labels(ex, graph.KindMethod), w) {
					t.Errorf("%s: missing method %q (got %v)", c.lang, w, labels(ex, graph.KindMethod))
				}
			}
			for _, w := range c.funcs {
				if !has(labels(ex, graph.KindFunction), w) {
					t.Errorf("%s: missing function %q (got %v)", c.lang, w, labels(ex, graph.KindFunction))
				}
			}
			if len(labels(ex, graph.KindModule)) == 0 {
				t.Errorf("%s: expected an imported module node", c.lang)
			}
			if len(ex.RawCalls) == 0 {
				t.Errorf("%s: expected at least one collected call", c.lang)
			}
		})
	}
}
