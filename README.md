# graffiti

One command turns your repository into a directed knowledge graph your AI coding
assistant reads instead of blindly grepping.

> **Status:** Plans 1–5. `graffiti .` builds a deterministic, schema-valid
> `.graffiti/map.json` (+ `MAP.md` + `map.html`) for a Go repository, with
> clustering/analysis, an LLM-free `query`, an MCP `serve`r, and Claude Code
> `init` integration. More languages and workspace federation are later plans.

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~8MB, Go grammar only)
make test       # runs the full test suite with required build tags
make xcompile   # cross-compiles static binaries for all v1 targets into dist/
```

The build tags `grammar_subset grammar_subset_go grammar_subset_gomod` ship only
the Go tree-sitter grammar (pure-Go runtime via `github.com/odvcencio/gotreesitter`,
no CGO, no WASM). They are required for the ~8MB size target; without them the code
still compiles but links the full grammar set (~31MB). Always pass them (the
Makefile does this for you).

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

## Guarantees (Plan 1)

- 0 API calls, $0, fully offline.
- Deterministic: same repo → byte-identical `map.json` modulo the single
  `generated_at` timestamp and the `root` basename.
- Single static binary, no runtime dependencies, no C toolchain.
