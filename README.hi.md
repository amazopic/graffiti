# 🕸️ graffiti — किसी भी repo को AI के लिए queryable code graph में बदलें

> एक ही command आपके repository को एक **directed knowledge graph** में बदल देती है,
> जिसे आपका AI coding assistant आँख मूँदकर grep करने के बजाय पढ़ता है। एक अकेली static Go binary —
> **शून्य API keys, $0, पूरी तरह offline, byte-deterministic.** **Go, Python,
> JavaScript, TypeScript, Rust, Java, और PHP** को parse करती है। एक LLM-रहित `query`,
> एक **MCP** server, **Claude Code** integration, एक interactive offline graph
> viewer, और multi-repo workspace federation साथ लाती है।

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**भाषाएँ:** English · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

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

## यह क्यों मौजूद है

एक AI coding assistant उतना ही अच्छा होता है जितना वह *देख* सकता है। उसे किसी बड़े
repo में डाल दीजिए और वह वही करेगा जो आप बिना नक्शे के करते: grep करता है, कुछ files खोलता है, अनुमान लगाता है।
वह कभी code की **आकृति** नहीं देख पाता — कौन-सा function किसको call करता है, कोई type कहाँ
परिभाषित है, कौन-सा module भार उठाने वाली दीवार है।

**graffiti वह नक्शा है जो वहाँ होना चाहिए था।** एक ही command repo को
[tree-sitter](https://tree-sitter.github.io/tree-sitter/) से parse करती है, edges को resolve करती है,
modules को cluster करती है, और एक graph लिखती है — मशीन के लिए JSON के रूप में, आपके लिए Markdown के रूप में,
और एक अकेली offline HTML के रूप में जिसे आप सचमुच देख सकते हैं। कोई keys नहीं। कोई cloud नहीं। कोई लागत नहीं।

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

किसी version या directory को pin करें:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

Installer आपके OS/arch के लिए सही static binary चुनता है, release manifest के विरुद्ध उसके SHA256
को verify करता है, और उसे install कर देता है। `graffiti version` से verify करें।
या source से build करें (नीचे देखें)।

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

`grammar_subset` build tags केवल वही grammars साथ लाते हैं जिन्हें graffiti support करता है (Go,
Python, JS, TS, Rust, Java, PHP, साथ ही go.mod) — pure-Go runtime
`github.com/odvcencio/gotreesitter` (कोई CGO नहीं, कोई WASM नहीं) के ज़रिए। ये binary को
~10 MB पर रखते हैं; इनके बिना code फिर भी compile होता है लेकिन पूरा grammar set link करता है
(~31 MB)। इन्हें हमेशा pass करें — Makefile आपके लिए यह कर देता है।

## Supported languages

| Language | निकाला गया |
|----------|-----------|
| Go | files, functions, methods (receiver के अनुसार), types, imports, resolved calls |
| Python, JavaScript, TypeScript, Rust, Java, PHP | files, functions, classes/structs/interfaces/enums/traits, methods (`Class.method`), imports, intra-repo calls |
| Markdown | doc nodes |

Non-Go extraction जानबूझकर ईमानदार है: यह सामान्य, उच्च-मूल्य वाली संरचना को पकड़ता है और
अनुमान लगाने के बजाय असाधारण constructs (decorators, generics, nested
definitions, dynamic dispatch) को **कम-निकालता** है।

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

पूरी command सूची के लिए `graffiti` को बिना किसी argument के चलाएँ।

## एक command, तीन artifacts

`graffiti .` सब कुछ `<repo>/.graffiti/` में लिखती है:

- **`map.json`** — graph स्वयं: nodes, edges, communities, जो
  `schema/map.schema.json` के विरुद्ध schema-checked हैं। यही वह है जो आपका AI पढ़ता है और जिसे `query`
  तथा MCP server traverse करते हैं।
- **`MAP.md`** — एक मानव-पठनीय सार: शीर्ष modules, सबसे अधिक जुड़े nodes,
  और तीन सबसे रोचक प्रश्न जिनका उत्तर आपका map दे सकता है।
- **`map.html`** — एक अकेली स्व-निहित, offline, interactive **force-directed
  graph**। कोई CDN नहीं, कोई server नहीं, कोई network नहीं — बस file खोल लें।

`map.html` में एक **2D/3D toggle** है (hover करने पर एक node और उसके पड़ोसी ऊपर उठते हैं), **node
search**, **click-to-copy `file:line`**, **sector zones**, **client / tests /
external** category toggles, और एक resizable **project → directory → file** tree
जिसमें show/hide checkboxes हैं। यह CSP-safe है और पूरी तरह offline काम करता है।

एक per-file content-hash cache `<repo>/.graffiti/cache/` के अंतर्गत रहता है, इसलिए दोबारा चलाने पर
केवल वही फिर से parse होता है जो बदला है।

## Claude Code integration

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` लिखता है:

- `.claude/skills/graffiti/SKILL.md` — एक छोटा skill ताकि Claude Code जान सके कि map को build/read/query करना है।
- एक `CLAUDE.md` block (`<!-- graffiti:start -->` / `<!-- graffiti:end -->` के बीच) जो
  assistant को बताता है कि map मौजूद होने पर grep के बजाय `graffiti query` को प्राथमिकता दे।
- `--hook` के साथ, एक `.claude/settings.json` PreToolUse प्रविष्टि जो `graffiti hook` चलाती है, जो
  `.graffiti/map.json` मौजूद होने पर `Grep`/`Glob` से पहले एक-पंक्ति का nudge जोड़ती है। यह hook कभी किसी tool को block नहीं करता।

यह idempotent है — किसी भी समय दोबारा चलाएँ; मौजूदा `CLAUDE.md` / `settings.json` सामग्री सुरक्षित रहती है।

## LLM के बिना query

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` एक soft ~2000-token node budget के भीतर graph का एक प्रासंगिक हिस्सा लौटाता है
— कोई model नहीं, कोई embeddings नहीं। प्रश्न को quotes में रखें।

## MCP server

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

किसी भी MCP-सक्षम client को इस पर इंगित करें और आपका assistant grep करने के बजाय tools के ज़रिए
graph को traverse करता है।

## Workspaces (multi-repo federation)

अलग-अलग repos को साथ-साथ रखें और उनमें आर-पार query करें — **बिना merge किए**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` एक committable registry (`.graffiti-workspace/workspace.json`)
और एक व्युत्पन्न, gitignorable cache (`.graffiti-workspace/overlay.json`) लिखता है। प्रत्येक repo की
अपनी `.graffiti/map.json` अपरिवर्तित रहती है और अभी भी standalone काम करती है — workspace एक
पतली computed overlay है, कभी भी merge किया हुआ blob नहीं।

**Cross-project links:** उन्हें `.graffiti-workspace/links` में स्पष्ट रूप से assert करें,
प्रति पंक्ति एक — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(`#` comments की अनुमति है; endpoints `alias::nodeid` हैं)। `graffiti links check` यह validate करता है
कि दोनों endpoints resolve होते हैं; `graffiti federate --explain` हर link को सूचीबद्ध करता है। Federated query
प्रत्येक node के आगे उसके member alias को prefix करती है और cross-links को traverse करती है। `graffiti workspace
render` एक `workspace.html` लिखता है — वही force-graph viewer जिसमें **projects tree के
शीर्ष स्तर** के रूप में होते हैं और cross-project links खींचे जाते हैं।

`.graffiti-workspace/overlay.json` को `.gitignore` में जोड़ें (यह व्युत्पन्न और पुनः-गणना योग्य है)।

## यह कैसे काम करता है

tree-sitter parsing (pure-Go, कोई CGO नहीं) → edge resolution → communities में clustering
→ हल्का-फुल्का analysis → deterministic serialization। कोई model नहीं, कोई
embeddings नहीं, कोई network नहीं — बस static analysis। यही कारण है कि यह मुफ़्त, निजी, और
reproducible है।

## Guarantees

- **0 API calls, $0, पूरी तरह offline.** आपके code के बारे में कुछ भी आपकी मशीन से बाहर नहीं जाता।
- **Deterministic:** वही repo → byte-identical `map.json`, सिवाय एकमात्र
  `generated_at` timestamp और `root` basename के। इसे commit करें; इसका diff देखें।
- **एक अकेली static binary**, कोई runtime dependencies नहीं, कोई C toolchain नहीं।

## License

Source-Available — graffiti को अपने स्वयं के repositories पर स्वतंत्र रूप से पढ़ें और चलाएँ, लेकिन किसी भी
प्रकार के पुनः-उपयोग, पुनर्वितरण, fork, या किसी अन्य project में शामिल करने के लिए लेखक से पूर्व लिखित
अनुमति आवश्यक है। देखें [LICENSE](LICENSE)।

## Author

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
