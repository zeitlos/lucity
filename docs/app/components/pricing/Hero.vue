<script setup lang="ts">
const appUrl = '/app';

const checks = [
  '14 days free, up to CHF 5 of usage',
  'No credit card to start',
  'Monthly, no contract',
];

function monthOfUsage() {
  let seed = 7;
  const noise = () => {
    seed = (seed * 16807) % 2147483647;
    return seed / 2147483647;
  };
  const level = (list: number[]) => {
    const mean = list.reduce((total, n) => total + n, 0) / list.length;
    return list.map(n => n / mean);
  };
  const cpu: number[] = [];
  const memory: number[] = [];
  for (let i = 0; i < 180; i += 1) {
    const day = i / 6;
    const hour = (i % 6) / 6;
    const week = Math.floor(day) % 7 < 5 ? 1 : 0.5;
    const daily = 0.45 + 0.75 * Math.max(0, Math.sin((hour - 0.2) * Math.PI * 1.25));
    const busy = Math.exp(-Math.pow((day - 17.4) / 0.55, 2));
    cpu.push((daily * week + busy * 2.6) * (0.85 + noise() * 0.3));
    memory.push(0.82 + 0.26 * ((day % 6) / 6) + 0.25 * Math.exp(-Math.pow((day - 17.4) / 0.9, 2)) + (noise() - 0.5) * 0.04);
  }
  return { cpu: level(cpu), memory: level(memory) };
}

function trace(list: number[], share: number) {
  return list.map((value, i) => `${i ? 'L' : 'M'}${((i / (list.length - 1)) * 300).toFixed(1)} ${(100 - Math.min(1, value * share) * 82).toFixed(1)}`).join('');
}

const usage = monthOfUsage();

const charts = [
  {
    name: 'CPU',
    head: '0.4 of 3 vCPU in use on average',
    size: 'Size picked: 3 vCPU',
    line: trace(usage.cpu, 0.4 / 3),
    label: 'CPU use over a month. It stays far below the size picked, apart from one busy day in the third week.',
    note: true,
  },
  {
    name: 'Memory',
    head: '3 of 6 GB in use on average',
    size: 'Size picked: 6 GB',
    line: trace(usage.memory, 3 / 6),
    label: 'Memory use over a month. It moves around half of the size picked.',
    note: false,
  },
];

const lines = [
  { name: 'CPU, 0.4 vCPU in use', amount: '8.00' },
  { name: 'Memory, 3 GB in use', amount: '30.00' },
  { name: 'Disk, 40 GB at full size', amount: '4.00' },
];

const creditMark = `${((25 / 124) * 100).toFixed(2)}%`;
</script>

