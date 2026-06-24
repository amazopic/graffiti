# 🕸️ graffiti — あらゆるリポジトリをAIのためのクエリ可能なコードグラフに

> たった1つのコマンドで、リポジトリを**有向の知識グラフ**へと変換します。AIコー
> ディングアシスタントは、やみくもにgrepする代わりにこのグラフを読み込みます。単
> 一の静的なGoバイナリ — **APIキー不要、$0、完全オフライン、バイト単位で決定論
> 的。Go、Python、JavaScript、TypeScript、Rust、Java、PHP** をパースします。LLM
> 不要の `query`、**MCP** サーバー、**Claude Code** 連携、インタラクティブなオフ
> ライングラフビューア、複数リポジトリのワークスペースフェデレーションを同梱して
> います。

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**言語:** English · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

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

## なぜ存在するのか

AIコーディングアシスタントは、*見える*ものの質に応じてしか役立ちません。大規模な
リポジトリに放り込めば、地図を持たないあなたと同じことをします。grepし、いくつか
ファイルを開き、推測する。コードの**形**を見ることは決してありません。どの関数が
どれを呼び出すのか、型がどこで定義されているのか、どのモジュールが屋台骨となる壁
なのか、を。

**graffiti は、そこに本来あるべきだった地図です。** たった1つのコマンドで、
[tree-sitter](https://tree-sitter.github.io/tree-sitter/) を使ってリポジトリを
パースし、エッジを解決し、モジュールをクラスタリングして、グラフを書き出します
— 機械のためのJSONとして、あなたのためのMarkdownとして、そして実際に眺められる単
一のオフラインHTMLとして。キー不要。クラウド不要。コスト不要。

## インストール

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

バージョンやディレクトリを固定するには:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

インストーラーは、お使いのOS/アーキテクチャに合った静的バイナリを選択し、リリース
マニフェストに対してSHA256を検証したうえでインストールします。`graffiti version`
で確認してください。あるいはソースからビルドすることもできます(下記参照)。

## ⚡ Claude Code でインストール(vibe-code)

<!-- vibe-install -->
ターミナルは不要 — すべてを **Claude Code** に任せましょう。次のプロンプトを
Claude Code のセッションに貼り付けて、各ステップで `y` と答えるだけです。適切なバ
イナリを取得し、あなたのリポジトリの地図をビルドし、連携をセットアップして、グラ
フを開いてくれます:

```text
amazopic の graffiti をインストールしてください。github.com/amazopic/graffiti の最新リリースから、私の OS/アーキテクチャに合った静的バイナリをダウンロードし(Go が使えるなら `make build` でソースからビルドしてもかまいません)、`graffiti` として私の PATH に置き、`graffiti version` で確認してください。次に、私のリポジトリのルートで `graffiti .` を実行して地図をビルドし、`graffiti init --hook` を実行して graffiti を Claude Code に組み込み、最後に `.graffiti/map.html` を開いてグラフを見せてください。各ステップの前に確認してください。
```

## ビルド

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

`grammar_subset` ビルドタグは、graffiti がサポートする文法だけ(Go、Python、JS、
TS、Rust、Java、PHP、加えて go.mod)を、ピュアGoのランタイム
`github.com/odvcencio/gotreesitter`(CGOなし、WASMなし)経由で同梱します。これに
よりバイナリは ~10 MB に保たれます。これらを付けない場合でもコードはコンパイルさ
れますが、完全な文法セット(~31 MB)がリンクされます。常にこれらを渡してください
— Makefile が代わりにやってくれます。

## サポート言語

| 言語 | 抽出される情報 |
|----------|-----------|
| Go | ファイル、関数、メソッド(レシーバ別)、型、インポート、解決済みの呼び出し |
| Python、JavaScript、TypeScript、Rust、Java、PHP | ファイル、関数、クラス/構造体/インターフェース/列挙型/トレイト、メソッド(`Class.method`)、インポート、リポジトリ内の呼び出し |
| Markdown | ドキュメントノード |

Go以外の抽出は、意図的に正直に作られています。一般的で価値の高い構造を捕捉し、推
測を吐き出すのではなく、特殊な構文(デコレータ、ジェネリクス、ネストした定義、動
的ディスパッチ)については**意図的に控えめに抽出**します。

## 使い方

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

コマンドの全一覧を見るには、引数なしで `graffiti` を実行してください。

## 1つのコマンド、3つの成果物

`graffiti .` は、すべてを `<repo>/.graffiti/` に書き出します:

- **`map.json`** — グラフそのもの: ノード、エッジ、コミュニティ。
  `schema/map.schema.json` に対してスキーマ検証されます。これは、あなたのAIが読み
  込み、`query` とMCPサーバーがたどるものです。
- **`MAP.md`** — 人間が読める要約: 主要なモジュール、最も多く接続されたノード、そ
  してあなたの地図が答えられる最も興味深い3つの質問。
- **`map.html`** — 単一の自己完結型・オフライン・インタラクティブな**力学的配置グ
  ラフ**。CDN不要、サーバー不要、ネットワーク不要 — ただファイルを開くだけ。

`map.html` には、**2D/3Dの切り替え**(ホバーするとノードとその近傍が浮き上がりま
す)、**ノード検索**、**`file:line` をクリックでコピー**、**セクターゾーン**、
**クライアント / テスト / 外部** のカテゴリ切り替え、そしてサイズ変更可能な
**プロジェクト → ディレクトリ → ファイル**のツリー(表示/非表示のチェックボックス
付き)があります。CSP対応で、完全にオフラインで動作します。

ファイルごとのコンテンツハッシュキャッシュが `<repo>/.graffiti/cache/` の下に置か
れるため、再実行時には変更された部分のみが再パースされます。

## Claude Code 連携

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` は次のものを書き出します:

- `.claude/skills/graffiti/SKILL.md` — Claude Code が地図のビルド/読み込み/クエリ
  を行えるようにする短いスキル。
- `CLAUDE.md` ブロック(`<!-- graffiti:start -->` / `<!-- graffiti:end -->` の間)
  — 地図が存在するときは grep よりも `graffiti query` を優先するようアシスタントに
  伝えます。
- `--hook` を付けると、`.claude/settings.json` に `graffiti hook` を実行する
  PreToolUse エントリが追加されます。これは `.graffiti/map.json` が存在する場合に
  `Grep`/`Glob` の前に1行のヒントを追加します。このフックがツールをブロックするこ
  とは決してありません。

これは冪等です — いつでも再実行できます。既存の `CLAUDE.md` / `settings.json` の
内容は保持されます。

## LLMなしのクエリ

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` は、ソフトな約2000トークンのノード予算の範囲内で、グラフの関連するスライ
スを返します — モデルなし、埋め込みなし。質問は引用符で囲んでください。

## MCP サーバー

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

MCP対応の任意のクライアントを向けると、あなたのアシスタントは grep する代わりに、
ツールを通じてグラフをたどります。

## ワークスペース(複数リポジトリのフェデレーション)

複数の別々のリポジトリを並べて配置し、それらをまたいでクエリできます —
**マージなしで**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` は、コミット可能なレジストリ(`.graffiti-workspace/workspace.json`)
と、そこから派生したgitignore可能なキャッシュ(`.graffiti-workspace/overlay.json`)
を書き出します。各リポジトリ自身の `.graffiti/map.json` は変更されず、単独でも引
き続き機能します — ワークスペースは薄い計算済みのオーバーレイであり、マージされた
塊では決してありません。

**プロジェクト間リンク:** `.graffiti-workspace/links` の中で、1行に1つずつ明示的
に宣言します —
`frontend::main-go:fetchcart -> backend::main-go:getcart calls`(`#` コメントが使
えます。エンドポイントは `alias::nodeid`)。`graffiti links check` は両方のエンド
ポイントが解決できることを検証します。`graffiti federate --explain` はすべてのリ
ンクを一覧表示します。フェデレーテッドクエリは各ノードにそのメンバーのエイリアス
を前置し、クロスリンクをたどります。`graffiti workspace render` は
`workspace.html` を書き出します — 同じ力学的配置グラフのビューアで、
**プロジェクトがツリーの最上位**に置かれ、プロジェクト間リンクが描画されます。

