// Canvas renderer. Each element is drawn the way the Go helper that codegen
// emits for it draws (drawImageContain, drawImageCover, drawText,
// drawVerticalGradientLines, drawRectangularAvatar, ...).

import { resolveSrc } from './model.js';

export const fontCss = (font, px) => `${px}px "${font}"`;

const FALLBACK_METRICS = {
  Aptos: { height: 1.220703125, ascent: 0.93896875, descent: 0.281734375 },
  Helvetica: { height: 1, ascent: 0.770015625, descent: 0.229984375 },
};

export class Measurer {
  constructor(metrics) {
    this.metrics = metrics || {};
    this.ctx = document.createElement('canvas').getContext('2d');
    this.cache = new Map();
  }
  m(font) {
    return this.metrics[font] || (font.startsWith('Aptos') ? FALLBACK_METRICS.Aptos : FALLBACK_METRICS.Helvetica);
  }
  width(text, font, px) {
    const key = font + '|' + px + '|' + text;
    let w = this.cache.get(key);
    if (w === undefined) {
      if (this.cache.size > 20000) this.cache.clear();
      this.ctx.font = fontCss(font, px);
      w = this.ctx.measureText(text).width;
      this.cache.set(key, w);
    }
    return w;
  }
  // gg's FontHeight with HintingFull: the font height quantized up to a whole pixel.
  fh(font, px) {
    return Math.ceil(this.m(font).height * px - 1e-4);
  }
  ascent(font, px) {
    return this.m(font).ascent * px;
  }
  descent(font, px) {
    return this.m(font).descent * px;
  }
}

// ------------------------------------------------------------------ images

export class ImageStore {
  constructor(onChange) {
    this.map = new Map();
    this.onChange = onChange;
  }
  static url(src) {
    return src.startsWith('data:') ? src : '/asset?src=' + encodeURIComponent(src);
  }
  get(src) {
    let e = this.map.get(src);
    if (!e) {
      e = { src, status: 'loading', img: new Image(), crop: null, tints: new Map(), error: '' };
      e.img.onload = () => {
        e.status = 'ok';
        this.onChange?.();
      };
      e.img.onerror = () => {
        e.status = 'error';
        fetch(ImageStore.url(src))
          .then((r) => r.json())
          .then((j) => (e.error = j.error || 'failed to load'))
          .catch(() => (e.error = 'failed to load'))
          .finally(() => this.onChange?.());
      };
      e.img.src = ImageStore.url(src);
      this.map.set(src, e);
    }
    return e;
  }
  invalidate(src) {
    this.map.delete(src);
  }
  // Bounding box of pixels with alpha > 2000/65535, like autoCropNonEmpty.
  crop(e) {
    if (e.crop) return e.crop;
    const { naturalWidth: w, naturalHeight: h } = e.img;
    const f = Math.max(1, Math.ceil(Math.sqrt((w * h) / 4e6)));
    const cw = Math.ceil(w / f), ch = Math.ceil(h / f);
    const c = document.createElement('canvas');
    c.width = cw;
    c.height = ch;
    const ctx = c.getContext('2d', { willReadFrequently: true });
    ctx.drawImage(e.img, 0, 0, cw, ch);
    const data = ctx.getImageData(0, 0, cw, ch).data;
    let x0 = cw, y0 = ch, x1 = -1, y1 = -1;
    for (let y = 0; y < ch; y++) {
      for (let x = 0; x < cw; x++) {
        if (data[(y * cw + x) * 4 + 3] >= 8) {
          if (x < x0) x0 = x;
          if (x > x1) x1 = x;
          if (y < y0) y0 = y;
          if (y > y1) y1 = y;
        }
      }
    }
    e.crop = x1 < 0 ? { x: 0, y: 0, w, h } : {
      x: x0 * f, y: y0 * f,
      w: Math.min(w, (x1 + 1) * f) - x0 * f,
      h: Math.min(h, (y1 + 1) * f) - y0 * f,
    };
    return e.crop;
  }
  // Like tintImageToColor: keep alpha, replace colour.
  tinted(e, color, autoCrop) {
    const key = color + '|' + autoCrop;
    let c = e.tints.get(key);
    if (!c) {
      const r = autoCrop ? this.crop(e) : { x: 0, y: 0, w: e.img.naturalWidth, h: e.img.naturalHeight };
      c = document.createElement('canvas');
      c.width = r.w;
      c.height = r.h;
      const ctx = c.getContext('2d');
      ctx.drawImage(e.img, r.x, r.y, r.w, r.h, 0, 0, r.w, r.h);
      ctx.globalCompositeOperation = 'source-in';
      ctx.fillStyle = color;
      ctx.fillRect(0, 0, r.w, r.h);
      e.tints.set(key, c);
    }
    return c;
  }
}

