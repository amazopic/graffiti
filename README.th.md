# 🕸️ graffiti — เปลี่ยน repo ใดก็ได้ให้กลายเป็นกราฟโค้ดที่ค้นหาได้สำหรับ AI

> คำสั่งเดียวเปลี่ยน repository ของคุณให้เป็น **กราฟความรู้แบบมีทิศทาง** ที่ผู้ช่วย
> เขียนโค้ด AI ของคุณอ่านแทนการ grep แบบมองไม่เห็นภาพ ไบนารี Go แบบ static เดี่ยว ๆ —
> **ไม่ต้องใช้ API key, ราคา $0, ทำงานออฟไลน์เต็มรูปแบบ, ให้ผลลัพธ์เหมือนเดิมระดับไบต์**
> วิเคราะห์ **Go, Python, JavaScript, TypeScript, Rust, Java และ PHP** มาพร้อมกับ
> `query` ที่ไม่ต้องใช้ LLM, เซิร์ฟเวอร์ **MCP**, การผสานรวมกับ **Claude Code**, ตัวดู
> กราฟแบบโต้ตอบที่ทำงานออฟไลน์ และการรวม workspace ข้ามหลาย repo

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**ภาษา:** English · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

🌐 **เว็บไซต์:** https://amazopic.github.io/graffiti/

```text
$ graffiti .
✓ Done. 0 API calls, $0.  214 files → 1,883 nodes, 4,102 edges, 12 communities.
  The 3 most interesting questions your map can answer:
    1) Which module is the load-bearing wall?
    2) What does the auth flow touch?
    3) Where are the cross-package call hotspots?
```

---

## ทำไมจึงมีเครื่องมือนี้

ผู้ช่วยเขียนโค้ด AI จะเก่งได้แค่เท่าที่มัน *มองเห็น* เท่านั้น โยนมันเข้าไปใน repo ขนาดใหญ่
แล้วมันก็จะทำสิ่งที่คุณจะทำเมื่อไม่มีแผนที่ นั่นคือ grep, เปิดไฟล์ดูสองสามไฟล์, แล้วเดา
มันไม่เคยเห็น **รูปทรง** ของโค้ด — ฟังก์ชันไหนเรียกฟังก์ชันไหน, type ถูกนิยามไว้ที่ใด,
โมดูลไหนคือกำแพงรับน้ำหนักของทั้งระบบ

