# 🕸️ graffiti — trasforma qualsiasi repository in un grafo di codice interrogabile per l'IA

> Un solo comando trasforma il tuo repository in un **grafo di conoscenza diretto**
> che il tuo assistente di programmazione IA legge invece di fare grep alla cieca.
> Un unico binario Go statico — **zero chiavi API, $0, completamente offline,
> deterministico a livello di byte.** Analizza **Go, Python, JavaScript,
> TypeScript, Rust, Java e PHP**. Include un `query` senza LLM, un server **MCP**,
> l'integrazione con **Claude Code**, un visualizzatore di grafi interattivo
> offline e la federazione di workspace multi-repository.

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**Lingue:** English · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

🌐 **Sito web:** https://amazopic.github.io/graffiti/

```text
$ graffiti .
✓ Done. 0 API calls, $0.  214 files → 1,883 nodes, 4,102 edges, 12 communities.
  The 3 most interesting questions your map can answer:
    1) Which module is the load-bearing wall?
    2) What does the auth flow touch?
    3) Where are the cross-package call hotspots?
```

---

## Perché esiste

Un assistente di programmazione IA è valido solo quanto ciò che riesce a *vedere*.
Mettilo in un repository di grandi dimensioni e farà ciò che faresti tu senza una
mappa: fa grep, apre qualche file, tira a indovinare. Non vede mai la **forma** del
codice — quale funzione chiama quale, dove è definito un tipo, quale modulo è il
muro portante.

