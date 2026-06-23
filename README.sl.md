# 🕸️ graffiti — vsak repozitorij spremenite v poizvedljiv graf kode za UI

> Z enim ukazom spremenite svoj repozitorij v **usmerjeni graf znanja**, ki ga
> vaš UI pomočnik za kodiranje bere namesto slepega grepanja. Ena sama statična
> binarna datoteka Go — **brez API-ključev, $0, povsem brez povezave,
> bajtno deterministična.** Razčlenjuje **Go, Python, JavaScript, TypeScript,
> Rust, Java in PHP**. Vključuje `query` brez LLM, strežnik **MCP**, integracijo
> s **Claude Code**, interaktivni pregledovalnik grafa brez povezave in
> federacijo delovnih prostorov med več repozitoriji.

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**Jeziki:** English · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

🌐 **Spletna stran:** https://amazopic.github.io/graffiti/

```text
$ graffiti .
✓ Done. 0 API calls, $0.  214 files → 1,883 nodes, 4,102 edges, 12 communities.
  The 3 most interesting questions your map can answer:
    1) Which module is the load-bearing wall?
    2) What does the auth flow touch?
    3) Where are the cross-package call hotspots?
```

---

## Zakaj to obstaja

UI pomočnik za kodiranje je le tako dober, kot je dobro tisto, kar lahko *vidi*.
Spustite ga v velik repozitorij in naredil bo enako, kot bi storili vi brez
zemljevida: grepa, odpre nekaj datotek, ugiba. Nikoli ne vidi **oblike** kode —
katera funkcija kliče katero, kje je definiran tip, kateri modul je nosilna stena.

