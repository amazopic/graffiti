package graph

import "testing"

func TestNormalizeID(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Hello", "hello"},
		{"HTTPRouting", "httprouting"},
		{"Auth & Sessions", "auth-sessions"},
		{"foo__bar", "foo-bar"},
		{"  leading-trailing  ", "leading-trailing"},
		{"a/b/c.go", "a-b-c-go"},
		{"already-clean", "already-clean"},
		{"Multiple   Spaces", "multiple-spaces"},
		{"", ""},
		{"!!!", ""},
	}
	for _, c := range cases {
		if got := NormalizeID(c.in); got != c.want {
			t.Errorf("NormalizeID(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNodeID_StableAndQualified(t *testing.T) {
	a := NodeID("greet/greet.go", "Hello")
	b := NodeID("greet/greet.go", "Hello")
	if a != b {
		t.Fatalf("NodeID not stable: %q vs %q", a, b)
	}
	c := NodeID("main.go", "Hello")
	if a == c {
		t.Fatalf("NodeID should differ across files: both %q", a)
	}
	if a != "greet-greet-go:hello" {
		t.Fatalf("NodeID format = %q, want %q", a, "greet-greet-go:hello")
	}
}
