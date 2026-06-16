---
name: graffiti
description: Use when exploring or answering questions about THIS codebase's structure — where something is defined, how components connect, what the architecture looks like. graffiti turns the repo into a queryable code map so you query the graph instead of grepping blind.
---

# graffiti — read the map, don't grep blind

`graffiti` is a CLI that turns this repository into a queryable code map (no API key, $0, offline).

## First time in a repo
1. Run `graffiti build .` — it writes `.graffiti/map.json`, `.graffiti/MAP.md`, and `.graffiti/map.html`.
2. Read `.graffiti/MAP.md`. Tell the user the **god nodes** (most-connected code) and the **surprising connections**, then offer to trace the single most interesting question the map suggests.

## Answering questions about the code
- Prefer `graffiti query "<question>"` over grep/read — it returns a scoped, ranked subgraph (relevant nodes + edges), not raw text matches.
- To locate a symbol, run `graffiti query "<symbol name>"`.

## After editing code
- Run `graffiti update` so later queries reflect the current code.

Keep this lightweight: the deterministic binary does the heavy lifting — build, read MAP.md, query.
