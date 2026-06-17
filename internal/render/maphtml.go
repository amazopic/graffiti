package render

import (
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/evgeniy-achin/graffiti/internal/analyze"
	"github.com/evgeniy-achin/graffiti/internal/graph"
)

// Viewer assets are inlined verbatim into map.html and hashed for the CSP. They
// are pure data here so the emitted sha256 hashes are a function only of these
// (deterministic) bytes — never of generated_at or graph content.
//
//go:embed viewer/app.css
var viewerCSS string

//go:embed viewer/app.js
var viewerJS string

// Keep embed referenced even if a future refactor drops one of the strings.
var _ embed.FS

// htmlEscape escapes interpolated labels/paths for safe inlining into HTML body
// text (XSS-safe self-contained file, spec §8.1).
func htmlEscape(s string) string { return html.EscapeString(s) }

// escapeScriptClose makes a JSON byte string safe to inline inside a
// <script type="application/json"> island: it escapes every '<' to <, which
// both (a) prevents a "</script" sequence in any label/path from prematurely
// closing the island and (b) is fully JSON-legal so the island still round-trips
// via JSON.parse. Applied BEFORE hashing, so the island sha256 stays a pure
// function of the graph.
func escapeScriptClose(jsonBytes []byte) string {
	return strings.ReplaceAll(string(jsonBytes), "<", `<`)
}

func sha256b64(s string) string {
	sum := sha256.Sum256([]byte(s))
	return base64.StdEncoding.EncodeToString(sum[:])
}

// WriteMapHTML renders map.html and writes it to <root>/.graffiti/map.html, next
// to map.json and MAP.md. generatedAt is read off doc.GeneratedAt.
func WriteMapHTML(doc *graph.Document, an analyze.Analysis, root string) error {
	dir := filepath.Join(root, ".graffiti")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	out := RenderMapHTML(doc, an, doc.GeneratedAt)
	return os.WriteFile(filepath.Join(dir, "map.html"), []byte(out), 0o644)
}

// WriteWorkspaceHTML renders a federated CombinedDocument (alias-prefixed nodes +
// file paths + cross-edges) to outPath, reusing the single-project force-graph
// renderer. The file-path prefixes make each project the top level of the tree.
func WriteWorkspaceHTML(doc *graph.Document, generatedAt, outPath string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	out := RenderMapHTML(doc, analyze.Analysis{}, generatedAt)
	return os.WriteFile(outPath, []byte(out), 0o644)
}

