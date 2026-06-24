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

// ─── Source heuristics: framework recognizers (INFERRED confidence) ─────────
//
// Covers, best-effort and per file role:
//   provides — generic router DSLs (Express/gin/chi/echo), Go net/http, Flask
//     `.route`, FastAPI/`@app.get` decorators, Spring (@*Mapping / @RequestMapping
//     with class-prefix), NestJS (@Controller prefix + @Get…/@MessagePattern/
//     @EventPattern), Kafka (@KafkaListener) / NATS / generic subscribe.
//   consumes — literal http(s):// URLs, frontend HTTP clients (fetch/axios/$fetch/
//     useFetch/HttpClient/HttpService), and queue producers (publish/emit/produce,
//     KafkaTemplate.send).
//
// React/Vue/Angular/Svelte/Nuxt files are detected as FRONTEND, where a
// `.get("/x")` call is a CONSUME (client request), not a route.

var (
	rxRoute    = regexp.MustCompile(`(?i)\.(get|post|put|delete|patch|head|options|route)\(\s*["']([^"']+)["']`)
	rxHandle   = regexp.MustCompile(`(?i)\b(?:handlefunc|handle)\(\s*["']([^"']+)["']`)
	rxSpring   = regexp.MustCompile(`@(Get|Post|Put|Delete|Patch)Mapping\(\s*["']([^"']+)["']`)
	rxSpringR  = regexp.MustCompile(`@RequestMapping\(([^)]*)\)`)
	rxReqVal   = regexp.MustCompile(`(?:value|path)\s*=\s*\{?\s*["']([^"']+)`)
	rxReqStr   = regexp.MustCompile(`["']([^"']+)["']`)
	rxReqMeth  = regexp.MustCompile(`RequestMethod\.(\w+)`)
	rxNestCtl  = regexp.MustCompile(`@Controller\(\s*["']?([^"')]*)`)
	rxNestRt   = regexp.MustCompile(`@(Get|Post|Put|Delete|Patch|All)\(\s*["']?([^"')]*)`)
	rxNestMsg  = regexp.MustCompile(`@(MessagePattern|EventPattern)\(\s*["']([^"']+)`)
	rxKafkaL   = regexp.MustCompile(`@KafkaListener\(([^)]*)\)`)
	rxKafkaTop = regexp.MustCompile(`(?:topics|destinations)\s*=\s*\{?\s*["']([^"']+)`)
	rxURL      = regexp.MustCompile(`["'](https?://[^"'\s]+)["']`)
	rxClient   = regexp.MustCompile(`(?i)\b(?:fetch|\$fetch|usefetch|axios(?:\.(?:get|post|put|delete|patch|request))?|http\.(?:get|post|put|delete|patch)|httpclient\.(?:get|post|put|delete|patch)|httpservice\.(?:get|post|put|delete|patch))\(\s*["']([^"']+)["']`)
	rxPub      = regexp.MustCompile(`(?i)\.(publish|emit|produce)\(\s*["']([^"']+)["']`)
	rxKafkaS   = regexp.MustCompile(`(?i)(?:kafkatemplate|kafkaproducer|producer)\.send\(\s*["']([^"']+)["']`)
	rxSub      = regexp.MustCompile(`(?i)\.(subscribe|consume|queuesubscribe)\(\s*["']([^"']+)["']`)
	rxFrontend = regexp.MustCompile(`(?i)from\s+["'](?:react|react-dom|vue|@angular/|svelte|next|nuxt|@tanstack/react-query)|createApp\(|defineComponent\(`)

	// Django/DRF (.py): urls.py route table
	rxDjango  = regexp.MustCompile(`(?i)\b(?:path|re_path|url)\(\s*r?["']([^"']+)["']`)
	rxReGroup = regexp.MustCompile(`\(\?P<[^>]*>[^)]*\)|\([^)]*\)`)
	// ASP.NET (.cs): attributes + minimal APIs
	rxAspAttr  = regexp.MustCompile(`\[Http(Get|Post|Put|Delete|Patch)(?:\(\s*"([^"]*)")?`)
	rxAspMap   = regexp.MustCompile(`\.Map(Get|Post|Put|Delete|Patch)\(\s*"([^"]+)"`)
	rxAspRoute = regexp.MustCompile(`\[Route\(\s*"([^"]+)"`)
	rxCsClass  = regexp.MustCompile(`class\s+(\w+?)Controller\b`)
	// Ktor (.kt): routing DSL (bare verb + path starting "/")
	rxKtor = regexp.MustCompile(`(?i)\b(get|post|put|delete|patch|route)\(\s*"(/[^"]*)"`)
	// gRPC clients: generated stub/client construction + method calls (consumes)
	rxGrpcNew  = regexp.MustCompile(`(\w+)\s*(?::=|=)\s*[\w.]*?New(\w+)Client\(`)
	rxGrpcStub = regexp.MustCompile(`(\w+)\s*=\s*[\w.]*?(\w+)Stub\(`)
	rxCall     = regexp.MustCompile(`\b(\w+)\.(\w+)\(`)
)

