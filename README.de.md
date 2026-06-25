# 🕸️ graffiti — verwandle jedes Repository in einen abfragbaren Code-Graphen für KI

> Ein einziger Befehl verwandelt dein Repository in einen **gerichteten
> Wissensgraphen**, den dein KI-Coding-Assistent liest, statt blind zu greppen. Ein
> einzelnes statisches Go-Binary —
> **keine API-Schlüssel, $0, vollständig offline, byte-deterministisch.** Parst **Go, Python,
> JavaScript, TypeScript, Rust, Java und PHP**. Liefert ein LLM-freies `query`, einen
> **MCP**-Server, **Claude Code**-Integration, einen interaktiven Offline-Graph-Viewer
> und Multi-Repo-Workspace-Föderation.

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**Sprachen:** English · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

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

## Warum es das gibt

Ein KI-Coding-Assistent ist nur so gut wie das, was er *sehen* kann. Setze ihn in ein
großes Repository und er macht, was du ohne Karte tun würdest: Er greppt, öffnet ein paar
Dateien, rät. Er sieht nie die **Gestalt** des Codes — welche Funktion welche aufruft, wo ein Typ
definiert ist, welches Modul die tragende Wand ist.

**graffiti ist die Karte, die schon immer hätte da sein sollen.** Ein einziger Befehl parst das Repository
mit [tree-sitter](https://tree-sitter.github.io/tree-sitter/), löst die Kanten auf,
clustert die Module und schreibt einen Graphen — als JSON für die Maschine, als Markdown für
dich und als einzelnes Offline-HTML, das du dir tatsächlich ansehen kannst. Keine Schlüssel. Keine Cloud. Keine Kosten.

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

Lege eine Version oder ein Verzeichnis fest:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

Der Installer wählt das passende statische Binary für dein OS/deine Architektur, verifiziert dessen SHA256
gegen das Release-Manifest und installiert es. Überprüfe mit `graffiti version`.
Oder baue aus dem Quellcode (siehe unten).

## ⚡ Installation mit Claude Code (Vibe-Coding)

<!-- vibe-install -->
Kein Terminal nötig — lass **Claude Code** die ganze Arbeit erledigen. Füge diesen einen Prompt
in eine Claude-Code-Sitzung ein und bestätige jeden Schritt mit `y`. Er lädt das passende Binary herunter,
baut die Karte für dein Repository, richtet die Integration ein und öffnet den Graphen:

```text
Installiere graffiti von amazopic für mich. Lade das passende statische Binary für mein OS/meine Architektur aus dem neuesten Release unter github.com/amazopic/graffiti herunter (oder baue es aus dem Quellcode mit `make build`, falls Go verfügbar ist), lege es als `graffiti` in meinen PATH und überprüfe es mit `graffiti version`. Führe dann `graffiti .` im Wurzelverzeichnis meines Repositorys aus, um die Karte zu bauen, führe `graffiti init --hook` aus, um graffiti in Claude Code einzubinden, und öffne schließlich `.graffiti/map.html`, damit ich den Graphen sehen kann. Frage vor jedem Schritt nach.
```

<!-- quickstart -->
## Schnellstart (60 Sekunden)

```bash
# 1 — installieren (oder aus dem Quellcode mit `make build` bauen)
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh

# 2 — dein Repository kartieren (schreibt .graffiti/map.json, MAP.md, map.html)
cd your-repo
graffiti .

# 3 — sieh dir den Graphen an
open .graffiti/map.html        # macOS — unter Linux `xdg-open`, unter Windows `start` verwenden

# 4 — stelle ihm Fragen: kein LLM, kein API-Schlüssel
graffiti query "where is the user authenticated"
```

Binde es dann einmalig in deinen KI-Assistenten ein:

```bash
graffiti init --hook    # Claude Code: Skill + CLAUDE.md + grep→query-Hinweis
graffiti serve          # oder die Karte jedem MCP-Client über stdio bereitstellen
```

**Weitere Beispielfragen** — `query` gibt einen eingegrenzten Teilgraphen innerhalb eines
weichen Budgets von ~2.000 Token zurück, sodass der Kontext klein und günstig bleibt (setze die Frage in Anführungszeichen):

```bash
graffiti query "login handler"
graffiti query "what does the checkout flow touch"
graffiti query "where is the cart fetched" ../shop   # einen anderen Pfad ansteuern
```
<!-- /quickstart -->

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

Die `grammar_subset`-Build-Tags liefern nur die Grammatiken, die graffiti unterstützt (Go,
Python, JS, TS, Rust, Java, PHP, plus go.mod), über die reine Go-Laufzeit
`github.com/odvcencio/gotreesitter` (kein CGO, kein WASM). Sie halten das Binary bei
~10 MB; ohne sie kompiliert der Code zwar weiterhin, linkt aber den vollständigen Grammatiksatz
(~31 MB). Übergib sie immer — das Makefile erledigt das für dich.

## Unterstützte Sprachen

| Sprache | Extrahiert |
|----------|-----------|
| Go | Dateien, Funktionen, Methoden (nach Receiver), Typen, Imports, aufgelöste Aufrufe |
| Python, JavaScript, TypeScript, Rust, Java, PHP | Dateien, Funktionen, Klassen/Structs/Interfaces/Enums/Traits, Methoden (`Class.method`), Imports, Repo-interne Aufrufe |
| Markdown | Dokumentationsknoten |

Die Nicht-Go-Extraktion ist bewusst ehrlich: Sie erfasst die gängige, hochwertige
Struktur und **unter-extrahiert** exotische Konstrukte (Decorators, Generics, verschachtelte
Definitionen, dynamisches Dispatch), statt Vermutungen auszugeben.

## Verwendung

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

Führe `graffiti` ohne Argumente aus, um die vollständige Befehlsliste zu erhalten.

## Ein Befehl, drei Artefakte

`graffiti .` schreibt alles nach `<repo>/.graffiti/`:

- **`map.json`** — der Graph selbst: Knoten, Kanten, Communities, schema-geprüft
  gegen `schema/map.schema.json`. Das ist es, was deine KI liest und was `query`
  und der MCP-Server durchlaufen.
- **`MAP.md`** — eine menschenlesbare Zusammenfassung: die wichtigsten Module, die am stärksten verbundenen Knoten
  und die drei interessantesten Fragen, die deine Karte beantworten kann.
- **`map.html`** — ein einzelner, in sich geschlossener, offline-fähiger, interaktiver **kräftegerichteter
  Graph**. Kein CDN, kein Server, kein Netzwerk — öffne einfach die Datei.

`map.html` hat einen **2D/3D-Umschalter** (Hovern hebt einen Knoten und seine Nachbarn an), **Knoten-Suche**,
**Klick-zum-Kopieren von `file:line`**, **Sektorzonen**, **client / tests /
external**-Kategorie-Umschalter und einen größenverstellbaren **Projekt → Verzeichnis → Datei**-Baum
mit Anzeigen/Ausblenden-Checkboxen. Er ist CSP-sicher und funktioniert vollständig offline.

Ein Content-Hash-Cache pro Datei liegt unter `<repo>/.graffiti/cache/`, sodass erneute Läufe
nur das neu parsen, was sich geändert hat.

## Claude Code-Integration

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` schreibt:

- `.claude/skills/graffiti/SKILL.md` — einen kurzen Skill, damit Claude Code weiß, wie es die Karte baut/liest/abfragt.
- einen `CLAUDE.md`-Block (zwischen `<!-- graffiti:start -->` / `<!-- graffiti:end -->`), der dem
  Assistenten sagt, `graffiti query` gegenüber grep zu bevorzugen, wenn eine Karte existiert.
- mit `--hook` einen `.claude/settings.json`-PreToolUse-Eintrag, der `graffiti hook` ausführt, was vor
  `Grep`/`Glob` einen einzeiligen Hinweis hinzufügt, wenn `.graffiti/map.json` vorhanden ist. Der Hook blockiert niemals ein Tool.

Es ist idempotent — führe es jederzeit erneut aus; bestehende `CLAUDE.md` / `settings.json`-Inhalte bleiben erhalten.

## Abfragen ohne LLM

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` gibt einen relevanten Ausschnitt des Graphen innerhalb eines weichen Budgets von ~2000 Token pro Knoten
zurück — kein Modell, keine Embeddings. Setze die Frage in Anführungszeichen.

## MCP-Server

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

Richte einen beliebigen MCP-fähigen Client darauf aus und dein Assistent durchläuft den Graphen über
Tools, statt zu greppen.

## Workspaces (Multi-Repo-Föderation)

Lege separate Repositories nebeneinander und frage über sie hinweg ab — **ohne sie zu mergen**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` schreibt eine commitbare Registry (`.graffiti-workspace/workspace.json`)
und einen abgeleiteten, per gitignore ausschließbaren Cache (`.graffiti-workspace/overlay.json`). Das eigene
`.graffiti/map.json` jedes Repositories bleibt unverändert und funktioniert weiterhin eigenständig — der Workspace ist ein
schlankes berechnetes Overlay, niemals ein zusammengeführter Blob.

**Projektübergreifende Verknüpfungen:** Deklariere sie explizit in `.graffiti-workspace/links`,
eine pro Zeile — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(`#`-Kommentare erlaubt; Endpunkte sind `alias::nodeid`). `graffiti links check` validiert,
dass beide Endpunkte auflösen; `graffiti federate --explain` listet jede Verknüpfung auf. Die föderierte Abfrage
versieht jeden Knoten mit dem Alias seines Mitglieds als Präfix und durchläuft die projektübergreifenden Verknüpfungen. `graffiti workspace
render` schreibt eine `workspace.html` — denselben kräftegerichteten Graph-Viewer mit den **Projekten als
oberster Ebene** des Baums und eingezeichneten projektübergreifenden Verknüpfungen.

Füge `.graffiti-workspace/overlay.json` zu `.gitignore` hinzu (es ist abgeleitet und neu berechenbar).

## 🛰️ System-Orchestrierung — viele Services, ein Graph

<!-- system-orchestration -->
Ein Microservice-System besteht aus vielen unabhängigen Repos, die zusammen ein Produkt
bilden. graffiti kartiert jedes davon und **entdeckt dann die Kanten zwischen ihnen** —
HTTP, gRPC, Queues — aus der *Vertragsoberfläche* jedes Services (was er `provides`
bereitstellt und was er `consumes` verbraucht). Kein händisches Verdrahten: Jeder Service
veröffentlicht seine eigene Karte; der Orchestrator föderiert die veröffentlichten Artefakte
und ordnet Konsumenten den Anbietern zu.

```bash
# in der CI jedes Services (oder lokal) — seine Karte in einen gemeinsamen Speicher veröffentlichen:
graffiti publish --to ../system-store --as carts

# dann, in der CI oder bei Bedarf, über das gesamte System hinweg:
graffiti system build       # föderieren + projektübergreifende Verknüpfungen automatisch entdecken
graffiti system render      # → .graffiti-system/system.html (Services als Spuren)
graffiti system impact carts::"GET /carts/{}"   # wer bricht, wenn sich dies ändert?
graffiti system audit       # hängende Konsumenten · verwaiste Anbieter · mehrdeutige (CI-Gate)
graffiti system query "where is the cart fetched and served"
```

Jede Karte trägt eine **Vertragsoberfläche**, die aus `openapi.json`, `.proto`,
Framework-Routen, Queue-Aufrufen oder einer expliziten `graffiti.contract.json` extrahiert wird.
Projektübergreifende Verknüpfungen werden nach Konfidenz bewertet; **mehrdeutige** und
**hängende** (Dead-Endpoint-)Konsumenten werden gemeldet, niemals stillschweigend verworfen.
Der System-Speicher ist einfach ein Verzeichnis oder git-Repo — $0, offline, neu berechenbar.

<!-- system-walkthrough -->
### Ein Ordner voller Services, Schritt für Schritt

Angenommen, deine Services liegen in einem gemeinsamen übergeordneten Ordner, jeder in seinem eigenen Verzeichnis:

```text
myproject/                ← übergeordneter Ordner = der gemeinsame "System-Speicher"
├── orders/               ← ein Service (Go)
├── web/                  ← ein Service (React/TS)
└── payments/             ← ein Service (Python)
```

**1. Baue und veröffentliche jeden Service** in einen Speicher im übergeordneten Ordner (`--to .`).
`publish` verwendet eine bestehende Karte wieder, baue also zuerst, um Code-Änderungen zu erfassen:

```bash
cd myproject
for d in */; do
  d=${d%/}
  graffiti build "$d" && graffiti publish "$d" --to .
done
```

Der Service-Name entspricht standardmäßig seinem Ordnernamen; überschreibe ihn mit `--as <name>`.

> ⚠️ **Beim erneuten Veröffentlichen:** `publish` baut eine bestehende Karte **nicht** neu. Nach
> Code-Änderungen führe immer zuerst `graffiti build <service>` aus (die Schleife oben tut
> das) und dann `publish` — andernfalls veröffentlichst du eine veraltete Karte.

**2. Baue den System-Graphen** — föderiere die Karten und entdecke die Verknüpfungen automatisch:

```bash
graffiti system build
# ✓ System "myproject": 3 services → 7 cross-service links (0 ambiguous, 0 dangling, 2 orphan). 0 API calls, $0.
```

**3. Nutze ihn:**

```bash
graffiti system render          # → .graffiti-system/system.html (Services als oberste Baumebene)
graffiti system impact orders   # wer bricht, wenn sich orders ändert (direkt + transitiv)
graffiti system audit           # hängende Konsumenten · verwaiste Anbieter · mehrdeutige (Exit-Code ≠ 0 → CI-Gate)
graffiti system status          # welche Services seit dem letzten Build abgewichen sind
graffiti system query "where is the order created"   # LLM-freie Abfrage über das gesamte System
graffiti system list            # registrierte Services
```

**Was im übergeordneten Ordner landet:**

```text
myproject/.graffiti-system/
├── system.json                 # die Registry der Services (committe dies)
├── overlay.json                # entdeckte Verknüpfungen (abgeleitet — kann gefahrlos in .gitignore)
├── system.html                 # die visuelle Systemkarte
└── services/<name>/map.json    # die veröffentlichte Karte jedes Services
```

**Verbessere die Verknüpfungsgenauigkeit.** Die Auto-Erkennung deckt Go (net/http, gin/chi/echo),
Flask, FastAPI, Django/DRF, Spring, NestJS, ASP.NET, Ktor, Frontend-Clients
(React/Vue/Angular/Svelte), gRPC und Kafka/NATS ab. Wo das nicht genügt, lege
eine dieser Dateien in ein Service-Wurzelverzeichnis (höchste Konfidenz zuerst):

| Datei | Liefert |
|------|-------|
| `graffiti.contract.json` | explizite `provides` / `consumes` — beliebiger Stack, höchste Konfidenz |
| `openapi.json` / `swagger.json` | HTTP-Routen als `provides` |
| `*.proto` | gRPC-Methoden als `provides` |

Minimale `graffiti.contract.json`:

```json
{
  "provides": [{ "kind": "http", "name": "GET /orders/{id}" }],
  "consumes": [{ "kind": "rpc",  "name": "Payments.Charge" }]
}
```

**Sichere die CI gegen tote Endpunkte ab** — `audit` beendet sich mit Exit-Code ≠ 0, wenn ein Konsument auf
einen Endpunkt zeigt, den nichts bereitstellt:

```bash
graffiti system build && graffiti system audit
```
<!-- /system-walkthrough -->

## Wie es funktioniert

tree-sitter-Parsing (reines Go, kein CGO) → Kantenauflösung → Clustern in
Communities → leichtgewichtige Analyse → deterministische Serialisierung. Kein Modell, keine
Embeddings, kein Netzwerk — nur statische Analyse. Deshalb ist es kostenlos, privat und
reproduzierbar.

## Garantien

- **0 API-Aufrufe, $0, vollständig offline.** Nichts über deinen Code verlässt deine Maschine.
- **Deterministisch:** dasselbe Repository → byte-identisches `map.json`, abgesehen vom einzelnen
  `generated_at`-Zeitstempel und dem `root`-Basisnamen. Committe es; diffe es.
- **Einzelnes statisches Binary**, keine Laufzeitabhängigkeiten, keine C-Toolchain.

## Lizenz

Source-Available — lies und führe graffiti frei auf deinen eigenen Repositories aus, aber jede
Wiederverwendung, Weiterverbreitung, Fork oder Aufnahme in ein anderes Projekt erfordert die vorherige schriftliche
Genehmigung des Autors. Siehe [LICENSE](LICENSE).

## Autor

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