**graffiti è la mappa che avrebbe dovuto esserci.** Un solo comando analizza il
repository con [tree-sitter](https://tree-sitter.github.io/tree-sitter/), risolve
gli archi, raggruppa i moduli e scrive un grafo — come JSON per la macchina, come
Markdown per te e come un unico HTML offline che puoi davvero guardare. Nessuna
chiave. Nessun cloud. Nessun costo.

## Installazione

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

Fissa una versione o una directory:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

L'installer sceglie il binario statico giusto per il tuo OS/architettura, ne
verifica lo SHA256 rispetto al manifest della release e lo installa. Verifica con
`graffiti version`. Oppure compila dai sorgenti (vedi sotto).

## ⚡ Installa con Claude Code (vibe-code)

<!-- vibe-install -->
Nessun terminale necessario — lascia che sia **Claude Code** a fare tutto. Incolla
questo singolo prompt in una sessione di Claude Code e rispondi `y` a ogni passaggio.
Scarica il binario giusto, costruisce la mappa per il tuo repository, configura
l'integrazione e apre il grafo:

```text
Installa graffiti di amazopic al posto mio. Scarica il binario statico giusto per il mio OS/architettura dall'ultima release su github.com/amazopic/graffiti (oppure compilalo dai sorgenti con `make build` se Go è disponibile), mettilo nel mio PATH come `graffiti` e verifica con `graffiti version`. Poi esegui `graffiti .` nella radice del mio repository per costruire la mappa, esegui `graffiti init --hook` per integrare graffiti in Claude Code e infine apri `.graffiti/map.html` così posso vedere il grafo. Chiedimi conferma prima di ogni passaggio.
```

<!-- quickstart -->
## Avvio rapido (60 secondi)

```bash
# 1 — installa (o compila dai sorgenti con `make build`)
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh

# 2 — mappa il tuo repository (scrive .graffiti/map.json, MAP.md, map.html)
cd your-repo
graffiti .

# 3 — guarda il grafo
open .graffiti/map.html        # macOS — usa `xdg-open` su Linux, `start` su Windows

# 4 — fagli delle domande: nessun LLM, nessuna chiave API
graffiti query "where is the user authenticated"
```

Poi integralo una volta nel tuo assistente IA:

```bash
graffiti init --hook    # Claude Code: skill + CLAUDE.md + suggerimento grep→query
graffiti serve          # oppure esponi la mappa a qualsiasi client MCP tramite stdio
```

**Altre domande di esempio** — `query` restituisce un sottografo circoscritto entro
un budget morbido di ~2.000 token, così il contesto resta piccolo ed economico
(metti la domanda tra virgolette):

```bash
graffiti query "login handler"
graffiti query "what does the checkout flow touch"
graffiti query "where is the cart fetched" ../shop   # punta a un altro percorso
```
<!-- /quickstart -->

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

I build tag `grammar_subset` includono solo le grammatiche supportate da graffiti
(Go, Python, JS, TS, Rust, Java, PHP, più go.mod) tramite il runtime puro-Go
`github.com/odvcencio/gotreesitter` (niente CGO, niente WASM). Mantengono il
binario a ~10 MB; senza di essi il codice si compila comunque, ma collega l'intero
set di grammatiche (~31 MB). Passali sempre — il Makefile lo fa per te.

## Linguaggi supportati

| Linguaggio | Estratto |
|----------|-----------|
| Go | file, funzioni, metodi (per receiver), tipi, import, chiamate risolte |
| Python, JavaScript, TypeScript, Rust, Java, PHP | file, funzioni, classi/struct/interfacce/enum/trait, metodi (`Class.method`), import, chiamate intra-repository |
| Markdown | nodi di documentazione |

L'estrazione non-Go è volutamente onesta: cattura la struttura comune e di alto
valore e **sotto-estrae** i costrutti esotici (decoratori, generici, definizioni
annidate, dispatch dinamico) invece di emettere ipotesi.

## Utilizzo

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

Esegui `graffiti` senza argomenti per l'elenco completo dei comandi.

## Un comando, tre artefatti

`graffiti .` scrive tutto in `<repo>/.graffiti/`:

- **`map.json`** — il grafo stesso: nodi, archi, comunità, validato rispetto allo
  schema `schema/map.schema.json`. È ciò che la tua IA legge e ciò che `query` e il
  server MCP attraversano.
- **`MAP.md`** — un riassunto leggibile dall'uomo: i moduli principali, i nodi più
  connessi e le tre domande più interessanti a cui la tua mappa può rispondere.
- **`map.html`** — un unico **grafo force-directed** interattivo, offline e
  autocontenuto. Niente CDN, niente server, niente rete — basta aprire il file.

`map.html` ha un **interruttore 2D/3D** (passandoci sopra il mouse solleva un nodo
e i suoi vicini), la **ricerca dei nodi**, il **click-per-copiare `file:line`**, le
**zone settoriali**, gli interruttori per le categorie **client / test / esterni** e
un albero ridimensionabile **progetto → directory → file** con caselle di
mostra/nascondi. È sicuro per la CSP e funziona interamente offline.

Una cache di hash dei contenuti per file risiede in `<repo>/.graffiti/cache/`,
così le riesecuzioni ri-analizzano solo ciò che è cambiato.

## Integrazione con Claude Code

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` scrive:

- `.claude/skills/graffiti/SKILL.md` — una skill breve così che Claude Code sappia
  come costruire/leggere/interrogare la mappa.
- un blocco `CLAUDE.md` (tra `<!-- graffiti:start -->` / `<!-- graffiti:end -->`)
  che indica all'assistente di preferire `graffiti query` a grep quando esiste una
  mappa.
- con `--hook`, una voce PreToolUse in `.claude/settings.json` che esegue
  `graffiti hook`, la quale aggiunge un suggerimento su una riga prima di
  `Grep`/`Glob` quando `.graffiti/map.json` è presente. L'hook non blocca mai uno
  strumento.

È idempotente — rieseguilo quando vuoi; i contenuti esistenti di `CLAUDE.md` /
`settings.json` vengono preservati.

## Interrogazione senza LLM

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` restituisce una porzione rilevante del grafo entro un budget morbido di
~2000 token per nodo — nessun modello, nessun embedding. Metti la domanda tra
virgolette.

## Server MCP

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

Punta qualsiasi client compatibile con MCP verso di esso e il tuo assistente
attraversa il grafo tramite strumenti invece di fare grep.

## Workspace (federazione multi-repository)

Disponi repository separati fianco a fianco e interrogali in modo trasversale —
**senza unirli**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` scrive un registro commit-abile (`.graffiti-workspace/workspace.json`)
e una cache derivata e ignorabile da git (`.graffiti-workspace/overlay.json`). La
`.graffiti/map.json` di ciascun repository rimane invariata e funziona ancora in
modo autonomo — il workspace è un sottile overlay calcolato, mai un blob unito.

**Link cross-progetto:** dichiarali esplicitamente in
`.graffiti-workspace/links`, uno per riga —
`frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(i commenti `#` sono ammessi; gli endpoint sono `alias::nodeid`).
`graffiti links check` verifica che entrambi gli endpoint si risolvano;
`graffiti federate --explain` elenca ogni link. La query federata prefigge ogni
nodo con l'alias del suo membro e attraversa i cross-link. `graffiti workspace
render` scrive un `workspace.html` — lo stesso visualizzatore force-graph con i
**progetti come livello superiore** dell'albero e i link cross-progetto disegnati.

Aggiungi `.graffiti-workspace/overlay.json` al `.gitignore` (è derivato e
ricalcolabile).

## 🛰️ Orchestrazione di sistema — molti servizi, un solo grafo

<!-- system-orchestration -->
Un sistema a microservizi è costituito da molti repository indipendenti che
formano un solo prodotto. graffiti mappa ciascuno di essi, poi **scopre gli archi
tra di loro** — HTTP, gRPC, code — a partire dalla *superficie di contratto* di
ogni servizio (ciò che `provides` e `consumes`, cioè ciò che offre e ciò che
consuma). Nessun cablaggio manuale: ogni servizio pubblica la propria mappa;
l'orchestratore federa gli artefatti pubblicati e abbina i consumatori ai
fornitori.

```bash
# nella CI di ogni servizio (o localmente) — pubblica la sua mappa in uno store condiviso:
graffiti publish --to ../system-store --as carts

# poi, nella CI o su richiesta, sull'intero sistema:
graffiti system build       # federate + auto-discover cross-service links
graffiti system render      # → .graffiti-system/system.html (services as lanes)
graffiti system impact carts::"GET /carts/{}"   # who breaks if this changes?
graffiti system audit       # dangling consumers · orphan providers · ambiguous (CI gate)
graffiti system query "where is the cart fetched and served"
```

Ogni mappa porta con sé una **superficie di contratto** estratta da `openapi.json`,
`.proto`, le rotte del framework, le chiamate alle code oppure un esplicito
`graffiti.contract.json`. I link cross-service vengono valutati per livello di
confidenza; i consumatori **ambigui** e **dangling** (endpoint morto) vengono
segnalati, mai scartati in silenzio. Lo store di sistema è semplicemente una
directory o un repository git — $0, offline, ricalcolabile. Vedi il
[documento di progettazione](docs/superpowers/specs/2026-06-24-graffiti-system-orchestration-design.md).

<!-- system-walkthrough -->
### Una cartella di servizi, passo dopo passo

Supponiamo che i tuoi servizi si trovino in un'unica cartella padre, ciascuno nella propria directory:

```text
myproject/                ← cartella padre = lo "store di sistema" condiviso
├── orders/               ← un servizio (Go)
├── web/                  ← un servizio (React/TS)
└── payments/             ← un servizio (Python)
```

**1. Compila e pubblica ogni servizio** in uno store nella cartella padre (`--to .`).
`publish` riutilizza una mappa esistente, quindi compila prima per recepire le modifiche al codice:

```bash
cd myproject
for d in */; do
  d=${d%/}
  graffiti build "$d" && graffiti publish "$d" --to .
done
```

Il nome del servizio assume per default il nome della sua cartella; sovrascrivilo con `--as <name>`.

> ⚠️ **Alla ri-pubblicazione:** `publish` **non** ricompila una mappa esistente. Dopo
> aver modificato il codice, esegui sempre prima `graffiti build <service>` (come fa
> il ciclo qui sopra) e poi `publish` — altrimenti pubblichi una mappa obsoleta.

**2. Costruisci il grafo di sistema** — federa le mappe e scopri automaticamente i link:

```bash
graffiti system build
# ✓ System "myproject": 3 services → 7 cross-service links (0 ambiguous, 0 dangling, 2 orphan). 0 API calls, $0.
```

**3. Usalo:**

```bash
graffiti system render          # → .graffiti-system/system.html (services as the top tree level)
graffiti system impact orders   # who breaks if orders changes (direct + transitive)
graffiti system audit           # dangling consumers · orphan providers · ambiguous (non-zero exit → CI gate)
graffiti system status          # which services drifted since the last build
graffiti system query "where is the order created"   # LLM-free retrieval across the whole system
graffiti system list            # registered services
```

**Cosa finisce nella cartella padre:**

```text
myproject/.graffiti-system/
├── system.json                 # il registro dei servizi (fanne il commit)
├── overlay.json                # link scoperti (derivato — sicuro da mettere in .gitignore)
├── system.html                 # la mappa visiva del sistema
└── services/<name>/map.json    # la mappa pubblicata di ciascun servizio
```

**Migliora l'accuratezza dei link.** Il rilevamento automatico copre Go (net/http,
gin/chi/echo), Flask, FastAPI, Django/DRF, Spring, NestJS, ASP.NET, Ktor, i client
frontend (React/Vue/Angular/Svelte), gRPC e Kafka/NATS. Dove ciò non basta, inserisci
uno di questi nella radice di un servizio (prima quelli a confidenza più alta):

| File | Fornisce |
|------|-----------|
| `graffiti.contract.json` | `provides` / `consumes` espliciti — qualsiasi stack, massima confidenza |
| `openapi.json` / `swagger.json` | rotte HTTP come `provides` |
| `*.proto` | metodi gRPC come `provides` |

`graffiti.contract.json` minimale:

```json
{
  "provides": [{ "kind": "http", "name": "GET /orders/{id}" }],
  "consumes": [{ "kind": "rpc",  "name": "Payments.Charge" }]
}
```

**Blocca la CI sugli endpoint morti** — `audit` esce con codice diverso da zero quando
un consumatore punta a un endpoint che nessuno fornisce:

```bash
graffiti system build && graffiti system audit
```
<!-- /system-walkthrough -->

## Come funziona

parsing con tree-sitter (puro-Go, niente CGO) → risoluzione degli archi →
clustering in comunità → analisi leggera → serializzazione deterministica. Nessun
modello, nessun embedding, nessuna rete — solo analisi statica. È per questo che è
gratuito, privato e riproducibile.

## Garanzie

- **0 chiamate API, $0, completamente offline.** Niente del tuo codice lascia la
  tua macchina.
- **Deterministico:** stesso repository → `map.json` identica a livello di byte, a
  meno del singolo timestamp `generated_at` e del basename `root`. Fanne il commit;
  fanne il diff.
- **Unico binario statico**, nessuna dipendenza a runtime, nessuna toolchain C.

## Licenza

Source-Available — leggi ed esegui graffiti liberamente sui tuoi repository, ma
qualsiasi riuso, ridistribuzione, fork o inclusione in un altro progetto richiede
il previo permesso scritto dell'autore. Vedi [LICENSE](LICENSE).

## Autore

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
