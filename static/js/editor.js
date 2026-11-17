// Interactive canvas: zoom/pan, selection, drag, resize, snapping and the
// Go-render comparison overlays.

import { drawFace } from './render.js';
import { canRotate, isBoxLike, isSquare } from './model.js';

const HANDLE_PX = 8;
const SNAP_PX = 6;
const round = (v) => Math.round(v * 100) / 100;

export class Stage {
  constructor(canvas, app) {
    this.canvas = canvas;
    this.ctx = canvas.getContext('2d');
    this.app = app;
    this.k = 0.1;
    this.panX = 0;
    this.panY = 0;
    this.drag = null;
    this.hoverId = null;
    this.guides = [];
    this.spaceDown = false;
    this.needsFit = true;
    this.raf = 0;

    new ResizeObserver(() => this.resize()).observe(canvas.parentElement);
    canvas.addEventListener('pointerdown', (e) => this.onDown(e));
    canvas.addEventListener('pointermove', (e) => this.onMove(e));
    canvas.addEventListener('pointerup', (e) => this.onUp(e));
    canvas.addEventListener('pointercancel', (e) => this.onUp(e));
    canvas.addEventListener('pointerleave', () => {
      this.hoverId = null;
      this.app.setPointer(null);
      this.invalidate();
    });
    canvas.addEventListener('wheel', (e) => this.onWheel(e), { passive: false });
    canvas.addEventListener('dblclick', (e) => this.onDblClick(e));
    window.addEventListener('keydown', (e) => {
      if (e.code === 'Space' && !isTyping(e)) {
        this.spaceDown = true;
        this.canvas.style.cursor = 'grab';
        e.preventDefault();
      }
    });
    window.addEventListener('keyup', (e) => {
      if (e.code === 'Space') {
        this.spaceDown = false;
        this.canvas.style.cursor = '';
      }
    });
  }

  // ---------------------------------------------------------------- view

  resize() {
    const r = this.canvas.parentElement.getBoundingClientRect();
    const dpr = window.devicePixelRatio || 1;
    this.canvas.width = Math.max(1, Math.round(r.width * dpr));
    this.canvas.height = Math.max(1, Math.round(r.height * dpr));
    this.canvas.style.width = r.width + 'px';
    this.canvas.style.height = r.height + 'px';
    if (this.needsFit) this.fit();
    this.invalidate();
  }

  fit() {
    const face = this.app.face();
    const r = this.canvas.parentElement.getBoundingClientRect();
    if (!face || r.width < 10) return;
    const margin = 48;
    this.k = Math.min((r.width - margin * 2) / face.outW, (r.height - margin * 2) / face.outH);
    this.panX = (r.width - face.outW * this.k) / 2;
    this.panY = (r.height - face.outH * this.k) / 2;
    this.needsFit = false;
    this.app.setZoom(this.k);
    this.invalidate();
  }

  zoomAt(factor, sx, sy) {
    const r = this.canvas.parentElement.getBoundingClientRect();
    sx ??= r.width / 2;
    sy ??= r.height / 2;
    const fx = (sx - this.panX) / this.k;
    const fy = (sy - this.panY) / this.k;
    this.k = Math.min(8, Math.max(0.01, this.k * factor));
    this.panX = sx - fx * this.k;
    this.panY = sy - fy * this.k;
    this.app.setZoom(this.k);
    this.invalidate();
  }

  zoomTo(k) {
    this.zoomAt(k / this.k);
  }

  point(e) {
    const r = this.canvas.getBoundingClientRect();
    const sx = e.clientX - r.left;
    const sy = e.clientY - r.top;
    return { x: (sx - this.panX) / this.k, y: (sy - this.panY) / this.k, sx, sy };
  }

  invalidate() {
    if (!this.raf) this.raf = requestAnimationFrame(() => {
      this.raf = 0;
      this.draw();
    });
  }

  // ---------------------------------------------------------------- hit testing

  local(L, p) {
    const deg = canRotate(L.el) ? +L.el.rotation || 0 : 0;
    if (!deg) return p;
    const a = (-deg * Math.PI) / 180;
    const dx = p.x - L.pivot.x;
    const dy = p.y - L.pivot.y;
    return { x: L.pivot.x + dx * Math.cos(a) - dy * Math.sin(a), y: L.pivot.y + dx * Math.sin(a) + dy * Math.cos(a) };
  }

