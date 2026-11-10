<script setup lang="ts">
const props = defineProps<{ plain?: boolean }>();

const appUrl = '/app';

const columns = [
  {
    title: 'Product',
    links: [
      { label: 'Features', to: '/features' },
      { label: 'No lock-in', to: '/docs/eject' },
      { label: 'Pricing', to: '/pricing' },
      { label: 'Self-hosting', to: '/docs/self-hosting' },
    ],
  },
  {
    title: 'Use cases',
    links: [
      { label: 'Agencies', to: '/use-cases/agencies' },
      { label: 'Startups', to: '/use-cases/startups' },
      { label: 'Platform teams', to: '/use-cases/platform-teams' },
    ],
  },
  {
    title: 'Compare',
    links: [
      { label: 'vs Railway', to: '/comparisons/vs-railway' },
      { label: 'vs Heroku', to: '/comparisons/vs-heroku' },
      { label: 'vs Render', to: '/comparisons/vs-render' },
      { label: 'vs Vercel', to: '/comparisons/vs-vercel' },
      { label: 'vs Fly.io', to: '/comparisons/vs-fly-io' },
      { label: 'vs Coolify', to: '/comparisons/vs-coolify' },
    ],
  },
  {
    title: 'Resources',
    links: [
      { label: 'Docs', to: '/docs/quickstart' },
      { label: 'Framework guides', to: '/docs/guides' },
      { label: 'CLI', to: '/docs/cli' },
      { label: 'Blog', to: '/blog' },
      { label: 'Changelog', href: 'https://github.com/zeitlos/lucity/releases' },
    ],
  },
  {
    title: 'Company',
    links: [
      { label: 'About zeitlos', href: 'https://zeitlos.software' },
      { label: 'Contact', href: 'mailto:christian@zeitlos.software' },
      { label: 'Report a vulnerability', href: 'https://github.com/zeitlos/lucity/security/policy' },
      { label: 'Privacy policy', to: '/legal/privacy-policy' },
      { label: 'Terms of use', to: '/legal/terms-of-use' },
      { label: 'Data processing', to: '/legal/data-processing-agreement' },
    ],
  },
];

const markPath = 'M 130 0 L 130 76 C 130 102.64, 114.4 112, 70 112 L 60 112 C 15.6 112, 0 121.36, 0 148 L 0 181 L 169 181 L 169 148 C 169 121.36, 178.23 112, 204.5 112 L 214.5 112 C 240.77 112, 250 102.64, 250 76 L 250 0 Z';
const wordmarkPath = 'M2167.8 -209.6Q2167.8 -136.8 2191.1 -88Q2214.5 -39.2 2258.9 -15Q2303.4 9.1 2366.7 9.1Q2430.1 9.1 2478.6 -22.1Q2527.1 -53.3 2550.6 -110.3H2552.5V40C2552.5 100.7 2413.8 61.3 2413.8 122V200H2594.9V122C2594.9 61.3 2681 100.7 2681 40V-532.4H2552.5V-251.5Q2552.5 -205.3 2534.9 -169.6Q2517.3 -133.9 2485.8 -114.3Q2454.3 -94.7 2414.5 -94.7Q2352.5 -94.7 2324.4 -131.5Q2296.2 -168.3 2296.2 -240.9V-532.4H2167.8ZM465.8 -209.6Q465.8 -136.8 489.1 -88Q512.5 -39.2 556.9 -15Q601.4 9.1 664.7 9.1Q728.1 9.1 776.6 -22.1Q825.1 -53.3 848.6 -110.3H850.6L855.9 0H979V-532.4H850.5V-251.5Q850.5 -205.3 832.9 -169.6Q815.3 -133.9 783.8 -114.3Q752.3 -94.7 712.5 -94.7Q650.5 -94.7 622.4 -131.5Q594.2 -168.3 594.2 -240.9V-532.4H465.8ZM1031.8 -266.4Q1031.8 -183.7 1065.4 -121.3Q1098.9 -59 1160.6 -24.9Q1222.3 9.1 1305.3 9.1Q1380.6 9.1 1438.6 -16.1Q1496.7 -41.4 1531.7 -88.9Q1566.8 -136.3 1574.3 -201.5H1443.3Q1433.9 -150.4 1398.3 -122.8Q1362.7 -95.2 1305.8 -95.2Q1261.3 -95.2 1229.1 -114.9Q1196.9 -134.6 1179.3 -172.9Q1161.6 -211.1 1161.6 -265.8Q1161.6 -318.1 1178.5 -356.4Q1195.3 -394.7 1227.8 -416Q1260.4 -437.4 1306.9 -437.4Q1360.6 -437.4 1395.2 -411.5Q1429.9 -385.6 1441.3 -334.9H1573.1Q1566.2 -398.6 1532 -445.2Q1497.9 -491.8 1441.1 -516.8Q1384.3 -541.7 1310 -541.7Q1225.7 -541.7 1162.9 -507.8Q1100.1 -473.9 1065.9 -411.5Q1031.8 -349.1 1031.8 -266.4ZM0 -729V0H435.7V-118.9H135.3V-729ZM1889.9 -129.3Q1889.9 -62.9 1921.4 -31.5Q1953 0 2018.4 0H2135.4V-106.6H2067Q2042.3 -106.6 2030 -117.6Q2017.7 -128.7 2017.7 -154.9V-429.9H2134.7V-532.4H2017.7V-729H1889.9V-532.4H1790.7V-429.9H1889.9ZM1626.8 -532.4V0H1754.9V-532.4ZM1631.1 -609.4H1750.7V-729H1631.1Z';
const wordmarkLayers = ['base', 'glow', 'light'];