`.graffiti-workspace/overlay.json` を `.gitignore` に追加してください(これは派生
物であり、再計算可能です)。

## 🛰️ システムオーケストレーション — 多数のサービス、1つのグラフ

<!-- system-orchestration -->
マイクロサービスシステムとは、1つのプロダクトを形作る多数の独立したリポジトリの集
まりです。graffiti はそれぞれを地図化し、次に各サービスの*コントラクトサーフェス*
(そのサービスが提供する `provides` ものと、消費する `consumes` もの)から、
**それらの間のエッジ** — HTTP、gRPC、キュー — を**発見します**。手作業の配線は不要
です。各サービスが自分自身の地図を公開し、オーケストレーターが公開されたアーティ
ファクトをフェデレートして、コンシューマーをプロバイダーにマッチングします。

```bash
# 各サービスの CI(またはローカル)で — その地図を共有ストアに公開する:
graffiti publish --to ../system-store --as carts

# 次に、CI またはオンデマンドで、システム全体に対して:
graffiti system build       # フェデレート + サービス間リンクの自動発見
graffiti system render      # → .graffiti-system/system.html(サービスをレーンとして表示)
graffiti system impact carts::"GET /carts/{}"   # これが変わると誰が壊れる?
graffiti system audit       # ぶら下がったコンシューマー · 孤立したプロバイダー · 曖昧なもの(CI ゲート)
graffiti system query "where is the cart fetched and served"
```

各地図は、`openapi.json`、`.proto`、フレームワークのルート、キュー呼び出し、また
は明示的な `graffiti.contract.json` から抽出された**コントラクトサーフェス**を担い
ます。サービス間リンクは確信度でスコアリングされ、**曖昧な(ambiguous)** コン
シューマーや**ぶら下がった(dangling、デッドエンドポイント)** コンシューマーは報告
され、決して黙って捨てられることはありません。システムストアは単なるディレクトリ
または git リポジトリにすぎません — $0、オフライン、再計算可能です。

## 仕組み

tree-sitter によるパース(ピュアGo、CGOなし)→ エッジ解決 → コミュニティへのクラ
スタリング → 軽量な分析 → 決定論的なシリアライズ。モデルなし、埋め込みなし、ネッ
トワークなし — ただの静的解析です。だからこそ、無料で、プライベートで、再現可能な
のです。

## 保証

- **0回のAPI呼び出し、$0、完全オフライン。** あなたのコードに関する何ものも、マシ
  ンの外へ出ることはありません。
- **決定論的:** 同じリポジトリ → バイト単位で同一の `map.json`(ただし単一の
  `generated_at` タイムスタンプと `root` のベース名を除く)。コミットして、差分を
  取りましょう。
- **単一の静的バイナリ**、ランタイム依存なし、Cツールチェーンなし。

## ライセンス

Source-Available — graffiti を自分自身のリポジトリ上で自由に読み、実行できます
が、再利用、再配布、フォーク、または別のプロジェクトへの組み込みには、作者による
事前の書面での許可が必要です。[LICENSE](LICENSE) を参照してください。

## 作者

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