// fileRole classifies a source file as "frontend" (HTTP-client consumer) or
// "backend" (route provider) so the same `.get("/x")` pattern is read correctly.
func fileRole(rel string, data []byte) string {
	switch strings.ToLower(filepathExt(rel)) {
	case ".tsx", ".jsx", ".vue", ".svelte":
		return "frontend"
	case ".ts", ".js", ".mjs":
		if rxFrontend.Match(data) {
			return "frontend"
		}
	}
	return "backend"
}

func filepathExt(p string) string {
	if i := strings.LastIndexByte(p, '.'); i >= 0 {
		return p[i:]
	}
	return ""
}

func meth(v string) string {
	if strings.EqualFold(v, "route") || strings.EqualFold(v, "all") {
		return "GET"
	}
	return strings.ToUpper(v)
}

func joinPath(prefix, p string) string {
	prefix, p = strings.Trim(prefix, "/"), strings.Trim(p, "/")
	full := prefix
	if p != "" {
		if full != "" {
			full += "/" + p
		} else {
			full = p
		}
	}
	return "/" + full
}

func validTopic(s string) bool { return s != "" && !strings.HasPrefix(s, "/") }

// djangoPath cleans a Django/DRF route pattern (path() or re_path() regex) into a
// match template: strip ^$ anchors, collapse regex groups to {} (named groups too).
// `<int:id>`-style converters are handled later by normPath.
func djangoPath(p string) string {
	p = strings.TrimPrefix(p, "^")
	p = strings.TrimSuffix(p, "$")
	return rxReGroup.ReplaceAllString(p, "{}")
}

// aspPath substitutes ASP.NET route tokens: [controller] → the controller class
// name (without the "Controller" suffix), [action] → dropped.
func aspPath(p, ctl string) string {
	if p == "" {
		return ""
	}
	p = strings.ReplaceAll(p, "[controller]", ctl)
	p = strings.ReplaceAll(p, "[action]", "")
	return p
}

// scanGrpcClients pre-scans a Go/Python file for generated gRPC client/stub
// constructions, returning var-name → service-name so method calls on those vars
// can be emitted as rpc consumes.
func scanGrpcClients(data []byte, ext string) map[string]string {
	vars := map[string]string{}
	if ext != ".go" && ext != ".py" {
		return vars
	}
	re := rxGrpcNew
	if ext == ".py" {
		re = rxGrpcStub
	}
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		for _, m := range re.FindAllStringSubmatch(sc.Text(), -1) {
			if m[2] != "" {
				vars[m[1]] = m[2]
			}
		}
	}
	return vars
}

var grpcSkipMethods = map[string]bool{"Close": true, "String": true, "Error": true, "Reset": true}