<template>
  <section class="mx-auto max-w-content px-6 pt-12 sm:pt-28" aria-labelledby="pricing-title">
    <h1 id="pricing-title" class="load-in font-display text-4xl/[1.15] text-neutral-800 max-md:text-balance sm:text-5xl/[1.1] lg:text-6xl/[1.1] xl:text-7xl/[1.1] dark:text-neutral-100" style="--load-order: 0">
      Pay for <span class="squiggle">what your apps use</span>,<br class="max-md:hidden">
      not for the size you picked.
    </h1>

    <p class="load-in mt-8 max-w-[800px] text-xl leading-relaxed text-stone-700 sm:text-2xl sm:leading-relaxed dark:text-stone-300" style="--load-order: 1">
      CPU and memory are billed as your services actually use them, so an app that sits idle costs very little. Plans start at CHF 5 a month, and the fee comes back as credit for that usage.
    </p>

    <div class="load-in mt-8 flex flex-col gap-3 sm:mt-11 sm:flex-row sm:items-center sm:gap-4" style="--load-order: 2">
      <a
        :href="`${appUrl}/login`"
        class="flex items-center justify-center gap-3 rounded-lg border border-transparent bg-primary px-7 py-3.5 text-lg font-semibold text-white shadow-sm transition-colors hover:bg-primary-600 sm:py-4 sm:text-xl"
      >
        Start free
        <UIcon name="i-lucide-arrow-right" class="size-5.5 shrink-0" />
      </a>
      <NuxtLink
        :to="{ hash: '#estimate' }"
        class="flex items-center justify-center rounded-lg border border-stone-300 bg-white/50 px-7 py-3.5 text-lg font-semibold text-neutral-800 transition-colors hover:bg-white sm:py-4 sm:text-xl dark:border-stone-700 dark:bg-white/5 dark:text-neutral-100 dark:hover:bg-white/10"
      >
        Estimate your bill
      </NuxtLink>
    </div>

    <ul class="load-in mt-5 flex flex-wrap gap-x-7 gap-y-1.5 text-base text-stone-600 sm:mt-6 dark:text-stone-400" style="--load-order: 3">
      <li v-for="check in checks" :key="check" class="flex items-center gap-2">
        <UIcon name="i-lucide-check" class="size-4.5 shrink-0 text-primary-700 dark:text-primary" />
        {{ check }}
      </li>
    </ul>

    <div
      class="load-in month mt-12 px-4 py-6 sm:mt-14 sm:p-9"
      style="--load-order: 4"
      role="group"
      aria-label="One month of a production app: two instances with 1 GB each and a 2 GB PostgreSQL database with a standby. Billed by use it comes to CHF 42. The same sizes billed in full at the same rates would come to CHF 124."
    >
      <p class="setup"><b>One month of a production app.</b> Two instances with 1 GB each, and a 2 GB PostgreSQL database with a standby and 20 GB of disk.</p>

      <div class="mt-7.5 flex flex-wrap items-start gap-x-16 gap-y-10">
        <div class="min-w-0 flex-[1.7_1_520px]">
          <template v-for="(chart, index) in charts" :key="chart.name">
            <div class="chart-head" :class="{ 'mt-7': index > 0 }">
              <b>{{ chart.name }}</b>
              <span>{{ chart.head }}</span>
            </div>
            <div class="chart" role="img" :aria-label="chart.label">
              <div class="chart-size" aria-hidden="true"><span>{{ chart.size }}</span></div>
              <i class="left-1/4" /><i class="left-1/2" /><i class="left-3/4" />
              <div class="pm-plot">
                <svg viewBox="0 0 300 100" preserveAspectRatio="none" aria-hidden="true">
                  <path class="chart-area" :d="`${chart.line}L300 100L0 100Z`" />
                  <path class="chart-line" :d="chart.line" />
                </svg>
              </div>
              <div class="pm-now" aria-hidden="true" />
              <div v-if="chart.note" class="pm-note" aria-hidden="true"><b>A busy day.</b> The bill follows it up, and comes back down the day after.</div>
            </div>
          </template>
          <div class="chart-days" aria-hidden="true">
            <span>Day 1</span><span>8</span><span>15</span><span>23</span><span>30</span>
          </div>
          <div class="chart-key">
            <span><i class="key-used" />What the app used. This is what you pay for.</span>
            <span><i class="key-size" />The size you picked. A fixed plan bills this, used or not.</span>
          </div>
        </div>

        <div class="bill min-w-0 flex-[1_1_340px]">
          <p class="bill-day">Day <span class="pm-d" aria-hidden="true" /><span class="sr-only">30</span> of 30</p>
          <p class="bill-cap">Billed by use</p>
          <p class="bill-amount">CHF <span class="pm-n" aria-hidden="true" /><span class="sr-only">42</span></p>
          <div class="bill-track">
            <i :style="{ left: creditMark }" />
            <div class="pm-ubar" />
          </div>
          <div class="bill-caps">
            <span class="pm-cap1" aria-hidden="true">So far inside the CHF 25 that Pro includes.</span>
            <span class="pm-cap2">Past the CHF 25 that Pro includes. Only the rest is added to the invoice.</span>
          </div>
          <ul class="bill-lines">
            <li v-for="line in lines" :key="line.name"><span>{{ line.name }}</span><em>{{ line.amount }}</em></li>
          </ul>
          <div class="mt-2">
            <p class="bill-cap flex items-baseline justify-between gap-4">
              <span>The same sizes, billed in full</span>
              <b>CHF <span class="pm-f" aria-hidden="true" /><span class="sr-only">124</span></b>
            </p>
            <div class="bill-track bill-track-fixed">
              <div class="pm-fbar" />
            </div>
          </div>
        </div>
      </div>

      <p class="fine">An example month. The use is assumed: half of the memory and about 13% of the CPU on average, with one busy day. Traffic is left out.</p>
    </div>
  </section>
