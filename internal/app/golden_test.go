package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/graph"
)

const fixtureGenAt = "2026-06-14T00:00:00Z"

func goldenPath() string {
	return filepath.Join("..", "..", "testdata", "golden", "gorepo.map.json")
}

// buildFixtureIntoTemp copies the committed fixture repo into a temp dir, builds
// it there (so we never write .graffiti into testdata), and returns the produced
// map.json bytes.
func buildFixtureIntoTemp(t *testing.T) []byte {
	t.Helper()
	src := filepath.Join("..", "..", "testdata", "fixtures", "gorepo")
	dst := t.TempDir()
	copyTree(t, src, dst)

	if _, err := Build(dst, fixtureGenAt); err != nil {
		t.Fatalf("Build fixture: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dst, ".graffiti", "map.json"))
	if err != nil {
		t.Fatalf("read produced map.json: %v", err)
	}
	return b
}

func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("copyTree: %v", err)
	}
}

var reGenAt = regexp.MustCompile(`("generated_at":\s*")[^"]*(")`)
var reRoot = regexp.MustCompile(`("root":\s*")[^"]*(")`)

func strip(b []byte) []byte {
	b = reGenAt.ReplaceAll(b, []byte(`${1}X${2}`))
	b = reRoot.ReplaceAll(b, []byte(`${1}X${2}`))
	return b
}

func TestGolden_GoRepoMapJSON(t *testing.T) {
	got := buildFixtureIntoTemp(t)

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(goldenPath(), got, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Log("golden updated")
		return
	}

	want, err := os.ReadFile(goldenPath())
	if err != nil {
		t.Fatalf("read golden (run UPDATE_GOLDEN=1 to create): %v", err)
	}
	if string(strip(got)) != string(strip(want)) {
		t.Fatalf("map.json differs from golden (modulo generated_at+root).\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestDeterminism_TwoBuildsByteIdentical(t *testing.T) {
	a := buildFixtureIntoTemp(t)
	b := buildFixtureIntoTemp(t)
	if string(strip(a)) != string(strip(b)) {
		t.Fatalf("two builds not byte-identical modulo generated_at+root")
	}
}

// TestGolden_StructuralShape enforces the EXPECTED GRAPH by code (not only by a
// frozen blob), so a wrong golden cannot silently pass.
func TestGolden_StructuralShape(t *testing.T) {
	var doc graph.Document
	if err := json.Unmarshal(buildFixtureIntoTemp(t), &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	byID := map[string]graph.Node{}
	for _, n := range doc.Nodes {
		byID[n.ID] = n
	}
	mustNode := func(file, label string, kind graph.Kind) string {
		id := graph.NodeID(file, label)
		n, ok := byID[id]
		if !ok {
			t.Fatalf("missing node %q (%s in %s)", id, label, file)
		}
		if n.Kind != kind {
			t.Fatalf("node %q kind = %q, want %q", id, n.Kind, kind)
		}
		return id
	}

	// file nodes
	mustNode("main.go", "main.go", graph.KindFile)
	mustNode("greet/greet.go", "greet/greet.go", graph.KindFile)
	mustNode("greet/greet_helper.go", "greet/greet_helper.go", graph.KindFile)

	// definitions
	mainID := mustNode("main.go", "main", graph.KindFunction)
	helloID := mustNode("greet/greet.go", "Hello", graph.KindFunction)
	mustNode("greet/greet.go", "upper", graph.KindFunction)
	mustNode("greet/greet_helper.go", "Formatter", graph.KindClass)
	fmtFormatID := mustNode("greet/greet_helper.go", "Formatter.Format", graph.KindMethod)

	// module nodes (keyed by full import path)
	fmtModID := graph.NodeID("module:fmt", "fmt")
	stringsModID := graph.NodeID("module:strings", "strings")
	greetModID := graph.NodeID("module:example.com/gorepo/greet", "greet")
	for _, id := range []string{fmtModID, stringsModID, greetModID} {
		if n, ok := byID[id]; !ok || n.Kind != graph.KindModule {
			t.Fatalf("missing module node %q", id)
		}
	}

	// expected calls edges (deliberate asymmetry, spec §5):
	//   main -> greet (module, EXTRACTED via import)   [greet.Hello selector]
	//   main -> fmt   (module, EXTRACTED via import)    [fmt.Println selector]
	//   Hello -> upper (function, INFERRED, same package)
	//   upper -> strings (module, EXTRACTED via import) [strings.ToUpper selector]
	//   Formatter.Format -> Hello (function, INFERRED)
	wantCalls := map[[2]string]graph.Confidence{
		{mainID, greetModID}: graph.ConfExtracted,
		{mainID, fmtModID}:   graph.ConfExtracted,
		{helloID, graph.NodeID("greet/greet.go", "upper")}:      graph.ConfInferred,
		{graph.NodeID("greet/greet.go", "upper"), stringsModID}: graph.ConfExtracted,
		{fmtFormatID, helloID}:                                  graph.ConfInferred,
	}
	gotCalls := map[[2]string]graph.Confidence{}
	for _, e := range doc.Edges {
		if e.Relation == graph.RelCalls {
			gotCalls[[2]string{e.From, e.To}] = e.Confidence
		}
	}
	for k, conf := range wantCalls {
		if got, ok := gotCalls[k]; !ok || got != conf {
			t.Fatalf("calls edge %v: got conf=%q ok=%v, want %q", k, got, ok, conf)
		}
	}
}
