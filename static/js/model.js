// Design model: element defaults, normalisation, field values and layout.
// Layout mirrors card-ui/codegen.py and the Go helpers it calls (drawText,
// splitNameToLines, cardUIFitLines, ...) so the preview matches the Go render.

export const TYPES = ['image', 'text', 'rich', 'rect', 'circle', 'qr', 'avatar'];
export const TYPE_LABELS = {
  image: 'Image', text: 'Text', rich: 'Rich text', rect: 'Rectangle',
  circle: 'Circle', qr: 'QR code', avatar: 'Photo',
};
export const SQUARE_TYPES = new Set(['circle', 'qr']);
export const BANGLA_QR = 'UAP_Student_Card_UI/1790672994-Bangla-QR.png';

let seq = 0;
export const newId = () => 'e' + Date.now().toString(36) + (seq++).toString(36);

export const isSquare = (el) => SQUARE_TYPES.has(el.type) || (el.type === 'avatar' && el.shape === 'circle');
export const isMultiline = (el) => el.type === 'text' && (el.wrap !== 'none' || (el.text || '').includes('\n'));
export const isBoxLike = (el) => !['text', 'rich'].includes(el.type);
export const canRotate = (el) => ['image', 'rect', 'text', 'rich'].includes(el.type);

// Same rule as codegen.is_drawable: hidden elements and empty shapes are skipped in Go.
export function isDrawable(el) {
  if (el.hidden) return false;
  if (el.type === 'rect' || el.type === 'circle') return !!(el.fill || (el.stroke && +el.strokeWidth > 0));
  return true;
}

export function defaultsFor(type, face) {
  const u = Math.max(1, Math.min(face.designW, face.designH) / 100); // 1% of the short side
  const r = (v) => Math.round(v * 100) / 100;
  const base = { type, name: TYPE_LABELS[type], hidden: false, locked: false, attach: null, rotation: 0, x: 0, y: 0 };
  switch (type) {
    case 'image':
      return { ...base, src: '', w: r(u * 30), h: r(u * 30), fit: 'contain', autoCrop: false, tint: '', fallbackColor: '' };
    case 'rect':
      return { ...base, w: r(u * 40), h: r(u * 20), radius: 0, fill: '#FFFFFF', fillOpacity: 1, stroke: '', strokeWidth: 0, strokeOpacity: 1 };
    case 'circle':
      return { ...base, w: r(u * 20), h: r(u * 20), fill: '#FFFFFF', fillOpacity: 1, stroke: '', strokeWidth: 0, strokeOpacity: 1 };
    case 'text':
      return {
        ...base, text: 'Text', font: 'AptosRegular', size: r(u * 5), color: '#FFFFFF', opacity: 1, ax: 0, ay: 0.5,
        gradient: { enabled: false, top: '#C99032', bottom: '#EAE2B1', stop: 0.2 },
        wrap: 'none', maxWidth: r(u * 60), maxLines: 2, lineSpacing: r(u * 6), minSize: r(u * 3), step: r(Math.max(1, u * 0.25)),
        valign: 'top', hideIfEmpty: false,
      };
    case 'rich':
      return {
        ...base, ax: 0, ay: 0.5, hideIfEmpty: false,
        segments: [
          { text: 'Phone: ', font: 'HelveticaRegular', size: r(u * 4), color: '#FFFFFF', opacity: 1 },
          { text: '{{PhoneNumber}}', font: 'HelveticaBold', size: r(u * 4), color: '#FFFFFF', opacity: 1 },
        ],
      };
    case 'qr':
      return {
        ...base, name: 'Payment QR', source: 'payment', w: r(u * 35), h: r(u * 35), panel: true, panelColor: '#FFFFFF',
        radius: r(u), padding: r(u), forceBlack: true,
        logo: { src: BANGLA_QR, size: r(u * 7), radius: r(u * 0.6), padding: r(u * 0.3), panelColor: '#FFFFFF' },
      };
    case 'avatar':
      return { ...base, src: '{{AvatarURL}}', shape: 'rect', w: r(u * 30), h: r(u * 36), stroke: '#FFFFFF', strokeWidth: 0 };
  }
  throw new Error('unknown type ' + type);
}

export function newFace(id, name, role, w, h, designW = w, designH = h) {
  return {
    id, name, role, designW, designH, outW: w, outH: h, background: '', clipRadius: 0,
    reference: { src: '', opacity: 0.5, visible: false }, elements: [],
  };
}