const headerOverDark = useHeaderOverDark();

const footer = ref<HTMLElement>();
const wordmark = ref<HTMLElement>();

let frame = 0;
let hovering = false;

function place(x: number, y: number) {
  if (!wordmark.value) return;
  wordmark.value.style.setProperty('--footer-spot-x', `${Math.max(0, Math.min(100, x)).toFixed(2)}%`);
  wordmark.value.style.setProperty('--footer-spot-y', `${Math.max(0, Math.min(100, y)).toFixed(2)}%`);
}

function follow() {
  if (!wordmark.value || !footer.value) return;
  headerOverDark.value = footer.value.getBoundingClientRect().top < -430;
  if (hovering) return;
  const rect = wordmark.value.getBoundingClientRect();
  const tail = footer.value.getBoundingClientRect().bottom - rect.bottom;
  const lead = 200;
  const progress = Math.max(0, Math.min(1, (window.innerHeight - rect.top + lead) / (rect.height + tail + lead)));
  place(6 + 62 * progress, 50);
}

function schedule() {
  if (frame) return;
  frame = requestAnimationFrame(() => {
    frame = 0;
    follow();
  });
}

function onPointerMove(event: PointerEvent) {
  if (!wordmark.value) return;
  const rect = wordmark.value.getBoundingClientRect();
  hovering = true;
  place(((event.clientX - rect.left) / rect.width) * 100, ((event.clientY - rect.top) / rect.height) * 100);
}

function onPointerLeave() {
  hovering = false;
  follow();
}

onMounted(() => {
  if (props.plain) return;
  window.addEventListener('scroll', schedule, { capture: true, passive: true });
  window.addEventListener('resize', schedule);
  follow();
});

onBeforeUnmount(() => {
  window.removeEventListener('scroll', schedule, { capture: true });
  window.removeEventListener('resize', schedule);
  if (frame) cancelAnimationFrame(frame);
  headerOverDark.value = false;
});
</script>

