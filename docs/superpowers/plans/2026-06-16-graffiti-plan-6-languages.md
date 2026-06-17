# graffiti Plan 6 — Additional languages (Python, JS, TS, Rust, Java, PHP)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend graffiti beyond Go to **Python, JavaScript, TypeScript, Rust, Java, and PHP**, so `graffiti build .` produces a useful directed graph (files, definitions, methods, imports, intra-repo call edges) for polyglot repos — the languages most vibe-coder repos actually use.

**Architecture:** A **table-driven extractor**. The existing Go path (`parse.ParseGo`, with its bespoke receiver handling) stays untouched. New languages are described by a small `LangSpec` (the tree-sitter node-kind/field vocabulary, empirically verified against `gotreesitter@v0.20.2`) and processed by one generic `parse.Extract` walk that emits: a file node, definition nodes (functions, classes/structs/interfaces/enums/traits → `KindClass`, methods labeled `Class.method`), `contains` edges, `imports` edges to synthesized module nodes, and raw call sites. Pass-2 `parse.ResolveCalls` is **already language-agnostic** (it resolves bare callees against same-repo definitions and drops unmatched/ambiguous ones), so it is reused unchanged — non-Go selector calls simply under-resolve, which is the correct honesty-first behavior. `app.Build` gains a lazy per-language parser cache and routes each file to Go / Markdown / generic.

**Tech Stack:** Go 1.26, pure-Go `gotreesitter@v0.20.2` (already a dep; six new grammars are gated behind `grammar_subset_<lang>` build tags). **No new dependencies.** The subset binary grows from ~9.6 MB to ~10.3 MB — well under the 16 MB size guard (measured in the feasibility spike).

---

## Feasibility spike results (2026-06-16, verified empirically — not assumed)

A throwaway spike parsed representative snippets in all six languages with the embedded grammars:

- **Availability:** `grammars.PythonLanguage()`, `JavascriptLanguage()`, `TypescriptLanguage()`, `RustLanguage()`, `JavaLanguage()`, `PhpLanguage()` all exist; each grammar ships behind `grammar_subset_<lang>`.
- **Fidelity:** **0 ERROR nodes** on representative snippets for every language; clean trees. (PHP, Python, Java additionally have upstream `*_parse_regression_test.go` suites inside gotreesitter.)
- **Size:** six-language subset binary = **10.3 MB** (< 16 MB guard); each grammar adds only ~70–130 KB.
- **`ChildByField("name")` returns the definition name for every class/function/method kind in all six languages.**

