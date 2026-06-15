"use strict";
(function () {
  // ---- 1. Load + rebuild the scene from the columnar data island. ----
  function loadData() {
    var el = document.getElementById("graffiti-data");
    if (!el) return null;
    try { return JSON.parse(el.textContent); } catch (e) { return null; }
  }
  function str(d, i) { return (d.strings && d.strings[i]) || ""; }
  function pts(flat, off, i) {
    var a = off[i] * 2, b = off[i + 1] * 2, out = [];
    for (var k = a; k < b; k += 2) out.push([flat[k], flat[k + 1]]);
    return out;
  }
  function rebuild(d) {
    var i, boxes = [], pins = [], bundles = [], arcs = [];
    for (i = 0; i < d.boxComm.length; i++) {
      boxes.push({ comm: d.boxComm[i], label: str(d, d.boxLabel[i]), count: d.boxCount[i],
        x: d.boxX[i], y: d.boxY[i], w: d.boxW[i], h: d.boxH[i], border: d.boxBorder[i] });
    }
    for (i = 0; i < d.pinComm.length; i++) {
      pins.push({ node: str(d, d.pinNode[i]), label: str(d, d.pinLabel[i]),
        comm: d.pinComm[i], x: d.pinX[i], y: d.pinY[i] });
    }
    for (i = 0; i < d.bundleFrom.length; i++) {
      bundles.push({ from: d.bundleFrom[i], to: d.bundleTo[i], count: d.bundleCount[i],
        pts: pts(d.bundlePts, d.bundleOff, i) });
    }
    for (i = 0; i < d.arcFrom.length; i++) {
      arcs.push({ from: d.arcFrom[i], to: d.arcTo[i], conf: str(d, d.arcConf[i]),
        pts: pts(d.arcPts, d.arcOff, i) });
    }
    return { w: d.w, h: d.h, boxes: boxes, pins: pins, bundles: bundles, arcs: arcs };
  }

  var data = loadData();
  if (!data) return;
  var scene = rebuild(data);

  // ---- 2. Canvas + HiDPI + camera (pan/zoom). No layout, no physics. ----
  var canvas = document.getElementById("canvas");
  var stage = document.getElementById("stage");
  if (!canvas || !stage) return;
  var ctx = canvas.getContext("2d");
  var dpr = window.devicePixelRatio || 1;
  var cam = { x: 0, y: 0, scale: 1 }; // world->screen: screen = (world - cam)/?, see toScreen
  var dirty = true;

  function fit() {
    var vw = stage.clientWidth, vh = stage.clientHeight;
    canvas.width = Math.round(vw * dpr);
    canvas.height = Math.round(vh * dpr);
    var s = Math.min(vw / scene.w, vh / scene.h);
    cam.scale = s;
    cam.x = (scene.w * s - vw) / 2 / s;
    cam.y = (scene.h * s - vh) / 2 / s;
    dirty = true;
  }
  function toScreen(wx, wy) { return [(wx - cam.x) * cam.scale, (wy - cam.y) * cam.scale]; }
  function toWorld(sx, sy) { return [sx / cam.scale + cam.x, sy / cam.scale + cam.y]; }

  // colorblind-safe confidence stroke styles (line style + alpha) per spec §8.5.
  function strokeForConf(conf) {
    if (conf === "AMBIGUOUS") return { dash: [6, 5], alpha: 0.55, width: 1 };
    if (conf === "INFERRED") return { dash: [], alpha: 0.55, width: 1 };
    return { dash: [], alpha: 1, width: 2 }; // EXTRACTED
  }

  var hot = null; // currently inspected box

  function draw() {
    var vw = stage.clientWidth, vh = stage.clientHeight;
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx.clearRect(0, 0, vw, vh);

    // 2a. Bundled flow-arrows (batched stroke; thickness = bundled count).
    ctx.strokeStyle = "#8893a5";
    var i, j, p;
    for (i = 0; i < scene.bundles.length; i++) {
      var bn = scene.bundles[i];
      ctx.lineWidth = Math.max(1, Math.min(8, Math.log2(bn.count + 1) * 2));
      ctx.beginPath();
      for (j = 0; j < bn.pts.length; j++) {
        p = toScreen(bn.pts[j][0], bn.pts[j][1]);
        if (j === 0) ctx.moveTo(p[0], p[1]); else ctx.lineTo(p[0], p[1]);
      }
      ctx.stroke();
    }

    // 2b. Surprising arcs (dashed, brightly-tinted).
    for (i = 0; i < scene.arcs.length; i++) {
      var ar = scene.arcs[i], st = strokeForConf(ar.conf);
      ctx.save();
      ctx.globalAlpha = st.alpha;
      ctx.strokeStyle = "#c0392b";
      ctx.lineWidth = st.width;
      ctx.setLineDash([6, 5]);
      ctx.beginPath();
      for (j = 0; j < ar.pts.length; j++) {
        p = toScreen(ar.pts[j][0], ar.pts[j][1]);
        if (j === 0) ctx.moveTo(p[0], p[1]); else ctx.lineTo(p[0], p[1]);
      }
      ctx.stroke();
      ctx.restore();
    }

    // 2c. District boxes (area = size, border weight = centrality).
    ctx.textBaseline = "top";
    ctx.font = "12px system-ui,sans-serif";
    for (i = 0; i < scene.boxes.length; i++) {
      var b = scene.boxes[i];
      var tl = toScreen(b.x, b.y);
      var w = b.w * cam.scale, h = b.h * cam.scale;
      ctx.fillStyle = (hot && hot.comm === b.comm) ? "#dbe7fb" : "#eef2f8";
      ctx.fillRect(tl[0], tl[1], w, h);
      ctx.strokeStyle = "#5b6b86";
      ctx.lineWidth = b.border;
      ctx.setLineDash([]);
      ctx.strokeRect(tl[0], tl[1], w, h);
      if (w > 46 && h > 22) {
        ctx.fillStyle = "#1c2330";
        ctx.fillText(b.label + " (" + b.count + ")", tl[0] + 6, tl[1] + 5);
      }
    }

    // 2d. God-node landmark pins (starred halo).
    for (i = 0; i < scene.pins.length; i++) {
      var pn = scene.pins[i], c = toScreen(pn.x, pn.y);
      ctx.fillStyle = "rgba(226,179,64,.35)";
      ctx.beginPath(); ctx.arc(c[0], c[1], 9, 0, Math.PI * 2); ctx.fill();
      ctx.fillStyle = "#d99a13";
      ctx.beginPath(); ctx.arc(c[0], c[1], 5, 0, Math.PI * 2); ctx.fill();
    }
    dirty = false;
  }

  function frame() { if (dirty) draw(); requestAnimationFrame(frame); }

  // ---- 3. Interaction: pan, zoom, click-to-inspect. ----
  var dragging = false, lastX = 0, lastY = 0;
  canvas.addEventListener("mousedown", function (e) {
    dragging = true; lastX = e.clientX; lastY = e.clientY; canvas.classList.add("drag");
  });
  window.addEventListener("mouseup", function () { dragging = false; canvas.classList.remove("drag"); });
  window.addEventListener("mousemove", function (e) {
    if (!dragging) return;
    cam.x -= (e.clientX - lastX) / cam.scale;
    cam.y -= (e.clientY - lastY) / cam.scale;
    lastX = e.clientX; lastY = e.clientY; dirty = true;
  });
  canvas.addEventListener("wheel", function (e) {
    e.preventDefault();
    var r = canvas.getBoundingClientRect();
    var sx = e.clientX - r.left, sy = e.clientY - r.top;
    var before = toWorld(sx, sy);
    var f = e.deltaY < 0 ? 1.1 : 1 / 1.1;
    cam.scale = Math.max(0.2, Math.min(8, cam.scale * f));
    var after = toWorld(sx, sy);
    cam.x += before[0] - after[0]; cam.y += before[1] - after[1];
    dirty = true;
  }, { passive: false });

  function boxAt(sx, sy) {
    var w = toWorld(sx, sy);
    for (var i = 0; i < scene.boxes.length; i++) {
      var b = scene.boxes[i];
      if (w[0] >= b.x && w[0] <= b.x + b.w && w[1] >= b.y && w[1] <= b.y + b.h) return b;
    }
    return null;
  }
  var inspect = document.getElementById("inspect");
  function showInspect(b) {
    hot = b; dirty = true;
    if (!inspect) return;
    if (!b) { inspect.className = ""; return; }
    inspect.textContent = "";
    var h = document.createElement("h3"); h.textContent = b.label; inspect.appendChild(h);
    var m = document.createElement("div"); m.className = "meta";
    m.textContent = b.count + " things · centrality " + b.border; inspect.appendChild(m);
    inspect.className = "show";
  }
  canvas.addEventListener("click", function (e) {
    var r = canvas.getBoundingClientRect();
    showInspect(boxAt(e.clientX - r.left, e.clientY - r.top));
  });

  // ---- 4. Left rail: search filter + landmark fly-to. ----
  var search = document.getElementById("search");
  if (search) {
    search.addEventListener("input", function () {
      var q = search.value.toLowerCase();
      var hit = null;
      for (var i = 0; i < scene.boxes.length; i++) {
        if (q && scene.boxes[i].label.toLowerCase().indexOf(q) >= 0) { hit = scene.boxes[i]; break; }
      }
      if (hit) flyTo(hit);
    });
  }
  function flyTo(b) {
    var vw = stage.clientWidth, vh = stage.clientHeight;
    cam.scale = Math.max(0.5, Math.min(4, Math.min(vw / (b.w * 2), vh / (b.h * 2))));
    cam.x = b.x + b.w / 2 - vw / 2 / cam.scale;
    cam.y = b.y + b.h / 2 - vh / 2 / cam.scale;
    showInspect(b);
  }
  // landmark / district chips emitted by the Go side carry data-comm.
  var chips = document.querySelectorAll("[data-comm]");
  for (var ci = 0; ci < chips.length; ci++) {
    (function (el) {
      el.addEventListener("click", function () {
        var cid = parseInt(el.getAttribute("data-comm"), 10);
        for (var i = 0; i < scene.boxes.length; i++) {
          if (scene.boxes[i].comm === cid) { flyTo(scene.boxes[i]); break; }
        }
      });
    })(chips[ci]);
  }

  window.addEventListener("resize", fit);
  fit();
  requestAnimationFrame(frame);
})();