  hit(p) {
    const face = this.app.face();
    const layouts = this.app.layouts();
    const pad = 3 / this.k;
    for (let i = face.elements.length - 1; i >= 0; i--) {
      const el = face.elements[i];
      if (el.hidden || el.locked) continue;
      const L = layouts.get(el.id);
      if (!L) continue;
      const q = this.local(L, p);
      const b = L.box;
      if (q.x >= b.x - pad && q.x <= b.x + b.w + pad && q.y >= b.y - pad && q.y <= b.y + b.h + pad) return el;
    }
    return null;
  }

  handles(L) {
    const el = L.el;
    if (!isBoxLike(el) || el.locked || (canRotate(el) && (+el.rotation || 0) % 360)) return [];
    const { x, y, w, h } = L.box;
    const all = {
      nw: [x, y], n: [x + w / 2, y], ne: [x + w, y], e: [x + w, y + h / 2],
      se: [x + w, y + h], s: [x + w / 2, y + h], sw: [x, y + h], w: [x, y + h / 2],
    };
    const keys = isSquare(el) ? ['nw', 'ne', 'se', 'sw'] : Object.keys(all);
    return keys.map((k) => ({ k, x: all[k][0], y: all[k][1] }));
  }

  maxWidthHandle(L) {
    const el = L.el;
    if (el.type !== 'text' || el.wrap === 'none' || +el.ax >= 1 || el.locked) return null;
    return { x: L.x + (1 - el.ax) * L.maxW, y: L.y };
  }

  handleAt(L, p) {
    const r = HANDLE_PX / this.k;
    const mw = this.maxWidthHandle(L);
    if (mw && Math.abs(p.x - mw.x) <= r && Math.abs(p.y - mw.y) <= r) return { k: 'maxw' };
    return this.handles(L).find((h) => Math.abs(p.x - h.x) <= r && Math.abs(p.y - h.y) <= r) || null;
  }

  // ---------------------------------------------------------------- pointer

  onDown(e) {
    this.canvas.setPointerCapture(e.pointerId);
    const p = this.point(e);
    if (e.button === 1 || (e.button === 0 && this.spaceDown)) {
      this.drag = { kind: 'pan', sx: p.sx, sy: p.sy, panX: this.panX, panY: this.panY };
      this.canvas.style.cursor = 'grabbing';
      return;
    }
    if (e.button !== 0) return;
    const layouts = this.app.layouts();
    const sel = this.app.selected();
    const selL = sel && layouts.get(sel.id);
    if (selL) {
      const handle = this.handleAt(selL, p);
      if (handle) {
        this.drag = { kind: handle.k === 'maxw' ? 'maxw' : 'resize', handle: handle.k, el: sel, L: selL, start: p, box0: { ...selL.box }, began: false };
        return;
      }
    }
    const hit = this.hit(p);
    if (!hit) {
      this.app.select(null);
      this.drag = { kind: 'pan', sx: p.sx, sy: p.sy, panX: this.panX, panY: this.panY, empty: true };
      return;
    }
    if (!sel || hit.id !== sel.id) this.app.select(hit.id);
    const L = this.app.layouts().get(hit.id);
    this.drag = { kind: 'move', el: hit, L, start: p, x0: hit.x, y0px: L.y, anchor0: L.x, box0: { ...L.box }, began: false };
  }

  onMove(e) {
    const p = this.point(e);
    const face = this.app.face();
    if (face) {
      const sx = face.outW / face.designW, sy = face.outH / face.designH;
      this.app.setPointer({ x: p.x / sx, y: p.y / sy });
    }
    const d = this.drag;
    if (!d) {
      this.updateHover(p);
      return;
    }
    if (d.kind === 'pan') {
      this.panX = d.panX + (p.sx - d.sx);
      this.panY = d.panY + (p.sy - d.sy);
      this.invalidate();
      return;
    }
    const dxs = (p.x - d.start.x) * this.k;
    const dys = (p.y - d.start.y) * this.k;
    if (!d.began) {
      if (Math.hypot(dxs, dys) < 3) return;
      d.began = true;
      this.app.beginEdit();
    }
    const L = d.L;
    const el = d.el;
    if (d.kind === 'move') {
      let dx = p.x - d.start.x;
      let dy = p.y - d.start.y;
      if (e.shiftKey) Math.abs(dx) > Math.abs(dy) ? (dy = 0) : (dx = 0);
      const box = { x: d.box0.x + dx, y: d.box0.y + dy, w: d.box0.w, h: d.box0.h };
      const anchor = el.type === 'text' || el.type === 'rich' ? { x: d.anchor0 + dx, y: d.y0px + dy } : null;
      const s = this.snap(box, anchor, el.id, !e.altKey);
      if (!(e.shiftKey && dx === 0)) dx += s.dx;
      if (!(e.shiftKey && dy === 0)) dy += s.dy;
      this.guides = s.guides;
      el.x = round(d.x0 + dx / L.sx);
      this.setY(el, L, d.y0px + dy);
    } else if (d.kind === 'maxw') {
      el.maxWidth = round(Math.max(1, (p.x - L.x) / (1 - el.ax)) / L.sx);
    } else if (d.kind === 'resize') {
      this.resizeTo(d, p, e.shiftKey);
    }
    this.app.changed({ live: true });
  }

