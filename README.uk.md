# 🕸️ graffiti — перетворіть будь-який репозиторій на запитуваний граф коду для AI

> Одна команда перетворює ваш репозиторій на **орієнтований граф знань**, який
> ваш AI-асистент із програмування читає замість сліпого grep. Єдиний статичний
> бінарник на Go — **жодних API-ключів, $0, повністю офлайн, побайтово
> детермінований.** Розбирає **Go, Python, JavaScript, TypeScript, Rust, Java та
> PHP**. Постачається з `query` без LLM, сервером **MCP**, інтеграцією з
> **Claude Code**, інтерактивним офлайн-переглядачем графа та федерацією
> мультирепозиторного робочого простору.

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**Мови:** English · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

🌐 **Вебсайт:** https://amazopic.github.io/graffiti/

```text
$ graffiti .
✓ Done. 0 API calls, $0.  214 files → 1,883 nodes, 4,102 edges, 12 communities.
  The 3 most interesting questions your map can answer:
    1) Which module is the load-bearing wall?
    2) What does the auth flow touch?
    3) Where are the cross-package call hotspots?
```

---

## Навіщо це потрібно

AI-асистент із програмування настільки хороший, наскільки добре він *бачить*.
Помістіть його у великий репозиторій — і він робить те саме, що зробили б ви без
мапи: він робить grep, відкриває кілька файлів, гадає. Він ніколи не бачить
**форму** коду — яка функція яку викликає, де визначено тип, який модуль є
несучою стіною.

