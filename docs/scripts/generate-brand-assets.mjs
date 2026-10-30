#!/usr/bin/env node
/**
 * Generates the Lucity logo lockups, favicons, app icons and the social
 * preview image from public/logo-mark.svg and public/logo-wordmark.svg.
 * The favicons and the touch icon are also written to the dashboard.
 *
 * Requires woff2_decompress, rsvg-convert and magick on PATH
 * (brew install woff2 librsvg imagemagick).
 *
 * Usage: node scripts/generate-brand-assets.mjs
 */

import { execFileSync } from 'node:child_process';
import { copyFileSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import opentype from 'opentype.js';

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const publicDir = join(root, 'public');
const repoDir = dirname(root);
const dashboardPublicDir = join(repoDir, 'services/dashboard/public');
const work = mkdtempSync(join(tmpdir(), 'lucity-brand-'));

const ink = '#301c0e';
const paper = '#e9dcc8';
const cream = '#f6ede0';
const markDark = '#a274ee';

const markSource = readFileSync(join(publicDir, 'logo-mark.svg'), 'utf8');
const [markWidth, markHeight] = markSource.match(/viewBox="0 0 ([\d.]+) ([\d.]+)"/).slice(1).map(Number);
const markPath = markSource.match(/ d="([^"]+)"/)[1];
const markColor = markSource.match(/fill="(#[0-9a-fA-F]{6})"/)[1];

const wordmarkSource = readFileSync(join(publicDir, 'logo-wordmark.svg'), 'utf8');
const [wordmarkTop, wordmarkWidth, wordmarkHeight] = wordmarkSource.match(/viewBox="0 (-[\d.]+) ([\d.]+) ([\d.]+)"/).slice(1).map(Number);
const wordmarkPath = wordmarkSource.match(/ d="([^"]+)"/)[1];
const wordmarkScale = markHeight / -wordmarkTop;
const wordmarkX = markWidth + markHeight * 0.61;

function font(file) {
  const source = join(work, file);
  copyFileSync(join(publicDir, 'fonts', file), source);
  execFileSync('woff2_decompress', [source]);
  const data = readFileSync(source.replace(/\.woff2$/, '.ttf'));
  return opentype.parse(data.buffer.slice(data.byteOffset, data.byteOffset + data.byteLength));
}

const n = v => +v.toFixed(2);