**Verified node vocabulary (the source of truth for Task 3's `LangSpec` table):**

| Lang | functions (top-level) | classes → KindClass | methods (in class) | impl | imports | call kinds → callee field |
|------|----|----|----|----|----|----|
| python | `function_definition` | `class_definition` | `function_definition` | — | `import_statement`, `import_from_statement` | `call`→`function` |
| javascript | `function_declaration` | `class_declaration` | `method_definition` | — | `import_statement` (module = `string` child) | `call_expression`→`function` |
| typescript | `function_declaration` | `class_declaration`, `interface_declaration` | `method_definition` | — | `import_statement` (module = `string` child) | `call_expression`→`function` |
| rust | `function_item` | `struct_item`, `enum_item`, `trait_item` | `function_item` | `impl_item` (qualifier = `type_identifier` child; body = `declaration_list`) | `use_declaration` | `call_expression`→`function` |
| java | *(none)* | `class_declaration`, `interface_declaration`, `enum_declaration` | `method_declaration` | — | `import_declaration` | `method_invocation`→`name` |
| php | `function_definition` | `class_declaration`, `interface_declaration`, `trait_declaration` | `method_declaration` | — | `namespace_use_declaration` | `function_call_expression`→`function`, `scoped_call_expression`→`name`, `member_call_expression`→`name` |

**Fidelity-gate honesty (deviation from the original memory mandate):** the mandate was "per-language parity tests vs upstream tree-sitter before freezing goldens." Running the upstream (C) tree-sitter requires a CLI/CGO toolchain not available in this offline pure-Go environment. Plan 6 therefore uses a **layered offline fidelity gate** instead: (a) gotreesitter's own upstream regression suites; (b) a committed **zero-ERROR-node assertion** per language on the fixtures; (c) **structural-shape assertions** that each language's expected functions/classes/methods/imports/one call edge are extracted; (d) golden determinism. True upstream-diff parity is **deferred to CI** (Plan 8), where a `tree-sitter` CLI can be installed. This limitation is stated, not hidden.

## File structure

```
internal/scan/scan.go            + extensions/Lang constants for 6 languages
internal/parse/registry.go       NewParser(scan.Lang) → Parser (lazy *ts.Language per lang)   [new]
internal/parse/langspec.go       LangSpec type + SpecFor(scan.Lang) table (6 specs)            [new]
internal/parse/extract.go        Extract(p, relPath, src, spec) generic table-driven walk       [new]
internal/app/app.go              route files: Go→ParseGo, Markdown→doc, others→Extract
Makefile                         TAGS += the 6 grammar_subset_<lang> tags
testdata/fixtures/polyglot/      one small file per language (py/js/ts/rs/java/php)             [new]
testdata/golden/polyglot.map.json + .MAP.md                                                    [new]
README.md                        supported-languages list
docs/superpowers/specs/2026-06-14-graffiti-design.md   §4 add PHP + architecture/parity note
```

Helpers reused from existing `internal/parse` (do **not** duplicate): `Walk`, `emitDef`, `fieldText`, `importBase`. `Extract` lives in the same package, so these are directly callable.

---

## Task 1: scan — extensions + Lang constants

**Files:**
- Modify: `internal/scan/scan.go`
- Test: `internal/scan/scan_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/scan/scan_test.go`:

```go
func TestScan_ClassifiesNewLanguages(t *testing.T) {
	dir := t.TempDir()
	files := map[string]Lang{
		"a.py":   LangPython,
		"b.js":   LangJavaScript,
		"c.jsx":  LangJavaScript,
		"d.mjs":  LangJavaScript,
		"e.ts":   LangTypeScript,
		"f.tsx":  LangTypeScript,
		"g.rs":   LangRust,
		"h.java": LangJava,
		"i.php":  LangPHP,
	}
	for name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	refs, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]Lang{}
	for _, r := range refs {
		got[r.RelPath] = r.Lang
	}
	for name, want := range files {
		if got[name] != want {
			t.Errorf("%s: lang = %q, want %q", name, got[name], want)
		}
	}
}
```

(Requires imports `os`, `path/filepath` in the test file — already present.)

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./internal/scan/ -run NewLanguages -v`
Expected: FAIL — undefined `LangPython` etc.

- [ ] **Step 3: Add the constants and extension map entries**

In `internal/scan/scan.go`, extend the `const` block:

```go
const (
	LangGo         Lang = "go"
	LangMarkdown   Lang = "markdown"
	LangPython     Lang = "python"
	LangJavaScript Lang = "javascript"
	LangTypeScript Lang = "typescript"
	LangRust       Lang = "rust"
	LangJava       Lang = "java"
	LangPHP        Lang = "php"
)
```

And extend `extLang`:

```go
var extLang = map[string]Lang{
	".go":   LangGo,
	".md":   LangMarkdown,
	".py":   LangPython,
	".js":   LangJavaScript,
	".jsx":  LangJavaScript,
	".mjs":  LangJavaScript,
	".cjs":  LangJavaScript,
	".ts":   LangTypeScript,
	".tsx":  LangTypeScript,
	".rs":   LangRust,
	".java": LangJava,
	".php":  LangPHP,
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/scan/ -run NewLanguages -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/scan/scan.go internal/scan/scan_test.go
git commit -m "feat(scan): classify python/js/ts/rust/java/php source files"
```

---

## Task 2: parse — multi-language parser registry + Makefile tags

**Files:**
- Create: `internal/parse/registry.go`
- Modify: `Makefile`
- Test: `internal/parse/registry_test.go`

- [ ] **Step 1: Add the grammar build tags to the Makefile**

In `Makefile`, replace the `TAGS :=` line with:

```makefile
TAGS := grammar_subset grammar_subset_go grammar_subset_gomod \
        grammar_subset_python grammar_subset_javascript grammar_subset_typescript \
        grammar_subset_rust grammar_subset_java grammar_subset_php
```

(Without these tags the subset build does not embed the new grammars and their parsers return nil. The default no-tag `go test ./...` embeds **all** grammars, so it is unaffected. Binary stays ~10.3 MB — verified.)

- [ ] **Step 2: Write the failing test**

`internal/parse/registry_test.go`:

```go
package parse

import (
	"testing"

	"github.com/amazopic/graffiti/internal/scan"
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
```

- [ ] **Step 3: Run it to verify it fails**

Run: `go test ./internal/parse/ -run NewParser -v`
Expected: FAIL — `NewParser` undefined.

- [ ] **Step 4: Write `registry.go`**

```go
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
```

(`NewGoParser` in `gotreesitter.go` remains for backward compatibility; it is equivalent to `NewParser(scan.LangGo)`.)

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/parse/ -run NewParser -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/parse/registry.go internal/parse/registry_test.go Makefile
git commit -m "feat(parse): NewParser language registry; add grammar build tags"
```

---

## Task 3: parse — `LangSpec` table

**Files:**
- Create: `internal/parse/langspec.go`
- Test: `internal/parse/langspec_test.go`

- [ ] **Step 1: Write the failing test**

`internal/parse/langspec_test.go`:

```go
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
```

- [ ] **Step 2: Run it to verify it fails**

Run: `go test ./internal/parse/ -run SpecFor -v`
Expected: FAIL — `SpecFor`/`LangSpec` undefined.

- [ ] **Step 3: Write `langspec.go`**

```go
package parse

import "github.com/amazopic/graffiti/internal/scan"

// LangSpec describes a language's tree-sitter node vocabulary for the generic
// extractor (Extract). Every definition kind exposes its name via the grammar
// field "name" (verified across all six languages in the Plan 6 feasibility spike).
type LangSpec struct {
	FuncKinds   []string          // top-level function definition kinds
	ClassKinds  []string          // class/struct/interface/enum/trait kinds → KindClass
	MethodKinds []string          // method definition kinds found inside a class body
	ImplKinds   []string          // impl blocks (Rust): qualifier from a type_identifier child
	ImportKinds []string          // import / use / namespace-use statement kinds
	ImportChild string           // child node kind holding the module string (JS/TS: "string"); "" = first named child
	CallKinds   map[string]string // call node kind -> grammar field holding the callee text
}

// SpecFor returns the extractor spec for a language, or ok=false for languages
// not handled by the generic extractor (Go uses ParseGo; Markdown is not parsed).
func SpecFor(l scan.Lang) (LangSpec, bool) {
	switch l {
	case scan.LangPython:
		return LangSpec{
			FuncKinds:   []string{"function_definition"},
			ClassKinds:  []string{"class_definition"},
			MethodKinds: []string{"function_definition"},
			ImportKinds: []string{"import_statement", "import_from_statement"},
			CallKinds:   map[string]string{"call": "function"},
		}, true
	case scan.LangJavaScript:
		return LangSpec{
			FuncKinds:   []string{"function_declaration"},
			ClassKinds:  []string{"class_declaration"},
			MethodKinds: []string{"method_definition"},
			ImportKinds: []string{"import_statement"},
			ImportChild: "string",
			CallKinds:   map[string]string{"call_expression": "function"},
		}, true
	case scan.LangTypeScript:
		return LangSpec{
			FuncKinds:   []string{"function_declaration"},
			ClassKinds:  []string{"class_declaration", "interface_declaration"},
			MethodKinds: []string{"method_definition"},
			ImportKinds: []string{"import_statement"},
			ImportChild: "string",
			CallKinds:   map[string]string{"call_expression": "function"},
		}, true
	case scan.LangRust:
		return LangSpec{
			FuncKinds:   []string{"function_item"},
			ClassKinds:  []string{"struct_item", "enum_item", "trait_item"},
			MethodKinds: []string{"function_item"},
			ImplKinds:   []string{"impl_item"},
			ImportKinds: []string{"use_declaration"},
			CallKinds:   map[string]string{"call_expression": "function"},
		}, true
	case scan.LangJava:
		return LangSpec{
			FuncKinds:   nil,
			ClassKinds:  []string{"class_declaration", "interface_declaration", "enum_declaration"},
			MethodKinds: []string{"method_declaration"},
			ImportKinds: []string{"import_declaration"},
			CallKinds:   map[string]string{"method_invocation": "name"},
		}, true
	case scan.LangPHP:
		return LangSpec{
			FuncKinds:   []string{"function_definition"},
			ClassKinds:  []string{"class_declaration", "interface_declaration", "trait_declaration"},
			MethodKinds: []string{"method_declaration"},
			ImportKinds: []string{"namespace_use_declaration"},
			CallKinds: map[string]string{
				"function_call_expression": "function",
				"scoped_call_expression":   "name",
				"member_call_expression":   "name",
			},
		}, true
	default:
		return LangSpec{}, false
	}
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/parse/ -run SpecFor -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/parse/langspec.go internal/parse/langspec_test.go
git commit -m "feat(parse): per-language node-kind LangSpec table"
```

---

## Task 4: parse — generic table-driven extractor

**Files:**
- Create: `internal/parse/extract.go`
- Test: `internal/parse/extract_test.go`

- [ ] **Step 1: Write the failing tests**

`internal/parse/extract_test.go`:

```go
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
```

- [ ] **Step 2: Run them to verify they fail**

Run: `go test ./internal/parse/ -run Extract -v`
Expected: FAIL — `Extract` undefined.

- [ ] **Step 3: Write `extract.go`**

```go
package parse

import (
	"strings"

	ts "github.com/odvcencio/gotreesitter"

	"github.com/amazopic/graffiti/internal/graph"
)

// Extract runs Pass 1 over one file using a table-driven LangSpec (spec §5). It
// emits a file node, definition nodes (functions, classes/structs/interfaces/
// enums/traits, and methods labeled "Class.method"), contains edges, imports
// edges to synthesized module nodes, and stashes raw call sites for Pass 2.
//
// Compared with ParseGo it deliberately UNDER-extracts (no Go-style receiver
// typing, no nested-definition recursion past a function body): honesty-first,
// per the §16 doctrine. The shared helpers emitDef/fieldText/Walk/importBase are
// reused from the package.
func Extract(p Parser, relPath string, src []byte, spec LangSpec) (*Extraction, error) {
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

	classSet := toSet(spec.ClassKinds)
	methodSet := toSet(spec.MethodKinds)
	funcSet := toSet(spec.FuncKinds)
	implSet := toSet(spec.ImplKinds)
	importSet := toSet(spec.ImportKinds)

	// 1) Imports (flat walk): emit a module node (keyed by raw import text) + edge.
	var importPaths []string
	Walk(root, func(n Node) {
		if !importSet[n.Type(lang)] {
			return
		}
		imp := importText(n, spec)
		if imp == "" {
			return
		}
		importPaths = append(importPaths, imp)
		label := lastSegment(imp)
		modID := graph.NodeID("module:"+imp, label)
		ex.Nodes = append(ex.Nodes, graph.Node{
			ID: modID, Label: label, Kind: graph.KindModule, File: relPath,
			Line: int(n.StartPoint().Row) + 1, Community: graph.UnclusteredCommunity,
		})
		ex.Edges = append(ex.Edges, graph.Edge{
			From: fileID, To: modID, Relation: graph.RelImports, Confidence: graph.ConfExtracted,
		})
	})

	// 2) Definitions: structured recursion so methods carry their class qualifier.
	var visit func(n Node, qualifier string)
	visit = func(n Node, qualifier string) {
		k := n.Type(lang)
		switch {
		case classSet[k]:
			name := fieldText(n, "name")
			if name != "" {
				defID := graph.NodeID(relPath, name)
				emitDef(ex, fileID, defID, name, graph.KindClass, relPath, line(n))
				for _, c := range n.NamedChildren() {
					visit(c, name)
				}
				return
			}
		case implSet[k]:
			// Rust impl: qualifier is the implemented type's name.
			q := namedChildText(n, "type_identifier", lang)
			for _, c := range n.NamedChildren() {
				visit(c, q)
			}
			return
		case methodSet[k] && qualifier != "":
			name := fieldText(n, "name")
			if name != "" {
				label := qualifier + "." + name
				defID := graph.NodeID(relPath, label)
				emitDef(ex, fileID, defID, label, graph.KindMethod, relPath, line(n))
				collectCallsSpec(ex, n, defID, relPath, importPaths, lang, spec)
				return
			}
		case funcSet[k] && qualifier == "":
			name := fieldText(n, "name")
			if name != "" {
				defID := graph.NodeID(relPath, name)
				emitDef(ex, fileID, defID, name, graph.KindFunction, relPath, line(n))
				collectCallsSpec(ex, n, defID, relPath, importPaths, lang, spec)
				return
			}
		}
		for _, c := range n.NamedChildren() {
			visit(c, qualifier)
		}
	}
	visit(root, "")

	return ex, nil
}

// collectCallsSpec walks a definition subtree and stashes each call site's callee
// text (extracted from the grammar field named by spec.CallKinds[kind]).
func collectCallsSpec(ex *Extraction, defNode Node, defID, file string, importPaths []string, lang *ts.Language, spec LangSpec) {
	Walk(defNode, func(n Node) {
		field, ok := spec.CallKinds[n.Type(lang)]
		if !ok {
			return
		}
		c := n.ChildByField(field)
		if c == nil {
			return
		}
		callee := strings.TrimSpace(c.Text())
		if callee == "" {
			return
		}
		ex.RawCalls = append(ex.RawCalls, RawCall{
			FromID: defID, Callee: callee, Line: int(n.StartPoint().Row) + 1, File: file,
			Imports: importPaths,
		})
	})
}

func line(n Node) int { return int(n.StartPoint().Row) + 1 }

func toSet(xs []string) map[string]bool {
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}
	return m
}

// namedChildText returns the text of the first named child of the given kind.
func namedChildText(n Node, kind string, lang *ts.Language) string {
	for _, c := range n.NamedChildren() {
		if c.Type(lang) == kind {
			return strings.TrimSpace(c.Text())
		}
	}
	return ""
}

// importText extracts a best-effort module path from an import node. JS/TS keep
// the module string in a dedicated child (spec.ImportChild); everything else uses
// the first named child (a dotted/scoped/namespace name).
func importText(n Node, spec LangSpec) string {
	if spec.ImportChild != "" {
		// find a descendant of the requested kind (the module string)
		var found string
		Walk(n, func(m Node) {
			if found != "" {
				return
			}
			// note: Type needs lang; importText has no lang, so match by quotes instead.
			t := strings.TrimSpace(m.Text())
			if len(t) >= 2 && (t[0] == '\'' || t[0] == '"' || t[0] == '`') {
				found = stripQuotes(t)
			}
		})
		if found != "" {
			return found
		}
	}
	kids := n.NamedChildren()
	if len(kids) == 0 {
		return ""
	}
	return strings.TrimSpace(kids[0].Text())
}

