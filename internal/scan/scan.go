// Package scan discovers and classifies source files under a repository root,
// honoring .gitignore, filtering to supported extensions, and returning a
// deterministically ordered slice (spec §5 scan stage).
package scan

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

// Lang classifies a discovered file.
type Lang string

const (
	LangGo       Lang = "go"
	LangMarkdown Lang = "markdown"
)

// extLang maps supported file extensions to their language (Plan 1 scope:
// Go + Markdown only).
var extLang = map[string]Lang{
	".go": LangGo,
	".md": LangMarkdown,
}

// FileRef is a discovered, classified file.
type FileRef struct {
	AbsPath string // absolute path on disk
	RelPath string // path relative to root, slash-separated
	Lang    Lang
}

// alwaysSkipDirs are directory names never descended into, regardless of .gitignore.
var alwaysSkipDirs = map[string]bool{
	".git":         true,
	".graffiti":    true,
	"node_modules": true,
	"vendor":       true,
}

// Scan walks root and returns supported files in deterministic order
// (by RelPath, slash-separated, ascending).
func Scan(root string) ([]FileRef, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	ign := loadGitignore(absRoot)

	var refs []FileRef
	walkErr := filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == absRoot {
			return nil
		}
		rel, relErr := filepath.Rel(absRoot, path)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)

		if d.IsDir() {
			if alwaysSkipDirs[d.Name()] {
				return filepath.SkipDir
			}
			if ign != nil && ign.MatchesPath(rel+"/") {
				return filepath.SkipDir
			}
			return nil
		}

		if ign != nil && ign.MatchesPath(rel) {
			return nil
		}
		lang, ok := extLang[strings.ToLower(filepath.Ext(rel))]
		if !ok {
			return nil
		}
		refs = append(refs, FileRef{AbsPath: path, RelPath: rel, Lang: lang})
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	sort.Slice(refs, func(i, j int) bool { return refs[i].RelPath < refs[j].RelPath })
	return refs, nil
}

// loadGitignore reads <root>/.gitignore if present. Returns nil if there is none.
func loadGitignore(absRoot string) *gitignore.GitIgnore {
	p := filepath.Join(absRoot, ".gitignore")
	if _, err := os.Stat(p); err != nil {
		return nil
	}
	ign, err := gitignore.CompileIgnoreFile(p)
	if err != nil {
		return nil
	}
	return ign
}
