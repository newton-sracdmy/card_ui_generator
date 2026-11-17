// card-ui entry point: state, undo/redo, autosave, keyboard shortcuts and the
// Go render round-trip.

import { Stage, isTyping } from './editor.js';
import { Measurer, ImageStore } from './render.js';
import {
  TYPES, TYPE_LABELS, defaultsFor, newId, normalizeDesign, computeFields, layoutFace, newFace, isSquare,
} from './model.js';
import {
  h, initPanels, renderFaces, renderLayers, renderProps, refreshProps, renderData, renderCard,
  openDesigns, openExport, openCode, showBuildError, toast,
} from './panels.js';

const $ = (sel) => document.querySelector(sel);
const storage = {
  get(k, d) {
    try { return localStorage.getItem(k) ?? d; } catch { return d; }
  },
  set(k, v) {
    try { localStorage.setItem(k, v); } catch { /* storage unavailable */ }
  },
};

// ------------------------------------------------------------------ API

async function request(method, url, body, raw = false) {
  const opts = { method, headers: {} };
  if (body !== undefined) {
    if (raw) opts.body = body;
    else {
      opts.body = JSON.stringify(body);
      opts.headers['Content-Type'] = 'application/json';
    }
  }
  const res = await fetch(url, opts);
  const text = await res.text();
  let data;
  try { data = text ? JSON.parse(text) : null; } catch { data = { error: text }; }
  if (!res.ok) {
    const err = new Error(data?.error || `${res.status} ${res.statusText}`);
    err.data = data;
    throw err;
  }
  return data;
}
const api = {
  get: (url) => request('GET', url),
  put: (url, body) => request('PUT', url, body),
  post: (url, body) => request('POST', url, body),
  upload: (key, file) => request('POST', `/api/upload?key=${encodeURIComponent(key)}`, file, true),
};

// ------------------------------------------------------------------ state

const S = {
  schema: null,
  design: null,
  slug: '',
  faceId: '',
  selectedId: null,
  view: { mode: 'design', overlayOpacity: 0.5, tab: storage.get('cardui.tab', 'props') },
  goRender: null,
  rendering: false,
  dirty: false,
  saveState: '',
  undo: [],
  redo: [],
  pending: null,
  zoom: 1,
  pointer: null,
};

let layoutCache = null;
let dataCache = null;
let saveTimer = 0;
let nudgeTimer = 0;
let panelFrame = 0;