func stripQuotes(s string) string {
	if len(s) >= 2 {
		switch s[0] {
		case '\'', '"', '`':
			if s[len(s)-1] == s[0] {
				return s[1 : len(s)-1]
			}
		}
	}
	return s
}

// lastSegment returns the final component of a module path, splitting on the
// separators used across our languages: "/", ".", "::", "\".
func lastSegment(imp string) string {
	imp = strings.TrimSpace(imp)
	for _, sep := range []string{"::", "\\", "/", "."} {
		if i := strings.LastIndex(imp, sep); i >= 0 {
			imp = imp[i+len(sep):]
		}
	}
	return imp
}
```

Note on `importText`: it has no `*ts.Language` available, so for the JS/TS module-string case it matches the first quote-delimited descendant rather than checking node kind. This is safe — the only quoted token in an `import_statement` is the module specifier.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/parse/ -run Extract -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/parse/extract.go internal/parse/extract_test.go
git commit -m "feat(parse): generic table-driven multi-language extractor"
```

---

## Task 5: app — route files through the right extractor

**Files:**
- Modify: `internal/app/app.go`
- Test: `internal/app/app_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/app/app_test.go`:

```go
func TestBuild_PolyglotRepo(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"svc.py":  "def handler():\n    return helper()\ndef helper():\n    return 1\n",
		"app.js":  "export function main(){ return util(); }\nfunction util(){ return 1; }\n",
		"Main.java": "public class Main {\n  public static void main(String[] a){ run(); }\n  static void run(){}\n}\n",
		"lib.rs":  "pub fn build(){ helper(); }\nfn helper(){}\n",
	}
	for name, src := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	stats, err := Build(dir, "2026-06-16T00:00:00Z")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if stats.Files != 4 {
		t.Errorf("files = %d, want 4", stats.Files)
	}
	// Each language should have contributed nodes (functions/methods).
	if stats.Nodes < 8 {
		t.Errorf("expected at least 8 nodes across 4 languages, got %d", stats.Nodes)
	}
}
```

