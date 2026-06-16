package workspace

import (
	"os"
	"path/filepath"
	"strings"
)

// CommonAncestor returns the deepest directory that is a prefix of every input
// path (the default workspace root, §16.4). Inputs should be absolute, cleaned
// paths. With a single input it returns that path unchanged.
func CommonAncestor(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	split := func(p string) []string {
		return strings.Split(filepath.Clean(p), string(filepath.Separator))
	}
	common := split(paths[0])
	for _, p := range paths[1:] {
		parts := split(p)
		n := len(common)
		if len(parts) < n {
			n = len(parts)
		}
		i := 0
		for i < n && common[i] == parts[i] {
			i++
		}
		common = common[:i]
	}
	joined := strings.Join(common, string(filepath.Separator))
	if joined == "" {
		return string(filepath.Separator)
	}
	return joined
}

// FindRoot searches cwd and its ancestors for a directory containing
// .graffiti-workspace/workspace.json and returns that directory, or "" if none.
func FindRoot(cwd string) string {
	dir := filepath.Clean(cwd)
	for {
		if _, err := os.Stat(filepath.Join(dir, WorkspaceDir, registryFile)); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