const app = {
  S,
  api,
  images: new ImageStore(() => {
    stage?.invalidate();
    refreshProps();
  }),
  measurer: null,

  face: () => S.design?.faces.find((f) => f.id === S.faceId) || S.design?.faces[0] || null,
  selected: () => app.face()?.elements.find((e) => e.id === S.selectedId) || null,
  data() {
    if (!dataCache) dataCache = computeFields(S.design, S.schema);
    return dataCache;
  },
  layouts() {
    const face = app.face();
    if (!face) return new Map();
    if (!layoutCache || layoutCache.face !== face) {
      const map = layoutFace(face, app.data(), app.measurer);
      // Keep the stored y of attached elements in sync; Go falls back to it.
      for (const L of map.values()) if (L.attachedTo) L.el.y = Math.round((L.y / L.sy) * 100) / 100;
      layoutCache = { face, map };
    }
    return layoutCache.map;
  },

  select(id) {
    S.selectedId = id;
    if (id && S.view.tab !== 'props') setTab('props');
    renderLayers();
    renderProps();
    stage.invalidate();
    updateStatus();
  },

  beginEdit() {
    if (!S.pending) S.pending = JSON.stringify(S.design);
  },
  commitEdit() {
    if (!S.pending) return;
    if (S.pending !== JSON.stringify(S.design)) {
      S.undo.push(S.pending);
      if (S.undo.length > 200) S.undo.shift();
      S.redo = [];
    }
    S.pending = null;
    updateUndoButtons();
  },
  // Every edit ends here: invalidate caches, schedule autosave, refresh views.
  changed({ live = false, structural = false } = {}) {
    layoutCache = null;
    dataCache = null;
    S.dirty = true;
    if (S.view.mode === 'go') setViewMode('design');
    scheduleSave();
    stage.invalidate();
    if (structural) renderPanels();
    else {
      refreshProps();
      if (!live) schedulePanels();
      else scheduleLayers();
    }
    renderViewbar();
    updateStatus();
  },
  mutate(fn, { refit = false } = {}) {
    app.beginEdit();
    fn();
    app.commitEdit();
    app.changed({ structural: true });
    if (refit) stage.fit();
  },

  setPointer(p) {
    S.pointer = p;
    updateStatus();
  },
  setZoom(k) {
    S.zoom = k;
    updateStatus();
  },
  focusProp(key) {
    setTab('props');
    const node = document.querySelector(`#panel-props [data-prop="${key}"]`);
    node?.focus();
    node?.select?.();
  },
  goImage(faceId) {
    return S.goRender?.faces[faceId]?.img || null;
  },

  // -- element actions
  addElement(type) {
    const face = app.face();
    if (!face) return;
    const el = { id: newId(), ...defaultsFor(type, face) };
    // Drop it in the middle of the visible area.
    const r = $('#canvas-wrap').getBoundingClientRect();
    const sx = face.outW / face.designW, sy = face.outH / face.designH;
    const cx = (r.width / 2 - stage.panX) / stage.k / sx;
    const cy = (r.height / 2 - stage.panY) / stage.k / sy;
    const w = el.w ?? 0, hh = el.h ?? 0;
    el.x = Math.round(Math.min(face.designW - w, Math.max(0, cx - (isSquare(el) || el.type !== 'text' ? w / 2 : 0))));
    el.y = Math.round(Math.min(face.designH - hh, Math.max(0, cy - hh / 2)));
    const used = new Set(face.elements.map((e) => e.name));
    let name = el.name, n = 2;
    while (used.has(name)) name = `${el.name} ${n++}`;
    el.name = name;
    app.mutate(() => face.elements.push(el));
    app.select(el.id);
  },
  deleteSelected() {
    const face = app.face();
    const el = app.selected();
    if (!el) return;
    app.mutate(() => {
      face.elements = face.elements.filter((e) => e.id !== el.id);
      for (const e of face.elements) if (e.attach?.to === el.id) e.attach = null;
    });
    app.select(null);
  },
  duplicateSelected() {
    const face = app.face();
    const el = app.selected();
    if (!el) return;
    const copy = JSON.parse(JSON.stringify(el));
    copy.id = newId();
    copy.name = el.name + ' copy';
    const off = Math.max(1, Math.round(Math.min(face.designW, face.designH) / 50));
    copy.x += off;
    if (copy.attach) copy.attach.offset = (+copy.attach.offset || 0) + off;
    else copy.y += off;
    app.mutate(() => face.elements.splice(face.elements.indexOf(el) + 1, 0, copy));
    app.select(copy.id);
  },
  moveSelected(dir) {
    const face = app.face();
    const el = app.selected();
    if (!el) return;
    const i = face.elements.indexOf(el);
    const j = Math.max(0, Math.min(face.elements.length - 1, i + dir));
    if (i === j) return;
    app.mutate(() => {
      face.elements.splice(i, 1);
      face.elements.splice(j, 0, el);
    });
  },
  reorder(dragId, targetId, above) {
    if (dragId === targetId) return;
    const face = app.face();
    const el = face.elements.find((e) => e.id === dragId);
    if (!el) return;
    app.mutate(() => {
      face.elements = face.elements.filter((e) => e.id !== dragId);
      const ti = face.elements.findIndex((e) => e.id === targetId);
      face.elements.splice(above ? ti + 1 : ti, 0, el);
    });
  },
  nudge(dx, dy) {
    const el = app.selected();
    if (!el || el.locked) return;
    const L = app.layouts().get(el.id);
    app.beginEdit();
    el.x = Math.round((el.x + dx) * 100) / 100;
    if (L?.attachedTo) el.attach.offset = Math.round(((+el.attach.offset || 0) + dy) * 100) / 100;
    el.y = Math.round((el.y + dy) * 100) / 100;
    app.changed({ live: true });
    clearTimeout(nudgeTimer);
    nudgeTimer = setTimeout(() => app.commitEdit(), 600);
  },

  // -- faces
  setFace(id) {
    S.faceId = id;
    S.selectedId = null;
    layoutCache = null;
    stage.needsFit = true;
    stage.fit();
    renderPanels();
    renderViewbar();
  },
  addFace() {
    const cur = app.face();
    const used = new Set(S.design.faces.map((f) => f.id));
    let n = S.design.faces.length + 1;
    while (used.has('face' + n)) n++;
    const face = newFace('face' + n, 'Face ' + n, 'none', cur.outW, cur.outH, cur.designW, cur.designH);
    app.mutate(() => S.design.faces.push(face));
    app.setFace(face.id);
  },
  duplicateFace(id) {
    const src = S.design.faces.find((f) => f.id === id);
    const copy = JSON.parse(JSON.stringify(src));
    const used = new Set(S.design.faces.map((f) => f.id));
    let n = 2;
    while (used.has(src.id + n)) n++;
    copy.id = src.id + n;
    copy.name = src.name + ' ' + n;
    copy.role = 'none';
    const idMap = new Map(copy.elements.map((e) => [e.id, newId()]));
    for (const e of copy.elements) {
      e.id = idMap.get(e.id);
      if (e.attach) e.attach.to = idMap.get(e.attach.to) || null;
    }
    app.mutate(() => S.design.faces.splice(S.design.faces.indexOf(src) + 1, 0, copy));
    app.setFace(copy.id);
  },
  deleteFace(id) {
    if (S.design.faces.length < 2) return;
    const face = S.design.faces.find((f) => f.id === id);
    if (!confirm(`Delete the “${face.name}” face and its ${face.elements.length} layers?`)) return;
    app.mutate(() => (S.design.faces = S.design.faces.filter((f) => f.id !== id)));
    app.setFace(S.design.faces[0].id);
  },

  copySelected() {
    const el = app.selected();
    if (!el) return;
    storage.set('cardui.clipboard', JSON.stringify(el));
    toast(`Copied “${el.name}” — paste with Ctrl+V on any face or design`);
  },
  paste() {
    const raw = storage.get('cardui.clipboard', '');
    const face = app.face();
    if (!raw || !face) return;
    let el;
    try { el = JSON.parse(raw); } catch { return; }
    el.id = newId();
    if (el.attach && !face.elements.some((e) => e.id === el.attach.to)) el.attach = null;
    if (face.elements.some((e) => e.name === el.name)) el.name += ' copy';
    app.mutate(() => face.elements.push(el));
    app.select(el.id);
  },
  async dropFiles(files, point) {
    const face = app.face();
    if (!face || !S.slug) return;
    const images = [...files].filter((f) => /\.(png|jpe?g|webp)$/i.test(f.name));
    if (!images.length) return toast('Drop PNG, JPG or WebP files', 'err');
    const sx = face.outW / face.designW, sy = face.outH / face.designH;
    for (const file of images) {
      const key = `CARD_UI/${S.slug}/${file.name.replace(/[^A-Za-z0-9@._()+-]+/g, '-')}`;
      let img;
      try {
        await api.upload(key, file);
        app.images.invalidate(key);
        img = await loadImage(ImageStore.url(key));
      } catch (err) {
        toast(`${file.name}: ${err.message}`, 'err');
        continue;
      }
      // Layers exported at full card size (common from Figma) cover the face exactly.
      let w = img.naturalWidth, hh = img.naturalHeight, x, y, fit = 'contain';
      if (Math.abs(w / hh - face.outW / face.outH) < 0.01 && w >= face.outW * 0.5) {
        [x, y, w, hh, fit] = [0, 0, face.outW, face.outH, 'stretch'];
      } else {
        const k = Math.min(1, (face.outW * 0.6) / w, (face.outH * 0.6) / hh);
        w *= k;
        hh *= k;
        x = point.x - w / 2;
        y = point.y - hh / 2;
      }
      const r = (v) => Math.round(v * 100) / 100;
      const name = file.name.replace(/\.[^.]+$/, '').replace(/^\d+[-_]/, '').replace(/@\dx$/i, '').replace(/[-_]+/g, ' ').trim();
      const el = { id: newId(), ...defaultsFor('image', face), name: name || 'Image', src: key, x: r(x / sx), y: r(y / sy), w: r(w / sx), h: r(hh / sy), fit };
      app.mutate(() => face.elements.push(el));
      app.select(el.id);
      toast(`Added ${key}`);
    }
  },

  // -- persistence & Go
  openDesign,
  saveNow,
  renderGo,
};

