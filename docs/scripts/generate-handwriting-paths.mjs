/**
 * One-off script to convert "(but don't want to)" in Caveat Bold
 * into per-character SVG path data for handwriting animation.
 *
 * Usage: node docs/scripts/generate-handwriting-paths.mjs
 *
 * Output: JSON array of { char, d, width } objects ready to paste
 * into the Vue component.
 */

import opentype from 'opentype.js';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __dirname = dirname(fileURLToPath(import.meta.url));
const fontPath = join(__dirname, 'Caveat-Bold.ttf');

const text = '(but don\'t want to)';
const fontSize = 28;

// Load the font (variable font — Caveat[wght].ttf)
const font = opentype.loadSync(fontPath);

// Get per-character paths with proper kerning
const glyphs = font.stringToGlyphs(text);
let x = 0;
const y = fontSize; // baseline

const chars = [];

for (let i = 0; i < glyphs.length; i++) {
  const glyph = glyphs[i];
  const path = glyph.getPath(x, y, fontSize);
  const d = path.toPathData(2);

  // Get kerning between this glyph and next
  let kerning = 0;
  if (i < glyphs.length - 1) {
    kerning = font.getKerningValue(glyph, glyphs[i + 1]);
  }

  const advanceWidth = (glyph.advanceWidth / font.unitsPerEm) * fontSize;

  chars.push({
    char: text[i],
    d,
    width: Math.round(advanceWidth * 100) / 100,
  });

  x += advanceWidth + (kerning / font.unitsPerEm) * fontSize;
}

// Total width for viewBox
const totalWidth = Math.ceil(x) + 5;
const totalHeight = Math.ceil(fontSize * 1.3);

console.log('// Paste this into index.vue <script setup>\n');
console.log(`const ASIDE_VIEWBOX = '0 0 ${totalWidth} ${totalHeight}';`);
console.log('');
console.log('const ASIDE_PATHS = [');
for (const c of chars) {
  if (c.d) {
    console.log(`  { char: ${JSON.stringify(c.char)}, d: '${c.d}' },`);
  }
}
console.log('];');
