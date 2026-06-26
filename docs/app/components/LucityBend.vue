<script setup lang="ts">
type AxisScores = { simplicity: number; scalability: number; sovereignty: number };

const cx = 220;
const cy = 190;
const R = 150;
const axes = {
  simplicity: { ux: 0, uy: -1 },
  sovereignty: { ux: 0.8660254, uy: 0.5 },
  scalability: { ux: -0.8660254, uy: 0.5 },
};

function coord(axis: keyof typeof axes, s: number) {
  const a = axes[axis];
  return { x: cx + R * s * a.ux, y: cy + R * s * a.uy };
}
function poly(scores: AxisScores) {
  return [coord('simplicity', scores.simplicity), coord('sovereignty', scores.sovereignty), coord('scalability', scores.scalability)]
    .map(p => `${p.x.toFixed(1)},${p.y.toFixed(1)}`)
    .join(' ');
}
function tri(s: number) {
  return poly({ simplicity: s, scalability: s, sovereignty: s });
}

const axisTips = [coord('simplicity', 1), coord('sovereignty', 1), coord('scalability', 1)];

// Lucity ships as a simple PaaS (simplicity-heavy) ...
const paasScores: AxisScores = { simplicity: 0.95, scalability: 0.55, sovereignty: 0.5 };
const paasPoly = poly(paasScores);
const paasDots = [coord('simplicity', paasScores.simplicity), coord('sovereignty', paasScores.sovereignty), coord('scalability', paasScores.scalability)];

// ... that you can eject onto the exact native Kubernetes evaluation.
const ejectPoly = poly({ simplicity: 0.18, scalability: 0.95, sovereignty: 0.85 });
</script>

<template>
  <section class="px-6">
    <div class="mx-auto grid max-w-352 items-center gap-12 lg:grid-cols-2 lg:gap-20">
      <div>
        <h2 class="font-display text-5xl leading-tight text-neutral-800 md:text-6xl dark:text-neutral-100">
          …Lucity is the one you can walk away from.
        </h2>

        <p class="mt-8 text-2xl leading-relaxed">
          Lucity can bend the software deployment trilemma, thanks to its unique ejectable architecture.
          <code class="rounded-md bg-[#00b87d] px-2 py-0.5 font-mono text-[0.9em] text-white">lucity eject</code>
          to native Kubernetes the moment you outgrow it.
        </p>
      </div>

      <div class="relative w-full">
        <svg viewBox="0 0 440 400" class="w-full" role="img" aria-label="How Lucity bends the deployment trilemma">
          <defs>
            <pattern id="bend-hatch" width="9" height="9" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">
              <line x1="0" y1="0" x2="0" y2="9" stroke="#00b87d" stroke-width="1.3" stroke-opacity="0.5" />
            </pattern>
          </defs>

          <g fill="none" stroke="#d2b78b" class="chart-grid">
            <polygon :points="tri(1)" stroke-width="1.25" />
            <polygon :points="tri(0.66)" stroke-opacity="0.55" />
            <polygon :points="tri(0.33)" stroke-opacity="0.55" />
            <line v-for="(tip, i) in axisTips" :key="i" :x1="cx" :y1="cy" :x2="tip.x" :y2="tip.y" stroke-opacity="0.55" />
          </g>

          <!-- Ejected reach (native Kubernetes) -->
          <polygon :points="ejectPoly" fill="url(#bend-hatch)" stroke="#00b87d" stroke-width="2" stroke-opacity="0.7" stroke-linejoin="round" />

          <!-- Lucity-as-PaaS shape -->
          <g style="color: #00b87d">
            <polygon
              :points="paasPoly"
              fill="currentColor"
              fill-opacity="0.2"
              stroke="currentColor"
              stroke-width="2.5"
              stroke-linejoin="round"
            />
            <circle v-for="(d, i) in paasDots" :key="i" :cx="d.x" :cy="d.y" r="5" fill="currentColor" />
          </g>
        </svg>

        <span class="absolute left-1/2 top-[1%] -translate-x-1/2 text-2xl font-medium text-neutral-800 dark:text-neutral-100">Simplicity</span>
        <span class="absolute left-[20.5%] top-[70%] -translate-x-1/2 text-2xl font-medium text-neutral-800 dark:text-neutral-100">Scalability</span>
        <span class="absolute left-[79.5%] top-[70%] -translate-x-1/2 text-2xl font-medium text-neutral-800 dark:text-neutral-100">Sovereignty</span>

        <div class="pointer-events-none absolute left-1/2 top-3/5 -translate-x-1/2 text-center">
          <code class="rounded-md bg-[#00b87d] px-1 py-0.5 font-mono text-xs text-white">lucity eject</code>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.dark .chart-grid {
  stroke: oklch(0.48 0.04 75);
}
</style>
