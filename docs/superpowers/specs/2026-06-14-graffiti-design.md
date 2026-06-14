# graffiti — Design Spec (v1 / MVP)

**Status:** Draft for review
**Date:** 2026-06-14
**Owner:** evgeniy.achin@gmail.com

---

## 1. One-liner & positioning

> **One command turns your repository into a map your AI coding assistant reads instead of blindly grepping — so it answers faster, cheaper, and right the first time.**

graffiti is a single self-contained binary that scans a code repository locally, builds a directed knowledge graph of its symbols and relationships, and emits three artifacts plus a deep, automatic integration with Claude Code. No Python, no API keys, no cost, fully offline.

The product is delivered to the developer *through their AI assistant*: after a one-time `graffiti init`, Claude Code prefers a scoped graph query over grepping and re-reading whole files on every future codebase question.

## 2. Target audience

"Vibe coders" — people who build software primarily by directing an AI coding assistant rather than hand-writing much code. They live inside a chat/IDE window, do not manage language runtimes or virtualenvs, and bounce off `command not found` / `ModuleNotFoundError` / PATH setup. The product must work from a single download with zero environment management.

## 3. The three wedges (what makes this "more effective")

1. **Radical simplicity** — one command, zero config, one success signal. A single static binary with no runtime dependencies.
2. **Graph visibility** — the default visual artifact is a *readable architecture map* (communities + call-flow), not a force-directed hairball; it scales by clustering, not collapsing, and opens offline with no CDN.
3. **Deep auto-integration** — an always-on loop installed into Claude Code so the assistant uses the graph automatically and keeps it fresh after edits. A single deterministic binary command replaces any multi-step procedure the model would otherwise have to execute.