// RenderMapHTML builds the single self-contained, offline, CSP-safe map.html: an
// interactive force-directed graph (layout computed in the browser). It is a pure
// function of (doc, generatedAt) for everything hashed: generatedAt appears ONLY
// in an HTML comment + a body data-attribute, both OUTSIDE every hashed body, so
// the CSP hashes never vary with time. The graph data ships as a compact columnar
// JSON island with the </script sequence escaped; the renderer (viewer/app.js) and
// theme (viewer/app.css) come from go:embed. A hidden ordered <nav id=a11y> mirrors
// the graph (grouped by directory) for screen readers and find-in-page (§8.7).
// The analyze.Analysis is currently unused by the renderer but kept in the
// signature for forward-compatible god-node/question surfacing.
func RenderMapHTML(doc *graph.Document, an analyze.Analysis, generatedAt string) string {
	_ = an
	dataJSON, err := json.Marshal(graphIsland(doc))
	if err != nil {
		panic(fmt.Sprintf("marshal graph island: %v", err)) // strings/ints only
	}
	island := escapeScriptClose(dataJSON) // escape </script BEFORE hashing

	// CSP hashes over the EXACT bytes inlined as each element body.
	scriptHash := sha256b64(viewerJS)
	styleHash := sha256b64(viewerCSS)
	dataHash := sha256b64(island)

	csp := strings.Join([]string{
		"default-src 'none'",
		fmt.Sprintf("script-src 'sha256-%s' 'sha256-%s'", scriptHash, dataHash),
		fmt.Sprintf("style-src 'sha256-%s'", styleHash),
		"img-src data:",
	}, "; ")

	esc := htmlEscape(generatedAt)
	var b strings.Builder
	w := func(s string) { b.WriteString(s) }

	w("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	w("<meta charset=\"utf-8\">\n")
	w("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	w("<meta http-equiv=\"Content-Security-Policy\" content=\"" + csp + "\">\n")
	w("<title>graffiti — " + htmlEscape(doc.Root) + "</title>\n")
	w("<!-- generated_at: " + esc + " -->\n") // OUTSIDE hashed bodies
	w("<style>" + viewerCSS + "</style>\n")
	w("</head>\n")
	w("<body data-generated-at=\"" + esc + "\">\n") // OUTSIDE hashed bodies

	w("<canvas id=\"c\"></canvas>\n")
	w("<div class=\"panel\" id=\"panel\"><div class=\"pad\">\n")
	w("<h1>" + htmlEscape(doc.Root) + "</h1><div class=\"muted\"><span id=\"nc\"></span> nodes · <span id=\"ec\"></span> edges · <span id=\"sc\"></span> sectors</div>\n")
	w("<h2>Code</h2><div id=\"cats\"></div>\n")
	w("<label class=\"row\"><input type=\"checkbox\" id=\"d3\" checked><span>3D depth (hover lift)</span></label>\n")
	w("<label class=\"row\"><input type=\"checkbox\" id=\"zones\" checked><span>show sector zones</span></label>\n")
	w("<h2>Structure <span class=\"acts\"><span id=\"exp\">expand</span> · <span id=\"col\">collapse</span></span></h2>\n")
	w("</div><div id=\"tree\"></div><div class=\"resizer\" id=\"resizer\"></div></div>\n")
	w("<div class=\"tip\" id=\"tip\"></div>\n")
	w("<button id=\"fit\" class=\"fitbtn\">⊡ fit to window (f)</button>\n")
	w("<div class=\"hint\">scroll = zoom · drag bg = pan · drag node = move · dbl-click = fit</div>\n")
	renderA11yMirror(&b, doc)

	// Data island (not executed). </script escaped above.
	w("<script type=\"application/json\" id=\"graffiti-data\">")
	w(island)
	w("</script>\n")
	w("<script>" + viewerJS + "</script>\n")
	w("</body>\n</html>\n")
	return b.String()
}

// renderA11yMirror emits the hidden, ordered DOM mirror of the graph (grouped by
// directory) so screen readers and find-in-page work over the Canvas (spec §8.7).
// Directories are sorted; nodes keep their (id-sorted) Document order.
func renderA11yMirror(b *strings.Builder, doc *graph.Document) {
	bySector := map[string][]graph.Node{}
	var order []string
	for _, n := range doc.Nodes {
		dir := n.File
		if i := strings.LastIndex(dir, "/"); i >= 0 {
			dir = dir[:i]
		} else {
			dir = "."
		}
		if _, seen := bySector[dir]; !seen {
			order = append(order, dir)
		}
		bySector[dir] = append(bySector[dir], n)
	}
	sort.Strings(order)

	b.WriteString("<nav id=\"a11y\" aria-label=\"Code graph (accessibility mirror)\">\n")
	for _, dir := range order {
		fmt.Fprintf(b, "<section><h3>%s (%d)</h3>\n<ul>\n", htmlEscape(dir), len(bySector[dir]))
		for _, n := range bySector[dir] {
			fmt.Fprintf(b, "<li>%s (%s, %s:%d)</li>\n",
				htmlEscape(n.Label), htmlEscape(string(n.Kind)), htmlEscape(n.File), n.Line)
		}
		b.WriteString("</ul>\n</section>\n")
	}
	b.WriteString("</nav>\n")
}
