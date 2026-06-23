# 🕸️ graffiti — transformez n'importe quel dépôt en graphe de code interrogeable pour l'IA

> Une seule commande transforme votre dépôt en un **graphe de connaissances orienté**
> que votre assistant de code IA lit au lieu de faire du grep à l'aveugle. Un unique
> binaire Go statique — **zéro clé d'API, 0 $, entièrement hors ligne, déterministe à
> l'octet près.** Analyse **Go, Python, JavaScript, TypeScript, Rust, Java et PHP**.
> Livré avec une commande `query` sans LLM, un serveur **MCP**, une intégration
> **Claude Code**, une visionneuse de graphe interactive hors ligne et la fédération
> multi-dépôts en workspace.

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**Langues :** [English](README.md) · [Русский](README.ru.md) · Français · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

🌐 **Site web :** https://amazopic.github.io/graffiti/

```text
$ graffiti .
✓ Done. 0 API calls, $0.  214 files → 1,883 nodes, 4,102 edges, 12 communities.
  The 3 most interesting questions your map can answer:
    1) Which module is the load-bearing wall?
    2) What does the auth flow touch?
    3) Where are the cross-package call hotspots?
```

---

## Pourquoi cet outil existe

Un assistant de code IA ne vaut que ce qu'il peut *voir*. Lâchez-le dans un grand
dépôt et il fait ce que vous feriez sans carte : il fait du grep, ouvre quelques
fichiers, devine. Il ne voit jamais la **forme** du code — quelle fonction appelle
quoi, où un type est défini, quel module est le mur porteur.

