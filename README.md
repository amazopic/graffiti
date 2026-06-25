# 🕸️ graffiti — turn any repo into a queryable code graph for AI

> One command turns your repository into a **directed knowledge graph** your AI
> coding assistant reads instead of blindly grepping. A single static Go binary —
> **zero API keys, $0, fully offline, byte-deterministic.** Parses **Go, Python,
> JavaScript, TypeScript, Rust, Java, and PHP**. Ships an LLM-free `query`, an
> **MCP** server, **Claude Code** integration, an interactive offline graph
> viewer, and multi-repo workspace federation.

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**Languages:** English · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

🌐 **Website:** https://amazopic.github.io/graffiti/

```text
$ graffiti .
✓ Done. 0 API calls, $0.  214 files → 1,883 nodes, 4,102 edges, 12 communities.
  The 3 most interesting questions your map can answer:
    1) Which module is the load-bearing wall?
    2) What does the auth flow touch?
    3) Where are the cross-package call hotspots?
```

---

## Why this exists

An AI coding assistant is only as good as what it can *see*. Drop it into a large
repo and it does what you'd do with no map: it greps, opens a few files, guesses.
It never sees the **shape** of the code — which function calls which, where a type
is defined, which module is the load-bearing wall.

**graffiti is the map that should have been there.** One command parses the repo
with [tree-sitter](https://tree-sitter.github.io/tree-sitter/), resolves the edges,
clusters the modules, and writes a graph — as JSON for the machine, as Markdown for
you, and as a single offline HTML you can actually look at. No keys. No cloud. No cost.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

Pin a version or directory:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

The installer picks the right static binary for your OS/arch, verifies its SHA256
against the release manifest, and installs it. Verify with `graffiti version`.
Or build from source (below).

## ⚡ Install with Claude Code (vibe-code)

<!-- vibe-install -->
No terminal needed — let **Claude Code** do the whole thing. Paste this one prompt
into a Claude Code session and answer `y` at each step. It fetches the right binary,
builds the map for your repo, wires up the integration, and opens the graph:

```text
Install graffiti by amazopic for me. Download the right static binary for my OS/arch from the latest release at github.com/amazopic/graffiti (or build it from source with `make build` if Go is available), put it on my PATH as `graffiti`, and verify with `graffiti version`. Then run `graffiti .` at my repo root to build the map, run `graffiti init --hook` to wire graffiti into Claude Code, and finally open `.graffiti/map.html` so I can see the graph. Ask before each step.
```

<!-- quickstart -->
## Quickstart (60 seconds)

```bash
# 1 — install (or build from source with `make build`)
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh

# 2 — map your repo (writes .graffiti/map.json, MAP.md, map.html)
cd your-repo
graffiti .

# 3 — look at the graph
open .graffiti/map.html        # macOS — use `xdg-open` on Linux, `start` on Windows

# 4 — ask it questions: no LLM, no API key
graffiti query "where is the user authenticated"
```

Then wire it into your AI assistant once:

```bash
graffiti init --hook    # Claude Code: skill + CLAUDE.md + grep→query nudge
graffiti serve          # or expose the map to any MCP client over stdio
```

**More example questions** — `query` returns a scoped subgraph within a soft
~2,000-token budget, so context stays small and cheap (quote the question):

```bash
graffiti query "login handler"
graffiti query "what does the checkout flow touch"
graffiti query "where is the cart fetched" ../shop   # target another path
```
<!-- /quickstart -->

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

The `grammar_subset` build tags ship only the grammars graffiti supports (Go,
Python, JS, TS, Rust, Java, PHP, plus go.mod) via the pure-Go runtime
`github.com/odvcencio/gotreesitter` (no CGO, no WASM). They keep the binary at
~10 MB; without them the code still compiles but links the full grammar set
(~31 MB). Always pass them — the Makefile does this for you.

## Supported languages

| Language | Extracted |
|----------|-----------|
| Go | files, functions, methods (by receiver), types, imports, resolved calls |
| Python, JavaScript, TypeScript, Rust, Java, PHP | files, functions, classes/structs/interfaces/enums/traits, methods (`Class.method`), imports, intra-repo calls |
| Markdown | doc nodes |

Non-Go extraction is intentionally honest: it captures the common, high-value
structure and **under-extracts** exotic constructs (decorators, generics, nested
definitions, dynamic dispatch) rather than emitting guesses.

## Usage

```bash
graffiti .                  # build the map for the current repo
graffiti build <path>       # build the map for <path>
graffiti <path>             # shorthand for `build <path>` when <path> is a directory
graffiti update [path]      # rebuild the map (full rebuild for now)
graffiti query "<q>" [path] # LLM-free scoped subgraph retrieval (soft token budget)
graffiti serve [path]       # MCP server over stdio (JSON-RPC 2.0)
graffiti init [--user] [--hook]  # install Claude Code integration
graffiti version            # print the version
```

Run `graffiti` with no arguments for the full command list.

## One command, three artifacts

`graffiti .` writes everything into `<repo>/.graffiti/`:

- **`map.json`** — the graph itself: nodes, edges, communities, schema-checked
  against `schema/map.schema.json`. This is what your AI reads and what `query`
  and the MCP server traverse.
- **`MAP.md`** — a human-readable digest: top modules, the most-connected nodes,
  and the three most interesting questions your map can answer.
- **`map.html`** — a single self-contained, offline, interactive **force-directed
  graph**. No CDN, no server, no network — just open the file.

`map.html` has a **2D/3D toggle** (hover lifts a node and its neighbours), **node
search**, **click-to-copy `file:line`**, **sector zones**, **client / tests /
external** category toggles, and a resizable **project → directory → file** tree
with show/hide checkboxes. It's CSP-safe and works entirely offline.

A per-file content-hash cache lives under `<repo>/.graffiti/cache/`, so re-runs
only re-parse what changed.

## Claude Code integration

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` writes:

- `.claude/skills/graffiti/SKILL.md` — a short skill so Claude Code knows to build/read/query the map.
- a `CLAUDE.md` block (between `<!-- graffiti:start -->` / `<!-- graffiti:end -->`) telling the
  assistant to prefer `graffiti query` over grep when a map exists.
- with `--hook`, a `.claude/settings.json` PreToolUse entry running `graffiti hook`, which adds a
  one-line nudge before `Grep`/`Glob` when `.graffiti/map.json` is present. The hook never blocks a tool.

It is idempotent — re-run any time; existing `CLAUDE.md` / `settings.json` content is preserved.

## Query without an LLM

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` returns a relevant slice of the graph within a soft ~2000-token node
budget — no model, no embeddings. Quote the question.

## MCP server

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

Point any MCP-capable client at it and your assistant traverses the graph through
tools instead of grepping.

## Workspaces (multi-repo federation)

Lay separate repos side by side and query across them — **without merging**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` writes a committable registry (`.graffiti-workspace/workspace.json`)
and a derived, gitignorable cache (`.graffiti-workspace/overlay.json`). Each repo's
own `.graffiti/map.json` is unchanged and still works standalone — the workspace is a
thin computed overlay, never a merged blob.

**Cross-project links:** assert them explicitly in `.graffiti-workspace/links`,
one per line — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(`#` comments allowed; endpoints are `alias::nodeid`). `graffiti links check` validates
both endpoints resolve; `graffiti federate --explain` lists every link. Federated query
prefixes each node with its member alias and traverses cross-links. `graffiti workspace
render` writes a `workspace.html` — the same force-graph viewer with the **projects as
the top level** of the tree and cross-project links drawn.

Add `.graffiti-workspace/overlay.json` to `.gitignore` (it is derived and recomputable).

## 🛰️ System orchestration — many services, one graph

<!-- system-orchestration -->
A microservice system is many independent repos that form one product. graffiti
maps each, then **discovers the edges between them** — HTTP, gRPC, queues — from
each service's *contract surface* (what it `provides` and `consumes`). No
hand-wiring: each service publishes its own map; the orchestrator federates the
published artifacts and matches consumers to providers.

```bash
# in each service's CI (or locally) — publish its map into a shared store:
graffiti publish --to ../system-store --as carts

# then, in CI or on demand, over the whole system:
graffiti system build       # federate + auto-discover cross-service links
graffiti system render      # → .graffiti-system/system.html (services as lanes)
graffiti system impact carts::"GET /carts/{}"   # who breaks if this changes?
graffiti system audit       # dangling consumers · orphan providers · ambiguous (CI gate)
graffiti system query "where is the cart fetched and served"
```

Each map carries a **contract surface** extracted from `openapi.json`, `.proto`,
framework routes, queue calls, or an explicit `graffiti.contract.json`. Cross-service
links are scored by confidence; **ambiguous** and **dangling** (dead-endpoint)
consumers are reported, never silently dropped. The system store is just a directory
or git repo — $0, offline, recomputable. See the
[design doc](docs/superpowers/specs/2026-06-24-graffiti-system-orchestration-design.md).

<!-- system-walkthrough -->
### A folder of services, step by step

Say your services live in one parent folder, each in its own directory:

```text
myproject/                ← parent folder = the shared "system store"
├── orders/               ← a service (Go)
├── web/                  ← a service (React/TS)
└── payments/             ← a service (Python)
```

**1. Build and publish each service** into a store at the parent folder (`--to .`).
`publish` reuses an existing map, so build first to pick up code changes:

```bash
cd myproject
for d in */; do
  d=${d%/}
  graffiti build "$d" && graffiti publish "$d" --to .
done
```

The service name defaults to its folder name; override with `--as <name>`.

> ⚠️ **On re-publish:** `publish` does **not** rebuild an existing map. After
> changing code, always run `graffiti build <service>` first (the loop above
> does) and then `publish` — otherwise you publish a stale map.

**2. Build the system graph** — federate the maps and auto-discover the links:

```bash
graffiti system build
# ✓ System "myproject": 3 services → 7 cross-service links (0 ambiguous, 0 dangling, 2 orphan). 0 API calls, $0.
```

**3. Use it:**

```bash
graffiti system render          # → .graffiti-system/system.html (services as the top tree level)
graffiti system impact orders   # who breaks if orders changes (direct + transitive)
graffiti system audit           # dangling consumers · orphan providers · ambiguous (non-zero exit → CI gate)
graffiti system status          # which services drifted since the last build
graffiti system query "where is the order created"   # LLM-free retrieval across the whole system
graffiti system list            # registered services
```

**What lands in the parent folder:**

```text
myproject/.graffiti-system/
├── system.json                 # the registry of services (commit this)
├── overlay.json                # discovered links (derived — safe to .gitignore)
├── system.html                 # the visual system map
└── services/<name>/map.json    # each service's published map
```

**Improve link accuracy.** Auto-detection covers Go (net/http, gin/chi/echo),
Flask, FastAPI, Django/DRF, Spring, NestJS, ASP.NET, Ktor, frontend clients
(React/Vue/Angular/Svelte), gRPC and Kafka/NATS. Where that is not enough, drop
one of these into a service root (highest confidence first):

| File | Gives |
|------|-------|
| `graffiti.contract.json` | explicit `provides` / `consumes` — any stack, highest confidence |
| `openapi.json` / `swagger.json` | HTTP routes as `provides` |
| `*.proto` | gRPC methods as `provides` |

Minimal `graffiti.contract.json`:

```json
{
  "provides": [{ "kind": "http", "name": "GET /orders/{id}" }],
  "consumes": [{ "kind": "rpc",  "name": "Payments.Charge" }]
}
```

**Gate CI on dead endpoints** — `audit` exits non-zero when a consumer points at
an endpoint nothing provides:

```bash
graffiti system build && graffiti system audit
```
<!-- /system-walkthrough -->

## How it works

tree-sitter parsing (pure-Go, no CGO) → edge resolution → clustering into
communities → lightweight analysis → deterministic serialization. No model, no
embeddings, no network — just static analysis. That's why it's free, private, and
reproducible.

## Guarantees

- **0 API calls, $0, fully offline.** Nothing about your code leaves your machine.
- **Deterministic:** same repo → byte-identical `map.json` modulo the single
  `generated_at` timestamp and the `root` basename. Commit it; diff it.
- **Single static binary**, no runtime dependencies, no C toolchain.

## License

Source-Available — read and run graffiti freely on your own repositories, but any
reuse, redistribution, fork, or inclusion in another project requires prior written
permission from the author. See [LICENSE](LICENSE).

## Author

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
