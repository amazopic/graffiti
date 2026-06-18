// ─────────────────────────────────────────────────────────────────────
// i18n — thin core. English dictionary is the permanent inline fallback.
// All other locales are lazy-loaded per-code chunks from ./locales/<code>.js
// ─────────────────────────────────────────────────────────────────────

export const ASSET_V = '3';

export const supportedLocales = [
  { code: 'en',   label: 'English',     native: 'English'    },
  { code: 'ru',   label: 'Russian',     native: 'Русский'    },
  { code: 'fr',   label: 'French',      native: 'Français'   },
  { code: 'de',   label: 'German',      native: 'Deutsch'    },
  { code: 'uk',   label: 'Ukrainian',   native: 'Українська' },
  { code: 'sl',   label: 'Slovenian',   native: 'Slovenščina'},
  { code: 'it',   label: 'Italian',     native: 'Italiano'   },
  { code: 'es',   label: 'Spanish',     native: 'Español'    },
  { code: 'zh',   label: 'Chinese',     native: '中文'        },
  { code: 'ja',   label: 'Japanese',    native: '日本語'      },
  { code: 'ko',   label: 'Korean',      native: '한국어'      },
  { code: 'ar',   label: 'Arabic',      native: 'العربية',     rtl: true },
  { code: 'pt',   label: 'Portuguese',  native: 'Português'  },
  { code: 'tr',   label: 'Turkish',     native: 'Türkçe'     },
  { code: 'id',   label: 'Indonesian',  native: 'Bahasa Indonesia' },
  { code: 'vi',   label: 'Vietnamese',  native: 'Tiếng Việt' },
  { code: 'hi',   label: 'Hindi',       native: 'हिन्दी'       },
  { code: 'zh-tw',label: 'Chinese (Traditional)', native: '繁體中文' },
  { code: 'pl',   label: 'Polish',      native: 'Polski'     },
  { code: 'th',   label: 'Thai',        native: 'ไทย' },
  { code: 'he',   label: 'Hebrew',      native: 'עברית',     rtl: true },
  { code: 'bn',   label: 'Bengali',     native: 'বাংলা' },
  { code: 'ur',   label: 'Urdu',        native: 'اردو',      rtl: true },
];

export const defaultLocale = 'en';

