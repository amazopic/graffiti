# 🕸️ graffiti — convierte cualquier repo en un grafo de código consultable para IA

> Un solo comando convierte tu repositorio en un **grafo de conocimiento dirigido**
> que tu asistente de codificación con IA lee en lugar de buscar a ciegas con grep.
> Un único binario estático de Go — **cero claves de API, $0, totalmente sin conexión,
> determinista a nivel de byte.** Analiza **Go, Python, JavaScript, TypeScript, Rust,
> Java y PHP**. Incluye un `query` sin LLM, un servidor **MCP**, integración con
> **Claude Code**, un visor de grafos interactivo sin conexión y federación de
> espacios de trabajo multi-repo.

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**Idiomas:** [English](README.md) · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · Español · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

🌐 **Sitio web:** https://amazopic.github.io/graffiti/

```text
$ graffiti .
✓ Done. 0 API calls, $0.  214 files → 1,883 nodes, 4,102 edges, 12 communities.
  The 3 most interesting questions your map can answer:
    1) Which module is the load-bearing wall?
    2) What does the auth flow touch?
    3) Where are the cross-package call hotspots?
```

---

## Por qué existe

Un asistente de codificación con IA es tan bueno como lo que puede *ver*. Suéltalo en
un repo grande y hará lo mismo que tú harías sin un mapa: busca con grep, abre unos
cuantos archivos, adivina. Nunca ve la **forma** del código — qué función llama a cuál,
dónde se define un tipo, qué módulo es el muro de carga.

