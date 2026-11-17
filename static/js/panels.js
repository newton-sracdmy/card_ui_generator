// DOM panels and dialogs: faces, layers, properties, data, card settings,
// asset browser, design list, code view and export.

import {
  TYPE_LABELS, BANGLA_QR, isSquare, isMultiline, isBoxLike, canRotate, isDrawable,
  templateErrors, resolveSrc, cardIdents, pascal, snake, newDesign, newFace, rescaleFace,
} from './model.js';
import { ImageStore } from './render.js';

let A; // the app (see main.js)
export const initPanels = (app) => (A = app);

// ------------------------------------------------------------------ DOM helpers

const PROPS = new Set(['value', 'checked', 'disabled', 'selected', 'indeterminate']);
export function h(tag, props, ...children) {
  const el = document.createElement(tag);
  for (const [k, v] of Object.entries(props || {})) {
    if (v == null || v === false) continue;
    if (k === 'class') el.className = v;
    else if (k === 'style' && typeof v === 'object') Object.assign(el.style, v);
    else if (k.startsWith('on') && typeof v === 'function') el.addEventListener(k.slice(2).toLowerCase(), v);
    else if (PROPS.has(k)) el[k] = v;
    else el.setAttribute(k, v === true ? '' : v);
  }
  for (const c of children.flat(Infinity)) {
    if (c == null || c === false) continue;
    el.append(c instanceof Node ? c : String(c));
  }
  return el;
}
const $ = (sel) => document.querySelector(sel);
const getPath = (o, p) => p.split('.').reduce((a, k) => (a == null ? a : a[k]), o);
function setPath(o, p, v) {
  const ks = p.split('.');
  const last = ks.pop();
  const t = ks.reduce((a, k) => (a[k] ??= {}), o);
  t[last] = v;
}
const fmtNum = (v) => (v === '' || v == null || Number.isNaN(+v) ? '' : String(Math.round(+v * 1000) / 1000));

// ------------------------------------------------------------------ bound controls

function bindInput(input, obj, path, { parse = (v) => v, format = (v) => v ?? '', structural = false, after } = {}) {
  input.value = format(getPath(obj, path));
  input._refresh = () => {
    if (document.activeElement !== input) input.value = format(getPath(obj, path));
  };
  input.addEventListener('input', () => {
    const v = parse(input.value, getPath(obj, path));
    if (v === undefined) return;
    A.beginEdit();
    setPath(obj, path, v);
    after?.(v);
    A.changed({ live: true });
  });
  input.addEventListener('change', () => {
    A.commitEdit();
    A.changed({ structural });
  });
  return input;
}

const parseNum = (v, prev) => (v.trim() === '' || Number.isNaN(+v) ? undefined : +v);

function num(label, obj, path, opts = {}) {
  const input = h('input', { type: 'number', step: opts.step ?? 'any', min: opts.min, max: opts.max, title: opts.title || label });
  bindInput(input, obj, path, { parse: parseNum, format: fmtNum, structural: opts.structural, after: opts.after });
  return h('label', { class: 'num' + (opts.wide ? ' wide' : '') }, h('span', null, label), input);
}

function text(obj, path, opts = {}) {
  const input = h('input', { type: 'text', placeholder: opts.placeholder || '', spellcheck: false });
  return bindInput(input, obj, path, opts);
}

function textarea(obj, path, opts = {}) {
  const input = h('textarea', { rows: opts.rows || 2, spellcheck: false, 'data-prop': opts.prop });
  return bindInput(input, obj, path, opts);
}

function select(obj, path, options, opts = {}) {
  const input = h('select', null, options.map(([v, l]) => h('option', { value: v }, l)));
  input.value = getPath(obj, path) ?? '';
  input._refresh = () => (input.value = getPath(obj, path) ?? '');
  input.addEventListener('change', () => {
    A.beginEdit();
    const v = opts.numeric ? +input.value : input.value;
    setPath(obj, path, v);
    opts.after?.(v);
    A.commitEdit();
    A.changed({ structural: opts.structural ?? true });
  });
  return input;
}

function check(label, obj, path, opts = {}) {
  const input = h('input', { type: 'checkbox', checked: !!getPath(obj, path) });
  input._refresh = () => (input.checked = !!getPath(obj, path));
  input.addEventListener('change', () => {
    A.beginEdit();
    setPath(obj, path, input.checked);
    opts.after?.(input.checked);
    A.commitEdit();
    A.changed({ structural: opts.structural ?? true });
  });
  return h('label', { class: 'check', title: opts.title }, input, h('span', null, label));
}