  onUp(e) {
    const d = this.drag;
    this.drag = null;
    this.guides = [];
    this.canvas.style.cursor = this.spaceDown ? 'grab' : '';
    if (d && d.began) this.app.commitEdit();
    if (d && d.kind !== 'pan') this.app.changed({});
    this.invalidate();
  }

  setY(el, L, ypx) {
    if (L.attachedTo) {
      el.attach.offset = round((ypx - L.attachedTo.anchorY) / L.sy);
    }
    el.y = round(ypx / L.sy);
  }

  resizeTo(d, p, keepAspect) {
    const { el, L, box0, handle } = d;
    const dx = p.x - d.start.x;
    const dy = p.y - d.start.y;
    let { x, y, w, h } = box0;
    if (isSquare(el)) {
      const sxn = handle.includes('w') ? -1 : 1;
      const syn = handle.includes('n') ? -1 : 1;
      const delta = Math.abs(dx) > Math.abs(dy) ? dx * sxn : dy * syn;
      const size = Math.max(1, box0.w + delta);
      x = handle.includes('w') ? box0.x + box0.w - size : box0.x;
      y = handle.includes('n') ? box0.y + box0.h - size : box0.y;
      el.x = round(x / L.sx);
      this.setY(el, L, y);
      el.w = el.h = round(size / L.sc);
      return;
    }
    if (handle.includes('w')) { x += dx; w -= dx; }
    if (handle.includes('e')) w += dx;
    if (handle.includes('n')) { y += dy; h -= dy; }
    if (handle.includes('s')) h += dy;
    if (keepAspect && box0.h > 0) {
      const ratio = box0.w / box0.h;
      if (handle === 'n' || handle === 's') w = h * ratio;
      else h = w / ratio;
      if (handle.includes('w')) x = box0.x + box0.w - w;
      if (handle.includes('n')) y = box0.y + box0.h - h;
    }
    if (w < 1) { if (handle.includes('w')) x -= 1 - w; w = 1; }
    if (h < 1) { if (handle.includes('n')) y -= 1 - h; h = 1; }
    el.x = round(x / L.sx);
    this.setY(el, L, y);
    el.w = round(w / L.sx);
    el.h = round(h / L.sy);
  }

  snap(box, anchor, selfId, enabled) {
    if (!enabled) return { dx: 0, dy: 0, guides: [] };
    const face = this.app.face();
    const t = SNAP_PX / this.k;
    const xs = [0, face.outW / 2, face.outW];
    const ys = [0, face.outH / 2, face.outH];
    for (const [id, L] of this.app.layouts()) {
      if (id === selfId || L.el.hidden) continue;
      xs.push(L.box.x, L.box.x + L.box.w / 2, L.box.x + L.box.w);
      ys.push(L.box.y, L.box.y + L.box.h / 2, L.box.y + L.box.h);
      if (L.el.type === 'text' || L.el.type === 'rich') {
        xs.push(L.x);
        ys.push(L.y);
      }
    }
    const candX = [box.x, box.x + box.w / 2, box.x + box.w];
    const candY = [box.y, box.y + box.h / 2, box.y + box.h];
    if (anchor) {
      candX.unshift(anchor.x);
      candY.unshift(anchor.y);
    }
    const best = (cands, targets) => {
      let b = null;
      for (const c of cands) for (const tv of targets) {
        const diff = tv - c;
        if (Math.abs(diff) <= t && (!b || Math.abs(diff) < Math.abs(b.d))) b = { d: diff, at: tv };
      }
      return b;
    };
    const bx = best(candX, xs);
    const by = best(candY, ys);
    const guides = [];
    if (bx) guides.push({ x: bx.at });
    if (by) guides.push({ y: by.at });
    return { dx: bx ? bx.d : 0, dy: by ? by.d : 0, guides };
  }