</template>

<style scoped>
@property --pm-n {
  syntax: '<integer>';
  inherits: true;
  initial-value: 0;
}

@property --pm-f {
  syntax: '<integer>';
  inherits: true;
  initial-value: 0;
}

@property --pm-d {
  syntax: '<integer>';
  inherits: true;
  initial-value: 0;
}

.squiggle {
  position: relative;
  white-space: nowrap;
}

.squiggle::after {
  content: "";
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 0.34em;
  background: url('/img/underline-joy.svg') center / 100% 100% no-repeat;
  clip-path: inset(0 100% 0 0);
  animation: highlighter 0.55s ease-out 0.35s forwards;
}

@keyframes highlighter {
  to {
    clip-path: inset(0 0 0 0);
  }
}

.month {
  --ink: oklch(0.25 0.03 55);
  --body: oklch(0.35 0.03 55);
  --muted: oklch(0.45 0.04 55);
  --meta: oklch(0.47 0.04 55);
  --rule-soft: oklch(0.91 0.015 80);
  --used: oklch(0.80 0.15 160);
  --used-edge: oklch(0.48 0.13 160);
  position: relative;
  overflow: hidden;
  border-radius: 24px;
  color: var(--body);
  background-color: oklch(0.975 0.012 80);
  background-image: radial-gradient(circle at 1px 1px, oklch(0.25 0.04 55 / .14) 1px, transparent 0);
  background-size: 22px 22px;
  box-shadow: inset 0 0 0 1px oklch(0.25 0.04 55 / .07);
}

.setup {
  max-width: 820px;
  font-size: 18px;
}

.setup b {
  font-weight: 600;
  color: var(--ink);
}

.chart-head {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 4px 14px;
  margin-bottom: 10px;
}

.chart-head b {
  font-size: 18px;
  font-weight: 600;
  color: var(--ink);
}

.chart-head span {
  font-size: 16px;
  color: var(--muted);
}

.chart {
  position: relative;
  height: 150px;
  box-shadow: 0 1px 0 var(--ink);
}

.chart-size,
.key-size,
.pm-fbar {
  border-top: 2px solid var(--ink);
  background-color: oklch(0.25 0.04 55 / .035);
  background-image: radial-gradient(circle at 1px 1px, oklch(0.25 0.04 55 / .30) 1px, transparent 0);
  background-size: 7px 7px;
}

.chart-size {
  position: absolute;
  left: 0;
  right: 0;
  top: 18%;
  bottom: 0;
}

.chart-size span {
  position: absolute;
  right: 0;
  bottom: calc(100% + 8px);
  font-size: 13.5px;
  font-weight: 600;
  color: var(--ink);
  white-space: nowrap;
}

.chart > i {
  position: absolute;
  top: 18%;
  bottom: 0;
  width: 1px;
  background: oklch(0.25 0.04 55 / .14);
}

.chart svg {
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
}

.chart-area {
  fill: var(--used);
}

.chart-line {
  fill: none;
  stroke: var(--used-edge);
  stroke-width: 1.5;
  stroke-linejoin: round;
  vector-effect: non-scaling-stroke;
}

.chart-days {
  display: flex;
  justify-content: space-between;
  margin-top: 8px;
  font-size: 13.5px;
  color: var(--meta);
}

.chart-key {
  display: flex;
  flex-wrap: wrap;
  gap: 10px 32px;
  margin-top: 22px;
  font-size: 16px;
}

.chart-key span {
  display: flex;
  align-items: center;
  gap: 10px;
}

.chart-key i {
  flex: none;
  box-sizing: border-box;
  width: 24px;
  height: 14px;
}

