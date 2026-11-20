<script setup lang="ts">
const appUrl = '/app';

const next = [
  { name: 'Data processing agreement', text: 'Sub-processors, measures and your rights.', to: '/legal/data-processing-agreement' },
  { name: 'How leaving works', text: 'Take a project with you in a standard format.', to: '/no-lock-in' },
  { name: 'The source code', text: 'Read how scanning and isolation are built.', to: 'https://github.com/zeitlos/lucity' },
];
</script>

<template>
  <section class="mx-auto max-w-content px-6" aria-labelledby="security-end">
    <div class="flex flex-wrap items-center gap-x-18 gap-y-10 border-y border-stone-300 py-11 dark:border-stone-800">
      <div class="min-w-0 flex-[1_1_520px]">
        <h2 id="security-end" class="font-display text-4xl/11 text-balance text-stone-900 md:text-5xl/14 lg:text-6xl/17 dark:text-stone-100">Hand this page to your security team.</h2>
        <p class="mt-5 max-w-[640px] text-xl leading-relaxed text-pretty text-stone-700 dark:text-stone-300">
          Send us your questionnaire and we will fill it in. Or deploy something small and open its Security tab, which is the quickest way to see all of this for yourself.
        </p>
        <div class="mt-8 flex flex-wrap items-center gap-x-6 gap-y-4">
          <a
            href="mailto:christian@zeitlos.software?subject=Security%20questionnaire"
            class="flex items-center justify-center rounded-lg border border-transparent bg-primary px-7 py-3.5 text-lg font-semibold text-white shadow-sm transition-colors hover:bg-primary-600 max-sm:w-full"
          >
            Send us your questionnaire
          </a>
          <a :href="`${appUrl}/login`" class="text-lg font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">Or deploy and open the Security tab</a>
        </div>
      </div>

      <nav aria-label="Read next" class="min-w-[min(280px,100%)] flex-[0_1_380px] border-b border-stone-300 dark:border-stone-800">
        <NuxtLink
          v-for="item in next"
          :key="item.name"
          :to="item.to"
          class="group block border-t border-stone-300 py-5.5 dark:border-stone-800"
        >
          <b class="block text-lg font-semibold text-stone-900 group-hover:underline group-hover:decoration-1 group-hover:underline-offset-4 dark:text-stone-100">{{ item.name }}</b>
          <span class="mt-1 block text-base text-stone-700 dark:text-stone-400">{{ item.text }}</span>
        </NuxtLink>
      </nav>
    </div>
  </section>
</template>
