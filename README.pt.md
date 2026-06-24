# 🕸️ graffiti — transforme qualquer repositório em um grafo de código consultável para IA

> Um único comando transforma seu repositório em um **grafo de conhecimento
> direcionado** que seu assistente de programação com IA lê em vez de fazer grep às
> cegas. Um único binário Go estático — **zero chaves de API, $0, totalmente offline,
> determinístico byte a byte.** Analisa **Go, Python, JavaScript, TypeScript, Rust,
> Java e PHP**. Inclui um `query` sem LLM, um servidor **MCP**, integração com
> **Claude Code**, um visualizador de grafo interativo e offline, e federação de
> workspace multi-repositório.

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**Idiomas:** [English](README.md) · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

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

## Por que isto existe

Um assistente de programação com IA é tão bom quanto aquilo que ele consegue *ver*.
Coloque-o em um repositório grande e ele faz o que você faria sem um mapa: faz grep,
abre alguns arquivos, adivinha. Ele nunca enxerga a **forma** do código — qual função
chama qual, onde um tipo é definido, qual módulo é a parede de sustentação.

**graffiti é o mapa que deveria estar ali.** Um único comando analisa o repositório
com [tree-sitter](https://tree-sitter.github.io/tree-sitter/), resolve as arestas,
agrupa os módulos e escreve um grafo — como JSON para a máquina, como Markdown para
você e como um único HTML offline que você pode realmente examinar. Sem chaves. Sem
nuvem. Sem custo.

## Instalação

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

Fixe uma versão ou diretório:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

O instalador escolhe o binário estático certo para seu sistema operacional/arquitetura,
verifica o SHA256 dele contra o manifesto de release, e o instala. Verifique com
`graffiti version`. Ou compile a partir do código-fonte (abaixo).

## ⚡ Instale com o Claude Code (vibe-code)

<!-- vibe-install -->
Sem precisar de terminal — deixe o **Claude Code** fazer tudo. Cole este único prompt
em uma sessão do Claude Code e responda `y` em cada etapa. Ele baixa o binário certo,
constrói o mapa do seu repositório, configura a integração e abre o grafo:

```text
Instale o graffiti, do amazopic, para mim. Baixe o binário estático certo para o meu SO/arquitetura a partir do último release em github.com/amazopic/graffiti (ou compile a partir do código-fonte com `make build` se o Go estiver disponível), coloque-o no meu PATH como `graffiti` e verifique com `graffiti version`. Em seguida, execute `graffiti .` na raiz do meu repositório para construir o mapa, execute `graffiti init --hook` para integrar o graffiti ao Claude Code e, por fim, abra o `.graffiti/map.html` para que eu possa ver o grafo. Pergunte antes de cada etapa.
```

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

As build tags `grammar_subset` incluem apenas as gramáticas que o graffiti suporta (Go,
Python, JS, TS, Rust, Java, PHP, além de go.mod) por meio do runtime puro em Go
`github.com/odvcencio/gotreesitter` (sem CGO, sem WASM). Elas mantêm o binário em
~10 MB; sem elas o código ainda compila, mas vincula o conjunto completo de gramáticas
(~31 MB). Sempre passe-as — o Makefile faz isso por você.

## Linguagens suportadas

| Linguagem | Extraído |
|----------|-----------|
| Go | arquivos, funções, métodos (por receiver), tipos, imports, chamadas resolvidas |
| Python, JavaScript, TypeScript, Rust, Java, PHP | arquivos, funções, classes/structs/interfaces/enums/traits, métodos (`Class.method`), imports, chamadas intra-repositório |
| Markdown | nós de documentação |

A extração de linguagens que não são Go é intencionalmente honesta: ela captura a
estrutura comum e de alto valor, e **subextrai** construções exóticas (decorators,
generics, definições aninhadas, dispatch dinâmico) em vez de emitir suposições.

## Uso

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

Execute `graffiti` sem argumentos para a lista completa de comandos.

## Um comando, três artefatos

`graffiti .` escreve tudo em `<repo>/.graffiti/`:

- **`map.json`** — o grafo em si: nós, arestas, comunidades, validados contra o schema
  `schema/map.schema.json`. É isto que sua IA lê e o que o `query` e o servidor MCP
  percorrem.
- **`MAP.md`** — um resumo legível por humanos: principais módulos, os nós mais
  conectados, e as três perguntas mais interessantes que seu mapa pode responder.
- **`map.html`** — um único **grafo de força** autocontido, offline e interativo.
  Sem CDN, sem servidor, sem rede — basta abrir o arquivo.

`map.html` tem um **toggle 2D/3D** (passar o mouse eleva um nó e seus vizinhos),
**busca de nós**, **clique para copiar `file:line`**, **zonas de setor**, toggles de
categoria **client / tests / external**, e uma árvore redimensionável **project →
directory → file** com checkboxes de mostrar/ocultar. É seguro para CSP e funciona
inteiramente offline.

Um cache de hash de conteúdo por arquivo fica em `<repo>/.graffiti/cache/`, então as
reexecuções só reanalisam o que mudou.

## Integração com Claude Code

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` escreve:

- `.claude/skills/graffiti/SKILL.md` — uma skill curta para que o Claude Code saiba
  construir/ler/consultar o mapa.
- um bloco `CLAUDE.md` (entre `<!-- graffiti:start -->` / `<!-- graffiti:end -->`)
  dizendo ao assistente para preferir `graffiti query` em vez de grep quando um mapa
  existir.
- com `--hook`, uma entrada PreToolUse em `.claude/settings.json` executando
  `graffiti hook`, que adiciona um aviso de uma linha antes de `Grep`/`Glob` quando
  `.graffiti/map.json` está presente. O hook nunca bloqueia uma ferramenta.

É idempotente — reexecute quando quiser; o conteúdo existente de `CLAUDE.md` /
`settings.json` é preservado.

## Consulta sem LLM

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` retorna uma fatia relevante do grafo dentro de um orçamento flexível de ~2000
tokens por nó — sem modelo, sem embeddings. Coloque a pergunta entre aspas.

## Servidor MCP

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

Aponte qualquer cliente compatível com MCP para ele e seu assistente percorre o grafo
através de ferramentas em vez de fazer grep.

## Workspaces (federação multi-repositório)

Coloque repositórios separados lado a lado e consulte através deles — **sem fundi-los**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` escreve um registro versionável (`.graffiti-workspace/workspace.json`)
e um cache derivado e ignorável pelo git (`.graffiti-workspace/overlay.json`). O próprio
`.graffiti/map.json` de cada repositório permanece inalterado e continua funcionando de
forma autônoma — o workspace é uma fina sobreposição calculada, nunca um blob fundido.

**Links entre projetos:** declare-os explicitamente em `.graffiti-workspace/links`, um
por linha — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(comentários com `#` são permitidos; os endpoints são `alias::nodeid`). `graffiti links check`
valida que ambos os endpoints resolvem; `graffiti federate --explain` lista cada link. A
consulta federada prefixa cada nó com o alias do seu membro e percorre os links cruzados.
`graffiti workspace render` escreve um `workspace.html` — o mesmo visualizador de grafo de
força com os **projetos como o nível superior** da árvore e os links entre projetos
desenhados.

Adicione `.graffiti-workspace/overlay.json` ao `.gitignore` (ele é derivado e recalculável).

## 🛰️ Orquestração de sistemas — muitos serviços, um grafo

<!-- system-orchestration -->
Um sistema de microsserviços é composto por muitos repositórios independentes que
formam um único produto. O graffiti mapeia cada um deles e então **descobre as arestas
entre eles** — HTTP, gRPC, filas — a partir da *superfície de contrato* de cada serviço
(o que ele fornece via `provides` e o que ele consome via `consumes`). Sem fiação
manual: cada serviço publica seu próprio mapa; o orquestrador federa os artefatos
publicados e casa os consumidores com os provedores.

```bash
# no CI de cada serviço (ou localmente) — publique seu mapa em um repositório compartilhado:
graffiti publish --to ../system-store --as carts

# então, no CI ou sob demanda, sobre todo o sistema:
graffiti system build       # federate + auto-discover cross-service links
graffiti system render      # → .graffiti-system/system.html (services as lanes)
graffiti system impact carts::"GET /carts/{}"   # who breaks if this changes?
graffiti system audit       # dangling consumers · orphan providers · ambiguous (CI gate)
graffiti system query "where is the cart fetched and served"
```

Cada mapa carrega uma **superfície de contrato** extraída de `openapi.json`, `.proto`,
rotas de framework, chamadas de fila ou de um `graffiti.contract.json` explícito. Os
links entre serviços são pontuados por confiança; consumidores **ambíguos** e
**pendentes** (com endpoint inexistente) são reportados, nunca descartados em silêncio.
O repositório do sistema é apenas um diretório ou um repositório git — $0, offline,
recalculável.

## Como funciona

análise com tree-sitter (puro Go, sem CGO) → resolução de arestas → agrupamento em
comunidades → análise leve → serialização determinística. Sem modelo, sem embeddings,
sem rede — apenas análise estática. É por isso que é gratuito, privado e reproduzível.

## Garantias

- **0 chamadas de API, $0, totalmente offline.** Nada sobre seu código sai da sua máquina.
- **Determinístico:** mesmo repositório → `map.json` idêntico byte a byte, exceto pelo
  único timestamp `generated_at` e pelo basename do `root`. Faça commit dele; faça diff dele.
- **Único binário estático**, sem dependências de runtime, sem toolchain C.

## Licença

Source-Available — leia e execute o graffiti livremente em seus próprios repositórios,
mas qualquer reúso, redistribuição, fork ou inclusão em outro projeto requer permissão
prévia por escrito do autor. Veja [LICENSE](LICENSE).

## Autor

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
