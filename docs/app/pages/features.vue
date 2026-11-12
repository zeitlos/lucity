<script setup lang="ts">
definePageMeta({
  layout: 'landing',
  header: false,
  footer: false,
});

useSeo({
  title: 'Features',
  description: 'The parts you need to run what you build: builds, deploys, services, domains, Postgres, storage, logs and metrics, tools, team, billing, security and ownership.',
  type: 'website',
});

interface Feature {
  name: string;
  text: string;
  to?: string;
  href?: string;
  limit?: string[];
}

interface Group {
  id: string;
  title: string;
  intro: string;
  features: Feature[];
}

const appUrl = '/app';

const sketches = import.meta.glob<string>('../assets/features/*.svg', {
  query: '?raw',
  import: 'default',
  eager: true,
});

function sketch(id: string) {
  return sketches[`../assets/features/${id}.svg`];
}

const groups: Group[] = [
  {
    id: 'builds',
    title: 'Builds',
    intro: 'From a repository or an image to something that runs, no Dockerfile required.',
    features: [
      { name: 'Automatic builds', to: '/docs/builds', text: 'Lucity detects the framework and builds it for you.' },
      { name: 'Prebuilt images', to: '/docs/services', text: 'Any public image, like Ghost or Grafana, found with a search in the dashboard.' },
      { name: 'Build cache', to: '/docs/builds#caching', text: 'Every build starts from the last one, so only what changed is rebuilt.' },
      { name: 'Monorepos', to: '/docs/builds#root-directory', text: 'Point each service at its own folder of the same repository.' },
      { name: 'Build settings', to: '/docs/builds#build-configuration', text: 'Override the install, build and start commands, or add system packages.' },
      { name: 'Framework guides', to: '/docs/guides', text: 'Next.js, Nuxt, SvelteKit, Astro, Express, FastAPI, Django, Laravel and Go.' },
    ],
  },
  {
    id: 'deploys',
    title: 'Deploys',
    intro: 'Getting a new version live, and back again if it misbehaves.',
    features: [
      { name: 'Push to deploy', to: '/docs/deployments', text: 'Every push to the tracked branch deploys, and one commit can roll out every service built from that repository.' },
      { name: 'Deploys from GitHub Actions', to: '/docs/deployments#triggering-a-deployment', text: 'Deploy once your tests pass, with no token stored anywhere.' },
      { name: 'Zero-downtime rollouts', to: '/docs/deployments#rollouts', text: "The new version takes traffic once it's healthy, and the old one finishes its requests." },
      { name: 'Changes in seconds', to: '/docs/deployments#configuration-changes', text: 'Variables, resources and the start command apply without a rebuild.' },
      { name: 'Failure reasons', to: '/docs/deployments#rollout-failures', text: 'A failed rollout says why, like out of memory or a crash loop, while the old version keeps serving.' },
      { name: 'Rollbacks', to: '/docs/deployments#rollbacks', text: 'Back to a previous deployment when a release goes wrong.' },
      { name: 'Environments', to: '/docs/environments', text: 'Development, staging and production as separate copies of one project.' },
    ],
  },
  {
    id: 'services',
    title: 'Services',
    intro: 'How each part of your app runs: its size, its count and its settings.',
    features: [
      { name: 'Start command', to: '/docs/services#start-command', text: 'Keep the detected one or set your own.' },
      { name: 'Health checks', to: '/docs/services#health-checks', text: 'A connection check by default, or an HTTP check with extra time for slow starts.' },
      { name: 'CPU and memory', to: '/docs/services#cpu-and-memory', text: 'Set each one separately instead of picking a fixed package, and change them without a rebuild.', limit: ['Hobby 2 vCPU, 4 GB', 'Pro 4 vCPU, 16 GB'] },
      { name: 'Replicas', to: '/docs/services#replicas-and-autoscaling', text: 'Run several instances, with traffic spread across the ones that are ready.', limit: ['Hobby 3 · Pro 10'] },
      { name: 'Autoscaling', to: '/docs/services#replicas-and-autoscaling', text: 'More instances as CPU climbs, fewer once it stays calm.' },
      { name: 'Private networking', to: '/docs/services#networking', text: 'Services reach each other by name, with no public address needed.' },
      { name: 'Variables', to: '/docs/variables', text: 'Per service, available while it builds and while it runs.' },
      { name: 'Shared variables', to: '/docs/variables#shared-variables', text: 'Set a value once for the whole environment.' },
      { name: 'References', to: '/docs/variables#dynamic-variables', text: "Pick a database's connection string from a list instead of copying passwords around." },
    ],
  },
  {
    id: 'domains',
    title: 'Domains',
    intro: 'How visitors reach your app.',
    features: [
      { name: 'Platform domains', to: '/docs/domains#platform-domains', text: 'A public address with HTTPS for every service, ready in seconds.' },
      { name: 'Custom domains', to: '/docs/domains#custom-domains', text: 'Your own hostname with two DNS records, shown ready to copy.', limit: ['Hobby 1 · Pro unlimited'] },
      { name: 'Apex domains', to: '/docs/domains#apex-domain', text: 'example.com itself, not only www.' },
      { name: 'Automatic certificates', to: '/docs/domains#verification-and-certificates', text: 'Issued as soon as a domain checks out, and renewed for you.' },
      { name: 'HTTPS only', to: '/docs/domains', text: 'Plain HTTP requests are redirected to HTTPS.' },
      { name: 'Domain redirects', to: '/docs/domains', text: 'Send www to your apex, or an old domain to a new one, keeping the path.' },
      { name: 'Global edge', text: 'Visitors connect to a location near them, with nothing to change in DNS.' },
    ],
  },
  {
    id: 'postgres',
    title: 'Postgres',
    intro: 'Managed PostgreSQL with failover and backups on every plan.',
    features: [
      { name: 'High availability', to: '/docs/postgres', text: 'Two instances on separate machines, with automatic failover.' },
      { name: 'Flexible sizing', to: '/docs/postgres#resources-and-metrics', text: 'CPU, memory and storage are set one by one, not bundled into fixed small, medium and large packages. Change any of them without downtime.', limit: ['Hobby 32 GB · Pro 128 GB'] },
      { name: 'Continuous backups', to: '/docs/postgres#backups', text: 'Every write is archived as it happens and kept for 30 days, with nothing to configure.' },
      { name: 'Point-in-time recovery', to: '/docs/postgres#point-in-time-restores', text: 'Restore to any second of the last 30 days, into a new database next to the original.' },
      { name: 'Backups on demand', to: '/docs/postgres#backups', text: "Take a full backup before a migration you're unsure about." },
      { name: 'Allowed addresses', to: '/docs/postgres#allowed-addresses', text: 'Private by default. Open it to the internet for the addresses you list, and nobody else.' },
      { name: 'Encrypted connections', to: '/docs/postgres#public-access', text: 'Public connections are encrypted all the way to the database, not to a proxy in front of it.' },
      { name: 'Table browser', to: '/docs/postgres#database-explorer', text: 'Read tables and rows in the dashboard to check a migration did what you expected.' },
      { name: 'SQL editor', to: '/docs/postgres#database-explorer', text: 'Run queries against the database from the dashboard.' },
      { name: 'Database metrics', to: '/docs/metrics#database-metrics', text: 'Connections, transactions, cache hits and replica lag.' },
      { name: 'Top queries', to: '/docs/postgres#resources-and-metrics', text: 'The statements taking the most time, so you know what to fix first.' },
      { name: 'Vector search', to: '/docs/postgres', text: 'pgvector is included, for embeddings and similarity search.' },
    ],
  },
  {
    id: 'storage',
    title: 'Storage',
    intro: "Somewhere for everything that isn't rows in a table.",
    features: [
      { name: 'Key-value store', to: '/docs/redis', text: 'Redis-compatible, for caches, sessions and job queues.', limit: ['Hobby 32 GB · Pro 128 GB'] },
      { name: 'Buckets', to: '/docs/object-storage', text: 'S3-compatible storage for uploads and files.' },
      { name: 'Public buckets', to: '/docs/object-storage#public-read', text: 'Serve files to everyone from a public address on a CDN.' },
      { name: 'File explorer', to: '/docs/object-storage#file-explorer', text: "Browse, upload and download a bucket's files in the dashboard." },
      { name: 'Volumes', to: '/docs/volumes', text: 'Disks that outlive deploys, for software that writes to its filesystem.', limit: ['Hobby 32 GB · Pro 128 GB'] },
      { name: 'Volume resizing', to: '/docs/volumes#size', text: 'Add space while the service keeps running.' },
      { name: 'Volume usage', to: '/docs/metrics#volume-usage', text: 'How full a volume is, over time.' },
    ],
  },
  {
    id: 'observe',
    title: 'Logs and metrics',
    intro: 'What your app is doing right now, and what it did last night.',
    features: [
      { name: 'Live logs', to: '/docs/logs#runtime-logs', text: 'Follow every instance as it prints, including the new ones a rollout starts.' },
      { name: 'Log history', to: '/docs/logs#runtime-logs', text: 'Look back an hour, a day or a week, even for instances that are long gone.' },
      { name: 'Log search', to: '/docs/logs#runtime-logs', text: 'Find a phrase across every instance, highlighted, with the lines around each match.' },
      { name: 'Instance filter', to: '/docs/logs#runtime-logs', text: 'Narrow the logs to one instance when several are running.' },
      { name: 'Log timeline', to: '/docs/logs#runtime-logs', text: 'How much was printed over time, with errors in their own color, so spikes stand out.' },
      { name: 'Platform events', to: '/docs/logs#runtime-logs', text: "Starts, restarts and failed health checks appear between your app's own lines." },
      { name: 'Log download', to: '/docs/logs#runtime-logs', text: "Save what you're looking at as a text file." },
      { name: 'Deployment logs', to: '/docs/logs#deployment-logs', text: 'Build and deploy output, streamed live and kept for a week.' },
      { name: 'Service metrics', to: '/docs/metrics#service-metrics', text: 'CPU and memory per service, over time.' },
    ],
  },
  {
    id: 'tools',
    title: 'Tools',
    intro: 'Work from a terminal, a pipeline or your agent.',
    features: [
      { name: 'CLI', to: '/docs/cli', text: 'Deploy, check status and manage variables from a terminal or any CI pipeline.' },
      { name: 'Claude plugin and MCP', to: '/docs/claude-plugin', text: 'Deploy and debug from Claude Code, Codex, Cursor or any other MCP client.' },
      { name: 'API', to: '/docs/api', text: 'Everything the dashboard does, with an interactive playground.' },
      { name: 'API tokens', to: '/docs/workspaces#api-tokens', text: 'Tokens for scripts and pipelines, each with its own role.' },
    ],
  },
  {
    id: 'team',
    title: 'Team',
    intro: 'Bring people in without sharing a password.',
    features: [
      { name: 'Workspaces', to: '/docs/workspaces', text: 'Your organization in Lucity, with its own projects and billing.' },
      { name: 'Members and roles', to: '/docs/workspaces#members-and-roles', text: 'Invite colleagues as admins or members.', limit: ['Pro'] },
      { name: 'Personal workspace', to: '/docs/workspaces#personal-workspaces', text: 'A space of your own for experiments and side projects.' },
    ],
  },
  {
    id: 'billing',
    title: 'Billing',
    intro: 'You pay for what your apps use, not for machines sitting idle.',
    features: [
      { name: 'Pay for what you use', to: '/docs/billing#what-you-pay-for', text: 'CPU and memory are billed on what your apps actually use, not a fixed price per app or machine, so an idle service costs very little.' },
      { name: 'Plan credits', to: '/docs/billing', text: "Your plan's monthly fee comes back as usage credit, so you only pay for usage beyond it." },
      { name: 'Usage so far', to: '/docs/billing#tracking-usage', text: 'What this billing period has cost so far, with an estimated total.' },
      { name: 'Free trial', to: '/docs/billing#the-free-trial', text: '14 days to try everything, no credit card needed.' },
    ],
  },
  {
    id: 'security',
    title: 'Security',
    intro: 'On for every app, without asking.',
    features: [
      { name: 'Secret scanning', to: '/docs/deployments#deployment-steps', text: 'Every deploy checks your repository and its history, and flags keys that still work.' },
      { name: 'Vulnerability checks', text: 'Running apps Lucity built are rechecked every two hours.' },
      { name: 'Workspace isolation', text: "Apps in one workspace can't reach another's apps or databases." },
      { name: 'Environment isolation', to: '/docs/environments#what-stays-separate', text: "Staging can't reach production's private network." },
      { name: 'Secret-free agent tools', to: '/docs/claude-plugin', text: 'Agent tools show variable names and references, never the values.' },
    ],
  },
  {
    id: 'ownership',
    title: 'Ownership',
    intro: "Where it lives, and how you'd leave.",
    features: [
      { name: 'Hosting in the EU', text: 'Apps and databases in Germany, files and backups in France.' },
      { name: 'Open source', href: 'https://github.com/zeitlos/lucity', text: 'Read every line of the platform.' },
      { name: 'Self-hosting', to: '/docs/self-hosting', text: 'Run the whole platform on your own infrastructure.' },
      { name: 'Eject', to: '/docs/eject', text: 'Download any project in a standard format and run it on any major cloud, without Lucity.' },
    ],
  },
];

