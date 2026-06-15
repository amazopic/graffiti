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
	"github.com/evgeniy-achin/graffiti/internal/layout"
)

// Viewer assets are inlined verbatim into map.html and hashed for the CSP. They
// are pure data here so the emitted sha256 hashes are a function only of these
// (deterministic) bytes — never of generated_at or scene content.
//
//go:embed viewer/app.css
var viewerCSS string

//go:embed viewer/app.js
var viewerJS string

// Keep embed referenced even if a future refactor drops one of the strings.
var _ embed.FS

// htmlEscape escapes interpolated labels/paths for safe inlining into HTML body
// text (XSS-safe self-contained file, spec §8.1/§5/§6).
func htmlEscape(s string) string { return html.EscapeString(s) }

// escapeScriptClose makes a JSON byte string safe to inline inside a
// <script type="application/json"> island: it escapes every '<' to the JSON
// unicode escape <, which both (a) prevents a "</script" sequence in any
// label/path from prematurely closing the island and (b) is fully JSON-legal so
// the island still round-trips via JSON.parse. Applied to the bytes BEFORE
// hashing, so the data-island sha256 stays a pure function of the scene.
func escapeScriptClose(jsonBytes []byte) string {
	return strings.ReplaceAll(string(jsonBytes), "<", `<`)
}

func sha256b64(s string) string {
	sum := sha256.Sum256([]byte(s))
	return base64.StdEncoding.EncodeToString(sum[:])
}

// WriteMapHTML renders map.html and writes it to <root>/.graffiti/map.html, next
// to map.json and MAP.md. generatedAt is read off doc.GeneratedAt (single source
// of truth, stamped by build.Assemble).
func WriteMapHTML(doc *graph.Document, an analyze.Analysis, scene layout.Scene, root string) error {
	dir := filepath.Join(root, ".graffiti")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	out := RenderMapHTML(doc, an, scene, doc.GeneratedAt)
	return os.WriteFile(filepath.Join(dir, "map.html"), []byte(out), 0o644)
}

// RenderMapHTML builds the single self-contained, offline, CSP-safe map.html
// (spec §8.1–§8.8). It is a pure function of (doc, an, scene, generatedAt):
// generatedAt appears ONLY in an HTML comment + a body data-attribute, both
// OUTSIDE every hashed body, so the CSP hashes never vary with time or scene.
// The inlined CSS/JS come from go:embed; the scene ships as a compact columnar
// JSON island with the </script sequence escaped. The hidden ordered <nav id=a11y>
// mirrors the districts for screen readers (spec §8.7).
func RenderMapHTML(doc *graph.Document, an analyze.Analysis, scene layout.Scene, generatedAt string) string {
	cs := toColumnar(scene)
	dataJSON, err := json.Marshal(cs) // deterministic for a fixed struct type
	if err != nil {
		panic(fmt.Sprintf("marshal columnar scene: %v", err)) // ints/strings only
	}
	island := escapeScriptClose(dataJSON) // escape </script BEFORE hashing

	// CSP hashes over the EXACT bytes inlined as each element body.
	scriptHash := sha256b64(viewerJS)
	styleHash := sha256b64(viewerCSS)
	dataHash := sha256b64(island) // belt-and-braces (island isn't executed)

	csp := strings.Join([]string{
		"default-src 'none'",
		fmt.Sprintf("script-src 'sha256-%s' 'sha256-%s'", scriptHash, dataHash),
		fmt.Sprintf("style-src 'sha256-%s'", styleHash),
		"img-src data:",
	}, "; ")

	var b strings.Builder
	w := func(s string) { b.WriteString(s) }

	w("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	w("<meta charset=\"utf-8\">\n")
	w("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	w("<meta http-equiv=\"Content-Security-Policy\" content=\"" + csp + "\">\n")
	w("<title>Graffiti Districts — " + htmlEscape(doc.Root) + "</title>\n")
	w("<!-- generated_at: " + htmlEscape(generatedAt) + " -->\n") // OUTSIDE hashed bodies
	w("<style>" + viewerCSS + "</style>\n")
	w("</head>\n")
	w("<body data-generated-at=\"" + htmlEscape(generatedAt) + "\">\n") // OUTSIDE hashed bodies

	// Top bar (the 0 API calls · $0 promise, spec §8.3).
	w("<div id=\"top\"><span class=\"title\">Districts — " + htmlEscape(doc.Root) + "</span>")
	w("<span class=\"cost\">0 API calls · $0</span></div>\n")

	w("<div id=\"wrap\">\n")
	renderRail(&b, doc, an)
	w("<div id=\"stage\"><canvas id=\"canvas\"></canvas>")
	w("<div id=\"inspect\"></div>")
	renderA11yMirror(&b, doc)
	w("</div>\n") // #stage
	w("</div>\n") // #wrap

	// Data island (data block; not executed). </script escaped above.
	w("<script type=\"application/json\" id=\"graffiti-data\">")
	w(island)
	w("</script>\n")
	w("<script>" + viewerJS + "</script>\n")
	w("</body>\n</html>\n")
	return b.String()
}