let stage = null;

// ------------------------------------------------------------------ panels

function setTab(tab) {
  S.view.tab = tab;
  storage.set('cardui.tab', tab);
  for (const b of document.querySelectorAll('#right-tabs button')) b.classList.toggle('on', b.dataset.tab === tab);
  for (const p of document.querySelectorAll('.panel')) p.hidden = p.id !== 'panel-' + tab;
  renderPanels();
}

function renderPanels() {
  cancelAnimationFrame(panelFrame);
  panelFrame = 0;
  renderFaces();
  renderLayers();
  if (S.view.tab === 'props') renderProps();
  else if (S.view.tab === 'data') renderData();
  else renderCard();
  $('#design-name').textContent = S.design?.name || '';
  document.title = S.design ? `${S.design.name} · card-ui` : 'card-ui';
}

function schedulePanels() {
  if (!panelFrame) panelFrame = requestAnimationFrame(() => {
    panelFrame = 0;
    renderLayers();
    renderFaces();
  });
}
const scheduleLayers = schedulePanels;

function setViewMode(mode) {
  S.view.mode = mode;
  renderViewbar();
  stage?.invalidate();
}

function renderViewbar() {
  const face = app.face();
  const g = S.goRender;
  const hasGo = !!(face && g?.faces[face.id]);
  const stale = g && g.designJSON !== JSON.stringify(S.design);
  const modes = [['design', 'Editor'], ['go', 'Go render'], ['overlay', 'Overlay'], ['diff', 'Difference']];
  const opacity = h('input', {
    type: 'range', min: 0, max: 1, step: 0.05, value: S.view.overlayOpacity,
    oninput: (e) => {
      S.view.overlayOpacity = +e.target.value;
      stage.invalidate();
    },
  });
  const ref = face?.reference;
  $('#viewbar').replaceChildren(...[
    h('div', { class: 'seg' }, modes.map(([m, l]) => h('button', {
      class: S.view.mode === m ? 'on' : '', disabled: m !== 'design' && !hasGo,
      title: m === 'design' ? 'Live preview (browser)' : hasGo ? '' : 'Render with Go first (Ctrl+Enter)',
      onclick: () => setViewMode(m),
    }, l))),
    S.view.mode === 'overlay' && hasGo && h('label', { class: 'inline' }, 'Go opacity', opacity),
    g && h('span', { class: 'render-info' + (stale ? ' stale' : '') },
      stale ? '● Go render is out of date' : '✓ Go render',
      h('small', null, ` ${g.at.toLocaleTimeString()} · build ${(g.report.buildMs / 1000).toFixed(1)}s` +
        (face && g.report.timings?.[face.id] !== undefined ? ` · draw ${(g.report.timings[face.id] / 1000).toFixed(1)}s` : ''))),
    h('div', { class: 'spacer' }),
    ref?.src && h('button', {
      class: 'small' + (ref.visible ? ' on' : ''), title: 'Toggle the reference mockup overlay',
      onclick: () => app.mutate(() => (ref.visible = !ref.visible)),
    }, 'Mockup'),
    h('div', { class: 'seg zoom' },
      h('button', { title: 'Zoom out (-)', onclick: () => stage.zoomAt(1 / 1.25) }, '−'),
      h('button', { title: 'Actual pixels (Ctrl+1)', onclick: () => stage.zoomTo(1) }, `${Math.round(S.zoom * 100)}%`),
      h('button', { title: 'Zoom in (+)', onclick: () => stage.zoomAt(1.25) }, '+'),
      h('button', { title: 'Fit (Ctrl+0)', onclick: () => stage.fit() }, 'Fit')),
  ].filter(Boolean));
}