export function newDesign(name, faces) {
  return { version: 1, name, ident: '', match: { designation: [], business: [] }, faces, fields: {}, sample: {} };
}

// Fill in properties added after a design was saved so the editor can rely on them.
export function normalizeDesign(d) {
  d.version ||= 1;
  d.match ||= { designation: [], business: [] };
  d.match.designation ||= [];
  d.match.business ||= [];
  d.fields ||= {};
  d.sample ||= {};
  d.faces ||= [];
  for (const face of d.faces) {
    face.reference ||= { src: '', opacity: 0.5, visible: false };
    face.background ||= '';
    face.clipRadius ||= 0;
    face.elements ||= [];
    face.elements = face.elements.map((el) => {
      const def = defaultsFor(el.type, face);
      const merged = { ...def, ...el };
      if (def.gradient) merged.gradient = { ...def.gradient, ...(el.gradient || {}) };
      if (def.logo) merged.logo = { ...def.logo, ...(el.logo || {}) };
      merged.id ||= newId();
      return merged;
    });
  }
  return d;
}

// ------------------------------------------------------------------ names (mirrors codegen.py)

const words = (s) => (s || '').match(/[A-Za-z0-9]+/g) || [];
export function pascal(name) {
  let out = words(name).map((w) => w[0].toUpperCase() + w.slice(1)).join('');
  if (!out || !/^[A-Za-z]/.test(out)) out = 'Card' + out;
  return out;
}
export function snake(name) {
  return words(name).map((w) => w.toLowerCase()).join('_') || 'card';
}
export function cardIdents(design) {
  const p = pascal(design.ident || pascal(design.name || 'Card'));
  return { pascal: p, snake: snake(p.replace(/([a-z0-9])([A-Z])/g, '$1 $2')) };
}

// ------------------------------------------------------------------ data fields

const MONTHS = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];

function goCardDate(iso) {
  const m = /^\s*(\d{4})-(\d{1,2})-(\d{1,2})\s*$/.exec(iso || '');
  if (!m) return 'N/A';
  return `${m[3].padStart(2, '0')} ${MONTHS[+m[2] - 1]} ${m[1]}`;
}

function sampleValue(expr, sample) {
  if (!expr) return '';
  if (expr.startsWith('date:')) return goCardDate(sample[expr.slice(5)]);
  for (const key of expr.split('||')) {
    const v = (sample[key] ?? '').trim();
    if (v !== '') return v;
  }
  return '';
}

export function buildInitials(name) {
  const parts = (name || '').trim().split(/\s+/).filter(Boolean);
  const out = parts.slice(0, 2).map((p) => p[0].toUpperCase()).join('');
  return out || 'NA';
}

// Template data exactly as build<Card>CardTemplateData would produce it for the sample inputs.
export function computeFields(design, schema) {
  const sample = {};
  for (const f of schema.sampleFields) sample[f.key] = design.sample[f.key] ?? f.default;
  const out = {};
  for (const f of schema.fields) {
    const cfg = design.fields[f.name] || {};
    const value = cfg.value ?? f.fallback;
    if (f.source === null || cfg.mode === 'fixed') {
      out[f.name] = value;
      continue;
    }
    const v = sampleValue(f.sample, sample);
    out[f.name] = v.trim() === '' ? value : v.trim();
  }
  out.Initials = buildInitials(out.MemberName);
  return out;
}

export function sampleWithDefaults(design, schema) {
  const s = {};
  for (const f of schema.sampleFields) s[f.key] = design.sample[f.key] ?? f.default;
  return s;
}

// ------------------------------------------------------------------ templates

const TEMPLATE_RE = /\{\{\s*([A-Za-z0-9_]+)((?:\s*\|\s*[a-z]+)*)\s*\}\}/g;
const SINGLE_FIELD_RE = /^\s*\{\{\s*([A-Za-z0-9_]+)\s*\}\}\s*$/;

function formatCardNumberOrRef(raw) {
  const cleaned = raw.trim().replaceAll('-', '').replaceAll(' ', '');
  if (cleaned.length === 16) return cleaned.match(/.{4}/g).join(' ');
  if (cleaned.length >= 12) return 'XXXX XXXX ' + cleaned.slice(-4);
  if (cleaned === '') return 'XXXX XXXX 1234';
  return raw;
}

