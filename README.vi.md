# 🕸️ graffiti — biến mọi repo thành đồ thị mã nguồn có thể truy vấn cho AI

> Chỉ một câu lệnh biến kho mã của bạn thành một **đồ thị tri thức có hướng** mà
> trợ lý lập trình AI sẽ đọc thay vì grep một cách mù quáng. Một binary Go tĩnh
> duy nhất — **không cần API key, $0, hoạt động hoàn toàn ngoại tuyến, tất định
> đến từng byte.** Phân tích **Go, Python, JavaScript, TypeScript, Rust, Java và
> PHP**. Đi kèm lệnh `query` không cần LLM, một máy chủ **MCP**, tích hợp **Claude
> Code**, trình xem đồ thị tương tác ngoại tuyến, và liên kết workspace đa repo.

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**Ngôn ngữ:** English · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

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

## Vì sao có công cụ này

Một trợ lý lập trình AI chỉ giỏi đúng bằng những gì nó có thể *nhìn thấy*. Thả nó
vào một repo lớn và nó sẽ làm đúng điều bạn làm khi không có bản đồ: nó grep, mở vài
tệp, rồi đoán. Nó không bao giờ thấy được **hình dạng** của mã — hàm nào gọi hàm
nào, kiểu dữ liệu được định nghĩa ở đâu, module nào là bức tường chịu lực.

