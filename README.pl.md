# 🕸️ graffiti — zamień dowolne repozytorium w przeszukiwalny graf kodu dla AI

> Jedno polecenie zamienia Twoje repozytorium w **skierowany graf wiedzy**, który
> Twój asystent AI do kodowania czyta zamiast ślepo grepować. Pojedynczy statyczny
> plik binarny Go — **zero kluczy API, $0, w pełni offline, deterministyczny co do
> bajta.** Parsuje **Go, Python, JavaScript, TypeScript, Rust, Java i PHP**. Dostarcza
> niezależny od LLM `query`, serwer **MCP**, integrację z **Claude Code**, interaktywną
> przeglądarkę grafu działającą offline oraz federację przestrzeni roboczych obejmujących
> wiele repozytoriów.

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**Języki:** English · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

🌐 **Strona:** https://amazopic.github.io/graffiti/

```text
$ graffiti .
✓ Done. 0 API calls, $0.  214 files → 1,883 nodes, 4,102 edges, 12 communities.
  The 3 most interesting questions your map can answer:
    1) Which module is the load-bearing wall?
    2) What does the auth flow touch?
    3) Where are the cross-package call hotspots?
```

---

## Po co to powstało

Asystent AI do kodowania jest tylko tak dobry, jak to, co potrafi *zobaczyć*. Wrzuć go
do dużego repozytorium, a zrobi to, co Ty zrobiłbyś bez mapy: grepuje, otwiera kilka
plików, zgaduje. Nigdy nie widzi **kształtu** kodu — która funkcja wywołuje którą, gdzie
zdefiniowany jest dany typ, który moduł jest ścianą nośną.