(`os`/`path/filepath` are already imported in app_test.go.)

- [ ] **Step 2: Run it to verify it fails**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod grammar_subset_python grammar_subset_javascript grammar_subset_typescript grammar_subset_rust grammar_subset_java grammar_subset_php" ./internal/app/ -run Polyglot -v`
Expected: FAIL — non-Go files are currently ignored, so node count is too low / files miscounted.

- [ ] **Step 3: Update `app.go` to route by language**

In `internal/app/app.go`, replace the per-file `switch ref.Lang` block (the `case scan.LangGo` / `case scan.LangMarkdown` switch) with:

```go
		switch ref.Lang {
		case scan.LangGo:
			ex, perr := parse.ParseGo(goParser, ref.RelPath, src)
			if perr != nil {
				return stats, perr
			}
			extractions = append(extractions, ex)
		case scan.LangMarkdown:
			extractions = append(extractions, markdownExtraction(ref.RelPath))
			stats.HasDocNode = true
		default:
			spec, ok := parse.SpecFor(ref.Lang)
			if !ok {
				continue // unsupported language reached scan but has no extractor; skip
			}
			p, perr := parserFor(parsers, ref.Lang)
			if perr != nil {
				return stats, perr
			}
			ex, perr := parse.Extract(p, ref.RelPath, src, spec)
			if perr != nil {
				return stats, perr
			}
			extractions = append(extractions, ex)
		}
```