// renderRail emits the left rail: search, Start here (3 questions), Landmarks
// (god nodes), Confidence legend. Deterministic; labels HTML-escaped.
func renderRail(b *strings.Builder, doc *graph.Document, an analyze.Analysis) {
	byID := make(map[string]graph.Node, len(doc.Nodes))
	for _, n := range doc.Nodes {
		byID[n.ID] = n
	}
	commOf := func(id string) int { return byID[id].Community }

	b.WriteString("<aside id=\"rail\">\n")
	b.WriteString("<input id=\"search\" type=\"text\" placeholder=\"Search districts…\" aria-label=\"Search districts\">\n")

	b.WriteString("<h2>Start here</h2>\n<ol>\n")
	for _, q := range an.Questions {
		b.WriteString("<li>" + htmlEscape(q) + "</li>\n")
	}
	b.WriteString("</ol>\n")

	b.WriteString("<h2>Landmarks</h2>\n")
	if len(an.GodNodes) == 0 {
		b.WriteString("<p>None.</p>\n")
	} else {
		for _, g := range an.GodNodes {
			fmt.Fprintf(b, "<button class=\"chip\" data-comm=\"%d\">%s — touched by %d things</button>\n",
				commOf(g.ID), htmlEscape(g.Label), g.Degree)
		}
	}

	b.WriteString("<h2>Confidence</h2>\n<div id=\"legend\">\n")
	b.WriteString("<div class=\"row\"><span class=\"swatch extracted\"></span>EXTRACTED — definite</div>\n")
	b.WriteString("<div class=\"row\"><span class=\"swatch inferred\"></span>INFERRED — inferred</div>\n")
	b.WriteString("<div class=\"row\"><span class=\"swatch ambiguous\"></span>AMBIGUOUS — guessed; verify</div>\n")
	b.WriteString("</div>\n")
	b.WriteString("</aside>\n")
}

// renderA11yMirror emits the hidden, ordered DOM mirror of districts (→ members)
// so screen readers and find-in-page work over the Canvas (spec §8.7). Ordered
// by community id; members are already sorted by cluster.NameCommunities.
func renderA11yMirror(b *strings.Builder, doc *graph.Document) {
	byID := make(map[string]graph.Node, len(doc.Nodes))
	for _, n := range doc.Nodes {
		byID[n.ID] = n
	}
	comms := append([]graph.Community(nil), doc.Communities...)
	sort.Slice(comms, func(i, j int) bool { return comms[i].ID < comms[j].ID })

	b.WriteString("<nav id=\"a11y\" aria-label=\"Districts (accessibility mirror)\">\n")
	for _, c := range comms {
		fmt.Fprintf(b, "<section><h3>%s (%d things)</h3>\n<ul>\n", htmlEscape(c.Label), len(c.Members))
		for _, id := range c.Members {
			n := byID[id]
			fmt.Fprintf(b, "<li>%s (%s, %s:%d)</li>\n",
				htmlEscape(n.Label), htmlEscape(string(n.Kind)), htmlEscape(n.File), n.Line)
		}
		b.WriteString("</ul>\n</section>\n")
	}
	b.WriteString("</nav>\n")
}
