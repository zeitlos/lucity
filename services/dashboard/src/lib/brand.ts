import { LOGO_HEIGHT, LOGO_PATH, LOGO_WIDTH, WORDMARK_CAP, WORDMARK_PATH, WORDMARK_WIDTH } from '@/lib/logo';

export const PURPLE = '#8843db';
export const LILAC = '#a274ee';
export const INK = '#301c0e';
export const PAPER = '#e9dcc8';
export const CREAM = '#f6ede0';
export const ESPRESSO = '#211208';
export const GREEN = '#00cf85';

const WORDMARK_DEPTH = 200;
const LOCKUP_GAP = 0.61;
const FLAT_ICON_MARK = 0.72;

export interface BrandAsset {
  file: string;
  width: number;
  height: number;
  viewBox: string;
  body: string;
}

const round = (value: number) => +value.toFixed(2);

export function brandSvg(asset: BrandAsset, width = asset.width, height = asset.height) {
  return `<svg xmlns="http://www.w3.org/2000/svg" width="${round(width)}" height="${round(height)}" viewBox="${asset.viewBox}" role="img" aria-label="Lucity">${asset.body}</svg>`;
}

export function brandSvgUrl(asset: BrandAsset, width?: number, height?: number) {
  return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(brandSvg(asset, width, height))}`;
}

function save(url: string, filename: string) {
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  link.click();
}

export function downloadSvg(asset: BrandAsset) {
  save(brandSvgUrl(asset), `${asset.file}.svg`);
}

export async function downloadPng(asset: BrandAsset, width: number) {
  const height = Math.round(width * asset.height / asset.width);
  const image = new Image();
  image.src = brandSvgUrl(asset, width, height);
  await image.decode();
  const canvas = document.createElement('canvas');
  canvas.width = width;
  canvas.height = height;
  canvas.getContext('2d')?.drawImage(image, 0, 0, width, height);
  save(canvas.toDataURL('image/png'), `${asset.file}.png`);
}

function markPath(fill: string, x = 0, y = 0, scale = 1) {
  const transform = x || y || scale !== 1 ? ` transform="translate(${round(x)} ${round(y)}) scale(${round(scale)})"` : '';
  return `<path${transform} d="${LOGO_PATH}" fill="${fill}"/>`;
}

export function markAsset(file: string, fill: string): BrandAsset {
  return {
    file,
    width: LOGO_WIDTH,
    height: LOGO_HEIGHT,
    viewBox: `0 0 ${LOGO_WIDTH} ${LOGO_HEIGHT}`,
    body: markPath(fill),
  };
}

export function wordmarkAsset(file: string, fill: string): BrandAsset {
  return {
    file,
    width: WORDMARK_WIDTH,
    height: WORDMARK_CAP + WORDMARK_DEPTH,
    viewBox: `0 -${WORDMARK_CAP} ${WORDMARK_WIDTH} ${WORDMARK_CAP + WORDMARK_DEPTH}`,
    body: `<path d="${WORDMARK_PATH}" fill="${fill}"/>`,
  };
}

const wordmarkScale = LOGO_HEIGHT / WORDMARK_CAP;
const wordmarkX = LOGO_WIDTH + LOGO_HEIGHT * LOCKUP_GAP;

export const LOCKUP_WIDTH = round(wordmarkX + WORDMARK_WIDTH * wordmarkScale);
export const LOCKUP_HEIGHT = round(LOGO_HEIGHT + WORDMARK_DEPTH * wordmarkScale);
export const LOCKUP_WORDMARK_TRANSFORM = `translate(${round(wordmarkX)} ${LOGO_HEIGHT}) scale(${+wordmarkScale.toFixed(5)})`;

export function lockupAsset(file: string, markFill: string, wordmarkFill: string): BrandAsset {
  return {
    file,
    width: LOCKUP_WIDTH,
    height: LOCKUP_HEIGHT,
    viewBox: `0 0 ${LOCKUP_WIDTH} ${LOCKUP_HEIGHT}`,
    body: markPath(markFill)
      + `<path transform="${LOCKUP_WORDMARK_TRANSFORM}" d="${WORDMARK_PATH}" fill="${wordmarkFill}"/>`,
  };
}

export interface IconTheme {
  name: string;
  file: string;
  tile: [string, string, string];
  mark: string;
  face: [string, string, string, string, string];
  edge: string;
  shadow: string;
  shadowOpacity: number;
  inset: number;
}

export const iconThemes: IconTheme[] = [
  {
    name: 'Cream on espresso',
    file: 'espresso',
    tile: ['#433229', '#2a1e18', '#1a1310'],
    mark: CREAM,
    face: ['#fffbf4', '#fbf2e3', '#e3cfb1', '#efe2cc', '#f4e9d6'],
    edge: '#5a3b1c',
    shadow: '#0c0705',
    shadowOpacity: 1,
    inset: 0.35,
  },
  {
    name: 'Cream on purple',
    file: 'purple',
    tile: ['#a06ae9', PURPLE, '#6a2cb8'],
    mark: CREAM,
    face: ['#fffbf4', '#fbf2e3', '#e3cfb1', '#efe2cc', '#f4e9d6'],
    edge: '#5a3b1c',
    shadow: '#2a0b52',
    shadowOpacity: 1,
    inset: 0.25,
  },
  {
    name: 'Purple on cream',
    file: 'cream',
    tile: ['#fffaf3', CREAM, '#e7d9c4'],
    mark: PURPLE,
    face: ['#ad80f2', '#9455e4', '#6f2fc2', '#8240d8', PURPLE],
    edge: '#2a0b52',
    shadow: '#5a3b1c',
    shadowOpacity: 0.5,
    inset: 0.08,
  },
  {
    name: 'Lilac on espresso',
    file: 'lilac',
    tile: ['#433229', '#2a1e18', '#1a1310'],
    mark: LILAC,
    face: ['#d2b6fa', '#b48ef3', '#8655d6', '#9a6deb', LILAC],
    edge: '#3a1670',
    shadow: '#0c0705',
    shadowOpacity: 1,
    inset: 0.35,
  },
];

function squircle(center = 512, radius = 412, exponent = 5, steps = 360) {
  const points: string[] = [];
  for (let i = 0; i < steps; i++) {
    const angle = (i / steps) * Math.PI * 2;
    const cos = Math.cos(angle);
    const sin = Math.sin(angle);
    const x = center + radius * Math.sign(cos) * Math.abs(cos) ** (2 / exponent);
    const y = center + radius * Math.sign(sin) * Math.abs(sin) ** (2 / exponent);
    points.push(`${x.toFixed(1)} ${y.toFixed(1)}`);
  }
  return `M ${points.join(' L ')} Z`;
}

const SQUIRCLE = squircle();

export function appIconAsset(theme: IconTheme): BrandAsset {
  const [tileTop, tileMiddle, tileBottom] = theme.tile;
  const [face0, face1, face2, face3, face4] = theme.face;
  return {
    file: `lucity-app-icon-${theme.file}`,
    width: 1024,
    height: 1024,
    viewBox: '0 0 1024 1024',
    body: `<defs>