// ─── English dictionary — eternal fallback, always resident ──────────────
const en = {
  "meta.title": "graffiti — turn any repo into a queryable code graph for AI",
  "meta.description": "graffiti turns a code repository into a directed knowledge graph your AI coding assistant reads instead of grepping. One static Go binary — zero API keys, $0, fully offline, byte-deterministic. Go, Python, JS, TS, Rust, Java, PHP.",
  "lang.label": "Language",

  "hero.brand": "graffiti · code graph",
  "hero.issue": "Issue 01 · 23 languages · 7 parsed",
  "hero.nameplate.sup": "A code-graph engine for AI coding assistants",
  "hero.pitch.badge": "★ Zero API keys · $0 · fully offline",
  "hero.pitch.title": "Stop grepping.<br/>Start reading the <em>graph</em>.",
  "hero.pitch.body": "Your AI assistant explores your codebase blind — one <code class=\"mono\">grep</code> at a time. <strong>graffiti</strong> turns the whole repo into a directed knowledge graph it can query: who calls what, what lives where, which modules matter. One static binary. No API key, no cloud, no cost.",
  "hero.cta.install": "Install in 30s",
  "hero.cta.graph": "See the graph",
  "hero.cta.github": "View on GitHub",
  "hero.meta.github": "GitHub",
  "hero.graph.label": "Live preview · a repo as a force-directed graph",
  "hero.terminal.label": "one command:",
  "hero.legend.client": "your code",
  "hero.legend.tests": "tests",
  "hero.legend.external": "external",
  "hero.det.eyebrow": "Determinism",
  "hero.det.title": "Same repo in — byte-identical map out.",
  "hero.det.body": "Every run sorts everything and stamps a single timestamp. <strong>Diff two builds and only the clock moves.</strong> Commit the map; review it like code.",

  "hero.ben.eyebrow": "Cheaper · faster · more effective",
  "hero.ben.tokens.lead": "Up to 50% fewer tokens",
  "hero.ben.tokens.desc": "Hand the model a scoped subgraph instead of whole files — smaller, cheaper calls.",
  "hero.ben.fast.lead": "Fewer round-trips",
  "hero.ben.fast.desc": "One graph query instead of a dozen greps — answers land faster.",
  "hero.ben.smart.lead": "More accurate answers",
  "hero.ben.smart.desc": "The assistant reads real structure — calls, defs, imports — not guesses.",

  "vibe.eyebrow": "Vibe-chill install",
  "vibe.title": "Why touch a terminal<br/>when you have <em>Claude Code</em>?",
  "vibe.intro": "Paste this one prompt into your Claude Code session. Say \"y\" when it asks for permission. Done.",
  "vibe.bonus": "No Go toolchain, no build flags, no config hunt — Claude downloads the binary, maps your repo, and wires up the integration, asking before each command.",
  "vibe.panel.label": "paste in Claude Code:",
  "vibe.prompt": "Install graffiti by amazopic for me. Download the right static binary for my OS/arch from the latest release at github.com/amazopic/graffiti (or build it from source with `make build` if Go is available), put it on my PATH as `graffiti`, and verify with `graffiti version`. Then run `graffiti .` at my repo root to build the map, run `graffiti init --hook` to wire graffiti into Claude Code (skill + CLAUDE.md + the grep→query nudge), and finally open `.graffiti/map.html` so I can see the graph. Ask before each step.",
  "vibe.note": "↳ Just say <code class=\"mono\">y</code> (yes) at every permission prompt — Claude will run each command one by one.",

  "note.label": "Editor's Note",
  "note.body.p1": "<span class=\"dropcap\">A</span>n AI coding assistant is only as good as what it can <em>see</em>. Drop it into a large repo and it does what you'd do with no map: it greps, opens a few files, guesses. It never sees the <strong>shape</strong> of the code — which function calls which, where a type is defined, which module is the load-bearing wall.",
  "note.body.p2": "graffiti is the map that should have been there. One command parses the repo with tree-sitter, resolves the edges, clusters the modules, and writes a graph — as JSON for the machine, as Markdown for you, and as a single offline HTML you can actually look at. <strong>No keys. No cloud. No cost.</strong>",
  "note.margin": "<em>Editorial note —</em> graffiti is source-available: read and run it freely on your own repos, but reuse requires the author's permission.",

  "numbers.title": "By the numbers",
  "numbers.langs": "languages parsed",
  "numbers.api": "API calls, ever",
  "numbers.cost": "dollars to run",
  "numbers.binary": "static binary (~10 MB)",
  "numbers.langdocs": "documented languages",
  "numbers.det": "byte-deterministic, offline",

  "map.meta": "Contents · three artifacts",
  "map.title": "One command,<br/><em>three artifacts.</em>",
  "map.intro": "<code class=\"mono\">graffiti .</code> writes everything into <code class=\"mono\">.graffiti/</code> — one for the machine, one for you, one to look at.",
  "map.json.title": "map.json",
  "map.json.desc": "The graph itself — nodes, edges, communities, schema-checked against a published contract. This is what your AI reads and what <code class=\"mono\">query</code> and the MCP server traverse.",
  "map.md.title": "MAP.md",
  "map.md.desc": "A human-readable digest: top modules, the most-connected nodes, and the three most interesting questions your map can answer.",
  "map.html.title": "map.html",
  "map.html.desc": "A single self-contained, offline, interactive force-directed graph. No CDN, no server, no network — just open the file.",
  "map.viewer.label": "map.html · interactive viewer",
  "map.viewer.f1": "2D / 3D toggle — hover lifts a node and its neighbours",
  "map.viewer.f2": "search nodes · click to copy file:line",
  "map.viewer.f3": "sector zones · client / tests / external toggles",
  "map.viewer.f4": "resizable project → directory → file tree",

  "compare.sub": "A side-by-side",
  "compare.title": "vs. grep,<br/>vs. the cloud",
  "compare.col.feature": "Capability",
  "compare.col.grep": "Grep / cloud RAG",
  "compare.col.ours": "graffiti",
  "compare.yes": "yes",
  "compare.no": "no",
  "cmp.f.graph": "Directed graph: calls, defs, imports",
  "cmp.f.offline": "Runs fully offline, no account",
  "cmp.f.cost": "Ongoing cost",
  "cmp.f.query": "LLM-free, token-budgeted retrieval",
  "cmp.f.mcp": "MCP server built in",
  "cmp.f.det": "Byte-deterministic output",
  "cmp.f.viewer": "Interactive offline graph viewer",
  "cmp.f.deps": "What it needs",
  "cmp.v.cost.grep": "$ per token",
  "cmp.v.cost.ours": "$0",
  "cmp.v.deps.grep": "cloud + API keys",
  "cmp.v.deps.ours": "one binary",

  "install.sub": "Get going",
  "install.title": "Install in 30 seconds",
  "install.intro": "One static binary does it all: build the map, query it, serve it over MCP, and wire itself into Claude Code.",
  "install.step1.title": "Get the binary",
  "install.step1.desc": "One <code class=\"mono\">curl</code> — or <code class=\"mono\">make build</code> from source.",
  "install.step2.title": "Map your repo",
  "install.step2.desc": "Run <code class=\"mono\">graffiti .</code> at the repo root.",
  "install.step3.title": "Wire up Claude Code",
  "install.step3.desc": "<code class=\"mono\">graffiti init --hook</code> — and your assistant reads the graph.",
  "install.claude.label": "Or in Claude Code — easiest",
  "install.claude.intro": "Have Claude Code open? Paste this one prompt and you're done — no terminal needed.",
  "install.claude.prompt1": "Install graffiti by amazopic for me. Download the right static binary for my OS/arch from the latest release at github.com/amazopic/graffiti, put it on my PATH as `graffiti`, and verify with `graffiti version`. Then run `graffiti .` to build the map, `graffiti init --hook` to wire it into Claude Code, and open `.graffiti/map.html`. Ask before each step.",
  "install.codelabel.main": "macOS / Linux / WSL",
  "install.codelabel.cc": "Claude Code integration",
  "install.note": "The installer picks the right static binary for your OS/arch, verifies its SHA256 against the signed release manifest, and installs it. Prefer source? <code class=\"mono\">make build</code> produces the same ~10 MB CGO-free binary.",

  "ws.eyebrow": "Multi-repo",
  "ws.title": "Frontend and backend,<br/><em>one graph.</em>",
  "ws.body": "Lay separate repos side by side and query across them — without merging. <code class=\"mono\">graffiti link ../frontend ../backend</code> federates them into a workspace; each repo keeps its own standalone map.",
  "ws.note": "Assert cross-project edges explicitly, validate they resolve, then <code class=\"mono\">graffiti query --workspace</code> traverses the whole federation — or render <code class=\"mono\">workspace.html</code> with the projects as the top level of the tree.",

  "faq.sub": "Q &amp; A",
  "faq.title": "Frequently<br/>Asked",
  "faq.q.what": "What is graffiti?",
  "faq.a.what": "A single static command-line binary that turns a code repository into a directed knowledge graph — nodes for files, functions, types and modules; edges for calls, definitions and imports. It writes the graph as <code>map.json</code>, a human digest <code>MAP.md</code>, and an interactive offline <code>map.html</code>, so an AI coding assistant can read structure instead of grepping blind.",
  "faq.q.how": "How does it build the graph without an LLM?",
  "faq.a.how": "It parses each file with <a href=\"https://tree-sitter.github.io/tree-sitter/\">tree-sitter</a> grammars (pure-Go, no CGO), extracts definitions and references, resolves the edges, clusters the result into modules, and serializes it. No model, no embeddings, no network — just static analysis. That's why it's free and deterministic.",
  "faq.q.langs": "Which languages does it support?",
  "faq.a.langs": "Go, Python, JavaScript, TypeScript, Rust, Java, and PHP (plus Markdown doc nodes). Go gets full call resolution; the others capture files, functions, classes/structs/interfaces, methods and imports, and intentionally <strong>under-extract</strong> exotic constructs rather than emit guesses.",
  "faq.q.offline": "Does it really need no API key or network?",
  "faq.a.offline": "Correct — zero API calls, $0, fully offline. Everything runs locally in one binary. Nothing about your code ever leaves your machine.",
  "faq.q.det": "What does \"byte-deterministic\" mean here?",
  "faq.a.det": "The same repository always produces a byte-identical <code>map.json</code> — modulo a single <code>generated_at</code> timestamp and the root folder name. Everything is sorted. You can commit the map and review changes to it in a diff like any other file.",
  "faq.q.cc": "How does the Claude Code integration work?",
  "faq.a.cc": "<code>graffiti init</code> installs a skill plus a <code>CLAUDE.md</code> block telling the assistant to prefer <code>graffiti query</code> over grep when a map exists. With <code>--hook</code> it also adds a PreToolUse nudge before <code>Grep</code>/<code>Glob</code>. It's idempotent and never blocks a tool.",
  "faq.q.mcp": "Can my assistant query it over MCP?",
  "faq.a.mcp": "Yes. <code>graffiti serve</code> exposes the map over an MCP stdio server (JSON-RPC 2.0), so any MCP-capable client can traverse the graph through tools. There's also an LLM-free <code>graffiti query \"…\"</code> that returns a scoped subgraph within a soft token budget.",
  "faq.q.viewer": "What is map.html?",
  "faq.a.viewer": "A single self-contained HTML file — an interactive force-directed graph rendered in the browser with a 2D/3D toggle, node search, click-to-copy <code>file:line</code>, sector zones, client/tests/external toggles, and a resizable project→directory→file tree. No CDN and no server: it's CSP-safe and works offline. Just open it.",
  "faq.q.big": "Will it handle a big repository?",
  "faq.a.big": "Yes. Parsing is fast static analysis and a per-file content-hash cache means re-runs only re-parse what changed. The graph and viewer stay responsive on large codebases.",
  "faq.q.free": "Is it free? Can I use it commercially?",
  "faq.a.free": "Building and running graffiti on your own repositories is free under the <a href=\"https://github.com/amazopic/graffiti/blob/main/LICENSE\">Source-Available License</a>. Any reuse, redistribution, fork, or inclusion in another project requires <strong>prior written permission</strong> from the author. Reasonable requests for personal, educational, and non-commercial use are typically granted.",

  "colo.title": "Graffiti",
  "colo.h.author": "Author",
  "colo.h.license": "License",
  "colo.h.set": "Set in",
  "colo.h.links": "Links",
  "colo.license.body": "Source-Available — reuse only with prior written permission.",
  "colo.license.read": "Read full license →",
  "colo.links.repo": "GitHub repo ↗",
  "colo.links.readme": "Read README ↗",
  "colo.links.viewer": "Viewer source",
  "colo.links.cc": "Claude Code ↗",
  "colo.meta.copyright": "© 2026 Yevgeniy Achin · Source-Available",
  "colo.meta.made": "Made for AI coding assistants",
  "colo.meta.issue": "Issue 01 · v0.1.0",

  "ui.copy": "Copy",
  "ui.copied": "✓ copied to clipboard",
  "ui.copyfail": "✗ copy failed",
};

