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
