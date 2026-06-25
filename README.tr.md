# 🕸️ graffiti — herhangi bir depoyu yapay zeka için sorgulanabilir bir kod grafiğine dönüştürün

> Tek bir komut, deponuzu yapay zeka kodlama asistanınızın körlemesine grep yapmak
> yerine okuyacağı bir **yönlü bilgi grafiğine** dönüştürür. Tek bir statik Go ikili
> dosyası — **sıfır API anahtarı, 0 dolar, tamamen çevrimdışı, bayt düzeyinde
> belirlenimci.** **Go, Python, JavaScript, TypeScript, Rust, Java ve PHP** ayrıştırır.
> LLM kullanmayan bir `query`, bir **MCP** sunucusu, **Claude Code** entegrasyonu,
> etkileşimli çevrimdışı bir grafik görüntüleyici ve çok depolu çalışma alanı
> federasyonu sunar.

[![License: Source-Available](https://img.shields.io/badge/license-Source--Available-orange.svg)](LICENSE)
[![Made for Claude Code](https://img.shields.io/badge/made%20for-Claude%20Code-7c3aed.svg)](https://claude.com/claude-code)
[![Languages](https://img.shields.io/badge/parses-Go·Python·JS·TS·Rust·Java·PHP-00a000.svg)](#supported-languages)
[![Single static binary](https://img.shields.io/badge/binary-static·CGO--free·~10MB-blue.svg)](#build)
[![Cost](https://img.shields.io/badge/%240·offline·deterministic-000.svg)](#guarantees)
[![Author](https://img.shields.io/badge/author-Yevgeniy%20Achin-blue.svg)](mailto:amazopic@gmail.com)

**Diller:** English · [Русский](README.ru.md) · [Français](README.fr.md) · [Deutsch](README.de.md) · [Українська](README.uk.md) · [Slovenščina](README.sl.md) · [Italiano](README.it.md) · [Español](README.es.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md) · [العربية](README.ar.md) · [Português](README.pt.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Tiếng Việt](README.vi.md) · [हिन्दी](README.hi.md) · [繁體中文](README.zh-tw.md) · [Polski](README.pl.md) · [ไทย](README.th.md) · [עברית](README.he.md) · [বাংলা](README.bn.md) · [اردو](README.ur.md)

🌐 **Web sitesi:** https://amazopic.github.io/graffiti/

```text
$ graffiti .
✓ Done. 0 API calls, $0.  214 files → 1,883 nodes, 4,102 edges, 12 communities.
  The 3 most interesting questions your map can answer:
    1) Which module is the load-bearing wall?
    2) What does the auth flow touch?
    3) Where are the cross-package call hotspots?
```

---

## Neden var

Bir yapay zeka kodlama asistanı, ancak *görebildiği* kadar iyidir. Onu büyük bir
depoya bırakın ve haritası olmadan sizin yapacağınız şeyi yapsın: grep yapar, birkaç
dosya açar, tahmin yürütür. Kodun **biçimini** asla görmez — hangi fonksiyon hangisini
çağırıyor, bir tür nerede tanımlanmış, hangi modül yük taşıyan duvar.

**graffiti, orada olması gereken haritadır.** Tek bir komut, depoyu
[tree-sitter](https://tree-sitter.github.io/tree-sitter/) ile ayrıştırır, kenarları
çözümler, modülleri kümeler ve bir grafik yazar — makine için JSON olarak, sizin için
Markdown olarak ve gerçekten bakabileceğiniz tek bir çevrimdışı HTML olarak. Anahtar
yok. Bulut yok. Maliyet yok.

## Kurulum

```bash
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh
```

Bir sürümü veya dizini sabitleyin:

```bash
GRAFFITI_VERSION=v0.1.0 INSTALL_DIR="$HOME/.local/bin" \
  sh -c "$(curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh)"
```

Yükleyici, işletim sisteminize/mimarinize uygun statik ikili dosyayı seçer, SHA256
değerini sürüm bildirimine göre doğrular ve kurar. `graffiti version` ile doğrulayın.
Ya da kaynaktan derleyin (aşağıda).

## ⚡ Claude Code ile kurun (vibe-code)

<!-- vibe-install -->
Terminale gerek yok — bırakın her şeyi **Claude Code** yapsın. Şu tek istemi bir Claude
Code oturumuna yapıştırın ve her adımda `y` ile yanıt verin. Doğru ikili dosyayı indirir,
deponuz için haritayı oluşturur, entegrasyonu bağlar ve grafiği açar:

```text
Benim için amazopic'in graffiti'sini kur. github.com/amazopic/graffiti adresindeki en son sürümden işletim sistemime/mimarime uygun statik ikili dosyayı indir (ya da Go varsa `make build` ile kaynaktan derle), `graffiti` olarak PATH'ime ekle ve `graffiti version` ile doğrula. Ardından haritayı oluşturmak için depo kökümde `graffiti .` çalıştır, graffiti'yi Claude Code'a bağlamak için `graffiti init --hook` çalıştır ve son olarak grafiği görebilmem için `.graffiti/map.html` dosyasını aç. Her adımdan önce sor.
```

<!-- quickstart -->
## Hızlı başlangıç (60 saniye)

```bash
# 1 — kur (ya da `make build` ile kaynaktan derle)
curl -fsSL https://raw.githubusercontent.com/amazopic/graffiti/main/scripts/install.sh | sh

# 2 — deponu haritala (.graffiti/map.json, MAP.md, map.html yazar)
cd your-repo
graffiti .

# 3 — grafiğe bak
open .graffiti/map.html        # macOS — Linux'ta `xdg-open`, Windows'ta `start` kullan

# 4 — ona sorular sor: LLM yok, API anahtarı yok
graffiti query "where is the user authenticated"
```

Ardından bir kez yapay zeka asistanına bağla:

```bash
graffiti init --hook    # Claude Code: beceri + CLAUDE.md + grep→query teşviki
graffiti serve          # ya da haritayı stdio üzerinden herhangi bir MCP istemcisine sun
```

**Daha fazla örnek soru** — `query`, yumuşak ~2.000 jetonluk bir bütçe içinde
kapsamlı bir alt grafik döndürür, böylece bağlam küçük ve ucuz kalır (soruyu tırnak
içine alın):

```bash
graffiti query "login handler"
graffiti query "what does the checkout flow touch"
graffiti query "where is the cart fetched" ../shop   # başka bir yolu hedefle
```
<!-- /quickstart -->

## Derleme

```bash
make build      # builds ./graffiti (CGO-free, ~10MB, 7 language grammars)
make test       # runs the full test suite with the required build tags
make xcompile   # cross-compiles static binaries for all targets into dist/
```

`grammar_subset` derleme etiketleri yalnızca graffiti'nin desteklediği gramerleri (Go,
Python, JS, TS, Rust, Java, PHP, ek olarak go.mod) saf Go çalışma zamanı
`github.com/odvcencio/gotreesitter` (CGO yok, WASM yok) aracılığıyla sunar. İkili
dosyayı ~10 MB'de tutarlar; bunlar olmadan kod yine derlenir ama tam gramer setini
(~31 MB) bağlar. Bunları her zaman geçirin — Makefile bunu sizin için yapar.

## Desteklenen diller

| Dil | Çıkarılanlar |
|----------|-----------|
| Go | dosyalar, fonksiyonlar, metotlar (alıcıya göre), türler, içe aktarımlar, çözümlenmiş çağrılar |
| Python, JavaScript, TypeScript, Rust, Java, PHP | dosyalar, fonksiyonlar, sınıflar/yapılar/arayüzler/enum'lar/trait'ler, metotlar (`Class.method`), içe aktarımlar, depo içi çağrılar |
| Markdown | belge düğümleri |

Go dışı çıkarım kasıtlı olarak dürüsttür: yaygın, yüksek değerli yapıyı yakalar ve
egzotik yapıları (dekoratörler, jenerikler, iç içe tanımlar, dinamik gönderim) tahmin
üretmek yerine **eksik çıkarır**.

## Kullanım

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

Tam komut listesi için `graffiti`'yi argümansız çalıştırın.

## Tek komut, üç çıktı

`graffiti .` her şeyi `<repo>/.graffiti/` içine yazar:

- **`map.json`** — grafiğin kendisi: düğümler, kenarlar, topluluklar,
  `schema/map.schema.json`'a göre şema denetimi yapılmış. Yapay zekanızın okuduğu ve
  `query` ile MCP sunucusunun gezindiği şey budur.
- **`MAP.md`** — insanların okuyabileceği bir özet: en üst modüller, en çok bağlantılı
  düğümler ve haritanızın yanıtlayabileceği en ilginç üç soru.
- **`map.html`** — tek başına yeterli, çevrimdışı, etkileşimli **kuvvet yönlendirmeli
  grafik**. CDN yok, sunucu yok, ağ yok — dosyayı açmanız yeterli.

`map.html` bir **2B/3B geçişine** (üzerine gelince bir düğümü ve komşularını
yükseltir), **düğüm aramasına**, **tıkla-kopyala `file:line`** özelliğine, **sektör
bölgelerine**, **istemci / testler / harici** kategori geçişlerine ve göster/gizle
onay kutularıyla yeniden boyutlandırılabilir bir **proje → dizin → dosya** ağacına
sahiptir. CSP açısından güvenlidir ve tamamen çevrimdışı çalışır.

`<repo>/.graffiti/cache/` altında dosya başına bir içerik karması önbelleği bulunur,
böylece yeniden çalıştırmalar yalnızca değişeni yeniden ayrıştırır.

## Claude Code entegrasyonu

```bash
graffiti init                 # install the skill + CLAUDE.md block (project)
graffiti init --hook          # also install the PreToolUse nudge (grep → graffiti query)
graffiti init --user          # install into ~/.claude instead of the repo
```

`graffiti init` şunları yazar:

- `.claude/skills/graffiti/SKILL.md` — Claude Code'un haritayı oluşturmayı/okumayı/sorgulamayı bilmesi için kısa bir beceri.
- bir `CLAUDE.md` bloğu (`<!-- graffiti:start -->` / `<!-- graffiti:end -->` arasında),
  asistana harita varken grep yerine `graffiti query`'yi tercih etmesini söyler.
- `--hook` ile, `.graffiti/map.json` mevcut olduğunda `Grep`/`Glob` öncesinde tek
  satırlık bir teşvik ekleyen `graffiti hook` çalıştıran bir `.claude/settings.json`
  PreToolUse girdisi. Kanca hiçbir zaman bir aracı engellemez.

İdempotenttir — istediğiniz zaman yeniden çalıştırın; mevcut `CLAUDE.md` /
`settings.json` içeriği korunur.

## LLM olmadan sorgulama

```bash
graffiti query "login handler"            # scoped subgraph for the current repo
graffiti query "where is cart fetched" ../shop
```

`query`, yumuşak ~2000 jetonluk bir düğüm bütçesi içinde grafiğin ilgili bir dilimini
döndürür — model yok, gömme yok. Soruyu tırnak içine alın.

## MCP sunucusu

```bash
graffiti serve                # MCP over stdio (JSON-RPC 2.0)
```

MCP yeteneğine sahip herhangi bir istemciyi ona yönlendirin; asistanınız grep yapmak
yerine araçlar aracılığıyla grafikte gezinir.

## Çalışma alanları (çok depolu federasyon)

Ayrı depoları yan yana koyun ve **birleştirmeden** aralarında sorgu yapın:

```bash
graffiti link ../frontend ../backend          # federate (builds members if needed)
graffiti query --workspace "where is the cart fetched and served"
graffiti serve  --workspace                    # MCP over the federation
graffiti update --workspace                    # rebuild changed members + recompute links
graffiti workspace render                      # → .graffiti-workspace/workspace.html
```

`graffiti link`, commit edilebilir bir kayıt (`.graffiti-workspace/workspace.json`) ve
ondan türetilmiş, gitignore'lanabilir bir önbellek (`.graffiti-workspace/overlay.json`)
yazar. Her deponun kendi `.graffiti/map.json` dosyası değişmez ve tek başına çalışmaya
devam eder — çalışma alanı, ince ve hesaplanmış bir katmandır, asla birleştirilmiş bir
blob değildir.

**Projeler arası bağlantılar:** bunları `.graffiti-workspace/links` içinde, satır
başına bir tane olacak şekilde açıkça beyan edin —
`frontend::main-go:fetchcart -> backend::main-go:getcart calls` (`#` yorumlarına izin
verilir; uç noktalar `alias::nodeid` biçimindedir). `graffiti links check` her iki uç
noktanın da çözümlendiğini doğrular; `graffiti federate --explain` her bağlantıyı
listeler. Federe sorgu her düğümün önüne üye takma adını ekler ve çapraz bağlantılarda
gezinir. `graffiti workspace render` bir `workspace.html` yazar — ağacın **üst
düzeyinde projeler** olan ve projeler arası bağlantıların çizildiği aynı kuvvet grafiği
görüntüleyicisi.

`.graffiti-workspace/overlay.json`'ı `.gitignore`'a ekleyin (türetilmiştir ve yeniden
hesaplanabilir).

## 🛰️ Sistem orkestrasyonu — birçok servis, tek bir grafik

<!-- system-orchestration -->
Bir mikroservis sistemi, tek bir ürünü oluşturan birçok bağımsız depodan oluşur.
graffiti her birini haritalandırır, ardından her servisin *sözleşme yüzeyinden*
(neyi `provides`, yani sunduğu ve neyi `consumes`, yani tükettiği) yola çıkarak
**aralarındaki kenarları keşfeder** — HTTP, gRPC, kuyruklar. Elle bağlama yok: her
servis kendi haritasını yayımlar; orkestratör yayımlanan yapıtları birleştirir ve
tüketicileri sağlayıcılarla eşleştirir.

```bash
# her servisin CI'ında (ya da yerelde) — haritasını paylaşılan bir depoya yayımlayın:
graffiti publish --to ../system-store --as carts

# ardından, CI'da veya talep üzerine, tüm sistem genelinde:
graffiti system build       # birleştir + servisler arası bağlantıları otomatik keşfet
graffiti system render      # → .graffiti-system/system.html (servisler şeritler olarak)
graffiti system impact carts::"GET /carts/{}"   # bu değişirse kim bozulur?
graffiti system audit       # boşta tüketiciler · yetim sağlayıcılar · belirsizler (CI kapısı)
graffiti system query "where is the cart fetched and served"
```

Her harita, `openapi.json`, `.proto`, çerçeve rotaları, kuyruk çağrıları ya da açık
bir `graffiti.contract.json` dosyasından çıkarılan bir **sözleşme yüzeyi** taşır.
Servisler arası bağlantılar güven düzeyine göre puanlanır; **belirsiz** ve **boşta**
(ölü uç noktalı) tüketiciler raporlanır, asla sessizce atılmaz. Sistem deposu yalnızca
bir dizin veya git deposudur — $0, çevrimdışı, yeniden hesaplanabilir.

<!-- system-walkthrough -->
### Adım adım, bir servisler klasörü

Diyelim ki servisleriniz tek bir üst klasörde, her biri kendi dizininde yaşıyor:

```text
myproject/                ← üst klasör = paylaşılan "sistem deposu"
├── orders/               ← bir servis (Go)
├── web/                  ← bir servis (React/TS)
└── payments/             ← bir servis (Python)
```

**1. Her servisi derleyin ve yayımlayın** — üst klasördeki bir depoya (`--to .`).
`publish` mevcut bir haritayı yeniden kullanır, bu yüzden kod değişikliklerini almak için önce derleyin:

```bash
cd myproject
for d in */; do
  d=${d%/}
  graffiti build "$d" && graffiti publish "$d" --to .
done
```

Servis adı varsayılan olarak klasör adıdır; `--as <name>` ile değiştirin.

> ⚠️ **Yeniden yayımlarken:** `publish` mevcut bir haritayı yeniden **derlemez**.
> Kodu değiştirdikten sonra her zaman önce `graffiti build <service>` çalıştırın
> (yukarıdaki döngü bunu yapar) ve ardından `publish` yapın — aksi halde eskimiş
> bir harita yayımlarsınız.

**2. Sistem grafiğini oluşturun** — haritaları birleştirin ve bağlantıları otomatik keşfedin:

```bash
graffiti system build
# ✓ System "myproject": 3 services → 7 cross-service links (0 ambiguous, 0 dangling, 2 orphan). 0 API calls, $0.
```

**3. Kullanın:**

```bash
graffiti system render          # → .graffiti-system/system.html (servisler en üst ağaç düzeyi olarak)
graffiti system impact orders   # orders değişirse kim bozulur (doğrudan + geçişli)
graffiti system audit           # boşta tüketiciler · yetim sağlayıcılar · belirsizler (sıfır olmayan çıkış → CI kapısı)
graffiti system status          # son derlemeden bu yana hangi servisler kaydı
graffiti system query "where is the order created"   # tüm sistem genelinde LLM kullanmayan erişim
graffiti system list            # kayıtlı servisler
```

**Üst klasöre ne iner:**

```text
myproject/.graffiti-system/
├── system.json                 # servislerin kaydı (bunu commit'leyin)
├── overlay.json                # keşfedilen bağlantılar (türetilmiş — .gitignore güvenli)
├── system.html                 # görsel sistem haritası
└── services/<name>/map.json    # her servisin yayımlanmış haritası
```

**Bağlantı doğruluğunu artırın.** Otomatik algılama şunları kapsar: Go (net/http,
gin/chi/echo), Flask, FastAPI, Django/DRF, Spring, NestJS, ASP.NET, Ktor, frontend
istemcileri (React/Vue/Angular/Svelte), gRPC ve Kafka/NATS. Bu yeterli olmadığında,
şunlardan birini bir servis köküne bırakın (önce en yüksek güven):

| Dosya | Sağladığı |
|------|-------|
| `graffiti.contract.json` | açık `provides` / `consumes` — herhangi bir yığın, en yüksek güven |
| `openapi.json` / `swagger.json` | HTTP rotaları `provides` olarak |
| `*.proto` | gRPC metotları `provides` olarak |

Asgari `graffiti.contract.json`:

```json
{
  "provides": [{ "kind": "http", "name": "GET /orders/{id}" }],
  "consumes": [{ "kind": "rpc",  "name": "Payments.Charge" }]
}
```

**CI'yı ölü uç noktalara karşı kapılayın** — bir tüketici hiçbir şeyin sağlamadığı
bir uç noktaya işaret ettiğinde `audit` sıfır olmayan çıkış yapar:

```bash
graffiti system build && graffiti system audit
```
<!-- /system-walkthrough -->

## Nasıl çalışır

tree-sitter ayrıştırma (saf Go, CGO yok) → kenar çözümleme → topluluklara kümeleme →
hafif analiz → belirlenimci serileştirme. Model yok, gömme yok, ağ yok — yalnızca
statik analiz. İşte bu yüzden ücretsiz, gizli ve yeniden üretilebilirdir.

## Garantiler

- **0 API çağrısı, 0 dolar, tamamen çevrimdışı.** Kodunuzla ilgili hiçbir şey
  makinenizden ayrılmaz.
- **Belirlenimci:** aynı depo → tek `generated_at` zaman damgası ve `root` taban adı
  dışında bayt düzeyinde aynı `map.json`. Commit edin; diff'ini alın.
- **Tek bir statik ikili dosya**, çalışma zamanı bağımlılığı yok, C araç zinciri yok.

## Lisans

Source-Available — graffiti'yi kendi depolarınızda özgürce okuyup çalıştırın, ancak
herhangi bir yeniden kullanım, yeniden dağıtım, fork veya başka bir projeye dahil etme,
yazardan önceden yazılı izin gerektirir. Bkz. [LICENSE](LICENSE).

## Yazar

Yevgeniy Achin · [amazopic@gmail.com](mailto:amazopic@gmail.com) · [github.com/amazopic](https://github.com/amazopic)
