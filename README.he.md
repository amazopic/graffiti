# 🕸️ graffiti — הפכו כל מאגר לגרף קוד הניתן לתשאול עבור בינה מלאכותית

> פקודה אחת הופכת את המאגר שלכם ל**גרף ידע מכוון** שעוזר הקוד מבוסס הבינה המלאכותית
> שלכם קורא במקום לבצע grep באופן עיוור. קובץ בינארי סטטי יחיד ב-Go —
> **אפס מפתחות API, ‏$0, עובד לחלוטין במצב לא מקוון, דטרמיניסטי ברמת הבית.** מנתח **Go, Python,
> JavaScript, TypeScript, Rust, Java ו-PHP**. כולל פקודת `query` ללא LLM, שרת
> **MCP**, אינטגרציה עם **Claude Code**, מציג גרף אינטראקטיבי שעובד במצב לא מקוון
> ופדרציית סביבות עבודה מרובות מאגרים.

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**שפות:** English · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

🌐 **אתר:** https://amazopic.github.io/graffiti/

```text
$ graffiti .
✓ Done. 0 API calls, $0.  214 files → 1,883 nodes, 4,102 edges, 12 communities.
  The 3 most interesting questions your map can answer:
    1) Which module is the load-bearing wall?
    2) What does the auth flow touch?
    3) Where are the cross-package call hotspots?
```

---

## למה זה קיים

עוזר קוד מבוסס בינה מלאכותית טוב בדיוק כמו מה שהוא מסוגל *לראות*. שחררו אותו לתוך מאגר
גדול והוא יעשה את מה שהייתם עושים בלי מפה: הוא מבצע grep, פותח כמה קבצים, מנחש.
הוא לעולם לא רואה את **הצורה** של הקוד — איזו פונקציה קוראת לאיזו, היכן טיפוס
מוגדר, איזה מודול הוא הקיר הנושא.