These three reinforce one moat: a **deterministic, local, zero-cost extraction core** plus an **LLM-free query path** (the assistant's own model does all natural-language reasoning over a scoped subgraph — no second key, no second bill).

## 4. Goals / Non-goals (v1)

**Goals**
- `graffiti .` builds a graph for a real repo in seconds, with **0 API calls / $0**, fully offline.
- Six source languages: **Python, JavaScript/TypeScript, Go, Rust, Java**, plus **Markdown**.
- Three output artifacts: `map.html`, `MAP.md`, `map.json` (with a published JSON Schema).
- LLM-free `graffiti query` returning a scoped subgraph under a token budget.
- `graffiti init` wires Claude Code (skill + always-on block + optional hook).
- One-command install of a single static binary on macOS/Linux/Windows.
- Optional **workspace** federation: link separate project graphs (e.g. frontend + backend + llm-router) with deterministic, zero-key cross-project edges — **without merging** them into one graph (see §16).

**Non-goals (explicitly deferred)**
- Multi-host install matrix (only Claude Code in v1; auto-detect later).
- Any LLM backend in the default path; semantic extraction of docs/PDFs/images; multimodal.
- A **merged** cross-repo "global" graph / second materialized store (the workspace in §16 is a computed overlay, not a merge); external databases (Neo4j/FalkorDB/Postgres), PR triage, HTTP MCP server with auth.
- Obsidian vault / wiki / SVG / GraphML / Canvas exports.
- Fuzzy/LLM entity dedup; README translations.
- **3D graph rendering** — considered and rejected for v1/default: it worsens legibility (occlusion, depth ambiguity), breaks the baked-deterministic Canvas2D layout (§8/§14), and blows the offline single-file / CSP / <1.5 MB budget. A fenced, optional "showcase" 3D view may be revisited later as a marketing artifact — never the default.

## 5. Architecture

A clean functional pipeline; stages communicate via plain structs/graph with no shared state.

```
scan → parse → build → cluster → analyze → render
                                   │
                                   └── query (separate LLM-free read path) ── mcp
```

| Stage | Responsibility | In → Out |
|-------|----------------|----------|
| `scan` | Discover & classify files; honor `.gitignore`; filter to supported extensions | dir → `[]FileRef` |
| `parse` | tree-sitter (via wazero, WASM grammars) AST walk; two-pass extraction | `FileRef` → `{nodes, edges}` |
| `build` | Validate against schema; assemble **directed** graph; deterministic IDs; merge-not-replace; content-hash cache | `[]extraction` → `Graph` |
| `cluster` | Community detection (Louvain) | `Graph` → `Graph` (+ `community` per node) |
| `analyze` | God nodes, surprising connections, suggested questions, import cycles | `Graph` → `Analysis` |
| `render` | Emit `map.json` + `MAP.md` + self-contained `map.html` | `Graph, Analysis` → 3 files |
| `query` | LLM-free BFS/DFS + IDF scoring → token-budgeted text subgraph | `Graph, question` → text |
| `mcp` | Expose `query`/`get_node`/`neighbors`/`path` as MCP tools over stdio | — |
| `integrate` | `graffiti init`: install Claude Code skill + CLAUDE.md block + optional hook | — |

### Extraction: two passes
- **Pass 1 (per file):** walk the AST → emit definition nodes (function/method/class/module) and structural edges (`imports` = EXTRACTED, `contains`, `inherits`/`implements`). Unresolved call sites are stashed as raw calls.
- **Pass 2 (cross-file):** build a global `label → [node_id]` index; resolve raw calls. A call backed by a matching `imports` edge is promoted to **EXTRACTED**, otherwise **INFERRED**. Ambiguous common names defined in ≥2 files with no disambiguating import are dropped (prevents "god node" inflation). Genuinely uncertain edges are tagged **AMBIGUOUS** and surfaced for review in `MAP.md`.

## 6. Data model

No database. The single source of truth is one file: `.graffiti/map.json`. The in-memory graph is **directed from day one** (no undirected-with-direction-in-attributes hack).

**Node**
```
{ "id": string,        // deterministic normalized slug (NFC + casefold + \w-collapse)
  "label": string,     // human name
  "kind": "function|method|class|module|file|doc|concept",
  "file": string,      // repo-relative path
  "line": number,      // 1-based
  "community": number  // assigned after clustering; -1 before
}
```
**Edge**
```
{ "from": string,      // node id
  "to": string,        // node id
  "relation": "calls|imports|inherits|implements|references|contains",
  "confidence": "EXTRACTED|INFERRED|AMBIGUOUS" }
```
**Document**
```
{ "version": string,
  "generated_at": string,   // RFC3339, stamped by the binary at write time
  "root": string,
  "nodes": [Node],
  "edges": [Edge],
  "communities": [ { "id": number, "label": string, "members": [string] } ] }
```

A JSON Schema is published at `schema/map.schema.json` and shipped so an assistant can validate/generate graphs reliably.

**Incremental update (`graffiti update`):** SHA256 content hash per file in `.graffiti/cache/`; unchanged files skip re-extraction. Build is **merge-not-replace** with an anti-shrink guard (never silently reduce node count without an explicit prune).

## 7. Query path (LLM-free)

`graffiti query "<question>"`:
1. Load `map.json`, rebuild the directed graph.
2. Tokenize the question; score nodes by IDF over labels/kinds.
3. Pick top seed nodes; BFS/DFS expand neighbors until a **token budget** (default ~2000) is reached.
4. Serialize the subgraph to compact text (nodes with file:line, edges with relation+confidence).

No inference call is made by graffiti. The Claude Code model reasons over the returned text. The same retrieval is exposed via MCP (`serve`) so Claude calls it natively.

## 8. Visual artifact (`map.html`)

The default artifact is a **readable architecture map — a "city map" of the codebase ("Districts")**, not a force-directed hairball. A vibe coder should be able to *name what the program does* before touching a control: the screen reads like a subway/city map of labeled neighborhoods, not a network of dots.

### 8.1 Self-contained, offline, CSP-safe
- All JS/CSS/data inlined via `embed.FS` into **one** `map.html`; opens from `file://` offline with **no CDN, no external `<link>`/`<script src>`**.
- **Zero third-party JS.** The only "library" is our own ~25–30KB hand-written vanilla-ES renderer (no modules, no `eval`, no `new Function`), embedded from `render/viewer/`.
- Data ships as a `<script type="application/json" id="graffiti-data">` island, read via `getElementById` + `JSON.parse` (never `eval`, never a JS literal). The `<` and closing-`</script>` sequences are escaped.
- A **strict `sha256` meta CSP** is emitted by the binary at write time (`default-src 'none'; script-src 'sha256-…'; style-src 'sha256-…'; img-src data:`). System font stack only (no font blob).
- HTML/JS emitters escape all interpolated labels/paths (XSS-safe self-contained file, per §5/§6).

### 8.2 Rendering & layout — Canvas2D, layout baked in Go
- **Rendering tech: a single full-window Canvas2D layer** (HiDPI via `devicePixelRatio`). Not SVG (collapses around 10–20k interactive elements on weak hardware), not WebGL for v1 (glyph atlas, picking buffer, context-loss, and shader strings fight byte-determinism and CSP cleanliness; revisit only above ~150k nodes). SVG is reserved for the rail, tooltips, search, and the accessibility mirror.
- **Layout is precomputed in Go and baked as integer coordinates** into the data island — the only way to satisfy the §14 byte-identical guarantee (a browser force simulation is seed/timing/float dependent and can never be byte-stable; it also reintroduces settling-jitter hairball). The browser does **no layout and no physics** — only camera transform, viewport cull, batched paint, and analytic hit-test.
- The Go layout pipeline is hierarchical and deterministic (seeded, sorted, integer-quantized): a **squarified treemap** packs the 10–40 community boxes (no top-tier force in v1); **file-shelf** layout lays files out inside an opened district; a bounded fixed-seed pass places symbols. Inter-community edges are **aggregated into one bundle per ordered `(A→B)` pair** and routed as orthogonal/elbow polylines in the gutters (transit-map look). Bundle control points and community hulls are baked in Go.
- Renderer rules (measured): **batch primitives by color** into single paths (per-node fill is ~6× slower); **labels are the cost center**, so cap them at ~300–500/frame with hide-on-collision and none below a pixel-size threshold; redraw only on camera change (`requestAnimationFrame` + dirty flag); cache the static district layer to an offscreen canvas. The **pick index** (uniform grid/quadtree) is rebuilt in-JS on load from the baked coords (smaller file, stays deterministic).

### 8.3 Default view
On open (already framed, no spinner, no settling), the screen shows **8–40 rounded, tinted, named district boxes** tiled like neighborhoods, on a **light theme** (deterministic and doc/PR-embeddable). Each box shows its human label (e.g. *Auth & Sessions*, *HTTP Routing*, *Parser*), a kind hint, and a count ("31 things"). **Box area** encodes subsystem size; **border weight** encodes centrality. A small number (typically 6–15) of **thick labeled flow-arrows** connect districts, arrow thickness = bundled edge count ("142 calls"). A few **god nodes** appear as starred **landmark pins** with a halo on their district; **surprising cross-community links** are dashed, brightly-tinted gutter arcs. No symbol-level dots at this tier. A left rail holds: search, a **Start here** list of the 3 suggested questions, a **Landmarks** (god-node) list, a **Surprising links** toggle, kind/confidence filters, and a legend. The top bar reinforces the `0 API calls · $0` promise.

**District names** come from a deterministic heuristic in the cluster/analyze stage (§5): the **dominant source directory** of the community's members (e.g. `internal/auth` → "Auth"), falling back to the **most-central member's label** when the directory is generic/root. The metaphor is only as good as the names, so this heuristic is part of v1 and is covered by the determinism test.

### 8.4 Semantic zoom (community → file → symbol)
Three tiers, driven by wheel/pinch zoom and click-to-drill (Esc / breadcrumb to back out):
- **Tier 1 — Districts (default):** named community boxes + bundled labeled flows + ~7 god-node pins (rest in the Landmarks list) + dashed surprising-link arcs. Visible primitives are `O(communities)` regardless of total node count, so the first frame is never a mesh.
- **Tier 2 — District interior:** the opened district shows its files as shelves with top symbols; siblings dim; edges to siblings appear as labeled boundary stubs. A district exceeding ~1500 visible cells sub-shelves and lazily paints only on-screen shelves.
- **Tier 3 — Symbol:** an individual symbol with `file:line`, kind icon, and its direct callers/callees (edges colored by relation); everything else desaturates for focus.

Click a bundled arrow to list the real underlying edges (true endpoints + true count — bundling never invents a route). Click a Start-here question or a Landmark/surprising-link chip to fly the camera and highlight the matching subgraph, closing the loop with `graffiti query`.

### 8.5 Encoding god nodes, surprises, and confidence
- **God nodes** (high-degree hubs from `analyze`): starred halo pins, capped at ~7 on the map; the rest in the Landmarks list with **degree in words** ("touched by 47 things — change carefully").
- **Surprising connections** (cross-community edges): dashed brightly-tinted arcs + an isolate toggle.
- **Confidence** (line style + plain-English legend): **EXTRACTED** = solid ("definite"); **INFERRED** = thin ("inferred"); **AMBIGUOUS** = dashed + muted + "?" ("guessed — verify"). **Relation** is a fixed, colorblind-safe hue per type. The assistant's uncertainty is always visible, never hidden (honesty-first, per §16.2's under-link policy).

