<script setup lang="ts">
import { comparisonFits, comparisonMatrix, comparisonPlatforms, comparisonSummaries, formatComparisonDate, comparisonCheckedOn } from '~/utils/comparisons';

definePageMeta({
  layout: 'landing',
  header: false,
  footer: false,
});

useSeo({
  title: 'How Lucity compares',
  description: 'Lucity next to Railway, Heroku, Render, Fly.io, Vercel, Deploio and Coolify, checked against each platform\'s own docs. Every comparison says where the other platform is ahead.',
  type: 'website',
});

const repository = 'https://github.com/zeitlos/lucity';

const rules = [
  { title: 'Their docs, linked and dated', text: "Every claim about another platform is checked against that platform's own docs and pricing, with a link next to it and the date we checked at the top of the page." },
  { title: 'Their features count too', text: "Each comparison covers Lucity's feature list and the other platform's headline features, so it isn't only about what we happen to do." },
  { title: 'Their strengths first', text: 'Each page says where the other platform is ahead before it says where Lucity is.' },
  { title: 'Our own screenshots', text: 'The same small app on both platforms, captured from our own accounts at the same window size, and not edited.' },
  { title: 'The same three setups', text: 'Every page prices the same side project, small production app and growing team, with the working shown.' },
  { title: 'Corrections within a week', text: 'Edit any page on GitHub or open an issue. We fix factual errors within a week, including ones the other teams send us.' },
];

const others = comparisonPlatforms.filter(platform => platform.key !== 'lucity');

function platformOf(key: string) {
  return comparisonPlatforms.find(platform => platform.key === key)!;
}
</script>

