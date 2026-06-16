package workspace

import (
	"path/filepath"
	"testing"
)

func TestCommonAncestor(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{[]string{"/a/b/frontend", "/a/b/backend"}, "/a/b"},
		{[]string{"/a/b/c", "/a/b/c/d"}, "/a/b/c"},
		{[]string{"/a/x", "/a/y", "/a/z/q"}, "/a"},
		{[]string{"/only/one"}, "/only/one"},
	}
	for _, c := range cases {
		in := make([]string, len(c.in))
		for i, p := range c.in {
			in[i] = filepath.FromSlash(p)
		}
		if got := CommonAncestor(in); got != filepath.FromSlash(c.want) {
			t.Errorf("CommonAncestor(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}