### 8.6 Scale: cluster, never collapse
Readability and performance at 10k+ nodes come from the same move: **always aggregate by meaning and only paint the current tier** — the inverse of collapsing above ~5,000. The default frame paints ~10–60 district boxes + ~30 bundled arrows + ~7 pins — trivial to draw and legible because a human reads named districts, not 10k dots. Viewport culling + LOD make off-screen and sub-pixel detail free; heavy repaints split across `requestAnimationFrame` with a low-detail proxy so the tab never freezes. **File weight, not render speed, is the real wall** (no `file://` transport gzip): the data island uses a **compact columnar encoding** (interned string table, integer coords, flat integer edge arrays, sorted keys) — ~1.3 MB at 10k nodes versus ~4.6 MB naive. v1 targets `map.html` < 1.5 MB at 10k. Above ~50k nodes the deepest symbol tier renders on-demand per visible district (deferred).

### 8.7 Accessibility
Canvas has no accessibility tree, so the binary also emits a **hidden, ordered DOM mirror** (districts → files → symbols) so screen readers and browser find-in-page work. Ships in **v1**.

### 8.8 Determinism
Same graph → **byte-identical `map.html` modulo the single `generated_at` timestamp**. Guaranteed by: layout baked in Go, integer-quantized coordinates, JSON keys and arrays sorted, and CSP hashes computed as pure functions of the (deterministic) inlined content (§14).

