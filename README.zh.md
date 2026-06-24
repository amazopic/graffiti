# 🕸️ graffiti — 将任意代码仓库变成可供 AI 查询的代码图谱

> 一条命令即可把你的代码仓库变成一张**有向知识图谱**，让你的 AI
> 编程助手去读取它，而不再盲目地 grep。一个静态链接的 Go 单一二进制文件 ——
> **零 API 密钥、$0 成本、完全离线、字节级确定性。** 可解析 **Go、Python、
> JavaScript、TypeScript、Rust、Java 和 PHP**。内置无需 LLM 的 `query`、一个
> **MCP** 服务器、**Claude Code** 集成、一个可交互的离线图谱查看器，以及
> 多仓库的工作区联邦。

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**语言:** English · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

🌐 **网站:** https://amazopic.github.io/graffiti/

```text
$ graffiti .
✓ Done. 0 API calls, $0.  214 files → 1,883 nodes, 4,102 edges, 12 communities.
  The 3 most interesting questions your map can answer:
    1) Which module is the load-bearing wall?
    2) What does the auth flow touch?
    3) Where are the cross-package call hotspots?
```

---

## 为什么需要它

AI 编程助手的能力，取决于它能*看到*什么。把它丢进一个庞大的代码仓库，它会做和你
在没有地图时一样的事:grep 一下、打开几个文件、靠猜。它永远看不到代码的**整体
结构** —— 哪个函数调用了哪个函数、某个类型在哪里定义、哪个模块是那堵承重墙。