func scanSource(rel string, data []byte) (provides, consumes []graph.Endpoint) {
	frontend := fileRole(rel, data) == "frontend"
	ext := strings.ToLower(filepathExt(rel))
	grpcVars := scanGrpcClients(data, ext)
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	line := 0
	prefix := "" // current NestJS @Controller / Spring / ASP.NET class prefix
	csCtl := ""  // ASP.NET controller name (for the [controller] token)
	if ext == ".cs" {
		if m := rxCsClass.FindSubmatch(data); m != nil { // pre-scan: [Route] often precedes the class line
			csCtl = strings.ToLower(string(m[1]))
		}
	}
	add := func(dst *[]graph.Endpoint, kind graph.EndpointKind, key, disp, source string, ln int) {
		*dst = append(*dst, graph.Endpoint{
			Kind: kind, Key: key, Display: disp, File: rel, Line: ln,
			Confidence: graph.ConfInferred, Source: source,
		})
	}
	httpC := func(method, raw string) {
		p := stripHost(raw)
		add(&consumes, graph.EPHTTP, httpKey(method, p), strings.ToUpper(method)+" "+p, "client", line)
	}

	for sc.Scan() {
		line++
		t := sc.Text()

		// ── consumes (every role) ──
		clientHit := false
		for _, m := range rxClient.FindAllStringSubmatch(t, -1) {
			clientHit = true
			httpC("GET", m[1])
		}
		for _, m := range rxURL.FindAllStringSubmatch(t, -1) {
			httpC("GET", m[1])
		}
		for _, m := range rxPub.FindAllStringSubmatch(t, -1) {
			if validTopic(m[2]) {
				add(&consumes, graph.EPQueue, m[2], "publish "+m[2], "literal", line)
			}
		}
		for _, m := range rxKafkaS.FindAllStringSubmatch(t, -1) {
			if validTopic(m[1]) {
				add(&consumes, graph.EPQueue, m[1], "publish "+m[1], "kafka", line)
			}
		}
		if len(grpcVars) > 0 {
			for _, m := range rxCall.FindAllStringSubmatch(t, -1) {
				if svc, ok := grpcVars[m[1]]; ok && !grpcSkipMethods[m[2]] {
					key := svc + "." + m[2]
					add(&consumes, graph.EPRPC, key, key, "grpc", line)
				}
			}
		}

		if frontend {
			// In frontend files, router-style calls are client CONSUMES, not routes.
			for _, m := range rxRoute.FindAllStringSubmatch(t, -1) {
				if strings.HasPrefix(m[2], "/") {
					httpC(meth(m[1]), m[2])
				}
			}
			continue
		}

		// ── backend provides ──
		if !clientHit {
			for _, m := range rxRoute.FindAllStringSubmatch(t, -1) {
				if strings.HasPrefix(m[2], "/") {
					add(&provides, graph.EPHTTP, httpKey(meth(m[1]), m[2]), strings.ToUpper(meth(m[1]))+" "+m[2], "route", line)
				}
			}
		}
		for _, m := range rxHandle.FindAllStringSubmatch(t, -1) {
			if strings.HasPrefix(m[1], "/") {
				add(&provides, graph.EPHTTP, httpKey("GET", m[1]), "GET "+m[1], "route", line)
			}
		}
		for _, m := range rxSpring.FindAllStringSubmatch(t, -1) {
			p := joinPath(prefix, m[2])
			add(&provides, graph.EPHTTP, httpKey(m[1], p), strings.ToUpper(m[1])+" "+p, "route", line)
		}
		for _, m := range rxSpringR.FindAllStringSubmatch(t, -1) {
			path := ""
			if pm := rxReqVal.FindStringSubmatch(m[1]); pm != nil {
				path = pm[1]
			} else if pm := rxReqStr.FindStringSubmatch(m[1]); pm != nil {
				path = pm[1]
			}
			if path == "" {
				continue
			}
			if mm := rxReqMeth.FindStringSubmatch(m[1]); mm != nil {
				method := strings.ToUpper(mm[1])
				full := joinPath(prefix, path)
				add(&provides, graph.EPHTTP, httpKey(method, full), method+" "+full, "route", line)
			} else {
				prefix = strings.Trim(path, "/") // class-level @RequestMapping → prefix
			}
		}
		if m := rxNestCtl.FindStringSubmatch(t); m != nil {
			prefix = strings.Trim(m[1], "/")
		}
		for _, m := range rxNestRt.FindAllStringSubmatch(t, -1) {
			method, full := meth(m[1]), joinPath(prefix, m[2])
			add(&provides, graph.EPHTTP, httpKey(method, full), method+" "+full, "route", line)
		}
		for _, m := range rxNestMsg.FindAllStringSubmatch(t, -1) {
			add(&provides, graph.EPQueue, m[2], strings.ToLower(m[1])+" "+m[2], "route", line)
		}
		for _, m := range rxSub.FindAllStringSubmatch(t, -1) {
			if validTopic(m[2]) {
				add(&provides, graph.EPQueue, m[2], "subscribe "+m[2], "literal", line)
			}
		}
		for _, m := range rxKafkaL.FindAllStringSubmatch(t, -1) {
			if tm := rxKafkaTop.FindStringSubmatch(m[1]); tm != nil {
				add(&provides, graph.EPQueue, tm[1], "subscribe "+tm[1], "kafka", line)
			}
		}

		// Django / DRF (.py): urls.py route table → http provides.
		if ext == ".py" {
			for _, m := range rxDjango.FindAllStringSubmatch(t, -1) {
				if p := djangoPath(m[1]); p != "" {
					full := joinPath(prefix, p)
					add(&provides, graph.EPHTTP, httpKey("GET", full), "GET "+full, "route", line)
				}
			}
		}
		// ASP.NET (.cs): [controller] class, [Route] prefix, [HttpGet] attrs, MapGet.
		if ext == ".cs" {
			if m := rxCsClass.FindStringSubmatch(t); m != nil {
				csCtl = strings.ToLower(m[1])
			}
			for _, m := range rxAspRoute.FindAllStringSubmatch(t, -1) {
				prefix = strings.Trim(aspPath(m[1], csCtl), "/")
			}
			for _, m := range rxAspAttr.FindAllStringSubmatch(t, -1) {
				full := joinPath(prefix, aspPath(m[2], csCtl))
				add(&provides, graph.EPHTTP, httpKey(m[1], full), strings.ToUpper(m[1])+" "+full, "route", line)
			}
			for _, m := range rxAspMap.FindAllStringSubmatch(t, -1) {
				add(&provides, graph.EPHTTP, httpKey(m[1], m[2]), strings.ToUpper(m[1])+" "+m[2], "route", line)
			}
		}
		// Ktor (.kt): routing DSL.
		if ext == ".kt" {
			for _, m := range rxKtor.FindAllStringSubmatch(t, -1) {
				add(&provides, graph.EPHTTP, httpKey(meth(m[1]), m[2]), strings.ToUpper(meth(m[1]))+" "+m[2], "route", line)
			}
		}
	}
	return provides, consumes
}