  updateHover(p) {
    const sel = this.app.selected();
    const L = sel && this.app.layouts().get(sel.id);
    const handle = L && this.handleAt(L, p);
    const cursors = { nw: 'nwse-resize', se: 'nwse-resize', ne: 'nesw-resize', sw: 'nesw-resize', n: 'ns-resize', s: 'ns-resize', e: 'ew-resize', w: 'ew-resize', maxw: 'ew-resize' };
    const hit = handle ? null : this.hit(p);
    if (!this.spaceDown) this.canvas.style.cursor = handle ? cursors[handle.k] : hit ? 'move' : '';
    const id = hit ? hit.id : null;
    if (id !== this.hoverId) {
      this.hoverId = id;
      this.invalidate();
    }
  }

  onWheel(e) {
    e.preventDefault();
    const p = this.point(e);
    if (e.ctrlKey || e.metaKey) {
      this.zoomAt(Math.exp(-e.deltaY * 0.0025), p.sx, p.sy);
    } else {
      this.panX -= e.shiftKey ? e.deltaY : e.deltaX;
      this.panY -= e.shiftKey ? 0 : e.deltaY;
      this.invalidate();
    }
  }

  onDblClick(e) {
    const hit = this.hit(this.point(e));
    if (hit && (hit.type === 'text' || hit.type === 'rich')) this.app.focusProp('text');
  }

  // ---------------------------------------------------------------- drawing

  draw() {
    const { ctx, canvas } = this;
    const dpr = window.devicePixelRatio || 1;
    ctx.setTransform(1, 0, 0, 1, 0, 0);
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    const face = this.app.face();
    if (!face) return;
    const layouts = this.app.layouts();
    const k = this.k;
    ctx.setTransform(dpr * k, 0, 0, dpr * k, dpr * this.panX, dpr * this.panY);

    ctx.save();
    ctx.shadowColor = 'rgba(0,0,0,0.45)';
    ctx.shadowBlur = 28;
    ctx.fillStyle = this.checker(ctx);
    ctx.fillRect(0, 0, face.outW, face.outH);
    ctx.restore();

    const view = this.app.S.view;
    const go = this.app.goImage(face.id);
    const env = { images: this.app.images, measurer: this.app.measurer, missing: [] };
    if (view.mode === 'go' && go) {
      ctx.drawImage(go, 0, 0, face.outW, face.outH);
    } else {
      drawFace(ctx, face, layouts, this.app.data(), env);
      if (go && view.mode === 'overlay') {
        ctx.globalAlpha = view.overlayOpacity;
        ctx.drawImage(go, 0, 0, face.outW, face.outH);
        ctx.globalAlpha = 1;
      } else if (go && view.mode === 'diff') {
        ctx.globalCompositeOperation = 'difference';
        ctx.drawImage(go, 0, 0, face.outW, face.outH);
        ctx.globalCompositeOperation = 'source-over';
      }
    }
    const ref = face.reference;
    if (ref?.visible && ref.src) {
      const e = this.app.images.get(ref.src);
      if (e.status === 'ok') {
        ctx.globalAlpha = +ref.opacity || 0.5;
        ctx.drawImage(e.img, 0, 0, face.outW, face.outH);
        ctx.globalAlpha = 1;
      }
    }
    this.drawChrome(ctx, face, layouts, env.missing);
  }

  checker(ctx) {
    if (!this.checkerCanvas) {
      const c = document.createElement('canvas');
      c.width = c.height = 16;
      const g = c.getContext('2d');
      g.fillStyle = '#2b2d33';
      g.fillRect(0, 0, 16, 16);
      g.fillStyle = '#34363d';
      g.fillRect(0, 0, 8, 8);
      g.fillRect(8, 8, 8, 8);
      this.checkerCanvas = c;
    }
    const pattern = ctx.createPattern(this.checkerCanvas, 'repeat');
    pattern.setTransform(new DOMMatrix().scale(1 / this.k));
    return pattern;
  }

  outline(ctx, L, color, dashed) {
    ctx.save();
    const deg = canRotate(L.el) ? +L.el.rotation || 0 : 0;
    if (deg % 360) {
      ctx.translate(L.pivot.x, L.pivot.y);
      ctx.rotate((deg * Math.PI) / 180);
      ctx.translate(-L.pivot.x, -L.pivot.y);
    }
    ctx.strokeStyle = color;
    ctx.lineWidth = 1.25 / this.k;
    if (dashed) ctx.setLineDash([4 / this.k, 3 / this.k]);
    ctx.strokeRect(L.box.x, L.box.y, L.box.w, L.box.h);
    ctx.restore();
  }

