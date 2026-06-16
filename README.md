# graffiti

One command turns your repository into a directed knowledge graph your AI coding
assistant reads instead of blindly grepping.

> **Status:** Plans 1–8. `graffiti .` builds a deterministic, schema-valid
> `.graffiti/map.json` (+ `MAP.md` + `map.html`) for **Go, Python, JavaScript,
> TypeScript, Rust, Java, and PHP** repositories, with clustering/analysis, an
> LLM-free `query`, an MCP `serve`r, Claude Code `init` integration, multi-repo
> `--workspace` federation, and one-command install.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/evgeniy-achin/graffiti/main/scripts/install.sh | sh
```

Pin a version or directory:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/evgeniy-achin/graffiti/main/scripts/install.sh)"
```

The installer picks the right static binary for your OS/arch, verifies its SHA256
against the release manifest, and installs it. Verify with `graffiti version`.
Or build from source (below).

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with required build tags
make xcompile   # cross-compiles static binaries for all v1 targets into dist/
```

The `grammar_subset` build tags ship only the grammars graffiti supports (Go,
Python, JS, TS, Rust, Java, PHP, plus go.mod) via the pure-Go runtime
`github.com/odvcencio/gotreesitter` (no CGO, no WASM). They keep the binary at
~10 MB; without them the code still compiles but links the full grammar set
(~31 MB). Always pass them (the Makefile does this for you).

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
graffiti .                 # build the map for the current repo
graffiti build <path>      # build the map for <path>
graffiti <path>            # shorthand for `build <path>` when <path> is a directory
graffiti update [path]     # rebuild the map (full rebuild for now)
graffiti query "<q>" [path] # LLM-free scoped subgraph retrieval (soft token budget)
graffiti serve [path]      # MCP server over stdio (JSON-RPC 2.0)
graffiti init [--user] [--hook]  # install Claude Code integration
```

Output: `<path>/.graffiti/map.json` (see `schema/map.schema.json` for the contract),
`MAP.md`, `map.html`, and a per-file content-hash cache under `<path>/.graffiti/cache/`.

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

It is idempotent — re-run any time; existing `CLAUDE.md`/`settings.json` content is preserved.

## Workspaces (multi-repo federation)

Lay separate repos side by side and query across them — **without merging**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
```

`graffiti link` writes a committable registry (`.graffiti-workspace/workspace.json`)
and a derived, gitignorable cache (`.graffiti-workspace/overlay.json`). Each repo's
own `.graffiti/map.json` is unchanged and still works standalone — the workspace is a
thin computed overlay, never a merged blob.

**Cross-project links (v1):** assert them explicitly in `.graffiti-workspace/links`,
one per line — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(`#` comments allowed; endpoints are `alias::nodeid`). `graffiti links check` validates
both endpoints resolve to real nodes; `graffiti federate --explain` lists every link.
Federated query prefixes each node with its member alias and traverses cross-links.
Automatic link discovery (shared symbols, HTTP routes, SDK base-URLs) and a
`workspace.html` lanes view are planned follow-ons.

Add `.graffiti-workspace/overlay.json` to `.gitignore` (it is derived and recomputable).

## Guarantees (Plan 1)

- 0 API calls, $0, fully offline.
- Deterministic: same repo → byte-identical `map.json` modulo the single
  `generated_at` timestamp and the `root` basename.
- Single static binary, no runtime dependencies, no C toolchain.