// ------------------------------------------------------------------ helpers

export function rgba(hex, opacity = 1) {
  const m = /^#?([0-9a-f]{6})$/i.exec((hex || '').trim());
  if (!m) return `rgba(0,0,0,${opacity})`;
  const n = parseInt(m[1], 16);
  return `rgba(${n >> 16},${(n >> 8) & 255},${n & 255},${opacity})`;
}

function roundRect(ctx, x, y, w, h, r) {
  ctx.beginPath();
  if (r > 0) ctx.roundRect(x, y, w, h, Math.min(r, w / 2, h / 2));
  else ctx.rect(x, y, w, h);
}

function rotate(ctx, el, L) {
  const deg = +el.rotation || 0;
  if (deg % 360 === 0) return;
  ctx.translate(L.pivot.x, L.pivot.y);
  ctx.rotate((deg * Math.PI) / 180);
  ctx.translate(-L.pivot.x, -L.pivot.y);
}

function drawFit(ctx, img, sx, sy, sw, sh, x, y, w, h, fit) {
  if (!sw || !sh || w <= 0 || h <= 0) return;
  if (fit === 'stretch') {
    ctx.drawImage(img, sx, sy, sw, sh, Math.round(x), Math.round(y), Math.max(1, Math.round(w)), Math.max(1, Math.round(h)));
    return;
  }
  if (fit === 'cover') {
    const s = Math.max(w / sw, h / sh);
    const dw = sw * s, dh = sh * s;
    ctx.save();
    ctx.beginPath();
    ctx.rect(x, y, w, h);
    ctx.clip();
    ctx.drawImage(img, sx, sy, sw, sh, Math.trunc(x + (w - dw) / 2), Math.trunc(y + (h - dh) / 2), Math.ceil(dw), Math.ceil(dh));
    ctx.restore();
    return;
  }
  const s = Math.min(w / sw, h / sh);
  const dw = sw * s, dh = sh * s;
  ctx.drawImage(img, sx, sy, sw, sh, x + (w - dw) / 2, y + (h - dh) / 2, Math.max(1, Math.round(dw)), Math.max(1, Math.round(dh)));
}

function source(env, e, autoCrop, tint) {
  if (tint) {
    const c = env.images.tinted(e, tint, autoCrop);
    return [c, 0, 0, c.width, c.height];
  }
  if (autoCrop) {
    const r = env.images.crop(e);
    return [e.img, r.x, r.y, r.w, r.h];
  }
  return [e.img, 0, 0, e.img.naturalWidth, e.img.naturalHeight];
}

const qrCache = new Map();
// Stand-in for the QR code: same footprint and colour, fake modules.
function qrPattern(color, seed) {
  const key = color + seed;
  if (qrCache.has(key)) return qrCache.get(key);
  const n = 33;
  const c = document.createElement('canvas');
  c.width = c.height = n;
  const ctx = c.getContext('2d');
  ctx.fillStyle = '#fff';
  ctx.fillRect(0, 0, n, n);
  ctx.fillStyle = color;
  let s = 0;
  for (const ch of seed) s = (s * 31 + ch.charCodeAt(0)) >>> 0;
  const rand = () => ((s = (s * 1664525 + 1013904223) >>> 0) / 4294967296);
  const finder = (x, y) => x < 8 && y < 8;
  for (let y = 0; y < n; y++) {
    for (let x = 0; x < n; x++) {
      if (finder(x, y) || finder(n - 1 - x, y) || finder(x, n - 1 - y)) continue;
      if (rand() < 0.5) ctx.fillRect(x, y, 1, 1);
    }
  }
  for (const [fx, fy] of [[0, 0], [n - 7, 0], [0, n - 7]]) {
    ctx.fillRect(fx, fy, 7, 7);
    ctx.clearRect(fx + 1, fy + 1, 5, 5);
    ctx.fillStyle = '#fff';
    ctx.fillRect(fx + 1, fy + 1, 5, 5);
    ctx.fillStyle = color;
    ctx.fillRect(fx + 2, fy + 2, 3, 3);
  }
  qrCache.set(key, c);
  return c;
}