const comparisons = [
  { label: 'vs Heroku', to: '/comparisons/vs-heroku' },
  { label: 'vs Vercel', to: '/comparisons/vs-vercel' },
  { label: 'vs Railway', to: '/comparisons/vs-railway' },
  { label: 'vs Render', to: '/comparisons/vs-render' },
  { label: 'vs Fly.io', to: '/comparisons/vs-fly-io' },
  { label: 'vs Coolify', to: '/comparisons/vs-coolify' },
  { label: 'vs Deploio', to: '/comparisons/vs-deploio' },
];
</script>

<template>
  <div class="pb-28 md:pb-36">
    <section class="mx-auto max-w-content px-6 pt-20 sm:pt-32" aria-labelledby="features-title">
      <h1
        id="features-title"
        class="load-in font-display text-4xl/[1.15] text-neutral-800 max-sm:text-balance sm:text-5xl/[1.1] md:text-6xl/[1.1] lg:text-7xl/22 dark:text-neutral-100"
        style="--load-order: 0"
      >
        The parts you need <br class="max-sm:hidden">
        to run what you build.
      </h1>
      <p class="load-in mt-6 max-w-3xl text-xl leading-relaxed text-neutral-700 sm:text-2xl sm:leading-relaxed dark:text-neutral-300" style="--load-order: 1">
        Add a service, a database or a domain to a project and it runs, already wired to the rest. You only pay for what it uses.
      </p>

      <div class="load-in mt-9 flex flex-col gap-3 sm:flex-row sm:items-center sm:gap-4" style="--load-order: 2">
        <a
          :href="`${appUrl}/login`"
          class="flex items-center justify-center gap-3 rounded-lg border border-transparent bg-primary px-6 py-3.5 text-lg font-semibold text-white shadow-sm transition-colors hover:bg-primary-600"
        >
          Deploy your first app
          <UIcon name="i-lucide-arrow-right" class="size-5 shrink-0" />
        </a>
        <NuxtLink
          to="/pricing"
          class="flex items-center justify-center rounded-lg border border-stone-300 bg-white/50 px-6 py-3.5 text-lg font-semibold text-neutral-800 transition-colors hover:bg-white dark:border-stone-700 dark:bg-white/5 dark:text-neutral-100 dark:hover:bg-white/10"
        >
          See what it costs
        </NuxtLink>
      </div>

      <nav
        aria-label="Groups on this page"
        class="load-in mt-18 grid grid-cols-2 gap-x-8 border-b border-default sm:grid-cols-3 md:grid-cols-4 xl:grid-cols-6"
        style="--load-order: 3"
      >
        <a
          v-for="group in groups"
          :key="group.id"
          :href="`#${group.id}`"
          class="group flex items-center justify-between gap-3 border-t border-default py-3 text-lg font-semibold text-neutral-800 dark:text-neutral-100"
        >
          {{ group.title }}
          <UIcon
            name="i-lucide-arrow-down"
            class="size-4.5 shrink-0 text-neutral-500 transition duration-200 group-hover:translate-y-0.5 group-hover:text-neutral-900 motion-reduce:transition-none dark:group-hover:text-neutral-100"
          />
        </a>
      </nav>
    </section>

    <section
      v-for="group in groups"
      :id="group.id"
      :key="group.id"
      :aria-labelledby="`${group.id}-title`"
      class="mx-auto mt-24 flex max-w-content scroll-mt-28 flex-col gap-8 px-6 lg:mt-28 lg:flex-row lg:items-start lg:gap-18"
    >
      <div class="lg:sticky lg:top-28 lg:w-85 lg:shrink-0">
        <h2
          :id="`${group.id}-title`"
          class="font-display text-4xl leading-tight text-neutral-800 lg:text-5xl lg:leading-tight dark:text-neutral-100"
        >
          {{ group.title }}
        </h2>
        <p class="mt-3.5 text-lg text-neutral-600 dark:text-neutral-400">
          {{ group.intro }}
        </p>
        <div class="feature-sketch mt-7 hidden lg:block" aria-hidden="true" v-html="sketch(group.id)" />
      </div>

      <ul class="min-w-0 flex-1 border-b border-default">
        <li
          v-for="feature in group.features"
          :key="feature.name"
          class="grid items-baseline gap-x-6 gap-y-0.5 border-t border-default py-4 sm:grid-cols-[1fr_auto]"
        >
          <div>
            <NuxtLink
              v-if="feature.to"
              :to="feature.to"
              class="text-lg font-semibold text-neutral-900 underline decoration-stone-400 decoration-1 underline-offset-4 transition-colors hover:decoration-current dark:text-neutral-100 dark:decoration-stone-600"
            >
              {{ feature.name }}
            </NuxtLink>
            <a
              v-else-if="feature.href"
              :href="feature.href"
              target="_blank"
              rel="noopener"
              class="text-lg font-semibold text-neutral-900 underline decoration-stone-400 decoration-1 underline-offset-4 transition-colors hover:decoration-current dark:text-neutral-100 dark:decoration-stone-600"
            >
              {{ feature.name }}
            </a>
            <span v-else class="text-lg font-semibold text-neutral-900 dark:text-neutral-100">
              {{ feature.name }}
            </span>
            <span class="mt-0.5 block text-base text-neutral-600 dark:text-neutral-400">
              {{ feature.text }}
            </span>
          </div>
          <span
            v-if="feature.limit"
            class="font-mono text-sm text-neutral-600 sm:text-right sm:whitespace-nowrap dark:text-neutral-400"
          >
            <span v-for="line in feature.limit" :key="line" class="block">{{ line }}</span>
          </span>
        </li>
      </ul>
    </section>

    <section class="mx-auto mt-28 max-w-content px-6 md:mt-36" aria-labelledby="switch">
      <h2 id="switch" class="font-display text-3xl leading-tight text-neutral-800 md:text-4xl dark:text-neutral-100">
        Coming from somewhere else?
      </h2>
      <p class="mt-3.5 text-lg text-neutral-600 dark:text-neutral-400">
        The same list, side by side with what you use today.
      </p>

      <div class="mt-7 grid grid-cols-2 gap-3.5 sm:grid-cols-4 xl:grid-cols-7">
        <NuxtLink
          v-for="comparison in comparisons"
          :key="comparison.to"
          :to="comparison.to"
          class="flex items-center justify-between gap-3 rounded-xl border border-default px-5 py-5 text-lg font-semibold text-neutral-800 transition-colors hover:bg-white/50 dark:text-neutral-100 dark:hover:bg-white/5"
        >
          {{ comparison.label }}
          <UIcon name="i-lucide-arrow-right" class="size-4.5 shrink-0" />
        </NuxtLink>
      </div>

      <div class="mt-12 flex flex-wrap justify-between gap-x-12 gap-y-4 border-t border-default pt-7 text-lg">
        <p class="text-neutral-600 dark:text-neutral-400">
          Missing something you need?
          <a
            href="https://github.com/zeitlos/lucity/issues"
            target="_blank"
            rel="noopener"
            class="font-semibold text-neutral-900 underline decoration-1 underline-offset-4 dark:text-neutral-100"
          >Tell us on GitHub</a>
          or
          <a
            href="mailto:christian@zeitlos.software"
            class="font-semibold text-neutral-900 underline decoration-1 underline-offset-4 dark:text-neutral-100"
          >by email</a>.
        </p>
        <NuxtLink
          to="/pricing"
          class="inline-flex items-center gap-2 font-semibold text-neutral-900 underline decoration-1 underline-offset-4 dark:text-neutral-100"
        >
          See what it costs
          <UIcon name="i-lucide-arrow-right" class="size-4.5 shrink-0" />
        </NuxtLink>
      </div>
    </section>
  </div>
