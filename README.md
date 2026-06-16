# graffiti

One command turns your repository into a directed knowledge graph your AI coding
assistant reads instead of blindly grepping.

> **Status:** Plans 1–6. `graffiti .` builds a deterministic, schema-valid
> `.graffiti/map.json` (+ `MAP.md` + `map.html`) for **Go, Python, JavaScript,
> TypeScript, Rust, Java, and PHP** repositories, with clustering/analysis, an
> LLM-free `query`, an MCP `serve`r, and Claude Code `init` integration. Workspace
> federation is a later plan.

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

## Guarantees (Plan 1)

- 0 API calls, $0, fully offline.
- Deterministic: same repo → byte-identical `map.json` modulo the single
  `generated_at` timestamp and the `root` basename.
- Single static binary, no runtime dependencies, no C toolchain.