### 8.9 Workspace extension (§16.7)
The renderer is authored as a reusable module with an alias-ready coordinate format. The workspace **lanes** view (a tinted, labeled column per project, each lane this same community-box map, with cross-project links as bundled counted gutter connectors and a workspace-overview zoom tier) drops in as a **pure additive layer** — no second renderer and no data-format change.

## 9. Claude Code integration (`graffiti init`)

- Installs a **short** skill at `.claude/skills/graffiti/SKILL.md` (project) or `~/.claude/skills/...` (user): "Run `graffiti build .`; read `MAP.md`; present god nodes + surprising connections + offer to trace the single most interesting question. For codebase questions, run `graffiti query`."
- Appends a tiny **always-on block** to `CLAUDE.md`: "If `.graffiti/map.json` exists, answer codebase questions via `graffiti query` instead of grep/read; after editing code, run `graffiti update`."
- Optional **PreToolUse hook** that nudges grep → `graffiti query` when a map exists.
- The skill is short and declarative — the heavy lifting is the deterministic binary, not a procedure the model must execute step-by-step.

## 10. Tech stack & distribution

- **Language:** Go 1.26.
- **Parsing (ratified 2026-06-14):** the **pure-Go tree-sitter runtime `github.com/odvcencio/gotreesitter`** (a CGO-free Go reimplementation with embedded grammars), behind graffiti's own swappable `parse.Parser` interface. A feasibility spike found no maintained way to load tree-sitter grammar `.wasm` for the 6 target languages today, so the original "tree-sitter→WASM via wazero" intent is **superseded** as the primary path. gotreesitter satisfies every hard constraint (pure-Go, no CGO, single static cross-compilable binary, offline, deterministic). Because the backend sits behind `parse.Parser`, a **`wazero` + tree-sitter-WASM backend remains the documented fallback** for any single language whose pure-Go grammar fidelity proves inadequate. The library is young/AI-generated, so: pin & vendor the version, and validate parse-tree fidelity per language against golden fixtures before trusting it (see §14).
- **Graph:** a small custom directed adjacency model in `internal/graph` (no heavy dependency, keeps the binary lean and the data model fully under our control); Louvain implemented in-package.
- **Viewer assets:** embedded via `embed.FS`.
- **MCP:** Go MCP SDK over stdio.
- **Distribution:** pure-Go static binaries cross-compiled in CI (GitHub Actions) for darwin/linux/windows × amd64/arm64; published as GitHub Releases. Install via `curl -fsSL .../install.sh | sh` (downloads the right binary, puts it on PATH); Homebrew tap and scoop/winget later. **Single binary, zero runtime dependencies.**

