# graffiti

One command turns your repository into a directed knowledge graph your AI coding
assistant reads instead of blindly grepping.

> **Status:** Plan 1 (walking skeleton). `graffiti .` builds a deterministic,
> schema-valid `.graffiti/map.json` for a Go repository. Clustering, MAP.md,
> map.html, query, MCP, init, more languages, and workspace federation are later
> plans.

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
graffiti .              # build the map for the current repo
graffiti build <path>   # build the map for <path>
graffiti <path>         # shorthand for `build <path>` when <path> is a directory
```

Output: `<path>/.graffiti/map.json` (see `schema/map.schema.json` for the contract)
and a per-file content-hash cache under `<path>/.graffiti/cache/`.

## Guarantees (Plan 1)

- 0 API calls, $0, fully offline.
- Deterministic: same repo → byte-identical `map.json` modulo the single
  `generated_at` timestamp and the `root` basename.
- Single static binary, no runtime dependencies, no C toolchain.