function normHex(v) {
  v = (v || '').trim();
  if (/^#?[0-9a-f]{6}$/i.test(v)) return ('#' + v.replace('#', '')).toUpperCase();
  if (/^#?[0-9a-f]{3}$/i.test(v)) return ('#' + v.replace('#', '').split('').map((c) => c + c).join('')).toUpperCase();
  return null;
}

function color(obj, path, opts = {}) {
  const current = () => getPath(obj, path) || '';
  const picker = h('input', { type: 'color', value: current() || '#000000' });
  const hex = h('input', { type: 'text', class: 'hex', value: current(), placeholder: opts.optional ? 'none' : '#RRGGBB', spellcheck: false });
  const sync = () => {
    picker.value = current() || '#000000';
    if (document.activeElement !== hex) hex.value = current();
    wrap.classList.toggle('empty', !current());
  };
  picker.addEventListener('input', () => {
    A.beginEdit();
    setPath(obj, path, picker.value.toUpperCase());
    hex.value = picker.value.toUpperCase();
    wrap.classList.remove('empty');
    A.changed({ live: true });
  });
  picker.addEventListener('change', () => {
    A.commitEdit();
    A.changed({ structural: opts.structural });
  });
  hex.addEventListener('change', () => {
    const v = normHex(hex.value);
    if (v === null && !(opts.optional && hex.value.trim() === '')) {
      hex.value = current();
      return;
    }
    A.beginEdit();
    setPath(obj, path, v || '');
    A.commitEdit();
    A.changed({ structural: opts.structural });
    sync();
  });
  const wrap = h('div', { class: 'color' + (current() ? '' : ' empty') }, picker, hex,
    opts.optional && h('button', {
      class: 'icon', title: 'None', onclick: () => {
        A.beginEdit();
        setPath(obj, path, '');
        A.commitEdit();
        A.changed({ structural: opts.structural });
        sync();
      },
    }, '×'));
  picker._refresh = sync;
  return wrap;
}

function row(label, ...controls) {
  return h('div', { class: 'row' }, h('span', { class: 'lbl' }, label), h('div', { class: 'ctl' }, controls));
}
function section(title, ...rows) {
  return h('div', { class: 'section' }, title && h('h4', null, title), rows);
}
function hint(textValue, kind = '') {
  return h('div', { class: 'hint ' + kind }, textValue);
}
function segmented(obj, path, options, opts = {}) {
  const wrap = h('div', { class: 'seg' });
  const draw = () => {
    wrap.replaceChildren(...options.map(([v, l, title]) => h('button', {
      class: getPath(obj, path) === v ? 'on' : '', title,
      onclick: () => {
        A.beginEdit();
        setPath(obj, path, v);
        A.commitEdit();
        A.changed({ structural: opts.structural ?? true });
        draw();
      },
    }, l)));
  };
  draw();
  wrap._refresh = draw;
  return wrap;
}

function fieldPicker(onPick, { filters = true, images = false } = {}) {
  const names = images ? ['AvatarURL', 'LOGO', 'SignatureURL'] : [...A.S.schema.fields.map((f) => f.name), ...A.S.schema.derivedFields];
  const opts = [['', images ? 'Use field…' : 'Insert field…'], ...names.map((n) => [`{{${n}}}`, n])];
  if (filters && !images) {
    for (const f of A.S.schema.filters) opts.push([`|${f}`, `filter |${f}`]);
  }
  const sel = h('select', { class: 'picker' }, opts.map(([v, l]) => h('option', { value: v }, l)));
  sel.addEventListener('change', () => {
    if (sel.value) onPick(sel.value);
    sel.value = '';
  });
  return sel;
}

function insertAtCursor(input, snippet, obj, path) {
  const start = input.selectionStart ?? input.value.length;
  const end = input.selectionEnd ?? start;
  let value = input.value;
  if (snippet.startsWith('|')) {
    // Put the filter inside the placeholder that ends at or after the cursor.
    const close = value.indexOf('}}', Math.max(0, start - 2));
    if (close >= 0) value = value.slice(0, close) + snippet + value.slice(close);
    else value += snippet;
  } else value = value.slice(0, start) + snippet + value.slice(end);
  A.beginEdit();
  setPath(obj, path, value);
  A.commitEdit();
  A.changed({ structural: true });
}

function assetField(obj, path, opts = {}) {
  const input = text(obj, path, { placeholder: opts.placeholder || 'S3 key, URL or {{Field}}', structural: true });
  const status = h('div', { class: 'asset-status' });
  const update = () => {
    const src = resolveSrc(getPath(obj, path), A.data());
    status.replaceChildren();
    if (!src) return;
    const e = A.images.get(src);
    if (e.status === 'ok') {
      const img = h('img', { src: ImageStore.url(src) });
      const info = `${e.img.naturalWidth}×${e.img.naturalHeight}`;
      status.append(img, h('span', null, info));
    } else if (e.status === 'error') status.append(h('span', { class: 'err' }, e.error || 'not found'));
    else status.append(h('span', { class: 'muted' }, 'loading…'));
  };
  update();
  input._refreshAsset = update;
  const browse = h('button', {
    onclick: () => openAssetPicker({
      current: getPath(obj, path),
      onPick: (key) => {
        A.beginEdit();
        setPath(obj, path, key);
        A.commitEdit();
        A.changed({ structural: true });
      },
    }),
  }, 'Browse…');
  const pick = opts.fields !== false && fieldPicker((v) => {
    A.beginEdit();
    setPath(obj, path, v);
    A.commitEdit();
    A.changed({ structural: true });
  }, { images: true });
  return h('div', { class: 'asset' }, h('div', { class: 'asset-row' }, input, browse), pick, status);
}

// ------------------------------------------------------------------ faces & layers

const ICONS = { image: '▣', text: 'T', rich: 'T⁺', rect: '▭', circle: '◯', qr: '▦', avatar: '☻' };

export function renderFaces() {
  const root = $('#faces');
  const d = A.S.design;
  if (!d) return root.replaceChildren();
  root.replaceChildren(
    ...d.faces.map((f) => h('button', {
      class: 'face-tab' + (f.id === A.S.faceId ? ' on' : ''),
      title: `${f.outW}×${f.outH} · role: ${f.role}`,
      onclick: () => A.setFace(f.id),
    }, f.name, h('small', null, roleLabel(f.role)))),
    h('button', { class: 'face-tab add', title: 'Add face', onclick: () => A.addFace() }, '+'),
  );
}

const roleLabel = (r) => ({ payment: 'Payment', oneid: 'OneID', thumbnail: 'Thumb', none: '—' })[r] || r;

export function renderLayers() {
  const root = $('#layers');
  const face = A.face();
  if (!face) return root.replaceChildren();
  const layouts = A.layouts();
  const items = [...face.elements].reverse().map((el) => {
    const L = layouts.get(el.id);
    const badAttach = el.attach && !L?.attachedTo;
    const li = h('li', {
      class: ['layer', el.id === A.S.selectedId && 'on', el.hidden && 'is-hidden', !isDrawable(el) && !el.hidden && 'is-empty'].filter(Boolean).join(' '),
      draggable: true,
      onclick: () => A.select(el.id),
      ondragstart: (e) => {
        e.dataTransfer.setData('text/plain', el.id);
        e.dataTransfer.effectAllowed = 'move';
        li.classList.add('dragging');
      },
      ondragend: () => li.classList.remove('dragging'),
      ondragover: (e) => {
        e.preventDefault();
        const r = li.getBoundingClientRect();
        li.classList.toggle('drop-above', e.clientY < r.top + r.height / 2);
        li.classList.toggle('drop-below', e.clientY >= r.top + r.height / 2);
      },
      ondragleave: () => li.classList.remove('drop-above', 'drop-below'),
      ondrop: (e) => {
        e.preventDefault();
        const above = li.classList.contains('drop-above');
        li.classList.remove('drop-above', 'drop-below');
        A.reorder(e.dataTransfer.getData('text/plain'), el.id, above);
      },
    },
    h('span', { class: 'ico t-' + el.type }, ICONS[el.type]),
    h('span', { class: 'name' }, el.name || TYPE_LABELS[el.type]),
    el.attach && h('span', { class: 'badge' + (badAttach ? ' warn' : ''), title: badAttach ? 'Attach target must be below this layer' : 'Y follows another element' }, '↧'),
    L?.hiddenByData && h('span', { class: 'badge', title: 'Hidden: its fields are empty in the sample data' }, '∅'),
    h('button', {
      class: 'icon tog' + (el.hidden ? ' off' : ''), title: el.hidden ? 'Show (included in Go)' : 'Hide (excluded from Go)',
      onclick: (e) => {
        e.stopPropagation();
        A.mutate(() => (el.hidden = !el.hidden));
      },
    }, el.hidden ? '◌' : '●'),
    h('button', {
      class: 'icon tog' + (el.locked ? ' on' : ''), title: el.locked ? 'Unlock' : 'Lock on canvas',
      onclick: (e) => {
        e.stopPropagation();
        A.mutate(() => (el.locked = !el.locked));
      },
    }, el.locked ? '🔒' : '🔓'));
    return li;
  });
  root.replaceChildren(
    items.length ? h('ul', null, items) : h('div', { class: 'empty-note' }, 'No layers yet. Add one from the toolbar.'),
  );
}

// ------------------------------------------------------------------ properties

export function renderProps() {
  const root = $('#panel-props');
  const el = A.selected();
  root.replaceChildren(el ? elementProps(el) : faceProps(A.face()));
}

export function refreshProps() {
  for (const node of document.querySelectorAll('#right input, #right select, #right textarea')) {
    node._refresh?.();
    node._refreshAsset?.();
  }
  for (const node of document.querySelectorAll('#right [data-live]')) node._live?.();
}

function live(fn) {
  const span = h('span', { 'data-live': '' });
  span._live = () => span.replaceChildren(fn());
  span._live();
  return span;
}

function elementProps(el) {
  const face = A.face();
  const L = A.layouts().get(el.id);
  const parts = [];
  parts.push(h('div', { class: 'el-head' },
    h('span', { class: 'ico t-' + el.type }, ICONS[el.type]),
    text(el, 'name', { placeholder: TYPE_LABELS[el.type] }),
    h('button', { class: 'icon', title: 'Duplicate (Ctrl+D)', onclick: () => A.duplicateSelected() }, '⧉'),
    h('button', { class: 'icon danger', title: 'Delete (Del)', onclick: () => A.deleteSelected() }, '🗑')));
  parts.push(h('div', { class: 'el-sub' }, TYPE_LABELS[el.type], ' · Go variable ', h('code', null, camelName(el)),
    !isDrawable(el) && !el.hidden ? h('span', { class: 'warn' }, ' · nothing to draw (no fill/stroke) — skipped in Go') : null));

  // position
  const earlier = face.elements.slice(0, face.elements.indexOf(el)).filter((t) => isDrawable(t));
  const attachOpts = [['', 'Fixed position'], ...earlier.slice().reverse().map((t) => [t.id, `Below “${t.name || TYPE_LABELS[t.type]}”`])];
  const attachSel = h('select', null, attachOpts.map(([v, l]) => h('option', { value: v }, l)));
  attachSel.value = el.attach?.to && earlier.some((t) => t.id === el.attach.to) ? el.attach.to : '';
  attachSel.addEventListener('change', () => {
    A.beginEdit();
    if (!attachSel.value) el.attach = null;
    else {
      const T = A.layouts().get(attachSel.value);
      el.attach = { to: attachSel.value, offset: Math.round(((L.y - T.anchorY) / L.sy) * 100) / 100 };
    }
    A.commitEdit();
    A.changed({ structural: true });
  });
  const pos = [
    h('div', { class: 'grid2' }, num('X', el, 'x'), el.attach && L?.attachedTo ? num('↧ Offset', el, 'attach.offset', { title: 'Distance below the anchor element' }) : num('Y', el, 'y')),
    row('Y follows', attachSel),
  ];
  if (el.attach && !L?.attachedTo) pos.push(hint('The anchor element is missing, hidden or above this layer — using the fixed Y.', 'warn'));
  else if (L?.attachedTo) {
    const t = L.attachedTo.el;
    pos.push(hint(t.type === 'text' && isMultiline(t)
      ? 'Y = last line of the anchor text + offset, so it moves when the text wraps.'
      : 'Y = bottom of the anchor element + offset.'));
  }
  if (isBoxLike(el)) {
    if (isSquare(el)) pos.push(h('div', { class: 'grid2' }, num('Size', el, 'w', { after: (v) => (el.h = v) })));
    else pos.push(h('div', { class: 'grid2' }, num('W', el, 'w'), num('H', el, 'h')));
  }
  if (canRotate(el)) pos.push(h('div', { class: 'grid2' }, num('Rotate°', el, 'rotation')));
  pos.push(h('div', { class: 'btns' },
    h('button', { class: 'small', title: 'Centre horizontally on the face', onclick: () => A.mutate(() => centre(el, face, 'x')) }, '↔ Centre on card'),
    isBoxLike(el) && !L?.attachedTo && h('button', { class: 'small', title: 'Centre vertically on the face', onclick: () => A.mutate(() => centre(el, face, 'y')) }, '↕ Middle of card')));
  parts.push(section('Position', pos));

  parts.push(...(TYPE_PROPS[el.type]?.(el, L) || []));
  return h('div', null, parts);
}

function centre(el, face, axis) {
  const sx = face.outW / face.designW, sy = face.outH / face.designH, sc = Math.min(sx, sy);
  const r = (v) => Math.round(v * 100) / 100;
  if (!isBoxLike(el)) {
    el.ax = 0.5;
    el.x = r(face.designW / 2);
    return;
  }
  if (axis === 'x') el.x = r((face.designW - (isSquare(el) ? (el.w * sc) / sx : el.w)) / 2);
  else el.y = r((face.designH - (isSquare(el) ? (el.w * sc) / sy : el.h)) / 2);
}

function camelName(el) {
  const p = pascal(el.name || el.type);
  return p[0].toLowerCase() + p.slice(1);
}

function textErrors(value) {
  const errs = templateErrors(value, A.S.schema);
  return errs.length ? hint(errs.join(', '), 'err') : null;
}

const FONT_OPTIONS = () => A.S.schema.fonts.map((f) => [f, f.replace(/([a-z])([A-Z])/g, '$1 $2')]);

const TYPE_PROPS = {
  image(el) {
    const e = el.src && A.images.get(resolveSrc(el.src, A.data()));
    return [
      section('Image',
        assetField(el, 'src'),
        row('Fit', select(el, 'fit', [['contain', 'Contain'], ['cover', 'Cover (crop)'], ['stretch', 'Stretch']])),
        check('Auto-crop transparent edges', el, 'autoCrop', { title: 'autoCropNonEmpty — use for @4x exports with padding' }),
        row('Tint', color(el, 'tint', { optional: true })),
        row('If missing', color(el, 'fallbackColor', { optional: true })),
        e && e.status === 'ok' && h('button', {
          class: 'small', onclick: () => {
            const r = el.autoCrop ? A.images.crop(e) : { w: e.img.naturalWidth, h: e.img.naturalHeight };
            const f = A.face();
            A.mutate(() => (el.h = Math.round(((el.w * r.h) / r.w) * (f.outW / f.designW) * (f.designH / f.outH) * 100) / 100));
          },
        }, 'Match height to image aspect'),
      ),
    ];
  },

  rect(el) {
    return [section('Shape',
      h('div', { class: 'grid2' }, num('Radius', el, 'radius', { min: 0 })),
      row('Fill', color(el, 'fill', { optional: true, structural: true })),
      h('div', { class: 'grid2' }, num('Fill α', el, 'fillOpacity', { min: 0, max: 1, step: 0.05 })),
      row('Stroke', color(el, 'stroke', { optional: true, structural: true })),
      h('div', { class: 'grid2' }, num('Stroke W', el, 'strokeWidth', { min: 0, structural: true }), num('Stroke α', el, 'strokeOpacity', { min: 0, max: 1, step: 0.05 })))];
  },

  circle(el) {
    return [section('Shape',
      row('Fill', color(el, 'fill', { optional: true, structural: true })),
      h('div', { class: 'grid2' }, num('Fill α', el, 'fillOpacity', { min: 0, max: 1, step: 0.05 })),
      row('Stroke', color(el, 'stroke', { optional: true, structural: true })),
      h('div', { class: 'grid2' }, num('Stroke W', el, 'strokeWidth', { min: 0, structural: true })))];
  },

  text(el) {
    const ta = textarea(el, 'text', { rows: 3, prop: 'text', structural: true });
    const grad = el.gradient;
    const wrapOpts = [
      ['none', 'No wrap (\\n = new line)'],
      ['balance', 'Balance 2 lines (names)'],
      ['wrap', 'Wrap to width'],
      ['fit', 'Shrink to fit'],
    ];
    const out = [
      section('Text',
        ta,
        h('div', { class: 'row-inline' }, fieldPicker((v) => insertAtCursor(ta, v, el, 'text'))),
        textErrors(el.text),
        live(() => h('div', { class: 'resolved' }, '→ ', (A.layouts().get(el.id)?.lines || []).map((l) => l.text).join(' ⏎ ') || '(empty)')),
        row('Font', select(el, 'font', FONT_OPTIONS())),
        h('div', { class: 'grid2' }, num('Size', el, 'size', { min: 1 }), num('Opacity', el, 'opacity', { min: 0, max: 1, step: 0.05 })),
        row('Color', color(el, 'color')),
        row('Align', segmented(el, 'ax', [[0, 'Left', 'x is the left edge'], [0.5, 'Center', 'x is the centre'], [1, 'Right', 'x is the right edge']])),
        !grad.enabled && h('div', { class: 'grid2' }, num('Anchor Y', el, 'ay', { step: 0.1, title: 'gg anchor: baseline = y + ay × font height (0.5 = centred)' })),
        check('Hide when its fields are empty', el, 'hideIfEmpty'),
      ),
      section('Gradient',
        check('Vertical gradient (drawVerticalGradientLines)', grad, 'enabled'),
        grad.enabled && row('Top', color(grad, 'top')),
        grad.enabled && row('Bottom', color(grad, 'bottom')),
        grad.enabled && h('div', { class: 'grid2' }, num('Stop', grad, 'stop', { min: 0, max: 1, step: 0.01, title: 'Fraction of the text height that stays the top colour' })),
      ),
      section('Lines',
        row('Wrap', select(el, 'wrap', wrapOpts)),
        el.wrap !== 'none' && h('div', { class: 'grid2' },
          num('Max W', el, 'maxWidth', { min: 1 }),
          el.wrap !== 'balance' && num('Max lines', el, 'maxLines', { min: 1, step: 1 })),
        el.wrap === 'fit' && h('div', { class: 'grid2' }, num('Min size', el, 'minSize', { min: 1 }), num('Step', el, 'step', { min: 0.1 })),
        isMultiline(el) && h('div', { class: 'grid2' }, num('Spacing', el, 'lineSpacing', { min: 0, title: 'Distance between line centres' })),
        isMultiline(el) && row('Y is the', select(el, 'valign', [['top', 'First line'], ['middle', 'Middle of block'], ['bottom', 'Last line']])),
        el.wrap === 'fit' && live(() => {
          const cur = A.layouts().get(el.id);
          return hint(cur ? `Fitted size: ${Math.round((cur.px / cur.sc) * 100) / 100}` : '');
        }),
      ),
    ];
    return out;
  },

  rich(el) {
    const segs = el.segments.map((s, i) => {
      const ta = text(s, 'text', { structural: false });
      return h('div', { class: 'segment' },
        h('div', { class: 'seg-head' }, h('span', null, `Run ${i + 1}`),
          h('button', { class: 'icon danger', title: 'Remove run', onclick: () => A.mutate(() => el.segments.splice(i, 1)) }, '×')),
        ta,
        h('div', { class: 'row-inline' }, fieldPicker((v) => insertAtCursor(ta, v, s, 'text'))),
        textErrors(s.text),
        row('Font', select(s, 'font', FONT_OPTIONS())),
        h('div', { class: 'grid2' }, num('Size', s, 'size', { min: 1 }), num('Opacity', s, 'opacity', { min: 0, max: 1, step: 0.05 })),
        row('Color', color(s, 'color')));
    });
    return [
      section('Runs (one line, side by side)', segs,
        h('button', {
          class: 'small', onclick: () => A.mutate(() => {
            const last = el.segments[el.segments.length - 1];
            el.segments.push(last ? { ...last, text: 'text' } : { text: 'text', font: 'AptosRegular', size: 40, color: '#FFFFFF', opacity: 1 });
          }),
        }, '+ Add run')),
      section('Layout',
        row('Align', segmented(el, 'ax', [[0, 'Left'], [0.5, 'Center'], [1, 'Right']])),
        h('div', { class: 'grid2' }, num('Anchor Y', el, 'ay', { step: 0.1 })),
        check('Hide when its fields are empty', el, 'hideIfEmpty')),
    ];
  },

  qr(el) {
    const logo = el.logo;
    return [
      section('QR code',
        row('Data', select(el, 'source', [['payment', 'Payment QR (DGePay)'], ['vds', 'VDS / OneID QR']])),
        el.source !== 'vds' && check('Force black (ensureBlackQRCodeImage)', el, 'forceBlack'),
        check('White panel behind', el, 'panel'),
        el.panel && row('Panel', color(el, 'panelColor')),
        h('div', { class: 'grid2' }, el.panel && num('Radius', el, 'radius', { min: 0 }), num('Padding', el, 'padding', { min: 0 })),
        hint('The editor shows a stand-in pattern; Render with Go draws the real code.')),
      section('Centre logo',
        assetField(logo, 'src', { fields: false, placeholder: 'none' }),
        !logo.src && h('button', { class: 'small', onclick: () => A.mutate(() => (logo.src = BANGLA_QR)) }, 'Use Bangla QR logo'),
        logo.src && h('div', { class: 'grid2' }, num('Size', logo, 'size', { min: 0 }), num('Radius', logo, 'radius', { min: 0 })),
        logo.src && h('div', { class: 'grid2' }, num('Padding', logo, 'padding', { min: 0 })),
        logo.src && row('Panel', color(logo, 'panelColor'))),
    ];
  },

  avatar(el) {
    return [section('Photo',
      assetField(el, 'src'),
      row('Shape', select(el, 'shape', [['rect', 'Rectangle (drawRectangularAvatar)'], ['circle', 'Circle']], {
        after: (v) => { if (v === 'circle') el.h = el.w; },
      })),
      el.shape === 'circle' && row('Border', color(el, 'stroke')),
      el.shape === 'circle' && h('div', { class: 'grid2' }, num('Border W', el, 'strokeWidth', { min: 0 })),
      hint('Cover fit. Falls back to the member initials when the photo cannot be loaded.'))];
  },
};

const SIZE_PRESETS = [
  ['', 'Presets…'],
  ['3942x6174', 'Portrait 3942×6174 (UAP student / faculty)'],
  ['4050x6300', 'Portrait 4050×6300 (default / MD)'],
  ['6337x4101', 'Landscape 6337×4101 (Saints Club)'],
  ['6174x3942', 'Landscape 6174×3942 (CR80 @1800 dpi)'],
  ['650x518', 'Thumbnail 650×518'],
];

function faceProps(face) {
  if (!face) return h('div');
  const d = A.S.design;
  const sizeSel = h('select', null, SIZE_PRESETS.map(([v, l]) => h('option', { value: v }, l)));
  sizeSel.addEventListener('change', () => {
    if (!sizeSel.value) return;
    const [w, hh] = sizeSel.value.split('x').map(Number);
    A.mutate(() => {
      const same = face.designW === face.outW && face.designH === face.outH;
      if (same) rescaleFace(face, w, hh);
      face.outW = w;
      face.outH = hh;
    }, { refit: true });
  });
  const designInputs = {};
  const designNum = (label, key) => {
    const input = h('input', { type: 'number', step: 'any', value: face[key] });
    input.addEventListener('change', () => {
      const v = +input.value;
      if (!(v > 0)) return (input.value = face[key]);
      A.mutate(() => {
        if (keep.checked) rescaleFace(face, key === 'designW' ? v : face.designW, key === 'designH' ? v : face.designH);
        else face[key] = v;
      });
    });
    designInputs[key] = input;
    return h('label', { class: 'num' }, h('span', null, label), input);
  };
  const keep = h('input', { type: 'checkbox', checked: true });
  const ref = face.reference;
  return h('div', null,
    h('div', { class: 'el-head' }, h('span', { class: 'ico' }, '◧'), text(face, 'name', { structural: true }),
      h('button', { class: 'icon', title: 'Duplicate face', onclick: () => A.duplicateFace(face.id) }, '⧉'),
      d.faces.length > 1 && h('button', { class: 'icon danger', title: 'Delete face', onclick: () => A.deleteFace(face.id) }, '🗑')),
    h('div', { class: 'el-sub' }, 'Nothing selected — face settings. Go function ', h('code', null, `draw${cardIdents(d).pascal}Card${pascal(face.name)}`)),
    section('Output',
      row('Role', select(face, 'role', [['payment', 'Payment (UniversityCardImages.Payment)'], ['oneid', 'OneID (UniversityCardImages.OneID)'], ['thumbnail', 'Thumbnail'], ['none', 'Not exported']])),
      h('div', { class: 'grid2' }, num('Out W', face, 'outW', { min: 1, step: 1, structural: true }), num('Out H', face, 'outH', { min: 1, step: 1, structural: true })),
      row('Size', sizeSel),
      row('Background', color(face, 'background', { optional: true })),
      h('div', { class: 'grid2' }, num('Corner R', face, 'clipRadius', { min: 0, title: 'Rounds the corners of the final PNG (transparent outside)' }))),
    section('Design units',
      hint('Coordinates are typed in design units and scaled to the output size (sx, sy, sc in Go). Use your Figma frame size here to copy numbers straight from Figma.'),
      h('div', { class: 'grid2' }, designNum('Design W', 'designW'), designNum('Design H', 'designH')),
      h('label', { class: 'check' }, keep, h('span', null, 'Keep layout when changing (rescale coordinates)'))),
    section('Reference mockup',
      hint('Overlay the designer’s full-card export to line elements up. Editor only — never exported.'),
      assetField(ref, 'src', { fields: false, placeholder: 'upload or pick the mockup PNG' }),
      ref.src && check('Show overlay', ref, 'visible'),
      ref.src && h('div', { class: 'grid2' }, num('Opacity', ref, 'opacity', { min: 0, max: 1, step: 0.05 }))));
}

// ------------------------------------------------------------------ data tab

export function renderData() {
  const root = $('#panel-data');
  const d = A.S.design;
  if (!d) return root.replaceChildren();
  const schema = A.S.schema;
  const groups = {};
  for (const f of schema.sampleFields) (groups[f.group] ||= []).push(f);
  const sample = Object.entries(groups).map(([g, fs]) => section(g, fs.map((f) => {
    if (!(f.key in d.sample)) d.sample[f.key] = f.default;
    return h('div', { class: 'row sample' }, h('span', { class: 'lbl', title: f.label }, f.label), text(d.sample, f.key, { placeholder: f.default }));
  })));
  const goData = A.S.goRender?.report?.data;
  const fieldRows = schema.fields.map((f) => {
    const cfg = (d.fields[f.name] ||= { mode: f.source === null ? 'fixed' : 'auto', value: f.fallback });
    if (cfg.value === undefined) cfg.value = f.fallback;
    const mode = f.source === null
      ? h('span', { class: 'muted small' }, 'fixed')
      : select(cfg, 'mode', [['auto', 'from data'], ['fixed', 'fixed']]);
    const value = text(cfg, 'value', { placeholder: cfg.mode === 'fixed' || f.source === null ? 'value' : 'fallback' });
    const preview = live(() => {
      const v = A.data()[f.name];
      const go = goData?.[f.name];
      const mismatch = goData && go !== undefined && go !== v;
      return h('span', { class: mismatch ? 'mismatch' : '', title: f.source ? `auto = ${f.source}` : 'static value' },
        v === '' ? h('i', { class: 'muted' }, 'empty') : v, mismatch ? ` (Go: ${go})` : '');
    });
    return h('div', { class: 'field' },
      h('div', { class: 'field-head' }, h('code', null, f.name), h('span', { class: 'pv' }, preview)),
      h('div', { class: 'field-ctl' }, mode, value));
  });
  root.replaceChildren(
    hint('Sample inputs feed the live preview, Render with Go and the generated test. Field rules become build…CardTemplateData in Go.'),
    ...sample,
    section('Card fields (data.X in Go)',
      hint('“from data” uses the member/merchant value and falls back to the text when it is empty; “fixed” always uses the text.'),
      fieldRows,
      hint('Initials is derived from MemberName. Use {{Field|filter}} in text; filters: ' + schema.filters.join(', ') + '.')),
  );
}

// ------------------------------------------------------------------ card tab

export function renderCard() {
  const root = $('#panel-card');
  const d = A.S.design;
  if (!d) return root.replaceChildren();
  const listInput = (key) => {
    const input = h('input', { type: 'text', value: d.match[key].join(', '), placeholder: 'comma separated, case-insensitive' });
    input.addEventListener('change', () => A.mutate(() => (d.match[key] = input.value.split(',').map((s) => s.trim()).filter(Boolean))));
    return input;
  };
  const ids = live(() => {
    const c = cardIdents(d);
    return h('div', { class: 'names' },
      h('div', null, 'File ', h('code', null, `card_generator/${c.snake}_card_image.go`)),
      h('div', null, 'Builder ', h('code', null, `Build${c.pascal}CardImages`)),
      h('div', null, 'Matcher ', h('code', null, `is${c.pascal}Card`)),
      h('div', null, 'Test ', h('code', null, `Test${c.pascal}CardUIGeneration`)));
  });
  root.replaceChildren(
    section('Card',
      row('Name', text(d, 'name', { structural: true })),
      row('Go name', text(d, 'ident', { placeholder: pascal(d.name), structural: true })),
      ids,
      hint(`Design file: card-ui/designs/${A.S.slug}.json`)),
    section('Which members get this card',
      hint('is…Card returns true when the designation or merchant business name contains any of these.'),
      row('Designation', listInput('designation')),
      row('Business', listInput('business'))),
    section('Faces',
      h('table', { class: 'fields' }, h('tbody', null, d.faces.map((f) => h('tr', null,
        h('td', null, f.name), h('td', null, roleLabel(f.role)), h('td', { class: 'muted' }, `${f.outW}×${f.outH}`)))))),
    section('Generate',
      h('div', { class: 'btns' },
        h('button', { onclick: () => openCode() }, 'View Go code'),
        h('button', { class: 'primary', onclick: () => A.renderGo() }, 'Render with Go'),
        h('button', { class: 'accent', onclick: () => openExport() }, 'Export to card_generator'))),
  );
}

// ------------------------------------------------------------------ modal & toast

export function modal({ title, body, actions = [], wide = false, onClose }) {
  const close = () => {
    overlay.remove();
    document.removeEventListener('keydown', onKey, true);
    onClose?.();
  };
  const onKey = (e) => {
    if (e.key === 'Escape') {
      e.stopPropagation();
      close();
    }
  };
  const footer = h('div', { class: 'modal-foot' }, actions.map((a) => h('button', {
    class: a.kind || '', disabled: a.disabled, onclick: () => a.onClick(close),
  }, a.label)));
  const overlay = h('div', { class: 'overlay', onmousedown: (e) => e.target === overlay && close() },
    h('div', { class: 'modal' + (wide ? ' wide' : '') },
      h('div', { class: 'modal-head' }, h('h3', null, title), h('button', { class: 'icon', onclick: close }, '×')),
      h('div', { class: 'modal-body' }, body),
      actions.length ? footer : null));
  document.addEventListener('keydown', onKey, true);
  $('#modal-root').append(overlay);
  return { close, overlay, footer };
}

export function toast(message, kind = '') {
  const t = h('div', { class: 'toast ' + kind }, message);
  $('#toast-root').append(t);
  setTimeout(() => t.classList.add('out'), kind === 'err' ? 7000 : 3500);
  setTimeout(() => t.remove(), kind === 'err' ? 7600 : 4100);
}

function copyButton(getText, label = 'Copy') {
  return h('button', {
    class: 'small', onclick: async (e) => {
      try {
        await navigator.clipboard.writeText(getText());
        e.target.textContent = 'Copied';
      } catch {
        e.target.textContent = 'Select & copy manually';
      }
      setTimeout(() => (e.target.textContent = label), 1500);
    },
  }, label);
}

function codeBlock(code, highlight = new Set()) {
  return h('pre', { class: 'code' }, code.split('\n').map((line, i) =>
    h('div', { class: highlight.has(i + 1) ? 'hl' : '' }, h('span', { class: 'ln' }, String(i + 1)), line || ' ')));
}

// ------------------------------------------------------------------ asset browser

let lastPrefix = '';
try { lastPrefix = localStorage.getItem('cardui.assetPrefix') || ''; } catch { /* storage unavailable */ }

export function openAssetPicker({ current, onPick }) {
  let prefix = current && !current.startsWith('{{') && !/^https?:|^data:/.test(current) && current.includes('/')
    ? current.slice(0, current.lastIndexOf('/') + 1)
    : lastPrefix;
  const list = h('div', { class: 'asset-list' });
  const crumbs = h('div', { class: 'crumbs' });
  const info = h('div', { class: 'muted small' });
  const folderInput = h('input', { type: 'text', value: '', spellcheck: false });
  const fileInput = h('input', { type: 'file', accept: '.png,.jpg,.jpeg,.webp', multiple: true });
  const manual = h('input', { type: 'text', placeholder: 'FOLDER/file.png or https://…', value: current && !current.startsWith('{{') ? current : '', spellcheck: false });

  const pick = (key, close) => {
    onPick(key);
    close();
  };
  let closeModal;
  const load = async (refresh = false) => {
    try { localStorage.setItem('cardui.assetPrefix', prefix); } catch { /* ignore */ }
    lastPrefix = prefix;
    folderInput.value = prefix || `CARD_UI/${A.S.slug}/`;
    const parts = prefix.split('/').filter(Boolean);
    crumbs.replaceChildren(
      h('button', { class: 'link', onclick: () => { prefix = ''; load(); } }, 'bucket'),
      ...parts.flatMap((p, i) => [' / ', h('button', { class: 'link', onclick: () => { prefix = parts.slice(0, i + 1).join('/') + '/'; load(); } }, p)]),
      h('button', { class: 'small', style: { marginLeft: 'auto' }, onclick: () => load(true) }, 'Refresh'));
    list.replaceChildren(h('div', { class: 'muted' }, 'Loading…'));
    let res;
    try {
      res = await A.api.get(`/api/assets?prefix=${encodeURIComponent(prefix)}${refresh ? '&refresh=1' : ''}`);
    } catch (err) {
      list.replaceChildren(h('div', { class: 'err' }, err.message));
      return;
    }
    info.textContent = res.error ? `S3 listing failed (${res.error}) — showing local files only.` : `Bucket ${res.bucket || ''}`;
    const folders = res.folders.map((f) => h('button', {
      class: 'folder', title: f, onclick: () => { prefix = f; load(); },
    }, '📁 ', f.slice(prefix.length)));
    const files = res.files.filter((f) => /\.(png|jpe?g|webp)$/i.test(f.key)).map((f) => h('button', {
      class: 'file' + (f.key === current ? ' on' : ''), title: f.key, onclick: () => pick(f.key, closeModal),
    },
    h('img', { loading: 'lazy', src: ImageStore.url(f.key) }),
    h('span', null, f.key.slice(prefix.length)),
    h('small', null, `${Math.round(f.size / 1024)} KB`, f.s3 ? '' : ' · local only')));
    list.replaceChildren(...folders, ...files);
    if (!folders.length && !files.length) list.append(h('div', { class: 'muted' }, 'Empty folder.'));
  };
  const upload = async () => {
    const files = [...fileInput.files];
    if (!files.length) return toast('Choose one or more image files first', 'err');
    let folder = folderInput.value.trim().replace(/^\/+/, '');
    if (folder && !folder.endsWith('/')) folder += '/';
    const keys = [];
    for (const file of files) {
      const key = folder + file.name.replace(/[^A-Za-z0-9@._()+-]+/g, '-');
      try {
        await A.api.upload(key, file);
        A.images.invalidate(key);
        keys.push(key);
      } catch (err) {
        toast(`${file.name}: ${err.message}`, 'err');
      }
    }
    if (keys.length === 1) return pick(keys[0], closeModal);
    prefix = folder;
    load(true);
  };
  const body = h('div', { class: 'asset-picker' },
    crumbs, info, list,
    h('div', { class: 'upload' },
      h('h4', null, 'Upload from your computer'),
      hint('Saved to card-ui/assets/<key> so previews work right away. Upload the same file to S3 under the same key before deploying — Export lists any that are missing.'),
      row('Folder', folderInput),
      h('div', { class: 'row-inline' }, fileInput, h('button', { class: 'primary', onclick: upload }, 'Upload'))),
    h('div', { class: 'upload' },
      h('h4', null, 'Use a key or URL'),
      h('div', { class: 'row-inline' }, manual, h('button', { onclick: () => manual.value.trim() && pick(manual.value.trim(), closeModal) }, 'Use'))));
  closeModal = modal({ title: 'Assets', body, wide: true }).close;
  load();
}

// ------------------------------------------------------------------ designs

const NEW_PRESETS = SIZE_PRESETS.filter(([v]) => v && v !== '650x518');

export async function openDesigns() {
  let designs = [];
  try {
    designs = await A.api.get('/api/designs');
  } catch (err) {
    toast(err.message, 'err');
  }
  const name = h('input', { type: 'text', placeholder: 'e.g. Dhaka Club Member', spellcheck: false });
  const preset = h('select', null, [...NEW_PRESETS, ['custom', 'Custom size']].map(([v, l]) => h('option', { value: v }, l)));
  const outW = h('input', { type: 'number', value: 3942 });
  const outH = h('input', { type: 'number', value: 6174 });
  preset.addEventListener('change', () => {
    if (preset.value !== 'custom') [outW.value, outH.value] = preset.value.split('x');
  });
  const units = h('input', { type: 'checkbox' });
  const dW = h('input', { type: 'number', placeholder: 'e.g. 401' });
  const dH = h('input', { type: 'number', placeholder: 'e.g. 637' });
  const unitsRow = h('div', { class: 'grid2', style: { display: 'none' } },
    h('label', { class: 'num' }, h('span', null, 'Design W'), dW), h('label', { class: 'num' }, h('span', null, 'Design H'), dH));
  units.addEventListener('change', () => (unitsRow.style.display = units.checked ? '' : 'none'));
  const back = h('input', { type: 'checkbox', checked: true });
  const thumb = h('input', { type: 'checkbox' });
  const err = h('div', { class: 'err' });

  const create = async (close) => {
    const n = name.value.trim();
    const slug = snake(n);
    if (!n) return (err.textContent = 'Give the card a name.');
    if (designs.some((d) => d.slug === slug)) return (err.textContent = `A design called ${slug} already exists.`);
    const w = +outW.value, hh = +outH.value;
    if (!(w > 0 && hh > 0)) return (err.textContent = 'Output size must be positive.');
    const dw = units.checked && +dW.value > 0 ? +dW.value : w;
    const dh = units.checked && +dH.value > 0 ? +dH.value : hh;
    const faces = [newFace('front', 'Front', 'payment', w, hh, dw, dh)];
    if (back.checked) faces.push(newFace('back', 'Back', 'oneid', w, hh, dw, dh));
    else faces[0].role = 'payment';
    if (thumb.checked) faces.push(newFace('thumbnail', 'Thumbnail', 'thumbnail', 650, 518));
    const d = newDesign(n, faces);
    d.match.designation = [n.toLowerCase()];
    try {
      await A.api.put(`/api/designs/${slug}?create=1`, d);
    } catch (e) {
      return (err.textContent = e.message);
    }
    close();
    A.openDesign(slug);
  };
  const duplicate = async (close) => {
    if (!A.S.design) return;
    const n = name.value.trim() || A.S.design.name + ' copy';
    const slug = snake(n);
    if (designs.some((d) => d.slug === slug)) return (err.textContent = `A design called ${slug} already exists.`);
    const d = JSON.parse(JSON.stringify(A.S.design));
    d.name = n;
    d.ident = '';
    try {
      await A.api.put(`/api/designs/${slug}?create=1`, d);
    } catch (e) {
      return (err.textContent = e.message);
    }
    close();
    A.openDesign(slug);
  };
  const body = h('div', { class: 'designs' },
    h('div', { class: 'design-list' },
      designs.length ? designs.map((d) => h('button', {
        class: 'design' + (d.slug === A.S.slug ? ' on' : ''),
        onclick: () => { m.close(); A.openDesign(d.slug); },
      }, h('strong', null, d.name), h('small', null, d.slug + '.json'), h('small', { class: 'muted' }, d.faces.join(' · '))))
        : h('div', { class: 'muted' }, 'No designs yet.')),
    h('div', { class: 'new-design' },
      h('h4', null, 'New card'),
      row('Name', name),
      row('Output size', preset),
      h('div', { class: 'grid2' }, h('label', { class: 'num' }, h('span', null, 'W px'), outW), h('label', { class: 'num' }, h('span', null, 'H px'), outH)),
      h('label', { class: 'check' }, units, h('span', null, 'Type coordinates in different units (e.g. Figma frame size)')),
      unitsRow,
      h('label', { class: 'check' }, back, h('span', null, 'Back face (OneID)')),
      h('label', { class: 'check' }, thumb, h('span', null, 'Thumbnail face (650×518)')),
      err));
  const m = modal({
    title: 'Designs', body, wide: true,
    actions: [
      A.S.design && { label: 'Duplicate current as new', onClick: duplicate },
      { label: 'Create card', kind: 'primary', onClick: create },
    ].filter(Boolean),
  });
  setTimeout(() => name.focus(), 50);
}

// ------------------------------------------------------------------ code, export, errors

export async function openCode() {
  await A.saveNow();
  let res;
  try {
    res = await A.api.post('/api/codegen', { design: A.S.design, slug: A.S.slug });
  } catch (err) {
    return showBuildError(err);
  }
  const tabs = [
    [res.fileName, res.code],
    ['cardui_helpers.go', res.helpers],
    [res.testName, res.test],
    ['Dispatcher snippet', `// Add near the top of BuildUniversityCardImages (university_card_image.go):\n${res.snippet}`],
  ];
  const view = h('div', { class: 'code-view' });
  const bar = h('div', { class: 'tabs' });
  const show = (i) => {
    bar.replaceChildren(...tabs.map(([n], j) => h('button', { class: j === i ? 'on' : '', onclick: () => show(j) }, n)),
      copyButton(() => tabs[i][1]));
    view.replaceChildren(codeBlock(tabs[i][1]));
  };
  show(0);
  modal({ title: 'Generated Go', body: h('div', null, bar, view), wide: true });
}

export async function openExport() {
  await A.saveNow();
  const d = A.S.design;
  const c = cardIdents(d);
  const appendTest = h('input', { type: 'checkbox', checked: true });
  const result = h('div');
  const body = h('div', { class: 'export' },
    h('p', null, 'Writes into the repository:'),
    h('ul', null,
      h('li', null, h('code', null, `card_generator/${c.snake}_card_image.go`)),
      h('li', null, h('code', null, 'card_generator/cardui_helpers.go'), h('span', { class: 'muted' }, ' (shared helpers, only if changed)'))),
    h('label', { class: 'check' }, appendTest, h('span', null, 'Add / update ', h('code', null, `Test${c.pascal}CardUIGeneration`), ' in university_card_image_test.go')),
    result);
  const m = modal({
    title: 'Export to card_generator', body, wide: true,
    actions: [{
      label: 'Export', kind: 'accent', onClick: async () => {
        result.replaceChildren(h('div', { class: 'muted' }, 'Exporting…'));
        let res;
        try {
          res = await A.api.post('/api/export', { design: A.S.design, slug: A.S.slug, appendTest: appendTest.checked });
        } catch (err) {
          result.replaceChildren(h('div', { class: 'err' }, err.message), err.data?.log ? codeBlock(err.data.log) : null);
          return;
        }
        toast('Exported ' + res.written.length + ' file(s)');
        result.replaceChildren(
          h('h4', null, 'Written'), h('ul', null, res.written.map((w) => h('li', null, h('code', null, w)))),
          ...res.warnings.map((w) => hint(w, 'warn')),
          h('h4', null, 'Wire it into BuildUniversityCardImages'),
          hint('Paste near the top of BuildUniversityCardImages in university_card_image.go, before the student/faculty checks if their designations could also match.'),
          h('div', { class: 'snippet' }, codeBlock(res.snippet), copyButton(() => res.snippet)),
          h('h4', null, 'Run the test'),
          h('div', { class: 'snippet' }, codeBlock(res.runCommand), copyButton(() => res.runCommand)));
      },
    }],
  });
  return m;
}

export function showBuildError(err) {
  const data = err.data || {};
  const log = data.log || '';
  const file = data.fileName;
  const lines = new Set();
  if (file) for (const m of log.matchAll(new RegExp(file.replace('.', '\\.') + ':(\\d+)', 'g'))) lines.add(+m[1]);
  modal({
    title: 'Go build failed', wide: true,
    body: h('div', null,
      h('div', { class: 'err' }, err.message),
      log && h('h4', null, 'Output'), log && codeBlock(log.trim()),
      data.code && h('h4', null, file), data.code && codeBlock(data.code, lines)),
  });
}