<template>
  <footer
    ref="footer"
    class="site-footer px-6"
    :class="plain ? 'pt-20 pb-16' : 'dark relative isolate overflow-hidden bg-stone-950 pb-9'"
  >
    <div v-if="!plain" class="site-footer__dusk pointer-events-none absolute inset-x-0 top-0 -z-20 h-150 sm:h-195" aria-hidden="true" />
    <svg v-if="!plain" class="site-footer__grain pointer-events-none absolute top-0 left-0 -z-10 h-150 w-full sm:h-195" aria-hidden="true">
      <filter id="site-footer-grain" x="0" y="0" width="100%" height="100%">
        <feTurbulence type="fractalNoise" baseFrequency="0.85" numOctaves="2" stitchTiles="stitch" />
        <feColorMatrix type="matrix" values="0.33 0.33 0.33 0 0 0.33 0.33 0.33 0 0 0.33 0.33 0.33 0 0 0 0 0 0 1" />
      </filter>
      <rect width="100%" height="100%" filter="url(#site-footer-grain)" />
    </svg>

    <div class="mx-auto max-w-content">
      <div v-if="!plain" class="pt-90 text-center sm:pt-125">
        <p class="mx-auto font-display text-4xl leading-tight text-balance text-stone-50 sm:text-5xl sm:leading-tight lg:text-6xl lg:leading-tight">
          Don’t <span class="whitespace-nowrap">half-ass</span> your DevOps.
        </p>
        <p class="mx-auto mt-6 max-w-2xl text-xl leading-normal text-balance text-stone-200">
          Get a <span class="whitespace-nowrap">whole-ass</span> platform that does it for you. The first 14 days are free, up to CHF 5 of usage.
        </p>
        <div class="mt-9 flex flex-col justify-center gap-3 sm:flex-row sm:gap-4">
          <a
            :href="`${appUrl}/login`"
            class="flex items-center justify-center gap-3 rounded-lg border border-transparent bg-stone-50 px-7 py-4 text-xl font-semibold text-stone-950 transition-colors hover:bg-white"
          >
            Deploy your first app
            <UIcon name="i-lucide-arrow-right" class="size-5.5 shrink-0" />
          </a>
          <NuxtLink
            to="/docs/quickstart"
            class="flex items-center justify-center rounded-lg border border-white/40 bg-white/10 px-7 py-4 text-xl font-semibold text-stone-50 transition-colors hover:bg-white/20"
          >
            Read the docs
          </NuxtLink>
        </div>
      </div>

      <nav aria-label="Footer" class="grid grid-cols-2 gap-x-8 gap-y-10 sm:grid-cols-3 lg:grid-cols-5" :class="{ 'mt-24 sm:mt-34': !plain }">
        <div v-for="column in columns" :key="column.title">
          <h2 class="text-base font-semibold text-neutral-900 dark:text-neutral-100">
            {{ column.title }}
          </h2>
          <ul class="mt-3.5 space-y-2.5">
            <li v-for="link in column.links" :key="link.label">
              <a
                v-if="link.href"
                :href="link.href"
                class="text-base text-neutral-600 transition-colors hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-neutral-100"
              >
                {{ link.label }}
              </a>
              <NuxtLink
                v-else
                :to="link.to"
                class="text-base text-neutral-600 transition-colors hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-neutral-100"
              >
                {{ link.label }}
              </NuxtLink>
            </li>
          </ul>
        </div>
      </nav>

      <div
        v-if="!plain"
        ref="wordmark"
        class="site-footer__wordmark relative mt-22"
        aria-hidden="true"
        @pointermove="onPointerMove"
        @pointerleave="onPointerLeave"
      >
        <div class="site-footer__spot pointer-events-none absolute -z-10" />
        <svg
          v-for="layer in wordmarkLayers"
          :key="layer"
          viewBox="-8 -737 4149 945"
          class="site-footer__layer"
          :class="`site-footer__layer--${layer}`"
        >
          <path :d="markPath" transform="translate(0 -729) scale(4.0276)" />
          <path :d="wordmarkPath" transform="translate(1452 0)" />
        </svg>
      </div>

      <div
        class="relative mt-14 flex flex-wrap items-center justify-between gap-x-8 gap-y-3 text-sm text-neutral-600 dark:text-neutral-400"
        :class="{ 'border-t border-default pt-8': plain }"
      >
        <span>Built in Switzerland, hosted in the EU</span>
        <div class="flex flex-wrap items-center gap-x-5 gap-y-3">
          <span>
            <a
              href="https://github.com/zeitlos/lucity"
              target="_blank"
              rel="noopener"
              class="underline underline-offset-3 transition-colors hover:text-neutral-900 dark:hover:text-neutral-100"
            >Open source on GitHub</a>
            · zeitlos software
          </span>
          <ThemeToggle />
        </div>
      </div>
    </div>
  </footer>
</template>

<style scoped>
@property --footer-spot-x {
  syntax: '<percentage>';
  inherits: true;
  initial-value: 68%;
}

@property --footer-spot-y {
  syntax: '<percentage>';
  inherits: true;
  initial-value: 50%;
}

.site-footer__dusk {
  background:
    linear-gradient(to bottom, oklch(0.95 0.02 80) 0%, oklch(0.95 0.02 80 / 0) 34%),
    linear-gradient(to bottom, oklch(0.20 0.03 55 / 0) 76%, oklch(0.20 0.03 55) 100%),
    radial-gradient(ellipse 30% 22% at 88% 55%, oklch(0.82 0.14 58 / 0.72) 0%, oklch(0.82 0.14 58 / 0.30) 50%, oklch(0.82 0.14 58 / 0) 100%),
    radial-gradient(ellipse 36% 22% at 40% 57%, oklch(0.70 0.21 0 / 0.40) 0%, oklch(0.70 0.21 0 / 0.15) 50%, oklch(0.70 0.21 0 / 0) 100%),
    radial-gradient(ellipse 62% 40% at 4% 84%, oklch(0.46 0.23 282 / 0.75) 0%, oklch(0.46 0.23 282 / 0.30) 50%, oklch(0.46 0.23 282 / 0) 100%),
    radial-gradient(ellipse 150% 100% at 50% 0% in oklab, oklch(0.95 0.02 80) 0%, oklch(0.95 0.02 80) 24%, oklch(0.925 0.045 72) 35%, oklch(0.86 0.09 52) 44%, oklch(0.74 0.15 18) 52%, oklch(0.60 0.19 335) 60%, oklch(0.46 0.19 302) 69%, oklch(0.30 0.10 305) 83%, oklch(0.20 0.03 55) 100%);
}

