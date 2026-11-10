<script setup lang="ts">
const appUrl = '/app';

const terms = [
  { name: '14 days or CHF 5 of usage', text: 'Whichever runs out first. A badge in the dashboard counts down both.' },
  { name: 'No credit card', text: 'You add a payment method when you choose Hobby or Pro.' },
  { name: 'Nothing is lost at the end', text: 'If the trial runs out first, the workspace pauses until you pick a plan. Your data stays put.' },
];
</script>

<template>
  <section class="mx-auto max-w-content px-6" aria-labelledby="end">
    <h2 id="end" class="max-w-[900px] font-display text-4xl/11 text-balance text-stone-900 md:text-5xl/14 lg:text-6xl/17 dark:text-stone-100">Start on the trial, pick a plan later.</h2>

    <div class="mt-10 grid grid-cols-[repeat(auto-fit,minmax(min(250px,100%),1fr))] gap-x-10 border-b border-stone-300 sm:mt-12 dark:border-stone-800">
      <div v-for="term in terms" :key="term.name" class="border-t border-stone-300 py-5.5 dark:border-stone-800">
        <b class="block text-lg font-semibold text-stone-900 dark:text-stone-100">{{ term.name }}</b>
        <span class="mt-1 block text-base text-stone-700 dark:text-stone-400">{{ term.text }}</span>
      </div>
    </div>

    <div class="mt-10 flex flex-col gap-x-6 gap-y-5 sm:mt-12 sm:flex-row sm:flex-wrap sm:items-center">
      <a
        :href="`${appUrl}/login`"
        class="flex items-center justify-center gap-3 rounded-lg border border-transparent bg-primary px-7 py-3.5 text-lg font-semibold text-white shadow-sm transition-colors hover:bg-primary-600 sm:py-4 sm:text-xl"
      >
        Start your trial
        <UIcon name="i-lucide-arrow-right" class="size-5.5 shrink-0" />
      </a>
      <span class="text-[17px] text-stone-700 dark:text-stone-400">
        Bigger than Pro, or need a contract?
        <a href="mailto:christian@zeitlos.software?subject=Larger%20plan" class="font-medium text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">Write to Christian</a>
      </span>
    </div>
  </section>
</template>
