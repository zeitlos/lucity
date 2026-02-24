<script setup lang="ts">
import { computed, getCurrentInstance } from 'vue';
import { cn } from '@/lib/utils';

type Point = [number, number];

const props = withDefaults(defineProps<{
  size?: number;
  debug?: boolean;
  variant?: 'default' | 'mark';
  class?: string;
}>(), {
  size: 40,
  debug: false,
  variant: 'default',
});

const CELL = 20;

// L rotated so corner points down in isometric view.
// Vertical stroke goes up-left, horizontal stroke goes up-right.
//
// Grid layout:
// [ ] [ ] [X]    row 0
// [ ] [ ] [X]    row 1
// [X] [X] [X]    row 2
//
// Corner at (2,2) → isometric (0, 2*CELL) = bottom center
const L_CELLS: Point[] = [
  [0, 2],
  [1, 2],
  [2, 2],
  [2, 1],
];

const TRIANGLE_VERTICES: Point[] = [
  [0, 1],
  [1, 0],
  [1, 1],
];

const uid = getCurrentInstance()?.uid ?? Math.random().toString(36).slice(2, 8);

function project(x: number, y: number): Point {
  return [
    (x - y) * CELL * 0.866,
    (x + y) * CELL * 0.5,
  ];
}

function pts(corners: Point[]): string {
  return corners.map(([x, y]) => `${x},${y}`).join(' ');
}

// Size-dependent optical corrections.
// At small sizes the triangle becomes hard to see, so we scale it up
// from its centroid. At large sizes we use standard proportions.
const triangleScale = computed(() => {
  if (props.size <= 32) return 2.0;
  if (props.size <= 64) return 1.4;
  return 1.0;
});

// Grid line stroke width: thinner at very large sizes for refinement.
const gridStrokeWidth = computed(() => {
  if (props.size >= 128) return 0.35;
  return 0.5;
});

// Whether to show gradient fills on tiles (large sizes only).
const showTileGradients = computed(() => props.size >= 96);

// Grid lines at detail sizes (≥96px) or in debug mode. Hidden at standard
// sizes where the scaled triangle no longer aligns with the grid.
const showGrid = computed(() => props.size >= 96 || props.debug);

const tiles = computed(() =>
  L_CELLS.map(([c, r]) => ({
    points: pts([
      project(c, r),
      project(c + 1, r),
      project(c + 1, r + 1),
      project(c, r + 1),
    ]),
  })),
);

// Scale triangle vertices outward from their centroid for small sizes.
function scaleFromCentroid(vertices: Point[], scale: number): Point[] {
  const cx = vertices.reduce((s, p) => s + p[0], 0) / vertices.length;
  const cy = vertices.reduce((s, p) => s + p[1], 0) / vertices.length;
  return vertices.map(([x, y]) => [
    cx + (x - cx) * scale,
    cy + (y - cy) * scale,
  ] as Point);
}

// Vertical nudge for the bold triangle to maintain consistent gap to the L.
// Scaling from centroid pushes the bottom edge closer to the tiles, so we
// shift the whole triangle upward to compensate.
const triangleNudgeY = computed(() => {
  if (props.size <= 32) return -3;
  if (props.size <= 64) return -1.5;
  return 0;
});

const triangle = computed(() => {
  const projected = TRIANGLE_VERTICES.map(([x, y]) => project(x, y));
  const scaled = triangleScale.value !== 1.0
    ? scaleFromCentroid(projected, triangleScale.value)
    : projected;
  const nudge = triangleNudgeY.value;
  const nudged = nudge !== 0
    ? scaled.map(([x, y]) => [x, y + nudge] as Point)
    : scaled;
  return { points: pts(nudged) };
});

// Bounding box of the tile gradient (top-left to bottom-right of all tiles).
const tileBounds = computed(() => {
  const allPts = L_CELLS.flatMap(([c, r]) => [
    project(c, r),
    project(c + 1, r),
    project(c + 1, r + 1),
    project(c, r + 1),
  ]);
  const xs = allPts.map(p => p[0]);
  const ys = allPts.map(p => p[1]);
  return {
    x1: Math.min(...xs),
    y1: Math.min(...ys),
    x2: Math.max(...xs),
    y2: Math.max(...ys),
  };
});

