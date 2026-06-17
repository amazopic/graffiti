# 🕸️ graffiti — যেকোনো রিপোকে AI-এর জন্য কোয়েরিযোগ্য কোড গ্রাফে পরিণত করুন

> একটি কমান্ড আপনার রিপোজিটরিকে একটি **নির্দেশিত নলেজ গ্রাফে** পরিণত করে, যা আপনার AI
> কোডিং সহকারী অন্ধভাবে grep করার বদলে পড়ে। একটি একক স্ট্যাটিক Go বাইনারি —
> **শূন্য API কী, $0, সম্পূর্ণ অফলাইন, বাইট-ডিটারমিনিস্টিক।** পার্স করে **Go, Python,
> JavaScript, TypeScript, Rust, Java, এবং PHP**। সরবরাহ করে একটি LLM-বিহীন `query`, একটি
> **MCP** সার্ভার, **Claude Code** ইন্টিগ্রেশন, একটি ইন্টার‍্যাক্টিভ অফলাইন গ্রাফ
> ভিউয়ার, এবং মাল্টি-রিপো ওয়ার্কস্পেস ফেডারেশন।

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**ভাষা:** English · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

🌐 **ওয়েবসাইট:** https://amazopic.github.io/graffiti/

```text
$ graffiti .
✓ Done. 0 API calls, $0.  214 files → 1,883 nodes, 4,102 edges, 12 communities.
  The 3 most interesting questions your map can answer:
    1) Which module is the load-bearing wall?
    2) What does the auth flow touch?
    3) Where are the cross-package call hotspots?
```

---

## এটি কেন আছে

একটি AI কোডিং সহকারী ঠিক ততটাই ভালো, যতটা সে *দেখতে* পারে। একে একটি বড়
রিপোতে ছেড়ে দিলে সে ঠিক তা-ই করে যা আপনি মানচিত্র ছাড়া করতেন: সে grep করে, কয়েকটি ফাইল
খোলে, অনুমান করে। সে কখনও কোডের **আকৃতি** দেখে না — কোন ফাংশন কোনটিকে কল করে, কোথায় একটি টাইপ
সংজ্ঞায়িত, কোন মডিউলটি ভার বহনকারী দেয়াল।

