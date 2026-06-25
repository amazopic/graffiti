# 🕸️ graffiti — کسی بھی repo کو AI کے لیے قابلِ کوئری کوڈ گراف میں بدلیں

> ایک کمانڈ آپ کی repository کو ایک **directed knowledge graph** میں بدل دیتی ہے
> جسے آپ کا AI کوڈنگ اسسٹنٹ اندھا دھند grep کرنے کے بجائے پڑھتا ہے۔ ایک واحد سٹیٹک Go binary —
> **صفر API keys، $0، مکمل طور پر آف لائن، byte-deterministic۔** **Go، Python،
> JavaScript، TypeScript، Rust، Java، اور PHP** کو پارس کرتا ہے۔ ایک LLM سے آزاد `query`، ایک
> **MCP** سرور، **Claude Code** انٹیگریشن، ایک انٹرایکٹو آف لائن گراف
> ویور، اور کثیر-repo workspace فیڈریشن فراہم کرتا ہے۔

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**زبانیں:** English · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

🌐 **ویب سائٹ:** https://amazopic.github.io/graffiti/

```text
$ graffiti .
✓ Done. 0 API calls, $0.  214 files → 1,883 nodes, 4,102 edges, 12 communities.
  The 3 most interesting questions your map can answer:
    1) Which module is the load-bearing wall?
    2) What does the auth flow touch?
    3) Where are the cross-package call hotspots?
```

---

## یہ کیوں موجود ہے

ایک AI کوڈنگ اسسٹنٹ صرف اتنا ہی اچھا ہوتا ہے جتنا وہ *دیکھ* سکتا ہے۔ اسے کسی بڑی
repo میں ڈالیں تو وہ وہی کرتا ہے جو آپ بغیر نقشے کے کرتے: یہ grep کرتا ہے، چند فائلیں کھولتا ہے، اندازے لگاتا ہے۔
یہ کبھی کوڈ کی **شکل** نہیں دیکھتا — کون سا فنکشن کس کو پکارتا ہے، کہاں کوئی type
متعین ہوا ہے، کون سا ماڈیول وہ دیوار ہے جس پر سارا بوجھ ٹکا ہے۔