</template>

<style>
.feature-sketch {
  --sketch-ink: #301c0e;
  --sketch-paper: #fbf7ef;
  --sketch-shade: #e9dfd0;
  --sketch-accent: var(--brand-mark);
  --sketch-accent-soft: #efe6fb;
  --sketch-accent-mid: #cdb4f2;
  --sketch-alert: #d9442f;
  --sketch-ok: #149e6c;
}

.dark .feature-sketch {
  --sketch-ink: oklch(0.88 0.03 80);
  --sketch-paper: oklch(0.25 0.03 55);
  --sketch-shade: oklch(0.33 0.03 55);
  --sketch-accent-soft: oklch(0.31 0.07 300);
  --sketch-accent-mid: oklch(0.46 0.13 300);
  --sketch-alert: oklch(0.68 0.19 30);
  --sketch-ok: oklch(0.72 0.16 160);
}

.feature-sketch svg {
  display: block;
  width: 100%;
  max-width: 340px;
  height: auto;
  overflow: visible;
}

.feature-sketch .scan {
  animation: illscan 4.5s ease-in-out infinite alternate;
}

.feature-sketch .fl {
  animation: illbob 4.6s ease-in-out infinite alternate;
}

.feature-sketch .hv {
  transition: transform 0.7s cubic-bezier(0.2, 0.7, 0.2, 1);
}

