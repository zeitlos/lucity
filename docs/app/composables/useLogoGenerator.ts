export const LOGO_FONTS = [
  { name: 'Redaction', family: '"Redaction", serif' },
  { name: 'Redaction 10', family: '"Redaction 10", serif' },
  { name: 'Redaction 20', family: '"Redaction 20", serif' },
  { name: 'Redaction 35', family: '"Redaction 35", serif' },
  { name: 'Redaction 50', family: '"Redaction 50", serif' },
  { name: 'Redaction 70', family: '"Redaction 70", serif' },
  { name: 'Redaction 100', family: '"Redaction 100", serif' },
  { name: 'Alliance No.1', family: '"Alliance No.1", "AllianceNo1", sans-serif' },
  { name: 'Alliance No.2', family: '"Alliance No.2", "AllianceNo2", sans-serif' },
  { name: 'SLTF Ryecroft', family: '"SLTF Ryecroft", "Ryecroft", sans-serif' },
  { name: 'Neue Haas Grotesk', family: '"Neue Haas Grotesk Display Pro", "Neue Haas Grotesk Display", "Neue Haas Grotesk", sans-serif' },
  { name: 'Suisse Int\'l', family: '"Suisse Intl", "Suisse Int\'l", sans-serif' },
  { name: 'Aeonik', family: '"Aeonik", sans-serif' },
  { name: 'Aeonik Extended', family: '"Aeonik Extended", "AeonikExtended", sans-serif' },
  { name: 'ABC Diatype', family: '"ABC Diatype", "Diatype", sans-serif' },
  { name: 'General Sans', family: '"General Sans", sans-serif' },
  { name: 'Space Grotesk', family: '"Space Grotesk", sans-serif' },
  { name: 'Bricolage Grotesque', family: '"Bricolage Grotesque", sans-serif' },
  { name: 'Funnel Display', family: '"Funnel Display", sans-serif' },
  { name: 'Inter', family: '"Inter", sans-serif' },
  { name: 'Mona Sans', family: '"Mona Sans", sans-serif' },
  { name: 'Helvetica Neue', family: '"Helvetica Neue", Helvetica, Arial, sans-serif' },
];

export const LOGO_TEXT_SCALE = 0.2;

export const LOGO_TEXT_SIZES: Record<string, number> = {
  'text-xs': 12,
  'text-sm': 14,
  'text-base': 16,
  'text-lg': 18,
  'text-xl': 20,
  'text-2xl': 24,
  'text-3xl': 30,
  'text-4xl': 36,
  'text-5xl': 48,
};

export function logoFontUnits(preset: string) {
  return (LOGO_TEXT_SIZES[preset] ?? LOGO_TEXT_SIZES['text-3xl']!) / LOGO_TEXT_SCALE;
}

export const LOGO_SWATCHES = [
  '#301c0e', '#0c0806', '#f6ede0', '#ede3d6', '#e3d5c2',
  '#00cf85', '#00b054', '#8843db', '#ff5199', '#e62b34', '#ffffff',
];

export const LOGO_PARAMS: Record<string, any> = {
  bendW: { label: 'Bend width', min: 20, max: 240, step: 1, value: 72 },
  bendH: { label: 'Bend height', min: 20, max: 56, step: 1, value: 43 },
  flat: { label: 'Flat length', min: 0, max: 320, step: 1, value: 121 },
  tension: { label: 'Curve tension', min: 0.15, max: 0.99, step: 0.01, value: 0.65 },
  topFlat: { label: 'Top vertical part', min: 0, max: 50, step: 1, value: 0 },
  botFlat: { label: 'Bottom vertical part', min: 0, max: 50, step: 1, value: 40 },
  depth: { label: 'Extrusion width', min: 10, max: 300, step: 1, value: 288 },
  text: { label: 'Text', type: 'text', value: 'Lucity' },
  font: { label: 'Font', type: 'select', value: 'Redaction', options: LOGO_FONTS.map(f => f.name) },
  weight: { label: 'Weight', type: 'select', value: '500', options: ['300', '400', '500', '600', '700', '800', '900'] },
  fstyle: { label: 'Style', type: 'select', value: 'Normal', options: ['Normal', 'Italic'] },
  lspace: { label: 'Letter spacing', min: -10, max: 40, step: 0.5, value: 0 },
  inFsize: { label: 'Font size', type: 'select', value: 'text-3xl', options: Object.keys(LOGO_TEXT_SIZES) },
  inGap: { label: 'Gap right of mark', min: 0, max: 200, step: 1, value: 60 },
  inDy: { label: 'Vertical offset', min: -20, max: 20, step: 1, value: 0 },
  bgOn: { label: 'Background', type: 'select', value: 'None', options: ['None', 'Color'] },
  bg: { label: 'Background color', type: 'swatch', value: '#f6ede0' },
  fill: { label: 'Logo color', type: 'swatch', value: '#301c0e' },
  textCol: { label: 'Text color', type: 'swatch', value: '#301c0e' },
  pixStyle: { label: 'Pixelate', type: 'select', value: 'Off', options: ['Off', 'Grid', 'Dither'] },
  pixSize: { label: 'Pixel size', min: 3, max: 40, step: 1, value: 10 },
  shade: { label: 'Shading', type: 'select', value: 'Off', options: ['Off', 'Gradient', 'Grain', 'Lines'] },
  shadeK: { label: 'Shade strength', min: 0.1, max: 1, step: 0.05, value: 0.7 },
  detail: { label: 'Shade detail', min: 2, max: 20, step: 1, value: 4 },
  zoom: { label: 'Header scale', min: 0.05, max: 0.22, step: 0.005, value: 0.2 },
  previewZoom: { label: 'Preview scale', min: 0.05, max: 1.5, step: 0.005, value: 0.22 },
};