.key-used {
  background: var(--used);
  box-shadow: inset 0 0 0 1.5px var(--used-edge);
}

.bill {
  box-sizing: border-box;
  padding: 26px 30px;
  border-radius: 14px;
  background: oklch(0.99 0.006 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .10), 0 24px 40px -24px oklch(0.25 0.04 55 / .35);
}

.bill-day {
  font-size: 15px;
  font-weight: 600;
  color: var(--meta);
}

.bill-cap {
  margin-top: 14px;
  font-size: 16px;
}

.bill-cap b {
  font-weight: 600;
  color: var(--ink);
  white-space: nowrap;
}

.bill-amount {
  margin-top: 2px;
  font-family: var(--font-display);
  font-size: 64px;
  line-height: 1.05;
  color: var(--ink);
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
}

.bill-track {
  position: relative;
  height: 14px;
  margin-top: 14px;
  overflow: hidden;
  background: oklch(0.25 0.04 55 / .06);
}

.bill-track > i {
  position: absolute;
  z-index: 1;
  top: 0;
  bottom: 0;
  width: 2px;
  margin-left: -1px;
  background: var(--ink);
}

.bill-track-fixed {
  background: none;
}

.bill-caps {
  display: grid;
  min-height: 42px;
  margin-top: 10px;
  font-size: 14.5px;
  line-height: 1.45;
  color: var(--muted);
}

.bill-caps span {
  grid-area: 1 / 1;
}

.bill-lines {
  margin-top: 18px;
}

.bill-lines li {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 0;
  border-top: 1px solid var(--rule-soft);
  font-size: 16px;
  color: var(--ink);
}

.bill-lines li:last-child {
  border-bottom: 1px solid var(--rule-soft);
}

.bill-lines em {
  font-family: var(--font-mono);
  font-size: 14.5px;
  font-style: normal;
  font-weight: 500;
  white-space: nowrap;
}

.fine {
  margin-top: 24px;
  font-size: 14.5px;
  color: var(--muted);
}

.pm-plot {
  position: absolute;
  inset: 0;
  clip-path: inset(0 0% 0 0);
  animation: pm-plot 20s linear infinite both;
}

.pm-now {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 100%;
  width: 2px;
  margin-left: -1px;
  background: var(--ink);
  opacity: 0;
  animation: pm-now 20s linear infinite both;
}

.pm-note {
  position: absolute;
  z-index: 1;
  left: 62.5%;
  top: 27%;
  width: min(35%, 236px);
  box-sizing: border-box;
  padding: 9px 12px;
  background: oklch(0.99 0.006 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .16), 0 10px 20px -12px oklch(0.25 0.04 55 / .4);
  font-size: 13.5px;
  line-height: 1.4;
  color: var(--body);
  animation: pm-note 20s linear infinite both;
}

.pm-note b {
  color: var(--ink);
}

.pm-n {
  --pm-n: 42;
  counter-reset: pmn var(--pm-n);
  animation: pm-n 20s linear infinite both;
}

.pm-n::after {
  content: counter(pmn);
}

.pm-f {
  --pm-f: 124;
  counter-reset: pmf var(--pm-f);
  animation: pm-f 20s linear infinite both;
}

.pm-f::after {
  content: counter(pmf);
}

.pm-d {
  --pm-d: 30;
  counter-reset: pmd var(--pm-d);
  animation: pm-d 20s linear infinite both;
}

.pm-d::after {
  content: counter(pmd);
}

.pm-ubar,
.pm-fbar {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 100%;
  box-sizing: border-box;
  transform-origin: left center;
}

.pm-ubar {
  background: var(--used);
  box-shadow: inset 0 0 0 1.5px var(--used-edge);
  transform: scaleX(0.3387);
  animation: pm-ubar 20s linear infinite both;
}

.pm-fbar {
  transform: scaleX(1);
  animation: pm-fbar 20s linear infinite both;
}

.pm-cap1 {
  opacity: 0;
  animation: pm-cap1 20s linear infinite both;
}

.pm-cap2 {
  opacity: 1;
  animation: pm-cap2 20s linear infinite both;
}

