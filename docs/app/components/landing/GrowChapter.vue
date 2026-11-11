<script setup lang="ts">
interface Part {
  name: string;
  text: string;
  ticks: string[];
  link: { to: string; label: string };
  sketch: string;
  sketchWidth?: string;
  shot: string;
  shotAlt: string;
  shotDark?: boolean;
}

const parts: Part[] = [
  {
    name: 'Databases',
    text: 'PostgreSQL on two machines with automatic failover. Every write is archived as it happens, so you can restore to any second of the last 30 days.',
    ticks: ['Table browser and SQL editor in the dashboard', 'Metrics, and the queries taking the most time', 'Private by default, open only to the addresses you list', 'Vector search included'],
    link: { to: '/docs/postgres', label: 'More on databases' },
    sketch: 'sketch-postgres.svg',
    shot: 'database-metrics.webp',
    shotAlt: 'Database metrics in the dashboard: CPU, memory, storage, connections, transactions, cache hit ratio and replication lag',
  },
  {
    name: 'Storage',
    text: 'A Redis-compatible store for caches, sessions and queues. S3-compatible buckets for uploads. Volumes for software that writes to disk.',
    ticks: ['Credentials are linked into your service, never copied around', 'Browse, upload and download files in the dashboard', 'Public buckets served from a CDN', 'Volumes grow while the service keeps running'],
    link: { to: '/docs/object-storage', label: 'More on storage' },
    sketch: 'sketch-storage.svg',
    shot: 'variables.webp',
    shotAlt: "The Variables tab of a service: DATABASE_URL linked to the database, and the bucket's endpoint, region and keys linked to the bucket",
  },
  {
    name: 'Scaling',
    text: 'Run several instances, with traffic spread across the ones that are ready. Autoscaling adds more as CPU climbs and removes them once it stays calm.',
    ticks: ['Up to ten instances per service on Pro', 'CPU and memory set separately, changed without a rebuild', 'Billed on what your apps actually use'],
    link: { to: '/docs/services#replicas-and-autoscaling', label: 'More on scaling' },
    sketch: 'sketch-scaling.svg',
    shot: 'scaling.webp',
    shotAlt: 'Scaling settings of a service: three replicas, an autoscaling switch, and CPU and memory limits per instance',
  },
  {
    name: 'Releases',
    text: "The new version takes traffic once it's healthy, and the old one finishes its requests. A broken deploy never replaces a working one.",
    ticks: ['Roll back to a previous deployment', 'Development, staging and production as copies of one project', 'A failed rollout says why, like out of memory or a crash loop'],
    link: { to: '/docs/deployments', label: 'More on releases' },
    sketch: 'sketch-releases.svg',
    shot: 'deployments.webp',
    shotAlt: 'The Deployments tab of a service with the active deployment and the history of superseded ones',
  },
  {
    name: 'Domains and edge',
    text: 'Your own hostname with two DNS records, shown ready to copy. Visitors connect to a location near them.',
    ticks: ['Certificates issued and renewed for you', 'example.com itself, not only www', 'Redirect an old domain to a new one, keeping the path'],
    link: { to: '/docs/domains', label: 'More on domains' },
    sketch: 'sketch-domains.svg',
    sketchWidth: 'min(100%, 300px)',
    shot: 'domains.webp',
    shotAlt: 'Domain settings of a service: the platform domain and a custom domain whose certificate is being issued',
  },
  {
    name: 'Logs and metrics',
    text: 'Follow every instance live, or search its history, even for instances that are long gone.',
    ticks: ['A timeline of log volume, with crashes marked', 'Starts, restarts and failed health checks between your own lines', 'CPU and memory per service, over time'],
    link: { to: '/docs/logs', label: 'More on logs' },
    sketch: 'sketch-logs.svg',
    shot: 'logs.webp',
    shotAlt: 'The log explorer: a search field, time ranges, a timeline with one lane per instance, and log lines colored by instance',
    shotDark: true,
  },
];

const extras = [
  { label: 'team roles', to: '/features#team' },
  { label: 'private networking', to: '/features#services' },
  { label: 'shared variables', to: '/features#services' },
  { label: 'monorepos', to: '/features#builds' },
  { label: 'an API', to: '/features#tools' },
];

const open = ref(0);
const current = computed(() => parts[open.value]!);
</script>