// Triangle gradient bounds.
const triangleBounds = computed(() => {
  const projected = TRIANGLE_VERTICES.map(([x, y]) => project(x, y));
  const xs = projected.map(p => p[0]);
  const ys = projected.map(p => p[1]);
  return {
    x1: Math.min(...xs),
    y1: Math.min(...ys),
    x2: Math.max(...xs),
    y2: Math.max(...ys),
  };
});

const EXT = 0.4;

const gridLines = computed(() => {
  const lines: { x1: number; y1: number; x2: number; y2: number; id: string }[] = [];

  for (let r = 0; r <= 3; r++) {
    const [x1, y1] = project(-EXT, r);
    const [x2, y2] = project(3 + EXT, r);
    lines.push({ x1, y1, x2, y2, id: `h${r}` });
  }

  for (let c = 0; c <= 3; c++) {
    const [x1, y1] = project(c, -EXT);
    const [x2, y2] = project(c, 3 + EXT);
    lines.push({ x1, y1, x2, y2, id: `v${c}` });
  }

  return lines;
});

// ViewBox always includes grid line extent so the logo is the same
// size whether or not grid lines are rendered.
const viewBox = computed(() => {
  const allPts: Point[] = [
    ...L_CELLS.flatMap(([c, r]) => [
      project(c, r),
      project(c + 1, r),
      project(c + 1, r + 1),
      project(c, r + 1),
    ]),
    ...TRIANGLE_VERTICES.map(([x, y]) => project(x, y)),
    ...gridLines.value.flatMap(l => [[l.x1, l.y1] as Point, [l.x2, l.y2] as Point]),
  ];

  const xs = allPts.map(p => p[0]);
  const ys = allPts.map(p => p[1]);
  const pad = props.debug ? 6 : 3;
  const minX = Math.min(...xs) - pad;
  const minY = Math.min(...ys) - pad;
  const w = Math.max(...xs) - Math.min(...xs) + pad * 2;
  const h = Math.max(...ys) - Math.min(...ys) + pad * 2;

  return { str: `${minX} ${minY} ${w} ${h}`, w, h };
});

const svgHeight = computed(() => Math.round(props.size * viewBox.value.h / viewBox.value.w));

const gridLabels = computed(() => {
  if (!props.debug) return [];
  const labels: { x: number; y: number; text: string }[] = [];
  for (let c = 0; c <= 3; c++) {
    for (let r = 0; r <= 3; r++) {
      const [px, py] = project(c, r);
      labels.push({ x: px, y: py, text: `${c},${r}` });
    }
  }
  return labels;
});

function gradId(lineId: string): string {
  return `fade-${uid}-${lineId}`;
}

// Size-dependent circle-to-logo ratio for the mark.
// Small sizes: the logo overflows the circle, clipping the L arms and
// triangle tips against the circle edge to form a distinctive silhouette
// with better legibility. Large sizes: fully contained with padding.
const markRadiusFactor = computed(() => {
  if (props.size <= 16) return 0.68;
  if (props.size <= 24) return 0.75;
  if (props.size <= 32) return 0.84;
  if (props.size <= 48) return 0.94;
  if (props.size <= 64) return 1.05;
  return 1.15;
});