function updateStatus() {
  const face = app.face();
  const el = app.selected();
  const parts = [];
  if (S.pointer) parts.push(`x ${Math.round(S.pointer.x * 10) / 10}  y ${Math.round(S.pointer.y * 10) / 10}`);
  if (face) parts.push(`${face.name}: design ${face.designW}×${face.designH} → output ${face.outW}×${face.outH} px`);
  if (el) {
    const L = app.layouts().get(el.id);
    parts.push(`${el.name}${L?.box ? `  (${Math.round(L.box.w / L.sx)}×${Math.round(L.box.h / L.sy)})` : ''}`);
  }
  $('#status-left').textContent = parts.join('   ·   ');
  $('#status-right').textContent = S.saveState;
  const zoomBtn = document.querySelector('#viewbar .zoom button:nth-child(2)');
  if (zoomBtn) zoomBtn.textContent = `${Math.round(S.zoom * 100)}%`;
}

function updateUndoButtons() {
  $('#btn-undo').disabled = !S.undo.length;
  $('#btn-redo').disabled = !S.redo.length;
}

// ------------------------------------------------------------------ persistence

function setSaveState(text) {
  S.saveState = text;
  updateStatus();
}

function scheduleSave() {
  setSaveState('Unsaved changes');
  clearTimeout(saveTimer);
  saveTimer = setTimeout(() => saveNow(), 1200);
}

