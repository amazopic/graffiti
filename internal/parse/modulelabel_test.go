package parse

import "testing"

func TestLastSegment(t *testing.T) {
	cases := map[string]string{
		"os":                          "os",
		"auth.session":                "session", // python from-import
		"./auth/session.js":           "session", // JS path import (extension stripped)
		"./auth/session":              "session", // TS path import (no extension)
		"std::collections::HashMap":   "HashMap", // rust
		"java.util.List":              "List",    // java
		"App\\Support\\Str":           "Str",     // php namespace
		"react":                       "react",   // bare package
		"@scope/pkg":                  "pkg",     // scoped npm package
	}
	for in, want := range cases {
		if got := lastSegment(in); got != want {
			t.Errorf("lastSegment(%q) = %q, want %q", in, got, want)
		}
	}
}
