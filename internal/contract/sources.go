package contract

import (
	"bufio"
	"bytes"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/amazopic/graffiti/internal/graph"
)

// ─── Normalization ───────────────────────────────────────────────────────

var rxPathParam = regexp.MustCompile(`\{[^/}]*\}|:[A-Za-z_]\w*|<[^/>]*>`)

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// normPath collapses a URL path to a match template: params → {}, numeric
// segments → {}, query/fragment stripped, trailing slash trimmed.
func normPath(p string) string {
	if i := strings.IndexAny(p, "?#"); i >= 0 {
		p = p[:i]
	}
	p = rxPathParam.ReplaceAllString(p, "{}")
	segs := strings.Split(p, "/")
	for i, s := range segs {
		if isAllDigits(s) {
			segs[i] = "{}"
		}
	}
	p = strings.Join(segs, "/")
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if len(p) > 1 {
		p = strings.TrimRight(p, "/")
	}
	if p == "" {
		p = "/"
	}
	return p
}

func httpKey(method, path string) string {
	m := strings.ToUpper(strings.TrimSpace(method))
	if m == "" {
		m = "GET"
	}
	return m + " " + normPath(path)
}

// stripHost turns "http://host:port/a/b?q" into "/a/b". A bare host with no path
// becomes "/".
func stripHost(u string) string {
	if i := strings.Index(u, "://"); i >= 0 {
		rest := u[i+3:]
		if j := strings.IndexByte(rest, '/'); j >= 0 {
			return rest[j:]
		}
		return "/"
	}
	return u
}

// ─── graffiti.contract.json (explicit override) ───────────────────────────

type cjEntry struct {
	Kind    string `json:"kind"`
	Method  string `json:"method"`
	Path    string `json:"path"`
	Name    string `json:"name"`   // rpc Service.Method
	Topic   string `json:"topic"`  // queue
	Symbol  string `json:"symbol"` // lib
	Key     string `json:"key"`    // optional explicit key override
	Display string `json:"display"`
	Node    string `json:"node"`
	File    string `json:"file"`
	Line    int    `json:"line"`
}

type cjFile struct {
	Provides []cjEntry `json:"provides"`
	Consumes []cjEntry `json:"consumes"`
}

func (e cjEntry) toEndpoint(declFile string) (graph.Endpoint, bool) {
	kind := graph.EndpointKind(strings.ToLower(e.Kind))
	if !graph.ValidEndpointKinds[kind] {
		return graph.Endpoint{}, false
	}
	key, disp := e.Key, e.Display
	if key == "" {
		switch kind {
		case graph.EPHTTP:
			key = httpKey(e.Method, e.Path)
		case graph.EPRPC:
			key = e.Name
		case graph.EPQueue:
			key = e.Topic
		case graph.EPLib:
			key = e.Symbol
		}
	}
	if key == "" {
		return graph.Endpoint{}, false
	}
	if disp == "" {
		disp = key
	}
	file := e.File
	if file == "" {
		file = declFile
	}
	return graph.Endpoint{
		Kind: kind, Key: key, Display: disp, Node: e.Node,
		File: file, Line: e.Line, Confidence: graph.ConfExtracted, Source: "contract",
	}, true
}

func parseContractJSON(rel string, data []byte) (provides, consumes []graph.Endpoint) {
	var f cjFile
	if json.Unmarshal(data, &f) != nil {
		return nil, nil
	}
	for _, e := range f.Provides {
		if ep, ok := e.toEndpoint(rel); ok {
			provides = append(provides, ep)
		}
	}
	for _, e := range f.Consumes {
		if ep, ok := e.toEndpoint(rel); ok {
			consumes = append(consumes, ep)
		}
	}
	return provides, consumes
}

// ─── OpenAPI / Swagger (declared HTTP provides) ────────────────────────────

var httpMethods = map[string]bool{
	"get": true, "post": true, "put": true, "delete": true,
	"patch": true, "head": true, "options": true, "trace": true,
}

