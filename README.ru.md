# 🕸️ graffiti — превратите любой репозиторий в запрашиваемый граф кода для ИИ

> Одна команда превращает ваш репозиторий в **ориентированный граф знаний**, который
> ваш ИИ-ассистент по программированию читает вместо слепого перебора через grep.
> Единый статический бинарник на Go —
> **ноль API-ключей, $0, полностью офлайн, побайтово детерминированный.** Разбирает **Go, Python,
> JavaScript, TypeScript, Rust, Java и PHP**. Поставляется с не использующим LLM `query`,
> сервером **MCP**, интеграцией с **Claude Code**, интерактивным офлайн-просмотрщиком
> графа и федерацией рабочих пространств для нескольких репозиториев.

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**Языки:** English · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

🌐 **Сайт:** https://amazopic.github.io/graffiti/

```text
$ graffiti .
✓ Done. 0 API calls, $0.  214 files → 1,883 nodes, 4,102 edges, 12 communities.
  The 3 most interesting questions your map can answer:
    1) Which module is the load-bearing wall?
    2) What does the auth flow touch?
    3) Where are the cross-package call hotspots?
```

---

## Зачем это нужно

ИИ-ассистент по программированию настолько хорош, насколько хорошо то, что он может *видеть*.
Поместите его в большой репозиторий — и он сделает то же, что сделали бы вы без карты:
прогоняет grep, открывает несколько файлов, гадает. Он никогда не видит **форму**
кода — какая функция какую вызывает, где определён тип, какой модуль является несущей стеной.