**graffiti وہ نقشہ ہے جو یہاں ہونا چاہیے تھا۔** ایک کمانڈ repo کو
[tree-sitter](https://tree-sitter.github.io/tree-sitter/) کے ساتھ پارس کرتی ہے، edges کو حل کرتی ہے،
ماڈیولز کو کلسٹر کرتی ہے، اور ایک گراف لکھتی ہے — مشین کے لیے JSON کی صورت میں، آپ کے لیے Markdown کی صورت میں،
اور ایک واحد آف لائن HTML کی صورت میں جسے آپ واقعی دیکھ سکتے ہیں۔ کوئی keys نہیں۔ کوئی cloud نہیں۔ کوئی لاگت نہیں۔

## انسٹال کریں

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

کوئی ورژن یا ڈائریکٹری مقرر کریں:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

انسٹالر آپ کے OS/arch کے لیے درست سٹیٹک binary منتخب کرتا ہے، اس کے SHA256 کی
ریلیز manifest کے ساتھ تصدیق کرتا ہے، اور اسے انسٹال کر دیتا ہے۔ `graffiti version` سے تصدیق کریں۔
یا سورس سے بنائیں (نیچے)۔

## ⚡ Claude Code کے ساتھ انسٹال کریں (vibe-code)

<!-- vibe-install -->
کسی terminal کی ضرورت نہیں — سارا کام **Claude Code** کو کرنے دیں۔ یہ ایک prompt
کسی Claude Code سیشن میں paste کریں اور ہر مرحلے پر `y` کا جواب دیں۔ یہ آپ کے OS/arch کے لیے درست binary حاصل کرتا ہے،
آپ کی repo کا نقشہ بناتا ہے، انٹیگریشن جوڑتا ہے، اور گراف کھول دیتا ہے:

```text
میرے لیے amazopic کا graffiti انسٹال کریں۔ github.com/amazopic/graffiti پر تازہ ترین ریلیز سے میرے OS/arch کے لیے درست سٹیٹک binary ڈاؤن لوڈ کریں (یا اگر Go موجود ہو تو `make build` سے سورس سے بنائیں)، اسے میرے PATH پر `graffiti` کے نام سے رکھیں، اور `graffiti version` سے تصدیق کریں۔ پھر میری repo کی جڑ میں نقشہ بنانے کے لیے `graffiti .` چلائیں، graffiti کو Claude Code میں جوڑنے کے لیے `graffiti init --hook` چلائیں، اور آخر میں `.graffiti/map.html` کھولیں تاکہ میں گراف دیکھ سکوں۔ ہر مرحلے سے پہلے پوچھیں۔
```

<!-- quickstart -->
## فوری آغاز (60 سیکنڈ)

```bash
# 1 — install (or build from source with `make build`)
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh

# 2 — map your repo (writes .graffiti/map.json, MAP.md, map.html)
cd your-repo
graffiti .

# 3 — look at the graph
open .graffiti/map.html        # macOS — use `xdg-open` on Linux, `start` on Windows

# 4 — ask it questions: no LLM, no API key
graffiti query "where is the user authenticated"
```

پھر اسے ایک بار اپنے AI اسسٹنٹ میں جوڑ دیں:

```bash
graffiti init --hook    # Claude Code: skill + CLAUDE.md + grep→query nudge
graffiti serve          # or expose the map to any MCP client over stdio
```

**مزید مثالی سوالات** — `query` ایک نرم ~2,000-token بجٹ کے اندر ایک محدود subgraph
واپس کرتی ہے، تاکہ context چھوٹا اور سستا رہے (سوال کو quotes میں رکھیں):

```bash
graffiti query "login handler"
graffiti query "what does the checkout flow touch"
graffiti query "where is the cart fetched" ../shop   # target another path
```
<!-- /quickstart -->

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

`grammar_subset` build tags صرف وہی grammars فراہم کرتے ہیں جنہیں graffiti سپورٹ کرتا ہے (Go،
Python، JS، TS، Rust، Java، PHP، اور go.mod) خالص-Go رن ٹائم
`github.com/odvcencio/gotreesitter` کے ذریعے (نہ CGO، نہ WASM)۔ یہ binary کو
~10 MB پر رکھتے ہیں؛ ان کے بغیر کوڈ پھر بھی کمپائل ہوتا ہے مگر پورا grammar سیٹ لنک کر دیتا ہے
(~31 MB)۔ انہیں ہمیشہ پاس کریں — Makefile یہ کام آپ کے لیے کرتی ہے۔

## سپورٹ یافتہ زبانیں

| زبان | جو نکالا جاتا ہے |
|----------|-----------|
| Go | files, functions, methods (by receiver), types, imports, resolved calls |
| Python، JavaScript، TypeScript، Rust، Java، PHP | files, functions, classes/structs/interfaces/enums/traits, methods (`Class.method`), imports, intra-repo calls |
| Markdown | doc nodes |

غیر-Go استخراج جان بوجھ کر کھرا ہے: یہ عام، اعلیٰ قدر والی ساخت کو پکڑتا ہے
اور اندازے پیش کرنے کے بجائے انوکھے constructs (decorators، generics، nested
definitions، dynamic dispatch) کو **کم نکالتا ہے**۔

## استعمال

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

مکمل کمانڈ فہرست کے لیے `graffiti` کو بغیر کسی argument کے چلائیں۔

## ایک کمانڈ، تین آرٹیفیکٹس

`graffiti .` سب کچھ `<repo>/.graffiti/` میں لکھتی ہے:

- **`map.json`** — گراف بذاتِ خود: nodes، edges، communities، جس کی schema
  `schema/map.schema.json` کے خلاف جانچی جاتی ہے۔ یہی وہ ہے جسے آپ کا AI پڑھتا ہے اور جسے `query`
  اور MCP سرور ٹریورس کرتے ہیں۔
- **`MAP.md`** — ایک انسانی-پڑھنے کے قابل خلاصہ: سرفہرست ماڈیولز، سب سے زیادہ جڑے ہوئے nodes،
  اور وہ تین سب سے دلچسپ سوالات جن کا جواب آپ کا نقشہ دے سکتا ہے۔
- **`map.html`** — ایک واحد خود مکتفی، آف لائن، انٹرایکٹو **force-directed
  graph**۔ کوئی CDN نہیں، کوئی سرور نہیں، کوئی نیٹ ورک نہیں — بس فائل کھولیں۔

`map.html` میں ایک **2D/3D toggle** ہے (hover کرنے پر کوئی node اور اس کے پڑوسی اٹھ جاتے ہیں)، **node
search**، **click-to-copy `file:line`**، **sector zones**، **client / tests /
external** کیٹیگری toggles، اور ایک قابلِ سائز تبدیلی **project → directory → file** ٹری
جس میں show/hide checkboxes ہیں۔ یہ CSP-safe ہے اور مکمل طور پر آف لائن کام کرتی ہے۔

ایک per-file content-hash cache `<repo>/.graffiti/cache/` کے نیچے رہتا ہے، تاکہ دوبارہ چلانے پر
صرف وہی پارس ہو جو بدلا ہے۔

## Claude Code انٹیگریشن

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` لکھتی ہے:

- `.claude/skills/graffiti/SKILL.md` — ایک مختصر skill تاکہ Claude Code کو معلوم ہو کہ نقشہ build/read/query کرنا ہے۔
- ایک `CLAUDE.md` بلاک (`<!-- graffiti:start -->` / `<!-- graffiti:end -->` کے درمیان) جو
  اسسٹنٹ کو بتاتا ہے کہ جب کوئی نقشہ موجود ہو تو grep پر `graffiti query` کو ترجیح دے۔
- `--hook` کے ساتھ، ایک `.claude/settings.json` PreToolUse اندراج جو `graffiti hook` چلاتا ہے، جو
  `Grep`/`Glob` سے پہلے ایک سطری اشارہ شامل کرتا ہے جب `.graffiti/map.json` موجود ہو۔ یہ hook کبھی کسی ٹول کو نہیں روکتا۔

یہ idempotent ہے — کسی بھی وقت دوبارہ چلائیں؛ موجودہ `CLAUDE.md` / `settings.json` مواد محفوظ رہتا ہے۔

## LLM کے بغیر Query کریں

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` ایک نرم ~2000-token node بجٹ کے اندر گراف کا ایک متعلقہ ٹکڑا واپس کرتی ہے
— نہ کوئی ماڈل، نہ embeddings۔ سوال کو quotes میں رکھیں۔

## MCP سرور

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

کسی بھی MCP-قابل کلائنٹ کو اس کی طرف اشارہ کریں اور آپ کا اسسٹنٹ grep کرنے کے بجائے
ٹولز کے ذریعے گراف کو ٹریورس کرتا ہے۔

## Workspaces (کثیر-repo فیڈریشن)

الگ الگ repos کو ساتھ ساتھ رکھیں اور ان میں سے گزر کر کوئری کریں — **بغیر merge کیے**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` ایک committable رجسٹری (`.graffiti-workspace/workspace.json`)
اور ایک مشتق، gitignorable cache (`.graffiti-workspace/overlay.json`) لکھتی ہے۔ ہر repo کا
اپنا `.graffiti/map.json` غیر متبدل رہتا ہے اور اب بھی الگ تھلگ کام کرتا ہے — workspace ایک
پتلی computed overlay ہے، کبھی کوئی merge شدہ blob نہیں۔

**کراس-پروجیکٹ links:** انہیں `.graffiti-workspace/links` میں واضح طور پر بیان کریں،
ہر سطر میں ایک — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(`#` کمنٹس کی اجازت ہے؛ endpoints کی صورت `alias::nodeid` ہے)۔ `graffiti links check` تصدیق کرتی ہے
کہ دونوں endpoints حل ہوتے ہیں؛ `graffiti federate --explain` ہر link کی فہرست دیتی ہے۔ Federated query
ہر node کو اس کے member alias کے ساتھ سابقہ لگاتی ہے اور کراس-links کو ٹریورس کرتی ہے۔ `graffiti workspace
render` ایک `workspace.html` لکھتی ہے — وہی force-graph ویور جس میں **projects ٹری کی سب سے اوپری
سطح** کے طور پر ہوتے ہیں اور کراس-پروجیکٹ links کھینچے ہوتے ہیں۔

`.graffiti-workspace/overlay.json` کو `.gitignore` میں شامل کریں (یہ مشتق اور دوبارہ قابلِ حساب ہے)۔

## 🛰️ System orchestration — بہت سی سروسز، ایک گراف

<!-- system-orchestration -->
ایک microservice سسٹم بہت سی آزاد repos پر مشتمل ہوتا ہے جو مل کر ایک پروڈکٹ بناتی ہیں۔ graffiti
ہر ایک کا نقشہ بناتی ہے، پھر **ان کے درمیان edges کو دریافت کرتی ہے** — HTTP، gRPC، queues — ہر
سروس کی *contract surface* سے (وہ کیا `provides` یعنی فراہم کرتی ہے اور کیا `consumes` یعنی
استعمال کرتی ہے)۔ کوئی ہاتھ سے wiring نہیں: ہر سروس اپنا نقشہ خود شائع کرتی ہے؛ orchestrator شائع
شدہ آرٹیفیکٹس کو federate کرتا ہے اور consumers کو providers کے ساتھ ملاتا ہے۔

```bash
# ہر سروس کی CI میں (یا مقامی طور پر) — اس کا نقشہ ایک مشترکہ store میں شائع کریں:
graffiti publish --to ../system-store --as carts

# پھر، CI میں یا طلب پر، پورے سسٹم کے اوپر:
graffiti system build       # federate + auto-discover cross-service links
graffiti system render      # → .graffiti-system/system.html (services as lanes)
graffiti system impact carts::"GET /carts/{}"   # who breaks if this changes?
graffiti system audit       # dangling consumers · orphan providers · ambiguous (CI gate)
graffiti system query "where is the cart fetched and served"
```

ہر نقشہ ایک **contract surface** اپنے ساتھ رکھتا ہے جو `openapi.json`، `.proto`،
framework routes، queue calls، یا کسی واضح `graffiti.contract.json` سے نکالی جاتی ہے۔ کراس-سروس
links کو اعتماد (confidence) کے لحاظ سے سکور کیا جاتا ہے؛ **ambiguous** اور **dangling**
(dead-endpoint) consumers کی رپورٹ دی جاتی ہے، انہیں کبھی خاموشی سے نظرانداز نہیں کیا جاتا۔ system store
محض ایک ڈائریکٹری یا git repo ہے — $0، آف لائن، دوبارہ قابلِ حساب۔

<!-- system-walkthrough -->
### سروسز کا ایک فولڈر، قدم بہ قدم

فرض کریں آپ کی سروسز ایک ہی parent فولڈر میں رہتی ہیں، ہر ایک اپنی ڈائریکٹری میں:

```text
myproject/                ← parent فولڈر = مشترکہ "system store"
├── orders/               ← ایک سروس (Go)
├── web/                  ← ایک سروس (React/TS)
└── payments/             ← ایک سروس (Python)
```

**1. ہر سروس کو بنائیں اور شائع کریں** parent فولڈر کے ایک store میں (`--to .`)۔
`publish` ایک موجودہ نقشہ دوبارہ استعمال کرتی ہے، لہٰذا کوڈ تبدیلیاں سمیٹنے کے لیے پہلے build کریں:

```bash
cd myproject
for d in */; do
  d=${d%/}
  graffiti build "$d" && graffiti publish "$d" --to .
done
```

سروس کا نام پہلے سے طے شدہ طور پر اس کے فولڈر کا نام ہوتا ہے؛ `--as <name>` سے اسے بدلیں۔

> ⚠️ **دوبارہ-شائع کرنے پر:** `publish` کسی موجودہ نقشے کو دوبارہ **نہیں** بناتی۔ کوڈ
> تبدیل کرنے کے بعد، ہمیشہ پہلے `graffiti build <service>` چلائیں (اوپر والا loop یہی
> کرتا ہے) اور پھر `publish` — ورنہ آپ ایک باسی نقشہ شائع کر دیں گے۔

**2. سسٹم گراف بنائیں** — نقشوں کو federate کریں اور links خودکار طور پر دریافت کریں:

```bash
graffiti system build
# ✓ System "myproject": 3 services → 7 cross-service links (0 ambiguous, 0 dangling, 2 orphan). 0 API calls, $0.
```

**3. اسے استعمال کریں:**

```bash
graffiti system render          # → .graffiti-system/system.html (services as the top tree level)
graffiti system impact orders   # who breaks if orders changes (direct + transitive)
graffiti system audit           # dangling consumers · orphan providers · ambiguous (non-zero exit → CI gate)
graffiti system status          # which services drifted since the last build
graffiti system query "where is the order created"   # LLM-free retrieval across the whole system
graffiti system list            # registered services
```

**parent فولڈر میں کیا آتا ہے:**

```text
myproject/.graffiti-system/
├── system.json                 # the registry of services (commit this)
├── overlay.json                # discovered links (derived — safe to .gitignore)
├── system.html                 # the visual system map
└── services/<name>/map.json    # each service's published map
```

**link کی درستگی بہتر کریں۔** خودکار شناخت Go (net/http، gin/chi/echo)،
Flask، FastAPI، Django/DRF، Spring، NestJS، ASP.NET، Ktor، frontend clients
(React/Vue/Angular/Svelte)، gRPC اور Kafka/NATS کا احاطہ کرتی ہے۔ جہاں یہ کافی نہ ہو، ان میں سے
ایک کو کسی سروس کی جڑ میں ڈالیں (سب سے زیادہ اعتماد والی پہلے):

| File | جو فراہم کرتی ہے |
|------|-------|
| `graffiti.contract.json` | واضح `provides` / `consumes` — کوئی بھی stack، سب سے زیادہ اعتماد |
| `openapi.json` / `swagger.json` | HTTP routes بطور `provides` |
| `*.proto` | gRPC methods بطور `provides` |

کم سے کم `graffiti.contract.json`:

```json
{
  "provides": [{ "kind": "http", "name": "GET /orders/{id}" }],
  "consumes": [{ "kind": "rpc",  "name": "Payments.Charge" }]
}
```

**dead endpoints پر CI کو gate کریں** — `audit` غیر-صفر کے ساتھ خارج ہوتی ہے جب کوئی consumer ایسے
endpoint کی طرف اشارہ کرے جسے کوئی فراہم نہیں کرتا:

```bash
graffiti system build && graffiti system audit
```
<!-- /system-walkthrough -->

## یہ کیسے کام کرتا ہے

tree-sitter پارسنگ (خالص-Go، نہ CGO) → edge resolution → communities میں
کلسٹرنگ → ہلکی پھلکی تجزیہ کاری → deterministic serialization۔ نہ کوئی ماڈل، نہ
embeddings، نہ نیٹ ورک — بس static analysis۔ یہی وجہ ہے کہ یہ مفت، نجی، اور
دوبارہ قابلِ تخلیق ہے۔

## ضمانتیں

- **0 API calls، $0، مکمل طور پر آف لائن۔** آپ کے کوڈ کا کچھ بھی آپ کی مشین سے باہر نہیں جاتا۔
- **Deterministic:** وہی repo → byte-identical `map.json`، سوائے واحد
  `generated_at` ٹائم اسٹیمپ اور `root` basename کے۔ اسے commit کریں؛ اس کا diff لیں۔
- **واحد سٹیٹک binary**، کوئی رن ٹائم انحصار نہیں، کوئی C toolchain نہیں۔

## لائسنس

Source-Available — graffiti کو اپنی repositories پر آزادانہ پڑھیں اور چلائیں، مگر کوئی بھی
دوبارہ استعمال، دوبارہ تقسیم، fork، یا کسی اور پروجیکٹ میں شمولیت کے لیے مصنف سے پیشگی تحریری
اجازت درکار ہے۔ دیکھیں [LICENSE](LICENSE)۔

## مصنف

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
