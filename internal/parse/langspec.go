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
	ImportChild string            // child node kind holding the module string (JS/TS: "string"); "" = first named child
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