export const LOGO_GROUPS: [string, string[]][] = [
  ['Profile', ['bendW', 'bendH', 'flat', 'tension', 'topFlat', 'botFlat']],
  ['Extrusion', ['depth']],
  ['Wordmark', ['text', 'font', 'weight', 'fstyle', 'lspace', 'inFsize', 'inGap', 'inDy']],
  ['Colors', ['bgOn', 'bg', 'fill', 'textCol']],
  ['Style', ['pixStyle', 'pixSize', 'shade', 'shadeK', 'detail']],
  ['Display', ['zoom', 'previewZoom']],
];

export const LOGO_STORE_KEY = 'lucity-logo-generator-v1';

type State = Record<string, any>;

const n2 = (v: number) => Math.round(v * 100) / 100;

const n = (v: number) => +(+v).toFixed(2);
const pt = (p: number[], o: number[] = [0, 0]) => `${n(p[0]! + o[0]!)} ${n(p[1]! + o[1]!)}`;

function pts(s: State) {
  const t = +s.tension;
  const bendW = +s.bendW, bendH = +s.bendH, flat = +s.flat;
  const top = +s.topFlat, bot = +s.botFlat;
  const W = 2 * bendW + flat;
  const m = (p: number[]) => [W - p[0]!, p[1]!];
  const A0 = [0, 0];
  const A = [0, top];
  const B = [bendW, top + bendH];
  const C = [bendW + flat, top + bendH];
  const D = [2 * bendW + flat, top + 2 * bendH];
  const D1 = [2 * bendW + flat, top + 2 * bendH + bot];
  return {
    A0: m(A0), A: m(A), B: m(B), C: m(C), D: m(D), D1: m(D1),
    b1c1: m([0, top + bendH * t]),
    b1c2: m([bendW * (1 - t), top + bendH]),
    b2c1: m([C[0]! + bendW * t, top + bendH]),
    b2c2: m([D[0]!, top + 2 * bendH - bendH * t]),
    W,
    H: top + 2 * bendH + bot,
  };
}

function frontPath(s: State) {
  const p = pts(s);
  return `M ${pt(p.A0)} L ${pt(p.A)} C ${pt(p.b1c1)}, ${pt(p.b1c2)}, ${pt(p.B)} L ${pt(p.C)} C ${pt(p.b2c1)}, ${pt(p.b2c2)}, ${pt(p.D)} L ${pt(p.D1)}`;
}

function sheetPath(s: State, o: number[]) {
  const p = pts(s);
  return frontPath(s)
    + ` L ${pt(p.D1, o)}`
    + ` L ${pt(p.D, o)}`
    + ` C ${pt(p.b2c2, o)}, ${pt(p.b2c1, o)}, ${pt(p.C, o)}`
    + ` L ${pt(p.B, o)}`
    + ` C ${pt(p.b1c2, o)}, ${pt(p.b1c1, o)}, ${pt(p.A, o)}`
    + ` L ${pt(p.A0, o)}`
    + ' Z';
}

const BAYER = [
  [0, 8, 2, 10], [12, 4, 14, 6], [3, 11, 1, 9], [15, 7, 13, 5],
].map(row => row.map(v => (v + 0.5) / 16));