async function saveNow() {
  clearTimeout(saveTimer);
  if (!S.design || !S.dirty) return;
  S.dirty = false;
  setSaveState('Saving…');
  try {
    await api.put(`/api/designs/${S.slug}`, S.design);
    setSaveState(`Saved ${new Date().toLocaleTimeString()}`);
  } catch (err) {
    S.dirty = true;
    setSaveState('Save failed');
    toast('Save failed: ' + err.message, 'err');
  }
}

async function openDesign(slug) {
  await saveNow();
  let d;
  try {
    d = await api.get(`/api/designs/${slug}`);
  } catch (err) {
    toast(err.message, 'err');
    return openDesigns();
  }
  S.design = normalizeDesign(d);
  S.slug = slug;
  S.faceId = S.design.faces[0]?.id || '';
  S.selectedId = null;
  S.undo = [];
  S.redo = [];
  S.pending = null;
  S.goRender = null;
  S.dirty = false;
  S.view.mode = 'design';
  layoutCache = null;
  dataCache = null;
  storage.set('cardui.last', slug);
  if (location.hash.slice(1) !== slug) history.replaceState(null, '', '#' + slug);
  setSaveState('Saved');
  updateUndoButtons();
  stage.needsFit = true;
  stage.fit();
  renderPanels();
  renderViewbar();
}

function undo() {
  app.commitEdit();
  if (!S.undo.length) return;
  S.redo.push(JSON.stringify(S.design));
  restore(S.undo.pop());
}

function redo() {
  if (!S.redo.length) return;
  S.undo.push(JSON.stringify(S.design));
  restore(S.redo.pop());
}

function restore(json) {
  S.design = normalizeDesign(JSON.parse(json));
  if (!S.design.faces.some((f) => f.id === S.faceId)) S.faceId = S.design.faces[0].id;
  if (!app.selected()) S.selectedId = null;
  updateUndoButtons();
  app.changed({ structural: true });
}

// ------------------------------------------------------------------ Go render

function loadImage(url) {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve(img);
    img.onerror = () => reject(new Error('could not load ' + url));
    img.src = url;
  });
}

async function renderGo() {
  if (!S.design || S.rendering) return;
  S.rendering = true;
  const btn = $('#btn-render');
  btn.classList.add('busy');
  btn.textContent = 'Rendering…';
  await saveNow();
  const designJSON = JSON.stringify(S.design);
  const started = performance.now();
  try {
    const report = await api.post('/api/render', { design: S.design, slug: S.slug });
    const faces = {};
    await Promise.all(Object.entries(report.faces || {}).map(async ([id, url]) => {
      faces[id] = { url, img: await loadImage(url) };
    }));
    S.goRender = { report, faces, designJSON, at: new Date() };
    if (report.error) toast('Go: ' + report.error, 'err');
    for (const [id, msg] of Object.entries(report.faceErrors || {})) toast(`Go ${id}: ${msg}`, 'err');
    if (report.matches === false) toast('Note: the sample designation/business does not match this card’s matcher (Card tab).', 'warn');
    toast(`Rendered with Go in ${((performance.now() - started) / 1000).toFixed(1)}s`);
    if (S.view.mode === 'design') S.view.mode = 'go';
    renderViewbar();
    if (S.view.tab === 'data') renderData();
    stage.invalidate();
  } catch (err) {
    showBuildError(err);
  } finally {
    S.rendering = false;
    btn.classList.remove('busy');
    btn.textContent = 'Render with Go';
  }
}

// ------------------------------------------------------------------ boot

async function loadFonts(names) {
  await Promise.all(names.map(async (name) => {
    const face = new FontFace(name, `url(/fonts/${name}.ttf)`);
    await face.load();
    document.fonts.add(face);
  }));
}