<template>
  <section class="mx-auto max-w-content px-6" aria-labelledby="grow">
    <LandingChapterHeading id="grow" beat="grow" promise="Built to grow" rest="well past the prototype.">
      Add what production needs, when it needs it. Everything lives in the same project and is wired to your services for you.
    </LandingChapterHeading>

    <div class="mt-14 flex flex-wrap items-start gap-x-16 gap-y-10">
      <div class="min-w-0 flex-[1_1_400px] border-b border-stone-300 dark:border-stone-800">
        <div v-for="(part, index) in parts" :key="part.name" class="border-t border-stone-300 dark:border-stone-800">
          <h3>
            <button
              :id="`part-${index}`"
              type="button"
              class="row-button flex min-h-11 w-full cursor-pointer items-center justify-between gap-6 py-4.75 text-left text-xl font-semibold transition-colors hover:text-stone-900 dark:hover:text-stone-100"
              :class="open === index ? 'text-stone-900 dark:text-stone-100' : 'text-stone-700 dark:text-stone-400'"
              :aria-expanded="open === index"
              :aria-controls="`part-panel-${index}`"
              @click="open = index"
            >
              <span>{{ part.name }}</span>
              <UIcon name="i-lucide-chevron-down" class="size-5 shrink-0 transition-transform duration-300 ease-out motion-reduce:transition-none" :class="{ 'rotate-180': open === index }" />
            </button>
          </h3>
          <div :id="`part-panel-${index}`" role="region" :aria-labelledby="`part-${index}`" :inert="open !== index" class="part-panel" :class="{ 'is-open': open === index }">
            <div class="min-h-0 overflow-hidden">
              <div class="part-body pb-6 sm:pr-10">
                <p class="text-lg/relaxed text-stone-700 dark:text-stone-400">{{ part.text }}</p>
                <ul class="ticks mt-3.5 flex flex-col gap-1.75 text-base text-stone-800 dark:text-stone-300">
                  <li v-for="tick in part.ticks" :key="tick" class="flex items-baseline gap-2.5">{{ tick }}</li>
                </ul>
                <p class="mt-4 text-base">
                  <NuxtLink :to="part.link.to" class="font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">{{ part.link.label }}</NuxtLink>
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <LandingStage tint="grow" class="min-w-0 flex-[1.55_1_560px] px-4 pt-6 sm:px-8">
        <div class="scene">
          <img class="sketch sketch-day" :src="`/img/landing/${current.sketch}`" alt="" width="340" height="260" :style="current.sketchWidth ? { width: current.sketchWidth } : undefined">
          <img class="sketch sketch-night" :src="`/img/landing/${current.sketch.replace('.svg', '-night.svg')}`" alt="" width="340" height="260" :style="current.sketchWidth ? { width: current.sketchWidth } : undefined">
          <div class="rise" :class="{ 'rise-dark': current.shotDark }">
            <img :src="`/img/landing/${current.shot}`" :alt="current.shotAlt">
          </div>
        </div>
      </LandingStage>
    </div>

    <p class="mt-8 text-base text-stone-700 dark:text-stone-400">
      Also included:
      <template v-for="(extra, index) in extras" :key="extra.label">
        <NuxtLink :to="extra.to" class="font-medium text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">{{ extra.label }}</NuxtLink>{{ index < extras.length - 2 ? ', ' : index === extras.length - 2 ? ' and ' : '. ' }}
      </template>
      <NuxtLink to="/features" class="font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">See every feature</NuxtLink>
    </p>
  </section>
</template>

<style scoped>
.part-panel {
  display: grid;
  grid-template-rows: 0fr;
  transition: grid-template-rows 0.35s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.part-panel.is-open {
  grid-template-rows: 1fr;
}

.part-body {
  opacity: 0;
  translate: 0 -0.375rem;
  transition:
    opacity 0.2s ease,
    translate 0.35s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.is-open .part-body {
  opacity: 1;
  translate: none;
  transition:
    opacity 0.3s ease 0.08s,
    translate 0.35s cubic-bezier(0.2, 0.8, 0.2, 1);
}

@media (prefers-reduced-motion: reduce) {
  .part-panel,
  .part-body,
  .is-open .part-body {
    transition: none;
  }
}

.row-button:focus-visible {
  outline: 2px solid oklch(0.55 0.15 300);
  outline-offset: 2px;
}

.ticks li::before {
  content: "";
  flex: none;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: oklch(0.58 0.17 35);
  transform: translateY(-2px);
}

.scene {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  min-height: 640px;
}

@media (max-width: 767px) {
  .scene {
    min-height: 0;
    gap: 28px;
  }
}

.sketch {
  display: block;
  width: min(100%, 370px);
  height: auto;
}

.sketch-night {
  display: none;
}

.dark .sketch-night {
  display: block;
}

.dark .sketch-day {
  display: none;
}

.rise {
  width: 100%;
  max-width: 776px;
  max-height: 372px;
  border-radius: 14px 14px 0 0;
  overflow: hidden;
  background: oklch(0.975 0.012 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .12), 0 28px 56px -22px oklch(0.25 0.04 55 / .42);
}

.rise-dark {
  background: oklch(0.16 0.01 55);
}

.rise img {
  display: block;
  width: 100%;
  height: auto;
}
</style>
