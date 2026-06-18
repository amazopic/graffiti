// ─────────────────────────────────────────────────────────────────────
// graffiti landing — interaction layer
//   · client-side i18n (23 locales, lazy-loaded)
//   · language switcher (persisted, ?lang=, RTL-aware)
//   · copy-to-clipboard buttons + toast
//   · scroll reveal
//   · a live force-directed graph in the hero (the product's signature)
// ─────────────────────────────────────────────────────────────────────

import {
  supportedLocales, ensureLocale, t, detectLocale, persistLocale, defaultLocale,
} from './i18n.js?v=3';

// ─── i18n application ────────────────────────────────────────────────
let current = defaultLocale;

function applyI18n(locale) {
  document.querySelectorAll('[data-i18n]').forEach(el => {
    el.textContent = t(el.getAttribute('data-i18n'), locale);
  });
  document.querySelectorAll('[data-i18n-html]').forEach(el => {
    el.innerHTML = t(el.getAttribute('data-i18n-html'), locale);
  });
  document.querySelectorAll('[data-i18n-aria-label]').forEach(el => {
    el.setAttribute('aria-label', t(el.getAttribute('data-i18n-aria-label'), locale));
  });
  document.title = t('meta.title', locale);
  const desc = document.querySelector('meta[name="description"]');
  if (desc) desc.setAttribute('content', t('meta.description', locale));
}

async function setLocale(code) {
  if (code !== 'en') await ensureLocale(code);
  current = code;
  const meta = supportedLocales.find(l => l.code === code) || {};
  document.documentElement.lang = code;
  document.documentElement.dir = meta.rtl ? 'rtl' : 'ltr';
  applyI18n(code);
  const cur = document.querySelector('[data-current-lang]');
  if (cur) cur.textContent = (meta.code || 'en').toUpperCase();
  persistLocale(code);
}

// ─── Language switcher ───────────────────────────────────────────────
function initLangSwitcher() {
  const root = document.querySelector('[data-lang-switcher]');
  if (!root) return;
  const toggle = root.querySelector('.lang-switcher__toggle');
  const menu = root.querySelector('[data-lang-menu]');

  menu.innerHTML = supportedLocales.map(l =>
    `<li role="option" data-code="${l.code}" tabindex="0">
       <span class="lang-switcher__native">${l.native}</span>
       <span class="lang-switcher__en">${l.label}</span>
     </li>`).join('');

  const close = () => { menu.hidden = true; toggle.setAttribute('aria-expanded', 'false'); };
  const open  = () => { menu.hidden = false; toggle.setAttribute('aria-expanded', 'true'); };

  toggle.addEventListener('click', e => {
    e.stopPropagation();
    menu.hidden ? open() : close();
  });
  menu.addEventListener('click', e => {
    const li = e.target.closest('li[data-code]');
    if (!li) return;
    setLocale(li.dataset.code);
    close();
  });
  menu.addEventListener('keydown', e => {
    if (e.key === 'Enter' && e.target.dataset.code) { setLocale(e.target.dataset.code); close(); toggle.focus(); }
  });
  document.addEventListener('click', () => { if (!menu.hidden) close(); });
  document.addEventListener('keydown', e => { if (e.key === 'Escape') close(); });
}

// ─── Copy buttons + toast ────────────────────────────────────────────
function toast(msg) {
  const el = document.querySelector('.toast');
  if (!el) return;
  el.textContent = msg;
  el.classList.add('show');
  clearTimeout(el._t);
  el._t = setTimeout(() => el.classList.remove('show'), 1800);
}

async function copyText(text) {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch (_) {
    try {
      const ta = document.createElement('textarea');
      ta.value = text; ta.style.position = 'fixed'; ta.style.opacity = '0';
      document.body.appendChild(ta); ta.select();
      const ok = document.execCommand('copy');
      document.body.removeChild(ta);
      return ok;
    } catch (_) { return false; }
  }
}

function initCopy() {
  document.querySelectorAll('[data-copy]').forEach(btn => {
    btn.addEventListener('click', async () => {
      const target = document.querySelector(btn.getAttribute('data-copy'));
      if (!target) return;
      const ok = await copyText(target.innerText.trim());
      toast(t(ok ? 'ui.copied' : 'ui.copyfail', current));
    });
  });
}

// ─── Scroll reveal ───────────────────────────────────────────────────
function initReveal() {
  const els = document.querySelectorAll('[data-io]');
  if (!('IntersectionObserver' in window) || !els.length) {
    els.forEach(el => el.classList.add('in'));
    return;
  }
  const io = new IntersectionObserver((entries) => {
    entries.forEach(en => {
      if (en.isIntersecting) { en.target.classList.add('in'); io.unobserve(en.target); }
    });
  }, { rootMargin: '0px 0px -10% 0px' });
  els.forEach(el => io.observe(el));
}