function buildToolbar() {
  $('#add-tools').replaceChildren(...TYPES.map((t) => h('button', {
    title: `Add ${TYPE_LABELS[t].toLowerCase()}`, onclick: () => app.addElement(t),
  }, '+ ', TYPE_LABELS[t])));
  $('#btn-open').onclick = () => openDesigns();
  $('#btn-undo').onclick = undo;
  $('#btn-redo').onclick = redo;
  $('#btn-code').onclick = () => openCode();
  $('#btn-render').onclick = () => renderGo();
  $('#btn-export').onclick = () => openExport();
  for (const b of document.querySelectorAll('#right-tabs button')) b.onclick = () => setTab(b.dataset.tab);
}

function onKey(e) {
  const mod = e.ctrlKey || e.metaKey;
  if (mod && e.key.toLowerCase() === 's') {
    e.preventDefault();
    S.dirty = true;
    saveNow();
    return;
  }
  if (mod && e.key === 'Enter') {
    e.preventDefault();
    renderGo();
    return;
  }
  if (isTyping(e) || document.querySelector('.overlay')) return;
  if (mod && e.key.toLowerCase() === 'z') {
    e.preventDefault();
    e.shiftKey ? redo() : undo();
  } else if (mod && e.key.toLowerCase() === 'y') {
    e.preventDefault();
    redo();
  } else if (mod && e.key.toLowerCase() === 'c') {
    app.copySelected();
  } else if (mod && e.key.toLowerCase() === 'v') {
    e.preventDefault();
    app.paste();
  } else if (mod && e.key.toLowerCase() === 'd') {
    e.preventDefault();
    app.duplicateSelected();
  } else if (mod && e.key === '0') {
    e.preventDefault();
    stage.fit();
  } else if (mod && e.key === '1') {
    e.preventDefault();
    stage.zoomTo(1);
  } else if (mod && e.key === ']') {
    e.preventDefault();
    app.moveSelected(1);
  } else if (mod && e.key === '[') {
    e.preventDefault();
    app.moveSelected(-1);
  } else if (e.key === 'Delete' || e.key === 'Backspace') {
    e.preventDefault();
    app.deleteSelected();
  } else if (e.key === 'Escape') {
    app.select(null);
  } else if (e.key === '+' || e.key === '=') {
    stage.zoomAt(1.25);
  } else if (e.key === '-') {
    stage.zoomAt(1 / 1.25);
  } else if (e.key.startsWith('Arrow')) {
    e.preventDefault();
    const step = e.shiftKey ? 10 : 1;
    const d = { ArrowLeft: [-step, 0], ArrowRight: [step, 0], ArrowUp: [0, -step], ArrowDown: [0, step] }[e.key];
    app.nudge(...d);
  }
}

async function boot() {
  try {
    S.schema = await api.get('/api/schema');
  } catch (err) {
    document.body.textContent = 'card-ui server is not reachable: ' + err.message;
    return;
  }
  if (S.schema.fontMetricsError) toast('Could not read Go font metrics; using defaults. ' + S.schema.fontMetricsError, 'err');
  app.measurer = new Measurer(S.schema.fontMetrics);
  await loadFonts(S.schema.fonts);
  initPanels(app);
  stage = new Stage($('#canvas'), app);
  buildToolbar();
  setTab(S.view.tab);
  updateUndoButtons();
  window.addEventListener('keydown', onKey);
  const wrap = $('#canvas-wrap');
  wrap.addEventListener('dragover', (e) => {
    if (![...e.dataTransfer.types].includes('Files')) return;
    e.preventDefault();
    wrap.classList.add('drop');
  });
  wrap.addEventListener('dragleave', () => wrap.classList.remove('drop'));
  wrap.addEventListener('drop', (e) => {
    e.preventDefault();
    wrap.classList.remove('drop');
    app.dropFiles(e.dataTransfer.files, stage.point(e));
  });
  window.addEventListener('beforeunload', () => {
    // keepalive lets the save finish after the tab closes.
    if (S.dirty && S.design) {
      fetch(`/api/designs/${S.slug}`, {
        method: 'PUT', keepalive: true, headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(S.design),
      });
    }
  });
  window.addEventListener('hashchange', () => {
    const slug = location.hash.slice(1);
    if (slug && slug !== S.slug) openDesign(slug);
  });
  let slug = location.hash.slice(1) || storage.get('cardui.last', '');
  if (!slug) {
    const list = await api.get('/api/designs').catch(() => []);
    slug = list[0]?.slug;
  }
  if (slug) await openDesign(slug);
  else openDesigns();
  window.cardui = app; // handy from the devtools console
}

boot();
