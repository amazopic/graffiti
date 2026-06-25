# 🕸️ graffiti — ubah repositori apa pun menjadi graf kode yang bisa di-query untuk AI

> Satu perintah mengubah repositori Anda menjadi **graf pengetahuan terarah** yang
> dibaca asisten coding AI Anda alih-alih melakukan grep secara membabi buta. Sebuah
> binary Go statis tunggal — **tanpa kunci API, $0, sepenuhnya offline, deterministik
> byte demi byte.** Mem-parsing **Go, Python, JavaScript, TypeScript, Rust, Java, dan
> PHP**. Dilengkapi `query` tanpa LLM, sebuah server **MCP**, integrasi **Claude Code**,
> penampil graf interaktif yang offline, dan federasi workspace multi-repo.

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**Bahasa:** [English](README.md) · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

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

## Mengapa ini ada

Sebuah asisten coding AI hanya sebaik apa yang bisa *dilihatnya*. Lemparkan ia ke
dalam repo yang besar dan ia melakukan apa yang akan Anda lakukan tanpa peta: ia
melakukan grep, membuka beberapa berkas, menebak. Ia tidak pernah melihat **bentuk**
dari kode itu — fungsi mana yang memanggil yang mana, di mana sebuah tipe didefinisikan,
modul mana yang menjadi dinding penopang.

