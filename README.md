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
