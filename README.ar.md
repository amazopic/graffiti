# 🕸️ graffiti — حوّل أي مستودع إلى رسم بياني للشيفرة قابل للاستعلام من أجل الذكاء الاصطناعي

> أمرٌ واحد يحوّل مستودعك إلى **رسم بياني معرفي موجّه** يقرؤه مساعد الترميز الذكي
> بدلاً من البحث الأعمى بـ grep. ملف Go تنفيذي ثابت واحد —
> **بلا مفاتيح API، بتكلفة 0$، يعمل دون اتصال بالكامل، وحتمي على مستوى البايت.** يحلّل **Go وPython
> وJavaScript وTypeScript وRust وJava وPHP**. يوفّر `query` خاليًا من نماذج اللغة، وخادم
> **MCP**، وتكاملًا مع **Claude Code**، وعارضًا تفاعليًا للرسم البياني يعمل دون اتصال،
> واتحاد مساحات عمل متعددة المستودعات.

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**اللغات:** English · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

🌐 **الموقع الإلكتروني:** https://amazopic.github.io/graffiti/

```text
$ graffiti .
✓ Done. 0 API calls, $0.  214 files → 1,883 nodes, 4,102 edges, 12 communities.
  The 3 most interesting questions your map can answer:
    1) Which module is the load-bearing wall?
    2) What does the auth flow touch?
    3) Where are the cross-package call hotspots?
```

---

## لماذا وُجدت هذه الأداة

لا يكون مساعد الترميز الذكي أفضل مما يستطيع *رؤيته*. أَلقِ به في مستودع كبير
فيفعل ما كنت لتفعله بلا خريطة: يبحث بـ grep، يفتح بضعة ملفات، يخمّن.
إنه لا يرى أبدًا **بنية** الشيفرة — أي دالة تستدعي أيها، أين يُعرَّف نوعٌ ما،
وأي وحدة هي الجدار الحامل.

