<script setup lang="ts">
interface Setup {
  label: string;
  text: string;
  plan: string;
  fee: number;
  web: number;
  webGb: number;
  webSize: number;
  cpu: number;
  dbGb: number;
  dbSize: number;
  dbCpu: number;
  disk: number;
  heroku: number[];
  herokuText: string;
  fly: number[];
  flyText: string;
  deploio: number[];
  deploioText: string;
}

const usdToEur = 0.89;
const chfToEur = 1.09;

const setups: Setup[] = [
  {
    label: 'Side project',
    text: 'One instance with 512 MB, and a small PostgreSQL database.',
    plan: 'Hobby', fee: 5, web: 1, webGb: 0.5, webSize: 0.25, cpu: 0.05, dbGb: 0.25, dbSize: 0.25, dbCpu: 0.05, disk: 1,
    heroku: [7, 5], herokuText: 'A Basic dyno and Postgres Essential-0',
    fly: [3.69, 38, 0.28], flyText: 'A shared-cpu-1x machine and Managed Postgres Basic',
    deploio: [17.5, 5.5], deploioText: 'A Mini app and an Economy database',
  },
  {
    label: 'Production app',
    text: 'Two instances with 1 GB each, and a 2 GB PostgreSQL database with a standby and 20 GB of disk.',
    plan: 'Pro', fee: 25, web: 2, webGb: 1, webSize: 0.5, cpu: 0.1, dbGb: 2, dbSize: 1, dbCpu: 0.1, disk: 20,
    heroku: [100, 200], herokuText: 'Two Standard-2X dynos and Postgres Premium-0',
    fly: [13.4, 72, 5.6], flyText: 'Two shared-cpu-1x machines and Managed Postgres Starter',
    deploio: [70, 68.2], deploioText: 'Two Standard-1 apps and a Business XS database',
  },
  {
    label: 'Growing product',
    text: 'Four instances with 2 GB each, and an 8 GB PostgreSQL database with a standby and 100 GB of disk.',
    plan: 'Pro', fee: 25, web: 4, webGb: 2, webSize: 1, cpu: 0.25, dbGb: 8, dbSize: 2, dbCpu: 0.25, disk: 100,
    heroku: [1000, 200], herokuText: 'Four Performance-M dynos and Postgres Premium-0, which has 4 GB',
    fly: [132, 282, 28], flyText: 'Four performance-1x machines and Managed Postgres Launch',
    deploio: [254, 103.2, 12.8], deploioText: 'Four Standard-2 apps and a Business S database',
  },
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

const usage = monthOfUsage();

const selected = ref(1);
const setup = computed(() => setups[selected.value]!);

function trace(list: number[], share: number) {
  return list.map((value, i) => `${i ? 'L' : 'M'}${((i / (list.length - 1)) * 300).toFixed(1)} ${(100 - Math.min(1, value * share) * 82).toFixed(1)}`).join('');
}

function round(n: number) {
  return String(Math.round(n * 100) / 100);
}

function sum(list: number[]) {
  return list.reduce((total, n) => total + n, 0);
}

const month = computed(() => {
  const s = setup.value;
  const cpuSize = s.web * s.webSize + 2 * s.dbSize;
  const cpuUsed = s.web * s.cpu + 2 * s.dbCpu;
  const memorySize = s.web * s.webGb + 2 * s.dbGb;
  const memoryUsed = memorySize * 0.5;
  const lucity = cpuUsed * 20 + memoryUsed * 10 + 2 * s.disk * 0.1;

  const others = [
    { name: 'Fly.io', detail: s.flyText, eur: sum(s.fly) * usdToEur, own: false },
    { name: 'Heroku', detail: s.herokuText, eur: sum(s.heroku) * usdToEur, own: false },
    { name: 'Deploio', detail: s.deploioText, eur: sum(s.deploio) * chfToEur, own: false },
  ].sort((a, b) => a.eur - b.eur);
  const rows = [
    { name: 'Lucity', detail: `${s.plan} plan, billed by usage, standby database included`, eur: Math.max(s.fee, lucity), own: true },
    ...others,
  ];
  const highest = Math.max(...rows.map(row => row.eur));

  const cpuLine = trace(usage.cpu, cpuUsed / cpuSize);
  const memoryLine = trace(usage.memory, memoryUsed / memorySize);

  return {
    cpuLine,
    cpuArea: `${cpuLine}L300 100L0 100Z`,
    memoryLine,
    memoryArea: `${memoryLine}L300 100L0 100Z`,
    cpuHead: `${round(cpuUsed)} of ${round(cpuSize)} vCPU in use on average`,
    memoryHead: `${round(memoryUsed)} of ${round(memorySize)} GB in use on average`,
    cpuSize: `Size picked: ${round(cpuSize)} vCPU`,
    memorySize: `Size picked: ${round(memorySize)} GB`,
    cpuShare: `${Math.round((cpuUsed / cpuSize) * 100)}%`,
    bars: rows.map(row => ({
      ...row,
      amount: `€${row.eur < 100 ? row.eur.toFixed(2) : Math.round(row.eur).toLocaleString('en-US')}`,
      width: `${Math.max(1, (row.eur / highest) * 100).toFixed(1)}%`,
    })),
  };
});

const comparisons = [
  { name: 'Heroku', to: '/comparisons/vs-heroku' },
  { name: 'Vercel', to: '/comparisons/vs-vercel' },
  { name: 'Railway', to: '/comparisons/vs-railway' },
  { name: 'Render', to: '/comparisons/vs-render' },
  { name: 'Fly.io', to: '/comparisons/vs-fly-io' },
  { name: 'Coolify', to: '/comparisons/vs-coolify' },
];
</script>

<template>
  <section class="price bg-stone-50 py-20 sm:pt-26 sm:pb-28 dark:bg-stone-900" aria-labelledby="price">
    <div class="mx-auto max-w-content px-6">
      <div class="flex flex-wrap items-end justify-between gap-x-18 gap-y-6">
        <div class="min-w-0 flex-[1_1_560px]">
          <h2 id="price" class="max-w-[720px] font-display text-3xl/tight text-stone-900 lg:text-4xl/tight dark:text-stone-100">You pay for what your app uses, not for the size you picked.</h2>
        </div>
        <p class="max-w-[540px] flex-[1_1_380px] text-lg text-stone-700 dark:text-stone-400">
          Plans start at €5 a month, and the fee comes back as usage credit. Pick a setup to see one month of it: what it used, what a fixed plan bills for, and what each comes to.
        </p>
      </div>

      <LandingTabs v-model="selected" :options="setups.map(item => item.label)" label="Setup" class="setup-tabs mt-9" />
      <p class="mt-4 text-lg text-stone-800 dark:text-stone-300" aria-live="polite">{{ setup.text }}</p>

      <div class="mt-10 flex flex-wrap items-start gap-x-18 gap-y-12">
        <div class="min-w-0 flex-[1.7_1_520px]">
          <div class="mb-2.5 flex flex-wrap items-baseline gap-x-3.5 gap-y-1">
            <b class="text-lg font-semibold text-stone-900 dark:text-stone-100">CPU</b>
            <span class="text-base text-stone-700 dark:text-stone-400">{{ month.cpuHead }}</span>
          </div>
          <div class="chart" role="img" aria-label="CPU use over a month. It stays far below the size picked, apart from one busy day in the third week.">
            <div class="chart-size" aria-hidden="true"><span>{{ month.cpuSize }}</span></div>
            <i class="left-1/4" /><i class="left-1/2" /><i class="left-3/4" />
            <svg viewBox="0 0 300 100" preserveAspectRatio="none" aria-hidden="true">
              <path class="chart-area" :d="month.cpuArea" />
              <path class="chart-line" :d="month.cpuLine" />
            </svg>
            <div class="chart-note" aria-hidden="true"><b>A busy day.</b> The bill follows it up, and comes back down the day after.</div>
          </div>

          <div class="mt-8.5 mb-2.5 flex flex-wrap items-baseline gap-x-3.5 gap-y-1">
            <b class="text-lg font-semibold text-stone-900 dark:text-stone-100">Memory</b>
            <span class="text-base text-stone-700 dark:text-stone-400">{{ month.memoryHead }}</span>
          </div>
          <div class="chart" role="img" aria-label="Memory use over a month. It moves around half of the size picked.">
            <div class="chart-size" aria-hidden="true"><span>{{ month.memorySize }}</span></div>
            <i class="left-1/4" /><i class="left-1/2" /><i class="left-3/4" />
            <svg viewBox="0 0 300 100" preserveAspectRatio="none" aria-hidden="true">
              <path class="chart-area" :d="month.memoryArea" />
              <path class="chart-line" :d="month.memoryLine" />
            </svg>
          </div>
          <div class="mt-2 flex justify-between text-sm text-stone-600 dark:text-stone-400" aria-hidden="true">
            <span>Day 1</span><span>8</span><span>15</span><span>23</span><span>30</span>
          </div>

          <div class="mt-5.5 flex flex-wrap gap-x-8 gap-y-2.5 text-base text-stone-800 dark:text-stone-300">
            <span class="flex items-center gap-2.5"><i class="key key-used" />What the app used. Lucity bills this.</span>
            <span class="flex items-center gap-2.5"><i class="key key-size" />The size you picked. Fixed plans bill this, used or not.</span>
          </div>
        </div>

        <div class="min-w-0 flex-[1_1_340px] border-b border-stone-300 dark:border-stone-700" aria-live="polite">
          <h3 class="mb-3.5 text-lg font-semibold text-stone-900 dark:text-stone-100">The same month, billed</h3>
          <div v-for="bar in month.bars" :key="bar.name" class="border-t border-stone-300 py-4 dark:border-stone-700">
            <div class="flex items-end justify-between gap-4">
              <div>
                <b class="block text-lg font-semibold text-stone-900 dark:text-stone-100">{{ bar.name }}</b>
                <small class="mt-0.5 block text-sm text-stone-700 dark:text-stone-400">{{ bar.detail }}</small>
              </div>
              <div class="font-display text-3xl/none whitespace-nowrap text-stone-900 dark:text-stone-100">{{ bar.amount }}</div>
            </div>
            <div class="bill-bar" :class="{ 'bill-own': bar.own }" :style="{ width: bar.width }" />
          </div>
        </div>
      </div>

      <div class="mt-6 flex flex-wrap items-start justify-between gap-x-12 gap-y-4">
        <p class="max-w-[860px] flex-[1_1_560px] text-sm text-stone-600 dark:text-stone-400">
          An example month at list prices on 10 Nov 2026, with USD converted at 0.82 and traffic left out. The usage is assumed: half of the memory and about {{ month.cpuShare }} of the CPU on average, with one busy day. Fly.io, Heroku and Deploio bill the size you pick, so the closest plan to each size is counted, with a standby database where one is listed.
        </p>
        <NuxtLink to="/pricing" class="text-lg font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">See pricing and estimate your bill</NuxtLink>
      </div>

      <p class="mt-7 text-lg text-stone-700 dark:text-stone-400">
        Coming from
        <template v-for="(comparison, index) in comparisons" :key="comparison.name">
          <NuxtLink :to="comparison.to" class="font-medium text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">{{ comparison.name }}</NuxtLink>{{ index < comparisons.length - 2 ? ', ' : index === comparisons.length - 2 ? ' or ' : '' }}
        </template>? See how Lucity compares.
      </p>
    </div>
  </section>
</template>

<style scoped>
.price {
  --chart-ink: oklch(0.25 0.03 55);
  --chart-dots: oklch(0.25 0.04 55 / .30);
  --chart-wash: oklch(0.25 0.04 55 / .035);
  --chart-grid: oklch(0.25 0.04 55 / .14);
  box-shadow: inset 0 1px 0 oklch(0.85 0.03 75);
}

.dark .price {
  --chart-ink: oklch(0.90 0.02 80);
  --chart-dots: oklch(0.90 0.02 80 / .30);
  --chart-wash: oklch(0.90 0.02 80 / .04);
  --chart-grid: oklch(0.90 0.02 80 / .12);
  box-shadow: inset 0 1px 0 oklch(0.30 0.03 55);
}

.dark .setup-tabs {
  background: oklch(1 0 0 / .07);
}

.setup-tabs :deep(.tab) {
  --tab-off: oklch(0.40 0.04 55);
}

.dark .setup-tabs :deep(.tab) {
  --tab-off: oklch(0.80 0.02 78);
}

.chart {
  position: relative;
  height: 168px;
  box-shadow: 0 1px 0 var(--chart-ink);
}

.chart-size,
.key-size,
.bill-bar {
  border-top: 2px solid var(--chart-ink);
  background-color: var(--chart-wash);
  background-image: radial-gradient(circle at 1px 1px, var(--chart-dots) 1px, transparent 0);
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
  color: var(--chart-ink);
  white-space: nowrap;
}

.chart > i {
  position: absolute;
  top: 18%;
  bottom: 0;
  width: 1px;
  background: var(--chart-grid);
}

.chart svg {
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
  height: 100%;
}

.chart-area {
  fill: oklch(0.80 0.15 160);
}

.chart-line {
  fill: none;
  stroke: oklch(0.48 0.13 160);
  stroke-width: 1.5;
  stroke-linejoin: round;
  vector-effect: non-scaling-stroke;
}

.chart-note {
  position: absolute;
  left: 62.5%;
  top: 27%;
  width: min(35%, 236px);
  box-sizing: border-box;
  padding: 9px 12px;
  background: oklch(0.99 0.006 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .16), 0 10px 20px -12px oklch(0.25 0.04 55 / .4);
  font-size: 13.5px;
  line-height: 1.4;
  color: oklch(0.35 0.03 55);
}

.chart-note b {
  color: oklch(0.25 0.03 55);
}

.key {
  flex: none;
  box-sizing: border-box;
  width: 24px;
  height: 14px;
}

.key-used,
.bill-own {
  border-top: 0;
  background: oklch(0.80 0.15 160);
  box-shadow: inset 0 0 0 1.5px oklch(0.48 0.13 160);
}

.bill-bar {
  box-sizing: border-box;
  height: 12px;
  margin-top: 12px;
  min-width: 4px;
  transition: width .45s cubic-bezier(.2, .7, .2, 1);
}

@media (max-width: 700px) {
  .chart-note {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .bill-bar {
    transition: none;
  }
}
</style>
