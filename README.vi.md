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

## ⚡ Cài đặt bằng Claude Code (vibe-code)

<!-- vibe-install -->
Không cần terminal — hãy để **Claude Code** làm tất cả. Dán đúng một prompt này vào
một phiên Claude Code và trả lời `y` ở mỗi bước. Nó sẽ tải đúng binary, build bản đồ
cho repo của bạn, kết nối phần tích hợp, rồi mở đồ thị:

```text
Cài graffiti của amazopic giúp tôi. Tải đúng static binary cho OS/kiến trúc của tôi từ bản phát hành mới nhất tại github.com/amazopic/graffiti (hoặc build từ mã nguồn bằng `make build` nếu có Go), đặt nó vào PATH với tên `graffiti`, và xác minh bằng `graffiti version`. Sau đó chạy `graffiti .` ở thư mục gốc repo của tôi để build bản đồ, chạy `graffiti init --hook` để kết nối graffiti vào Claude Code, và cuối cùng mở `.graffiti/map.html` để tôi xem được đồ thị. Hãy hỏi tôi trước mỗi bước.
```

<!-- quickstart -->
## Bắt đầu nhanh (60 giây)

```bash
# 1 — cài đặt (hoặc build từ mã nguồn bằng `make build`)
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh

# 2 — lập bản đồ cho repo của bạn (ghi ra .graffiti/map.json, MAP.md, map.html)
cd your-repo
graffiti .

# 3 — xem đồ thị
open .graffiti/map.html        # macOS — dùng `xdg-open` trên Linux, `start` trên Windows

# 4 — đặt câu hỏi cho nó: không cần LLM, không cần API key
graffiti query "where is the user authenticated"
```

Sau đó chỉ cần kết nối nó vào trợ lý AI của bạn một lần:

```bash
graffiti init --hook    # Claude Code: skill + CLAUDE.md + gợi ý grep→query
graffiti serve          # hoặc đưa bản đồ ra cho bất kỳ client MCP nào qua stdio
```

**Thêm câu hỏi ví dụ** — `query` trả về một đồ thị con đã được khoanh vùng trong
ngân sách mềm khoảng ~2.000 token, nên ngữ cảnh giữ được nhỏ gọn và rẻ (hãy đặt câu
hỏi trong dấu ngoặc kép):

```bash
graffiti query "login handler"
graffiti query "what does the checkout flow touch"
graffiti query "where is the cart fetched" ../shop   # nhắm tới một đường dẫn khác
```
<!-- /quickstart -->

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

## 🛰️ Điều phối hệ thống — nhiều dịch vụ, một đồ thị

<!-- system-orchestration -->
Một hệ thống microservice là nhiều repo độc lập cùng tạo nên một sản phẩm. graffiti
lập bản đồ cho từng repo, rồi **khám phá các cạnh giữa chúng** — HTTP, gRPC, hàng đợi —
từ *bề mặt hợp đồng (contract surface)* của mỗi dịch vụ (những gì nó `provides` (cung cấp)
và `consumes` (tiêu thụ)). Không cần đấu nối thủ công: mỗi dịch vụ tự công bố bản đồ
của riêng nó; bộ điều phối liên kết các sản phẩm đã công bố và khớp bên tiêu thụ với
bên cung cấp.

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

Mỗi bản đồ mang theo một **bề mặt hợp đồng (contract surface)** được trích xuất từ
`openapi.json`, `.proto`, các route của framework, các lời gọi hàng đợi, hoặc một
tệp `graffiti.contract.json` tường minh. Các liên kết xuyên dịch vụ được chấm điểm
theo độ tin cậy; các bên tiêu thụ **mơ hồ (ambiguous)** và **lơ lửng (dangling)** —
trỏ tới điểm cuối đã chết — đều được báo cáo, không bao giờ bị âm thầm bỏ qua. Kho
lưu trữ hệ thống chỉ đơn thuần là một thư mục hoặc một repo git — $0, ngoại tuyến,
có thể tính toán lại.