@keyframes pm-plot {
  0% { clip-path: inset(0 100% 0 0); }
  5% { clip-path: inset(0 100% 0 0); }
  65% { clip-path: inset(0 0% 0 0); }
  92.5% { clip-path: inset(0 0% 0 0); }
  95% { clip-path: inset(0 100% 0 0); }
  100% { clip-path: inset(0 100% 0 0); }
}

@keyframes pm-now {
  0% { left: 0%; opacity: 0; }
  4% { left: 0%; opacity: 0; }
  5% { left: 0%; opacity: 1; }
  65% { left: 100%; opacity: 1; }
  66% { left: 100%; opacity: 1; }
  68% { left: 100%; opacity: 0; }
  94.5% { left: 100%; opacity: 0; }
  94.6% { left: 0%; opacity: 0; }
  100% { left: 0%; opacity: 0; }
}

@keyframes pm-note {
  0% { opacity: 0; top: 32%; }
  39.8% { opacity: 0; top: 32%; }
  41.55% { opacity: 1; top: 27%; }
  91% { opacity: 1; top: 27%; }
  92.5% { opacity: 0; top: 32%; }
  100% { opacity: 0; top: 32%; }
}

@keyframes pm-n {
  0% { --pm-n: 0; }
  5% { --pm-n: 0; }
  7% { --pm-n: 1; }
  9% { --pm-n: 3; }
  11% { --pm-n: 4; }
  13% { --pm-n: 5; }
  15% { --pm-n: 7; }
  17% { --pm-n: 8; }
  19% { --pm-n: 9; }
  21% { --pm-n: 11; }
  23% { --pm-n: 12; }
  25% { --pm-n: 13; }
  27% { --pm-n: 15; }
  29% { --pm-n: 16; }
  31% { --pm-n: 18; }
  33% { --pm-n: 19; }
  35% { --pm-n: 20; }
  37% { --pm-n: 22; }
  39% { --pm-n: 23; }
  41% { --pm-n: 26; }
  43% { --pm-n: 27; }
  45% { --pm-n: 28; }
  47% { --pm-n: 29; }
  49% { --pm-n: 31; }
  51% { --pm-n: 32; }
  53% { --pm-n: 34; }
  55% { --pm-n: 35; }
  57% { --pm-n: 37; }
  59% { --pm-n: 38; }
  61% { --pm-n: 39; }
  63% { --pm-n: 40; }
  65% { --pm-n: 42; }
  95% { --pm-n: 42; }
  95.1% { --pm-n: 0; }
  100% { --pm-n: 0; }
}

@keyframes pm-f {
  0% { --pm-f: 0; }
  5% { --pm-f: 0; }
  7% { --pm-f: 4; }
  9% { --pm-f: 8; }
  11% { --pm-f: 12; }
  13% { --pm-f: 17; }
  15% { --pm-f: 21; }
  17% { --pm-f: 25; }
  19% { --pm-f: 29; }
  21% { --pm-f: 33; }
  23% { --pm-f: 37; }
  25% { --pm-f: 41; }
  27% { --pm-f: 45; }
  29% { --pm-f: 50; }
  31% { --pm-f: 54; }
  33% { --pm-f: 58; }
  35% { --pm-f: 62; }
  37% { --pm-f: 66; }
  39% { --pm-f: 70; }
  41% { --pm-f: 74; }
  43% { --pm-f: 79; }
  45% { --pm-f: 83; }
  47% { --pm-f: 87; }
  49% { --pm-f: 91; }
  51% { --pm-f: 95; }
  53% { --pm-f: 99; }
  55% { --pm-f: 103; }
  57% { --pm-f: 107; }
  59% { --pm-f: 112; }
  61% { --pm-f: 116; }
  63% { --pm-f: 120; }
  65% { --pm-f: 124; }
  95% { --pm-f: 124; }
  95.1% { --pm-f: 0; }
  100% { --pm-f: 0; }
}