html.dark .site-footer__dusk {
  background:
    linear-gradient(to bottom, oklch(0.20 0.03 55) 0%, oklch(0.20 0.03 55 / 0) 34%),
    linear-gradient(to bottom, oklch(0.20 0.03 55 / 0) 76%, oklch(0.20 0.03 55) 100%),
    radial-gradient(ellipse 30% 22% at 88% 55%, oklch(0.72 0.15 50 / 0.42) 0%, oklch(0.72 0.15 50 / 0.16) 50%, oklch(0.72 0.15 50 / 0) 100%),
    radial-gradient(ellipse 36% 22% at 40% 57%, oklch(0.62 0.21 0 / 0.34) 0%, oklch(0.62 0.21 0 / 0.12) 50%, oklch(0.62 0.21 0 / 0) 100%),
    radial-gradient(ellipse 62% 40% at 4% 84%, oklch(0.46 0.23 282 / 0.60) 0%, oklch(0.46 0.23 282 / 0.24) 50%, oklch(0.46 0.23 282 / 0) 100%),
    radial-gradient(ellipse 150% 100% at 50% 0% in oklab, oklch(0.20 0.03 55) 0%, oklch(0.20 0.03 55) 24%, oklch(0.24 0.05 20) 35%, oklch(0.32 0.11 5) 44%, oklch(0.42 0.16 350) 52%, oklch(0.45 0.19 328) 60%, oklch(0.40 0.18 302) 69%, oklch(0.28 0.10 305) 83%, oklch(0.20 0.03 55) 100%);
}

.site-footer__grain {
  mix-blend-mode: soft-light;
  opacity: 0.55;
  -webkit-mask-image: linear-gradient(to bottom, transparent 28%, #000 60%, #000 82%, transparent 100%);
  mask-image: linear-gradient(to bottom, transparent 28%, #000 60%, #000 82%, transparent 100%);
}

.site-footer__wordmark {
  transition:
    --footer-spot-x 0.45s cubic-bezier(0.2, 0.7, 0.2, 1),
    --footer-spot-y 0.45s cubic-bezier(0.2, 0.7, 0.2, 1);
}

.site-footer__spot {
  inset: -95% -40%;
  background: radial-gradient(ellipse 19.3% 30.1% at calc(22.22% + var(--footer-spot-x, 68%) / 1.8) calc(32.76% + var(--footer-spot-y, 50%) / 2.9), oklch(0.50 0.21 300 / 0.34) 0%, oklch(0.50 0.21 300 / 0) 100%);
}

.site-footer__layer {
  display: block;
  width: 100%;
  height: auto;
  fill: none;
  stroke-width: 1.5px;
}

.site-footer__layer path {
  vector-effect: non-scaling-stroke;
}

.site-footer__layer--base {
  position: relative;
  stroke: oklch(0.40 0.04 62);
}

.site-footer__layer--glow,
.site-footer__layer--light {
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
}

.site-footer__layer--glow {
  fill: oklch(0.55 0.21 300 / 0.13);
  stroke: oklch(0.68 0.19 302);
  -webkit-mask-image: radial-gradient(ellipse 40% 150% at var(--footer-spot-x, 68%) var(--footer-spot-y, 50%), #000 0%, transparent 100%);
  mask-image: radial-gradient(ellipse 40% 150% at var(--footer-spot-x, 68%) var(--footer-spot-y, 50%), #000 0%, transparent 100%);
}

.site-footer__layer--light {
  stroke: oklch(0.95 0.035 80);
  -webkit-mask-image: radial-gradient(ellipse 19% 78% at var(--footer-spot-x, 68%) var(--footer-spot-y, 50%), #000 0%, transparent 100%);
  mask-image: radial-gradient(ellipse 19% 78% at var(--footer-spot-x, 68%) var(--footer-spot-y, 50%), #000 0%, transparent 100%);
}

@media (prefers-reduced-motion: reduce) {
  .site-footer__wordmark {
    transition: none;
  }
}
</style>