## 11. CLI surface (minimal)

```
graffiti .                 # build the map for the current repo (alias: graffiti build .)
graffiti query "<q>"       # LLM-free scoped subgraph retrieval
graffiti update            # incremental AST-only rebuild
graffiti init              # install Claude Code integration
graffiti serve             # MCP server (stdio)
```

One success signal after build:
```
✓ Done. 0 API calls, $0.  N files → M nodes, K edges, C communities.
  The 3 most interesting questions your map can answer:
    1) …  2) …  3) …
```

## 12. Project structure

```
graffiti/
├── cmd/graffiti/main.go
├── internal/
│   ├── scan/        parse/        graph/
│   ├── cluster/     analyze/      render/    (render/viewer embedded assets)
│   ├── query/       mcp/          integrate/
├── schema/map.schema.json
├── docs/
├── go.mod
└── README.md
```

## 13. Success criteria (MVP acceptance)

1. `graffiti .` on a mixed Go/Python/JS repo finishes in seconds, makes **0 API calls**, writes 3 artifacts.
2. `graffiti query "where is auth handled"` returns a relevant, on-budget subgraph.
3. After `graffiti init`, Claude Code answers a codebase question by calling `graffiti query` rather than grepping.
4. `map.html` opens offline and shows a readable architecture map on a ~10k-node repo without degrading to a blob.
5. Install is a single command with no PATH/Python/runtime errors on a clean machine.

## 14. Testing

- Per-package Go unit tests; per-language parse fixtures.
- Golden tests for `map.json` and `MAP.md` on fixture repos.
- Query relevance tests (seed/expansion/budget).
- Determinism test: same input → byte-identical `map.json` (modulo `generated_at`).

## 15. Risks & mitigations

- **WASM grammar quality/coverage** for 6 languages → validate each grammar against fixtures up front; start with the 2 strongest (Go, Python) and expand.
- **wazero parse performance** on large repos → parallelize per-file parsing; benchmark early; cache aggressively.
- **Large-graph layout quality** of the architecture view → community-first rendering with drill-in; treat the readable map as the hero, not raw force layout.
- **Scope creep** → the Non-goals list is binding for v1.

## 16. Workspace / multi-project linking

> **Many repos, one system — without merging anything.** A *workspace* lays your separate project graphs side by side and draws only the wires between them. Each repo keeps its own map; nothing is rebuilt into a blob.

