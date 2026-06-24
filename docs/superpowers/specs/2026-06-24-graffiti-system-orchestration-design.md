# graffiti — System Orchestration (multi-service code graph) · Design

> **Status:** design → implementing. Builds on Plan 7 workspace federation.
> The killer feature for service architectures: many independent service repos,
> each publishing its own map, federated into ONE system graph with **automatic
> cross-service links** discovered from contracts.

## Problem

A microservice system is N independent repos that together form one product. Each
repo's `graffiti` map is useful alone, but the system-level questions — *what calls
this endpoint? if I change it, who breaks? what's the whole architecture? is any
consumer pointing at a dead endpoint?* — need the **combined** graph with edges
that cross repo boundaries. Plan 7 federates maps and supports **explicit**
cross-links; it cannot **discover** them. Discovery needs each map to carry a
**contract surface**, which today it does not.

## Decisions (locked with the user)

- **Topology:** each service publishes its own map artifact (CI-native push); an
  orchestrator federates the published artifacts. Storage-agnostic; **git-as-registry**
  for v1 ($0, offline, versioned, diffable).
- **Links:** **auto-discover from contracts** (HTTP / gRPC / queues / shared lib)
  with confidence scoring; **explicit links remain as overrides**.
- **Outcomes:** impact analysis, architecture onboarding map, contract/dependency
  audit, cross-service navigation — all four.

## Three layers (each independently shippable)

### A. Contract surface (single-repo, foundational)

Every `map.json` gains two arrays (deterministic, sorted, optional):

```jsonc
"provides": [ { "kind":"http","key":"GET /carts/{}","display":"GET /carts/{id}",
                "node":"...","file":"...","line":42,"confidence":"EXTRACTED","source":"openapi" } ],
"consumes": [ { "kind":"http","key":"GET /carts/{}","display":"GET http://carts/{id}",
                "node":"...","file":"...","line":88,"confidence":"INFERRED","source":"literal" } ]
```

- `kind`: `http` | `rpc` | `queue` | `lib`
- `key`: normalized **match key** (the join key between consumes↔provides).
  - http → `METHOD /path/with/{}` (path params collapsed to `{}`, host stripped)
  - rpc → `Service.Method`
  - queue → `topic`
  - lib → `package:Symbol`
- `confidence`: `EXTRACTED` (declared spec / explicit) · `INFERRED` (code heuristic) · `AMBIGUOUS`
- `source`: `openapi` | `proto` | `contract` | `route` | `literal`
- `node`: nearest graph node in the same file at/above the match line (handler / call site).

**Extractors (confidence order):**
1. **`graffiti.contract.json`** (explicit, repo-authored) — EXTRACTED. Always works, any stack.
2. **OpenAPI** (`openapi.json` / `swagger.json`, stdlib JSON) → http provides — EXTRACTED.
3. **`.proto`** (small line scanner) → rpc provides — EXTRACTED.
4. **Framework recognizers** (INFERRED), per detected file role:
   - **Backend providers** — Go net/http, gin/chi/echo, Flask `.route`, FastAPI/`@app.get`,
     **Django/DRF** (`path`/`re_path`/`url` in urls.py), **Spring** (`@*Mapping` +
     `@RequestMapping` with class-prefix tracking), **NestJS** (`@Controller` prefix +
     `@Get`/`@Post`/… + `@MessagePattern`/`@EventPattern`), **ASP.NET** (`[HttpGet]`/
     `[Route]`/`MapGet` with `[controller]` substitution), **Ktor** (routing DSL),
     **Kafka** (`@KafkaListener`) / **NATS** / generic `.subscribe` → http/queue provides.
   - **Frontend consumers** — React / Vue / Angular / Svelte / Nuxt files (by extension or
     `react`/`vue`/`@angular` import) where `fetch`/`axios`/`$fetch`/`useFetch`/`HttpClient`
     and `.get("/x")` calls are CONSUMES, not routes.
   - **gRPC clients** — Go `New<Svc>Client(conn).Method(...)` (chained) and
     `c := New<Svc>Client(...)` + `c.Method(...)`, Python `<Svc>Stub(...)` → rpc
     CONSUMES (`Svc.Method`).
   - **gRPC providers** are attributed by **server registration** (Go
     `RegisterXServer`, Python `add_XServicer_to_server`, C# `MapGrpcService<X>`,
     Java `XImplBase`, Node `.X.service` — call sites only, not generated definitions)
     + the rpc method set from the `.proto` or generated `XServer`/`XServicer` stub.
     A repo that merely vendors a shared multi-service proto/stub as a client (no
     registration) provides nothing — defeating false ambiguity.