<path id="body" d="${SQUIRCLE}"/>
<path id="bend" d="${LOGO_PATH}"/>
<clipPath id="bodyClip"><use href="#body"/></clipPath>
<clipPath id="markClip"><use href="#bend"/></clipPath>
<filter id="iconShadow" x="-20%" y="-20%" width="140%" height="140%">
<feDropShadow dx="0" dy="2" stdDeviation="2" flood-color="#000" flood-opacity="0.3"/>
<feDropShadow dx="0" dy="12" stdDeviation="14" flood-color="#000" flood-opacity="0.24"/>
</filter>
<linearGradient id="tile" x1="0" y1="100" x2="0" y2="924" gradientUnits="userSpaceOnUse">
<stop offset="0" stop-color="${tileTop}"/>
<stop offset="0.5" stop-color="${tileMiddle}"/>
<stop offset="1" stop-color="${tileBottom}"/>
</linearGradient>
<radialGradient id="sheen" cx="512" cy="60" r="700" gradientUnits="userSpaceOnUse">
<stop offset="0" stop-color="#fff" stop-opacity="0.12"/>
<stop offset="1" stop-color="#fff" stop-opacity="0"/>
</radialGradient>
<linearGradient id="rimV" x1="0" y1="100" x2="0" y2="924" gradientUnits="userSpaceOnUse">
<stop offset="0" stop-color="#fff" stop-opacity="0.62"/>
<stop offset="0.12" stop-color="#fff" stop-opacity="0.26"/>
<stop offset="0.5" stop-color="#fff" stop-opacity="0.1"/>
<stop offset="0.88" stop-color="#fff" stop-opacity="0.16"/>
<stop offset="1" stop-color="#fff" stop-opacity="0.36"/>
</linearGradient>
<linearGradient id="rimD" x1="100" y1="100" x2="924" y2="924" gradientUnits="userSpaceOnUse">
<stop offset="0" stop-color="#fff" stop-opacity="0.34"/>
<stop offset="0.3" stop-color="#fff" stop-opacity="0"/>
<stop offset="0.75" stop-color="#fff" stop-opacity="0"/>
<stop offset="1" stop-color="#fff" stop-opacity="0.18"/>
</linearGradient>
<filter id="soft6" x="-10%" y="-10%" width="120%" height="120%"><feGaussianBlur stdDeviation="6"/></filter>
<filter id="soft14" x="-10%" y="-10%" width="120%" height="120%"><feGaussianBlur stdDeviation="14"/></filter>
<filter id="markShadow" x="-30%" y="-30%" width="160%" height="180%">
<feDropShadow dx="0" dy="0.9" stdDeviation="0.9" flood-color="${theme.shadow}" flood-opacity="${round(0.5 * theme.shadowOpacity)}"/>
<feDropShadow dx="0" dy="6.4" stdDeviation="8.2" flood-color="${theme.shadow}" flood-opacity="${round(0.4 * theme.shadowOpacity)}"/>
</filter>
<linearGradient id="face" x1="0" y1="0" x2="0" y2="181" gradientUnits="userSpaceOnUse">
<stop offset="0" stop-color="${face0}"/>
<stop offset="0.38" stop-color="${face1}"/>
<stop offset="0.5" stop-color="${face2}"/>
<stop offset="0.8" stop-color="${face3}"/>
<stop offset="1" stop-color="${face4}"/>
</linearGradient>
<linearGradient id="edgeHi" x1="40" y1="0" x2="150" y2="181" gradientUnits="userSpaceOnUse">
<stop offset="0" stop-color="#fff" stop-opacity="1"/>
<stop offset="0.45" stop-color="#fff" stop-opacity="0.35"/>
<stop offset="1" stop-color="#fff" stop-opacity="0.15"/>
</linearGradient>
<linearGradient id="edgeLo" x1="0" y1="90" x2="0" y2="181" gradientUnits="userSpaceOnUse">
<stop offset="0" stop-color="${theme.edge}" stop-opacity="0"/>
<stop offset="1" stop-color="${theme.edge}" stop-opacity="0.28"/>
</linearGradient>
</defs>
<g filter="url(#iconShadow)"><use href="#body" fill="url(#tile)"/></g>
<g clip-path="url(#bodyClip)">
<rect width="1024" height="1024" fill="url(#sheen)"/>
<use href="#body" fill="none" stroke="#000" stroke-opacity="${theme.inset}" stroke-width="40" filter="url(#soft14)"/>
<use href="#body" fill="none" stroke="url(#rimV)" stroke-width="9"/>
<use href="#body" fill="none" stroke="url(#rimD)" stroke-width="9"/>
<use href="#body" fill="none" stroke="#fff" stroke-opacity="0.07" stroke-width="30" filter="url(#soft6)"/>
<g transform="translate(237 306.9) scale(2.2)">
<use href="#bend" fill="url(#face)" filter="url(#markShadow)"/>
<g clip-path="url(#markClip)">
<use href="#bend" fill="none" stroke="url(#edgeLo)" stroke-width="6.24"/>
<use href="#bend" fill="none" stroke="url(#edgeHi)" stroke-width="2.6"/>
</g>
</g>
</g>`,
  };
}

export function flatIconAsset(theme: IconTheme): BrandAsset {
  const side = round(Math.max(LOGO_WIDTH, LOGO_HEIGHT) / FLAT_ICON_MARK);
  return {
    file: `lucity-icon-${theme.file}`,
    width: side,
    height: side,
    viewBox: `0 0 ${side} ${side}`,
    body: `<rect width="${side}" height="${side}" fill="${theme.tile[1]}"/>`
      + markPath(theme.mark, (side - LOGO_WIDTH) / 2, (side - LOGO_HEIGHT) / 2),
  };
}