function svg([x, y, width, height], body) {
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="${n(x)} ${n(y)} ${n(width)} ${n(height)}" role="img" aria-label="Lucity">\n  ${body.join('\n  ')}\n</svg>\n`;
}

function mark(fill, x = 0, y = 0, scale = 1) {
  const transform = x || y || scale !== 1 ? ` transform="translate(${n(x)} ${n(y)}) scale(${+scale.toFixed(4)})"` : '';
  return `<path${transform} d="${markPath}" fill="${fill}"/>`;
}

function wordmark(fill) {
  return `<path transform="translate(${n(wordmarkX)} ${n(markHeight)}) scale(${+wordmarkScale.toFixed(5)})" d="${wordmarkPath}" fill="${fill}"/>`;
}

function pathData({ commands }) {
  return commands.map(({ type, x, y, x1, y1, x2, y2 }) => {
    if (type === 'Z') return 'Z';
    if (type === 'C') return `C${n(x1)} ${n(y1)} ${n(x2)} ${n(y2)} ${n(x)} ${n(y)}`;
    if (type === 'Q') return `Q${n(x1)} ${n(y1)} ${n(x)} ${n(y)}`;
    return `${type}${n(x)} ${n(y)}`;
  }).join('');
}

function text(face, value, x, baseline, size, fill, opacity = 1) {
  const alpha = opacity < 1 ? ` fill-opacity="${opacity}"` : '';
  return `<path d="${pathData(face.getPath(value, x, baseline, size))}" fill="${fill}"${alpha}/>`;
}

function centeredText(face, value, center, baseline, size, fill, opacity) {
  return text(face, value, center - face.getAdvanceWidth(value, size) / 2, baseline, size, fill, opacity);
}

function rasterize(source, width, height, output) {
  const input = join(work, 'raster.svg');
  writeFileSync(input, source);
  execFileSync('rsvg-convert', ['-w', String(width), '-h', String(height), input, '-o', output]);
  return output;
}

const written = [];

function target(...parts) {
  const path = join(...parts);
  written.push(path);
  return path;
}

const headlineFace = font('Redaction_20-Regular.woff2');
const sansFace = font('MonaSans-Variable.woff2');
sansFace.variation.set({ wght: 400 });

const lockupBox = [0, 0, wordmarkX + wordmarkWidth * wordmarkScale, markHeight + (wordmarkHeight + wordmarkTop) * wordmarkScale];
const lockupPngHeight = 240;

for (const [name, color, fill, readme] of [['light', ink, markColor, 'logo.svg'], ['dark', paper, markDark, 'logo-dark.svg']]) {
  const lockup = svg(lockupBox, [mark(fill), wordmark(color)]);
  writeFileSync(target(publicDir, `logo-${name}.svg`), lockup);
  writeFileSync(target(repoDir, 'assets', readme), lockup);
  rasterize(lockup, Math.round(lockupPngHeight * lockupBox[2] / lockupBox[3]), lockupPngHeight, target(publicDir, `logo-${name}.png`));
}

const faviconSide = Math.max(markWidth, markHeight);
const faviconBox = [(markWidth - faviconSide) / 2, (markHeight - faviconSide) / 2, faviconSide, faviconSide];
const faviconPlain = svg(faviconBox, [mark(markColor)]);
const faviconAdaptive = svg(faviconBox, [`<style>@media (prefers-color-scheme: dark) { path { fill: ${markDark}; } }</style>`, mark(markColor)]);
const icoSizes = [16, 32, 48].map(size => rasterize(faviconPlain, size, size, join(work, `favicon-${size}.png`)));
for (const dir of [publicDir, dashboardPublicDir]) {
  writeFileSync(target(dir, 'favicon.svg'), faviconAdaptive);
  rasterize(faviconPlain, 32, 32, target(dir, 'favicon-32x32.png'));
  rasterize(faviconPlain, 96, 96, target(dir, 'favicon-96x96.png'));
  execFileSync('magick', [...icoSizes, target(dir, 'favicon.ico')]);
}

function tile(scale) {
  const side = Math.max(markWidth, markHeight) / scale;
  return svg([0, 0, side, side], [
    `<rect width="${n(side)}" height="${n(side)}" fill="${markColor}"/>`,
    mark(cream, (side - markWidth) / 2, (side - markHeight) / 2),
  ]);
}
for (const dir of [publicDir, dashboardPublicDir]) {
  rasterize(tile(0.72), 180, 180, target(dir, 'apple-touch-icon.png'));
}
rasterize(tile(0.72), 192, 192, target(publicDir, 'web-app-manifest-192x192.png'));
rasterize(tile(0.72), 512, 512, target(publicDir, 'web-app-manifest-512x512.png'));
rasterize(tile(0.62), 512, 512, target(publicDir, 'web-app-manifest-maskable-512x512.png'));

const og = { width: 1280, height: 640, center: 640 };
const ogMarkHeight = 84, ogMarkWidth = ogMarkHeight * markWidth / markHeight;
const headline = { size: 96, lineHeight: 106 };
const ogBlock = ogMarkHeight + 62 + 0.7 * headline.size + headline.lineHeight + 76 + 8;
const ogTop = (og.height - ogBlock) / 2;
const ogFirstLine = ogTop + ogMarkHeight + 62 + 0.7 * headline.size;
const ogOverlay = svg([0, 0, og.width, og.height], [
  mark(markDark, og.center - ogMarkWidth / 2, ogTop, ogMarkHeight / markHeight),
  centeredText(headlineFace, 'Deploy anything.', og.center, ogFirstLine, headline.size, cream),
  centeredText(headlineFace, 'Own your stack.', og.center, ogFirstLine + headline.lineHeight, headline.size, cream),
  centeredText(sansFace, 'Open-source · Kubernetes-native · Ejectable', og.center, ogFirstLine + headline.lineHeight + 76, 30, cream, 0.78),
]);
execFileSync('magick', [
  join(repoDir, 'services/dashboard/assets/img/cablecar.webp'),
  '-resize', `${og.width}x${og.height}^`,
  '-gravity', 'center',
  '-extent', `${og.width}x${og.height}`,
  '-fill', 'black',
  '-colorize', '55%',
  rasterize(ogOverlay, og.width, og.height, join(work, 'og-overlay.png')),
  '-composite',
  '-quality', '86',
  target(publicDir, 'img/og.jpg'),
]);

rmSync(work, { recursive: true, force: true });
console.log(written.map(path => `wrote ${path.replace(`${repoDir}/`, '')}`).join('\n'));