**graffiti হলো সেই মানচিত্র যা সেখানে থাকা উচিত ছিল।** একটি কমান্ড রিপোটিকে
[tree-sitter](https://tree-sitter.github.io/tree-sitter/) দিয়ে পার্স করে, এজগুলো রিজলভ করে,
মডিউলগুলো ক্লাস্টার করে, এবং একটি গ্রাফ লেখে — মেশিনের জন্য JSON হিসেবে, আপনার জন্য Markdown হিসেবে,
এবং একটি একক অফলাইন HTML হিসেবে যা আপনি সত্যিই দেখতে পারেন। কোনো কী নেই। কোনো ক্লাউড নেই। কোনো খরচ নেই।

## ইনস্টল

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

একটি ভার্সন বা ডিরেক্টরি পিন করুন:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

ইনস্টলার আপনার OS/arch-এর জন্য সঠিক স্ট্যাটিক বাইনারি বেছে নেয়, রিলিজ ম্যানিফেস্টের বিপরীতে এর
SHA256 যাচাই করে, এবং এটি ইনস্টল করে। `graffiti version` দিয়ে যাচাই করুন।
অথবা সোর্স থেকে বিল্ড করুন (নিচে)।

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

`grammar_subset` বিল্ড ট্যাগগুলো কেবল সেই গ্রামারগুলোই সরবরাহ করে যা graffiti সমর্থন করে (Go,
Python, JS, TS, Rust, Java, PHP, এবং go.mod), বিশুদ্ধ-Go রানটাইম
`github.com/odvcencio/gotreesitter`-এর মাধ্যমে (কোনো CGO নেই, কোনো WASM নেই)। এগুলো বাইনারিটিকে
~10 MB-এ রাখে; এগুলো ছাড়া কোডটি এখনও কম্পাইল হয় কিন্তু সম্পূর্ণ গ্রামার সেটের সাথে লিঙ্ক করে
(~31 MB)। সবসময় এগুলো পাস করুন — Makefile আপনার জন্য এটি করে দেয়।

## সমর্থিত ভাষা

| ভাষা | নিষ্কাশিত |
|----------|-----------|
| Go | files, functions, methods (by receiver), types, imports, resolved calls |
| Python, JavaScript, TypeScript, Rust, Java, PHP | files, functions, classes/structs/interfaces/enums/traits, methods (`Class.method`), imports, intra-repo calls |
| Markdown | doc nodes |

Go-বহির্ভূত নিষ্কাশন উদ্দেশ্যমূলকভাবে সৎ: এটি সাধারণ, উচ্চ-মূল্যের
গঠন ধরে এবং অনুমান নির্গত করার বদলে বিদেশী কনস্ট্রাক্ট (ডেকোরেটর, জেনেরিক, নেস্টেড
সংজ্ঞা, ডায়নামিক ডিসপ্যাচ) **কম-নিষ্কাশন** করে।

## ব্যবহার

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

সম্পূর্ণ কমান্ড তালিকার জন্য কোনো আর্গুমেন্ট ছাড়াই `graffiti` চালান।

## একটি কমান্ড, তিনটি আর্টিফ্যাক্ট

`graffiti .` সবকিছু `<repo>/.graffiti/`-এ লেখে:

- **`map.json`** — গ্রাফটি নিজেই: নোড, এজ, কমিউনিটি, `schema/map.schema.json`-এর
  বিপরীতে স্কিমা-যাচাইকৃত। এটিই আপনার AI পড়ে এবং `query`
  ও MCP সার্ভার ট্রাভার্স করে।
- **`MAP.md`** — একটি মানব-পাঠযোগ্য সারসংক্ষেপ: শীর্ষ মডিউল, সর্বাধিক-সংযুক্ত নোড,
  এবং আপনার মানচিত্র উত্তর দিতে পারে এমন তিনটি সবচেয়ে আকর্ষণীয় প্রশ্ন।
- **`map.html`** — একটি একক স্বয়ংসম্পূর্ণ, অফলাইন, ইন্টার‍্যাক্টিভ **ফোর্স-ডিরেক্টেড
  গ্রাফ**। কোনো CDN নেই, কোনো সার্ভার নেই, কোনো নেটওয়ার্ক নেই — শুধু ফাইলটি খুলুন।

`map.html`-এ আছে একটি **2D/3D টগল** (হোভার একটি নোড ও তার প্রতিবেশীদের তুলে ধরে), **নোড
সার্চ**, **ক্লিক-করে-কপি `file:line`**, **সেক্টর জোন**, **client / tests /
external** ক্যাটাগরি টগল, এবং শো/হাইড চেকবক্সসহ একটি রিসাইজযোগ্য **project → directory → file**
ট্রি। এটি CSP-নিরাপদ এবং সম্পূর্ণরূপে অফলাইনে কাজ করে।

প্রতি-ফাইল কন্টেন্ট-হ্যাশ ক্যাশ `<repo>/.graffiti/cache/`-এর অধীনে থাকে, ফলে পুনরায়-রান
শুধু যা পরিবর্তিত হয়েছে তা-ই পুনরায়-পার্স করে।

## Claude Code ইন্টিগ্রেশন

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` লেখে:

- `.claude/skills/graffiti/SKILL.md` — একটি ছোট স্কিল যাতে Claude Code জানে কীভাবে মানচিত্র বিল্ড/পড়া/কোয়েরি করতে হয়।
- একটি `CLAUDE.md` ব্লক (`<!-- graffiti:start -->` / `<!-- graffiti:end -->`-এর মাঝে) যা
  সহকারীকে বলে যে মানচিত্র থাকলে grep-এর চেয়ে `graffiti query` পছন্দ করতে।
- `--hook` সহ, একটি `.claude/settings.json` PreToolUse এন্ট্রি যা `graffiti hook` চালায়, যেটি
  `.graffiti/map.json` উপস্থিত থাকলে `Grep`/`Glob`-এর আগে এক-লাইনের নাজ যোগ করে। হুকটি কখনও কোনো টুল ব্লক করে না।

এটি আইডেমপোটেন্ট — যেকোনো সময় পুনরায় চালান; বিদ্যমান `CLAUDE.md` / `settings.json` কন্টেন্ট সংরক্ষিত থাকে।

## LLM ছাড়াই কোয়েরি

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` একটি সফট ~2000-টোকেন নোড বাজেটের মধ্যে গ্রাফের একটি প্রাসঙ্গিক অংশ ফেরত দেয়
— কোনো মডেল নেই, কোনো এম্বেডিং নেই। প্রশ্নটি উদ্ধৃতিতে রাখুন।

## MCP সার্ভার

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

যেকোনো MCP-সক্ষম ক্লায়েন্টকে এর দিকে নির্দেশ করুন এবং আপনার সহকারী grep করার বদলে
টুলের মাধ্যমে গ্রাফটি ট্রাভার্স করে।

## ওয়ার্কস্পেস (মাল্টি-রিপো ফেডারেশন)

আলাদা রিপোগুলোকে পাশাপাশি রাখুন এবং সেগুলোর মধ্যে কোয়েরি করুন — **মার্জ না করেই**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` একটি কমিটযোগ্য রেজিস্ট্রি (`.graffiti-workspace/workspace.json`)
এবং একটি উদ্ভূত, gitignore-যোগ্য ক্যাশ (`.graffiti-workspace/overlay.json`) লেখে। প্রতিটি রিপোর
নিজস্ব `.graffiti/map.json` অপরিবর্তিত থাকে এবং এখনও স্বতন্ত্রভাবে কাজ করে — ওয়ার্কস্পেসটি একটি
পাতলা গণনাকৃত ওভারলে, কখনও একটি মার্জ করা ব্লব নয়।

**ক্রস-প্রজেক্ট লিঙ্ক:** এগুলো `.graffiti-workspace/links`-এ স্পষ্টভাবে নির্ধারণ করুন,
প্রতি লাইনে একটি — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(`#` কমেন্ট অনুমোদিত; এন্ডপয়েন্ট হলো `alias::nodeid`)। `graffiti links check` যাচাই করে যে
উভয় এন্ডপয়েন্ট রিজলভ হয়; `graffiti federate --explain` প্রতিটি লিঙ্ক তালিকাভুক্ত করে। ফেডারেটেড কোয়েরি
প্রতিটি নোডের আগে তার সদস্য alias যুক্ত করে এবং ক্রস-লিঙ্ক ট্রাভার্স করে। `graffiti workspace
render` একটি `workspace.html` লেখে — একই ফোর্স-গ্রাফ ভিউয়ার যেখানে **প্রজেক্টগুলো ট্রির
শীর্ষ স্তরে** থাকে এবং ক্রস-প্রজেক্ট লিঙ্ক আঁকা হয়।

`.graffiti-workspace/overlay.json`-কে `.gitignore`-এ যোগ করুন (এটি উদ্ভূত এবং পুনরায়-গণনাযোগ্য)।

## এটি কীভাবে কাজ করে

tree-sitter পার্সিং (বিশুদ্ধ-Go, কোনো CGO নেই) → এজ রিজলিউশন → কমিউনিটিতে
ক্লাস্টারিং → হালকা বিশ্লেষণ → ডিটারমিনিস্টিক সিরিয়ালাইজেশন। কোনো মডেল নেই, কোনো
এম্বেডিং নেই, কোনো নেটওয়ার্ক নেই — শুধু স্ট্যাটিক বিশ্লেষণ। এই কারণেই এটি ফ্রি, ব্যক্তিগত, এবং
পুনরুৎপাদনযোগ্য।

## গ্যারান্টি

- **0 API কল, $0, সম্পূর্ণ অফলাইন।** আপনার কোড সম্পর্কে কিছুই আপনার মেশিন ছেড়ে যায় না।
- **ডিটারমিনিস্টিক:** একই রিপো → বাইট-অভিন্ন `map.json`, কেবল একক
  `generated_at` টাইমস্ট্যাম্প এবং `root` বেসনেম বাদে। এটি কমিট করুন; এটি diff করুন।
- **একক স্ট্যাটিক বাইনারি**, কোনো রানটাইম নির্ভরতা নেই, কোনো C টুলচেইন নেই।

## লাইসেন্স

Source-Available — আপনার নিজের রিপোজিটরিগুলোতে graffiti অবাধে পড়ুন এবং চালান, তবে যেকোনো
পুনঃব্যবহার, পুনর্বণ্টন, ফর্ক, বা অন্য কোনো প্রজেক্টে অন্তর্ভুক্তির জন্য লেখকের পূর্ব লিখিত
অনুমতি প্রয়োজন। দেখুন [LICENSE](LICENSE)।

## লেখক

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
