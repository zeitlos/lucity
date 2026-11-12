<script setup lang="ts">
defineProps<{
  signInUrl: string;
  signUpUrl: string;
  githubUrl: string;
}>();

const groups = [
  {
    key: 'features',
    label: 'Features',
    links: [
      { to: '/features', label: 'Every feature' },
      { to: '/security', label: 'Security' },
      { to: '/no-lock-in', label: 'No lock-in' },
    ],
  },
  {
    key: 'solutions',
    label: 'Solutions',
    links: [
      { to: '/use-cases/startups', label: 'Startups and small teams' },
      { to: '/use-cases/agencies', label: 'Agencies' },
      { to: '/use-cases/platform-teams', label: 'Platform teams' },
      { to: '/agents', label: 'Agent hosting', soon: true },
      { to: '/comparisons', label: 'Switching from another platform' },
    ],
  },
  {
    key: 'resources',
    label: 'Resources',
    links: [
      { to: '/docs/quickstart', label: 'Docs' },
      { to: '/blog', label: 'Blog' },
      { to: '/docs/guides', label: 'Guides' },
      { to: 'https://github.com/zeitlos/lucity/releases', label: 'Changelog' },
    ],
  },
];

const expanded = ref<string | null>('solutions');

const row = 'flex min-h-15.5 items-center justify-between border-b border-stone-300/70 text-xl text-stone-900 focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-(--nav-accent) dark:border-stone-800 dark:text-stone-100';
</script>

<template>
  <div class="px-6 pt-1 pb-8 text-left">
    <nav aria-label="Main" class="flex flex-col">
      <div v-for="group in groups" :key="group.key" class="border-b border-stone-300/70 dark:border-stone-800">
        <button
          type="button"
          :aria-expanded="expanded === group.key"
          :aria-controls="`sheet-${group.key}`"
          class="flex min-h-15.5 w-full cursor-pointer items-center justify-between text-xl text-stone-900 focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-(--nav-accent) dark:text-stone-100"
          @click="expanded = expanded === group.key ? null : group.key"
        >
          {{ group.label }}
          <UIcon
            name="i-lucide-chevron-down"
            class="size-5 text-stone-500 transition-transform duration-200 motion-reduce:transition-none"
            :class="{ 'rotate-180': expanded === group.key }"
          />
        </button>
        <div
          :id="`sheet-${group.key}`"
          class="sheet-group"
          :class="{ 'is-open': expanded === group.key }"
        >
          <div class="min-h-0 overflow-hidden">
            <div class="pb-3.5">
              <NuxtLink
                v-for="link in group.links"
                :key="link.to"
                :to="link.to"
                class="flex min-h-11.5 items-center text-[17px] text-stone-800 focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-(--nav-accent) dark:text-stone-300"
              >
                <span>{{ link.label }}<em v-if="link.soon" class="ml-2 text-[19px] font-display font-normal italic text-[oklch(0.58_0.20_0)] dark:text-[oklch(0.76_0.13_0)]">soon</em></span>
              </NuxtLink>
            </div>
          </div>
        </div>
      </div>

      <NuxtLink to="/pricing" :class="row">
        Pricing
        <UIcon name="i-lucide-chevron-right" class="size-5 text-stone-500" />
      </NuxtLink>
      <a :href="githubUrl" target="_blank" rel="noopener" :class="row">
        GitHub
        <UIcon name="i-lucide-arrow-up-right" class="size-5 text-stone-500" />
      </a>
    </nav>

    <div class="mt-7 flex flex-col gap-3">
      <a
        :href="signInUrl"
        class="flex min-h-13.5 items-center justify-center rounded-lg border border-stone-300 bg-stone-50/50 text-lg font-semibold text-stone-900 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-(--nav-accent) dark:border-stone-700 dark:bg-white/5 dark:text-stone-100"
      >
        Sign in
      </a>
      <a
        :href="signUpUrl"
        class="flex min-h-13.5 items-center justify-center rounded-lg bg-primary text-lg font-semibold text-white transition-colors hover:bg-primary-600 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-(--nav-accent)"
      >
        Get started
      </a>
      <p class="mt-1 text-center text-base text-stone-700 dark:text-stone-400">
        14 days free, up to €5 of usage
      </p>
    </div>
  </div>
</template>

<style scoped>
.sheet-group {
  display: grid;
  grid-template-rows: 0fr;
  visibility: hidden;
  transition:
    grid-template-rows 0.25s cubic-bezier(0.2, 0.8, 0.2, 1),
    visibility 0s 0.25s;
}

.sheet-group.is-open {
  grid-template-rows: 1fr;
  visibility: visible;
  transition:
    grid-template-rows 0.25s cubic-bezier(0.2, 0.8, 0.2, 1),
    visibility 0s;
}

@media (prefers-reduced-motion: reduce) {
  .sheet-group,
  .sheet-group.is-open {
    transition: none;
  }
}
</style>