Add a lazy parser cache. Just before the `for _, ref := range refs {` loop, add:

```go
	parsers := map[scan.Lang]parse.Parser{}
```

And add this helper at the end of `app.go`:

```go
// parserFor lazily constructs and caches a parser per language so each grammar
// blob is loaded at most once per build.
func parserFor(cache map[scan.Lang]parse.Parser, l scan.Lang) (parse.Parser, error) {
	if p, ok := cache[l]; ok {
		return p, nil
	}
	p, err := parse.NewParser(l)
	if err != nil {
		return nil, err
	}
	cache[l] = p
	return p, nil
}
```

Ensure `internal/app/app.go` imports `"github.com/amazopic/graffiti/internal/scan"` (it already does for `scan.LangGo`).

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod grammar_subset_python grammar_subset_javascript grammar_subset_typescript grammar_subset_rust grammar_subset_java grammar_subset_php" ./internal/app/ -run Polyglot -v`
Expected: PASS

- [ ] **Step 5: Confirm the existing Go golden still passes (no regression)**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod grammar_subset_python grammar_subset_javascript grammar_subset_typescript grammar_subset_rust grammar_subset_java grammar_subset_php" ./internal/app/ -v`
Expected: PASS (the gorepo golden is unchanged — Go routing is untouched).