**graffiti est la carte qui aurait dû être là.** Une seule commande analyse le dépôt
avec [tree-sitter](https://tree-sitter.github.io/tree-sitter/), résout les arêtes,
regroupe les modules et écrit un graphe — en JSON pour la machine, en Markdown pour
vous, et en un unique fichier HTML hors ligne que vous pouvez réellement regarder.
Aucune clé. Aucun cloud. Aucun coût.

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

Épinglez une version ou un répertoire :

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

L'installeur choisit le bon binaire statique pour votre OS/architecture, vérifie son
SHA256 par rapport au manifeste de la release et l'installe. Vérifiez avec `graffiti
version`. Ou compilez depuis les sources (ci-dessous).

## ⚡ Installation avec Claude Code (vibe-code)

<!-- vibe-install -->
Pas besoin de terminal — laissez **Claude Code** tout faire à votre place. Collez ce
seul prompt dans une session Claude Code et répondez `y` à chaque étape. Il récupère le
bon binaire, construit la carte de votre dépôt, met en place l'intégration et ouvre le
graphe :

```text
Installe graffiti d'amazopic pour moi. Télécharge le bon binaire statique pour mon OS/architecture depuis la dernière release sur github.com/amazopic/graffiti (ou compile-le depuis les sources avec `make build` si Go est disponible), place-le sur mon PATH sous le nom `graffiti`, et vérifie avec `graffiti version`. Ensuite, lance `graffiti .` à la racine de mon dépôt pour construire la carte, lance `graffiti init --hook` pour intégrer graffiti à Claude Code, et enfin ouvre `.graffiti/map.html` pour que je puisse voir le graphe. Demande-moi confirmation avant chaque étape.
```

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

Les tags de build `grammar_subset` ne livrent que les grammaires prises en charge par
graffiti (Go, Python, JS, TS, Rust, Java, PHP, plus go.mod) via le runtime pur Go
`github.com/odvcencio/gotreesitter` (sans CGO, sans WASM). Ils maintiennent le binaire
à ~10 Mo ; sans eux, le code compile toujours mais lie l'ensemble complet des
grammaires (~31 Mo). Passez-les toujours — le Makefile le fait pour vous.

## Langages pris en charge

| Langage | Extrait |
|----------|-----------|
| Go | fichiers, fonctions, méthodes (par récepteur), types, imports, appels résolus |
| Python, JavaScript, TypeScript, Rust, Java, PHP | fichiers, fonctions, classes/structs/interfaces/enums/traits, méthodes (`Class.method`), imports, appels intra-dépôt |
| Markdown | nœuds de documentation |

L'extraction hors Go est volontairement honnête : elle capture la structure courante à
forte valeur et **sous-extrait** les constructions exotiques (décorateurs, génériques,
définitions imbriquées, dispatch dynamique) plutôt que d'émettre des suppositions.

## Utilisation

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

Lancez `graffiti` sans argument pour obtenir la liste complète des commandes.

## Une commande, trois artefacts

`graffiti .` écrit tout dans `<repo>/.graffiti/` :

- **`map.json`** — le graphe lui-même : nœuds, arêtes, communautés, validés par le
  schéma `schema/map.schema.json`. C'est ce que lit votre IA et ce que parcourent
  `query` et le serveur MCP.
- **`MAP.md`** — un digest lisible par un humain : les principaux modules, les nœuds
  les plus connectés, et les trois questions les plus intéressantes auxquelles votre
  carte peut répondre.
- **`map.html`** — un unique **graphe à force dirigée** interactif, autonome et hors
  ligne. Aucun CDN, aucun serveur, aucun réseau — il suffit d'ouvrir le fichier.

`map.html` possède un **basculement 2D/3D** (le survol soulève un nœud et ses
voisins), une **recherche de nœuds**, le **clic pour copier `file:line`**, des **zones
de secteur**, des bascules de catégorie **client / tests / externe**, et une
arborescence redimensionnable **projet → répertoire → fichier** avec des cases à cocher
afficher/masquer. Il est conforme à la CSP et fonctionne entièrement hors ligne.

Un cache de hachage de contenu par fichier réside sous `<repo>/.graffiti/cache/`, de
sorte que les exécutions ultérieures ne réanalysent que ce qui a changé.

## Intégration Claude Code

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` écrit :

- `.claude/skills/graffiti/SKILL.md` — une courte compétence pour que Claude Code sache construire/lire/interroger la carte.
- un bloc `CLAUDE.md` (entre `<!-- graffiti:start -->` / `<!-- graffiti:end -->`) qui indique à
  l'assistant de préférer `graffiti query` à grep lorsqu'une carte existe.
- avec `--hook`, une entrée PreToolUse dans `.claude/settings.json` exécutant `graffiti hook`, qui ajoute un
  conseil d'une ligne avant `Grep`/`Glob` lorsque `.graffiti/map.json` est présent. Le hook ne bloque jamais un outil.

C'est idempotent — relancez à tout moment ; le contenu existant de `CLAUDE.md` / `settings.json` est préservé.

## Interroger sans LLM

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` renvoie une tranche pertinente du graphe dans une limite souple d'environ 2000
tokens de nœuds — sans modèle, sans embeddings. Mettez la question entre guillemets.

## Serveur MCP

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

Pointez n'importe quel client compatible MCP vers lui et votre assistant parcourt le
graphe au moyen d'outils au lieu de faire du grep.

## Workspaces (fédération multi-dépôts)

Posez des dépôts distincts côte à côte et interrogez-les ensemble — **sans les
fusionner** :

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` écrit un registre versionnable (`.graffiti-workspace/workspace.json`)
et un cache dérivé et ignorable par git (`.graffiti-workspace/overlay.json`). Le propre
`.graffiti/map.json` de chaque dépôt reste inchangé et fonctionne toujours de manière
autonome — le workspace est une fine surcouche calculée, jamais un agrégat fusionné.

**Liens inter-projets :** déclarez-les explicitement dans `.graffiti-workspace/links`,
un par ligne — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(les commentaires `#` sont autorisés ; les extrémités sont des `alias::nodeid`).
`graffiti links check` valide que les deux extrémités se résolvent ; `graffiti federate
--explain` liste chaque lien. La requête fédérée préfixe chaque nœud par l'alias de son
membre et parcourt les liens croisés. `graffiti workspace render` écrit un
`workspace.html` — la même visionneuse de graphe à force dirigée avec les **projets
comme niveau supérieur** de l'arborescence et les liens inter-projets tracés.

Ajoutez `.graffiti-workspace/overlay.json` à votre `.gitignore` (il est dérivé et recalculable).

## Comment ça marche

Analyse tree-sitter (pur Go, sans CGO) → résolution des arêtes → regroupement en
communautés → analyse légère → sérialisation déterministe. Aucun modèle, aucun
embedding, aucun réseau — uniquement de l'analyse statique. Voilà pourquoi c'est
gratuit, privé et reproductible.

## Garanties

- **0 appel d'API, 0 $, entièrement hors ligne.** Rien de votre code ne quitte votre machine.
- **Déterministe :** même dépôt → `map.json` identique à l'octet près, à l'exception du
  seul horodatage `generated_at` et du nom de base de `root`. Versionnez-le ; faites-en le diff.
- **Un unique binaire statique**, sans dépendance d'exécution, sans chaîne d'outils C.

## Licence

Source-Available — lisez et exécutez graffiti librement sur vos propres dépôts, mais
toute réutilisation, redistribution, fork ou inclusion dans un autre projet nécessite
l'autorisation écrite préalable de l'auteur. Voir [LICENSE](LICENSE).

## Auteur

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