// Locale registry. 'en' is always present; others fill in lazily.
const registry = { en };

// Lazy-load a locale chunk. Idempotent; safe to await repeatedly.
// On failure, registers an empty dict so callers transparently fall back to en.
export async function ensureLocale(code) {
  if (registry[code]) return registry[code];
  try {
    registry[code] = (await import(`./locales/${code}.js?v=${ASSET_V}`)).default;
  } catch (e) {
    console.warn('i18n: failed to load locale', code, e);
    registry[code] = {};
  }
  return registry[code];
}

// ─── Helpers ───────────────────────────────────────────────────────────

export function t(key, locale = defaultLocale) {
  return (registry[locale] && registry[locale][key]) ?? en[key] ?? key;
}

export function detectLocale() {
  // English is the default. The user explicitly opts into another language
  // via the switcher or via ?lang=xx — we do NOT auto-detect from the browser.
  const params = new URLSearchParams(location.search);
  const q = params.get('lang');
  if (q && supportedLocales.some(l => l.code === q)) return q;
  try {
    const saved = localStorage.getItem('lang');
    if (saved && supportedLocales.some(l => l.code === saved)) return saved;
  } catch (_) { /* private mode etc. */ }
  return defaultLocale;
}

export function persistLocale(code) {
  try { localStorage.setItem('lang', code); } catch (_) {}
}

export const enDict = en;