**graffiti adalah peta yang seharusnya ada di sana.** Satu perintah mem-parsing repo
dengan [tree-sitter](https://tree-sitter.github.io/tree-sitter/), menyelesaikan edge-nya,
mengelompokkan modul-modulnya, dan menulis sebuah graf — sebagai JSON untuk mesin, sebagai
Markdown untuk Anda, dan sebagai satu HTML offline yang benar-benar bisa Anda lihat. Tanpa
kunci. Tanpa cloud. Tanpa biaya.

## Instalasi

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

Sematkan sebuah versi atau direktori:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

Installer memilih binary statis yang tepat untuk OS/arsitektur Anda, memverifikasi SHA256-nya
terhadap manifes rilis, dan menginstalnya. Verifikasi dengan `graffiti version`.
Atau build dari sumber (di bawah).

## ⚡ Install dengan Claude Code (vibe-code)

<!-- vibe-install -->
Tidak perlu terminal — biarkan **Claude Code** yang mengerjakan semuanya. Tempelkan satu prompt
ini ke dalam sesi Claude Code dan jawab `y` di setiap langkah. Ia mengunduh binary yang tepat,
membangun peta untuk repo Anda, menyambungkan integrasinya, dan membuka grafnya:

```text
Install graffiti buatan amazopic untuk saya. Unduh binary statis yang tepat untuk OS/arsitektur saya dari rilis terbaru di github.com/amazopic/graffiti (atau build dari sumber dengan `make build` jika Go tersedia), letakkan di PATH saya sebagai `graffiti`, dan verifikasi dengan `graffiti version`. Lalu jalankan `graffiti .` di root repo saya untuk membangun peta, jalankan `graffiti init --hook` untuk menyambungkan graffiti ke Claude Code, dan terakhir buka `.graffiti/map.html` agar saya bisa melihat grafnya. Tanyakan sebelum setiap langkah.
```

<!-- quickstart -->
## Mulai cepat (60 detik)

```bash
# 1 — instal (atau build dari sumber dengan `make build`)
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh

# 2 — petakan repo Anda (menulis .graffiti/map.json, MAP.md, map.html)
cd your-repo
graffiti .

# 3 — lihat grafnya
open .graffiti/map.html        # macOS — gunakan `xdg-open` di Linux, `start` di Windows

# 4 — ajukan pertanyaan padanya: tanpa LLM, tanpa kunci API
graffiti query "where is the user authenticated"
```

Lalu sambungkan ke asisten AI Anda sekali saja:

```bash
graffiti init --hook    # Claude Code: skill + CLAUDE.md + dorongan grep→query
graffiti serve          # atau ekspos peta ke klien MCP apa pun melalui stdio
```

**Lebih banyak contoh pertanyaan** — `query` mengembalikan subgraf tercakup dalam
anggaran lunak ~2.000-token, sehingga konteks tetap kecil dan murah (kutip
pertanyaannya):

```bash
graffiti query "login handler"
graffiti query "what does the checkout flow touch"
graffiti query "where is the cart fetched" ../shop   # targetkan path lain
```
<!-- /quickstart -->

## Build

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

Build tag `grammar_subset` hanya menyertakan grammar yang didukung graffiti (Go,
Python, JS, TS, Rust, Java, PHP, ditambah go.mod) melalui runtime Go murni
`github.com/odvcencio/gotreesitter` (tanpa CGO, tanpa WASM). Mereka menjaga binary tetap
~10 MB; tanpanya kode masih bisa dikompilasi tetapi me-link seluruh set grammar
(~31 MB). Selalu sertakan keduanya — Makefile melakukan ini untuk Anda.

## Bahasa yang didukung

| Bahasa | Yang diekstrak |
|----------|-----------|
| Go | berkas, fungsi, metode (berdasarkan receiver), tipe, import, panggilan yang teresolusi |
| Python, JavaScript, TypeScript, Rust, Java, PHP | berkas, fungsi, class/struct/interface/enum/trait, metode (`Class.method`), import, panggilan dalam-repo |
| Markdown | node dokumen |

Ekstraksi non-Go sengaja dibuat jujur: ia menangkap struktur umum bernilai tinggi dan
**kurang mengekstrak** konstruksi eksotis (decorator, generic, definisi bersarang,
dynamic dispatch) alih-alih mengeluarkan tebakan.

## Penggunaan

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

Jalankan `graffiti` tanpa argumen untuk daftar perintah lengkap.

## Satu perintah, tiga artefak

`graffiti .` menulis semuanya ke dalam `<repo>/.graffiti/`:

- **`map.json`** — graf itu sendiri: node, edge, komunitas, diperiksa skemanya
  terhadap `schema/map.schema.json`. Inilah yang dibaca AI Anda dan yang ditelusuri oleh
  `query` serta server MCP.
- **`MAP.md`** — ringkasan yang mudah dibaca manusia: modul teratas, node yang paling
  banyak terhubung, dan tiga pertanyaan paling menarik yang bisa dijawab peta Anda.
- **`map.html`** — satu **graf force-directed** interaktif, offline, dan mandiri.
  Tanpa CDN, tanpa server, tanpa jaringan — cukup buka berkasnya.

`map.html` memiliki **toggle 2D/3D** (mengarahkan kursor mengangkat sebuah node beserta
tetangganya), **pencarian node**, **klik-untuk-menyalin `file:line`**, **zona sektor**,
toggle kategori **client / tests / external**, dan pohon **project → directory → file**
yang bisa diubah ukurannya dengan kotak centang tampilkan/sembunyikan. Ia aman-CSP dan
bekerja sepenuhnya offline.

Sebuah cache content-hash per-berkas tersimpan di bawah `<repo>/.graffiti/cache/`, sehingga
proses ulang hanya mem-parsing kembali apa yang berubah.

## Integrasi Claude Code

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` menulis:

- `.claude/skills/graffiti/SKILL.md` — sebuah skill singkat agar Claude Code tahu cara build/baca/query peta.
- sebuah blok `CLAUDE.md` (di antara `<!-- graffiti:start -->` / `<!-- graffiti:end -->`) yang memberi tahu
  asisten untuk lebih memilih `graffiti query` daripada grep ketika sebuah peta ada.
- dengan `--hook`, sebuah entri PreToolUse di `.claude/settings.json` yang menjalankan `graffiti hook`, yang menambahkan
  saran satu baris sebelum `Grep`/`Glob` ketika `.graffiti/map.json` tersedia. Hook tidak pernah memblokir sebuah tool.

Ia idempoten — jalankan ulang kapan saja; konten `CLAUDE.md` / `settings.json` yang sudah ada tetap dipertahankan.

## Query tanpa LLM

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query` mengembalikan irisan graf yang relevan dalam anggaran node lunak ~2000-token —
tanpa model, tanpa embedding. Kutip pertanyaannya.

## Server MCP

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

Arahkan klien apa pun yang mendukung MCP ke server itu dan asisten Anda akan menelusuri graf
melalui tool alih-alih melakukan grep.

## Workspace (federasi multi-repo)

Letakkan repo-repo terpisah berdampingan dan query lintas mereka — **tanpa menggabungkan**:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link` menulis sebuah registry yang bisa di-commit (`.graffiti-workspace/workspace.json`)
dan sebuah cache turunan yang bisa di-gitignore (`.graffiti-workspace/overlay.json`). Berkas
`.graffiti/map.json` milik setiap repo tidak berubah dan tetap berfungsi mandiri — workspace adalah
overlay tipis yang dihitung, bukan blob gabungan.

**Tautan lintas-proyek:** nyatakan secara eksplisit di `.graffiti-workspace/links`,
satu per baris — `frontend::main-go:fetchcart -> backend::main-go:getcart calls`
(komentar `#` diperbolehkan; endpoint berupa `alias::nodeid`). `graffiti links check` memvalidasi
bahwa kedua endpoint teresolusi; `graffiti federate --explain` mencantumkan setiap tautan. Query terfederasi
memberi prefiks setiap node dengan alias anggotanya dan menelusuri tautan lintas-repo. `graffiti workspace
render` menulis sebuah `workspace.html` — penampil force-graph yang sama dengan **proyek sebagai
level teratas** pohon dan tautan lintas-proyek tergambar.

Tambahkan `.graffiti-workspace/overlay.json` ke `.gitignore` (ia turunan dan bisa dihitung ulang).

## 🛰️ Orkestrasi sistem — banyak layanan, satu graf

<!-- system-orchestration -->
Sebuah sistem microservice adalah banyak repo independen yang membentuk satu produk.
graffiti memetakan masing-masing, lalu **menemukan edge di antara mereka** — HTTP, gRPC,
queue — dari *permukaan kontrak* setiap layanan (apa yang `provides` (disediakan) dan
`consumes` (dikonsumsi)). Tanpa penyambungan manual: setiap layanan menerbitkan petanya
sendiri; orchestrator menyatukan artefak yang diterbitkan dan mencocokkan consumer dengan
provider.

```bash
# di CI setiap layanan (atau secara lokal) — terbitkan petanya ke store bersama:
graffiti publish --to ../system-store --as carts

# lalu, di CI atau sesuai permintaan, atas keseluruhan sistem:
graffiti system build       # federate + auto-discover cross-service links
graffiti system render      # → .graffiti-system/system.html (services as lanes)
graffiti system impact carts::"GET /carts/{}"   # who breaks if this changes?
graffiti system audit       # dangling consumers · orphan providers · ambiguous (CI gate)
graffiti system query "where is the cart fetched and served"
```

Setiap peta membawa **permukaan kontrak** yang diekstrak dari `openapi.json`, `.proto`,
route framework, panggilan queue, atau `graffiti.contract.json` eksplisit. Tautan lintas-layanan
diberi skor berdasarkan keyakinan; consumer yang **ambigu** dan **dangling** (endpoint-mati)
dilaporkan, tidak pernah dibuang secara diam-diam. System store hanyalah sebuah direktori
atau repo git — $0, offline, bisa dihitung ulang.

<!-- system-walkthrough -->
### Sebuah folder layanan, langkah demi langkah

Misalkan layanan-layanan Anda berada dalam satu folder induk, masing-masing di direktorinya sendiri:

```text
myproject/                ← folder induk = "system store" bersama
├── orders/               ← sebuah layanan (Go)
├── web/                  ← sebuah layanan (React/TS)
└── payments/             ← sebuah layanan (Python)
```

**1. Build dan publish setiap layanan** ke sebuah store di folder induk (`--to .`).
`publish` menggunakan kembali peta yang sudah ada, jadi build dulu untuk menangkap perubahan kode:

```bash
cd myproject
for d in */; do
  d=${d%/}
  graffiti build "$d" && graffiti publish "$d" --to .
done
```

Nama layanan secara default mengikuti nama foldernya; timpa dengan `--as <name>`.

> ⚠️ **Saat publish ulang:** `publish` **tidak** membangun ulang peta yang sudah ada.
> Setelah mengubah kode, selalu jalankan `graffiti build <service>` terlebih dahulu (loop di atas
> melakukannya) lalu `publish` — jika tidak, Anda mem-publish peta yang basi.

**2. Build graf sistem** — federasikan peta-peta itu dan temukan tautannya secara otomatis:

```bash
graffiti system build
# ✓ System "myproject": 3 services → 7 cross-service links (0 ambiguous, 0 dangling, 2 orphan). 0 API calls, $0.
```

**3. Gunakan:**

```bash
graffiti system render          # → .graffiti-system/system.html (services as the top tree level)
graffiti system impact orders   # who breaks if orders changes (direct + transitive)
graffiti system audit           # dangling consumers · orphan providers · ambiguous (non-zero exit → CI gate)
graffiti system status          # which services drifted since the last build
graffiti system query "where is the order created"   # LLM-free retrieval across the whole system
graffiti system list            # registered services
```

**Apa yang muncul di folder induk:**

```text
myproject/.graffiti-system/
├── system.json                 # the registry of services (commit this)
├── overlay.json                # discovered links (derived — safe to .gitignore)
├── system.html                 # the visual system map
└── services/<name>/map.json    # each service's published map
```

**Tingkatkan akurasi tautan.** Deteksi otomatis mencakup Go (net/http, gin/chi/echo),
Flask, FastAPI, Django/DRF, Spring, NestJS, ASP.NET, Ktor, klien frontend
(React/Vue/Angular/Svelte), gRPC dan Kafka/NATS. Bila itu tidak cukup, letakkan
salah satu dari ini ke root sebuah layanan (keyakinan tertinggi lebih dahulu):

| File | Memberikan |
|------|-------|
| `graffiti.contract.json` | `provides` / `consumes` eksplisit — stack apa pun, keyakinan tertinggi |
| `openapi.json` / `swagger.json` | route HTTP sebagai `provides` |
| `*.proto` | metode gRPC sebagai `provides` |

`graffiti.contract.json` minimal:

```json
{
  "provides": [{ "kind": "http", "name": "GET /orders/{id}" }],
  "consumes": [{ "kind": "rpc",  "name": "Payments.Charge" }]
}
```

**Jadikan endpoint mati sebagai gate CI** — `audit` keluar dengan kode non-nol ketika sebuah
consumer menunjuk ke endpoint yang tidak disediakan apa pun:

```bash
graffiti system build && graffiti system audit
```
<!-- /system-walkthrough -->

## Cara kerjanya

Parsing tree-sitter (Go murni, tanpa CGO) → resolusi edge → pengelompokan ke dalam
komunitas → analisis ringan → serialisasi deterministik. Tanpa model, tanpa
embedding, tanpa jaringan — hanya analisis statis. Itulah mengapa ia gratis, privat, dan
dapat direproduksi.

## Jaminan

- **0 panggilan API, $0, sepenuhnya offline.** Tidak ada apa pun tentang kode Anda yang meninggalkan mesin Anda.
- **Deterministik:** repo yang sama → `map.json` yang identik byte demi byte selain timestamp
  `generated_at` tunggal dan basename `root`. Commit-lah; diff-lah.
- **Binary statis tunggal**, tanpa dependensi runtime, tanpa toolchain C.

## Lisensi

Source-Available — baca dan jalankan graffiti dengan bebas pada repositori Anda sendiri, tetapi setiap
penggunaan ulang, redistribusi, fork, atau penyertaan dalam proyek lain memerlukan izin tertulis
sebelumnya dari penulis. Lihat [LICENSE](LICENSE).

## Penulis

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