function pixelRects(s: State, markW: number, markH: number) {
  const px = +s.pixSize;
  const SS = 4;
  const cols = Math.ceil(markW / px);
  const rows = Math.ceil(markH / px);
  const cv = document.createElement('canvas');
  cv.width = cols * SS;
  cv.height = rows * SS;
  const ctx = cv.getContext('2d', { willReadFrequently: true })!;
  ctx.scale(SS / px, SS / px);
  ctx.fillStyle = '#000';
  ctx.fill(new Path2D(sheetPath(s, [+s.depth, 0])));
  const data = ctx.getImageData(0, 0, cv.width, cv.height).data;
  const dither = s.pixStyle === 'Dither';
  const on: boolean[][] = [];
  for (let r = 0; r < rows; r++) {
    on.push(new Array(cols).fill(false));
    for (let c = 0; c < cols; c++) {
      let sum = 0;
      for (let y = 0; y < SS; y++) {
        for (let x = 0; x < SS; x++) {
          sum += data[(((r * SS + y) * cv.width) + (c * SS + x)) * 4 + 3]!;
        }
      }
      const cov = sum / (SS * SS * 255);
      const thr = dither ? BAYER[r % 4]![c % 4]! : 0.5;
      on[r]![c] = cov >= thr;
    }
  }
  const rects = [];
  for (let r = 0; r < rows; r++) {
    let c = 0;
    while (c < cols) {
      if (!on[r]![c]) { c++; continue; }
      let c2 = c;
      while (c2 < cols && on[r]![c2]) c2++;
      rects.push(`<rect x="${n(c * px)}" y="${n(r * px)}" width="${n((c2 - c) * px)}" height="${n(px)}"/>`);
      c = c2;
    }
  }
  return rects.join('\n    ');
}

function hexRgb(h: string) {
  return [parseInt(h.slice(1, 3), 16), parseInt(h.slice(3, 5), 16), parseInt(h.slice(5, 7), 16)];
}

function mixHex(a: string, b: string, t: number) {
  const A = hexRgb(a), B = hexRgb(b);
  return '#' + A.map((v, i) => Math.round(v + (B[i]! - v) * t).toString(16).padStart(2, '0')).join('');
}

function shadeColors(s: State) {
  const K = +s.shadeK;
  return { dark: mixHex(s.fill, '#000000', 0.6 * K), light: mixHex(s.fill, '#ffffff', 0.4 * K) };
}

function shadeTable(s: State) {
  const t = +s.tension, W = +s.bendW, H = +s.bendH;
  const top = +s.topFlat, bot = +s.botFlat;
  const FLOOR = 0.15, N = 32;
  const rows: number[][] = [];
  rows.push([0, FLOOR]);
  if (top > 0) rows.push([top, FLOOR]);
  const bend1: number[][] = [];
  for (let i = 0; i <= N; i++) {
    const u = i / N, iu = 1 - u;
    const y = 3 * iu * iu * u * (H * t) + 3 * iu * u * u * H + u * u * u * H;
    const dx = 6 * iu * u * (W * (1 - t)) + 3 * u * u * (W - W * (1 - t));
    const dy = 3 * iu * iu * (H * t) + 6 * iu * u * (H - H * t);
    const b = FLOOR + (1 - FLOOR) * Math.abs(dx) / Math.hypot(dx, dy);
    bend1.push([y, b]);
    rows.push([top + y, b]);
  }
  for (let i = 0; i <= N; i++) {
    const [y, b] = bend1[N - i]!;
    rows.push([top + 2 * H - y!, 0.08 + (b! - 0.08) * 0.72]);
  }
  if (bot > 0) rows.push([top + 2 * H + bot, 0.10]);
  return rows;
}

function makeBAt(s: State) {
  const table = shadeTable(s);
  return (y: number) => {
    if (y <= table[0]![0]!) return table[0]![1]!;
    for (let i = 1; i < table.length; i++) {
      if (y <= table[i]![0]!) {
        const [y0, b0] = table[i - 1]!;
        const [y1, b1] = table[i]!;
        return y1 === y0 ? b0! : b0! + (b1! - b0!) * (y - y0!) / (y1! - y0!);
      }
    }
    return table[table.length - 1]![1]!;
  };
}

function maskGrid(s: State, markW: number, markH: number, px: number) {
  const SS = 2;
  const cols = Math.ceil(markW / px), rows = Math.ceil(markH / px);
  const cv = document.createElement('canvas');
  cv.width = cols * SS;
  cv.height = rows * SS;
  const ctx = cv.getContext('2d', { willReadFrequently: true })!;
  ctx.scale(SS / px, SS / px);
  ctx.fillStyle = '#000';
  ctx.fill(new Path2D(sheetPath(s, [+s.depth, 0])));
  const data = ctx.getImageData(0, 0, cv.width, cv.height).data;
  const on: boolean[][] = [];
  for (let r = 0; r < rows; r++) {
    on.push(new Array(cols));
    for (let c = 0; c < cols; c++) {
      let sum = 0;
      for (let y = 0; y < SS; y++) {
        for (let x = 0; x < SS; x++) {
          sum += data[(((r * SS + y) * cv.width) + (c * SS + x)) * 4 + 3]!;
        }
      }
      on[r]![c] = sum / (SS * SS * 255) >= 0.4;
    }
  }
  return { cols, rows, on };
}