- [ ] **Step 6: Commit**

```bash
git add internal/app/app.go internal/app/app_test.go
git commit -m "feat(app): route python/js/ts/rust/java/php files through the generic extractor"
```

---

## Task 6: per-language fixtures + fidelity tests

This is the offline fidelity gate (see "Fidelity-gate honesty" above): zero-ERROR-node + structural-shape assertions per language.

**Files:**
- Create: `testdata/fixtures/polyglot/svc.py`, `app.js`, `model.ts`, `lib.rs`, `Greeter.java`, `helper.php`
- Create: `internal/parse/fidelity_test.go`

- [ ] **Step 1: Create the fixtures**

`testdata/fixtures/polyglot/svc.py`:
```python
import os
from auth.session import Session

class Service:
    def handle(self, req):
        return validate(req)

def validate(req):
    return os.path.exists(req)
```

`testdata/fixtures/polyglot/app.js`:
```javascript
import { Session } from './auth/session.js';

export class App {
  run(req) {
    return validate(req);
  }
}

function validate(req) {
  return Boolean(req);
}
```

`testdata/fixtures/polyglot/model.ts`:
```typescript
import { Session } from './auth/session';

export interface User { id: number }

export class Model {
  load(id: number): User {
    return fetchUser(id);
  }
}

function fetchUser(id: number): User {
  return { id };
}
```

`testdata/fixtures/polyglot/lib.rs`:
```rust
use std::collections::HashMap;

pub struct Store { data: HashMap<String, String> }

impl Store {
    pub fn get(&self, k: &str) -> String {
        normalize(k)
    }
}

fn normalize(k: &str) -> String {
    k.to_string()
}
```

`testdata/fixtures/polyglot/Greeter.java`:
```java
package com.example;

import java.util.List;

public class Greeter {
    public String greet(String name) {
        return sanitize(name);
    }

    private String sanitize(String n) {
        return n.trim();
    }
}
```

`testdata/fixtures/polyglot/helper.php`:
```php
<?php
namespace App;

use App\Support\Str;

class Helper {
    public function clean($value) {
        return normalize($value);
    }
}

function normalize($v) {
    return Str::lower($v);
}
```

- [ ] **Step 2: Write the fidelity test**

`internal/parse/fidelity_test.go`:

```go
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
			// at least one import module node + one collected call
			if len(labels(ex, graph.KindModule)) == 0 {
				t.Errorf("%s: expected an imported module node", c.lang)
			}
			if len(ex.RawCalls) == 0 {
				t.Errorf("%s: expected at least one collected call", c.lang)
			}
		})
	}
}
```