<template>
  <div class="pb-28 md:pb-36">
    <section class="mx-auto max-w-content px-6 pt-20 sm:pt-32" aria-labelledby="title">
      <h1 id="title" class="font-display text-4xl/[1.15] text-neutral-800 sm:text-5xl/[1.1] md:text-6xl/[1.1] lg:text-7xl/22 dark:text-neutral-100">
        How Lucity compares
      </h1>
      <p class="mt-7 max-w-4xl text-xl leading-relaxed text-neutral-700 sm:text-2xl sm:leading-relaxed dark:text-neutral-300">
        These pages explain where Lucity differs from other platforms, to help you choose one. Every platform here does something better than we do, and each comparison says what.
      </p>
      <p class="mt-4 max-w-4xl text-lg/relaxed text-neutral-600 dark:text-neutral-400">
        We build Lucity, so we aren't neutral. We try to be as objective as we can anyway, and none of these is a “versus” page. If you spot something wrong or out of date,
        <a :href="`${repository}/tree/main/docs/app/utils/comparisons`" target="_blank" rel="noopener" class="font-semibold text-neutral-900 underline decoration-1 underline-offset-4 dark:text-neutral-100">edit it on GitHub</a>
        or
        <a :href="`${repository}/issues/new`" target="_blank" rel="noopener" class="font-semibold text-neutral-900 underline decoration-1 underline-offset-4 dark:text-neutral-100">open an issue</a>.
      </p>
    </section>

    <section class="mx-auto mt-24 max-w-content px-6" aria-labelledby="rules">
      <h2 id="rules" class="font-display text-3xl leading-tight text-neutral-800 md:text-4xl dark:text-neutral-100">
        How we compare
      </h2>
      <div class="mt-7 grid gap-x-12 sm:grid-cols-2 lg:grid-cols-3">
        <div v-for="rule in rules" :key="rule.title" class="border-t border-default pt-5 pb-6">
          <h3 class="text-lg font-semibold text-neutral-900 dark:text-neutral-100">{{ rule.title }}</h3>
          <p class="mt-2 text-base/relaxed text-neutral-600 dark:text-neutral-400">{{ rule.text }}</p>
        </div>
      </div>
    </section>

    <section class="mx-auto mt-24 max-w-[96rem] px-6" aria-labelledby="glance">
      <div class="mx-auto max-w-content">
        <h2 id="glance" class="font-display text-4xl leading-tight text-neutral-800 lg:text-5xl lg:leading-tight dark:text-neutral-100">
          At a glance
        </h2>
        <p class="mt-5 max-w-3xl text-lg text-neutral-600 dark:text-neutral-400">
          The questions people ask most when they choose, for every platform we compare. Each column links to the full comparison, with the sources.
        </p>
        <div class="mt-6 mb-8 flex flex-wrap items-center gap-x-6 gap-y-2 text-[15px]">
          <span class="inline-flex items-center gap-1.5"><CompareMark mark="yes" /><span class="font-semibold">Yes</span></span>
          <span class="inline-flex items-center gap-1.5"><CompareMark mark="partly" /><span class="font-semibold">Partly, or with a catch</span></span>
          <span class="inline-flex items-center gap-1.5"><CompareMark mark="no" /><span class="font-semibold">No</span></span>
        </div>
      </div>
      <CompareMatrix :groups="comparisonMatrix" :platforms="comparisonPlatforms" />
    </section>

    <section class="mx-auto mt-28 max-w-content px-6" aria-labelledby="fit">
      <h2 id="fit" class="font-display text-4xl leading-tight text-neutral-800 lg:text-5xl lg:leading-tight dark:text-neutral-100">
        Which one fits
      </h2>
      <p class="mt-5 max-w-3xl text-lg text-neutral-600 dark:text-neutral-400">
        Start from what matters most to you. More than one answer is often fine.
      </p>
      <ul class="mt-8 border-b border-default">
        <li v-for="fit in comparisonFits" :key="fit.need" class="grid gap-x-10 gap-y-2 border-t border-default py-4.5 md:grid-cols-[minmax(0,1fr)_minmax(0,1.3fr)]">
          <span class="text-lg/snug font-semibold text-neutral-900 dark:text-neutral-100">{{ fit.need }}</span>
          <span class="text-[17px]/normal text-neutral-700 dark:text-neutral-300">
            <span class="mr-1.5 inline-flex items-center gap-2 align-[-3px] font-semibold text-neutral-900 dark:text-neutral-100">
              <ComparePlatformIcon v-for="key in fit.platforms" :key="key" :name="platformOf(key).name" :icon="platformOf(key).icon" :lucity="key === 'lucity'" />
              {{ fit.platforms.map(key => platformOf(key).name).join(' and ') }}.
            </span>
            {{ fit.why }}
          </span>
        </li>
      </ul>
    </section>

    <section class="mx-auto mt-28 max-w-content px-6" aria-labelledby="all">
      <h2 id="all" class="font-display text-4xl leading-tight text-neutral-800 lg:text-5xl lg:leading-tight dark:text-neutral-100">
        All comparisons
      </h2>
      <ul class="mt-8 border-b border-default">
        <li v-for="platform in others" :key="platform.key" class="grid items-baseline gap-x-8 gap-y-1.5 border-t border-default py-5 md:grid-cols-[14rem_minmax(0,1fr)_auto]">
          <NuxtLink :to="platform.to!" class="inline-flex items-center gap-2.5 text-xl font-semibold text-neutral-900 underline decoration-stone-400 decoration-1 underline-offset-4 hover:decoration-current dark:text-neutral-100">
            <ComparePlatformIcon :name="platform.name" :icon="platform.icon" />Lucity and {{ platform.name }}
          </NuxtLink>
          <span class="text-[17px]/normal text-neutral-700 dark:text-neutral-300">{{ comparisonSummaries[platform.key] }}</span>
          <span class="text-sm whitespace-nowrap text-neutral-600 dark:text-neutral-400">Checked {{ formatComparisonDate(comparisonCheckedOn) }}</span>
        </li>
      </ul>
      <p class="mt-6 text-lg text-neutral-700 dark:text-neutral-300">
        Missing a platform you're weighing up?
        <a :href="`${repository}/issues/new`" target="_blank" rel="noopener" class="font-semibold text-neutral-900 underline decoration-1 underline-offset-4 dark:text-neutral-100">Tell us which one</a>.
      </p>
    </section>
  </div>
</template>
