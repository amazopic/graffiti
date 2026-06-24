# 🕸️ graffiti — 어떤 저장소든 AI가 질의할 수 있는 코드 그래프로

> 한 번의 명령으로 저장소를 AI 코딩 어시스턴트가 무작정 grep하는 대신 읽어 들이는
> **방향성 지식 그래프**로 바꿉니다. 단일 정적 Go 바이너리 —
> **API 키 0개, 비용 $0, 완전 오프라인, 바이트 단위로 결정적**입니다. **Go, Python,
> JavaScript, TypeScript, Rust, Java, PHP**를 파싱합니다. LLM이 필요 없는 `query`,
> **MCP** 서버, **Claude Code** 통합, 인터랙티브 오프라인 그래프
> 뷰어, 그리고 다중 저장소 워크스페이스 페더레이션을 함께 제공합니다.

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**언어:** [English](README.md) · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

🌐 **웹사이트:** https://amazopic.github.io/graffiti/

```text
$ graffiti .
✓ Done. 0 API calls, $0.  214 files → 1,883 nodes, 4,102 edges, 12 communities.
  The 3 most interesting questions your map can answer:
    1) Which module is the load-bearing wall?
    2) What does the auth flow touch?
    3) Where are the cross-package call hotspots?
```

---

## 왜 만들었나

AI 코딩 어시스턴트는 *볼 수 있는* 만큼만 똑똑합니다. 거대한 저장소에 던져 넣으면
지도 없이 여러분이 할 법한 일을 그대로 합니다. grep을 돌리고, 파일 몇 개를 열어 보고,
추측합니다. 코드의 **형태** — 어떤 함수가 어떤 함수를 호출하는지, 어디에 타입이
정의되어 있는지, 어느 모듈이 구조를 떠받치는 내력벽인지 — 는 결코 보지 못합니다.

