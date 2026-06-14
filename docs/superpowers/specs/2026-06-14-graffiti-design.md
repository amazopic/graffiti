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

- **Self-contained**: all JS/CSS/data inlined via `embed.FS`; opens from `file://` offline, no CDN, CSP-safe.
- **Default view = readable architecture map**: communities rendered as labeled boxes; inter-community edges drawn as call-flow; intra-community detail hidden until drill-in.
- **Semantic zoom**: community → file → symbol. Large graphs render at community level first and drill in on click — scales past 5,000 nodes by clustering, never by collapsing into an unreadable blob.
- HTML/JS emitters escape all interpolated content (XSS-safe self-contained file).

## 9. Claude Code integration (`graffiti init`)

- Installs a **short** skill at `.claude/skills/graffiti/SKILL.md` (project) or `~/.claude/skills/...` (user): "Run `graffiti build .`; read `MAP.md`; present god nodes + surprising connections + offer to trace the single most interesting question. For codebase questions, run `graffiti query`."
- Appends a tiny **always-on block** to `CLAUDE.md`: "If `.graffiti/map.json` exists, answer codebase questions via `graffiti query` instead of grep/read; after editing code, run `graffiti update`."
- Optional **PreToolUse hook** that nudges grep → `graffiti query` when a map exists.
- The skill is short and declarative — the heavy lifting is the deterministic binary, not a procedure the model must execute step-by-step.

## 10. Tech stack & distribution

- **Language:** Go 1.26.
- **Parsing:** tree-sitter grammars compiled to **WASM**, executed via **wazero** (pure-Go WASM runtime). No CGO → trivial static cross-compilation. Grammars embedded via `embed.FS`.
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
4. **Literal-path route↔call-site matcher** → **INFERRED**, emitted only when *all* hold: (a) client method is statically known (fetch/axios verb, `requests`/`httpx`, Go `net/http`); (b) the client path resolves to a literal — or a literal joined to a **resolved base-URL/env constant that attributes the call to a specific member**; (c) the path is not in a frozen generic-path denylist (`/`, `/health`, `/login`, `/users`, `/api`, `/v1`, …); and the server route template has **router-prefix/group assembly** applied (FastAPI `include_router(prefix=)`, Flask `Blueprint url_prefix`, Gin `r.Group`). v1 framework scope is deliberately small (client: JS/TS fetch+axios, Python requests/httpx, Go net/http; server: Express/Fastify, FastAPI/Flask, Gin). **If base-URL resolution and prefix assembly cannot both land in v1, this matcher is dropped** and the workspace relies on signals 1–3 — the weak version is a net-negative false-edge generator and will not ship.

**Confidence** reuses the §5 ladder: **EXTRACTED** (explicit; unique symbol match; OpenAPI + generated client); **INFERRED** (OpenAPI + hand-written client by path; gated literal-path match); **AMBIGUOUS** (multi-match, partial match, unresolved base URL, or generic path) — surfaced in a "review these" report and `MAP.md`, never drawn as a confident edge.

**False-positive guards (binding):** require both method and path for any auto HTTP edge; multi-match → AMBIGUOUS (never silently pick one); require resolved base-URL → member attribution before any INFERRED HTTP edge (this is what stops a call to an un-federated third party such as Stripe/OpenAI from matching a federated route); record the normalized join key in `match` for auditability.

**Known limitation (documented, not hidden):** the backend↔LLM-router case is typically an SDK call with an overridden `base_url` and no extractable path/verb, so v1 will frequently detect **nothing** there; cover it with an explicit `links.yml` entry. A small known-SDK recognizer is a candidate enhancement.

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

### 16.5 Determinism

The overlay is a pure function of (member `map.json` files) + (optional contract files) + (`links.yml`), computed by exact-after-normalization key matching — no network, no model, no API key. The path-template normalization grammar (param-syntax collapsing, trailing-slash/query handling, prefix/group assembly) is **frozen in this spec** so byte-identical output is guaranteed across languages. Members are sorted by alias and links by `(from, to, relation)`; the §14 determinism test extends directly: same N inputs → byte-identical `overlay.json` modulo `generated_at`. Any LLM assistance (e.g. a future `--suggest` for AMBIGUOUS candidates) is explicitly optional, off by default, human-confirmed, written back only as explicit `links.yml` rules, and never on the default query/render path.

### 16.6 Identity

Each project carries a small **immutable project id** in its `.graffiti/` (generated on first build), distinct from the human-facing `alias`. Committed `links.yml` rules and overlay edges resolve members by this id, so renaming a folder or alias never breaks committed links. The alias remains the display/query token; the id is the durable identity.

### 16.7 Visualization (fast-follow)

v1 ships the text/MCP federated path. A separate, self-contained `workspace.html` (`graffiti workspace render` → `.graffiti-workspace/`) follows: deterministic **lanes** (one tinted, labeled column per project, each lane the project's §8 community-box map), cross-project links drawn as **bundled, counted connectors in the gutters**, a new **workspace-overview** zoom tier above community→file→symbol, AMBIGUOUS links dashed in a "suspected links" panel. Per-project `map.html` files stay untouched; all assets embedded via `embed.FS`, offline, CSP-safe, deterministic.
```
