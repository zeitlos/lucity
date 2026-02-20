<script setup lang="ts">
import { onMounted, ref } from 'vue';

const canvasEl = ref<HTMLCanvasElement | null>(null);

function generateTopo(canvas: HTMLCanvasElement) {
  const rect = canvas.getBoundingClientRect();
  const w = Math.floor(rect.width);
  const h = Math.floor(rect.height);
  if (w === 0 || h === 0) return;

  canvas.width = w;
  canvas.height = h;
  const ctx = canvas.getContext('2d');
  if (!ctx) return;

  // Permutation table
  const perm = new Uint8Array(512);
  const p = new Uint8Array(256);
  for (let i = 0; i < 256; i++) p[i] = i;
  for (let i = 255; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [p[i], p[j]] = [p[j], p[i]];
  }
  for (let i = 0; i < 512; i++) perm[i] = p[i & 255];

  const grad = [
    [1, 1], [-1, 1], [1, -1], [-1, -1],
    [1, 0], [-1, 0], [0, 1], [0, -1],
  ];
  function dot(gi: number, x: number, y: number) {
    const g = grad[gi % 8];
    return g[0] * x + g[1] * y;
  }
  function fade(t: number) { return t * t * t * (t * (t * 6 - 15) + 10); }
  function lerp(a: number, b: number, t: number) { return a + t * (b - a); }

  function noise(x: number, y: number) {
    const xi = Math.floor(x) & 255;
    const yi = Math.floor(y) & 255;
    const xf = x - Math.floor(x);
    const yf = y - Math.floor(y);
    const u = fade(xf);
    const v = fade(yf);
    const aa = perm[perm[xi] + yi];
    const ab = perm[perm[xi] + yi + 1];
    const ba = perm[perm[xi + 1] + yi];
    const bb = perm[perm[xi + 1] + yi + 1];
    return lerp(
      lerp(dot(aa, xf, yf), dot(ba, xf - 1, yf), u),
      lerp(dot(ab, xf, yf - 1), dot(bb, xf - 1, yf - 1), u),
      v,
    );
  }

  function fbm(x: number, y: number) {
    let val = 0, amp = 0.5, freq = 1;
    for (let i = 0; i < 5; i++) {
      val += amp * noise(x * freq, y * freq);
      amp *= 0.5;
      freq *= 2;
    }
    return val;
  }

  // Generate noise field
  const step = 4;
  const cols = Math.ceil(w / step) + 1;
  const rows = Math.ceil(h / step) + 1;
  const field = new Float32Array(cols * rows);
  const scale = 0.012;
  for (let j = 0; j < rows; j++) {
    for (let i = 0; i < cols; i++) {
      field[j * cols + i] = fbm(i * step * scale, j * step * scale);
    }
  }

  // Marching squares
  const isDark = document.documentElement.classList.contains('dark');
  ctx.strokeStyle = isDark ? 'rgba(160, 150, 140, 0.25)' : 'rgba(150, 140, 120, 0.22)';
  ctx.lineWidth = 1;
  ctx.lineJoin = 'round';
  ctx.lineCap = 'round';

  const levels = 16;
  const minV = -0.6, maxV = 0.6;
  for (let l = 0; l < levels; l++) {
    const threshold = minV + (maxV - minV) * (l / (levels - 1));
    ctx.beginPath();
    for (let j = 0; j < rows - 1; j++) {
      for (let i = 0; i < cols - 1; i++) {
        const tl = field[j * cols + i];
        const tr = field[j * cols + i + 1];
        const br = field[(j + 1) * cols + i + 1];
        const bl = field[(j + 1) * cols + i];
        let code = 0;
        if (tl >= threshold) code |= 8;
        if (tr >= threshold) code |= 4;
        if (br >= threshold) code |= 2;
        if (bl >= threshold) code |= 1;
        if (code === 0 || code === 15) continue;
        const x = i * step, y = j * step;
        const interp = (a: number, b: number) => {
          const d = b - a;
          return Math.abs(d) < 1e-10 ? 0.5 : (threshold - a) / d;
        };
        const top = x + interp(tl, tr) * step;
        const right = y + interp(tr, br) * step;
        const bottom = x + interp(bl, br) * step;
        const left = y + interp(tl, bl) * step;
        const segments: [number, number, number, number][] = [];
        switch (code) {
          case 1: case 14: segments.push([x, left, bottom, y + step]); break;
          case 2: case 13: segments.push([bottom, y + step, x + step, right]); break;
          case 3: case 12: segments.push([x, left, x + step, right]); break;
          case 4: case 11: segments.push([top, y, x + step, right]); break;
          case 5: segments.push([top, y, x + step, right]); segments.push([x, left, bottom, y + step]); break;
          case 6: case 9: segments.push([top, y, bottom, y + step]); break;
          case 7: case 8: segments.push([x, left, top, y]); break;
          case 10: segments.push([x, left, top, y]); segments.push([bottom, y + step, x + step, right]); break;
        }
        for (const [x1, y1, x2, y2] of segments) {
          ctx.moveTo(x1, y1);
          ctx.lineTo(x2, y2);
        }
      }
    }
    ctx.stroke();
  }
}

onMounted(() => {
  if (canvasEl.value) {
    requestAnimationFrame(() => {
      if (canvasEl.value) generateTopo(canvasEl.value);
    });
  }
});
</script>

<template>
  <canvas ref="canvasEl" class="pointer-events-none absolute inset-0 h-full w-full" />
</template>