// ─── Force-directed graph (hero showpiece) ───────────────────────────
// A tiny self-contained physics sim + pseudo-3D sphere rendering, echoing
// graffiti's actual map.html viewer. No deps, no WebGL.
function mulberry32(a) {
  return function () {
    a |= 0; a = (a + 0x6D2B79F5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

function buildDemoGraph(rand) {
  // 3 sectors of "your code" + a tests cluster + external libs.
  const sectors = [
    { n: 7, cat: 0 }, { n: 6, cat: 0 }, { n: 5, cat: 0 },
    { n: 5, cat: 1 }, { n: 5, cat: 2 },
  ];
  const nodes = [], edges = [];
  let id = 0;
  const groups = [];
  for (const s of sectors) {
    const g = [];
    for (let i = 0; i < s.n; i++) {
      nodes.push({
        id: id, cat: s.cat,
        x: (rand() - 0.5) * 320, y: (rand() - 0.5) * 320, vx: 0, vy: 0,
        r: 3 + rand() * 7, deg: 0,
      });
      g.push(id); id++;
    }
    groups.push(g);
    // intra-sector edges (loose tree + a couple extra)
    for (let i = 1; i < g.length; i++) {
      edges.push([g[i], g[Math.floor(rand() * i)]]);
    }
    if (g.length > 3) edges.push([g[1], g[g.length - 1]]);
  }
  // a few cross-sector edges, incl. into tests + external
  edges.push([groups[0][0], groups[1][0]]);
  edges.push([groups[1][0], groups[2][0]]);
  edges.push([groups[3][0], groups[0][2]]); // tests -> code
  edges.push([groups[3][1], groups[1][2]]);
  edges.push([groups[0][1], groups[4][0]]); // code -> external
  edges.push([groups[2][1], groups[4][1]]);
  edges.push([groups[1][2], groups[4][2]]);
  for (const [a, b] of edges) { nodes[a].deg++; nodes[b].deg++; }
  nodes.forEach(n => { n.r = 3.5 + Math.min(n.deg, 7) * 1.6; });
  return { nodes, edges };
}

const CAT_COLORS = [
  { core: '#2a2f6b', glow: '#5e6ad2' }, // your code → lavender (brand accent)
  { core: '#14431f', glow: '#27a644' }, // tests → success green
  { core: '#15406e', glow: '#4ea7ff' }, // external → blue
];

function spriteFor(cache, cat, r, dpr) {
  const key = cat + ':' + Math.round(r) + ':' + dpr;
  if (cache[key]) return cache[key];
  const pad = 6;
  const size = Math.ceil((r + pad) * 2 * dpr);
  const c = document.createElement('canvas');
  c.width = c.height = size;
  const g = c.getContext('2d');
  const cx = size / 2, R = r * dpr;
  const col = CAT_COLORS[cat] || CAT_COLORS[0];
  // soft glow
  const glow = g.createRadialGradient(cx, cx, R * 0.4, cx, cx, R + pad * dpr);
  glow.addColorStop(0, col.glow + '88');
  glow.addColorStop(1, col.glow + '00');
  g.fillStyle = glow;
  g.beginPath(); g.arc(cx, cx, R + pad * dpr, 0, Math.PI * 2); g.fill();
  // sphere body with a light from top-left → pseudo-3D
  const body = g.createRadialGradient(cx - R * 0.35, cx - R * 0.35, R * 0.1, cx, cx, R);
  body.addColorStop(0, col.glow);
  body.addColorStop(0.55, col.core);
  body.addColorStop(1, '#06060a');
  g.fillStyle = body;
  g.beginPath(); g.arc(cx, cx, R, 0, Math.PI * 2); g.fill();
  cache[key] = { canvas: c, half: size / 2 / dpr };
  return cache[key];
}

function initGraph() {
  const canvas = document.getElementById('graph');
  if (!canvas) return;
  const ctx = canvas.getContext('2d');
  const reduce = matchMedia('(prefers-reduced-motion: reduce)').matches;
  const rand = mulberry32(0x9e3779b9);
  const { nodes, edges } = buildDemoGraph(rand);
  const sprites = {};
  let dpr = 1, W = 0, H = 0, hover = -1, raf = 0;
  const neighbors = nodes.map(() => new Set());
  edges.forEach(([a, b]) => { neighbors[a].add(b); neighbors[b].add(a); });

  function resize() {
    const rect = canvas.getBoundingClientRect();
    dpr = Math.min(window.devicePixelRatio || 1, 2);
    W = rect.width; H = rect.height;
    canvas.width = Math.max(1, Math.round(W * dpr));
    canvas.height = Math.max(1, Math.round(H * dpr));
  }

  function step() {
    // physics: repulsion + spring + gravity, in graph space centered at 0
    for (let i = 0; i < nodes.length; i++) {
      const a = nodes[i];
      for (let j = i + 1; j < nodes.length; j++) {
        const b = nodes[j];
        let dx = a.x - b.x, dy = a.y - b.y;
        let d2 = dx * dx + dy * dy + 0.01;
        const f = 900 / d2;
        const d = Math.sqrt(d2);
        const ux = dx / d, uy = dy / d;
        a.vx += ux * f; a.vy += uy * f;
        b.vx -= ux * f; b.vy -= uy * f;
      }
    }
    for (const [ai, bi] of edges) {
      const a = nodes[ai], b = nodes[bi];
      const dx = b.x - a.x, dy = b.y - a.y;
      const d = Math.sqrt(dx * dx + dy * dy) || 0.01;
      const f = (d - 64) * 0.012;
      const ux = dx / d, uy = dy / d;
      a.vx += ux * f; a.vy += uy * f;
      b.vx -= ux * f; b.vy -= uy * f;
    }
    for (const n of nodes) {
      n.vx -= n.x * 0.0016; n.vy -= n.y * 0.0016; // gravity to center
      n.vx *= 0.86; n.vy *= 0.86;                 // damping
      n.x += n.vx; n.y += n.vy;
    }
  }

  function fitTransform() {
    let minX = 1e9, minY = 1e9, maxX = -1e9, maxY = -1e9;
    for (const n of nodes) {
      if (n.x < minX) minX = n.x; if (n.x > maxX) maxX = n.x;
      if (n.y < minY) minY = n.y; if (n.y > maxY) maxY = n.y;
    }
    const gw = (maxX - minX) || 1, gh = (maxY - minY) || 1;
    const pad = 36;
    const scale = Math.min((W - pad * 2) / gw, (H - pad * 2) / gh, 2.2);
    const cx = (minX + maxX) / 2, cy = (minY + maxY) / 2;
    return { scale, cx, cy };
  }

  function draw() {
    const { scale, cx, cy } = fitTransform();
    const toX = x => W / 2 + (x - cx) * scale;
    const toY = y => H / 2 + (y - cy) * scale;
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx.clearRect(0, 0, W, H);

    // edges
    ctx.lineWidth = 1;
    for (const [ai, bi] of edges) {
      const a = nodes[ai], b = nodes[bi];
      const lit = hover >= 0 && (ai === hover || bi === hover);
      ctx.strokeStyle = lit ? 'rgba(130,143,255,0.7)' : 'rgba(255,255,255,0.10)';
      ctx.beginPath();
      ctx.moveTo(toX(a.x), toY(a.y));
      ctx.lineTo(toX(b.x), toY(b.y));
      ctx.stroke();
    }

    // nodes (sorted by y for a faint depth order)
    const order = nodes.map((_, i) => i).sort((p, q) => nodes[p].y - nodes[q].y);
    for (const i of order) {
      const n = nodes[i];
      const dim = hover >= 0 && i !== hover && !neighbors[hover].has(i);
      const lift = (i === hover) ? 1.35 : (hover >= 0 && neighbors[hover].has(i) ? 1.12 : 1);
      const r = n.r * scale * lift;
      const sp = spriteFor(sprites, n.cat, r, dpr);
      ctx.globalAlpha = dim ? 0.28 : 1;
      ctx.drawImage(sp.canvas, toX(n.x) - sp.half, toY(n.y) - sp.half, sp.half * 2, sp.half * 2);
    }
    ctx.globalAlpha = 1;
  }

  let warm = reduce ? 220 : 0;
  function frame() {
    if (!reduce) step();
    draw();
    if (reduce) return; // static after one settle below
    raf = requestAnimationFrame(frame);
  }

  // hover hit-test
  canvas.addEventListener('pointermove', e => {
    const rect = canvas.getBoundingClientRect();
    const mx = e.clientX - rect.left, my = e.clientY - rect.top;
    const { scale, cx, cy } = fitTransform();
    let best = -1, bestD = 1e9;
    for (let i = 0; i < nodes.length; i++) {
      const px = W / 2 + (nodes[i].x - cx) * scale;
      const py = H / 2 + (nodes[i].y - cy) * scale;
      const d = (px - mx) ** 2 + (py - my) ** 2;
      const rr = (nodes[i].r * scale + 6) ** 2;
      if (d < rr && d < bestD) { bestD = d; best = i; }
    }
    if (best !== hover) { hover = best; canvas.style.cursor = best >= 0 ? 'pointer' : 'default'; }
  });
  canvas.addEventListener('pointerleave', () => { hover = -1; });

  resize();
  window.addEventListener('resize', () => { resize(); if (reduce) { for (let k = 0; k < 200; k++) step(); draw(); } });
  if ('ResizeObserver' in window) new ResizeObserver(() => resize()).observe(canvas);

  if (reduce) { for (let k = 0; k < warm; k++) step(); draw(); }
  else frame();
}

// ─── Boot ────────────────────────────────────────────────────────────
function boot() {
  initLangSwitcher();
  initCopy();
  initReveal();
  initGraph();
  setLocale(detectLocale());
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', boot);
} else {
  boot();
}