function formatSaintsDate(rawDate) {
  const raw = rawDate.trim();
  if (raw === '' || raw === 'N/A') return 'DD/MM/YYYY';
  if (raw.includes('/')) return raw;
  let m = /^(\d{1,2}) ([A-Za-z]+) (\d{4})$/.exec(raw);
  if (m) {
    const idx = MONTHS.findIndex((name) => name === m[2] || name.slice(0, 3) === m[2]);
    if (idx >= 0) return `${m[1].padStart(2, '0')}/${String(idx + 1).padStart(2, '0')}/${m[3]}`;
  }
  m = /^(\d{4})-(\d{2})-(\d{2})/.exec(raw);
  if (m) return `${m[3]}/${m[2]}/${m[1]}`;
  return raw;
}

export const FILTERS = {
  upper: (s) => s.toUpperCase(),
  lower: (s) => s.toLowerCase(),
  title: (s) => s.toLowerCase().replace(/(^|[^\p{L}\p{N}_])(\p{L})/gu, (_, a, b) => a + b.toUpperCase()),
  card: formatCardNumberOrRef,
  date: formatSaintsDate,
  trim: (s) => s.trim(),
};

export function resolveTemplate(text, data) {
  return (text || '').replace(TEMPLATE_RE, (whole, field, filters) => {
    if (!(field in data)) return whole;
    let v = String(data[field] ?? '');
    for (const f of filters.split('|').map((x) => x.trim()).filter(Boolean)) {
      if (FILTERS[f]) v = FILTERS[f](v);
    }
    return v;
  });
}

export function templateFields(text) {
  return [...(text || '').matchAll(TEMPLATE_RE)].map((m) => m[1]);
}

export function templateErrors(text, schema) {
  const known = new Set([...schema.fields.map((f) => f.name), ...schema.derivedFields]);
  const errors = [];
  for (const m of (text || '').matchAll(TEMPLATE_RE)) {
    if (!known.has(m[1])) errors.push(`unknown field {{${m[1]}}}`);
    for (const f of m[2].split('|').map((x) => x.trim()).filter(Boolean)) {
      if (!FILTERS[f]) errors.push(`unknown filter |${f}`);
    }
  }
  return errors;
}

// Image sources may be an S3 key, a URL, or a single {{Field}}.
export function resolveSrc(src, data) {
  const m = SINGLE_FIELD_RE.exec(src || '');
  if (m) return (data[m[1]] || '').trim();
  return (src || '').trim();
}

// ------------------------------------------------------------------ text splitting (Go ports)

const fields = (s) => (s || '').split(/\s+/).filter(Boolean);

// Port of splitNameToLines (student_card_image.go).
export function splitNameToLines(name, maxWidth, measure) {
  const ws = fields(name);
  if (ws.length <= 1) return ws;
  if (measure(name) <= maxWidth) return [name];
  if (ws.length >= 5 && ws[0].includes('.')) {
    let splitIdx = 3;
    if (ws[1].length <= 2 && ws[1].includes('.') && measure(ws.slice(0, 4).join(' ')) <= maxWidth) splitIdx = 4;
    const first = ws.slice(0, splitIdx).join(' ');
    const second = ws.slice(splitIdx).join(' ');
    if (second.trim() !== '' && measure(second) <= maxWidth) return [first, second];
  }
  let bestSplit = 1;
  let bestDiff = Number.MAX_VALUE;
  for (let i = 1; i < ws.length; i++) {
    if (i === 1 && ws[0].length <= 3 && !ws[0].includes('.') && ws.length > 2) continue;
    const w1 = measure(ws.slice(0, i).join(' '));
    const w2 = measure(ws.slice(i).join(' '));
    if (w1 > maxWidth || w2 > maxWidth) continue;
    const diff = Math.abs(w1 - w2);
    if (diff < bestDiff) {
      bestDiff = diff;
      bestSplit = i;
    }
  }
  if (bestSplit === 1 && ws.length > 1 && ws[0].length <= 3 && !ws[0].includes('.') && ws.length > 2) {
    const l1 = ws.slice(0, 2).join(' ');
    const l2 = ws.slice(2).join(' ');
    if (measure(l1) <= maxWidth && measure(l2) <= maxWidth) return [l1, l2];
  }
  if (bestSplit === 1) return [name];
  return [ws.slice(0, bestSplit).join(' '), ws.slice(bestSplit).join(' ')];
}