.feature-sketch:hover .hv {
  transform: translateY(var(--hy, -8px));
}

.feature-sketch .pl {
  animation: illpulse 1.6s ease-in-out infinite alternate;
}

.feature-sketch .gh {
  animation: illghost 7s ease-in-out infinite;
}

.feature-sketch .fw {
  animation: illflow 1.7s linear infinite;
}

.feature-sketch:hover .fw {
  animation-duration: 0.9s;
}

.feature-sketch .ants {
  animation: illants 4s linear infinite;
}

.feature-sketch .rp {
  transform-box: fill-box;
  transform-origin: center;
  animation: illripple 2.6s ease-out infinite;
}

.feature-sketch .tm {
  animation: illtime 12s linear infinite;
}

.feature-sketch .tl {
  animation: illt0 12s steps(1, end) infinite;
}

.feature-sketch .hit {
  pointer-events: all;
}

.feature-sketch .lyi {
  transition: transform 0.45s cubic-bezier(0.2, 0.7, 0.2, 1);
}

.feature-sketch .ly:hover .lyi {
  transform: translateY(-9px);
}

.feature-sketch .ly:hover ~ .ly .lyi {
  transform: translateY(-18px);
}

.feature-sketch .rf {
  transition: fill 0.15s;
}

.feature-sketch .rl {
  transition: stroke 0.15s;
}