**graffiti هي الخريطة التي كان ينبغي أن تكون موجودة.** أمرٌ واحد يحلّل المستودع
بواسطة [tree-sitter](https://tree-sitter.github.io/tree-sitter/)، يحلّ الحوافّ،
يجمّع الوحدات في عناقيد، ويكتب رسمًا بيانيًا — بصيغة JSON للآلة، وبصيغة Markdown لك،
وكملف HTML واحد يعمل دون اتصال يمكنك فعلاً النظر فيه. لا مفاتيح. لا سحابة. لا تكلفة.

## التثبيت

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

ثبّت إصدارًا أو دليلًا محدّدًا:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

يختار المثبِّت الملف التنفيذي الثابت المناسب لنظام التشغيل/المعمارية لديك، ويتحقق من
بصمة SHA256 الخاصة به مقابل بيان الإصدار، ثم يثبّته. تحقّق بـ `graffiti version`.
أو ابنِ من المصدر (أدناه).

## ⚡ التثبيت عبر Claude Code (بأسلوب vibe-code)

<!-- vibe-install -->
لا حاجة إلى الطرفية — دع **Claude Code** يتولّى كل شيء. الصِق هذا الأمر الواحد
في جلسة Claude Code وأجب بـ `y` عند كل خطوة. سيجلب الملف التنفيذي المناسب،
ويبني الخريطة لمستودعك، ويربط التكامل، ثم يفتح الرسم البياني:

```text
ثبّت لي graffiti من amazopic. نزّل الملف التنفيذي الثابت المناسب لنظام التشغيل/المعمارية لديّ من أحدث إصدار على github.com/amazopic/graffiti (أو ابنِه من المصدر بـ `make build` إن كان Go متاحًا)، وضعه في PATH باسم `graffiti`، وتحقّق منه بـ `graffiti version`. ثم شغّل `graffiti .` في جذر مستودعي لبناء الخريطة، وشغّل `graffiti init --hook` لربط graffiti بـ Claude Code، وأخيرًا افتح `.graffiti/map.html` كي أرى الرسم البياني. اسأل قبل كل خطوة.
```

<!-- quickstart -->
## بدء سريع (60 ثانية)

```bash
# 1 — التثبيت (أو البناء من المصدر بـ `make build`)
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh

# 2 — خرائط مستودعك (تكتب .graffiti/map.json وMAP.md وmap.html)
cd your-repo
graffiti .

# 3 — انظر إلى الرسم البياني
open .graffiti/map.html        # macOS — استخدم `xdg-open` على Linux، و`start` على Windows

# 4 — اطرح عليه الأسئلة: بلا نموذج لغوي، بلا مفتاح API
graffiti query "where is the user authenticated"
```

ثم اربطه بمساعدك الذكي مرة واحدة:

```bash
graffiti init --hook    # Claude Code: skill + CLAUDE.md + grep→query nudge
graffiti serve          # أو اعرض الخريطة لأي عميل MCP عبر stdio
```

**مزيد من الأسئلة كأمثلة** — يُرجِع `query` رسمًا بيانيًا فرعيًا محدَّد النطاق ضمن ميزانية
مرنة تبلغ ~2000 رمز، فيبقى السياق صغيرًا وزهيدًا (ضع السؤال بين علامتي اقتباس):

```bash
graffiti query "login handler"
graffiti query "what does the checkout flow touch"
graffiti query "where is the cart fetched" ../shop   # استهدف مسارًا آخر
```
<!-- /quickstart -->

## البناء

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

تشحن أعلام البناء `grammar_subset` القواعد النحوية التي يدعمها graffiti فقط (Go
وPython وJS وTS وRust وJava وPHP، بالإضافة إلى go.mod) عبر زمن تشغيل Go الخالص
`github.com/odvcencio/gotreesitter` (بلا CGO، بلا WASM). تبقي حجم الملف التنفيذي عند
~10 ميغابايت؛ من دونها تظل الشيفرة تُترجَم لكنها تربط مجموعة القواعد النحوية الكاملة
(~31 ميغابايت). مرّرها دائمًا — وملف Makefile يفعل ذلك نيابةً عنك.

## اللغات المدعومة

| اللغة | ما يُستخرَج |
|----------|-----------|
| Go | الملفات، الدوال، التوابع (حسب المُستقبِل)، الأنواع، الاستيرادات، الاستدعاءات المحلولة |
| Python، JavaScript، TypeScript، Rust، Java، PHP | الملفات، الدوال، الأصناف/البِنى/الواجهات/التعدادات/السمات، التوابع (`Class.method`)، الاستيرادات، الاستدعاءات داخل المستودع |
| Markdown | عُقَد التوثيق |

الاستخراج لغير Go صادق عن قصد: فهو يلتقط البنية الشائعة عالية القيمة، **ويستخرج أقل من اللازم**
للبُنى النادرة (المزخرِفات، الأنواع العامة، التعريفات المتداخلة، الإرسال الديناميكي)
بدلًا من إصدار تخمينات.

## الاستخدام

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

شغّل `graffiti` بلا وسائط للحصول على قائمة الأوامر الكاملة.

## أمرٌ واحد، ثلاثة مُخرَجات

يكتب `graffiti .` كل شيء داخل `<repo>/.graffiti/`:

- **`map.json`** — الرسم البياني نفسه: العُقَد، الحوافّ، المجتمعات، مُتحقَّق منه مقابل
  المخطط `schema/map.schema.json`. هذا ما يقرؤه ذكاؤك الاصطناعي وما يجتازه `query`
  وخادم MCP.
- **`MAP.md`** — خلاصة مقروءة للبشر: أهم الوحدات، أكثر العُقَد ترابطًا،
  والأسئلة الثلاثة الأكثر إثارة للاهتمام التي يمكن لخريطتك الإجابة عنها.
- **`map.html`** — **رسم بياني موجَّه بالقوى** واحد ومكتفٍ ذاتيًا وتفاعلي يعمل دون اتصال.
  لا CDN، لا خادم، لا شبكة — فقط افتح الملف.

يحتوي `map.html` على **مُبدِّل ثنائي/ثلاثي الأبعاد** (التمرير فوق عُقدة يرفعها مع جيرانها)،
و**بحث في العُقَد**، و**نسخ `file:line` بنقرة**، و**مناطق قطاعية**، ومبدّلات لفئات
**شيفرتك / الاختبارات / الخارجية**، وشجرة قابلة لتغيير الحجم **المشروع ← الدليل ← الملف**
مع مربعات إظهار/إخفاء. وهو آمن وفق سياسة CSP ويعمل دون اتصال بالكامل.

تعيش ذاكرة تخزين مؤقت لبصمة محتوى كل ملف تحت `<repo>/.graffiti/cache/`، لذا فإن عمليات
إعادة التشغيل تعيد تحليل ما تغيّر فقط.

## التكامل مع Claude Code

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

يكتب `graffiti init`:

- `.claude/skills/graffiti/SKILL.md` — مهارة قصيرة كي يعرف Claude Code أن يبني/يقرأ/يستعلم الخريطة.
- كتلة `CLAUDE.md` (بين `<!-- graffiti:start -->` / `<!-- graffiti:end -->`) تُخبر
  المساعد بتفضيل `graffiti query` على grep متى وُجدت خريطة.
- مع `--hook`، مُدخل PreToolUse في `.claude/settings.json` يشغّل `graffiti hook`، الذي يضيف
  تلميحًا من سطر واحد قبل `Grep`/`Glob` عند وجود `.graffiti/map.json`. لا يحجب الخطّاف أداةً أبدًا.

وهو عملية عديمة الأثر التراكمي (idempotent) — أعد تشغيله في أي وقت؛ يُحفَظ محتوى
`CLAUDE.md` / `settings.json` الموجود.

## الاستعلام دون نموذج لغوي

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

يُرجِع `query` شريحة ذات صلة من الرسم البياني ضمن ميزانية عُقَد مرنة تبلغ ~2000 رمز
(token) — بلا نموذج، بلا تضمينات. ضع السؤال بين علامتي اقتباس.

## خادم MCP

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

وجّه أي عميل قادر على MCP إليه، فيجتاز مساعدك الرسم البياني عبر الأدوات بدلًا من البحث بـ grep.

## مساحات العمل (اتحاد متعدد المستودعات)

ضع مستودعات منفصلة جنبًا إلى جنب واستعلم عبرها — **دون دمج**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

يكتب `graffiti link` سجلًّا قابلًا للإيداع (`.graffiti-workspace/workspace.json`)
وذاكرة تخزين مؤقت مشتقّة قابلة للتجاهل في git (`.graffiti-workspace/overlay.json`). يبقى
ملف `.graffiti/map.json` الخاص بكل مستودع دون تغيير ولا يزال يعمل بمفرده — فمساحة العمل
طبقة محسوبة رقيقة، وليست أبدًا كتلة مدموجة.

**الروابط بين المشاريع:** أكّدها صراحةً في `.graffiti-workspace/links`،
واحدًا في كل سطر — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(تعليقات `#` مسموح بها؛ النهايات الطرفية هي `alias::nodeid`). يتحقق `graffiti links check`
من أن كلتا النهايتين تُحلّان؛ ويسرد `graffiti federate --explain` كل رابط. يضيف الاستعلام
المُتّحِد بادئةً لكل عُقدة باسم العضو المستعار ويجتاز الروابط المتقاطعة. يكتب `graffiti workspace
render` ملف `workspace.html` — العارض ذاته للرسم البياني الموجَّه بالقوى مع **المشاريع
بوصفها المستوى الأعلى** من الشجرة، ورسم الروابط بين المشاريع.

أضف `.graffiti-workspace/overlay.json` إلى `.gitignore` (فهو مشتق وقابل لإعادة الحساب).

## 🛰️ تنسيق الأنظمة — خدمات كثيرة، رسم بياني واحد

<!-- system-orchestration -->
نظام الخدمات المصغّرة هو مستودعات مستقلّة كثيرة تشكّل منتجًا واحدًا. تُخطِّط graffiti
كلًّا منها، ثم **تكتشف الحوافّ فيما بينها** — HTTP وgRPC والطوابير — انطلاقًا من
*سطح العقد* لكل خدمة (ما الذي `provides` تُوفّره وما الذي `consumes` تستهلكه). بلا
توصيل يدوي: تنشر كل خدمة خريطتها الخاصة؛ ويتولّى المُنسِّق اتحاد المُخرَجات المنشورة
ومطابقة المستهلِكين بالمُوفِّرين.

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

تحمل كل خريطة **سطح عقد** مُستخرَجًا من `openapi.json` أو `.proto` أو مسارات إطار
العمل أو استدعاءات الطوابير أو ملف `graffiti.contract.json` صريح. تُسجَّل درجات الثقة
للروابط بين الخدمات؛ ويُبلَّغ عن المستهلِكين **الملتبِسين** و**المعلّقين** (ذوي النقاط
الطرفية الميتة)، ولا يُسقَطون أبدًا بصمت. مخزن النظام ليس سوى دليل أو مستودع git —
بتكلفة 0$، يعمل دون اتصال، وقابل لإعادة الحساب.

<!-- system-walkthrough -->
### مجلد من الخدمات، خطوة بخطوة

لنفترض أن خدماتك تعيش في مجلد أب واحد، كل واحدة في دليلها الخاص:

```text
myproject/                ← parent folder = the shared "system store"
├── orders/               ← خدمة (Go)
├── web/                  ← خدمة (React/TS)
└── payments/             ← خدمة (Python)
```

**1. ابنِ كل خدمة وانشرها** في مخزن عند المجلد الأب (`--to .`).
يعيد `publish` استخدام خريطة موجودة، لذا ابنِ أولًا لالتقاط تغييرات الشيفرة:

```bash
cd myproject
for d in */; do
  d=${d%/}
  graffiti build "$d" && graffiti publish "$d" --to .
done
```

يأخذ اسم الخدمة افتراضيًا اسم مجلدها؛ تجاوزه بـ `--as <name>`.

> ⚠️ **عند إعادة النشر:** لا يعيد `publish` بناء خريطة موجودة. بعد
> تغيير الشيفرة، شغّل دائمًا `graffiti build <service>` أولًا (كما تفعل الحلقة
> أعلاه) ثم `publish` — وإلا فإنك تنشر خريطة قديمة.

**2. ابنِ رسم النظام البياني** — اتّحِد الخرائط واكتشف الروابط تلقائيًا:

```bash
graffiti system build
# ✓ System "myproject": 3 services → 7 cross-service links (0 ambiguous, 0 dangling, 2 orphan). 0 API calls, $0.
```

**3. استخدمه:**

```bash
graffiti system render          # → .graffiti-system/system.html (services as the top tree level)
graffiti system impact orders   # من ينكسر إذا تغيّرت orders (مباشر + متعدٍّ)
graffiti system audit           # dangling consumers · orphan providers · ambiguous (non-zero exit → CI gate)
graffiti system status          # أي خدمات انحرفت منذ آخر بناء
graffiti system query "where is the order created"   # استرجاع خالٍ من نماذج اللغة عبر النظام بأكمله
graffiti system list            # الخدمات المسجَّلة
```

**ما الذي يَحُلّ في المجلد الأب:**

```text
myproject/.graffiti-system/
├── system.json                 # سجل الخدمات (أودِع هذا)
├── overlay.json                # الروابط المكتشَفة (مشتقة — آمن تجاهلها بـ .gitignore)
├── system.html                 # خريطة النظام المرئية
└── services/<name>/map.json    # خريطة كل خدمة المنشورة
```

**حسّن دقة الروابط.** يغطي الاكتشاف التلقائي Go (net/http، gin/chi/echo)،
وFlask وFastAPI وDjango/DRF وSpring وNestJS وASP.NET وKtor وعملاء الواجهة الأمامية
(React/Vue/Angular/Svelte) وgRPC وKafka/NATS. وحيث لا يكفي ذلك، أَسقِط
أحد هذه في جذر خدمة (الأعلى ثقةً أولًا):

| File | Gives |
|------|-------|
| `graffiti.contract.json` | تصريح `provides` / `consumes` — أي حزمة تقنية، أعلى ثقة |
| `openapi.json` / `swagger.json` | مسارات HTTP كـ `provides` |
| `*.proto` | توابع gRPC كـ `provides` |

أدنى `graffiti.contract.json`:

```json
{
  "provides": [{ "kind": "http", "name": "GET /orders/{id}" }],
  "consumes": [{ "kind": "rpc",  "name": "Payments.Charge" }]
}
```

**اجعل النقاط الطرفية الميتة بوابةً لـ CI** — يخرج `audit` بقيمة غير صفرية حين يشير
مستهلِك إلى نقطة طرفية لا يوفّرها أحد:

```bash
graffiti system build && graffiti system audit
```
<!-- /system-walkthrough -->

## كيف تعمل

تحليل بـ tree-sitter (Go خالص، بلا CGO) ← حلّ الحوافّ ← التجميع في
مجتمعات ← تحليل خفيف ← تسلسل حتمي. لا نموذج، لا
تضمينات، لا شبكة — فقط تحليل ثابت. لهذا فهي مجانية، خاصّة،
وقابلة لإعادة الإنتاج.

## الضمانات

- **0 استدعاء API، 0$، يعمل دون اتصال بالكامل.** لا شيء من شيفرتك يغادر جهازك.
- **حتمية:** المستودع نفسه ← `map.json` مطابق على مستوى البايت باستثناء الطابع الزمني
  المفرد `generated_at` واسم أساس المجلد `root`. أَودِعه؛ قارِن فروقاته.
- **ملف تنفيذي ثابت واحد**، بلا اعتماديات وقت تشغيل، بلا سلسلة أدوات C.

## الترخيص

متاح المصدر (Source-Available) — اقرأ وشغّل graffiti بحرية على مستودعاتك الخاصة، لكن أي
إعادة استخدام أو إعادة توزيع أو تفريع أو تضمين في مشروع آخر يتطلب إذنًا خطيًا مسبقًا
من المؤلف. انظر [LICENSE](LICENSE).

## المؤلف

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