- [ ] **Step 3: Run the fidelity test**

Run: `go test ./internal/parse/ -run Fidelity -v`
Expected: PASS — 0 ERROR nodes per language; all expected defs present.

- [ ] **Step 4: Commit**

```bash
git add testdata/fixtures/polyglot internal/parse/fidelity_test.go
git commit -m "test(parse): per-language fidelity fixtures (0 ERROR nodes + expected defs)"
```

---

## Task 7: polyglot golden

**Files:**
- Create: `testdata/golden/polyglot.map.json`
- Create: `internal/app/polyglot_golden_test.go`

- [ ] **Step 1: Write the golden test (build the fixtures dir, compare map.json)**

`internal/app/polyglot_golden_test.go`:

```go
package app

import (
	"os"
	"path/filepath"
	"testing"
)

// buildPolyglotInto copies the committed polyglot fixtures into a temp dir and
// builds, returning the produced map.json bytes (root is stable: the temp base).
func TestPolyglot_Golden(t *testing.T) {
	srcDir := filepath.FromSlash("../../testdata/fixtures/polyglot")
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for _, e := range entries {
		b, err := os.ReadFile(filepath.Join(srcDir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, e.Name()), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Build(dir, "2026-06-16T00:00:00Z"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, ".graffiti", "map.json"))
	if err != nil {
		t.Fatal(err)
	}

	goldenPath := filepath.FromSlash("../../testdata/golden/polyglot.map.json")
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden (generate it per the plan's step): %v", err)
	}
	// map.json root is filepath.Base(dir) (a temp name) → normalize before compare.
	gs := normalizeRoot(t, got)
	ws := normalizeRoot(t, want)
	if gs != ws {
		t.Fatalf("polyglot map.json differs from golden.\n--- got ---\n%s", gs)
	}

	// Determinism: a second build of the same inputs is byte-identical.
	dir2 := t.TempDir()
	for _, e := range entries {
		b, _ := os.ReadFile(filepath.Join(srcDir, e.Name()))
		_ = os.WriteFile(filepath.Join(dir2, e.Name()), b, 0o644)
	}
	if _, err := Build(dir2, "2026-06-16T00:00:00Z"); err != nil {
		t.Fatal(err)
	}
	got2, _ := os.ReadFile(filepath.Join(dir2, ".graffiti", "map.json"))
	if normalizeRoot(t, got) != normalizeRoot(t, got2) {
		t.Fatal("non-deterministic: two builds of the same inputs differ")
	}
}
```

Add the `normalizeRoot` helper (replaces the `"root":"<temp>"` value with a constant so the golden is location-independent — mirrors how the gorepo golden is handled):

```go
import (
	"encoding/json"
)

func normalizeRoot(t *testing.T, b []byte) string {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal map.json: %v", err)
	}
	m["root"] = "polyglot"
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}
```

(Combine the two `import` blocks into one when writing the file.)

- [ ] **Step 2: Generate the golden from the tool**

```bash
make build
TMP=$(mktemp -d)
cp testdata/fixtures/polyglot/* "$TMP"/
./graffiti build "$TMP" >/dev/null
cp "$TMP/.graffiti/map.json" testdata/golden/polyglot.map.json
rm -rf "$TMP"
```

Then **read** `testdata/golden/polyglot.map.json` and sanity-check it by eye: it should contain file nodes for all six fixtures, class/method/function nodes (e.g. `Service`, `Service.handle`, `App.run`, `Store.get`, `Greeter.greet`, `Helper.clean`), module nodes from imports, `contains`/`imports` edges, and any resolved intra-file `calls` edges (e.g. `validate`, `normalize`).

- [ ] **Step 3: Run the golden test**

Run: `go test -tags "grammar_subset grammar_subset_go grammar_subset_gomod grammar_subset_python grammar_subset_javascript grammar_subset_typescript grammar_subset_rust grammar_subset_java grammar_subset_php" ./internal/app/ -run Polyglot_Golden -v`
Expected: PASS (golden matches; determinism holds).

- [ ] **Step 4: Commit**

```bash
git add testdata/golden/polyglot.map.json internal/app/polyglot_golden_test.go
git commit -m "test(app): byte-deterministic polyglot map.json golden"
```

---

## Task 8: docs + full verification

