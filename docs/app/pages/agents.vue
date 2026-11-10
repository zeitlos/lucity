<script setup lang="ts">
definePageMeta({
  layout: 'landing',
  header: false,
  footer: false,
});

useSeo({
  title: 'Agent hosting',
  description: 'Host the agents you already run, right next to your app. Lucity runs the agent from your terminal around the clock: it wakes on the events you choose, reaches only what you allow and keeps a record of every call.',
  type: 'website',
});

const appUrl = '/app';

const checks = [
  'No SDK, nothing to rewrite',
  'Your models, on your own keys',
  'No lock-in, leave anytime',
];

const beats = ['joy', 'grow', 'keep'] as const;

const beatColor = {
  joy: 'oklch(0.55 0.15 300)',
  grow: 'oklch(0.58 0.17 35)',
  keep: 'oklch(0.58 0.20 0)',
};

const promises = [
  { beat: 'joy', name: 'Always on', text: 'It runs on Lucity, not on a laptop. Asleep until something wakes it, at work seconds later.' },
  { beat: 'grow', name: 'Under your control', text: 'It starts with nothing. You grant each model, tool and host, the keys stay with you, and every call is on record.' },
  { beat: 'keep', name: 'Built from parts', text: 'Triggers, connectors, model providers and channels. Connect each once, combine them in plain words.' },
];

const facts = [
  { name: 'Billed while awake', text: 'An agent is billed like a service, by the memory and CPU it uses while it works. Asleep, it costs next to nothing.' },
  { name: 'Models at cost', text: 'Model calls go on your own provider bill, at their price. We add no markup and resell no tokens.' },
  { name: 'In Europe', text: 'Agents run in the EU, operated from Switzerland. Pair one with a European model provider and its work stays in Europe.' },
  { name: 'Nothing of ours inside', text: 'The harness is open and the instructions are a text file. Export an agent and run it wherever a container runs.' },
];

const steps = [
  { name: 'Pick its harness', text: 'The one you run in your terminal today.' },
  { name: 'Paste its instructions', text: 'The instructions file you already have works as it is.' },
  { name: 'Tell it when to wake', text: 'A deploy fails, an alert fires, or every weekday at 06:00.' },
];
</script>