function rnd(i: number) {
  let x = Math.imul(i ^ (i >>> 16), 2246822507);
  x = Math.imul(x ^ (x >>> 13), 3266489909);
  return ((x ^ (x >>> 16)) >>> 0) / 4294967296;
}

function shadedMark(s: State, markW: number, markH: number, idSuffix: string) {
  const { dark, light } = shadeColors(s);
  const bAt = makeBAt(s);
  const p = pts(s);
  const o = [+s.depth, 0];
  const clipId = `lucity-clip-${idSuffix}`;
  const clip = `<clipPath id="${clipId}"><path d="${sheetPath(s, o)}"/></clipPath>`;

  if (s.shade === 'Gradient') {
    const gid = `lucity-grad-${idSuffix}`;
    const stops = shadeTable(s).map(([y, b]) =>
      `<stop offset="${n(Math.min(1, y! / p.H))}" stop-color="${mixHex(dark, light, b!)}"/>`).join('');
    const defs = `<defs><linearGradient id="${gid}" x1="0" y1="0" x2="0" y2="1">${stops}</linearGradient></defs>`;
    return `${defs}<path d="${sheetPath(s, o)}" fill="url(#${gid})"/>`;
  }

  const K = +s.shadeK;
  const px = +s.detail;
  const base = `<path d="${sheetPath(s, o)}" fill="${light}"/>`;
  let tex = '';
  if (s.shade === 'Grain') {
    const g = maskGrid(s, markW, markH, px);
    const parts = [];
    for (let r = 0; r < g.rows; r++) {
      const density = (1 - bAt((r + 0.5) * px)) * K;
      for (let c = 0; c < g.cols; c++) {
        if (g.on[r]![c] && rnd(r * g.cols + c) < density) {
          parts.push(`<rect x="${n(c * px)}" y="${n(r * px)}" width="${n(px)}" height="${n(px)}"/>`);
        }
      }
    }
    tex = parts.join('');
  } else {
    const spacing = px * 3;
    const g = maskGrid(s, markW, markH, spacing);
    const parts = [];
    for (let r = 0; r < g.rows; r++) {
      const shadow = (1 - bAt((r + 0.5) * spacing)) * K;
      const th = Math.max(0.5, Math.min(spacing * 0.9, spacing * shadow));
      const y = r * spacing + (spacing - th) / 2;
      let c = 0;
      while (c < g.cols) {
        if (!g.on[r]![c]) { c++; continue; }
        let c2 = c;
        while (c2 < g.cols && g.on[r]![c2]) c2++;
        parts.push(`<rect x="${n(c * spacing)}" y="${n(y)}" width="${n((c2 - c) * spacing)}" height="${n(th)}"/>`);
        c = c2;
      }
    }
    tex = parts.join('');
  }
  return `<defs>${clip}</defs>${base}<g fill="${dark}" clip-path="url(#${clipId})">${tex}</g>`;
}

let measureCtx: CanvasRenderingContext2D | null = null;

function fontFamily(name: string) {
  const f = LOGO_FONTS.find(f => f.name === name);
  return f ? f.family : name;
}

function textMetrics(s: State, fsize: number) {
  if (!measureCtx) measureCtx = document.createElement('canvas').getContext('2d');
  const ctx = measureCtx!;
  ctx.font = `${s.fstyle === 'Italic' ? 'italic ' : ''}${s.weight} ${fsize}px ${fontFamily(s.font)}`;
  const m = ctx.measureText(s.text);
  const w = m.width + Math.max(0, s.text.length - 1) * (+s.lspace);
  const asc = m.actualBoundingBoxAscent ?? fsize * 0.75;
  const desc = m.actualBoundingBoxDescent ?? fsize * 0.25;
  const fontAsc = m.fontBoundingBoxAscent ?? fsize * 0.8;
  const fontDesc = m.fontBoundingBoxDescent ?? fsize * 0.2;
  return { w, asc, desc, fontAsc, fontDesc };
}

