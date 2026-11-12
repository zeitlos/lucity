<script setup lang="ts">
definePageMeta({
  layout: 'landing',
  header: false,
  footer: false,
});

useSeo({
  title: 'Deploy anything. Own your stack.',
  description: 'Get your apps online and running in production, without the lock-in. Lucity is the European open source alternative to Vercel, Heroku, and Railway: the same effortless developer experience, but your whole stack stays yours to move anywhere, anytime.',
  type: 'website',
});

const appUrl = '/app';

const trialPoints = [
  '14 days free, up to €5 of usage',
  'No lock-in, leave anytime',
];
</script>

<template>
  <section class="mx-auto px-6 mt-48 max-sm:mt-12 flex flex-col gap-16 3xl:grid 3xl:grid-cols-12 3xl:items-end 3xl:gap-24">
    <div class="3xl:col-span-5 3xl:mb-24">
      <h1 class="load-in font-display text-4xl/[1.15] sm:text-5xl/[1.1] md:text-6xl/[1.1] lg:text-7xl/22 text-neutral-800 max-sm:text-balance dark:text-neutral-100" style="--load-order: 0">
        The Deployment Platform <br class="max-sm:hidden">
        that is a <span class="squiggle squiggle-joy">joy to use</span>, <br class="max-sm:hidden">
        <span class="squiggle squiggle-grow">built to grow</span>, and always <br class="max-sm:hidden">
        <span class="squiggle squiggle-keep">yours to keep</span>.
      </h1>

      <p class="load-in mt-8 max-w-3xl text-xl leading-relaxed text-stone-700 sm:text-2xl sm:leading-relaxed dark:text-stone-300" style="--load-order: 1">
        Connect a repo and Lucity runs it, with databases, domains, backups and scaling built in. A cloud platform made in Switzerland and hosted in the EU.
      </p>

      <div class="load-in mt-8 flex flex-col gap-3 sm:mt-11 sm:flex-row sm:items-center sm:gap-4" style="--load-order: 2">
        <a
          :href="`${appUrl}/login`"
          class="flex items-center justify-center gap-3 rounded-lg border border-transparent bg-primary px-7 py-3.5 text-lg font-semibold text-white shadow-sm transition-colors hover:bg-primary-600 sm:py-4 sm:text-xl"
        >
          Deploy your first app
          <UIcon name="i-lucide-arrow-right" class="size-5.5 shrink-0" />
        </a>
        <NuxtLink
          to="/docs/quickstart"
          class="flex items-center justify-center rounded-lg border border-stone-300 bg-white/50 px-7 py-3.5 text-lg font-semibold text-neutral-800 transition-colors hover:bg-white sm:py-4 sm:text-xl dark:border-stone-700 dark:bg-white/5 dark:text-neutral-100 dark:hover:bg-white/10"
        >
          Read the docs
        </NuxtLink>
      </div>

      <ul class="load-in mt-5 flex flex-wrap gap-x-7 gap-y-1.5 text-base text-stone-600 sm:mt-6 dark:text-stone-400" style="--load-order: 3">
        <li v-for="point in trialPoints" :key="point" class="flex items-center gap-2">
          <UIcon name="i-lucide-check" class="size-4.5 shrink-0 text-primary-700 dark:text-primary" />
          {{ point }}
        </li>
      </ul>
    </div>

    <HeroDemo class="load-in w-full 3xl:col-span-7" style="--load-order: 4; --load-blur: 0px" />
  </section>

  <LandingTechMarquee class="mt-20 sm:mt-26" />

  <LandingJoyChapter class="mt-24 sm:mt-38" />

  <LandingGrowChapter class="mt-24 sm:mt-38" />

  <LandingGuardRails class="mt-24 sm:mt-34" />

  <LandingKeepChapter class="mt-24 sm:mt-38" />

  <LandingPriceCompare class="mt-24 sm:mt-34" />

  <LandingRecentlyShipped />

  <LandingFaq class="mt-20 sm:mt-28" />
</template>

<style scoped>
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
  background-repeat: no-repeat;
  background-position: center;
  background-size: 100% 100%;
  clip-path: inset(0 100% 0 0);
  animation: highlighter 0.55s ease-out forwards;
}

.squiggle-joy::after {
  background-image: url('/img/underline-joy.svg');
  animation-delay: 0.75s;
}

.squiggle-grow::after {
  background-image: url('/img/underline-grow.svg');
  animation-delay: 1.3s;
}

.squiggle-keep::after {
  background-image: url('/img/underline-keep.svg');
  animation-delay: 1.85s;
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