**graffiti 就是那张本应存在的地图。** 一条命令就用
[tree-sitter](https://tree-sitter.github.io/tree-sitter/) 解析整个仓库、解析边的
关系、聚类各个模块，并写出一张图谱 —— 给机器看的 JSON、给你看的 Markdown，以及一个
你真的能打开来看的单文件离线 HTML。无需密钥。无需云端。零成本。

## 安装

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

固定某个版本或安装目录:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

安装程序会为你的操作系统/架构挑选合适的静态二进制文件，依据发布清单校验其 SHA256，
然后完成安装。可用 `graffiti version` 进行验证。或者从源码构建(见下文)。

## ⚡ 使用 Claude Code 安装(vibe-code)

<!-- vibe-install -->
无需终端 —— 让 **Claude Code** 替你完成全部工作。把下面这段提示词粘贴到 Claude Code
会话中，并在每一步都回答 `y` 即可。它会拉取合适的二进制文件、为你的仓库构建地图、
接好集成，并打开图谱:

```text
帮我安装 amazopic 的 graffiti。从 github.com/amazopic/graffiti 的最新发布中下载适配我操作系统/架构的静态二进制文件(如果有 Go，也可以用 `make build` 从源码构建),把它作为 `graffiti` 放到我的 PATH 上，并用 `graffiti version` 验证。然后在我的仓库根目录运行 `graffiti .` 构建地图,运行 `graffiti init --hook` 把 graffiti 接入 Claude Code,最后打开 `.graffiti/map.html` 让我看到图谱。每一步执行前都先问我。
```

## 构建

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

`grammar_subset` 构建标签只打包 graffiti 支持的语法(Go、Python、JS、TS、Rust、
Java、PHP，外加 go.mod),它们通过纯 Go 运行时
`github.com/odvcencio/gotreesitter`(无 CGO、无 WASM)实现。这些标签把二进制文件
保持在约 10 MB;不加它们代码同样能编译，但会链接进完整的语法集(约 31 MB)。请
始终传入它们 —— Makefile 已经替你做好了。

## 支持的语言

| 语言 | 提取内容 |
|----------|-----------|
| Go | 文件、函数、方法(按接收者)、类型、导入、已解析的调用 |
| Python、JavaScript、TypeScript、Rust、Java、PHP | 文件、函数、类/结构体/接口/枚举/trait、方法(`Class.method`)、导入、仓库内调用 |
| Markdown | 文档节点 |

非 Go 语言的提取刻意保持诚实:它会捕获常见且高价值的结构，并且对那些复杂构造
(装饰器、泛型、嵌套定义、动态分派)**有意少提取**，而不是凭空给出猜测。

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

不带任何参数运行 `graffiti` 即可查看完整的命令列表。

## 一条命令，三件产物

`graffiti .` 会把所有内容写入 `<repo>/.graffiti/`:

- **`map.json`** —— 图谱本身:节点、边、社区,并依据 `schema/map.schema.json`
  进行模式校验。这就是你的 AI 读取的东西，也是 `query` 与 MCP 服务器遍历的对象。
- **`MAP.md`** —— 一份人类可读的摘要:顶层模块、连接最多的节点,以及你的地图能
  回答的三个最有意思的问题。
- **`map.html`** —— 一个自包含、离线、可交互的**力导向图**。无 CDN、无服务器、
  无网络 —— 直接打开文件即可。

`map.html` 带有 **2D/3D 切换**(悬停会抬起一个节点及其相邻节点)、**节点搜索**、
**点击复制 `file:line`**、**扇区分区**、**客户端 / 测试 / 外部** 类别切换,以及一个
可调整大小的 **项目 → 目录 → 文件** 树形结构，并配有显示/隐藏复选框。它符合 CSP
要求，并且完全离线工作。

一份按文件内容哈希的缓存位于 `<repo>/.graffiti/cache/` 下，因此重新运行时只会
重新解析发生变化的部分。

## Claude Code 集成

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` 会写入:

- `.claude/skills/graffiti/SKILL.md` —— 一个简短的技能，让 Claude Code 知道要去构建/读取/查询地图。
- 一段 `CLAUDE.md` 区块(位于 `<!-- graffiti:start -->` / `<!-- graffiti:end -->` 之间),告诉
  助手在地图存在时优先使用 `graffiti query` 而非 grep。
- 加上 `--hook` 时，会添加一条 `.claude/settings.json` 的 PreToolUse 条目，运行 `graffiti hook`,
  当 `.graffiti/map.json` 存在时，它会在 `Grep`/`Glob` 之前加上一行提示。该钩子从不会阻断任何工具。

它是幂等的 —— 随时可以重新运行;现有的 `CLAUDE.md` / `settings.json` 内容会被保留。

## 无需 LLM 的查询

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` 会在约 2000 个 token 的软节点预算内返回图谱中相关的一个切片 —— 无模型、
无嵌入向量。记得给问题加上引号。

## MCP 服务器

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

把任意支持 MCP 的客户端指向它，你的助手就能通过工具来遍历图谱，而不再 grep。

## 工作区(多仓库联邦)

把多个独立的仓库并排摆放，并跨仓库查询 —— **无需合并**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` 会写出一个可提交到版本库的注册表(`.graffiti-workspace/workspace.json`),
以及一个派生的、可被 gitignore 的缓存(`.graffiti-workspace/overlay.json`)。每个仓库
自己的 `.graffiti/map.json` 保持不变，仍可独立工作 —— 工作区只是一个轻量的计算
叠加层，绝不是合并后的大块数据。

**跨项目链接:** 在 `.graffiti-workspace/links` 中显式声明它们，每行一条 ——
`frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(允许使用 `#` 注释;端点格式为 `alias::nodeid`)。`graffiti links check` 会校验
两端是否都能解析;`graffiti federate --explain` 会列出每一条链接。联邦查询会为每个
节点加上其成员别名前缀，并遍历跨链接。`graffiti workspace render` 会写出一个
`workspace.html` —— 同样的力导向图查看器，但**以项目作为树的顶层**，并绘制出
跨项目的链接。

把 `.graffiti-workspace/overlay.json` 加入 `.gitignore`(它是派生出来的，可以重新计算)。

## 🛰️ 系统编排 —— 众多服务，一张图谱

<!-- system-orchestration -->
一个微服务系统就是众多相互独立、却共同构成一个产品的仓库。graffiti 会分别为
每一个仓库绘制地图，然后从每个服务的*契约面*(它所 `provides`(提供)与
`consumes`(消费)的内容)出发，**发现它们之间的边** —— HTTP、gRPC、消息队列。
无需手工接线:每个服务发布自己的地图;编排器把已发布的产物联邦起来，并将消费者
与提供者进行匹配。

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

每张地图都带有一个**契约面**,它从 `openapi.json`、`.proto`、框架路由、队列调用，
或一个显式的 `graffiti.contract.json` 中提取而来。跨服务链接会按置信度评分;
**有歧义的(ambiguous)**与**悬空的(dangling，即死端点)**消费者都会被报告出来，
绝不会被悄无声息地丢弃。系统存储不过是一个目录或一个 git 仓库 —— $0 成本、完全
离线、可重新计算。

## 工作原理

tree-sitter 解析(纯 Go，无 CGO)→ 边解析 → 聚类成社区 → 轻量级分析 →
确定性序列化。无模型、无嵌入向量、无网络 —— 只有静态分析。这正是它免费、私密、
可复现的原因。

## 保证

- **0 次 API 调用、$0 成本、完全离线。** 你的代码不会有任何内容离开你的机器。
- **确定性:** 同一个仓库 → 字节级完全相同的 `map.json`,仅有单个 `generated_at`
  时间戳和 `root` 基础名(basename)会变。把它提交到版本库;对它做 diff。
- **单一静态二进制文件**,无运行时依赖，无需 C 工具链。

## 许可证

Source-Available —— 你可以在自己的代码仓库上自由地阅读和运行 graffiti，但任何
重用、再分发、fork，或将其纳入另一个项目，都需要事先获得作者的书面许可。参见
[LICENSE](LICENSE)。

## 作者

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
