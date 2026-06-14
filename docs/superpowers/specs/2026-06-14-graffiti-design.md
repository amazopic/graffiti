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

**Non-goals (explicitly deferred)**
- Multi-host install matrix (only Claude Code in v1; auto-detect later).
- Any LLM backend in the default path; semantic extraction of docs/PDFs/images; multimodal.
- External databases (Neo4j/FalkorDB/Postgres), PR triage, cross-repo "global" graph, HTTP MCP server with auth.
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
```