function drawCenteredInitials(ctx, text, cx, cy, font, px, color, M) {
  ctx.font = fontCss(font, px);
  ctx.fillStyle = color;
  const w = M.width(text, font, px);
  ctx.fillText(text, cx - w / 2, cy + 0.5 * M.fh(font, px));
}

// ------------------------------------------------------------------ elements

const DRAW = {
  image(ctx, el, L, data, env) {
    const src = resolveSrc(el.src, data);
    const e = src ? env.images.get(src) : null;
    const { x, y, w, h } = L.box;
    if (!e || e.status !== 'ok') {
      if ((!e || e.status === 'error') && el.fallbackColor) {
        ctx.fillStyle = rgba(el.fallbackColor);
        ctx.fillRect(x, y, w, h);
      } else if (env.missing && (!e || e.status === 'error')) env.missing.push(L);
      return;
    }
    rotate(ctx, el, L);
    const [img, sx, sy, sw, sh] = source(env, e, el.autoCrop, el.tint);
    drawFit(ctx, img, sx, sy, sw, sh, x, y, w, h, el.fit);
  },

  rect(ctx, el, L) {
    rotate(ctx, el, L);
    const { x, y, w, h } = L.box;
    roundRect(ctx, x, y, w, h, (+el.radius || 0) * L.sc);
    if (el.fill) {
      ctx.fillStyle = rgba(el.fill, el.fillOpacity ?? 1);
      ctx.fill();
    }
    if (el.stroke && +el.strokeWidth > 0) {
      ctx.strokeStyle = rgba(el.stroke, el.strokeOpacity ?? 1);
      ctx.lineWidth = el.strokeWidth * L.sc;
      ctx.stroke();
    }
  },

  circle(ctx, el, L) {
    const { x, y, w } = L.box;
    ctx.beginPath();
    ctx.arc(x + w / 2, y + w / 2, w / 2, 0, Math.PI * 2);
    if (el.fill) {
      ctx.fillStyle = rgba(el.fill, el.fillOpacity ?? 1);
      ctx.fill();
    }
    if (el.stroke && +el.strokeWidth > 0) {
      ctx.strokeStyle = rgba(el.stroke, el.strokeOpacity ?? 1);
      ctx.lineWidth = el.strokeWidth * L.sc;
      ctx.stroke();
    }
  },

  text(ctx, el, L) {
    rotate(ctx, el, L);
    ctx.font = fontCss(L.font, L.px);
    ctx.textBaseline = 'alphabetic';
    if (!L.gradient) {
      ctx.fillStyle = rgba(el.color, el.opacity ?? 1);
      for (const line of L.lines) ctx.fillText(line.text, line.x, line.baseline);
      return;
    }
    const g = el.gradient;
    const ink = (line) => {
      const m = ctx.measureText(line.text);
      return [line.baseline - m.actualBoundingBoxAscent, line.baseline + m.actualBoundingBoxDescent];
    };
    const paint = (top, bottom) => {
      const grad = ctx.createLinearGradient(0, top, 0, Math.max(bottom, top + 1));
      grad.addColorStop(0, g.top);
      grad.addColorStop(Math.min(1, Math.max(0, +g.stop || 0)), g.top);
      grad.addColorStop(1, g.bottom);
      return grad;
    };
    if (L.gradientBlock) {
      const spans = L.lines.filter((l) => l.text.trim()).map(ink);
      if (!spans.length) return;
      ctx.fillStyle = paint(Math.min(...spans.map((s) => s[0])), Math.max(...spans.map((s) => s[1])));
      for (const line of L.lines) ctx.fillText(line.text, line.x, line.baseline);
      return;
    }
    for (const line of L.lines) {
      if (!line.text.trim()) continue;
      const [top, bottom] = ink(line);
      ctx.fillStyle = paint(top, bottom);
      ctx.fillText(line.text, line.x, line.baseline);
    }
  },

  rich(ctx, el, L) {
    rotate(ctx, el, L);
    ctx.textBaseline = 'alphabetic';
    for (const s of L.segments) {
      ctx.font = fontCss(s.font, s.px);
      ctx.fillStyle = rgba(s.color, s.opacity ?? 1);
      ctx.fillText(s.text, s.x, s.baseline);
    }
  },

  qr(ctx, el, L, data, env) {
    const { x, y, w: size } = L.box;
    const sc = L.sc;
    if (el.panel) {
      roundRect(ctx, x, y, size, size, (+el.radius || 0) * sc);
      ctx.fillStyle = rgba(el.panelColor || '#FFFFFF');
      ctx.fill();
    }
    const pad = (+el.padding || 0) * sc;
    const color = el.source === 'vds' || el.forceBlack ? '#000000' : '#0F4C9C';
    ctx.imageSmoothingEnabled = false;
    if (size - pad * 2 > 0) ctx.drawImage(qrPattern(color, el.source || 'payment'), x + pad, y + pad, size - pad * 2, size - pad * 2);
    ctx.imageSmoothingEnabled = true;
    const logo = el.logo || {};
    if ((logo.src || '').trim()) {
      const ls = (+logo.size || 0) * sc;
      const lx = x + (size - ls) / 2, ly = y + (size - ls) / 2;
      roundRect(ctx, lx, ly, ls, ls, (+logo.radius || 0) * sc);
      ctx.fillStyle = rgba(logo.panelColor || '#FFFFFF');
      ctx.fill();
      const e = env.images.get(resolveSrc(logo.src, data));
      if (e.status === 'ok') {
        const lp = (+logo.padding || 0) * sc;
        drawFit(ctx, e.img, 0, 0, e.img.naturalWidth, e.img.naturalHeight, lx + lp, ly + lp, ls - lp * 2, ls - lp * 2, 'contain');
      }
    }
  },

  avatar(ctx, el, L, data, env) {
    const src = resolveSrc(el.src, data);
    const e = src ? env.images.get(src) : null;
    const { x, y, w, h } = L.box;
    const M = env.measurer;
    if (el.shape === 'circle') {
      const r = w / 2, cx = x + r, cy = y + r;
      if (e && e.status === 'ok') {
        ctx.save();
        ctx.beginPath();
        ctx.arc(cx, cy, r, 0, Math.PI * 2);
        ctx.clip();
        drawFit(ctx, e.img, 0, 0, e.img.naturalWidth, e.img.naturalHeight, x, y, w, w, 'cover');
        ctx.restore();
      } else if (!e || e.status === 'error') {
        ctx.beginPath();
        ctx.arc(cx, cy, r, 0, Math.PI * 2);
        ctx.fillStyle = '#ede9fe';
        ctx.fill();
        drawCenteredInitials(ctx, data.Initials || 'NA', cx, cy, 'AptosBold', r * 0.75, '#6d28d9', M);
      }
      if (+el.strokeWidth > 0) {
        ctx.beginPath();
        ctx.arc(cx, cy, r, 0, Math.PI * 2);
        ctx.strokeStyle = rgba(el.stroke || '#FFFFFF');
        ctx.lineWidth = el.strokeWidth * L.sc;
        ctx.stroke();
      }
      return;
    }
    if (e && e.status === 'ok') {
      drawFit(ctx, e.img, 0, 0, e.img.naturalWidth, e.img.naturalHeight, x, y, w, h, 'cover');
    } else if (!e || e.status === 'error') {
      ctx.fillStyle = '#1B1B1E';
      ctx.fillRect(x, y, w, h);
      drawCenteredInitials(ctx, data.Initials || 'SC', x + w / 2, y + h / 2, 'HelveticaBold', Math.min(w, h) * 0.4, '#F6D686', M);
    }
  },
};

// Draws a face in output-pixel coordinates (the caller sets up the transform).
export function drawFace(ctx, face, layouts, data, env) {
  const sc = Math.min(face.outW / face.designW, face.outH / face.designH);
  ctx.save();
  if (+face.clipRadius > 0) {
    roundRect(ctx, 0, 0, face.outW, face.outH, face.clipRadius * sc);
    ctx.clip();
  }
  if (face.background) {
    ctx.fillStyle = rgba(face.background);
    ctx.fillRect(0, 0, face.outW, face.outH);
  }
  for (const el of face.elements) {
    const L = layouts.get(el.id);
    if (!L || !L.drawable || L.hiddenByData) continue;
    ctx.save();
    try {
      DRAW[el.type]?.(ctx, el, L, data, env);
    } finally {
      ctx.restore();
    }
  }
  ctx.restore();
}