### 16.1 Model: federation, not merge

A workspace is a thin **computed overlay** over N independent per-project graphs. There are three artifacts with strict ownership:

1. **`.graffiti/map.json` per project — unchanged and authoritative.** Local, unprefixed IDs exactly as in §6. A project that happens to be in a workspace still builds, queries, and renders standalone with zero workspace awareness. `schema/map.schema.json` does **not** change.
2. **`.graffiti/workspace.json` — a human-legible registry** (committable). It holds **no graph data**: just the member list (alias + relative path + last-seen content hash) and optional declarative link rules. This is the only new file a user ever edits; it contains pointers and intent, never nodes or edges, so it cannot drift the way a merged store would.
3. **`.graffiti-workspace/overlay.json` — a derived cache** (gitignorable, never authored, never a source of truth). It holds the computed cross-project **links** as alias-qualified edges plus the source hashes they were built from, and is transparently recomputed when any member changes.

The alias qualifier (`alias::nodeId`) exists **only** in the derived overlay; it is never written back into any project's nodes. The overlay is produced by a pass that runs **after** each project's own deterministic build, so cross-project links can never corrupt or shrink a single project's `map.json` (the merge-not-replace and anti-shrink guards of §6 are preserved). Because every link is recomputed from each side's *current* facts, rebuilding one project independently can never produce a stale overlay — the next workspace query/render self-heals.

The contract surface used for matching (HTTP routes a project **serves**, HTTP call-sites it **consumes**, exported/imported symbols) is recorded during each project's normal deterministic parse as ordinary nodes/edges plus an optional per-node `meta` field (e.g. `{ "http_method": "GET", "http_path": "/cart" }`). The exact carrier (an optional node `meta` object vs. dedicated `exports`/`imports_external` arrays) is finalized in the implementation plan; either way it lives in each project's own `map.json` as a local fact and the `map.schema.json` core node/edge shape stays backward-compatible.

```jsonc
// .graffiti/workspace.json  (registry — no graph data)
{ "version": "1", "name": "shop", "generated_at": "RFC3339",
  "members": [
    { "alias": "web", "path": "../frontend", "map_hash": "sha256…" },
    { "alias": "api", "path": "../backend",  "map_hash": "sha256…" } ],
  "link_rules": { "match": ["http", "symbol"], "use_contracts": true, "first_party": ["@shop/types"] } }
```

```jsonc
// .graffiti-workspace/overlay.json  (derived — recomputable, gitignorable)
{ "version": "1", "generated_at": "RFC3339",
  "source_hashes": { "web": "sha256…", "api": "sha256…" },
  "links": [
    { "from": "web::apiclient.fetchcart", "to": "api::routes.get_cart",
      "relation": "calls", "confidence": "INFERRED", "via": "route-match",
      "match": { "method": "GET", "path": "/cart", "operationId": null } } ],
  "ambiguous": [ /* same shape; surfaced for review, never traversed as confident */ ] }
```

A **CrossEdge** is just an `Edge` (§6) whose endpoints carry an alias prefix, so the query serializer and viewer need no new edge type. The `relation` and `confidence` vocabularies are reused verbatim; **no new enum values in v1**. The `via` field records discovery provenance for auditability. A `schema/workspace.schema.json` is published for both derived/registry files.

### 16.2 Link discovery (deterministic, zero-key, honesty-first)

Links are derived by pure static analysis in **strict precedence order**, reusing the Pass-2 `label → node` index (§5) extended across project boundaries. The governing policy is **under-link**: a wrong confident cross-edge silently misleads the assistant, whereas a missing one falls back to grep.

