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
)

func extractBetween(t *testing.T, s, open, close string) string {
	t.Helper()
	i := strings.Index(s, open)
	if i < 0 {
		t.Fatalf("open marker %q not found", open)
	}
	i += len(open)
	j := strings.Index(s[i:], close)
	if j < 0 {
		t.Fatalf("close marker %q not found", close)
	}
	return s[i : i+j]
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
	doc, an := sampleClustered()
	a := RenderMapHTML(doc, an, "2026-06-17T00:00:00Z")
	b := RenderMapHTML(doc, an, "2026-06-17T00:00:00Z")
	if a != b {
		t.Fatalf("not byte-identical for identical inputs (len %d vs %d)", len(a), len(b))
	}
}

// --- 1b. Different generated_at => identical EXCEPT the 2 generated_at lines. ---
func TestMapHTML_DiffersOnlyByGeneratedAt(t *testing.T) {
	doc, an := sampleClustered()
	a := RenderMapHTML(doc, an, "2026-06-17T00:00:00Z")
	b := RenderMapHTML(doc, an, "1999-12-31T23:59:59Z")
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

// --- 2. CSP correctness: recompute the inlined bodies' hashes independently. ---
func TestMapHTML_CSPHashesMatchInlinedBodies(t *testing.T) {
	doc, an := sampleClustered()
	html := RenderMapHTML(doc, an, "2026-06-17T00:00:00Z")
	cspLine := lineContaining(t, strings.Split(html, "\n"), "Content-Security-Policy")

	styleBody := extractBetween(t, html, "<style>", "</style>")
	scriptBody := extractBetween(t, html, "<script>", "</script>") // renderer (island tag has attrs)
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
	doc, an := sampleClustered()
	html := RenderMapHTML(doc, an, "2026-06-17T00:00:00Z")
	for _, b := range []string{"http://", "https://", "src=", "<link", "@import"} {
		if strings.Contains(html, b) {
			t.Fatalf("self-containment violated: found %q", b)
		}
	}
	if regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9+.-]*://`).MatchString(html) {
		t.Fatal("self-containment violated: found a scheme:// URL")
	}
}

// --- 4. Data island parses to the columnar graph shape. ---
func TestMapHTML_DataIslandParses(t *testing.T) {
	doc, an := sampleClustered()
	html := RenderMapHTML(doc, an, "2026-06-17T00:00:00Z")
	island := extractBetween(t, html, `<script type="application/json" id="graffiti-data">`, "</script>")
	var gd graphData
	if err := json.Unmarshal([]byte(island), &gd); err != nil {
		t.Fatalf("data island did not parse as JSON: %v", err)
	}
	n := len(gd.Label)
	if n == 0 || len(gd.Kind) != n || len(gd.File) != n || len(gd.Deg) != n || len(gd.Cat) != n || len(gd.Line) != n {
		t.Fatalf("columnar arrays inconsistent (label=%d)", n)
	}
}

// --- 5. Structural presence: canvas, a11y, tree, the toggles, fit. ---
func TestMapHTML_StructuralPresence(t *testing.T) {
	doc, an := sampleClustered()
	html := RenderMapHTML(doc, an, "2026-06-17T00:00:00Z")
	for _, must := range []string{`<canvas id="c"`, `id="a11y"`, `id="graffiti-data"`, `id="tree"`, "3D depth", "fit to window", "show sector zones"} {
		if !strings.Contains(html, must) {
			t.Fatalf("missing structural element %q", must)
		}
	}
	// a11y mirror lists every node label.
	mirror := extractBetween(t, html, `<nav id="a11y"`, "</nav>")
	for _, nd := range doc.Nodes {
		if !strings.Contains(mirror, htmlEscape(nd.Label)) {
			t.Fatalf("a11y mirror missing node %q", nd.Label)
		}
	}
}

// --- 6. XSS escaping of labels AND the </script island escape. ---
func TestMapHTML_EscapesLabelsAndScriptClose(t *testing.T) {
	doc := graph.NewDocument("demo")
	doc.GeneratedAt = "2026-06-17T00:00:00Z"
	evil := `</script><img src=x onerror=alert(1)>`
	doc.Nodes = []graph.Node{
		{ID: "n0", Label: evil, Kind: graph.KindFunction, File: "a.go", Line: 1, Community: 0},
	}
	an := analyze.Analyze(doc, analyze.Degrees(doc))
	html := RenderMapHTML(doc, an, "2026-06-17T00:00:00Z")

	island := extractBetween(t, html, `<script type="application/json" id="graffiti-data">`, "</script>")
	if strings.Contains(island, "</script") {
		t.Fatalf("data island contains an unescaped </script sequence:\n%s", island)
	}
	if !strings.Contains(island, "\\u003c") {
		t.Fatalf("expected '<' to be escaped to \\u003c in the island")
	}
	var gd graphData
	if err := json.Unmarshal([]byte(island), &gd); err != nil {
		t.Fatalf("escaped island did not parse: %v", err)
	}
	found := false
	for _, l := range gd.Label {
		if l == evil {
			found = true
		}
	}
	if !found {
		t.Fatalf("evil label did not round-trip through the island")
	}
	// In the visible body (a11y mirror) the label must be HTML-escaped.
	body := html[strings.Index(html, "<body"):]
	if strings.Contains(body, "<img src=x onerror") {
		t.Fatalf("label was not HTML-escaped in the body (XSS):\n%s", body)
	}
}

// --- 7. Writer writes next to map.json/MAP.md. ---
func TestWriteMapHTML_WritesNextToJSON(t *testing.T) {
	doc, an := sampleClustered()
	dir := t.TempDir()
	if err := WriteMapHTML(doc, an, dir); err != nil {
		t.Fatalf("WriteMapHTML: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, ".graffiti", "map.html"))
	if err != nil {
		t.Fatalf("read map.html: %v", err)
	}
	if !strings.Contains(string(b), `<canvas id="c"`) {
		t.Fatalf("written map.html missing <canvas>")
	}
}