function esc(t: string) {
  return t.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

const PAD = 6;

function layoutOf(s: State) {
  const p = pts(s);
  const markW = p.W + (+s.depth);
  const markH = p.H;
  const fsize = logoFontUnits(s.inFsize);
  const gap = +s.inGap;
  const tm = s.text
    ? textMetrics(s, fsize)
    : { w: 0, asc: 0, desc: 0, fontAsc: fsize * 0.8, fontDesc: fsize * 0.2 };

  const baseY = markH / 2 + (tm.asc - tm.desc) / 2 + (+s.inDy);
  const minY = Math.min(0, baseY - tm.asc) - PAD;
  const maxY = Math.max(markH, baseY + tm.desc) + PAD;

  return {
    vb: [-PAD, minY, markW + (s.text ? gap + tm.w : 0) + 2 * PAD, maxY - minY],
    textX: markW + gap,
    baseY,
    fsize,
    tm,
  };
}

export function buildLogoMark(s: State, idSuffix = '') {
  if (!import.meta.client) return '';

  const p = pts(s);
  const o = [+s.depth, 0];
  const markW = p.W + (+s.depth);
  const markH = p.H;
  const l = layoutOf(s);
  const vb = l.vb.map(n).join(' ');

  const bgEl = s.bgOn === 'Color'
    ? `\n  <rect x="${n(l.vb[0]!)}" y="${n(l.vb[1]!)}" width="${n(l.vb[2]!)}" height="${n(l.vb[3]!)}" fill="${s.bg}"/>`
    : '';

  let mark;
  if (s.pixStyle !== 'Off') {
    mark = `<g fill="${s.fill}">\n    ${pixelRects(s, markW, markH)}\n  </g>`;
  } else if (s.shade !== 'Off') {
    mark = shadedMark(s, markW, markH, idSuffix);
  } else {
    mark = `<path d="${sheetPath(s, o)}" fill="${s.fill}"/>`;
  }

  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="${vb}">${bgEl}\n  ${mark}\n</svg>`;
}

export function buildLogoLockup(s: State, scale: number, idSuffix = '') {
  if (!import.meta.client) return null;

  const l = layoutOf(s);
  const [vx, vy, vw, vh] = l.vb as [number, number, number, number];
  const spacing = (+s.lspace) * scale;
  const baselineFromTop = (l.fsize - (l.tm.fontAsc + l.tm.fontDesc)) / 2 + l.tm.fontAsc;

  return {
    size: { width: `${n2(vw * scale)}px`, height: `${n2(vh * scale)}px` },
    mark: buildLogoMark(s, idSuffix),
    text: {
      left: `${n2((l.textX - vx) * scale)}px`,
      top: `${n2((l.baseY - vy - baselineFromTop) * scale)}px`,
      fontFamily: fontFamily(s.font),
      fontSize: `${n2(l.fsize * scale)}px`,
      fontWeight: String(s.weight),
      fontStyle: s.fstyle === 'Italic' ? 'italic' : 'normal',
      letterSpacing: `${n2(spacing)}px`,
      color: s.textCol,
    },
  };
}

export function useLogoGenerator() {
  const defaults = () => Object.fromEntries(Object.entries(LOGO_PARAMS).map(([k, v]) => [k, v.value]));

  const state = useState<State>('logo-generator-state', defaults);
  const saved = useState<Record<string, State>>('logo-generator-saved', () => ({}));
  const panelOpen = useState<boolean>('logo-generator-open', () => false);
  const ready = useState<boolean>('logo-generator-ready', () => false);
  const fontTick = useState<number>('logo-generator-fonts', () => 0);

  const headerLockup = computed(() => {
    fontTick.value;
    return ready.value ? buildLogoLockup(state.value, +state.value.zoom, 'header') : null;
  });

  function persist() {
    try {
      localStorage.setItem(LOGO_STORE_KEY, JSON.stringify({ saved: saved.value, current: state.value }));
    } catch {}
  }

  function clampParam(key: string, value: any) {
    const param = LOGO_PARAMS[key];
    if (!param) return value;
    if (param.options) return param.options.includes(value) ? value : param.value;
    if (param.type || typeof value !== 'number') return value;
    return Math.min(Math.max(value, param.min), param.max);
  }

  function restore() {
    try {
      const raw = localStorage.getItem(LOGO_STORE_KEY);
      if (!raw) return;
      const data = JSON.parse(raw);
      if (data.saved) saved.value = { ...saved.value, ...data.saved };
      if (data.current) {
        for (const k in state.value) {
          if (k in data.current) state.value[k] = clampParam(k, data.current[k]);
        }
      }
    } catch {}
  }

  function reset() {
    state.value = defaults();
  }

  return { state, saved, panelOpen, ready, fontTick, headerLockup, persist, restore, reset };
}
