/**
 * One-off script to convert "(but don't want to)" in Caveat Bold
 * into per-character SVG center-line paths for handwriting animation.
 *
 * Uses opentype.js for glyph outlines, then flo-mat (Medial Axis Transform)
 * to extract center-line skeleton paths that can be animated with
 * stroke-dashoffset for a realistic handwriting effect.
 *
 * Usage: node docs/scripts/generate-handwriting-paths.mjs
 *
 * Output: JSON array ready to paste into the Vue component.
 */

import opentype from 'opentype.js';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import {
  findMats,
  getPathsFromStr,
  toScaleAxis,
  CpNodeFs,
} from 'flo-mat';

const __dirname = dirname(fileURLToPath(import.meta.url));
const fontPath = join(__dirname, 'Caveat-Bold.ttf');

const text = '(but don\'t want to)';
const fontSize = 28;

// Load the font
const font = opentype.loadSync(fontPath);

// Get per-character paths with proper kerning
const glyphs = font.stringToGlyphs(text);
let x = 0;
const y = fontSize; // baseline

const chars = [];

for (let i = 0; i < glyphs.length; i++) {
  const glyph = glyphs[i];
  const path = glyph.getPath(x, y, fontSize);
  const outlineD = path.toPathData(2);

  // Get kerning between this glyph and next
  let kerning = 0;
  if (i < glyphs.length - 1) {
    kerning = font.getKerningValue(glyph, glyphs[i + 1]);
  }

  const advanceWidth = (glyph.advanceWidth / font.unitsPerEm) * fontSize;

  // Extract center-line via Medial Axis Transform
  let centerlineD = '';
  let strokeWidth = 2;

  if (outlineD && text[i] !== ' ') {
    try {
      const bezierLoops = getPathsFromStr(outlineD);
      const mats = findMats(bezierLoops, {
        maxCurviness: 0.4,
        maxLength: 8,
      });

      if (mats.length > 0) {
        // Apply Scale Axis Transform to simplify (remove tiny branches)
        const sat = toScaleAxis(mats[0], 1.5);

        // Extract skeleton curves and maximal disk radii
        const curves = [];
        const radii = [];

        CpNodeFs.traverseEdges(sat.cpNode, (cpNode) => {
          if (CpNodeFs.isTerminating(cpNode)) return;

          // Get the MAT curve from this node to the next
          const curve = CpNodeFs.getMatCurveToNext(cpNode);
          if (curve && curve.length >= 2) {
            curves.push(curve);
            // Record maximal disk radius (= half stroke width)
            radii.push(cpNode.cp.circle.radius);
          }
        });

        if (curves.length > 0) {
          centerlineD = curvesToSvgPath(curves);
          // Average radius gives us half the stroke width
          const avgRadius = radii.reduce((a, b) => a + b, 0) / radii.length;
          strokeWidth = Math.round(avgRadius * 2 * 100) / 100;
        }
      }
    } catch (err) {
      console.error(`  [WARN] FloMat failed for '${text[i]}': ${err.message}`);
      // Fall back to outline path
      centerlineD = '';
    }
  }

  chars.push({
    char: text[i],
    outlineD,
    centerlineD,
    strokeWidth,
    width: Math.round(advanceWidth * 100) / 100,
  });

  x += advanceWidth + (kerning / font.unitsPerEm) * fontSize;
}

// Total width for viewBox
const totalWidth = Math.ceil(x) + 5;
const totalHeight = Math.ceil(fontSize * 1.3);

// Output results
console.log('// Paste this into index.vue <script setup>\n');
console.log(`const ASIDE_VIEWBOX = '0 0 ${totalWidth} ${totalHeight}';`);
console.log('');
console.log('const ASIDE_PATHS = [');
for (const c of chars) {
  const d = c.centerlineD || c.outlineD;
  const hasCenterline = !!c.centerlineD;
  if (d) {
    console.log(`  { char: ${JSON.stringify(c.char)}, d: '${d}', strokeWidth: ${c.strokeWidth}, centerline: ${hasCenterline} },`);
  } else {
    console.log(`  { char: ${JSON.stringify(c.char)}, d: '' },`);
  }
}
console.log('];');

// Summary
const total = chars.filter(c => c.char !== ' ').length;
const extracted = chars.filter(c => c.centerlineD).length;
console.log(`\n// ${extracted}/${total} characters got center-line paths`);

/* ── Helpers ── */

/**
 * Convert an array of bezier curves to an SVG path string.
 * Each curve is [[x0,y0], [x1,y1], ...] with 2-4 control points.
 */
function curvesToSvgPath(curves) {
  if (!curves.length) return '';

  const parts = [];
  let lastPoint = null;

  for (const curve of curves) {
    const start = curve[0];

    // Move to start if not continuous with previous curve
    if (!lastPoint || dist(lastPoint, start) > 0.5) {
      parts.push(`M${r(start[0])} ${r(start[1])}`);
    }

    if (curve.length === 2) {
      // Line
      const [, p1] = curve;
      parts.push(`L${r(p1[0])} ${r(p1[1])}`);
      lastPoint = p1;
    } else if (curve.length === 3) {
      // Quadratic bezier
      const [, cp, p2] = curve;
      parts.push(`Q${r(cp[0])} ${r(cp[1])} ${r(p2[0])} ${r(p2[1])}`);
      lastPoint = p2;
    } else if (curve.length === 4) {
      // Cubic bezier
      const [, cp1, cp2, p3] = curve;
      parts.push(`C${r(cp1[0])} ${r(cp1[1])} ${r(cp2[0])} ${r(cp2[1])} ${r(p3[0])} ${r(p3[1])}`);
      lastPoint = p3;
    }
  }

  return parts.join('');
}

function r(n) {
  return Math.round(n * 100) / 100;
}

function dist(a, b) {
  return Math.hypot(a[0] - b[0], a[1] - b[1]);
}