**graffiti คือแผนที่ที่ควรจะมีอยู่ตั้งแต่แรก** คำสั่งเดียววิเคราะห์ repo ด้วย
[tree-sitter](https://tree-sitter.github.io/tree-sitter/), แก้ปัญหาเส้นเชื่อม (edge),
จัดกลุ่มโมดูล แล้วเขียนกราฟออกมา — เป็น JSON สำหรับเครื่อง, เป็น Markdown สำหรับคุณ,
และเป็น HTML ออฟไลน์ไฟล์เดียวที่คุณดูได้จริง ๆ ไม่ต้องใช้ key ไม่ต้องใช้คลาวด์ ไม่มีค่าใช้จ่าย

## การติดตั้ง

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

ปักหมุดเวอร์ชันหรือไดเรกทอรี:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

ตัวติดตั้งจะเลือกไบนารี static ที่ถูกต้องสำหรับ OS/สถาปัตยกรรมของคุณ, ตรวจสอบค่า SHA256
เทียบกับ manifest ของรีลีส แล้วติดตั้งให้ ตรวจสอบด้วย `graffiti version`
หรือจะ build จากซอร์สก็ได้ (ด้านล่าง)

## ⚡ ติดตั้งด้วย Claude Code (vibe-code)

<!-- vibe-install -->
ไม่ต้องใช้เทอร์มินัล — ปล่อยให้ **Claude Code** ทำให้ทั้งหมด แค่วางพรอมต์เดียวนี้
ลงในเซสชัน Claude Code แล้วตอบ `y` ในแต่ละขั้นตอน มันจะดึงไบนารีที่ถูกต้องมาให้,
สร้างแผนที่สำหรับ repo ของคุณ, ตั้งค่าการผสานรวมให้เรียบร้อย แล้วเปิดกราฟขึ้นมา:

```text
ช่วยติดตั้ง graffiti ของ amazopic ให้ที โดยดาวน์โหลดไบนารี static ที่ถูกต้องสำหรับ OS/สถาปัตยกรรมของฉันจากรีลีสล่าสุดที่ github.com/amazopic/graffiti (หรือ build จากซอร์สด้วย `make build` ถ้ามี Go อยู่แล้ว), วางมันไว้บน PATH ของฉันในชื่อ `graffiti` แล้วตรวจสอบด้วย `graffiti version` จากนั้นรัน `graffiti .` ที่รากของ repo ฉันเพื่อสร้างแผนที่, รัน `graffiti init --hook` เพื่อตั้งค่า graffiti ให้เข้ากับ Claude Code และสุดท้ายเปิด `.graffiti/map.html` เพื่อให้ฉันเห็นกราฟ ถามฉันก่อนทุกขั้นตอนด้วย
```

<!-- quickstart -->
## เริ่มใช้งานเร็ว (60 วินาที)

```bash
# 1 — ติดตั้ง (หรือ build จากซอร์สด้วย `make build`)
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh

# 2 — สร้างแผนที่ของ repo คุณ (เขียน .graffiti/map.json, MAP.md, map.html)
cd your-repo
graffiti .

# 3 — ดูกราฟ
open .graffiti/map.html        # macOS — ใช้ `xdg-open` บน Linux, `start` บน Windows

# 4 — ถามคำถามกับมัน: ไม่ต้องใช้ LLM, ไม่ต้องใช้ API key
graffiti query "where is the user authenticated"
```

จากนั้นเชื่อมมันเข้ากับผู้ช่วย AI ของคุณเพียงครั้งเดียว:

```bash
graffiti init --hook    # Claude Code: skill + CLAUDE.md + grep→query nudge
graffiti serve          # หรือเปิดเผยแผนที่ให้ MCP client ใดก็ได้ผ่าน stdio
```

**ตัวอย่างคำถามเพิ่มเติม** — `query` คืนซับกราฟที่จำกัดขอบเขตภายในงบประมาณแบบยืดหยุ่น
ที่ประมาณ ~2,000 โทเค็น ดังนั้นบริบทจึงเล็กและประหยัด (ใส่เครื่องหมายคำพูดให้กับคำถาม):

```bash
graffiti query "login handler"
graffiti query "what does the checkout flow touch"
graffiti query "where is the cart fetched" ../shop   # ระบุพาธอื่น
```
<!-- /quickstart -->

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

build tag `grammar_subset` จะรวมเฉพาะ grammar ที่ graffiti รองรับเท่านั้น (Go,
Python, JS, TS, Rust, Java, PHP รวมถึง go.mod) ผ่าน runtime แบบ pure-Go
`github.com/odvcencio/gotreesitter` (ไม่มี CGO, ไม่มี WASM) ซึ่งทำให้ไบนารีมีขนาด
~10 MB หากไม่มี tag เหล่านี้ โค้ดก็ยังคอมไพล์ได้แต่จะลิงก์ชุด grammar เต็ม
(~31 MB) ส่งพวกมันเสมอ — Makefile จัดการเรื่องนี้ให้คุณแล้ว

## ภาษาที่รองรับ

| ภาษา | สิ่งที่สกัดออกมา |
|----------|-----------|
| Go | ไฟล์, ฟังก์ชัน, เมธอด (ตาม receiver), type, import, การเรียกที่แก้แล้ว |
| Python, JavaScript, TypeScript, Rust, Java, PHP | ไฟล์, ฟังก์ชัน, class/struct/interface/enum/trait, เมธอด (`Class.method`), import, การเรียกภายใน repo |
| Markdown | โหนดเอกสาร |

การสกัดสำหรับภาษาที่ไม่ใช่ Go ตั้งใจให้ซื่อตรง คือจับโครงสร้างที่พบบ่อยและมีคุณค่าสูง
แล้ว **สกัดน้อยกว่าความจริง** สำหรับโครงสร้างแปลก ๆ (decorator, generic, การนิยามแบบซ้อน,
dynamic dispatch) แทนที่จะปล่อยให้มันเดา

## การใช้งาน

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

รัน `graffiti` โดยไม่ใส่อาร์กิวเมนต์เพื่อดูรายการคำสั่งทั้งหมด

## คำสั่งเดียว, สามผลผลิต

`graffiti .` เขียนทุกอย่างลงใน `<repo>/.graffiti/`:

- **`map.json`** — ตัวกราฟเอง: โหนด, เส้นเชื่อม, community, ตรวจสอบ schema เทียบกับ
  `schema/map.schema.json` นี่คือสิ่งที่ AI ของคุณอ่าน และเป็นสิ่งที่ `query`
  กับเซิร์ฟเวอร์ MCP เดินสำรวจ
- **`MAP.md`** — บทสรุปที่มนุษย์อ่านได้: โมดูลอันดับต้น ๆ, โหนดที่เชื่อมต่อมากที่สุด,
  และสามคำถามที่น่าสนใจที่สุดที่แผนที่ของคุณตอบได้
- **`map.html`** — **กราฟแบบ force-directed** ไฟล์เดียวที่ครบในตัว ทำงานออฟไลน์ และโต้ตอบได้
  ไม่ต้องใช้ CDN, ไม่ต้องใช้เซิร์ฟเวอร์, ไม่ต้องใช้เครือข่าย — แค่เปิดไฟล์ขึ้นมา

`map.html` มี **ปุ่มสลับ 2D/3D** (เลื่อนเมาส์ไปวางจะยกโหนดนั้นกับโหนดข้างเคียงขึ้นมา),
**ค้นหาโหนด**, **คลิกเพื่อคัดลอก `file:line`**, **โซนภาค (sector)**, ปุ่มสลับหมวด
**client / tests / external** และต้นไม้ **project → directory → file** ที่ปรับขนาดได้
พร้อมช่องทำเครื่องหมายแสดง/ซ่อน มันปลอดภัยต่อ CSP และทำงานออฟไลน์ได้อย่างสมบูรณ์

แคชแฮชเนื้อหารายไฟล์อยู่ภายใต้ `<repo>/.graffiti/cache/` ดังนั้นการรันซ้ำจะวิเคราะห์ใหม่
เฉพาะส่วนที่เปลี่ยนไปเท่านั้น

## การผสานรวมกับ Claude Code

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` จะเขียน:

- `.claude/skills/graffiti/SKILL.md` — สกิลสั้น ๆ เพื่อให้ Claude Code รู้ว่าต้อง build/อ่าน/query แผนที่
- บล็อก `CLAUDE.md` (ระหว่าง `<!-- graffiti:start -->` / `<!-- graffiti:end -->`) ที่บอกให้
  ผู้ช่วยเลือกใช้ `graffiti query` แทน grep เมื่อมีแผนที่อยู่
- เมื่อใช้ `--hook` จะเพิ่มรายการ PreToolUse ใน `.claude/settings.json` ที่รัน `graffiti hook`
  ซึ่งจะเพิ่มข้อความเตือนหนึ่งบรรทัดก่อน `Grep`/`Glob` เมื่อมี `.graffiti/map.json` อยู่ hook นี้ไม่เคยบล็อกเครื่องมือใด ๆ

มันเป็น idempotent — รันซ้ำได้ทุกเมื่อ เนื้อหา `CLAUDE.md` / `settings.json` ที่มีอยู่จะถูกรักษาไว้

## ค้นหาโดยไม่ต้องใช้ LLM

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` คืนสไลซ์ของกราฟที่เกี่ยวข้องภายในงบประมาณโหนดแบบยืดหยุ่นที่ประมาณ ~2000 โทเค็น —
ไม่มีโมเดล, ไม่มี embedding ใส่เครื่องหมายคำพูดให้กับคำถามด้วย

## เซิร์ฟเวอร์ MCP

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

ชี้ไคลเอนต์ที่รองรับ MCP ตัวใดก็ได้มาที่มัน แล้วผู้ช่วยของคุณก็จะเดินสำรวจกราฟผ่านเครื่องมือ
แทนการ grep

## Workspaces (การรวมข้ามหลาย repo)

วาง repo แยกกันไว้เคียงข้างกันแล้ว query ข้ามกันได้ — **โดยไม่ต้องรวมเข้าด้วยกัน**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` จะเขียน registry ที่ commit ได้ (`.graffiti-workspace/workspace.json`)
และแคชที่ derive ขึ้นมา ซึ่ง gitignore ได้ (`.graffiti-workspace/overlay.json`) ไฟล์
`.graffiti/map.json` ของแต่ละ repo ยังไม่เปลี่ยนแปลงและยังทำงานแบบเดี่ยวได้ — workspace
เป็นเพียงโอเวอร์เลย์ที่คำนวณขึ้นแบบบาง ๆ ไม่ใช่ก้อนที่รวมเข้าด้วยกัน

**ลิงก์ข้ามโปรเจกต์:** ระบุมันอย่างชัดเจนใน `.graffiti-workspace/links`
หนึ่งบรรทัดต่อหนึ่งลิงก์ — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(อนุญาตให้มีคอมเมนต์ `#`; endpoint คือ `alias::nodeid`) `graffiti links check` ตรวจสอบว่า
ปลายทางทั้งสองด้านแก้ค่าได้; `graffiti federate --explain` แสดงทุกลิงก์ การ query แบบรวม
จะนำหน้าแต่ละโหนดด้วย alias ของสมาชิกแล้วเดินสำรวจลิงก์ข้ามกัน `graffiti workspace
render` เขียน `workspace.html` — ตัวดูกราฟแบบ force-graph เดียวกันโดยมี **โปรเจกต์เป็น
ระดับบนสุด** ของต้นไม้ และวาดลิงก์ข้ามโปรเจกต์ไว้

เพิ่ม `.graffiti-workspace/overlay.json` ลงใน `.gitignore` (เพราะมัน derive มาและคำนวณใหม่ได้)

## 🛰️ การจัดวงดนตรีของระบบ (system orchestration) — หลายเซอร์วิส, กราฟเดียว

<!-- system-orchestration -->
ระบบไมโครเซอร์วิสคือ repo อิสระจำนวนมากที่ประกอบกันเป็นผลิตภัณฑ์เดียว graffiti
สร้างแผนที่ของแต่ละตัว แล้ว **ค้นพบเส้นเชื่อมระหว่างกัน** — HTTP, gRPC, คิว — จาก
*พื้นผิวสัญญา (contract surface)* ของแต่ละเซอร์วิส (สิ่งที่มัน `provides` หรือให้บริการ
และสิ่งที่มัน `consumes` หรือเรียกใช้) ไม่ต้องเดินสายเชื่อมด้วยมือ คือแต่ละเซอร์วิส
เผยแพร่แผนที่ของตัวเอง แล้วตัวจัดวงดนตรี (orchestrator) จะรวม artifact ที่เผยแพร่ไว้
เข้าด้วยกัน และจับคู่ผู้เรียกใช้ (consumer) กับผู้ให้บริการ (provider)

```bash
# ใน CI ของแต่ละเซอร์วิส (หรือบนเครื่องโลคัล) — เผยแพร่แผนที่ของมันลงในที่เก็บที่ใช้ร่วมกัน:
graffiti publish --to ../system-store --as carts

# จากนั้นใน CI หรือเมื่อต้องการ ทำกับทั้งระบบ:
graffiti system build       # federate + auto-discover cross-service links
graffiti system render      # → .graffiti-system/system.html (services as lanes)
graffiti system impact carts::"GET /carts/{}"   # who breaks if this changes?
graffiti system audit       # dangling consumers · orphan providers · ambiguous (CI gate)
graffiti system query "where is the cart fetched and served"
```

แผนที่แต่ละตัวพก **พื้นผิวสัญญา (contract surface)** ที่สกัดมาจาก `openapi.json`,
`.proto`, เส้นทาง (route) ของเฟรมเวิร์ก, การเรียกคิว หรือ `graffiti.contract.json`
แบบระบุชัดเจน ลิงก์ข้ามเซอร์วิสจะถูกให้คะแนนตามความเชื่อมั่น (confidence) ผู้เรียกใช้
ที่ **กำกวม (ambiguous)** และ **ลอยค้าง (dangling — ปลายทางตาย)** จะถูกรายงานออกมา
ไม่ถูกทิ้งไปอย่างเงียบ ๆ ที่เก็บของระบบเป็นเพียงไดเรกทอรีหรือ git repo เท่านั้น — ราคา
$0, ออฟไลน์, คำนวณใหม่ได้

<!-- system-walkthrough -->
### โฟลเดอร์ของเซอร์วิสต่าง ๆ ทีละขั้นตอน

สมมติว่าเซอร์วิสของคุณอยู่ในโฟลเดอร์แม่เดียวกัน แต่ละตัวอยู่ในไดเรกทอรีของตัวเอง:

```text
myproject/                ← โฟลเดอร์แม่ = "ที่เก็บของระบบ" ที่ใช้ร่วมกัน
├── orders/               ← เซอร์วิสหนึ่งตัว (Go)
├── web/                  ← เซอร์วิสหนึ่งตัว (React/TS)
└── payments/             ← เซอร์วิสหนึ่งตัว (Python)
```

**1. Build และ publish แต่ละเซอร์วิส** ลงในที่เก็บที่โฟลเดอร์แม่ (`--to .`)
`publish` จะนำแผนที่ที่มีอยู่มาใช้ซ้ำ ดังนั้นให้ build ก่อนเพื่อรับการเปลี่ยนแปลงของโค้ด:

```bash
cd myproject
for d in */; do
  d=${d%/}
  graffiti build "$d" && graffiti publish "$d" --to .
done
```

ชื่อเซอร์วิสจะตั้งค่าเริ่มต้นเป็นชื่อโฟลเดอร์ของมัน; ใช้ `--as <name>` เพื่อกำหนดเอง

> ⚠️ **เมื่อ publish ซ้ำ:** `publish` **ไม่** build แผนที่ที่มีอยู่ใหม่ หลังจาก
> เปลี่ยนโค้ด ให้รัน `graffiti build <service>` ก่อนเสมอ (ลูปด้านบนทำให้แล้ว)
> แล้วจึง `publish` — มิฉะนั้นคุณจะ publish แผนที่ที่ล้าสมัย

**2. Build กราฟของระบบ** — รวมแผนที่เข้าด้วยกันและค้นพบลิงก์โดยอัตโนมัติ:

```bash
graffiti system build
# ✓ System "myproject": 3 services → 7 cross-service links (0 ambiguous, 0 dangling, 2 orphan). 0 API calls, $0.
```

**3. ใช้งานมัน:**

```bash
graffiti system render          # → .graffiti-system/system.html (services as the top tree level)
graffiti system impact orders   # who breaks if orders changes (direct + transitive)
graffiti system audit           # dangling consumers · orphan providers · ambiguous (non-zero exit → CI gate)
graffiti system status          # which services drifted since the last build
graffiti system query "where is the order created"   # LLM-free retrieval across the whole system
graffiti system list            # registered services
```

**สิ่งที่ปรากฏในโฟลเดอร์แม่:**

```text
myproject/.graffiti-system/
├── system.json                 # the registry of services (commit this)
├── overlay.json                # discovered links (derived — safe to .gitignore)
├── system.html                 # the visual system map
└── services/<name>/map.json    # each service's published map
```

**ปรับปรุงความแม่นยำของลิงก์** การตรวจจับอัตโนมัติครอบคลุม Go (net/http, gin/chi/echo),
Flask, FastAPI, Django/DRF, Spring, NestJS, ASP.NET, Ktor, ไคลเอนต์ฝั่งหน้า
(React/Vue/Angular/Svelte), gRPC และ Kafka/NATS เมื่อแค่นั้นยังไม่พอ ให้วาง
ไฟล์อย่างใดอย่างหนึ่งต่อไปนี้ลงในรากของเซอร์วิส (เรียงตามความเชื่อมั่นสูงสุดก่อน):

| File | ให้อะไร |
|------|-------|
| `graffiti.contract.json` | ระบุ `provides` / `consumes` อย่างชัดเจน — stack ใดก็ได้, ความเชื่อมั่นสูงสุด |
| `openapi.json` / `swagger.json` | เส้นทาง HTTP เป็น `provides` |
| `*.proto` | เมธอด gRPC เป็น `provides` |

`graffiti.contract.json` ขั้นต่ำ:

```json
{
  "provides": [{ "kind": "http", "name": "GET /orders/{id}" }],
  "consumes": [{ "kind": "rpc",  "name": "Payments.Charge" }]
}
```

**กั้น CI ไว้ที่ endpoint ที่ตายแล้ว** — `audit` จะออกด้วยค่าไม่เป็นศูนย์เมื่อผู้เรียกใช้
ชี้ไปยัง endpoint ที่ไม่มีใครให้บริการ:

```bash
graffiti system build && graffiti system audit
```
<!-- /system-walkthrough -->

## มันทำงานอย่างไร

การวิเคราะห์ด้วย tree-sitter (pure-Go, ไม่มี CGO) → การแก้เส้นเชื่อม → การจัดกลุ่มเป็น
community → การวิเคราะห์แบบเบา → การ serialize แบบ deterministic ไม่มีโมเดล, ไม่มี
embedding, ไม่มีเครือข่าย — เป็นเพียงการวิเคราะห์เชิงสถิต (static analysis) นั่นคือเหตุผล
ที่มันฟรี, เป็นส่วนตัว, และทำซ้ำผลได้

## การรับประกัน

- **เรียก API 0 ครั้ง, ราคา $0, ออฟไลน์เต็มรูปแบบ** ไม่มีอะไรเกี่ยวกับโค้ดของคุณที่หลุดออกจากเครื่องของคุณ
- **Deterministic:** repo เดียวกัน → `map.json` ที่เหมือนกันทุกไบต์ ยกเว้นเพียง timestamp
  `generated_at` ตัวเดียวและชื่อ basename ของ `root` เท่านั้น commit มันได้; diff มันได้
- **ไบนารี static เดี่ยว ๆ**, ไม่มี dependency ตอน runtime, ไม่ต้องใช้ C toolchain

## สัญญาอนุญาต

Source-Available — อ่านและรัน graffiti บน repository ของคุณเองได้อย่างอิสระ แต่การนำ
ไปใช้ซ้ำ, แจกจ่ายซ้ำ, fork หรือรวมเข้ากับโปรเจกต์อื่นใด ต้องได้รับอนุญาตเป็นลายลักษณ์
อักษรจากผู้สร้างก่อน ดู [LICENSE](LICENSE)

## ผู้สร้าง

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
