<script setup lang="ts">
import { releasesUrl } from './releases';
</script>

<template>
  <section class="mx-auto max-w-content px-6 pb-16 sm:pb-22" aria-labelledby="follow">
    <h2 id="follow" class="font-display text-3xl/tight text-stone-900 lg:text-4xl/tight dark:text-stone-100">Follow along.</h2>
    <div class="mt-9 grid gap-x-10 border-b border-stone-300 sm:grid-cols-2 dark:border-stone-800">
      <div class="flex flex-col border-t border-stone-300 py-5.5 dark:border-stone-800">
        <b class="block text-[19px] font-semibold text-stone-900 dark:text-stone-100">Release notes</b>
        <span class="mt-1 mb-3 block max-w-[560px] text-base text-stone-700 dark:text-stone-400">
          Every release comes with notes written for the people who use the platform. Watch the repository on GitHub to get them by email.
        </span>
        <a :href="releasesUrl" class="mt-auto self-start text-base font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">All releases</a>
      </div>
      <div class="flex flex-col border-t border-stone-300 py-5.5 dark:border-stone-800">
        <b class="block text-[19px] font-semibold text-stone-900 dark:text-stone-100">Write back</b>
        <span class="mt-1 mb-3 block max-w-[560px] text-base text-stone-700 dark:text-stone-400">
          Questions, corrections and disagreement are welcome, and a person answers.
        </span>
        <a href="mailto:christian@zeitlos.software" class="mt-auto self-start text-base font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">christian@zeitlos.software</a>
      </div>
    </div>
  </section>
</template>
