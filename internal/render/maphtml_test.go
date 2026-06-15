package render

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/evgeniy-achin/graffiti/internal/analyze"
	"github.com/evgeniy-achin/graffiti/internal/graph"
	"github.com/evgeniy-achin/graffiti/internal/layout"
)

// sampleClusteredScene builds a small clustered doc + analysis + baked scene.
// (sampleClustered without the scene already exists in mapmd_test.go.)
func sampleClusteredScene(t *testing.T) (*graph.Document, analyze.Analysis, layout.Scene) {
	t.Helper()
	doc, an := sampleClustered()
	return doc, an, layout.Layout(doc, an)
}

func extractBetween(t *testing.T, html, open, close string) string {
	t.Helper()
	i := strings.Index(html, open)
	if i < 0 {
		t.Fatalf("open marker %q not found", open)
	}
	i += len(open)
	j := strings.Index(html[i:], close)
	if j < 0 {
		t.Fatalf("close marker %q not found", close)
	}
	return html[i : i+j]
}
func recompSha256B64(s string) string {
	sum := sha256.Sum256([]byte(s))
	return base64.StdEncoding.EncodeToString(sum[:])
}
func lineContaining(t *testing.T, lines []string, sub string) string {
	t.Helper()
	for _, l := range lines {
		if strings.Contains(l, sub) {
			return l
		}
	}
	t.Fatalf("no line containing %q", sub)
	return ""
}

// --- 1. Determinism: same generated_at => byte-identical. ---
func TestMapHTML_DeterministicSameTime(t *testing.T) {
	doc, an, sc := sampleClusteredScene(t)
	a := RenderMapHTML(doc, an, sc, "2026-06-15T00:00:00Z")
	b := RenderMapHTML(doc, an, sc, "2026-06-15T00:00:00Z")
	if a != b {
		t.Fatalf("not byte-identical for identical inputs (len %d vs %d)", len(a), len(b))
	}
}

// --- 1b. Determinism: different generated_at => identical EXCEPT the two
// generated_at-bearing lines; CSP hashes unchanged. ---
func TestMapHTML_DiffersOnlyByGeneratedAt(t *testing.T) {
	doc, an, sc := sampleClusteredScene(t)
	t1, t2 := "2026-06-15T00:00:00Z", "1999-12-31T23:59:59Z"
	a := RenderMapHTML(doc, an, sc, t1)
	b := RenderMapHTML(doc, an, sc, t2)
	if a == b {
		t.Fatal("expected difference when generated_at changes")
	}
	la, lb := strings.Split(a, "\n"), strings.Split(b, "\n")
	if len(la) != len(lb) {
		t.Fatalf("line count differs: %d vs %d", len(la), len(lb))
	}
	var diff []int
	for i := range la {
		if la[i] != lb[i] {
			diff = append(diff, i)
		}
	}
	if len(diff) != 2 {
		t.Fatalf("expected exactly 2 differing lines, got %d: %v", len(diff), diff)
	}
	for _, i := range diff {
		if !strings.Contains(la[i], "generated") {
			t.Fatalf("differing line %d is not a generated_at carrier: %q", i, la[i])
		}
	}
	if lineContaining(t, la, "Content-Security-Policy") != lineContaining(t, lb, "Content-Security-Policy") {
		t.Fatal("CSP line changed across generated_at")
	}
}

// --- 2. CSP correctness: independently recompute the inlined bodies' hashes. ---
func TestMapHTML_CSPHashesMatchInlinedBodies(t *testing.T) {
	doc, an, sc := sampleClusteredScene(t)
	html := RenderMapHTML(doc, an, sc, "2026-06-15T00:00:00Z")
	cspLine := lineContaining(t, strings.Split(html, "\n"), "Content-Security-Policy")

	styleBody := extractBetween(t, html, "<style>", "</style>")
	scriptBody := extractBetween(t, html, "<script>", "</script>")
	dataBody := extractBetween(t, html, `<script type="application/json" id="graffiti-data">`, "</script>")

	for _, want := range []string{
		"sha256-" + recompSha256B64(styleBody),
		"sha256-" + recompSha256B64(scriptBody),
		"sha256-" + recompSha256B64(dataBody),
	} {
		if !strings.Contains(cspLine, want) {
			t.Fatalf("hash %q not in CSP: %s", want, cspLine)
		}
	}
	for _, must := range []string{"default-src 'none'", "style-src 'sha256-", "img-src data:"} {
		if !strings.Contains(cspLine, must) {
			t.Fatalf("CSP missing directive %q: %s", must, cspLine)
		}
	}
}

// --- 3. Self-containment: no external refs anywhere. ---
func TestMapHTML_SelfContained(t *testing.T) {
	doc, an, sc := sampleClusteredScene(t)
	html := RenderMapHTML(doc, an, sc, "2026-06-15T00:00:00Z")
	for _, b := range []string{"http://", "https://", "src=", "<link", "@import"} {
		if strings.Contains(html, b) {
			t.Fatalf("self-containment violated: found %q", b)
		}
	}
	if regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9+.-]*://`).MatchString(html) {
		t.Fatal("self-containment violated: found a scheme:// URL")
	}
}