**graffiti는 원래 있었어야 할 그 지도입니다.** 한 번의 명령으로
[tree-sitter](https://tree-sitter.github.io/tree-sitter/)로 저장소를 파싱하고,
엣지를 해석하고, 모듈을 클러스터링한 다음, 그래프를 기록합니다 — 기계를 위한 JSON으로,
여러분을 위한 Markdown으로, 그리고 실제로 들여다볼 수 있는 단일 오프라인 HTML로요.
키도, 클라우드도, 비용도 없습니다.

## 설치

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

버전이나 디렉터리를 고정하려면:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

설치 프로그램은 여러분의 OS/아키텍처에 맞는 정적 바이너리를 골라, 릴리스 매니페스트에
대조해 SHA256을 검증한 뒤 설치합니다. `graffiti version`으로 확인하세요.
또는 소스에서 직접 빌드할 수도 있습니다(아래 참고).

## ⚡ Claude Code로 설치하기 (바이브 코딩)

<!-- vibe-install -->
터미널이 필요 없습니다 — **Claude Code**가 전부 알아서 처리하게 하세요. 아래 프롬프트
하나를 Claude Code 세션에 붙여 넣고 각 단계마다 `y`로 답하기만 하면 됩니다. 알맞은
바이너리를 받아 와, 여러분의 저장소에 맞는 지도를 빌드하고, 통합을 연결한 뒤, 그래프를
열어 줍니다:

```text
amazopic이 만든 graffiti를 설치해 줘. github.com/amazopic/graffiti의 최신 릴리스에서 내 OS/아키텍처에 맞는 정적 바이너리를 받거나(Go가 있으면 `make build`로 소스에서 빌드해도 돼), `graffiti`라는 이름으로 내 PATH에 올린 다음 `graffiti version`으로 확인해 줘. 그다음 내 저장소 루트에서 `graffiti .`를 실행해 지도를 빌드하고, `graffiti init --hook`을 실행해 graffiti를 Claude Code에 연결하고, 마지막으로 `.graffiti/map.html`을 열어서 그래프를 볼 수 있게 해 줘. 각 단계 전에 먼저 물어봐 줘.
```

## 빌드

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

`grammar_subset` 빌드 태그는 graffiti가 지원하는 문법(Go, Python, JS, TS, Rust,
Java, PHP, 그리고 go.mod)만을, 순수 Go 런타임
`github.com/odvcencio/gotreesitter`(CGO 없음, WASM 없음)를 통해 포함시킵니다.
이 태그 덕분에 바이너리는 ~10 MB로 유지됩니다. 태그가 없어도 코드는 컴파일되지만 전체
문법 집합(~31 MB)이 링크됩니다. 항상 이 태그를 넘기세요 — Makefile이 대신
처리해 줍니다.

## 지원 언어

| 언어 | 추출 대상 |
|----------|-----------|
| Go | 파일, 함수, 메서드(리시버 기준), 타입, import, 해석된 호출 |
| Python, JavaScript, TypeScript, Rust, Java, PHP | 파일, 함수, 클래스/구조체/인터페이스/열거형/트레이트, 메서드(`Class.method`), import, 저장소 내부 호출 |
| Markdown | 문서 노드 |

Go 외 언어의 추출은 의도적으로 솔직합니다. 흔하고 가치가 높은 구조는 포착하되, 특이한
구문(데코레이터, 제네릭, 중첩 정의, 동적 디스패치)은 추측을 내놓는 대신 **과소
추출**합니다.

## 사용법

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

인자 없이 `graffiti`를 실행하면 전체 명령 목록을 볼 수 있습니다.

## 한 번의 명령, 세 가지 산출물

`graffiti .`은 모든 것을 `<repo>/.graffiti/` 안에 기록합니다:

- **`map.json`** — 그래프 그 자체: 노드, 엣지, 커뮤니티이며,
  `schema/map.schema.json`에 대해 스키마 검증을 거칩니다. 여러분의 AI가 읽는 것이자
  `query`와 MCP 서버가 순회하는 대상입니다.
- **`MAP.md`** — 사람이 읽을 수 있는 요약: 핵심 모듈, 가장 많이 연결된 노드,
  그리고 여러분의 지도가 답할 수 있는 가장 흥미로운 질문 세 가지.
- **`map.html`** — 단일 파일로 완결되는 오프라인 인터랙티브 **힘 기반 레이아웃
  그래프**입니다. CDN도, 서버도, 네트워크도 없이 — 그냥 파일을 열기만 하면 됩니다.

`map.html`에는 **2D/3D 토글**(노드 위에 마우스를 올리면 그 노드와 이웃이 떠오릅니다),
**노드 검색**, **클릭하면 `file:line` 복사**, **섹터 영역**, **client / tests /
external** 카테고리 토글, 그리고 표시/숨김 체크박스가 달린 크기 조절 가능한
**project → directory → file** 트리가 있습니다. CSP에 안전하며 완전히 오프라인으로
동작합니다.

파일별 콘텐츠 해시 캐시가 `<repo>/.graffiti/cache/` 아래에 자리 잡아, 다시 실행할 때
변경된 부분만 재파싱합니다.

## Claude Code 통합

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init`이 기록하는 것:

- `.claude/skills/graffiti/SKILL.md` — Claude Code가 지도를 빌드/읽기/질의할 줄 알도록 하는 짧은 스킬.
- `CLAUDE.md` 블록(`<!-- graffiti:start -->` / `<!-- graffiti:end -->` 사이)으로, 지도가 존재할 때
  어시스턴트가 grep보다 `graffiti query`를 우선하도록 안내합니다.
- `--hook`을 주면 `.claude/settings.json`에 `graffiti hook`을 실행하는 PreToolUse 항목이 추가되며,
  `.graffiti/map.json`이 있을 때 `Grep`/`Glob` 앞에 한 줄짜리 권유를 덧붙입니다. 이 훅은 절대로 도구를 막지 않습니다.

멱등적입니다 — 언제든 다시 실행해도 됩니다. 기존 `CLAUDE.md` / `settings.json` 내용은 보존됩니다.

## LLM 없이 질의하기

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query`는 약 2000 토큰의 느슨한 노드 예산 안에서 그래프의 관련 부분을 반환합니다 —
모델도, 임베딩도 없습니다. 질문은 따옴표로 감싸세요.

## MCP 서버

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

MCP를 지원하는 클라이언트를 여기에 연결하면, 어시스턴트가 grep 대신 도구를 통해
그래프를 순회합니다.

## 워크스페이스 (다중 저장소 페더레이션)

별개의 저장소들을 나란히 두고 그 너머로 질의하세요 — **병합하지 않고**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link`는 커밋 가능한 레지스트리(`.graffiti-workspace/workspace.json`)와
그로부터 파생되어 gitignore할 수 있는 캐시(`.graffiti-workspace/overlay.json`)를
기록합니다. 각 저장소의 자체 `.graffiti/map.json`은 변경되지 않으며 단독으로도 여전히
동작합니다 — 워크스페이스는 얇게 계산된 오버레이일 뿐, 병합된 덩어리가 아닙니다.

**프로젝트 간 링크:** `.graffiti-workspace/links`에 한 줄에 하나씩 명시적으로
선언하세요 — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(`#` 주석 허용; 엔드포인트는 `alias::nodeid` 형식). `graffiti links check`는
양쪽 엔드포인트가 모두 해석되는지 검증하고, `graffiti federate --explain`은 모든
링크를 나열합니다. 페더레이션 질의는 각 노드 앞에 멤버 별칭을 붙이고 교차 링크를
순회합니다. `graffiti workspace render`는 `workspace.html`을 기록합니다 — **프로젝트를
트리의 최상위 레벨로** 둔 동일한 힘 기반 그래프 뷰어이며, 프로젝트 간 링크가 그려집니다.

`.graffiti-workspace/overlay.json`을 `.gitignore`에 추가하세요(파생물이며 다시 계산할 수 있습니다).

## 🛰️ 시스템 오케스트레이션 — 여러 서비스, 하나의 그래프

<!-- system-orchestration -->
마이크로서비스 시스템은 하나의 제품을 이루는 여러 개의 독립된 저장소들입니다. graffiti는
각각을 지도화한 다음, 각 서비스의 *계약 표면*(무엇을 제공(`provides`)하고 무엇을
소비(`consumes`)하는지)으로부터 **그들 사이의 엣지를 발견합니다** — HTTP, gRPC, 큐를요.
수작업 배선은 없습니다: 각 서비스가 자신의 지도를 게시하면, 오케스트레이터가 게시된
아티팩트들을 페더레이션하고 소비자를 제공자에 매칭합니다.

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

각 지도는 `openapi.json`, `.proto`, 프레임워크 라우트, 큐 호출, 또는 명시적인
`graffiti.contract.json`에서 추출된 **계약 표면**을 지닙니다. 서비스 간 링크는 신뢰도로
점수가 매겨지며, **모호한(ambiguous)** 소비자와 **끊어진(dangling)**(죽은 엔드포인트)
소비자는 보고될 뿐, 결코 조용히 버려지지 않습니다. 시스템 스토어는 그저 디렉터리나 git
저장소일 뿐입니다 — $0, 오프라인, 다시 계산 가능합니다.

## 작동 방식

tree-sitter 파싱(순수 Go, CGO 없음) → 엣지 해석 → 커뮤니티로 클러스터링 →
경량 분석 → 결정적 직렬화. 모델도, 임베딩도, 네트워크도 없습니다 — 그저 정적
분석일 뿐입니다. 그래서 무료이고, 사적이며, 재현 가능합니다.

## 보증

- **API 호출 0건, 비용 $0, 완전 오프라인.** 여러분의 코드에 관한 그 무엇도 기기를 떠나지 않습니다.
- **결정적:** 같은 저장소 → 단 하나의 `generated_at` 타임스탬프와 `root` 베이스명을
  제외하면 바이트 단위로 동일한 `map.json`. 커밋하세요, diff로 비교하세요.
- **단일 정적 바이너리**, 런타임 의존성 없음, C 툴체인 없음.

## 라이선스

Source-Available — 여러분 자신의 저장소에서 graffiti를 자유롭게 읽고 실행할 수 있지만,
재사용, 재배포, 포크, 또는 다른 프로젝트에 포함하려면 저작자의 사전 서면 허가가
필요합니다. [LICENSE](LICENSE)를 참고하세요.

## 저작자

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
