package app

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

// buildFixtureMapMD builds the fixture into a temp dir and returns the produced
// MAP.md bytes (companion to buildFixtureIntoTemp, which returns map.json).
func buildFixtureMapMD(t *testing.T) []byte {
	t.Helper()
	src := filepath.Join("..", "..", "testdata", "fixtures", "gorepo")
	dst := t.TempDir()
	copyTree(t, src, dst)
	if _, err := Build(dst, fixtureGenAt); err != nil {
		t.Fatalf("Build fixture: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dst, ".graffiti", "MAP.md"))
	if err != nil {
		t.Fatalf("read produced MAP.md: %v", err)
	}
	return b
}

func mapMDGoldenPath() string {
	return filepath.Join("..", "..", "testdata", "golden", "gorepo.MAP.md")
}

func TestGolden_MapMD(t *testing.T) {
	got := buildFixtureMapMD(t)
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(mapMDGoldenPath(), got, 0o644); err != nil {
			t.Fatalf("write MAP.md golden: %v", err)
		}
		t.Log("MAP.md golden updated")
		return
	}
	want, err := os.ReadFile(mapMDGoldenPath())
	if err != nil {
		t.Fatalf("read MAP.md golden (run UPDATE_GOLDEN=1 to create): %v", err)
	}
	// MAP.md prints the root in its title; normalize it like map.json's root.
	norm := func(b []byte) string {
		s := string(b)
		if i := strings.IndexByte(s, '\n'); i >= 0 {
			s = "# X — Map" + s[i:] // replace the title line (root-dependent)
		}
		return s
	}
	if norm(got) != norm(want) {
		t.Fatalf("MAP.md differs from golden.\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// TestClustering_StructuralInvariants asserts the clustering CONTRACT on the real
// fixture (not blind community numbers): every node has a community >= 0; labels
// are contiguous 0..K-1; communities[] matches node assignments; members sorted;
// intra-file coupled symbols co-locate (Hello+upper; Formatter+Formatter.Format);
// and the weak cross-file Formatter.Format->Hello call stays cross-community. A
// broken clustering (unclustered nodes, gaps, or collapse-everything) fails loudly.
func TestClustering_StructuralInvariants(t *testing.T) {
	var doc graph.Document
	if err := json.Unmarshal(buildFixtureIntoTemp(t), &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// 1. Every node clustered (>= 0) and labels contiguous 0..K-1.
	maxC := -1
	seen := map[int]bool{}
	commOf := map[string]int{}
	for _, n := range doc.Nodes {
		if n.Community < 0 {
			t.Fatalf("node %q still unclustered (community=%d)", n.ID, n.Community)
		}
		seen[n.Community] = true
		commOf[n.ID] = n.Community
		if n.Community > maxC {
			maxC = n.Community
		}
	}
	for i := 0; i <= maxC; i++ {
		if !seen[i] {
			t.Fatalf("community %d missing — labels not contiguous", i)
		}
	}

	// 2. communities[] is consistent with node assignments, ids contiguous,
	//    members sorted and complete.
	if len(doc.Communities) != maxC+1 {
		t.Fatalf("communities len = %d, want %d", len(doc.Communities), maxC+1)
	}
	for i, c := range doc.Communities {
		if c.ID != i {
			t.Fatalf("communities[%d].id = %d, want %d (contiguous, sorted)", i, c.ID, i)
		}
		if c.Label == "" {
			t.Fatalf("community %d has empty label", c.ID)
		}
		if !sortedStrings(c.Members) {
			t.Fatalf("community %d members not sorted: %v", c.ID, c.Members)
		}
		for _, m := range c.Members {
			if commOf[m] != c.ID {
				t.Fatalf("member %q listed in community %d but node says %d", m, c.ID, commOf[m])
			}
		}
	}

	// 3. Tightly-coupled symbols co-locate (validated against the real fixture).
	//    - greet.go's own defs Hello and upper are joined by the intra-file call
	//      Hello->upper (plus both anchored to their file node), so they share a
	//      community.
	//    - Formatter and Formatter.Format live in greet_helper.go and are joined
	//      by the file `contains` edges, so they share a community.
	hello := graph.NodeID("greet/greet.go", "Hello")
	upper := graph.NodeID("greet/greet.go", "upper")
	formatter := graph.NodeID("greet/greet_helper.go", "Formatter")
	format := graph.NodeID("greet/greet_helper.go", "Formatter.Format")
	if commOf[hello] != commOf[upper] {
		t.Fatalf("Hello (c=%d) and upper (c=%d) should share a community", commOf[hello], commOf[upper])
	}
	if commOf[formatter] != commOf[format] {
		t.Fatalf("Formatter (c=%d) and Formatter.Format (c=%d) should share a community", commOf[formatter], commOf[format])
	}

	// 4. The Formatter.Format -> Hello call crosses community boundaries on this
	//    fixture (greet.go defs and greet_helper.go defs cluster apart), so the
	//    weak cross-package coupling is genuinely cross-community: an invariant a
	//    broken (collapse-everything) clustering would violate.
	if commOf[format] == commOf[hello] {
		t.Fatalf("Formatter.Format (c=%d) and Hello (c=%d) unexpectedly share a community; "+
			"the cross-file INFERRED call should remain a cross-community (surprising) edge", commOf[format], commOf[hello])
	}
}

func sortedStrings(s []string) bool {
	for i := 1; i < len(s); i++ {
		if s[i-1] > s[i] {
			return false
		}
	}
	return true
}

// buildFixtureMapHTML builds the fixture into a temp dir and returns the produced
// map.html bytes (companion to buildFixtureIntoTemp / buildFixtureMapMD).
func buildFixtureMapHTML(t *testing.T) []byte {
	t.Helper()
	src := filepath.Join("..", "..", "testdata", "fixtures", "gorepo")
	dst := t.TempDir()
	copyTree(t, src, dst)
	if _, err := Build(dst, fixtureGenAt); err != nil {
		t.Fatalf("Build fixture: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dst, ".graffiti", "map.html"))
	if err != nil {
		t.Fatalf("read produced map.html: %v", err)
	}
	return b
}

// stripGeneratedAt blanks the two generated_at-bearing carriers (HTML comment +
// body data-attribute) so two builds compare byte-equal modulo the timestamp.
// The root also appears in the <title> and top-bar span; strip those too so
// builds into different temp dirs (different roots) still compare equal.
var reHTMLComment = regexp.MustCompile(`<!-- generated_at: [^>]*-->`)
var reHTMLDataAttr = regexp.MustCompile(`data-generated-at="[^"]*"`)
var reHTMLTitle = regexp.MustCompile(`<title>Graffiti Districts — [^<]*</title>`)
var reHTMLTopBar = regexp.MustCompile(`Districts — [^<]*</span>`)

func stripHTML(b []byte) []byte {
	b = reHTMLComment.ReplaceAll(b, []byte(`<!-- generated_at: X -->`))
	b = reHTMLDataAttr.ReplaceAll(b, []byte(`data-generated-at="X"`))
	b = reHTMLTitle.ReplaceAll(b, []byte(`<title>Graffiti Districts — X</title>`))
	b = reHTMLTopBar.ReplaceAll(b, []byte(`Districts — X</span>`))
	return b
}

func TestMapHTML_TwoBuildsByteIdenticalModuloGeneratedAt(t *testing.T) {
	a := stripHTML(buildFixtureMapHTML(t))
	b := stripHTML(buildFixtureMapHTML(t))
	if string(a) != string(b) {
		t.Fatalf("two map.html builds not byte-identical modulo generated_at")
	}
}

func TestMapHTML_SelfContainedAndCSP(t *testing.T) {
	html := string(buildFixtureMapHTML(t))

	for _, banned := range []string{"http://", "https://", "src=", "<link", "@import"} {
		if strings.Contains(html, banned) {
			t.Fatalf("self-containment violated: found %q", banned)
		}
	}
	if regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9+.-]*://`).MatchString(html) {
		t.Fatalf("self-containment violated: scheme:// URL present")
	}

	// Recompute the inlined style/script hashes and assert they're in the CSP.
	extract := func(open, close string) string {
		i := strings.Index(html, open)
		if i < 0 {
			t.Fatalf("marker %q not found", open)
		}
		i += len(open)
		j := strings.Index(html[i:], close)
		if j < 0 {
			t.Fatalf("close %q not found", close)
		}
		return html[i : i+j]
	}
	h := func(s string) string {
		sum := sha256.Sum256([]byte(s))
		return "sha256-" + base64.StdEncoding.EncodeToString(sum[:])
	}
	var csp string
	for _, l := range strings.Split(html, "\n") {
		if strings.Contains(l, "Content-Security-Policy") {
			csp = l
		}
	}
	if csp == "" {
		t.Fatal("no CSP meta line")
	}
	for _, want := range []string{
		h(extract("<style>", "</style>")),
		h(extract("<script>", "</script>")),
		h(extract(`<script type="application/json" id="graffiti-data">`, "</script>")),
	} {
		if !strings.Contains(csp, want) {
			t.Fatalf("CSP missing hash %q\nCSP=%s", want, csp)
		}
	}
}

func TestMapHTML_StructuralAndA11yMirror(t *testing.T) {
	html := string(buildFixtureMapHTML(t))
	for _, must := range []string{`<canvas id="canvas"`, `<nav id="a11y"`, `id="graffiti-data"`, "Start here", "Landmarks", "Confidence"} {
		if !strings.Contains(html, must) {
			t.Fatalf("map.html missing %q", must)
		}
	}
	// Every clustered community label from the fixture appears in the a11y mirror.
	var doc graph.Document
	if err := json.Unmarshal(buildFixtureIntoTemp(t), &doc); err != nil {
		t.Fatalf("unmarshal map.json: %v", err)
	}
	mi := strings.Index(html, `<nav id="a11y"`)
	mirror := html[mi : mi+strings.Index(html[mi:], "</nav>")]
	for _, c := range doc.Communities {
		if !strings.Contains(mirror, c.Label) {
			t.Fatalf("a11y mirror missing district %q", c.Label)
		}
	}
}

func mapHTMLGoldenPath() string {
	return filepath.Join("..", "..", "testdata", "golden", "gorepo.map.html.strip")
}

// TestGolden_MapHTMLStrip locks the byte-exact map.html modulo generated_at. The
// golden stores the stripped bytes; regenerate via UPDATE_GOLDEN=1.
func TestGolden_MapHTMLStrip(t *testing.T) {
	got := stripHTML(buildFixtureMapHTML(t))
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(mapHTMLGoldenPath(), got, 0o644); err != nil {
			t.Fatalf("write map.html golden: %v", err)
		}
		t.Log("map.html strip golden updated")
		return
	}
	want, err := os.ReadFile(mapHTMLGoldenPath())
	if err != nil {
		t.Fatalf("read map.html golden (run UPDATE_GOLDEN=1 to create): %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("map.html differs from golden (modulo generated_at)")
	}
}

// TestSuggestedQuestions_Shape asserts exactly 3 deterministic questions, with the
// fixture's actual top god node surfaced in question 1. Validated against real
// build output: the highest-degree node is the greet/greet.go FILE node (degree 3:
// it contains Hello+upper and imports strings), so question 1 names it. (Hello,
// upper and main are also degree-3 but the file node sorts first by the god-node
// total order, so it is the one surfaced.)
func TestSuggestedQuestions_Shape(t *testing.T) {
	src := filepath.Join("..", "..", "testdata", "fixtures", "gorepo")
	dst := t.TempDir()
	copyTree(t, src, dst)
	stats, err := Build(dst, fixtureGenAt)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(stats.Questions) != 3 {
		t.Fatalf("questions = %d, want exactly 3", len(stats.Questions))
	}
	if !strings.Contains(stats.Questions[0], "greet/greet.go") {
		t.Fatalf("question 1 should name the top hub greet/greet.go, got %q", stats.Questions[0])
	}
}