// Port of cardUIWrapLines.
export function wrapLines(text, maxWidth, maxLines, measure) {
  const ws = fields(text);
  if (ws.length === 0) return [''];
  const lines = [];
  let current = ws[0];
  for (const word of ws.slice(1)) {
    const candidate = current + ' ' + word;
    if (measure(candidate) <= maxWidth) {
      current = candidate;
      continue;
    }
    lines.push(current);
    current = word;
  }
  lines.push(current);
  if (maxLines > 0 && lines.length > maxLines) {
    const head = lines.slice(0, maxLines - 1);
    head.push(lines.slice(maxLines - 1).join(' '));
    return head;
  }
  return lines;
}

// Port of cardUIFitLines.
export function fitLines(text, maxWidth, size, minSize, step, maxLines, measureAt) {
  if (step <= 0) step = 1;
  for (let s = size; s >= minSize; s -= step) {
    const m = (t) => measureAt(t, s);
    const lines = wrapLines(text, maxWidth, 0, m);
    if (lines.length <= maxLines && lines.every((l) => m(l) <= maxWidth)) return { lines, px: s };
  }
  return { lines: wrapLines(text, maxWidth, maxLines, (t) => measureAt(t, minSize)), px: minSize };
}

// ------------------------------------------------------------------ layout

// Lays out every element of a face in output pixels. `M` measures text:
// { width(text, font, px), fh(font, px), ascent(font, px), descent(font, px) }.
export function layoutFace(face, data, M) {
  const sx = face.outW / face.designW;
  const sy = face.outH / face.designH;
  const sc = Math.min(sx, sy);
  const out = new Map();
  face.elements.forEach((el, index) => {
    const L = { el, index, sx, sy, sc, drawable: isDrawable(el), hiddenByData: false, errors: [] };
    const target = el.attach && out.get(el.attach.to);
    let y = el.y * sy;
    if (target && target.drawable && target.index < index) {
      y = target.anchorY + (+el.attach.offset || 0) * sy;
      L.attachedTo = target;
    }
    L.x = el.x * sx;
    L.y = y;
    if (el.type === 'text') layoutText(el, L, data, M);
    else if (el.type === 'rich') layoutRich(el, L, data, M);
    else if (isSquare(el)) {
      L.w = L.h = el.w * sc;
      L.box = { x: L.x, y: L.y, w: L.w, h: L.h };
      L.anchorY = L.y + L.h;
    } else {
      L.w = el.w * sx;
      L.h = el.h * sy;
      L.box = { x: L.x, y: L.y, w: L.w, h: L.h };
      L.anchorY = L.y + L.h;
    }
    if (!L.pivot) L.pivot = { x: L.box.x + L.box.w / 2, y: L.box.y + L.box.h / 2 };
    out.set(el.id, L);
  });
  return out;
}

function hiddenByData(el, texts, data) {
  if (!el.hideIfEmpty) return false;
  const fs = texts.flatMap(templateFields);
  return fs.length > 0 && fs.some((f) => !(data[f] || '').trim());
}

function layoutText(el, L, data, M) {
  const { sx, sy, sc } = L;
  const font = el.font;
  const text = resolveTemplate(el.text, data);
  const multiline = isMultiline(el);
  let px = el.size * sc;
  let lines;
  const maxW = (+el.maxWidth || 0) * sx;
  if (!multiline) lines = [text];
  else if (el.wrap === 'fit') {
    const r = fitLines(text, maxW, px, (+el.minSize || el.size) * sc, (+el.step || 1) * sc, +el.maxLines || 2, (t, s) => M.width(t, font, s));
    lines = r.lines;
    px = r.px;
  } else if (el.wrap === 'balance') lines = splitNameToLines(text, maxW, (t) => M.width(t, font, px));
  else if (el.wrap === 'wrap') lines = wrapLines(text, maxW, +el.maxLines || 2, (t) => M.width(t, font, px));
  else lines = text.split('\n');

  const spacing = multiline ? (+el.lineSpacing || +el.size || 12) * sy : 0;
  let startY = L.y;
  if (multiline && el.valign === 'middle') startY = L.y - ((lines.length - 1) * spacing) / 2;
  else if (multiline && el.valign === 'bottom') startY = L.y - (lines.length - 1) * spacing;

  const gradient = !!el.gradient?.enabled;
  const ax = +el.ax || 0;
  const ay = gradient ? 0.5 : +el.ay;
  const fh = M.fh(font, px);
  const asc = M.ascent(font, px);
  const desc = M.descent(font, px);
  L.font = font;
  L.px = px;
  L.fh = fh;
  L.spacing = spacing;
  L.multiline = multiline;
  L.gradient = gradient;
  L.gradientBlock = gradient && multiline && ax === 0;
  L.lines = lines.map((t, i) => {
    const w = M.width(t, font, px);
    const baseline = startY + i * spacing + ay * fh;
    return { text: t, w, x: L.x - ax * w, baseline, top: baseline - asc, bottom: baseline + desc };
  });
  L.hiddenByData = hiddenByData(el, [el.text], data);
  // Go: <name>BottomY for wrapped text (left at Y when hidden), plain Y otherwise.
  L.anchorY = multiline && !L.hiddenByData ? startY + (lines.length - 1) * spacing : L.y;
  L.box = unionLines(L.lines, L.x, L.y);
  L.pivot = { x: L.x, y: L.y };
  L.maxW = maxW;
}