**Files:**
- Modify: `README.md`, `docs/superpowers/specs/2026-06-14-graffiti-design.md`

- [ ] **Step 1: Update README supported languages**

In `README.md`, update the Status/Usage area to note supported languages: **Go, Python, JavaScript, TypeScript, Rust, Java, PHP** (+ Markdown doc nodes). Add a short "Supported languages" subsection listing them and noting non-Go extraction is files + definitions + methods + imports + intra-repo call edges (under-extracts exotic constructs by design).

- [ ] **Step 2: Amend the spec**

In `docs/superpowers/specs/2026-06-14-graffiti-design.md`:
- §4 (Goals): change the language line to include **PHP** (now Python, JS/TS, Go, Rust, Java, **PHP** + Markdown).
- Append a short "Implemented (Plan 6, 2026-06-16)" note recording: the table-driven `LangSpec`/`Extract` architecture; Go keeps its bespoke `ParseGo`; the six grammars are gated by `grammar_subset_<lang>` tags (subset binary ~10.3 MB); Pass-2 `ResolveCalls` is reused unchanged (non-Go selector calls under-resolve by design); and the **offline fidelity gate** (gotreesitter regression suites + zero-ERROR-node + structural assertions + golden determinism), with true upstream-tree-sitter parity **deferred to CI (Plan 8)**.

- [ ] **Step 3: Full verification**

Run each and confirm:

```bash
# subset config MUST carry all grammar tags (use the Makefile, which now includes them):
make vet
make test
go test ./...                                  # no-tags config (full grammar embed)
go mod tidy && git diff --exit-code go.mod go.sum
make build && make xcompile                    # size guard (expect ~10.3 MB, < 16 MB)
```

Expected: vet clean; both configs green (all packages incl. scan/parse/app); `go mod tidy` a no-op (zero new deps); all five cross-compiled binaries under the 16 MB guard.

- [ ] **Step 4: Polyglot smoke**

```bash
SMOKE=$(mktemp -d)
cp testdata/fixtures/polyglot/* "$SMOKE"/
./graffiti build "$SMOKE"            # success line: 6 files → many nodes/edges/communities
./graffiti query "validate request" "$SMOKE" | head -20
rm -rf "$SMOKE"
```

Expected: build reports 6 files and a non-trivial node/edge/community count; query returns a scoped subgraph mentioning the relevant nodes.

- [ ] **Step 5: Commit**

```bash
git add README.md docs/superpowers/specs/2026-06-14-graffiti-design.md
git commit -m "docs: document Plan 6 multi-language support (incl. PHP)"
```

---

## Self-review checklist (run before merge)

1. **Scope:** six languages added via one table-driven extractor; Go path untouched (gorepo golden still green — Task 5 step 5).
2. **Determinism (§14):** `Extract` walks named children in source order; nodes/edges sorted in `build.Assemble`; polyglot golden byte-stable across two builds (Task 7).
3. **No new deps:** stdlib + existing gotreesitter; `go mod tidy` no-op (Task 8). Six grammar tags added to the Makefile; binary ~10.3 MB < guard.
4. **Fidelity gate (honest):** zero-ERROR-node + structural-shape per language (Task 6); upstream-diff parity explicitly deferred to CI, not silently skipped.
5. **Honesty-first extraction:** non-Go selector calls under-resolve (Pass-2 drops unmatched) rather than emit false edges; nested-in-function definitions under-extracted by design.
6. **Type consistency:** `LangSpec` field names match between `langspec.go` and `extract.go`; `SpecFor`/`NewParser` take `scan.Lang`; `Extract` signature matches every call site (tests + app).

## Deferred follow-ups (record in memory, non-blocking)

- **Upstream tree-sitter parity diffing** belongs in CI (Plan 8): install the `tree-sitter` CLI, parse each fixture with both engines, diff the s-expressions.
- Non-Go **call→module** edges don't form (callees are bare/under-resolved); cross-file/import-aware resolution per language is future work.
- **Nested definitions** (functions/classes inside a function body) are flattened/skipped past the first def level.
- **TSX/JSX** route to the TS/JS grammar respectively; dedicated `tsx`/`jsx` grammars exist if fidelity proves inadequate.
- **Rust trait impls** (`impl Trait for Type`) qualify methods by the first `type_identifier` (the trait), not the implementing type — revisit if it mislabels.
- **PHP/Java visibility & static** and **Python decorators** are ignored (under-extraction).
- Consider migrating Go to the `LangSpec` table later (its receiver handling is the only blocker).