**graffiti — это та самая карта, которой там должно было быть.** Одна команда разбирает репозиторий
с помощью [tree-sitter](https://tree-sitter.github.io/tree-sitter/), разрешает рёбра,
группирует модули в кластеры и записывает граф — как JSON для машины, как Markdown для
вас и как единый офлайн-HTML, на который вы действительно можете взглянуть. Никаких ключей. Никакого облака. Никаких затрат.

## Установка

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

Зафиксируйте версию или каталог:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

Установщик выбирает подходящий статический бинарник для вашей ОС/архитектуры, сверяет его SHA256
с манифестом релиза и устанавливает его. Проверьте командой `graffiti version`.
Или соберите из исходного кода (ниже).

## ⚡ Установка через Claude Code (vibe-code)

<!-- vibe-install -->
Терминал не нужен — пусть всю работу сделает **Claude Code**. Вставьте этот единственный
промпт в сессию Claude Code и отвечайте `y` на каждом шаге. Он скачает нужный бинарник,
построит карту для вашего репозитория, настроит интеграцию и откроет граф:

```text
Установи мне graffiti от amazopic. Скачай подходящий статический бинарник для моей ОС/архитектуры из последнего релиза на github.com/amazopic/graffiti (или собери из исходного кода с помощью `make build`, если установлен Go), помести его в мой PATH под именем `graffiti` и проверь командой `graffiti version`. Затем запусти `graffiti .` в корне моего репозитория, чтобы построить карту, запусти `graffiti init --hook`, чтобы подключить graffiti к Claude Code, и в конце открой `.graffiti/map.html`, чтобы я увидел граф. Спрашивай перед каждым шагом.
```

## Сборка

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

Теги сборки `grammar_subset` включают только те грамматики, которые поддерживает graffiti (Go,
Python, JS, TS, Rust, Java, PHP, плюс go.mod), через чисто-Go-рантайм
`github.com/odvcencio/gotreesitter` (без CGO, без WASM). Они удерживают бинарник на
~10 МБ; без них код всё равно компилируется, но компонуется с полным набором грамматик
(~31 МБ). Всегда указывайте их — Makefile делает это за вас.

## Поддерживаемые языки

| Язык | Что извлекается |
|----------|-----------|
| Go | файлы, функции, методы (по получателю), типы, импорты, разрешённые вызовы |
| Python, JavaScript, TypeScript, Rust, Java, PHP | файлы, функции, классы/структуры/интерфейсы/перечисления/трейты, методы (`Class.method`), импорты, вызовы внутри репозитория |
| Markdown | узлы документации |

Извлечение для не-Go языков намеренно честное: оно захватывает распространённую, высокоценную
структуру и **недоизвлекает** экзотические конструкции (декораторы, дженерики, вложенные
определения, динамическую диспетчеризацию), а не выдаёт догадки.

## Использование

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

Запустите `graffiti` без аргументов, чтобы получить полный список команд.

## Одна команда, три артефакта

`graffiti .` записывает всё в `<repo>/.graffiti/`:

- **`map.json`** — сам граф: узлы, рёбра, сообщества, проверенные по схеме
  `schema/map.schema.json`. Именно это читает ваш ИИ и именно это обходят `query`
  и сервер MCP.
- **`MAP.md`** — удобная для человека сводка: главные модули, наиболее связанные узлы
  и три самых интересных вопроса, на которые может ответить ваша карта.
- **`map.html`** — единый самодостаточный, офлайн, интерактивный **граф с силовой
  раскладкой**. Никакого CDN, никакого сервера, никакой сети — просто откройте файл.

`map.html` имеет **переключатель 2D/3D** (наведение приподнимает узел и его соседей), **поиск
узлов**, **копирование `file:line` по клику**, **секторные зоны**, переключатели категорий **client /
tests / external** и изменяемое по размеру дерево **project → directory → file**
с флажками показа/скрытия. Он безопасен с точки зрения CSP и работает полностью офлайн.

Кэш на основе хеша содержимого по каждому файлу хранится в `<repo>/.graffiti/cache/`, так что повторные
запуски заново разбирают только то, что изменилось.

## Интеграция с Claude Code

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` записывает:

- `.claude/skills/graffiti/SKILL.md` — короткий навык, чтобы Claude Code знал, как собирать/читать/запрашивать карту.
- блок `CLAUDE.md` (между `<!-- graffiti:start -->` / `<!-- graffiti:end -->`), который сообщает
  ассистенту предпочитать `graffiti query` вместо grep, когда карта существует.
- с `--hook` — запись PreToolUse в `.claude/settings.json`, запускающую `graffiti hook`, которая добавляет
  однострочную подсказку перед `Grep`/`Glob`, когда присутствует `.graffiti/map.json`. Хук никогда не блокирует инструмент.

Команда идемпотентна — запускайте повторно в любое время; существующее содержимое `CLAUDE.md` / `settings.json` сохраняется.

## Запрос без LLM

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` возвращает релевантный срез графа в рамках мягкого бюджета ~2000 токенов на узлы —
без модели, без эмбеддингов. Заключайте вопрос в кавычки.

## Сервер MCP

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

Направьте на него любой MCP-совместимый клиент, и ваш ассистент будет обходить граф через
инструменты вместо grep.

## Рабочие пространства (федерация нескольких репозиториев)

Расположите отдельные репозитории рядом и запрашивайте по ним — **без слияния**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` записывает фиксируемый в системе контроля версий реестр (`.graffiti-workspace/workspace.json`)
и производный, исключаемый из git кэш (`.graffiti-workspace/overlay.json`). Собственный
`.graffiti/map.json` каждого репозитория остаётся неизменным и по-прежнему работает автономно — рабочее пространство представляет собой
тонкий вычисляемый оверлей, а не объединённый монолит.

**Межпроектные связи:** объявляйте их явно в `.graffiti-workspace/links`,
по одной на строку — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(допускаются комментарии `#`; конечные точки имеют вид `alias::nodeid`). `graffiti links check` проверяет,
что обе конечные точки разрешаются; `graffiti federate --explain` перечисляет каждую связь. Федеративный запрос
снабжает каждый узел префиксом псевдонима его участника и обходит межпроектные связи. `graffiti workspace
render` записывает `workspace.html` — тот же просмотрщик графа с силовой раскладкой, где **проекты являются
верхним уровнем** дерева, и с отрисованными межпроектными связями.

Добавьте `.graffiti-workspace/overlay.json` в `.gitignore` (он производный и пересчитываемый).

## 🛰️ Оркестрация системы — много сервисов, один граф

<!-- system-orchestration -->
Микросервисная система — это множество независимых репозиториев, которые образуют один
продукт. graffiti строит карту каждого из них, а затем **обнаруживает рёбра между ними** —
HTTP, gRPC, очереди — исходя из *контрактной поверхности* каждого сервиса (того, что он
предоставляет, `provides`, и потребляет, `consumes`). Никакого ручного связывания: каждый
сервис публикует собственную карту; оркестратор федерирует опубликованные артефакты и
сопоставляет потребителей с поставщиками.

```bash
# in each service's CI (or locally) — publish its map into a shared store:
graffiti publish --to ../system-store --as carts

# then, in CI or on demand, over the whole system:
graffiti system build       # federate + auto-discover cross-service links
graffiti system render      # → .graffiti-system/system.html (services as lanes)
graffiti system impact carts::"GET /carts/{}"   # who breaks if this changes?
graffiti system audit       # dangling consumers · orphan providers · ambiguous (CI gate)
graffiti system query "where is the cart fetched and served"
```

Каждая карта несёт **контрактную поверхность**, извлечённую из `openapi.json`, `.proto`,
маршрутов фреймворка, вызовов очередей или явного `graffiti.contract.json`. Межсервисные
связи оцениваются по уверенности; **неоднозначные** и **повисшие** (с мёртвой конечной точкой)
потребители сообщаются, а не молча отбрасываются. Системное хранилище — это просто каталог
или git-репозиторий — $0, офлайн, пересчитываемое.

## Как это работает

разбор tree-sitter (чисто Go, без CGO) → разрешение рёбер → кластеризация в
сообщества → лёгкий анализ → детерминированная сериализация. Никакой модели, никаких
эмбеддингов, никакой сети — только статический анализ. Вот почему это бесплатно, приватно и
воспроизводимо.

## Гарантии

- **0 API-вызовов, $0, полностью офлайн.** Ничто из вашего кода не покидает вашу машину.
- **Детерминированность:** один и тот же репозиторий → побайтово идентичный `map.json` с точностью до единственной
  метки времени `generated_at` и базового имени `root`. Фиксируйте его в репозитории; смотрите его диффы.
- **Единый статический бинарник**, без зависимостей времени выполнения, без C-инструментария.

## Лицензия

Source-Available — читайте и запускайте graffiti свободно на своих собственных репозиториях, но любое
повторное использование, распространение, форк или включение в другой проект требует предварительного письменного
разрешения автора. См. [LICENSE](LICENSE).

## Автор

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