func parseOpenAPI(rel string, data []byte) []graph.Endpoint {
	var doc struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if json.Unmarshal(data, &doc) != nil {
		return nil
	}
	var out []graph.Endpoint
	for path, ops := range doc.Paths {
		for method := range ops {
			if !httpMethods[strings.ToLower(method)] {
				continue
			}
			m := strings.ToUpper(method)
			out = append(out, graph.Endpoint{
				Kind: graph.EPHTTP, Key: httpKey(m, path), Display: m + " " + path,
				File: rel, Line: 0, Confidence: graph.ConfExtracted, Source: "openapi",
			})
		}
	}
	return out
}

// ─── .proto (declared gRPC provides) ───────────────────────────────────────

var (
	rxService = regexp.MustCompile(`^\s*service\s+(\w+)\s*\{`)
	rxRPC     = regexp.MustCompile(`^\s*rpc\s+(\w+)\s*\(`)
)

func parseProto(rel string, data []byte) []graph.Endpoint {
	var out []graph.Endpoint
	svc := ""
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	line := 0
	for sc.Scan() {
		line++
		t := sc.Text()
		if m := rxService.FindStringSubmatch(t); m != nil {
			svc = m[1]
			continue
		}
		if svc == "" {
			continue
		}
		if strings.Contains(t, "}") && !strings.Contains(t, "rpc") {
			svc = ""
		}
		if m := rxRPC.FindStringSubmatch(t); m != nil {
			key := svc + "." + m[1]
			out = append(out, graph.Endpoint{
				Kind: graph.EPRPC, Key: key, Display: key,
				File: rel, Line: line, Confidence: graph.ConfExtracted, Source: "proto",
			})
		}
	}
	return out
}

// ─── Source heuristics (router DSLs, literal URLs, queue calls) ────────────

var (
	rxRoute  = regexp.MustCompile(`(?i)\.(get|post|put|delete|patch|head|options)\(\s*["']([^"']+)["']`)
	rxHandle = regexp.MustCompile(`(?i)\b(?:handlefunc|handle)\(\s*["']([^"']+)["']`)
	rxSpring = regexp.MustCompile(`@(Get|Post|Put|Delete|Patch)Mapping\(\s*["']([^"']+)["']`)
	rxURL    = regexp.MustCompile(`["'](https?://[^"'\s]+)["']`)
	rxPub    = regexp.MustCompile(`(?i)\.(publish|emit|produce)\(\s*["']([^"']+)["']`)
	rxSub    = regexp.MustCompile(`(?i)\.(subscribe|consume|queuesubscribe)\(\s*["']([^"']+)["']`)
)

func scanSource(rel string, data []byte) (provides, consumes []graph.Endpoint) {
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	line := 0
	add := func(dst *[]graph.Endpoint, kind graph.EndpointKind, key, disp, source string, ln int) {
		*dst = append(*dst, graph.Endpoint{
			Kind: kind, Key: key, Display: disp, File: rel, Line: ln,
			Confidence: graph.ConfInferred, Source: source,
		})
	}
	for sc.Scan() {
		line++
		t := sc.Text()

		for _, m := range rxRoute.FindAllStringSubmatch(t, -1) {
			if strings.HasPrefix(m[2], "/") {
				add(&provides, graph.EPHTTP, httpKey(m[1], m[2]), strings.ToUpper(m[1])+" "+m[2], "route", line)
			}
		}
		for _, m := range rxSpring.FindAllStringSubmatch(t, -1) {
			add(&provides, graph.EPHTTP, httpKey(m[1], m[2]), strings.ToUpper(m[1])+" "+m[2], "route", line)
		}
		for _, m := range rxHandle.FindAllStringSubmatch(t, -1) {
			if strings.HasPrefix(m[1], "/") {
				add(&provides, graph.EPHTTP, httpKey("GET", m[1]), "GET "+m[1], "route", line)
			}
		}
		for _, m := range rxSub.FindAllStringSubmatch(t, -1) {
			add(&provides, graph.EPQueue, m[2], "subscribe "+m[2], "literal", line)
		}

		for _, m := range rxURL.FindAllStringSubmatch(t, -1) {
			p := stripHost(m[1])
			add(&consumes, graph.EPHTTP, httpKey("GET", p), "GET "+m[1], "literal", line)
		}
		for _, m := range rxPub.FindAllStringSubmatch(t, -1) {
			add(&consumes, graph.EPQueue, m[2], "publish "+m[2], "literal", line)
		}
	}
	return provides, consumes
}
