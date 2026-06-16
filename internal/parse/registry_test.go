package parse

import (
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/scan"
)

func TestNewParser_AllLanguagesLoad(t *testing.T) {
	for _, l := range []scan.Lang{
		scan.LangGo, scan.LangPython, scan.LangJavaScript,
		scan.LangTypeScript, scan.LangRust, scan.LangJava, scan.LangPHP,
	} {
		p, err := NewParser(l)
		if err != nil {
			t.Fatalf("%s: %v", l, err)
		}
		tree, err := p.Parse([]byte(""))
		if err != nil {
			t.Fatalf("%s parse empty: %v", l, err)
		}
		if tree.Root() == nil {
			t.Fatalf("%s: nil root", l)
		}
	}
}

func TestNewParser_Unsupported(t *testing.T) {
	if _, err := NewParser(scan.LangMarkdown); err == nil {
		t.Fatal("markdown has no tree-sitter parser; expected an error")
	}
}