<!-- system-walkthrough -->
### Một thư mục chứa các dịch vụ, từng bước một

Giả sử các dịch vụ của bạn nằm trong một thư mục cha, mỗi dịch vụ ở thư mục riêng của nó:

```text
myproject/                ← thư mục cha = "kho lưu trữ hệ thống" dùng chung
├── orders/               ← một dịch vụ (Go)
├── web/                  ← một dịch vụ (React/TS)
└── payments/             ← một dịch vụ (Python)
```

**1. Build và publish từng dịch vụ** vào một kho lưu trữ tại thư mục cha (`--to .`).
`publish` tái sử dụng bản đồ hiện có, nên hãy build trước để cập nhật các thay đổi mã:

```bash
cd myproject
for d in */; do
  d=${d%/}
  graffiti build "$d" && graffiti publish "$d" --to .
done
```

Tên dịch vụ mặc định lấy theo tên thư mục của nó; ghi đè bằng `--as <name>`.

> ⚠️ **Khi publish lại:** `publish` **không** build lại một bản đồ đã có. Sau khi
> thay đổi mã, hãy luôn chạy `graffiti build <service>` trước (vòng lặp ở trên đã
> làm vậy) rồi mới `publish` — nếu không bạn sẽ publish một bản đồ cũ.

**2. Build đồ thị hệ thống** — liên kết các bản đồ và tự động khám phá các liên kết:

```bash
graffiti system build
# ✓ System "myproject": 3 services → 7 cross-service links (0 ambiguous, 0 dangling, 2 orphan). 0 API calls, $0.
```

**3. Dùng nó:**

```bash
graffiti system render          # → .graffiti-system/system.html (các dịch vụ là cấp cao nhất của cây)
graffiti system impact orders   # ai bị hỏng nếu orders thay đổi (trực tiếp + bắc cầu)
graffiti system audit           # bên tiêu thụ lơ lửng · bên cung cấp mồ côi · mơ hồ (thoát khác 0 → cổng CI)
graffiti system status          # những dịch vụ nào đã trôi dạt kể từ lần build trước
graffiti system query "where is the order created"   # truy xuất không cần LLM xuyên suốt cả hệ thống
graffiti system list            # các dịch vụ đã đăng ký
```

**Những gì xuất hiện trong thư mục cha:**

```text
myproject/.graffiti-system/
├── system.json                 # sổ đăng ký các dịch vụ (hãy commit tệp này)
├── overlay.json                # các liên kết đã khám phá (dẫn xuất — an toàn để .gitignore)
├── system.html                 # bản đồ hệ thống trực quan
└── services/<name>/map.json    # bản đồ đã publish của mỗi dịch vụ
```

**Cải thiện độ chính xác của liên kết.** Tự động phát hiện bao phủ Go (net/http,
gin/chi/echo), Flask, FastAPI, Django/DRF, Spring, NestJS, ASP.NET, Ktor, các client
frontend (React/Vue/Angular/Svelte), gRPC và Kafka/NATS. Ở những chỗ chừng đó là chưa
đủ, hãy thả một trong các tệp này vào thư mục gốc của dịch vụ (độ tin cậy cao nhất trước):

| Tệp | Cung cấp |
|------|-------|
| `graffiti.contract.json` | `provides` / `consumes` tường minh — bất kỳ stack nào, độ tin cậy cao nhất |
| `openapi.json` / `swagger.json` | các route HTTP dưới dạng `provides` |
| `*.proto` | các phương thức gRPC dưới dạng `provides` |

`graffiti.contract.json` tối thiểu:

```json
{
  "provides": [{ "kind": "http", "name": "GET /orders/{id}" }],
  "consumes": [{ "kind": "rpc",  "name": "Payments.Charge" }]
}
```

**Đặt cổng CI dựa trên các điểm cuối đã chết** — `audit` thoát với mã khác 0 khi một
bên tiêu thụ trỏ tới một điểm cuối mà không gì cung cấp:

```bash
graffiti system build && graffiti system audit
```
<!-- /system-walkthrough -->

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