**graffiti chính là tấm bản đồ lẽ ra phải có ở đó.** Một câu lệnh phân tích repo
bằng [tree-sitter](https://tree-sitter.github.io/tree-sitter/), giải quyết các cạnh,
gom cụm các module, rồi viết ra một đồ thị — dưới dạng JSON cho máy, dưới dạng
Markdown cho bạn, và dưới dạng một tệp HTML ngoại tuyến duy nhất mà bạn thực sự có
thể xem được. Không cần key. Không cần cloud. Không tốn chi phí.

## Cài đặt

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

Ghim một phiên bản hoặc thư mục cụ thể:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

Trình cài đặt chọn đúng binary tĩnh cho OS/kiến trúc của bạn, xác minh SHA256 của nó
dựa trên manifest bản phát hành, rồi cài đặt. Kiểm chứng bằng `graffiti version`.
Hoặc build từ mã nguồn (bên dưới).

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

Các build tag `grammar_subset` chỉ đóng gói những grammar mà graffiti hỗ trợ (Go,
Python, JS, TS, Rust, Java, PHP, cộng thêm go.mod) thông qua runtime thuần Go
`github.com/odvcencio/gotreesitter` (không CGO, không WASM). Chúng giữ binary ở mức
~10 MB; nếu không có chúng thì mã vẫn biên dịch được nhưng sẽ liên kết với toàn bộ
tập grammar (~31 MB). Hãy luôn truyền chúng — Makefile đã làm việc này giúp bạn.

## Các ngôn ngữ được hỗ trợ

| Ngôn ngữ | Trích xuất |
|----------|-----------|
| Go | tệp, hàm, phương thức (theo receiver), kiểu, import, lời gọi đã giải quyết |
| Python, JavaScript, TypeScript, Rust, Java, PHP | tệp, hàm, class/struct/interface/enum/trait, phương thức (`Class.method`), import, lời gọi trong nội bộ repo |
| Markdown | nút tài liệu |

Việc trích xuất ngoài Go cố ý trung thực: nó nắm bắt cấu trúc phổ biến, giá trị cao
và **trích xuất thiếu** những kết cấu kỳ lạ (decorator, generic, định nghĩa lồng
nhau, dynamic dispatch) thay vì đưa ra phỏng đoán.

## Cách dùng

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

Chạy `graffiti` không kèm tham số nào để xem danh sách lệnh đầy đủ.

## Một câu lệnh, ba sản phẩm

`graffiti .` ghi mọi thứ vào `<repo>/.graffiti/`:

- **`map.json`** — bản thân đồ thị: các nút, các cạnh, các cộng đồng, được kiểm tra
  theo schema dựa trên `schema/map.schema.json`. Đây là thứ AI của bạn đọc và là thứ
  mà `query` cùng máy chủ MCP duyệt qua.
- **`MAP.md`** — bản tóm tắt dễ đọc cho con người: các module hàng đầu, các nút được
  kết nối nhiều nhất, và ba câu hỏi thú vị nhất mà bản đồ của bạn có thể trả lời.
- **`map.html`** — một **đồ thị định hướng bằng lực (force-directed)** tương tác,
  ngoại tuyến, khép kín trong một tệp duy nhất. Không CDN, không máy chủ, không mạng
  — chỉ cần mở tệp ra.

`map.html` có **nút chuyển 2D/3D** (di chuột để nâng một nút cùng các nút lân cận),
**tìm kiếm nút**, **nhấp để sao chép `file:line`**, **các vùng sector**, công tắc
phân loại **client / tests / external**, và một cây **dự án → thư mục → tệp** có thể
thay đổi kích thước với các ô đánh dấu ẩn/hiện. Nó an toàn với CSP và hoạt động hoàn
toàn ngoại tuyến.

Một bộ nhớ đệm băm nội dung theo từng tệp nằm trong `<repo>/.graffiti/cache/`, nên
các lần chạy lại chỉ phân tích lại những gì đã thay đổi.

## Tích hợp Claude Code

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` ghi:

- `.claude/skills/graffiti/SKILL.md` — một skill ngắn để Claude Code biết cách build/đọc/truy vấn bản đồ.
- một khối `CLAUDE.md` (giữa `<!-- graffiti:start -->` / `<!-- graffiti:end -->`) nhắc trợ lý
  ưu tiên dùng `graffiti query` thay vì grep khi đã có bản đồ.
- với `--hook`, một mục PreToolUse trong `.claude/settings.json` chạy `graffiti hook`, nó thêm
  một dòng gợi ý trước `Grep`/`Glob` khi có mặt `.graffiti/map.json`. Hook không bao giờ chặn một công cụ.

Lệnh này có tính idempotent — chạy lại bất cứ lúc nào; nội dung `CLAUDE.md` / `settings.json` hiện có sẽ được giữ nguyên.

## Truy vấn mà không cần LLM

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` trả về một lát cắt liên quan của đồ thị trong một ngân sách mềm khoảng
~2000 token cho các nút — không cần mô hình, không cần embedding. Hãy đặt câu hỏi
trong dấu ngoặc kép.

## Máy chủ MCP

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

Trỏ bất kỳ client hỗ trợ MCP nào vào nó và trợ lý của bạn sẽ duyệt đồ thị thông qua
các công cụ thay vì grep.

## Workspaces (liên kết đa repo)

Đặt các repo riêng biệt cạnh nhau và truy vấn xuyên qua chúng — **mà không cần hợp nhất**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` ghi một sổ đăng ký có thể commit (`.graffiti-workspace/workspace.json`)
và một bộ nhớ đệm dẫn xuất, có thể đưa vào gitignore (`.graffiti-workspace/overlay.json`).
Tệp `.graffiti/map.json` riêng của mỗi repo vẫn không bị thay đổi và vẫn hoạt động độc
lập — workspace chỉ là một lớp phủ tính toán mỏng, không bao giờ là một khối hợp nhất.

**Liên kết xuyên dự án:** khẳng định chúng một cách tường minh trong
`.graffiti-workspace/links`, mỗi liên kết một dòng — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(cho phép chú thích `#`; điểm cuối có dạng `alias::nodeid`). `graffiti links check` xác thực
rằng cả hai điểm cuối đều giải quyết được; `graffiti federate --explain` liệt kê mọi liên kết.
Truy vấn liên kết gắn tiền tố cho mỗi nút bằng alias thành viên của nó và duyệt qua các liên kết
chéo. `graffiti workspace render` ghi một `workspace.html` — vẫn là trình xem đồ thị lực đó nhưng
với **các dự án nằm ở cấp cao nhất** của cây và các liên kết xuyên dự án được vẽ ra.

Thêm `.graffiti-workspace/overlay.json` vào `.gitignore` (nó là dẫn xuất và có thể tính lại được).

## Cách hoạt động

phân tích tree-sitter (thuần Go, không CGO) → giải quyết cạnh → gom cụm thành các
cộng đồng → phân tích nhẹ → tuần tự hóa tất định. Không mô hình, không embedding,
không mạng — chỉ là phân tích tĩnh. Đó là lý do vì sao nó miễn phí, riêng tư và có
thể tái lập.

## Các đảm bảo

- **0 lời gọi API, $0, hoàn toàn ngoại tuyến.** Không có gì về mã của bạn rời khỏi máy của bạn.
- **Tất định:** cùng một repo → `map.json` giống hệt đến từng byte, ngoại trừ duy nhất
  dấu thời gian `generated_at` và tên cơ sở (basename) của `root`. Hãy commit nó; hãy diff nó.
- **Một binary tĩnh duy nhất**, không phụ thuộc runtime, không cần bộ công cụ C.

## Giấy phép

Source-Available — bạn được tự do đọc và chạy graffiti trên repo của riêng mình,
nhưng mọi hành vi tái sử dụng, phân phối lại, fork, hay đưa vào một dự án khác đều
cần có sự cho phép trước bằng văn bản từ tác giả. Xem [LICENSE](LICENSE).

## Tác giả

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