<template>
  <div class="agents-page">
    <section class="mx-auto max-w-content px-6 pt-12 sm:pt-22" aria-labelledby="agents-title">
      <h1 id="agents-title" class="load-in font-display text-4xl/[1.15] text-stone-900 max-sm:text-balance md:text-5xl/[1.1] lg:text-6xl/[1.1] xl:text-7xl/[1.1] dark:text-stone-100" style="--load-order: 0">
        Host the agents you <span class="squiggle">already run</span>, <br class="max-sm:hidden">
        right next to your app.
      </h1>

      <p class="load-in mt-8 max-w-[800px] text-xl leading-relaxed text-stone-700 sm:text-[23px] sm:leading-relaxed dark:text-stone-300" style="--load-order: 1">
        Lucity runs the agent from your terminal as it is, around the clock. It wakes on the events you choose, reaches only what you allow and keeps a record of every call. A cloud platform made in Switzerland and hosted in the EU.
      </p>

      <div class="load-in mt-8 flex flex-col gap-3 sm:mt-11 sm:flex-row sm:items-center sm:gap-4" style="--load-order: 2">
        <a
          :href="`${appUrl}/login`"
          class="flex items-center justify-center gap-3 rounded-lg border border-transparent bg-primary px-7 py-3.5 text-lg font-semibold text-white shadow-sm transition-colors hover:bg-primary-600 sm:py-4 sm:text-xl"
        >
          Host your first agent
          <UIcon name="i-lucide-arrow-right" class="size-5.5 shrink-0" />
        </a>
        <NuxtLink
          to="/docs/agents"
          class="flex items-center justify-center rounded-lg border border-stone-300 bg-white/50 px-7 py-3.5 text-lg font-semibold text-neutral-800 transition-colors hover:bg-white sm:py-4 sm:text-xl dark:border-stone-700 dark:bg-white/5 dark:text-neutral-100 dark:hover:bg-white/10"
        >
          Read the docs
        </NuxtLink>
      </div>

      <ul class="load-in mt-5 flex flex-wrap gap-x-7 gap-y-1.5 text-base text-stone-600 sm:mt-6 dark:text-stone-400" style="--load-order: 3">
        <li v-for="check in checks" :key="check" class="flex items-center gap-2">
          <UIcon name="i-lucide-check" class="size-4.5 shrink-0 text-primary-700 dark:text-primary" />
          {{ check }}
        </li>
      </ul>

      <AgentsHeroStage class="load-in mt-12 sm:mt-14" style="--load-order: 4; --load-blur: 0px" />

      <div class="mt-12 grid gap-x-14 sm:mt-18 lg:grid-cols-3">
        <div v-for="promise in promises" :key="promise.name" class="border-t border-stone-300 pt-5.5 max-lg:pb-5.5 dark:border-stone-800">
          <div class="beats mb-5.5 flex gap-1.5" aria-hidden="true">
            <svg v-for="name in beats" :key="name" width="34" height="10" viewBox="0 0 34 10" fill="none" :stroke="name === promise.beat ? beatColor[name] : 'var(--beat-off)'" stroke-width="2.5" stroke-linecap="round"><path d="M2 6c5-6 10 6 15 0s10 6 15 0" /></svg>
          </div>
          <b class="block text-lg font-semibold text-stone-900 dark:text-stone-100">{{ promise.name }}</b>
          <span class="mt-1 block text-base text-stone-700 dark:text-stone-400">{{ promise.text }}</span>
        </div>
      </div>
    </section>

    <AgentsWorksWith class="mt-20 sm:mt-24" />

    <AgentsAlwaysOn class="mt-24 sm:mt-30" />

    <AgentsControl class="mt-24 sm:mt-30" />

    <AgentsParts class="mt-24 sm:mt-30" />

    <AgentsShip class="mt-24 sm:mt-30" />

    <section class="facts mt-24 bg-stone-50 py-16 sm:mt-30 sm:pt-22 sm:pb-20 dark:bg-stone-900" aria-labelledby="facts">
      <div class="mx-auto max-w-content px-6">
        <h2 id="facts" class="font-display text-3xl/tight text-stone-900 lg:text-4xl/tight dark:text-stone-100">What it costs, and where it runs.</h2>
        <div class="mt-10 grid grid-cols-[repeat(auto-fit,minmax(min(250px,100%),1fr))] gap-x-10 border-b border-stone-300 dark:border-stone-700">
          <div v-for="fact in facts" :key="fact.name" class="border-t border-stone-300 py-5.5 dark:border-stone-700">
            <b class="block text-lg font-semibold text-stone-900 dark:text-stone-100">{{ fact.name }}</b>
            <span class="mt-1 block text-base text-stone-700 dark:text-stone-400">{{ fact.text }}</span>
          </div>
        </div>
      </div>
    </section>

    <AgentsFaq class="mt-20 sm:mt-28" />

    <section class="mx-auto mt-24 max-w-content px-6 pb-14 sm:mt-30" aria-labelledby="end">
      <h2 id="end" class="max-w-[900px] font-display text-4xl/11 text-balance text-stone-900 md:text-5xl/14 lg:text-6xl/17 dark:text-stone-100">Move one agent off your laptop tonight.</h2>
      <div class="mt-10 grid gap-x-12 gap-y-8 sm:mt-14 lg:grid-cols-3">
        <div v-for="(step, index) in steps" :key="step.name" class="grid grid-cols-[48px_minmax(0,1fr)] items-baseline gap-x-4">
          <span class="font-display text-4xl/none text-(--brand-mark)">{{ index + 1 }}</span>
          <div>
            <b class="block text-lg font-semibold text-stone-900 dark:text-stone-100">{{ step.name }}</b>
            <p class="mt-1.5 text-base text-stone-700 dark:text-stone-400">{{ step.text }}</p>
          </div>
        </div>
      </div>
      <div class="mt-10 flex flex-col gap-x-5 gap-y-4 sm:mt-14 sm:flex-row sm:items-center">
        <a
          :href="`${appUrl}/login`"
          class="flex items-center justify-center gap-3 rounded-lg border border-transparent bg-primary px-7 py-3.5 text-lg font-semibold text-white shadow-sm transition-colors hover:bg-primary-600 sm:py-4 sm:text-xl"
        >
          Host your first agent
          <UIcon name="i-lucide-arrow-right" class="size-5.5 shrink-0" />
        </a>
        <span class="text-lg text-stone-700 max-sm:text-center dark:text-stone-400">Then close the lid.</span>
      </div>
    </section>
  </div>
</template>

<style scoped>
.agents-page {
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
  background: url('/img/underline-joy.svg') center / 100% 100% no-repeat;
  clip-path: inset(0 100% 0 0);
  animation: highlighter 0.55s ease-out 0.55s forwards;
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

.beats {
  --beat-off: oklch(0.82 0.03 75);
}

.dark .beats {
  --beat-off: oklch(0.40 0.03 55);
}

.facts {
  box-shadow: inset 0 1px 0 oklch(0.91 0.015 80), inset 0 -1px 0 oklch(0.91 0.015 80);
}

.dark .facts {
  box-shadow: inset 0 1px 0 oklch(0.30 0.03 55), inset 0 -1px 0 oklch(0.30 0.03 55);
}
</style>
