<script setup lang="ts">
definePageMeta({
  layout: 'landing',
  header: false,
  footer: false,
});

useSeo({
  title: 'Security',
  description: 'Every app Lucity builds from your code is checked for leaked keys on each deploy and for known vulnerabilities while it runs, and every workspace is walled off from the next. With answers for your security review.',
  type: 'website',
});

const appUrl = '/app';

const checks = [
  'On every plan, the trial included',
  'Never holds a release back',
  'Open source, so you can read how it works',
];

const promises = [
  { name: 'Leaked keys', text: 'Every deploy reads your repository and its history, and tells you which keys still work.' },
  { name: 'Vulnerabilities', text: 'Apps Lucity built are rechecked every two hours while they run, not only when you ship.' },
  { name: 'Isolation', text: 'Apps in one workspace cannot reach the apps or databases of another.' },
];
</script>

<template>
  <div class="security-page">
    <section class="mx-auto max-w-content px-6 pt-12 sm:pt-28" aria-labelledby="security-title">
      <h1 id="security-title" class="load-in font-display text-4xl/[1.15] text-neutral-800 sm:text-5xl/[1.1] md:text-6xl/[1.1] lg:text-7xl/[1.1] dark:text-neutral-100" style="--load-order: 0">
        Guard rails <span class="squiggle">come standard</span>.
      </h1>

      <p class="load-in mt-8 max-w-[820px] text-xl leading-relaxed text-stone-700 sm:text-2xl sm:leading-relaxed dark:text-stone-300" style="--load-order: 1">
        Every app Lucity builds from your code is checked for leaked keys on each deploy, and for known vulnerabilities for as long as it runs. Every workspace is walled off from the next. Nothing to install and nothing to configure, on every plan.
      </p>

      <div class="load-in mt-8 flex flex-col gap-3 sm:mt-11 sm:flex-row sm:flex-wrap sm:items-center sm:gap-4" style="--load-order: 2">
        <a
          :href="`${appUrl}/login`"
          class="flex items-center justify-center gap-3 rounded-lg border border-transparent bg-primary px-7 py-3.5 text-lg font-semibold text-white shadow-sm transition-colors hover:bg-primary-600 sm:py-4 sm:text-xl sm:whitespace-nowrap"
        >
          Deploy and open the Security tab
          <UIcon name="i-lucide-arrow-right" class="size-5.5 shrink-0" />
        </a>
        <a
          href="#review"
          class="flex items-center justify-center rounded-lg border border-stone-300 bg-white/50 px-7 py-3.5 text-lg font-semibold text-neutral-800 transition-colors hover:bg-white sm:py-4 sm:text-xl dark:border-stone-700 sm:whitespace-nowrap dark:bg-white/5 dark:text-neutral-100 dark:hover:bg-white/10"
        >
          Answers for your security review
        </a>
      </div>

      <ul class="load-in mt-5 flex flex-wrap gap-x-7 gap-y-1.5 text-base text-stone-600 sm:mt-6 dark:text-stone-400" style="--load-order: 3">
        <li v-for="check in checks" :key="check" class="flex items-center gap-2">
          <UIcon name="i-lucide-check" class="size-4.5 shrink-0 text-primary-700 dark:text-primary" />
          {{ check }}
        </li>
      </ul>

      <div class="load-in mt-12 sm:mt-14" style="--load-order: 4; --load-blur: 0px">
        <SecurityHeroStage />
      </div>

      <div class="mt-10 grid gap-x-14 sm:mt-14 min-[1060px]:grid-cols-3">
        <div v-for="promise in promises" :key="promise.name" class="border-t border-stone-300 py-5.5 min-[1060px]:pb-0 dark:border-stone-800">
          <b class="block text-lg font-semibold text-stone-900 dark:text-stone-100">{{ promise.name }}</b>
          <span class="mt-1 block text-base text-stone-700 dark:text-stone-400">{{ promise.text }}</span>
        </div>
      </div>
    </section>

    <SecurityKeysChapter class="mt-20 sm:mt-31" />

    <SecurityVulnerabilitiesChapter class="mt-20 sm:mt-31" />

    <SecurityWallsChapter class="mt-20 sm:mt-31" />

    <SecurityAgentsChapter class="mt-20 sm:mt-31" />

    <SecurityReviewTable class="mt-20 sm:mt-31" />

    <SecurityFaq class="mt-20 sm:mt-28" />

    <SecurityEnding class="mt-20 pb-16 sm:mt-28 sm:pb-24" />
  </div>
</template>

<style scoped>
.security-page {
  overflow-x: clip;
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
  background: url('/img/underline-grow.svg') center / 100% 100% no-repeat;
  clip-path: inset(0 100% 0 0);
  animation: highlighter 0.55s ease-out 0.35s forwards;
}

@keyframes highlighter {
  to {
    clip-path: inset(0 0 0 0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .squiggle::after {
    animation: none;
    clip-path: inset(0 0 0 0);
  }
}
</style>