// --- 4. Data island parses to the columnar shape. ---
func TestMapHTML_DataIslandParses(t *testing.T) {
	doc, an, sc := sampleClusteredScene(t)
	html := RenderMapHTML(doc, an, sc, "2026-06-15T00:00:00Z")
	island := extractBetween(t, html, `<script type="application/json" id="graffiti-data">`, "</script>")
	var cs ColumnarScene
	if err := json.Unmarshal([]byte(island), &cs); err != nil {
		t.Fatalf("data island did not parse as JSON: %v", err)
	}
	if cs.W != sc.W || cs.H2 != sc.H {
		t.Fatalf("canvas dims mismatch: %dx%d want %dx%d", cs.W, cs.H2, sc.W, sc.H)
	}
	if len(cs.BoxComm) != len(sc.Boxes) {
		t.Fatalf("boxes in island = %d, want %d", len(cs.BoxComm), len(sc.Boxes))
	}
	if len(cs.Strings) == 0 || cs.Strings[0] != "" {
		t.Fatalf("string table must start with \"\"")
	}
}

// --- 5. Structural presence: <canvas>, a11y mirror lists every district, 3 Qs. ---
func TestMapHTML_StructuralPresence(t *testing.T) {
	doc, an, sc := sampleClusteredScene(t)
	html := RenderMapHTML(doc, an, sc, "2026-06-15T00:00:00Z")
	for _, must := range []string{`<canvas id="canvas"`, `id="a11y"`, `id="graffiti-data"`, "Start here", "Landmarks", "Confidence"} {
		if !strings.Contains(html, must) {
			t.Fatalf("missing structural element %q", must)
		}
	}
	// a11y mirror lists every district label.
	mirror := extractBetween(t, html, `<nav id="a11y"`, "</nav>")
	for _, c := range doc.Communities {
		if !strings.Contains(mirror, c.Label) {
			t.Fatalf("a11y mirror missing district %q", c.Label)
		}
	}
	// exactly the 3 questions appear (as <li> items in the Start here block).
	for _, q := range an.Questions {
		if !strings.Contains(html, htmlEscape(q)) {
			t.Fatalf("question missing from map.html: %q", q)
		}
	}
}

// --- 6. XSS escaping of labels AND the </script island escape. ---
func TestMapHTML_EscapesLabelsAndScriptClose(t *testing.T) {
	doc := graph.NewDocument("demo")
	doc.GeneratedAt = "2026-06-15T00:00:00Z"
	// A malicious label containing both an HTML tag and a script-close sequence.
	evil := `</script><img src=x onerror=alert(1)>`
	doc.Nodes = []graph.Node{
		{ID: "n0", Label: evil, Kind: graph.KindFunction, File: "a.go", Line: 1, Community: 0},
	}
	doc.Communities = []graph.Community{{ID: 0, Label: evil, Members: []string{"n0"}}}
	an := analyze.Analyze(doc, analyze.Degrees(doc))
	sc := layout.Layout(doc, an)
	html := RenderMapHTML(doc, an, sc, "2026-06-15T00:00:00Z")

	// The raw closing-script sequence must not appear inside the JSON island.
	island := extractBetween(t, html, `<script type="application/json" id="graffiti-data">`, "</script>")
	if strings.Contains(island, "</script") {
		t.Fatalf("data island contains an unescaped </script sequence:\n%s", island)
	}
	if !strings.Contains(island, "\\u003c") {
		t.Fatalf("expected '<' to be escaped to \\u003c in the island")
	}
	// The island still parses (escape is JSON-legal) and round-trips the label.
	var cs ColumnarScene
	if err := json.Unmarshal([]byte(island), &cs); err != nil {
		t.Fatalf("escaped island did not parse: %v", err)
	}
	found := false
	for _, s := range cs.Strings {
		if s == evil {
			found = true
		}
	}
	if !found {
		t.Fatalf("evil label did not round-trip through the island")
	}
	// In the visible HTML body (rail/a11y mirror) the label must be HTML-escaped:
	// no raw "<img" and no raw onerror attribute should survive.
	body := html[strings.Index(html, "<body"):]
	if strings.Contains(body, "<img src=x onerror") {
		t.Fatalf("label was not HTML-escaped in the body (XSS):\n%s", body)
	}
}

// --- 7. Writer writes next to map.json/MAP.md. ---
func TestWriteMapHTML_WritesNextToJSON(t *testing.T) {
	doc, an, sc := sampleClusteredScene(t)
	dir := t.TempDir()
	if err := WriteMapHTML(doc, an, sc, dir); err != nil {
		t.Fatalf("WriteMapHTML: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, ".graffiti", "map.html"))
	if err != nil {
		t.Fatalf("read map.html: %v", err)
	}
	if !strings.Contains(string(b), `<canvas id="canvas"`) {
		t.Fatalf("written map.html missing <canvas>")
	}
}
