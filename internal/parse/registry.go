package parse

import (
	"fmt"

	ts "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"

	"github.com/amazopic/graffiti/internal/scan"
)

// langConstructors maps a scan.Lang to its gotreesitter language constructor.
// Markdown is intentionally absent (it is handled as a doc node, not parsed).
var langConstructors = map[scan.Lang]func() *ts.Language{
	scan.LangGo:         grammars.GoLanguage,
	scan.LangPython:     grammars.PythonLanguage,
	scan.LangJavaScript: grammars.JavascriptLanguage,
	scan.LangTypeScript: grammars.TypescriptLanguage,
	scan.LangRust:       grammars.RustLanguage,
	scan.LangJava:       grammars.JavaLanguage,
	scan.LangPHP:        grammars.PhpLanguage,
}

// NewParser returns a Parser for the given language. It returns an error for a
// language without a tree-sitter grammar (e.g. Markdown).
func NewParser(l scan.Lang) (Parser, error) {
	ctor, ok := langConstructors[l]
	if !ok {
		return nil, fmt.Errorf("parse: no tree-sitter parser for language %q", l)
	}
	lang := ctor()
	if lang == nil {
		return nil, fmt.Errorf("parse: grammar for %q is not embedded (missing build tag grammar_subset_%s?)", l, l)
	}
	return &gtsParser{lang: lang}, nil
}
