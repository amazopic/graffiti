# 🕸️ graffiti — 將任何儲存庫轉換成 AI 可查詢的程式碼圖譜

> 一道指令就能將你的儲存庫轉換成一張**有向知識圖譜**，讓你的 AI
> 程式設計助手用它來閱讀，而不是盲目地 grep。一個靜態的 Go 單一執行檔 —
> **零 API 金鑰、$0、完全離線、位元組層級可決定。**可解析 **Go、Python、
> JavaScript、TypeScript、Rust、Java 與 PHP**。隨附一個免 LLM 的 `query`、一個
> **MCP** 伺服器、**Claude Code** 整合、一個互動式離線圖譜檢視器，以及
> 多儲存庫工作區聯邦。

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**語言：** English · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

🌐 **網站：** https://amazopic.github.io/graffiti/

```text
$ graffiti .
✓ Done. 0 API calls, $0.  214 files → 1,883 nodes, 4,102 edges, 12 communities.
  The 3 most interesting questions your map can answer:
    1) Which module is the load-bearing wall?
    2) What does the auth flow touch?
    3) Where are the cross-package call hotspots?
```

---

## 為什麼需要它

一個 AI 程式設計助手的能力上限，取決於它能*看見*多少。把它丟進一個龐大的
儲存庫，它只能做你在沒有地圖時會做的事：grep、開幾個檔案、然後用猜的。
它永遠看不見程式碼的**整體形狀** — 哪個函式呼叫哪個、某個型別定義在哪裡、
哪個模組是承重牆。