@keyframes pm-d {
  0% { --pm-d: 0; }
  5% { --pm-d: 0; }
  7% { --pm-d: 1; }
  9% { --pm-d: 2; }
  11% { --pm-d: 3; }
  13% { --pm-d: 4; }
  15% { --pm-d: 5; }
  17% { --pm-d: 6; }
  19% { --pm-d: 7; }
  21% { --pm-d: 8; }
  23% { --pm-d: 9; }
  25% { --pm-d: 10; }
  27% { --pm-d: 11; }
  29% { --pm-d: 12; }
  31% { --pm-d: 13; }
  33% { --pm-d: 14; }
  35% { --pm-d: 15; }
  37% { --pm-d: 16; }
  39% { --pm-d: 17; }
  41% { --pm-d: 18; }
  43% { --pm-d: 19; }
  45% { --pm-d: 20; }
  47% { --pm-d: 21; }
  49% { --pm-d: 22; }
  51% { --pm-d: 23; }
  53% { --pm-d: 24; }
  55% { --pm-d: 25; }
  57% { --pm-d: 26; }
  59% { --pm-d: 27; }
  61% { --pm-d: 28; }
  63% { --pm-d: 29; }
  65% { --pm-d: 30; }
  95% { --pm-d: 30; }
  95.1% { --pm-d: 0; }
  100% { --pm-d: 0; }
}

@keyframes pm-ubar {
  0% { transform: scaleX(0); }
  5% { transform: scaleX(0); }
  7% { transform: scaleX(0.0104); }
  9% { transform: scaleX(0.0211); }
  11% { transform: scaleX(0.0323); }
  13% { transform: scaleX(0.0436); }
  15% { transform: scaleX(0.0554); }
  17% { transform: scaleX(0.0665); }
  19% { transform: scaleX(0.0757); }
  21% { transform: scaleX(0.0863); }
  23% { transform: scaleX(0.0973); }
  25% { transform: scaleX(0.1087); }
  27% { transform: scaleX(0.1204); }
  29% { transform: scaleX(0.1324); }
  31% { transform: scaleX(0.1417); }
  33% { transform: scaleX(0.1514); }
  35% { transform: scaleX(0.1625); }
  37% { transform: scaleX(0.1740); }
  39% { transform: scaleX(0.1873); }
  41% { transform: scaleX(0.2065); }
  43% { transform: scaleX(0.2181); }
  45% { transform: scaleX(0.2277); }
  47% { transform: scaleX(0.2376); }
  49% { transform: scaleX(0.2491); }
  51% { transform: scaleX(0.2611); }
  53% { transform: scaleX(0.2733); }
  55% { transform: scaleX(0.2838); }
  57% { transform: scaleX(0.2945); }
  59% { transform: scaleX(0.3044); }
  61% { transform: scaleX(0.3147); }
  63% { transform: scaleX(0.3265); }
  65% { transform: scaleX(0.3387); }
  92.5% { transform: scaleX(0.3387); }
  95% { transform: scaleX(0); }
  100% { transform: scaleX(0); }
}

@keyframes pm-fbar {
  0% { transform: scaleX(0); }
  5% { transform: scaleX(0); }
  65% { transform: scaleX(1); }
  92.5% { transform: scaleX(1); }
  95% { transform: scaleX(0); }
  100% { transform: scaleX(0); }
}

@keyframes pm-cap1 {
  0% { opacity: 0; }
  5% { opacity: 0; }
  6.5% { opacity: 1; }
  41% { opacity: 1; }
  42% { opacity: 0; }
  100% { opacity: 0; }
}

@keyframes pm-cap2 {
  0% { opacity: 0; }
  42% { opacity: 0; }
  43.5% { opacity: 1; }
  92.5% { opacity: 1; }
  94.75% { opacity: 0; }
  100% { opacity: 0; }
}
@media (max-width: 700px) {
  .pm-note {
    display: none;
  }

  .bill {
    padding: 22px 20px;
  }

  .bill-amount {
    font-size: 52px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .squiggle::after {
    animation: none;
    clip-path: inset(0 0 0 0);
  }

  .month * {
    animation: none !important;
  }
}
</style>