**graffiti היא המפה שהייתה צריכה להיות שם.** פקודה אחת מנתחת את המאגר
באמצעות [tree-sitter](https://tree-sitter.github.io/tree-sitter/), פותרת את הקשתות,
מקבצת את המודולים וכותבת גרף — כ-JSON עבור המכונה, כ-Markdown עבורכם,
וכקובץ HTML יחיד שעובד במצב לא מקוון שאפשר באמת להסתכל עליו. ללא מפתחות. ללא ענן. ללא עלות.

## התקנה

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

קיבעו גרסה או ספרייה:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

מתקין ההתקנה בוחר את הקובץ הבינארי הסטטי המתאים למערכת ההפעלה/הארכיטקטורה שלכם, מאמת את ה-SHA256 שלו
מול מניפסט הגרסה ומתקין אותו. אמתו באמצעות `graffiti version`.
או בנו מהמקור (להלן).

## ⚡ התקנה באמצעות Claude Code (vibe-code)

<!-- vibe-install -->
לא צריך טרמינל — תנו ל**Claude Code** לעשות את כל העבודה. הדביקו את ההנחיה הזו
בסשן של Claude Code וענו `y` בכל שלב. היא מורידה את הקובץ הבינארי המתאים,
בונה את המפה עבור המאגר שלכם, מחברת את האינטגרציה ופותחת את הגרף:

```text
התקן לי את graffiti מאת amazopic. הורד את הקובץ הבינארי הסטטי המתאים למערכת ההפעלה/הארכיטקטורה שלי מהגרסה האחרונה ב-github.com/amazopic/graffiti (או בנה אותו מהמקור באמצעות `make build` אם Go זמין), הצב אותו ב-PATH שלי בשם `graffiti`, ואמת באמצעות `graffiti version`. לאחר מכן הרץ `graffiti .` בשורש המאגר שלי כדי לבנות את המפה, הרץ `graffiti init --hook` כדי לחבר את graffiti אל Claude Code, ולבסוף פתח את `.graffiti/map.html` כדי שאוכל לראות את הגרף. שאל לפני כל שלב.
```

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

תגי ה-build בשם `grammar_subset` כוללים רק את הדקדוקים ש-graffiti תומכת בהם (Go,
Python, JS, TS, Rust, Java, PHP, וכן go.mod) באמצעות זמן הריצה הטהור ב-Go
‏`github.com/odvcencio/gotreesitter` (ללא CGO, ללא WASM). הם שומרים על הקובץ הבינארי בגודל
‏~10 MB; בלעדיהם הקוד עדיין מתהדר אך מקשר את מערכת הדקדוקים המלאה
‏(~31 MB). תמיד העבירו אותם — ה-Makefile עושה זאת עבורכם.

## שפות נתמכות

| שפה | מה מחולץ |
|----------|-----------|
| Go | קבצים, פונקציות, מתודות (לפי receiver), טיפוסים, imports, קריאות שנפתרו |
| Python, JavaScript, TypeScript, Rust, Java, PHP | קבצים, פונקציות, classes/structs/interfaces/enums/traits, מתודות (`Class.method`), imports, קריאות בתוך המאגר |
| Markdown | צמתי תיעוד |

החילוץ של שפות שאינן Go הוא כן במכוון: הוא לוכד את המבנה הנפוץ ובעל הערך הגבוה,
ו**מחלץ בחֶסר** מבנים אקזוטיים (decorators, generics, הגדרות מקוננות, dynamic dispatch)
במקום לפלוט ניחושים.

## שימוש

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

הריצו את `graffiti` ללא ארגומנטים לקבלת רשימת הפקודות המלאה.

## פקודה אחת, שלושה תוצרים

‏`graffiti .` כותבת הכול לתוך `<repo>/.graffiti/`:

- **`map.json`** — הגרף עצמו: צמתים, קשתות, קהילות, מאומת מול הסכמה
  ‏`schema/map.schema.json`. זה מה שהבינה המלאכותית שלכם קוראת ומה ש-`query`
  ושרת ה-MCP חוצים.
- **`MAP.md`** — תקציר קריא לבני אדם: המודולים המובילים, הצמתים המקושרים ביותר,
  ושלוש השאלות המעניינות ביותר שהמפה שלכם יכולה לענות עליהן.
- **`map.html`** — **גרף בעל כוחות מכוונים** (force-directed) יחיד, עצמאי, אינטראקטיבי, שעובד
  במצב לא מקוון. ללא CDN, ללא שרת, ללא רשת — פשוט פתחו את הקובץ.

ל-`map.html` יש **מתג 2D/3D** (ריחוף מרים צומת ואת שכניו), **חיפוש
צמתים**, **לחיצה להעתקת `file:line`**, **אזורי מגזרים**, מתגי קטגוריות עבור **client / tests /
external**, ועץ **פרויקט ← ספרייה ← קובץ** הניתן לשינוי גודל
עם תיבות סימון להצגה/הסתרה. הוא בטוח מבחינת CSP ועובד לחלוטין במצב לא מקוון.

מטמון לפי תוכן (content-hash) ברמת הקובץ נמצא תחת `<repo>/.graffiti/cache/`, כך שהרצות חוזרות
מנתחות מחדש רק את מה שהשתנה.

## אינטגרציה עם Claude Code

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

‏`graffiti init` כותבת:

- ‏`.claude/skills/graffiti/SKILL.md` — מיומנות קצרה כדי ש-Claude Code יידע לבנות/לקרוא/לתשאל את המפה.
- בלוק `CLAUDE.md` (בין `<!-- graffiti:start -->` ל-`<!-- graffiti:end -->`) שמורה
  לעוזר להעדיף את `graffiti query` על פני grep כשקיימת מפה.
- עם `--hook`, רשומת PreToolUse בקובץ `.claude/settings.json` שמריצה את `graffiti hook`, אשר מוסיפה
  רמז של שורה אחת לפני `Grep`/`Glob` כאשר `.graffiti/map.json` קיים. ה-hook לעולם אינו חוסם כלי.

זה אידמפוטנטי — הריצו שוב בכל עת; התוכן הקיים של `CLAUDE.md` / `settings.json` נשמר.

## תשאול ללא LLM

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

‏`query` מחזיר פלח רלוונטי של הגרף בתוך תקציב צמתים רך של ~2000 טוקנים
— ללא מודל, ללא embeddings. הקיפו את השאלה במרכאות.

## שרת MCP

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

כוונו אליו כל לקוח התומך ב-MCP והעוזר שלכם יחצה את הגרף דרך
כלים במקום לבצע grep.

## סביבות עבודה (פדרציה מרובת מאגרים)

הניחו מאגרים נפרדים זה לצד זה ותשאלו לרוחבם — **בלי מיזוג**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

‏`graffiti link` כותבת רישום הניתן לקומיט (`.graffiti-workspace/workspace.json`)
ומטמון נגזר שניתן להחריג מ-git‏ (`.graffiti-workspace/overlay.json`). ה-`.graffiti/map.json`
של כל מאגר נותר ללא שינוי ועדיין עובד באופן עצמאי — סביבת העבודה היא
שכבת-על מחושבת ודקה, ולעולם לא גוש ממוזג.

**קישורים בין פרויקטים:** הצהירו עליהם במפורש ב-`.graffiti-workspace/links`,
אחד בכל שורה — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
‏(הערות `#` מותרות; נקודות הקצה הן `alias::nodeid`). `graffiti links check` מאמת
ששתי נקודות הקצה נפתרות; `graffiti federate --explain` מציג כל קישור. תשאול מפודרר
מקדים לכל צומת את כינוי החבר שלו וחוצה קישורים צולבים. `graffiti workspace
render` כותבת `workspace.html` — אותו מציג גרף-כוחות, עם ה**פרויקטים כרמה
העליונה** של העץ ועם הקישורים בין הפרויקטים מצוירים.

הוסיפו את `.graffiti-workspace/overlay.json` ל-`.gitignore` (הוא נגזר וניתן לחישוב מחדש).

## 🛰️ תזמור מערכות — שירותים רבים, גרף אחד

<!-- system-orchestration -->
מערכת מבוססת מיקרו-שירותים היא מאגרים רבים ועצמאיים שיוצרים יחד מוצר אחד. graffiti
ממפה כל אחד מהם, ולאחר מכן **מגלה את הקשתות שביניהם** — HTTP, gRPC, תורים — מתוך
*משטח החוזה* (contract surface) של כל שירות (מה שהוא מספק `provides` ומה שהוא צורך
`consumes`). ללא חיווט ידני: כל שירות מפרסם את המפה שלו; המתזמר מאחד את התוצרים
שפורסמו ומתאים בין צרכנים לספקים.

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

כל מפה נושאת **משטח חוזה** (contract surface) המחולץ מתוך `openapi.json`, ‏`.proto`,
נתיבי framework, קריאות לתורים, או קובץ `graffiti.contract.json` מפורש. הקישורים
בין השירותים מדורגים לפי רמת ביטחון; צרכנים **דו-משמעיים** (ambiguous) ו**תלושים**
(dangling — נקודת קצה מתה) מדווחים, ולעולם אינם מושמטים בשקט. מאגר המערכת הוא פשוט
ספרייה או מאגר git — ‏$0, לא מקוון, ניתן לחישוב מחדש.

## איך זה עובד

ניתוח tree-sitter (טהור ב-Go, ללא CGO) ← פתרון קשתות ← קיבוץ
לקהילות ← ניתוח קל-משקל ← סריאליזציה דטרמיניסטית. ללא מודל, ללא
embeddings, ללא רשת — רק ניתוח סטטי. זו הסיבה שזה חינמי, פרטי
וניתן לשחזור.

## הבטחות

- **0 קריאות API, ‏$0, עובד לחלוטין במצב לא מקוון.** דבר בנוגע לקוד שלכם אינו עוזב את המכונה שלכם.
- **דטרמיניסטי:** אותו מאגר ← `map.json` זהה ברמת הבית, פרט לחותמת הזמן היחידה
  ‏`generated_at` ולשם הבסיס של ה-`root`. עשו לו קומיט; השוו אותו ב-diff.
- **קובץ בינארי סטטי יחיד**, ללא תלויות זמן ריצה, ללא ערכת כלים של C.

## רישיון

Source-Available — קראו והריצו את graffiti בחופשיות על המאגרים שלכם, אך כל
שימוש חוזר, הפצה מחדש, fork, או הכללה בפרויקט אחר מחייבים אישור בכתב מראש
מהמחבר. ראו [LICENSE](LICENSE).

## מחבר

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