**graffiti je zemljevid, ki bi moral biti tam.** En ukaz razčleni repozitorij s
[tree-sitter](https://tree-sitter.github.io/tree-sitter/), razreši povezave,
združi module v gruče in zapiše graf — kot JSON za stroj, kot Markdown za vas in
kot eno samo datoteko HTML brez povezave, ki si jo lahko dejansko ogledate.
Brez ključev. Brez oblaka. Brez stroškov.

## Namestitev

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

Pripnite različico ali imenik:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

Namestitveni program izbere pravo statično binarno datoteko za vaš OS/arhitekturo,
preveri njen SHA256 glede na manifest izdaje in jo namesti. Preverite z ukazom
`graffiti version`. Ali pa zgradite iz izvorne kode (spodaj).

## ⚡ Namestitev s Claude Code (vibe-code)

<!-- vibe-install -->
Brez terminala — pustite, da vse opravi **Claude Code**. Prilepite ta en sam poziv
v sejo Claude Code in pri vsakem koraku odgovorite z `y`. Prenese pravo binarno
datoteko, zgradi zemljevid za vaš repozitorij, vzpostavi integracijo in odpre graf:

```text
Namesti mi graffiti od amazopic. Iz zadnje izdaje na github.com/amazopic/graffiti prenesi pravo statično binarno datoteko za moj OS/arhitekturo (ali pa jo zgradi iz izvorne kode z `make build`, če je na voljo Go), jo postavi v moj PATH kot `graffiti` in preveri z `graffiti version`. Nato v korenu mojega repozitorija zaženi `graffiti .`, da zgradiš zemljevid, zaženi `graffiti init --hook`, da vključiš graffiti v Claude Code, in na koncu odpri `.graffiti/map.html`, da vidim graf. Pred vsakim korakom vprašaj.
```

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

Gradbene oznake `grammar_subset` vključujejo le slovnice, ki jih graffiti podpira
(Go, Python, JS, TS, Rust, Java, PHP, plus go.mod), prek čistega izvajalnega
okolja Go `github.com/odvcencio/gotreesitter` (brez CGO, brez WASM). Te ohranjajo
binarno datoteko pri ~10 MB; brez njih se koda še vedno prevede, vendar poveže
celoten nabor slovnic (~31 MB). Vedno jih posredujte — Makefile to stori namesto vas.

## Podprti jeziki

| Jezik | Izvlečeno |
|----------|-----------|
| Go | datoteke, funkcije, metode (po prejemniku), tipi, uvozi, razrešeni klici |
| Python, JavaScript, TypeScript, Rust, Java, PHP | datoteke, funkcije, razredi/strukture/vmesniki/enumi/traiti, metode (`Class.method`), uvozi, klici znotraj repozitorija |
| Markdown | dokumentacijska vozlišča |

Izvlek za jezike, ki niso Go, je namerno pošten: zajame pogosto, dragoceno
strukturo in **premalo izvleče** eksotične konstrukte (dekoratorje, generike,
gnezdene definicije, dinamično razpošiljanje), namesto da bi proizvajal ugibanja.

## Uporaba

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

Zaženite `graffiti` brez argumentov za celoten seznam ukazov.

## En ukaz, trije artefakti

`graffiti .` zapiše vse v `<repo>/.graffiti/`:

- **`map.json`** — sam graf: vozlišča, povezave, skupnosti, preverjeno glede na
  shemo `schema/map.schema.json`. To je tisto, kar bere vaš UI in kar prečkata
  `query` ter strežnik MCP.
- **`MAP.md`** — človeku berljiv povzetek: glavni moduli, najbolj povezana
  vozlišča in tri najbolj zanimiva vprašanja, na katera lahko odgovori vaš zemljevid.
- **`map.html`** — ena samostojna, interaktivna **silno usmerjena graf** brez
  povezave. Brez CDN, brez strežnika, brez omrežja — samo odprite datoteko.

`map.html` ima **preklop 2D/3D** (premik miške dvigne vozlišče in njegove sosede),
**iskanje vozlišč**, **klik za kopiranje `file:line`**, **sektorska območja**,
preklope kategorij **vaša koda / testi / zunanje** in spremenljivo veliko drevo
**projekt → imenik → datoteka** s potrditvenimi polji za prikaz/skrivanje. Je
varen za CSP in deluje povsem brez povezave.

Predpomnilnik vsebinskih zgoščenih vrednosti za vsako datoteko se nahaja v
`<repo>/.graffiti/cache/`, zato ponovni zagoni ponovno razčlenijo le tisto, kar se je spremenilo.

## Integracija s Claude Code

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` zapiše:

- `.claude/skills/graffiti/SKILL.md` — kratko spretnost, da Claude Code ve, kako
  zgraditi/brati/poizvedovati zemljevid.
- blok `CLAUDE.md` (med `<!-- graffiti:start -->` / `<!-- graffiti:end -->`), ki
  pomočniku naroči, naj raje uporabi `graffiti query` kot grep, kadar zemljevid obstaja.
- z `--hook` vnos PreToolUse v `.claude/settings.json`, ki zažene `graffiti hook`,
  ki doda enovrstični namig pred `Grep`/`Glob`, kadar je prisoten `.graffiti/map.json`.
  Kavelj nikoli ne blokira orodja.

Je idempotenten — znova ga zaženite kadarkoli; obstoječa vsebina `CLAUDE.md` /
`settings.json` se ohrani.

## Poizvedovanje brez LLM

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` vrne ustrezno rezino grafa znotraj mehkega proračuna vozlišč ~2000 žetonov
— brez modela, brez vstavitev. Vprašanje dajte v narekovaje.

## Strežnik MCP

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

Usmerite katerega koli odjemalca, ki podpira MCP, nanj in vaš pomočnik bo prečkal
graf prek orodij namesto z grepanjem.

## Delovni prostori (federacija med več repozitoriji)

Postavite ločene repozitorije drug ob drugega in poizvedujte po njih — **brez združevanja**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` zapiše register, ki ga je mogoče zavarovati v sistemu za nadzor
različic (`.graffiti-workspace/workspace.json`), in izpeljani predpomnilnik, ki ga
je mogoče izpustiti iz gita (`.graffiti-workspace/overlay.json`). Lasten
`.graffiti/map.json` vsakega repozitorija ostane nespremenjen in še vedno deluje
samostojno — delovni prostor je tanka izračunana prekrivna plast, nikoli združena gruda.

**Povezave med projekti:** izrecno jih navedite v `.graffiti-workspace/links`, po
eno na vrstico — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(komentarji `#` so dovoljeni; končne točke so `alias::nodeid`).
`graffiti links check` preveri, da se obe končni točki razrešita; `graffiti federate
--explain` našteje vsako povezavo. Federirana poizvedba vsako vozlišče opremi s
predpono z aliasom člana in prečka navzkrižne povezave. `graffiti workspace render`
zapiše `workspace.html` — isti pregledovalnik silnega grafa s **projekti kot
najvišjo ravnjo** drevesa in izrisanimi navzkrižnimi povezavami med projekti.

Dodajte `.graffiti-workspace/overlay.json` v `.gitignore` (je izpeljan in ga je mogoče znova izračunati).

## Kako deluje

razčlenjevanje s tree-sitter (čisti Go, brez CGO) → razrešitev povezav →
združevanje v skupnosti → lahka analiza → deterministična serializacija. Brez
modela, brez vstavitev, brez omrežja — samo statična analiza. Zato je brezplačno,
zasebno in ponovljivo.

## Jamstva

- **0 API-klicev, $0, povsem brez povezave.** Nič o vaši kodi ne zapusti vašega računalnika.
- **Deterministično:** isti repozitorij → bajtno identičen `map.json`, razen
  enega samega časovnega žiga `generated_at` in osnovnega imena `root`. Zavarujte
  ga; primerjajte ga.
- **Ena sama statična binarna datoteka**, brez izvajalnih odvisnosti, brez orodjarne C.

## Licenca

Source-Available — graffiti lahko prosto berete in zaganjate na svojih
repozitorijih, vendar vsakršna ponovna uporaba, redistribucija, razvejitev ali
vključitev v drug projekt zahteva predhodno pisno dovoljenje avtorja. Glejte
[LICENSE](LICENSE).

## Avtor

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
