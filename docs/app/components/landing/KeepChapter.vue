<script setup lang="ts">
interface Entry {
  name: string;
  note: string;
  folder?: 'open' | 'closed';
  nested?: boolean;
}

const entries: Entry[] = [
  { name: 'build.sh', note: 'Rebuilds your images' },
  { name: 'chart', note: 'How every service runs', folder: 'closed' },
  { name: 'README.md', note: 'How to install it' },
  { name: 'values', note: 'One file per environment', folder: 'open' },
  { name: 'development.yaml', note: '', nested: true },
  { name: 'production.yaml', note: 'Services, variables, databases, domains', nested: true },
];

const promises = [
  { name: 'Leave anytime', text: 'One click downloads your services, variables, databases and domains. Your data is yours to copy over, and the guide shows how.', link: { to: '/docs/eject', label: 'How leaving works' } },
  { name: 'Hosted in Europe', text: 'Lucity is a Swiss company. Apps and databases run in Germany, files and backups in France.', link: { to: '/features#ownership', label: 'Where your data lives' } },
  { name: 'Open source', text: 'Read every line of the platform, or run the whole thing on your own infrastructure for free.', link: { to: '/docs/self-hosting', label: 'Self-hosting' } },
];

const folderPath = 'M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.7-.9l-.8-1.2A2 2 0 0 0 7.9 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z';
</script>

<template>
  <section class="mx-auto max-w-content px-6" aria-labelledby="keep">
    <LandingChapterHeading id="keep" beat="keep" promise="Yours to keep" rest="even if you leave.">
      Every project has an Eject button. You'll probably never use it. We'd rather keep you by being good than by being hard to leave.
    </LandingChapterHeading>

    <LandingStage tint="keep" class="mt-14 flex flex-wrap items-end gap-x-10 gap-y-7 px-4 pt-11 sm:px-10">
      <div class="min-w-0 flex-[1.5_1_460px] self-center pb-2 xl:pb-11">
        <img src="/img/landing/sketch-eject.svg" alt="Your project on Lucity, and the same project as a copy running anywhere else, carried over as one zip" width="760" height="300" class="mx-auto block h-auto w-full max-w-[820px] dark:hidden">
        <img src="/img/landing/sketch-eject-night.svg" alt="" width="760" height="300" class="mx-auto hidden h-auto w-full max-w-[820px] dark:block">
        <p class="mt-5 flex items-start gap-2.5 text-lg font-medium text-[oklch(0.32_0.05_0)] sm:items-center dark:text-[oklch(0.90_0.03_0)] sm:justify-center">
          <UIcon name="i-lucide-check" class="mt-1.5 size-4 shrink-0 text-[oklch(0.50_0.18_0)] sm:mt-0 dark:text-[oklch(0.78_0.14_0)]" />
          No account, no license key, no call home.
        </p>
      </div>

      <div class="finder mx-auto min-w-0 max-w-[560px] flex-[1_1_400px]" role="group" aria-label="What is in the zip">
        <div class="finder-bar" aria-hidden="true">
          <i /><i /><i />
          <b><svg width="18" height="18" viewBox="0 0 24 24" fill="oklch(0.90 0.06 0)" stroke="oklch(0.58 0.20 0)" stroke-width="1.8" stroke-linejoin="round"><path :d="folderPath" /></svg>vouch-ejected</b>
        </div>
        <div class="finder-head" aria-hidden="true"><span>Name</span><span>What it's for</span></div>
        <div>
          <div v-for="entry in entries" :key="entry.name" class="finder-row">
            <span :class="{ 'pl-5.5': entry.nested }">
              <svg v-if="entry.folder" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="oklch(0.50 0.03 55)" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path :d="entry.folder === 'open' ? 'm6 9 6 6 6-6' : 'm9 6 6 6-6 6'" />
              </svg>
              <span v-else class="w-3 shrink-0" />
              <svg v-if="entry.folder" width="18" height="18" viewBox="0 0 24 24" fill="oklch(0.90 0.06 0)" stroke="oklch(0.58 0.20 0)" stroke-width="1.8" stroke-linejoin="round" aria-hidden="true"><path :d="folderPath" /></svg>
              <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="oklch(0.99 0.004 80)" stroke="oklch(0.45 0.04 55)" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z" /><path d="M14 2v4a2 2 0 0 0 2 2h4" /></svg>
              {{ entry.name }}
            </span>
            <span>{{ entry.note }}</span>
          </div>
          <div v-for="filler in 3" :key="filler" class="finder-row filler" aria-hidden="true"><span /><span /></div>
        </div>
      </div>
    </LandingStage>

    <div class="mt-10 grid grid-cols-[repeat(auto-fit,minmax(min(300px,100%),1fr))] gap-x-12 border-b border-stone-300 dark:border-stone-800">
      <div v-for="promise in promises" :key="promise.name" class="border-t border-stone-300 py-5.5 dark:border-stone-800">
        <b class="block text-lg font-semibold text-stone-900 dark:text-stone-100">{{ promise.name }}</b>
        <span class="mt-1 block text-base text-stone-700 dark:text-stone-400">{{ promise.text }}</span>
        <NuxtLink :to="promise.link.to" class="mt-3 inline-block text-base font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">{{ promise.link.label }}</NuxtLink>
      </div>
    </div>
  </section>
</template>

<style scoped>
.finder {
  border-radius: 12px 12px 0 0;
  overflow: hidden;
  background: oklch(0.988 0.005 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .14), 0 28px 56px -22px oklch(0.25 0.04 55 / .42);
  font-size: 15px;
  color: oklch(0.25 0.03 55);
}

.finder-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 44px;
  padding: 0 16px;
  background: oklch(0.955 0.012 60);
  box-shadow: 0 1px 0 oklch(0.25 0.04 55 / .10);
}

.finder-bar i {
  flex: none;
  width: 11px;
  height: 11px;
  border-radius: 50%;
  background: oklch(0.80 0.03 40);
}

.finder-bar b {
  margin-left: 10px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
}

.finder-head,
.finder-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.3fr);
  gap: 16px;
  align-items: center;
  padding: 0 18px;
}

.finder-head {
  height: 32px;
  font-size: 13px;
  color: oklch(0.47 0.04 55);
  box-shadow: 0 1px 0 oklch(0.25 0.04 55 / .10);
}

.finder-row {
  height: 40px;
}

.finder-row:nth-child(even) {
  background: oklch(0.25 0.04 55 / .04);
}

.finder-row > span:first-child {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  white-space: nowrap;
}

.finder-row > span:last-child {
  font-size: 14px;
  color: oklch(0.45 0.04 55);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.finder-row svg {
  flex: none;
}

@media (max-width: 520px) {
  .finder-head,
  .finder-row {
    grid-template-columns: minmax(0, 1fr);
  }

  .finder-head > span:last-child,
  .finder-row > span:last-child,
  .filler + .filler {
    display: none;
  }
}
</style>