// Mark variant: circle with logo cut out as negative space.
// The circle's central axis is aligned with the L corner at grid(2,2),
// which projects to x=0 in isometric space. The triangle's lowest
// vertex also sits at x=0, so both key points land on the axis.
const mark = computed(() => {
  const allPts: Point[] = L_CELLS.flatMap(([c, r]) => [
    project(c, r),
    project(c + 1, r),
    project(c + 1, r + 1),
    project(c, r + 1),
  ]);

  // Use small-size triangle (scale=2.0, nudge=-3) for the mark since
  // it's always rendered at compact sizes.
  const triProjected = TRIANGLE_VERTICES.map(([x, y]) => project(x, y));
  const triScaled = scaleFromCentroid(triProjected, 2.0);
  const triNudged = triScaled.map(([x, y]) => [x, y - 3] as Point);
  triNudged.forEach(p => allPts.push(p));

  const ys = allPts.map(p => p[1]);

  // Align the circle's central axis with the L corner at grid(2,2).
  const cx = 0;
  const cy = (Math.min(...ys) + Math.max(...ys)) / 2;

  let maxDist = 0;
  for (const [px, py] of allPts) {
    const d = Math.sqrt((px - cx) ** 2 + (py - cy) ** 2);
    if (d > maxDist) maxDist = d;
  }
  const r = maxDist * markRadiusFactor.value;

  return {
    cx, cy, r,
    viewBox: `${cx - r} ${cy - r} ${r * 2} ${r * 2}`,
    tiles: L_CELLS.map(([c, row]) => pts([
      project(c, row),
      project(c + 1, row),
      project(c + 1, row + 1),
      project(c, row + 1),
    ])),
    triangle: pts(triNudged),
    lCorner: project(2, 2) as Point,
    triBottom: triNudged[2],
  };
});
</script>

