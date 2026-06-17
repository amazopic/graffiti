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