**graffiti — це мапа, яка мала б там бути.** Одна команда розбирає репозиторій за
допомогою [tree-sitter](https://tree-sitter.github.io/tree-sitter/), визначає
зв'язки, кластеризує модулі та записує граф — як JSON для машини, як Markdown для
вас і як єдиний офлайн-HTML, на який ви справді можете подивитися. Жодних ключів.
Жодної хмари. Жодних витрат.

## Встановлення

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

Зафіксуйте версію або каталог:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

Інсталятор обирає правильний статичний бінарник для вашої ОС/архітектури,
перевіряє його SHA256 за маніфестом релізу та встановлює його. Перевірте за
допомогою `graffiti version`. Або зберіть із вихідного коду (нижче).

## ⚡ Встановлення за допомогою Claude Code (vibe-code)

<!-- vibe-install -->
Термінал не потрібен — дозвольте **Claude Code** зробити все за вас. Вставте цю
єдину підказку в сесію Claude Code й відповідайте `y` на кожному кроці. Вона
завантажує потрібний бінарник, будує мапу для вашого репозиторію, налаштовує
інтеграцію та відкриває граф:

```text
Встанови для мене graffiti від amazopic. Завантаж правильний статичний бінарник для моєї ОС/архітектури з останнього релізу на github.com/amazopic/graffiti (або збери його з вихідного коду за допомогою `make build`, якщо доступний Go), додай його до мого PATH як `graffiti` й перевір за допомогою `graffiti version`. Потім запусти `graffiti .` у корені мого репозиторію, щоб побудувати мапу, запусти `graffiti init --hook`, щоб підключити graffiti до Claude Code, і нарешті відкрий `.graffiti/map.html`, щоб я побачив граф. Питай перед кожним кроком.
```

<!-- quickstart -->
## Швидкий старт (60 секунд)

```bash
# 1 — встановіть (або зберіть із вихідного коду за допомогою `make build`)
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh

# 2 — змапте свій репозиторій (записує .graffiti/map.json, MAP.md, map.html)
cd your-repo
graffiti .

# 3 — подивіться на граф
open .graffiti/map.html        # macOS — на Linux використовуйте `xdg-open`, на Windows `start`

# 4 — ставте йому запитання: без LLM, без API-ключа
graffiti query "where is the user authenticated"
```

Потім один раз підключіть його до свого AI-асистента:

```bash
graffiti init --hook    # Claude Code: skill + CLAUDE.md + підказка grep→query
graffiti serve          # або надайте мапу будь-якому MCP-клієнту через stdio
```

**Більше прикладів запитань** — `query` повертає обмежений підграф у межах м'якого
бюджету ~2000 токенів, тож контекст залишається малим і дешевим (беріть запитання
в лапки):

```bash
graffiti query "login handler"
graffiti query "what does the checkout flow touch"
graffiti query "where is the cart fetched" ../shop   # цільтеся в інший шлях
```
<!-- /quickstart -->

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

Теги збірки `grammar_subset` постачають лише ті граматики, які підтримує graffiti
(Go, Python, JS, TS, Rust, Java, PHP, а також go.mod) через рантайм на чистому Go
`github.com/odvcencio/gotreesitter` (без CGO, без WASM). Вони утримують бінарник
на рівні ~10 МБ; без них код усе одно компілюється, але лінкується з повним
набором граматик (~31 МБ). Завжди передавайте їх — Makefile робить це за вас.

## Supported languages

| Мова | Що витягується |
|----------|-----------|
| Go | файли, функції, методи (за отримувачем), типи, імпорти, розв'язані виклики |
| Python, JavaScript, TypeScript, Rust, Java, PHP | файли, функції, класи/структури/інтерфейси/переліки/трейти, методи (`Class.method`), імпорти, виклики всередині репозиторію |
| Markdown | вузли документації |

Витягування для не-Go навмисно чесне: воно фіксує поширену, цінну структуру і
**недовитягує** екзотичні конструкції (декоратори, узагальнення, вкладені
визначення, динамічну диспетчеризацію), а не видає здогадки.

## Usage

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

Запустіть `graffiti` без аргументів, щоб побачити повний список команд.

## Одна команда, три артефакти

`graffiti .` записує все у `<repo>/.graffiti/`:

- **`map.json`** — сам граф: вузли, зв'язки, спільноти, перевірені за схемою
  `schema/map.schema.json`. Це те, що читає ваш AI і що обходять `query` та сервер
  MCP.
- **`MAP.md`** — зрозумілий людині дайджест: топові модулі, найбільш пов'язані
  вузли та три найцікавіші запитання, на які може відповісти ваша мапа.
- **`map.html`** — єдиний самодостатній, офлайновий, інтерактивний
  **силоспрямований граф**. Жодної CDN, жодного сервера, жодної мережі — просто
  відкрийте файл.

`map.html` має **перемикач 2D/3D** (наведення піднімає вузол та його сусідів),
**пошук вузлів**, **копіювання `file:line` за кліком**, **секторні зони**,
перемикачі категорій **client / tests / external** і дерево **проєкт → каталог →
файл** зі зміною розміру та прапорцями показу/приховування. Воно сумісне з CSP і
працює повністю офлайн.

Кеш геш-вмісту для кожного файлу розташовується у `<repo>/.graffiti/cache/`, тож
повторні запуски перерозбирають лише те, що змінилося.

## Claude Code integration

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` записує:

- `.claude/skills/graffiti/SKILL.md` — короткий навик, щоб Claude Code знав, як
  будувати/читати/запитувати мапу.
- блок `CLAUDE.md` (між `<!-- graffiti:start -->` / `<!-- graffiti:end -->`), який
  каже асистенту віддавати перевагу `graffiti query` над grep, коли мапа існує.
- з `--hook` — запис PreToolUse у `.claude/settings.json`, що запускає
  `graffiti hook`, який додає однорядкову підказку перед `Grep`/`Glob`, коли
  присутній `.graffiti/map.json`. Хук ніколи не блокує інструмент.

Вона ідемпотентна — перезапускайте будь-коли; наявний вміст `CLAUDE.md` /
`settings.json` зберігається.

## Запит без LLM

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` повертає релевантний зріз графа в межах м'якого бюджету ~2000 токенів на
вузли — без моделі, без вкладень. Беріть запитання в лапки.

## MCP server

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

Спрямуйте на нього будь-який клієнт із підтримкою MCP — і ваш асистент обходить
граф через інструменти замість grep.

## Робочі простори (мультирепозиторна федерація)

Розмістіть окремі репозиторії поряд і запитуйте крізь них — **без злиття**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` записує придатний для комітів реєстр
(`.graffiti-workspace/workspace.json`) і похідний кеш, який можна додати до
gitignore (`.graffiti-workspace/overlay.json`). Власний `.graffiti/map.json`
кожного репозиторію залишається незмінним і досі працює окремо — робочий простір є
тонким обчисленим накладенням, ніколи не злитою грудкою.

**Міжпроєктні зв'язки:** задавайте їх явно у `.graffiti-workspace/links`, по
одному на рядок — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(коментарі `#` дозволені; кінцеві точки мають вигляд `alias::nodeid`).
`graffiti links check` перевіряє, що обидві кінцеві точки розв'язуються;
`graffiti federate --explain` перелічує кожен зв'язок. Федеративний запит додає до
кожного вузла префікс із псевдонімом учасника й обходить міжзв'язки.
`graffiti workspace render` записує `workspace.html` — той самий переглядач
силографа з **проєктами як верхнім рівнем** дерева та намальованими міжпроєктними
зв'язками.

Додайте `.graffiti-workspace/overlay.json` до `.gitignore` (він похідний і
переобчислюваний).

## 🛰️ Оркестрація системи — багато сервісів, один граф

<!-- system-orchestration -->
Мікросервісна система — це безліч незалежних репозиторіїв, що утворюють один
продукт. graffiti мапить кожен із них, а потім **виявляє зв'язки між ними** —
HTTP, gRPC, черги — на основі *контрактної поверхні* кожного сервісу (того, що він
`provides`, тобто надає, і `consumes`, тобто споживає). Жодного ручного
з'єднування: кожен сервіс публікує власну мапу, а оркестратор федерує опубліковані
артефакти й зіставляє споживачів із постачальниками.

```bash
# у CI кожного сервісу (або локально) — опублікуйте його мапу до спільного сховища:
graffiti publish --to ../system-store --as carts

# потім, у CI або на вимогу, для всієї системи:
graffiti system build       # федерація + автовиявлення міжсервісних зв'язків
graffiti system render      # → .graffiti-system/system.html (сервіси як доріжки)
graffiti system impact carts::"GET /carts/{}"   # хто зламається, якщо це зміниться?
graffiti system audit       # підвислі споживачі · постачальники-сироти · неоднозначні (gate для CI)
graffiti system query "where is the cart fetched and served"
```

Кожна мапа несе **контрактну поверхню**, витягнуту з `openapi.json`, `.proto`,
маршрутів фреймворка, викликів черг або явного `graffiti.contract.json`.
Міжсервісні зв'язки оцінюються за рівнем впевненості; **неоднозначні** та
**підвислі** (з мертвими кінцевими точками) споживачі повідомляються, а не
відкидаються мовчки. Системне сховище — це просто каталог або git-репозиторій:
$0, офлайн, переобчислюване.

<!-- system-walkthrough -->
### Тека сервісів, крок за кроком

Припустімо, ваші сервіси лежать в одній батьківській теці, кожен у власному
каталозі:

```text
myproject/                ← батьківська тека = спільне «системне сховище»
├── orders/               ← сервіс (Go)
├── web/                  ← сервіс (React/TS)
└── payments/             ← сервіс (Python)
```

**1. Зберіть і опублікуйте кожен сервіс** до сховища в батьківській теці (`--to .`).
`publish` повторно використовує наявну мапу, тож спершу зробіть build, щоб підхопити
зміни в коді:

```bash
cd myproject
for d in */; do
  d=${d%/}
  graffiti build "$d" && graffiti publish "$d" --to .
done
```

Ім'я сервісу типово береться з імені його теки; перевизначте через `--as <name>`.

> ⚠️ **При повторній публікації:** `publish` **не** перебудовує наявну мапу. Після
> зміни коду завжди спершу запускайте `graffiti build <service>` (цикл вище це
> робить), а потім `publish` — інакше ви публікуєте застарілу мапу.

**2. Побудуйте системний граф** — федеруйте мапи й автоматично виявіть зв'язки:

```bash
graffiti system build
# ✓ System "myproject": 3 services → 7 cross-service links (0 ambiguous, 0 dangling, 2 orphan). 0 API calls, $0.
```

**3. Використовуйте це:**

```bash
graffiti system render          # → .graffiti-system/system.html (services as the top tree level)
graffiti system impact orders   # who breaks if orders changes (direct + transitive)
graffiti system audit           # dangling consumers · orphan providers · ambiguous (non-zero exit → CI gate)
graffiti system status          # which services drifted since the last build
graffiti system query "where is the order created"   # LLM-free retrieval across the whole system
graffiti system list            # registered services
```

**Що з'являється в батьківській теці:**

```text
myproject/.graffiti-system/
├── system.json                 # the registry of services (commit this)
├── overlay.json                # discovered links (derived — safe to .gitignore)
├── system.html                 # the visual system map
└── services/<name>/map.json    # each service's published map
```

**Підвищте точність зв'язків.** Автовиявлення охоплює Go (net/http, gin/chi/echo),
Flask, FastAPI, Django/DRF, Spring, NestJS, ASP.NET, Ktor, фронтенд-клієнти
(React/Vue/Angular/Svelte), gRPC і Kafka/NATS. Там, де цього недостатньо, додайте
один із цих файлів у корінь сервісу (спочатку найвища впевненість):

| Файл | Що дає |
|------|-------|
| `graffiti.contract.json` | явні `provides` / `consumes` — будь-який стек, найвища впевненість |
| `openapi.json` / `swagger.json` | HTTP-маршрути як `provides` |
| `*.proto` | методи gRPC як `provides` |

Мінімальний `graffiti.contract.json`:

```json
{
  "provides": [{ "kind": "http", "name": "GET /orders/{id}" }],
  "consumes": [{ "kind": "rpc",  "name": "Payments.Charge" }]
}
```

**Заблокуйте CI на мертвих кінцевих точках** — `audit` завершується з ненульовим
кодом, коли споживач указує на кінцеву точку, яку ніхто не надає:

```bash
graffiti system build && graffiti system audit
```
<!-- /system-walkthrough -->

## Як це працює

Розбір tree-sitter (чистий Go, без CGO) → розв'язання зв'язків → кластеризація у
спільноти → легкий аналіз → детермінована серіалізація. Жодної моделі, жодних
вкладень, жодної мережі — лише статичний аналіз. Ось чому це безкоштовно,
приватно та відтворювано.

## Guarantees

- **0 викликів API, $0, повністю офлайн.** Нічого про ваш код не залишає вашу
  машину.
- **Детерміновано:** той самий репозиторій → побайтово ідентичний `map.json` з
  точністю до єдиної мітки часу `generated_at` та базового імені `root`.
  Комітьте його; робіть diff.
- **Єдиний статичний бінарник**, без залежностей часу виконання, без C-тулчейну.

## License

Source-Available — читайте та запускайте graffiti вільно на власних репозиторіях,
але будь-яке повторне використання, поширення, форк або включення в інший проєкт
потребує попереднього письмового дозволу автора. Див. [LICENSE](LICENSE).

## Author

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