<template>
  <!-- Mark variant: circle with logo knocked out -->
  <svg
    v-if="variant === 'mark'"
    :width="props.size"
    :height="props.size"
    :viewBox="mark.viewBox"
    :class="cn('inline-block', props.class)"
    xmlns="http://www.w3.org/2000/svg"
    role="img"
    aria-label="Lucity logo"
  >
    <defs>
      <mask :id="`mark-mask-${uid}`">
        <circle :cx="mark.cx" :cy="mark.cy" :r="mark.r" fill="white" />
        <polygon
          v-for="(tilePts, i) in mark.tiles"
          :key="'mt-' + i"
          :points="tilePts"
          fill="black"
          stroke="black"
          stroke-width="0.5"
          stroke-linejoin="round"
        />
        <polygon :points="mark.triangle" fill="black" />
      </mask>
    </defs>
    <circle
      :cx="mark.cx"
      :cy="mark.cy"
      :r="mark.r"
      fill="currentColor"
      :mask="`url(#mark-mask-${uid})`"
    />
    <!-- Construction grid for alignment verification -->
    <g v-if="props.debug">
      <!-- Vertical center axis -->
      <line
        :x1="mark.cx" :y1="mark.cy - mark.r"
        :x2="mark.cx" :y2="mark.cy + mark.r"
        stroke="#ff4444" stroke-width="0.4" stroke-dasharray="3 3" opacity="0.7"
      />
      <!-- Horizontal center axis -->
      <line
        :x1="mark.cx - mark.r" :y1="mark.cy"
        :x2="mark.cx + mark.r" :y2="mark.cy"
        stroke="#ff4444" stroke-width="0.4" stroke-dasharray="3 3" opacity="0.7"
      />
      <!-- Circle center -->
      <circle :cx="mark.cx" :cy="mark.cy" r="1.2" fill="#ff4444" opacity="0.8" />
      <!-- L corner — should sit on vertical axis -->
      <circle :cx="mark.lCorner[0]" :cy="mark.lCorner[1]" r="1.5" fill="#ff4444" />
      <!-- Triangle lowest vertex — should sit on vertical axis -->
      <circle :cx="mark.triBottom[0]" :cy="mark.triBottom[1]" r="1.5" fill="#4488ff" />
    </g>
  </svg>

  <!-- Default variant: colored tiles + triangle -->
  <svg
    v-else
    :width="props.size"
    :height="svgHeight"
    :viewBox="viewBox.str"
    :class="cn('inline-block', props.class)"
    xmlns="http://www.w3.org/2000/svg"
    role="img"
    aria-label="Lucity logo"
  >
    <defs>
      <!-- Grid line fade gradients (detail sizes only) -->
      <template v-if="showGrid">
        <linearGradient
          v-for="line in gridLines"
          :key="'grad-' + line.id"
          :id="gradId(line.id)"
          gradientUnits="userSpaceOnUse"
          :x1="line.x1"
          :y1="line.y1"
          :x2="line.x2"
          :y2="line.y2"
        >
          <stop
            offset="0%"
            stop-color="currentColor"
            stop-opacity="0"
          />
          <stop
            offset="15%"
            stop-color="currentColor"
            stop-opacity="0.25"
          />
          <stop
            offset="85%"
            stop-color="currentColor"
            stop-opacity="0.25"
          />
          <stop
            offset="100%"
            stop-color="currentColor"
            stop-opacity="0"
          />
        </linearGradient>
      </template>

      <!-- Tile depth gradient for large sizes -->
      <template v-if="showTileGradients">
        <linearGradient
          :id="`tile-depth-${uid}`"
          gradientUnits="userSpaceOnUse"
          :x1="tileBounds.x1"
          :y1="tileBounds.y1"
          :x2="tileBounds.x1"
          :y2="tileBounds.y2"
        >
          <stop
            offset="0%"
            class="tile-grad-start"
            stop-opacity="1"
          />
          <stop
            offset="100%"
            class="tile-grad-end"
            stop-opacity="1"
          />
        </linearGradient>

        <linearGradient
          :id="`tri-depth-${uid}`"
          gradientUnits="userSpaceOnUse"
          :x1="triangleBounds.x1"
          :y1="triangleBounds.y1"
          :x2="triangleBounds.x1"
          :y2="triangleBounds.y2"
        >
          <stop
            offset="0%"
            class="tri-grad-start"
            stop-opacity="1"
          />
          <stop
            offset="100%"
            class="tri-grad-end"
            stop-opacity="1"
          />
        </linearGradient>
      </template>
    </defs>

    <g
      v-if="showGrid"
      class="text-muted-foreground"
    >
      <line
        v-for="line in gridLines"
        :key="line.id"
        :x1="line.x1"
        :y1="line.y1"
        :x2="line.x2"
        :y2="line.y2"
        :stroke="`url(#${gradId(line.id)})`"
        :stroke-width="gridStrokeWidth"
        fill="none"
      />
    </g>

    <g>
      <polygon
        v-for="(tile, i) in tiles"
        :key="'tile-' + i"
        :points="tile.points"
        :fill="showTileGradients ? `url(#tile-depth-${uid})` : undefined"
        :stroke="showTileGradients ? `url(#tile-depth-${uid})` : undefined"
        :class="{ 'tile-flat': !showTileGradients }"
        stroke-width="0.5"
        stroke-linejoin="round"
      />
    </g>

    <g class="triangle-group">
      <polygon
        :points="triangle.points"
        :fill="showTileGradients ? `url(#tri-depth-${uid})` : undefined"
        :class="{ 'logo-triangle': !showTileGradients }"
        stroke="none"
      />
    </g>

    <g v-if="props.debug">
      <template
        v-for="label in gridLabels"
        :key="label.text"
      >
        <circle
          :cx="label.x"
          :cy="label.y"
          r="0.8"
          class="fill-destructive"
        />
        <text
          :x="label.x + 1.5"
          :y="label.y - 1"
          font-size="2.5"
          font-family="var(--font-mono)"
          class="fill-muted-foreground"
        >
          {{ label.text }}
        </text>
      </template>
    </g>
  </svg>
</template>

<style scoped>
/* Matching stroke eliminates anti-aliasing gaps between adjacent tiles. */
.tile-flat {
  fill: var(--primary);
  stroke: var(--primary);
}

.logo-triangle {
  fill: var(--accent);
}

:global(.dark) .logo-triangle {
  fill: var(--accent-foreground);
}

.triangle-group {
  transition: transform 0.3s ease;
}

/* Tile depth gradient: primary lightened → primary darkened */
.tile-grad-start {
  stop-color: oklch(from var(--primary) calc(l + 0.06) c h);
}

.tile-grad-end {
  stop-color: oklch(from var(--primary) calc(l - 0.06) c h);
}

/* Triangle depth gradient: accent lightened → accent */
.tri-grad-start {
  stop-color: oklch(from var(--accent) calc(l + 0.06) c h);
}

.tri-grad-end {
  stop-color: oklch(from var(--accent) calc(l - 0.04) c h);
}

:global(.dark) .tri-grad-start {
  stop-color: oklch(from var(--accent-foreground) calc(l + 0.06) c h);
}

:global(.dark) .tri-grad-end {
  stop-color: oklch(from var(--accent-foreground) calc(l - 0.04) c h);
}
</style>