.feature-sketch .mk {
  transition: opacity 0.15s;
}

.feature-sketch .rg:hover .rf {
  fill: var(--sketch-accent-soft);
}

.feature-sketch .rg:hover .rl {
  stroke: var(--sketch-accent);
  stroke-width: 1.4px;
}

.feature-sketch svg:has(.rg:hover) .rg.hot:not(:hover) .rf {
  fill: var(--sketch-paper);
}

.feature-sketch svg:has(.rg:hover) .rg.hot:not(:hover) .rl {
  stroke: var(--sketch-ink);
  stroke-width: 1px;
}

.feature-sketch svg:has(.rg:not(.hot):hover) .mk {
  opacity: 0;
}

@keyframes illscan {
  from { transform: translate(-22px, -12.7px); }
  to { transform: translate(20.8px, 12px); }
}

@keyframes illbob {
  from { transform: translateY(-2.5px); }
  to { transform: translateY(2.5px); }
}

@keyframes illpulse {
  from { opacity: 1; }
  to { opacity: 0.2; }
}

@keyframes illghost {
  0%, 12% { opacity: 0.15; }
  35%, 80% { opacity: 1; }
  100% { opacity: 0.15; }
}

@keyframes illflow {
  to { stroke-dashoffset: -18.1; }
}