**graffiti to mapa, której tam brakowało.** Jedno polecenie parsuje repozytorium za
pomocą [tree-sitter](https://tree-sitter.github.io/tree-sitter/), rozwiązuje krawędzie,
grupuje moduły i zapisuje graf — jako JSON dla maszyny, jako Markdown dla Ciebie i jako
pojedynczy plik HTML działający offline, na który możesz naprawdę spojrzeć. Bez kluczy.
Bez chmury. Bez kosztów.

## Instalacja

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

Przypnij wersję lub katalog:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

Instalator wybiera właściwy statyczny plik binarny dla Twojego systemu/architektury,
weryfikuje jego SHA256 względem manifestu wydania i instaluje go. Sprawdź poleceniem
`graffiti version`. Albo zbuduj ze źródeł (poniżej).

## ⚡ Instalacja przez Claude Code (vibe-code)

<!-- vibe-install -->
Bez terminala — niech **Claude Code** zrobi wszystko za Ciebie. Wklej ten jeden prompt
do sesji Claude Code i odpowiadaj `y` na każdym kroku. Pobierze właściwy plik binarny,
zbuduje mapę Twojego repozytorium, podłączy integrację i otworzy graf:

```text
Zainstaluj dla mnie graffiti autorstwa amazopic. Pobierz właściwy statyczny plik binarny dla mojego systemu/architektury z najnowszego wydania na github.com/amazopic/graffiti (albo zbuduj ze źródeł poleceniem `make build`, jeśli dostępne jest Go), umieść go w mojej zmiennej PATH jako `graffiti` i zweryfikuj poleceniem `graffiti version`. Następnie uruchom `graffiti .` w katalogu głównym mojego repozytorium, aby zbudować mapę, uruchom `graffiti init --hook`, aby podłączyć graffiti do Claude Code, a na koniec otwórz `.graffiti/map.html`, abym mógł zobaczyć graf. Pytaj przed każdym krokiem.
```

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

Tagi kompilacji `grammar_subset` dostarczają tylko gramatyki obsługiwane przez graffiti
(Go, Python, JS, TS, Rust, Java, PHP oraz go.mod) za pośrednictwem czystego środowiska
uruchomieniowego Go `github.com/odvcencio/gotreesitter` (bez CGO, bez WASM). Utrzymują
plik binarny na poziomie ~10 MB; bez nich kod nadal się kompiluje, ale linkuje pełny
zestaw gramatyk (~31 MB). Zawsze je przekazuj — Makefile robi to za Ciebie.

## Obsługiwane języki

| Język | Co jest wyodrębniane |
|----------|-----------|
| Go | pliki, funkcje, metody (po odbiorniku), typy, importy, rozwiązane wywołania |
| Python, JavaScript, TypeScript, Rust, Java, PHP | pliki, funkcje, klasy/struktury/interfejsy/enumy/traity, metody (`Class.method`), importy, wywołania wewnątrz repozytorium |
| Markdown | węzły dokumentacji |

Wyodrębnianie poza Go jest celowo uczciwe: wychwytuje powszechną, wartościową strukturę
i **niedoszacowuje** egzotyczne konstrukcje (dekoratory, generyki, zagnieżdżone
definicje, dynamiczne wiązanie) zamiast emitować domysły.

## Użycie

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

Uruchom `graffiti` bez argumentów, aby zobaczyć pełną listę poleceń.

## Jedno polecenie, trzy artefakty

`graffiti .` zapisuje wszystko do `<repo>/.graffiti/`:

- **`map.json`** — sam graf: węzły, krawędzie, społeczności, sprawdzone względem schematu
  `schema/map.schema.json`. To właśnie czyta Twoje AI oraz po czym poruszają się `query`
  i serwer MCP.
- **`MAP.md`** — czytelne dla człowieka streszczenie: najważniejsze moduły, najbardziej
  połączone węzły oraz trzy najciekawsze pytania, na które Twoja mapa potrafi odpowiedzieć.
- **`map.html`** — pojedynczy, samowystarczalny, offline'owy, interaktywny **graf typu
  force-directed**. Bez CDN, bez serwera, bez sieci — po prostu otwórz plik.

`map.html` ma **przełącznik 2D/3D** (najechanie unosi węzeł i jego sąsiadów),
**wyszukiwanie węzłów**, **kliknięcie kopiuje `file:line`**, **strefy sektorów**,
przełączniki kategorii **client / tests / external** oraz skalowalne drzewo
**projekt → katalog → plik** z polami wyboru pokaż/ukryj. Jest zgodny z CSP i działa
całkowicie offline.

Pamięć podręczna oparta na haszu treści każdego pliku znajduje się w
`<repo>/.graffiti/cache/`, więc ponowne uruchomienia ponownie parsują tylko to, co
się zmieniło.

## Integracja z Claude Code

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` zapisuje:

- `.claude/skills/graffiti/SKILL.md` — krótki skill, dzięki któremu Claude Code wie, jak
  budować/czytać/odpytywać mapę.
- blok `CLAUDE.md` (pomiędzy `<!-- graffiti:start -->` / `<!-- graffiti:end -->`),
  który informuje asystenta, aby przedkładał `graffiti query` nad grep, gdy mapa istnieje.
- z `--hook`, wpis PreToolUse w `.claude/settings.json` uruchamiający `graffiti hook`,
  który dodaje jednowierszową podpowiedź przed `Grep`/`Glob`, gdy obecny jest
  `.graffiti/map.json`. Hook nigdy nie blokuje narzędzia.

Jest idempotentny — uruchamiaj ponownie kiedykolwiek; istniejąca zawartość
`CLAUDE.md` / `settings.json` jest zachowywana.

## Odpytywanie bez LLM

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` zwraca istotny wycinek grafu w ramach miękkiego budżetu ~2000 tokenów na węzły —
bez modelu, bez osadzeń. Ujmij pytanie w cudzysłów.

## Serwer MCP

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

Skieruj na niego dowolnego klienta obsługującego MCP, a Twój asystent będzie poruszał się
po grafie za pomocą narzędzi zamiast grepować.

## Przestrzenie robocze (federacja wielu repozytoriów)

Ustaw osobne repozytoria obok siebie i odpytuj je łącznie — **bez scalania**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` zapisuje rejestr nadający się do commitowania
(`.graffiti-workspace/workspace.json`) oraz pochodną, możliwą do dodania do gitignore
pamięć podręczną (`.graffiti-workspace/overlay.json`). Własny `.graffiti/map.json`
każdego repozytorium pozostaje niezmieniony i nadal działa samodzielnie — przestrzeń
robocza to cienka, wyliczana nakładka, nigdy scalony blok.

**Połączenia między projektami:** deklaruj je jawnie w `.graffiti-workspace/links`,
po jednym w wierszu — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(komentarze `#` dozwolone; punkty końcowe to `alias::nodeid`). `graffiti links check`
sprawdza, czy oba punkty końcowe się rozwiązują; `graffiti federate --explain` wypisuje
każde połączenie. Sfederowane zapytanie poprzedza każdy węzeł aliasem jego członka i
przemierza połączenia między projektami. `graffiti workspace render` zapisuje
`workspace.html` — tę samą przeglądarkę grafu typu force-graph z **projektami jako
najwyższym poziomem** drzewa oraz narysowanymi połączeniami między projektami.

Dodaj `.graffiti-workspace/overlay.json` do `.gitignore` (jest pochodny i można go
przeliczyć ponownie).

## Jak to działa

parsowanie tree-sitter (czysty Go, bez CGO) → rozwiązywanie krawędzi → grupowanie w
społeczności → lekka analiza → deterministyczna serializacja. Bez modelu, bez osadzeń,
bez sieci — wyłącznie analiza statyczna. Dlatego jest darmowy, prywatny i powtarzalny.

## Gwarancje

- **0 wywołań API, $0, w pełni offline.** Nic z Twojego kodu nie opuszcza Twojej maszyny.
- **Deterministyczny:** to samo repozytorium → identyczny co do bajta `map.json` z
  dokładnością do pojedynczego znacznika czasu `generated_at` i nazwy katalogu `root`.
  Commituj go; porównuj różnice.
- **Pojedynczy statyczny plik binarny**, bez zależności środowiska uruchomieniowego,
  bez zestawu narzędzi C.

## Licencja

Source-Available — czytaj i uruchamiaj graffiti swobodnie na własnych repozytoriach, ale
wszelkie ponowne wykorzystanie, redystrybucja, fork lub włączenie do innego projektu
wymaga uprzedniej pisemnej zgody autora. Zobacz [LICENSE](LICENSE).

## Autor

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