1. **Explicit — `.graffiti/links.yml`** (human/assistant-asserted) → **EXTRACTED**, `via: explicit`. The only 100%-precision signal because it does zero inference. **Ships first.** `graffiti links check` validates that both endpoints resolve to real nodes so stale links never point at ghosts.
2. **Symbol export × unresolved import** — a member's unresolved external reference matched against another member's exported definition. Unique match → **EXTRACTED** (`via: symbol`); multiple → **AMBIGUOUS**. A first-party allowlist (link rules) prevents incidental third-party dependencies from linking.
3. **OpenAPI/Swagger server route table** — parsed (YAML/JSON reader, no new grammar) into canonical `{method, fully-assembled path, operationId}`. Authoritative for **server** endpoints; sidesteps router-prefix assembly on the server side. **EXTRACTED** only when joined to a *detected generated client*; otherwise the client join is path-based → **INFERRED**.
4. **Literal-path route↔call-site matcher** → **INFERRED**, emitted only when *all* hold: (a) client method is statically known (fetch/axios verb, `requests`/`httpx`, Go `net/http`); (b) the client path resolves to a literal — or a literal joined to a **resolved base-URL/env constant that attributes the call to a specific member**; (c) the path is not in a frozen generic-path denylist (`/`, `/health`, `/login`, `/users`, `/api`, `/v1`, …); and the server route template has **router-prefix/group assembly** applied (FastAPI `include_router(prefix=)`, Flask `Blueprint url_prefix`, Gin `r.Group`). v1 framework scope is deliberately small (client: JS/TS fetch+axios, Python requests/httpx, Go net/http; server: Express/Fastify, FastAPI/Flask, Gin). **Decision: this matcher is TIME-BOXED.** It ships in v1 only if both base-URL/env resolution and router-prefix/group assembly land by the v1 cutoff; otherwise it **auto-defers** to a fast-follow and the workspace relies on signals 1–3. The weak version (without those two prerequisites) is a net-negative false-edge generator and must not ship.
5. **Known-SDK base-URL-override recognizer** (v1, for the canonical backend↔LLM-router case) → **INFERRED**, `via: sdk`. Recognizes a small set of well-known service clients (`openai`, `anthropic`, `langchain`) constructed with an overridden `base_url`/endpoint; when that base URL resolves to a federated member, attribute the call to it. No path/verb is required (these calls usually have none). Multi-match or unresolved base URL → **AMBIGUOUS**, never a confident edge.

**Confidence** reuses the §5 ladder: **EXTRACTED** (explicit; unique symbol match; OpenAPI + generated client); **INFERRED** (OpenAPI + hand-written client by path; gated literal-path match); **AMBIGUOUS** (multi-match, partial match, unresolved base URL, or generic path) — surfaced in a "review these" report and `MAP.md`, never drawn as a confident edge.

**False-positive guards (binding):** require both method and path for any auto HTTP edge; multi-match → AMBIGUOUS (never silently pick one); require resolved base-URL → member attribution before any INFERRED HTTP edge (this is what stops a call to an un-federated third party such as Stripe/OpenAI from matching a federated route); record the normalized join key in `match` for auditability.

**Backend↔LLM-router (the canonical hard case):** these are usually SDK calls with an overridden `base_url` and no extractable path/verb. v1 ships the **known-SDK recognizer (signal 5)** to attribute them to a federated member when the base URL resolves; anything it can't resolve degrades to **AMBIGUOUS** or is covered by an explicit `links.yml` entry. Broadening the recognized SDK set beyond `openai`/`anthropic`/`langchain` is a fast-follow.

**Dropped from v1 auto-detection** (kept only as resolution inputs or deferred): bare env-var/base-URL matching as a standalone edge producer; queue/topic/table-name matching; gRPC `.proto` and GraphQL SDL contracts.

### 16.3 Query, MCP, and update

`graffiti query` stays single-project and identical to §7 by default. `--workspace` opts into the overlay: each member loads into its own alias namespace, IDF seeding runs across all members, and expansion uses a **two-tier budget** — free intra-project BFS/DFS plus a separate small **cross-hop budget** (default 2) for traversing links, so the caller project's context stays dominant. The serialized subgraph prefixes every node with its alias (`api:routes.get_cart`) and annotates each cross-link with relation + confidence + `via`. AMBIGUOUS links are not traversed as confident edges but may appear in a "suspected links" footer. As in §7, graffiti makes no inference call; the host model reasons over the returned text. The same retrieval is exposed via `graffiti serve --workspace` (MCP tools accept and return `alias::id`).