@keyframes illants {
  to { stroke-dashoffset: -30; }
}

@keyframes illripple {
  from { transform: scale(0.5); opacity: 0.9; }
  to { transform: scale(2.2); opacity: 0; }
}

@keyframes illtime {
  0%, 100% { transform: translateY(21.5px); }
  50% { transform: translateY(-21.5px); }
}

@keyframes illt0 {
  0%, 93.75%, 100% { opacity: 1; }
  6.25%, 87.5% { opacity: 0; }
}

@keyframes illt1 {
  0%, 12.5%, 81.25%, 93.75%, 100% { opacity: 0; }
  6.25%, 87.5% { opacity: 1; }
}

@keyframes illt2 {
  0%, 6.25%, 18.75%, 75%, 87.5%, 100% { opacity: 0; }
  12.5%, 81.25% { opacity: 1; }
}

@keyframes illt3 {
  0%, 12.5%, 25%, 68.75%, 81.25%, 100% { opacity: 0; }
  18.75%, 75% { opacity: 1; }
}

@keyframes illt4 {
  0%, 18.75%, 31.25%, 62.5%, 75%, 100% { opacity: 0; }
  25%, 68.75% { opacity: 1; }
}

@keyframes illt5 {
  0%, 25%, 37.5%, 56.25%, 68.75%, 100% { opacity: 0; }
  31.25%, 62.5% { opacity: 1; }
}

@keyframes illt6 {
  0%, 31.25%, 43.75%, 50%, 62.5%, 100% { opacity: 0; }
  37.5%, 56.25% { opacity: 1; }
}

@keyframes illt7 {
  0%, 37.5%, 56.25%, 100% { opacity: 0; }
  43.75%, 50% { opacity: 1; }
}

@media (prefers-reduced-motion: reduce) {
  .feature-sketch * {
    animation: none !important;
    transition: none !important;
  }
}
</style>