> **Validated on a real polyglot system.** Run against Google's *Online Boutique*
> (`microservices-demo`, 12 services in Go/C#/Node/Python/Java over gRPC), `graffiti
> system build` reconstructed the actual service dependency graph — 14 cross-service
> links, **0 ambiguous** (checkout→{cart,currency,email,payment,productcatalog,
> shipping}, frontend→{ad,cart,recommendation,shipping}, recommendation→
> productcatalog) — fully offline, $0. Server-registration attribution was the key:
> the demo vendors the full `demo.proto` into several services, which naïvely makes
> every one a provider of every rpc; scoping provides to registered servers
> eliminated the ambiguity.
5. **Literal / producer heuristic** — `http(s)://…` literals and queue producers
   (`publish`/`emit`/`produce`, `KafkaTemplate.send`, NATS `Publish`) → http/queue consumes.

Role detection resolves the key ambiguity: `.get("/x")` is a *route* (provide) in a
backend file but a *client call* (consume) in a frontend file; a client call
(`axios.get`) on a backend gateway line is also read as a consume.

Honest under-extraction: recognize the common, mark confidence, let `graffiti.contract.json`
fill gaps. New package `internal/contract`; wired into `app.Build` after assembly.

### B. Registry · publish · federate · match

- **`.graffiti-system/system.json`** (committable): the list of services —
  `{name, fetch (git path / dir / url), language, owner, pinned commit}`. The
  multi-repo analog of Plan 7's `workspace.json`.
- **`graffiti publish [--to <dir>] [--as <name>]`**: writes this repo's artifact
  (`map.json` + contract + commit SHA metadata) into the system store (a dir / git
  checkout). Deterministic.
- **`graffiti system build`**: collect artifacts (cache by SHA) → federate
  (reuse `workspace.CombinedDocument`, alias = service name) → **match**.
- **The matcher** (`internal/system/match.go`): index every `provides` by
  `(kind,key)` across services; for each `consumes`, find providers in **other**
  services with the same key (most-specific http template wins). Emit a cross-service
  link `consumerNode → providerNode` (relation `calls`/`references`, kind carried).
  - `>1` provider → **ambiguous** (recorded, not drawn confidently).
  - `0` providers → **dangling** (consumer points at nothing → audit signal).
  - explicit `links` entries apply as EXTRACTED overrides.
- **System overlay** (`.graffiti-system/overlay.json`, derived/gitignorable):
  confident links + ambiguous + dangling + per-link confidence + provenance.

### C. Views · queries

- **`graffiti system render`** → `.graffiti-system/system.html`: the Plan 9
  force-graph viewer with **services as the top tree level** and cross-service edges
  drawn/colored (reuse `render.WriteWorkspaceHTML`).
- **`graffiti system impact <service::key>`**: reverse-traverse cross links →
  every dependent service & caller. (impact analysis)
- **`graffiti system audit`**: dangling consumers, orphan providers (provided,
  never consumed), ambiguous matches → text report. (audit)
- **`graffiti system status`**: which services drifted since last federation (SHA).
- **`graffiti system query "<q>"`** / MCP over the federated index: cross-service
  navigation (reuse query engine + `CombinedIndex`).

## Determinism

The system overlay is a pure function of the **set of input artifacts** (sorted by
service name); each service pinned by commit SHA in `system.json`. Contract arrays
are sorted (kind, key, file, line). `system.html` follows Plan 9's data-island
determinism.

## Doctrine fit

CLI-first, single static binary, $0, fully offline (git-as-registry), no daemon.
Orchestration = CI jobs + committable registry + derived overlay — Plan 7's
philosophy, scaled to published artifacts and automatic links.

## Honest limits (v1)

- Contract extraction is framework-shaped. v1 covers declared specs (OpenAPI/proto)
  + `graffiti.contract.json` + recognizers for Go net/http·gin·chi·echo, Flask,
  FastAPI, Spring (`@*Mapping`/`@RequestMapping`/`@KafkaListener`), NestJS
  (`@Controller`/`@Get…`/`@MessagePattern`), Kafka/NATS producers & subscribers, and
  frontend HTTP clients (React/Vue/Angular/Svelte/Nuxt). Frameworks outside this set
  under-extract → declare them in `graffiti.contract.json`. Confidence is always
  surfaced; low-confidence is never asserted as fact. Class-prefix combination is
  best-effort (Spring/Nest); deeply nested or programmatically-built routes may be missed.
- Object-store/artifact-registry adapters beyond git/dir are follow-ups.