`graffiti update --workspace` rebuilds only changed members (per-project SHA256 cache, §6) and then recomputes only the overlay (`--links-only` skips member rebuild). Each link carries the member `map_hash` it was computed against; if a member changed since, query output adds a one-line nudge `(overlay stale — run: graffiti update --workspace)` rather than returning silently wrong links.

### 16.4 CLI & success signal

```
graffiti link <pathA> <pathB> [<pathC> …]   # build any unbuilt member, write workspace.json, compute overlay
graffiti link --name shop <pathA> <pathB>   # name the workspace
graffiti workspace add <path> --as <alias>  # register a member (add / rm / list)
graffiti links check                        # validate links.yml resolves to real nodes
graffiti federate --explain                 # show the matched join key per link; AMBIGUOUS listed separately
graffiti query --workspace [name] "<q>"     # federated, alias-prefixed, two-tier budget
graffiti update --workspace [name]          # rebuild changed members + recompute overlay (--links-only)
graffiti serve  --workspace [name]          # MCP over the federation
```

One success signal mirrors §11:
```
✓ Linked 2 projects. 14 cross-project links (web→api 9, api→web 5;
  11 EXTRACTED, 3 INFERRED, 2 AMBIGUOUS for review). 0 API calls, $0.
```

Setup stays 1–2 commands with one success line. The single-project happy path is untouched: adopting a workspace adds capability without adding runtime, a database, a service, or anything to keep in sync by hand. `graffiti init` makes the always-on `CLAUDE.md` block workspace-aware so the assistant prefers `graffiti query --workspace` for system-spanning questions and plain `graffiti query` for single-repo ones.

**Workspace root location (decision):** `workspace.json` + `.graffiti-workspace/` are **auto-discovered** — `graffiti link`/`--workspace` walk up to the nearest common ancestor of the member paths and default the workspace root there, then search upward from the cwd on later runs. No manual placement or config is required; a dedicated orchestrator dir or a hub repo still works if the user prefers, but is never mandatory.

### 16.5 Determinism

The overlay is a pure function of (member `map.json` files) + (optional contract files) + (`links.yml`), computed by exact-after-normalization key matching — no network, no model, no API key. The path-template normalization grammar (param-syntax collapsing, trailing-slash/query handling, prefix/group assembly) is **frozen in this spec** so byte-identical output is guaranteed across languages. Members are sorted by alias and links by `(from, to, relation)`; the §14 determinism test extends directly: same N inputs → byte-identical `overlay.json` modulo `generated_at`. Any LLM assistance (e.g. a future `--suggest` for AMBIGUOUS candidates) is explicitly optional, off by default, human-confirmed, written back only as explicit `links.yml` rules, and never on the default query/render path.

### 16.6 Identity

Each project carries a small **immutable project id** in its `.graffiti/` (generated on first build), distinct from the human-facing `alias`. Committed `links.yml` rules and overlay edges resolve members by this id, so renaming a folder or alias never breaks committed links. The alias remains the display/query token; the id is the durable identity.

### 16.7 Visualization (fast-follow)

v1 ships the text/MCP federated path. A separate, self-contained `workspace.html` (`graffiti workspace render` → `.graffiti-workspace/`) follows: deterministic **lanes** (one tinted, labeled column per project, each lane the project's §8 community-box map), cross-project links drawn as **bundled, counted connectors in the gutters**, a new **workspace-overview** zoom tier above community→file→symbol, AMBIGUOUS links dashed in a "suspected links" panel. Per-project `map.html` files stay untouched; all assets embedded via `embed.FS`, offline, CSP-safe, deterministic.
```
