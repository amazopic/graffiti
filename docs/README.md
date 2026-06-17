# graffiti — landing site

This folder is the source of the project's landing page, published with
**GitHub Pages** (Settings → Pages → *Deploy from a branch* → `main` / `/docs`).

Live: **https://amazopic.github.io/graffiti/**

## Layout

```
docs/
├── index.html          # the page (semantic HTML + JSON-LD SEO + hreflang)
├── css/
│   ├── tokens.css       # design tokens (color, type, spacing, motion)
│   ├── base.css         # reset + editorial primitives + grain + reveal
│   └── sections.css     # per-section styling
├── js/
│   ├── i18n.js          # locale loader + canonical English dictionary
│   ├── main.js          # i18n apply, language switcher, copy, reveal, force-graph canvas
│   └── locales/<code>.js # one flat dictionary per language (lazy-loaded)
├── favicon.svg
├── og-image.svg         # social card (1200×630)
├── robots.txt           # AI/LLM crawler allowlist + sitemap
├── sitemap.xml          # with hreflang alternates
├── llms.txt             # AI crawler index
├── llms-full.txt        # AI ingestion corpus
└── .nojekyll            # serve files verbatim (skip Jekyll)
```

## How it works

- **No build step.** Plain HTML/CSS/ES-modules — open `index.html` or serve the
  folder statically.
- **i18n** is client-side: `js/i18n.js` holds the English dictionary (the
  fallback that always ships) and lazy-loads `js/locales/<code>.js` when a reader
  picks another language. Strings are bound to the DOM via `data-i18n`
  (textContent), `data-i18n-html` (innerHTML), and `data-i18n-aria-label`.
- **The hero graph** is a tiny self-contained force-directed simulation drawn on
  a `<canvas>` — the same idea as graffiti's real `map.html` viewer, no deps.

## Editing

1. Change copy → edit the English entries in `js/i18n.js`, then mirror the key in
   each `js/locales/<code>.js`.
2. Bump the `?v=` query string on the asset links in `index.html` (and `ASSET_V`
   in `i18n.js`) to bust caches.
3. Add a language → add it to `supportedLocales` in `js/i18n.js`, drop a
   `js/locales/<code>.js`, and add the `hreflang`/sitemap entries.