function layoutRich(el, L, data, M) {
  const { sc } = L;
  const ax = +el.ax || 0;
  const ay = +el.ay;
  const segs = (el.segments || []).map((s) => {
    const px = (+s.size || 12) * sc;
    const text = resolveTemplate(s.text, data);
    return { ...s, text, px, w: M.width(text, s.font, px), fh: M.fh(s.font, px), asc: M.ascent(s.font, px), desc: M.descent(s.font, px) };
  });
  const total = segs.reduce((a, s) => a + s.w, 0);
  let cursor = L.x - ax * total;
  L.segments = segs.map((s) => {
    const baseline = L.y + ay * s.fh;
    const seg = { ...s, x: cursor, baseline, top: baseline - s.asc, bottom: baseline + s.desc };
    cursor += s.w;
    return seg;
  });
  L.hiddenByData = hiddenByData(el, (el.segments || []).map((s) => s.text), data);
  L.anchorY = L.y;
  L.box = unionLines(L.segments.map((s) => ({ x: s.x, w: s.w, top: s.top, bottom: s.bottom })), L.x, L.y);
  L.pivot = { x: L.x, y: L.y };
}

function unionLines(lines, x, y) {
  if (!lines.length) return { x: x - 10, y: y - 10, w: 20, h: 20 };
  let x0 = Infinity, y0 = Infinity, x1 = -Infinity, y1 = -Infinity;
  for (const l of lines) {
    x0 = Math.min(x0, l.x);
    x1 = Math.max(x1, l.x + Math.max(l.w, 1));
    y0 = Math.min(y0, l.top);
    y1 = Math.max(y1, l.bottom);
  }
  return { x: x0, y: y0, w: x1 - x0, h: y1 - y0 };
}

// ------------------------------------------------------------------ geometry helpers

export function rescaleFace(face, newDesignW, newDesignH) {
  const fx = newDesignW / face.designW;
  const fy = newDesignH / face.designH;
  const oldSc = Math.min(face.outW / face.designW, face.outH / face.designH);
  const newSc = Math.min(face.outW / newDesignW, face.outH / newDesignH);
  const fs = oldSc / newSc;
  const r = (v) => Math.round(v * 1000) / 1000;
  for (const el of face.elements) {
    el.x = r(el.x * fx);
    el.y = r(el.y * fy);
    if (el.attach) el.attach.offset = r((+el.attach.offset || 0) * fy);
    if (isSquare(el)) {
      el.w = el.h = r(el.w * fs);
    } else if (el.w !== undefined) {
      el.w = r(el.w * fx);
      el.h = r(el.h * fy);
    }
    for (const k of ['size', 'minSize', 'step', 'radius', 'padding', 'strokeWidth']) {
      if (typeof el[k] === 'number') el[k] = r(el[k] * fs);
    }
    if (el.type === 'text') {
      el.maxWidth = r(el.maxWidth * fx);
      el.lineSpacing = r(el.lineSpacing * fy);
    }
    if (el.logo) for (const k of ['size', 'radius', 'padding']) el.logo[k] = r(el.logo[k] * fs);
    if (el.segments) for (const s of el.segments) s.size = r(s.size * fs);
  }
  face.clipRadius = r(face.clipRadius * fs);
  face.designW = newDesignW;
  face.designH = newDesignH;
}