**graffiti es el mapa que debería haber estado ahí.** Un solo comando analiza el repo
con [tree-sitter](https://tree-sitter.github.io/tree-sitter/), resuelve las aristas,
agrupa los módulos y escribe un grafo — como JSON para la máquina, como Markdown para
ti y como un único HTML sin conexión que de verdad puedes mirar. Sin claves. Sin nube.
Sin costo.

## Instalación

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

Fija una versión o un directorio:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

El instalador elige el binario estático adecuado para tu SO/arquitectura, verifica su
SHA256 contra el manifiesto de la versión e instala. Verifica con `graffiti version`.
O compila desde el código fuente (abajo).

## ⚡ Instalar con Claude Code (vibe-code)

<!-- vibe-install -->
No necesitas la terminal — deja que **Claude Code** se encargue de todo. Pega este único
prompt en una sesión de Claude Code y responde `y` en cada paso. Descarga el binario
correcto, construye el mapa de tu repo, configura la integración y abre el grafo:

```text
Instálame graffiti de amazopic. Descarga el binario estático correcto para mi SO/arquitectura desde la última release en github.com/amazopic/graffiti (o compílalo desde el código fuente con `make build` si Go está disponible), ponlo en mi PATH como `graffiti` y verifícalo con `graffiti version`. Después ejecuta `graffiti .` en la raíz de mi repo para construir el mapa, ejecuta `graffiti init --hook` para integrar graffiti en Claude Code y, por último, abre `.graffiti/map.html` para que pueda ver el grafo. Pregunta antes de cada paso.
```

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

Los build tags `grammar_subset` incluyen solo las gramáticas que graffiti soporta (Go,
Python, JS, TS, Rust, Java, PHP, más go.mod) a través del runtime puro en Go
`github.com/odvcencio/gotreesitter` (sin CGO, sin WASM). Mantienen el binario en
~10 MB; sin ellos el código aún compila, pero enlaza el conjunto completo de gramáticas
(~31 MB). Pásalos siempre — el Makefile lo hace por ti.

## Lenguajes soportados

| Lenguaje | Extraído |
|----------|-----------|
| Go | archivos, funciones, métodos (por receptor), tipos, imports, llamadas resueltas |
| Python, JavaScript, TypeScript, Rust, Java, PHP | archivos, funciones, clases/structs/interfaces/enums/traits, métodos (`Class.method`), imports, llamadas dentro del repo |
| Markdown | nodos de documentación |

La extracción para lenguajes distintos de Go es deliberadamente honesta: captura la
estructura común y de alto valor, y **subextrae** las construcciones exóticas
(decoradores, genéricos, definiciones anidadas, despacho dinámico) en lugar de emitir
conjeturas.

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

Ejecuta `graffiti` sin argumentos para ver la lista completa de comandos.

## Un comando, tres artefactos

`graffiti .` escribe todo dentro de `<repo>/.graffiti/`:

- **`map.json`** — el grafo en sí: nodos, aristas, comunidades, validado contra el
  esquema `schema/map.schema.json`. Esto es lo que lee tu IA y lo que recorren `query`
  y el servidor MCP.
- **`MAP.md`** — un resumen legible por humanos: los módulos principales, los nodos más
  conectados y las tres preguntas más interesantes que tu mapa puede responder.
- **`map.html`** — un único **grafo dirigido por fuerzas** interactivo, autónomo y sin
  conexión. Sin CDN, sin servidor, sin red — solo abre el archivo.

`map.html` tiene un **conmutador 2D/3D** (al pasar el cursor se eleva un nodo y sus
vecinos), **búsqueda de nodos**, **clic para copiar `file:line`**, **zonas por sector**,
conmutadores de categoría **cliente / tests / externo**, y un árbol redimensionable
**proyecto → directorio → archivo** con casillas para mostrar/ocultar. Es seguro para
CSP y funciona completamente sin conexión.

Una caché de hash de contenido por archivo vive en `<repo>/.graffiti/cache/`, de modo
que las reejecuciones solo vuelven a analizar lo que cambió.

## Integración con Claude Code

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` escribe:

- `.claude/skills/graffiti/SKILL.md` — una skill breve para que Claude Code sepa que debe construir/leer/consultar el mapa.
- un bloque en `CLAUDE.md` (entre `<!-- graffiti:start -->` / `<!-- graffiti:end -->`) que le dice al
  asistente que prefiera `graffiti query` antes que grep cuando exista un mapa.
- con `--hook`, una entrada PreToolUse en `.claude/settings.json` que ejecuta `graffiti hook`, la cual añade
  una sugerencia de una línea antes de `Grep`/`Glob` cuando `.graffiti/map.json` está presente. El hook nunca bloquea una herramienta.

Es idempotente — reejecútalo en cualquier momento; se preserva el contenido existente de `CLAUDE.md` / `settings.json`.

## Consulta sin un LLM

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` devuelve una porción relevante del grafo dentro de un presupuesto suave de
~2000 tokens por nodo — sin modelo, sin embeddings. Pon la pregunta entre comillas.

## Servidor MCP

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

Apunta cualquier cliente compatible con MCP hacia él y tu asistente recorrerá el grafo
a través de herramientas en lugar de buscar con grep.

## Espacios de trabajo (federación multi-repo)

Coloca repos separados uno junto a otro y consúltalos en conjunto — **sin fusionarlos**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` escribe un registro versionable (`.graffiti-workspace/workspace.json`)
y una caché derivada e ignorable por git (`.graffiti-workspace/overlay.json`). El propio
`.graffiti/map.json` de cada repo permanece sin cambios y sigue funcionando de forma
autónoma — el espacio de trabajo es una fina capa calculada, nunca un bloque fusionado.

**Enlaces entre proyectos:** decláralos explícitamente en `.graffiti-workspace/links`,
uno por línea — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(se permiten comentarios con `#`; los extremos son `alias::nodeid`). `graffiti links check` valida
que ambos extremos se resuelvan; `graffiti federate --explain` lista cada enlace. La consulta federada
antepone a cada nodo el alias de su miembro y recorre los enlaces cruzados. `graffiti workspace
render` escribe un `workspace.html` — el mismo visor de grafo dirigido por fuerzas con los **proyectos como
nivel superior** del árbol y los enlaces entre proyectos dibujados.

Añade `.graffiti-workspace/overlay.json` a tu `.gitignore` (es derivado y recalculable).

## Cómo funciona

análisis con tree-sitter (puro en Go, sin CGO) → resolución de aristas → agrupamiento en
comunidades → análisis ligero → serialización determinista. Sin modelo, sin embeddings,
sin red — solo análisis estático. Por eso es gratis, privado y reproducible.

## Garantías

- **0 llamadas a la API, $0, totalmente sin conexión.** Nada de tu código sale de tu máquina.
- **Determinista:** mismo repo → `map.json` idéntico a nivel de byte salvo por el único
  timestamp `generated_at` y el basename de `root`. Hazle commit; haz diff.
- **Un único binario estático**, sin dependencias en tiempo de ejecución, sin toolchain de C.

## Licencia

Source-Available — lee y ejecuta graffiti libremente en tus propios repositorios, pero
cualquier reutilización, redistribución, fork o inclusión en otro proyecto requiere
permiso previo por escrito del autor. Consulta [LICENSE](LICENSE).

## Autor

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
