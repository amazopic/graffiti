package query_test // external test package: imports app (which links parse) — needs the build tags

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/app"
	"github.com/evgeniy-achin/graffiti/internal/query"
	"github.com/evgeniy-achin/graffiti/internal/store"
)

const fixtureGenAt = "2026-06-16T00:00:00Z"

// goldenQuestion matches real fixture symbols: "formatter"/"format" hit
// Formatter.Format (method), "hello" hits Hello (function), "greeting" has no
// exact match but "formatter"/"format" are high-IDF seeds. BFS expansion pulls
// in greet/greet.go (contains Hello, upper) and greet/greet_helper.go (contains
// Formatter, Formatter.Format), producing a non-empty NODES + EDGES subgraph.
const goldenQuestion = "how does formatter format hello greeting"

// copyTree mirrors the helper Plans 1–3 use in app tests.
func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		s, d := filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())
		if e.IsDir() {
			copyTree(t, s, d)
			continue
		}
		b, err := os.ReadFile(s)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(d, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func runFixtureQuery(t *testing.T) string {
	t.Helper()
	src := filepath.Join("..", "..", "testdata", "fixtures", "gorepo")
	dst := t.TempDir()
	copyTree(t, src, dst)
	if _, err := app.Build(dst, fixtureGenAt); err != nil {
		t.Fatalf("build fixture: %v", err)
	}
	doc, err := store.Load(filepath.Join(dst, ".graffiti", "map.json"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return query.Query(store.NewIndex(doc), goldenQuestion, query.DefaultTokenBudget)
}

func TestGolden_GoRepoQuery(t *testing.T) {
	got := runFixtureQuery(t)
	path := filepath.Join("..", "..", "testdata", "golden", "gorepo.query.txt")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Log("query golden updated")
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden (run UPDATE_GOLDEN=1 to create): %v", err)
	}
	if got != string(want) {
		t.Fatalf("query output differs from golden:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestGolden_GoRepoQuery_Deterministic(t *testing.T) {
	if runFixtureQuery(t) != runFixtureQuery(t) {
		t.Fatal("two fixture queries not byte-identical")
	}
}
