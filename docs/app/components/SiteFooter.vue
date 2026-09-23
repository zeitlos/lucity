<script setup lang="ts">
const columns = [
  {
    title: 'Docs',
    links: [
      { label: 'Quickstart', to: '/quickstart' },
      { label: 'Framework guides', to: '/guides' },
      { label: 'Deployments', to: '/deployments' },
      { label: 'CLI', to: '/cli' },
    ],
  },
  {
    title: 'Use Cases',
    links: [
      { label: 'Agencies', to: '/use-cases/agencies' },
      { label: 'Startups', to: '/use-cases/startups' },
      { label: 'Platform Teams', to: '/use-cases/platform-teams' },
    ],
  },
  {
    title: 'Comparison',
    links: [
      { label: 'Lucity vs. Railway', to: '/comparisons/vs-railway' },
      { label: 'Lucity vs. Heroku', to: '/comparisons/vs-heroku' },
      { label: 'Lucity vs. Render', to: '/comparisons/vs-render' },
      { label: 'Lucity vs. Vercel', to: '/comparisons/vs-vercel' },
      { label: 'Lucity vs. Fly.io', to: '/comparisons/vs-fly-io' },
      { label: 'Lucity vs. Coolify', to: '/comparisons/vs-coolify' },
    ],
  },
  {
    title: 'Company',
    links: [
      { label: 'About zeitlos', href: 'https://zeitlos.software' },
      { label: 'Contact', href: 'mailto:christian@zeitlos.software' },
      { label: 'Blog', to: '/blog' },
    ],
  },
  {
    title: 'Legal',
    links: [
      { label: 'Privacy Policy', to: '/legal/privacy-policy' },
      { label: 'Terms of Use', to: '/legal/terms-of-use' },
      { label: 'Data Processing', to: '/legal/data-processing-agreement' },
    ],
  },
];
</script>

<template>
  <footer class="px-6 pb-16 pt-20">
    <div class="mx-auto grid max-w-wide grid-cols-2 gap-x-8 gap-y-12 sm:grid-cols-3 lg:grid-cols-5">
      <div v-for="col in columns" :key="col.title">
        <h3 class="text-lg font-bold text-neutral-900 dark:text-neutral-100">
          {{ col.title }}
        </h3>
        <ul class="mt-4 space-y-3">
          <li v-for="link in col.links" :key="link.label">
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
    </div>

    <div class="mx-auto mt-14 flex max-w-wide flex-col items-center gap-4 border-t border-default pt-8 sm:flex-row sm:justify-between">
      <p class="text-base text-neutral-600 dark:text-neutral-400">
        Built in Switzerland, hosted in the EU
      </p>
      <div class="flex items-center gap-5">
        <a
          href="https://github.com/zeitlos/lucity"
          target="_blank"
          rel="noopener"
          aria-label="Lucity on GitHub"
          class="inline-flex items-center text-neutral-600 transition-colors hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-neutral-100"
        >
          <UIcon name="i-simple-icons-github" class="size-5" />
        </a>
        <a
          href="https://www.linkedin.com/company/lucity"
          target="_blank"
          rel="noopener"
          aria-label="Lucity on LinkedIn"
          class="inline-flex items-center text-neutral-600 transition-colors hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-neutral-100"
        >
          <UIcon name="i-simple-icons-linkedin" class="size-5" />
        </a>
        <ThemeToggle />
      </div>
    </div>
  </footer>
</template>