  drawChrome(ctx, face, layouts, missing) {
    const k = this.k;
    const px = 1 / k;
    ctx.lineWidth = px;
    ctx.strokeStyle = 'rgba(255,255,255,0.18)';
    ctx.strokeRect(0, 0, face.outW, face.outH);

    for (const L of missing) {
      const { x, y, w, h } = L.box;
      ctx.save();
      ctx.strokeStyle = '#f5a524';
      ctx.setLineDash([5 * px, 4 * px]);
      ctx.strokeRect(x, y, w, h);
      ctx.beginPath();
      ctx.moveTo(x, y);
      ctx.lineTo(x + w, y + h);
      ctx.moveTo(x + w, y);
      ctx.lineTo(x, y + h);
      ctx.stroke();
      ctx.restore();
    }

    const sel = this.app.selected();
    if (this.hoverId && (!sel || this.hoverId !== sel.id)) {
      const L = layouts.get(this.hoverId);
      if (L) this.outline(ctx, L, 'rgba(94,160,255,0.7)');
    }
    const L = sel && layouts.get(sel.id);
    if (L) {
      const accent = '#4f8cff';
      if (L.attachedTo) {
        ctx.save();
        ctx.strokeStyle = '#c084fc';
        ctx.setLineDash([4 * px, 3 * px]);
        ctx.lineWidth = 1.25 * px;
        const ax = L.box.x + Math.min(24 * px, L.box.w / 2);
        ctx.beginPath();
        ctx.moveTo(ax, L.attachedTo.anchorY);
        ctx.lineTo(ax, L.y);
        ctx.stroke();
        ctx.beginPath();
        ctx.moveTo(L.attachedTo.box.x, L.attachedTo.anchorY);
        ctx.lineTo(L.attachedTo.box.x + L.attachedTo.box.w, L.attachedTo.anchorY);
        ctx.stroke();
        ctx.restore();
      }
      if (sel.type === 'text' && sel.wrap !== 'none') {
        ctx.save();
        ctx.strokeStyle = 'rgba(79,140,255,0.55)';
        ctx.setLineDash([3 * px, 3 * px]);
        const left = L.x - sel.ax * L.maxW;
        ctx.strokeRect(left, L.box.y, L.maxW, L.box.h);
        ctx.restore();
      }
      this.outline(ctx, L, sel.locked ? '#888' : accent);
      if (sel.type === 'text' || sel.type === 'rich') {
        ctx.save();
        ctx.strokeStyle = accent;
        ctx.lineWidth = 1.5 * px;
        const r = 5 * px;
        ctx.beginPath();
        ctx.moveTo(L.x - r, L.y);
        ctx.lineTo(L.x + r, L.y);
        ctx.moveTo(L.x, L.y - r);
        ctx.lineTo(L.x, L.y + r);
        ctx.stroke();
        ctx.restore();
      }
      const hs = HANDLE_PX * px;
      ctx.fillStyle = '#fff';
      ctx.strokeStyle = accent;
      ctx.lineWidth = 1.25 * px;
      for (const h of this.handles(L)) {
        ctx.fillRect(h.x - hs / 2, h.y - hs / 2, hs, hs);
        ctx.strokeRect(h.x - hs / 2, h.y - hs / 2, hs, hs);
      }
      const mw = this.maxWidthHandle(L);
      if (mw) {
        ctx.save();
        ctx.translate(mw.x, mw.y);
        ctx.rotate(Math.PI / 4);
        ctx.fillRect(-hs / 2, -hs / 2, hs, hs);
        ctx.strokeRect(-hs / 2, -hs / 2, hs, hs);
        ctx.restore();
      }
    }

    ctx.save();
    ctx.strokeStyle = '#ff4fd8';
    ctx.lineWidth = px;
    for (const g of this.guides) {
      ctx.beginPath();
      if (g.x !== undefined) {
        ctx.moveTo(g.x, -20 * px);
        ctx.lineTo(g.x, face.outH + 20 * px);
      } else {
        ctx.moveTo(-20 * px, g.y);
        ctx.lineTo(face.outW + 20 * px, g.y);
      }
      ctx.stroke();
    }
    ctx.restore();
  }
}

export function isTyping(e) {
  const t = e.target;
  return !!(t && t.closest && t.closest('input, textarea, select, [contenteditable="true"]'));
}
