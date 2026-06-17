package parse

import (
	"testing"

	"github.com/amazopic/graffiti/internal/scan"
)

func TestSpecFor_CoversNewLanguages(t *testing.T) {
	for _, l := range []scan.Lang{
		scan.LangPython, scan.LangJavaScript, scan.LangTypeScript,
		scan.LangRust, scan.LangJava, scan.LangPHP,
	} {
		spec, ok := SpecFor(l)
		if !ok {
			t.Fatalf("%s: no spec", l)
		}
		if len(spec.ClassKinds) == 0 {
			t.Errorf("%s: no class kinds", l)
		}
		if len(spec.ImportKinds) == 0 {
			t.Errorf("%s: no import kinds", l)
		}
		if len(spec.CallKinds) == 0 {
			t.Errorf("%s: no call kinds", l)
		}
	}
}

func TestSpecFor_GoAndMarkdownAbsent(t *testing.T) {
	if _, ok := SpecFor(scan.LangGo); ok {
		t.Error("Go uses ParseGo, not the generic extractor; SpecFor(Go) must be absent")
	}
	if _, ok := SpecFor(scan.LangMarkdown); ok {
		t.Error("Markdown is not parsed")
	}
}

func TestSpecFor_JavaHasNoTopLevelFunctions(t *testing.T) {
	spec, _ := SpecFor(scan.LangJava)
	if len(spec.FuncKinds) != 0 {
		t.Errorf("Java has no top-level functions; FuncKinds = %v", spec.FuncKinds)
	}
}