**graffiti 就是那張本應存在的地圖。**一道指令就以
[tree-sitter](https://tree-sitter.github.io/tree-sitter/) 解析整個儲存庫，
解析出邊、把模組分群，並寫出一張圖 — 給機器看的 JSON、給你看的 Markdown，
還有一個你真的可以親眼觀看的單一離線 HTML。沒有金鑰。沒有雲端。沒有費用。

## 安裝

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

指定版本或目錄：

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

安裝程式會為你的 OS/架構挑選正確的靜態執行檔，對照發行版本資訊檔驗證其
SHA256，然後完成安裝。可用 `graffiti version` 驗證。
或從原始碼建置（見下文）。

## ⚡ 用 Claude Code 安裝（vibe-code）

<!-- vibe-install -->
不需要終端機 — 讓 **Claude Code** 幫你一手包辦。把下面這段提示詞貼進一個
Claude Code 工作階段，並在每個步驟回答 `y`。它會抓取正確的執行檔、為你的
儲存庫建置地圖、串接好整合，然後打開圖譜：

```text
幫我安裝 amazopic 的 graffiti。從 github.com/amazopic/graffiti 的最新發行版本下載適合我 OS/架構的靜態執行檔（若已安裝 Go，也可以用 `make build` 從原始碼建置），把它以 `graffiti` 之名放到我的 PATH 上，並用 `graffiti version` 驗證。接著在我的儲存庫根目錄執行 `graffiti .` 來建置地圖，執行 `graffiti init --hook` 把 graffiti 串接進 Claude Code，最後打開 `.graffiti/map.html`，讓我可以看到圖譜。每個步驟前都先問我。
```

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

`grammar_subset` build tags 只會透過純 Go 執行環境
`github.com/odvcencio/gotreesitter`（無 CGO、無 WASM）隨附 graffiti 支援的文法
（Go、Python、JS、TS、Rust、Java、PHP，外加 go.mod）。它們讓執行檔維持在
約 10 MB；少了它們，程式碼仍可編譯，但會連結完整的文法集（約 31 MB）。請
總是傳入它們 — Makefile 已替你做好這件事。

## 支援的語言

| 語言 | 擷取內容 |
|----------|-----------|
| Go | 檔案、函式、方法（依接收者）、型別、imports、已解析的呼叫 |
| Python、JavaScript、TypeScript、Rust、Java、PHP | 檔案、函式、類別／結構／介面／列舉／trait、方法（`Class.method`）、imports、儲存庫內呼叫 |
| Markdown | 文件節點 |

非 Go 的擷取刻意保持誠實：它擷取常見且高價值的結構，並對冷僻的構造
（裝飾器、泛型、巢狀定義、動態分派）**刻意少擷取**，而不是輸出臆測。

## 用法

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

不帶任何引數執行 `graffiti` 即可看到完整的指令清單。

## 一道指令，三件產物

`graffiti .` 會把所有東西寫進 `<repo>/.graffiti/`：

- **`map.json`** — 圖本身：節點、邊、社群，並對照
  `schema/map.schema.json` 進行 schema 檢查。這是你的 AI 所讀取的內容，也是
  `query` 與 MCP 伺服器所遍歷的對象。
- **`MAP.md`** — 一份人類可讀的摘要：頂層模組、連結最多的節點，
  以及你的地圖能回答的三個最有意思的問題。
- **`map.html`** — 一個單一、自成一體、離線、互動式的**力導向
  圖譜**。無 CDN、無伺服器、無網路 — 直接打開檔案即可。

`map.html` 具備 **2D/3D 切換**（懸停會抬起一個節點及其鄰居）、**節點
搜尋**、**點擊複製 `file:line`**、**區段分區**、**client／tests／
external** 類別切換，以及一棵可調整大小的 **project → directory → file** 樹，
附帶顯示／隱藏核取方塊。它符合 CSP，並可完全離線運作。

每個檔案的內容雜湊快取存放在 `<repo>/.graffiti/cache/` 之下，因此重新執行
時只會重新解析有變動的部分。

## Claude Code 整合

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` 會寫入：

- `.claude/skills/graffiti/SKILL.md` — 一個簡短的技能，讓 Claude Code 知道要建置／讀取／查詢地圖。
- 一段 `CLAUDE.md` 區塊（位於 `<!-- graffiti:start -->` / `<!-- graffiti:end -->` 之間），告訴
  助手在地圖存在時優先使用 `graffiti query` 而非 grep。
- 加上 `--hook` 後，會新增一筆執行 `graffiti hook` 的 `.claude/settings.json` PreToolUse 項目，
  它會在 `.graffiti/map.json` 存在時，於 `Grep`／`Glob` 之前加上一行提示。此 hook 絕不會阻擋工具。

它具備冪等性 — 隨時可重新執行；既有的 `CLAUDE.md` / `settings.json` 內容都會被保留。

## 不用 LLM 也能查詢

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` 會在約 2000 token 的軟性節點預算內，回傳圖中相關的一個切片 — 無
模型、無嵌入向量。問題請加上引號。

## MCP 伺服器

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

把任何具備 MCP 能力的用戶端指向它，你的助手就能透過工具遍歷圖譜，而不是用
grep。

## 工作區（多儲存庫聯邦）

把各自獨立的儲存庫並排擺放，並跨它們查詢 — **不需合併**：

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` 會寫入一份可提交的登錄檔（`.graffiti-workspace/workspace.json`）
以及一份衍生、可加入 gitignore 的快取（`.graffiti-workspace/overlay.json`）。每個儲存庫
自己的 `.graffiti/map.json` 都不會被改動，且仍可獨立運作 — 工作區只是一層薄薄的計算
覆蓋層，絕非合併後的大塊資料。

**跨專案連結：**在 `.graffiti-workspace/links` 中明確宣告它們，
每行一條 — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
（允許 `#` 註解；端點為 `alias::nodeid`）。`graffiti links check` 會驗證
兩個端點皆可解析；`graffiti federate --explain` 會列出每一條連結。聯邦查詢
會在每個節點前面加上其成員別名，並遍歷跨連結。`graffiti workspace
render` 會寫出一個 `workspace.html` — 同樣的力導向圖譜檢視器，但**以專案作為
樹的最上層**，並繪出跨專案連結。

把 `.graffiti-workspace/overlay.json` 加入 `.gitignore`（它是衍生且可重新計算的）。

## 🛰️ 系統編排 — 眾多服務，一張圖

<!-- system-orchestration -->
一套微服務系統，是由許多獨立的儲存庫共同構成一項產品。graffiti
會逐一描繪每一個，接著**探查它們彼此之間的邊** — HTTP、gRPC、佇列 —
這些都從每個服務的*契約面*（也就是它所 `provides`（提供）與
`consumes`（消費）的內容）推導而來。不需要手動接線：每個服務各自發布
自己的地圖；編排器再將已發布的產物聯邦起來，並把消費端對應到提供端。

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

每張地圖都帶有一份**契約面**，從 `openapi.json`、`.proto`、框架路由、
佇列呼叫，或一份明確的 `graffiti.contract.json` 中擷取而來。跨服務的
連結會依信心程度評分；**有歧義**與**懸空**（端點失效）的消費端都會被
回報，絕不會被悄悄丟棄。系統儲存區不過是一個目錄或一個 git 儲存庫 —
$0、離線、可重新計算。

## 運作原理

tree-sitter 解析（純 Go、無 CGO）→ 邊解析 → 分群成
社群 → 輕量分析 → 可決定的序列化。無模型、無
嵌入向量、無網路 — 純粹是靜態分析。這就是為什麼它免費、私密又
可重現。

## 保證

- **0 次 API 呼叫、$0、完全離線。**關於你程式碼的任何資訊都不會離開你的機器。
- **可決定：**相同儲存庫 → 位元組層級完全相同的 `map.json`，僅有單一的
  `generated_at` 時間戳記與 `root` 基底名稱例外。把它提交進版控；用 diff 檢視它。
- **單一靜態執行檔**，無執行階段相依、無 C 工具鏈。

## 授權

Source-Available — 你可以自由地在自己的儲存庫上閱讀並執行 graffiti，但任何
重複使用、再散布、fork，或納入另一個專案，都需要事先取得作者的書面
許可。詳見 [LICENSE](LICENSE)。

## 作者

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
