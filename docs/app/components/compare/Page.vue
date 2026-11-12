<script setup lang="ts">
import { comparisonGroups, formatComparisonDate } from '~/utils/comparisons';
import type { Comparison } from '~/utils/comparisons/types';

const props = defineProps<{
  comparison: Comparison;
}>();

const repository = 'https://github.com/zeitlos/lucity';

const groups = computed(() => comparisonGroups(props.comparison));
const checked = computed(() => formatComparisonDate(props.comparison.checked));
const editUrl = computed(() => `${repository}/edit/main/docs/app/utils/comparisons/${props.comparison.key}.ts`);
const issueUrl = `${repository}/issues/new`;

useSeo({
  title: `Lucity and ${props.comparison.name}`,
  description: props.comparison.description,
  type: 'article',
});
</script>

<template>
  <article class="pb-28 md:pb-36">
    <header class="mx-auto max-w-content px-6 pt-20 sm:pt-28">
      <nav aria-label="Breadcrumb" class="text-[15px] text-neutral-600 dark:text-neutral-400">
        <NuxtLink to="/comparisons" class="underline decoration-stone-400 decoration-1 underline-offset-4 hover:decoration-current">Comparisons</NuxtLink>
        <span aria-hidden="true" class="mx-2">/</span>
        <span>{{ comparison.name }}</span>
      </nav>

      <div class="mt-8 flex items-center gap-3.5 text-[26px]" aria-hidden="true">
        <span class="flex size-14 items-center justify-center rounded-2xl border border-default bg-white/60 dark:bg-white/5"><ComparePlatformIcon name="Lucity" icon="" lucity /></span>
        <span class="font-display text-2xl text-neutral-500 italic dark:text-neutral-400">and</span>
        <span class="flex size-14 items-center justify-center rounded-2xl border border-default bg-white/60 text-[30px] text-neutral-900 dark:bg-white/5 dark:text-neutral-100"><ComparePlatformIcon :name="comparison.name" :icon="comparison.icon" /></span>
      </div>

      <h1 class="mt-7 font-display text-4xl/[1.15] text-neutral-800 sm:text-5xl/[1.1] md:text-6xl/[1.1] lg:text-7xl/22 dark:text-neutral-100">
        Lucity and {{ comparison.name }}
      </h1>
      <p class="mt-7 max-w-4xl text-xl leading-relaxed text-neutral-700 sm:text-2xl sm:leading-relaxed dark:text-neutral-300">
        {{ comparison.intro }}
      </p>
      <p class="mt-4 max-w-4xl text-lg/relaxed text-neutral-600 dark:text-neutral-400">
        We build Lucity, so we aren't neutral. We try to be as objective as we can anyway: this is not a “versus” page, and {{ comparison.respect }}
        If you spot something wrong or out of date,
        <a :href="editUrl" target="_blank" rel="noopener" class="font-semibold text-neutral-900 underline decoration-1 underline-offset-4 dark:text-neutral-100">edit this page on GitHub</a>
        or
        <a :href="issueUrl" target="_blank" rel="noopener" class="font-semibold text-neutral-900 underline decoration-1 underline-offset-4 dark:text-neutral-100">open an issue</a>.
        We correct facts within a week, including ones the {{ comparison.name }} team sends us.
      </p>

      <dl class="mt-11 grid grid-cols-1 border-y border-default sm:grid-cols-2 lg:grid-cols-4">
        <div class="py-4 pr-5">
          <dt class="text-sm text-neutral-600 dark:text-neutral-400">Last checked</dt>
          <dd class="mt-1 font-semibold text-neutral-900 dark:text-neutral-100">{{ checked }}</dd>
        </div>
        <div class="py-4 pr-5">
          <dt class="text-sm text-neutral-600 dark:text-neutral-400">Checked against</dt>
          <dd class="mt-1 font-semibold text-neutral-900 dark:text-neutral-100">{{ comparison.sources }}</dd>
        </div>
        <div class="py-4 pr-5">
          <dt class="text-sm text-neutral-600 dark:text-neutral-400">Screenshots</dt>
          <dd class="mt-1 font-semibold text-neutral-900 dark:text-neutral-100">Our own accounts on both</dd>
        </div>
        <div class="py-4">
          <dt class="text-sm text-neutral-600 dark:text-neutral-400">Changes</dt>
          <dd class="mt-1 font-semibold text-neutral-900 dark:text-neutral-100">
            <a href="#history" class="underline decoration-1 underline-offset-4">Page history</a>
          </dd>
        </div>
      </dl>
    </header>

    <section class="mx-auto mt-24 max-w-content px-6 lg:mt-28" aria-labelledby="short">
      <h2 id="short" class="font-display text-4xl leading-tight text-neutral-800 lg:text-5xl lg:leading-tight dark:text-neutral-100">
        In short
      </h2>
      <p class="mt-5 max-w-4xl text-xl/relaxed text-neutral-700 dark:text-neutral-300">
        {{ comparison.short }}
      </p>
      <div class="mt-11 grid gap-x-14 gap-y-10 lg:grid-cols-2">
        <div>
          <h3 class="flex items-center gap-2.5 text-lg font-semibold text-neutral-900 dark:text-neutral-100">
            <ComparePlatformIcon :name="comparison.name" :icon="comparison.icon" />{{ comparison.name }} is the better fit if
          </h3>
          <ul class="mt-3.5 border-b border-default">
            <li v-for="item in comparison.fitOther" :key="item" class="border-t border-default py-3 text-[17px]/normal text-neutral-700 dark:text-neutral-300">
              {{ item }}
            </li>
          </ul>
        </div>
        <div>
          <h3 class="flex items-center gap-2.5 text-lg font-semibold text-neutral-900 dark:text-neutral-100">
            <ComparePlatformIcon name="Lucity" icon="" lucity />Lucity is the better fit if
          </h3>
          <ul class="mt-3.5 border-b border-default">
            <li v-for="item in comparison.fitLucity" :key="item" class="border-t border-default py-3 text-[17px]/normal text-neutral-700 dark:text-neutral-300">
              {{ item }}
            </li>
          </ul>
        </div>
      </div>
    </section>

    <section class="mx-auto mt-28 max-w-content px-6" aria-labelledby="screens">
      <h2 id="screens" class="font-display text-4xl leading-tight text-neutral-800 lg:text-5xl lg:leading-tight dark:text-neutral-100">
        See them side by side
      </h2>
      <p class="mt-5 max-w-3xl text-lg text-neutral-600 dark:text-neutral-400">
        {{ comparison.shotsIntro }}
      </p>
      <CompareShots :platform="comparison.key" :name="comparison.name" :icon="comparison.icon" :shots="comparison.shots" />
    </section>

    <section class="mx-auto mt-28 max-w-content px-6" aria-labelledby="features">
      <h2 id="features" class="font-display text-4xl leading-tight text-neutral-800 lg:text-5xl lg:leading-tight dark:text-neutral-100">
        Feature by feature
      </h2>
      <p class="mt-5 max-w-3xl text-lg text-neutral-600 dark:text-neutral-400">
        {{ comparison.tableIntro }}
      </p>
      <CompareTable :groups="groups" :name="comparison.name" :icon="comparison.icon" />
      <p class="mt-6 max-w-3xl text-[15px] text-neutral-600 dark:text-neutral-400">
        {{ comparison.tableFoot }}
      </p>
    </section>

    <section class="mx-auto mt-28 max-w-content px-6" aria-labelledby="ahead">
      <div class="grid gap-x-16 gap-y-14 lg:grid-cols-2">
        <div v-for="(column, index) in [{ title: `Where ${comparison.name} is ahead`, points: comparison.aheadOther }, { title: 'Where Lucity is ahead', points: comparison.aheadLucity }]" :key="column.title">
          <h2 :id="index === 0 ? 'ahead' : undefined" class="font-display text-3xl leading-tight text-neutral-800 md:text-4xl dark:text-neutral-100">
            {{ column.title }}
          </h2>
          <ul class="mt-6 border-b border-default">
            <li v-for="point in column.points" :key="point.title" class="border-t border-default py-3.5 text-[17px]/normal text-neutral-700 dark:text-neutral-300">
              <strong class="font-semibold text-neutral-900 dark:text-neutral-100">{{ point.title }}</strong> {{ point.text }}
            </li>
          </ul>
        </div>
      </div>
      <div class="mt-16 max-w-4xl">
        <h2 class="font-display text-3xl leading-tight text-neutral-800 md:text-4xl dark:text-neutral-100">
          Where they simply differ
        </h2>
        <ul class="mt-6 border-b border-default">
          <li v-for="point in comparison.differ" :key="point.title" class="border-t border-default py-3.5 text-[17px]/normal text-neutral-700 dark:text-neutral-300">
            <strong class="font-semibold text-neutral-900 dark:text-neutral-100">{{ point.title }}</strong> {{ point.text }}
          </li>
        </ul>
      </div>
    </section>

    <section class="mx-auto mt-28 max-w-content px-6" aria-labelledby="cost">
      <h2 id="cost" class="font-display text-4xl leading-tight text-neutral-800 lg:text-5xl lg:leading-tight dark:text-neutral-100">
        What the same app costs
      </h2>
      <p class="mt-5 max-w-4xl text-lg/relaxed text-neutral-600 dark:text-neutral-400">
        {{ comparison.priceIntro }}
      </p>
      <div class="mt-9 overflow-x-auto">
        <table class="w-full min-w-2xl border-collapse text-left">
          <colgroup>
            <col class="w-[46%]">
            <col class="w-[27%]">
            <col class="w-[27%]">
          </colgroup>
          <thead>
            <tr class="text-[15px] text-neutral-600 dark:text-neutral-400">
              <th scope="col" class="py-3 pr-4 font-semibold">Setup, per month</th>
              <th scope="col" class="py-3 pr-4 text-right font-semibold">
                <span class="inline-flex items-center gap-2"><ComparePlatformIcon :name="comparison.name" :icon="comparison.icon" />{{ comparison.name }}</span>
              </th>
              <th scope="col" class="py-3 text-right font-semibold">
                <span class="inline-flex items-center gap-2"><ComparePlatformIcon name="Lucity" icon="" lucity />Lucity</span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="price in comparison.prices" :key="price.setup" class="border-t border-default align-top">
              <th scope="row" class="py-4 pr-6 font-normal">
                <span class="block text-[17px] font-semibold text-neutral-900 dark:text-neutral-100">{{ price.setup }}</span>
                <span class="mt-0.5 block text-[15px] text-neutral-600 dark:text-neutral-400">{{ price.detail }}</span>
              </th>
              <td class="py-4 pr-4 text-right">
                <span class="block font-display text-3xl text-neutral-800 tabular-nums dark:text-neutral-100">{{ price.other }}</span>
                <span class="mt-1 block text-sm text-neutral-600 dark:text-neutral-400">{{ price.otherNote }}</span>
              </td>
              <td class="py-4 text-right">
                <span class="block font-display text-3xl text-neutral-800 tabular-nums dark:text-neutral-100">{{ price.lucity }}</span>
                <span class="mt-1 block text-sm text-neutral-600 dark:text-neutral-400">{{ price.lucityNote }}</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <details class="mt-6 border-y border-default py-1">
        <summary class="flex min-h-11 cursor-pointer items-center font-semibold text-neutral-900 dark:text-neutral-100">
          How we worked these out
        </summary>
        <p class="max-w-4xl pb-4 text-base/relaxed text-neutral-700 dark:text-neutral-300">
          {{ comparison.priceMethod }}
          For your own numbers, use the <NuxtLink to="/pricing" class="font-semibold underline decoration-1 underline-offset-4">calculator on the pricing page</NuxtLink>.
        </p>
      </details>
    </section>

    <section class="mx-auto mt-28 max-w-content px-6" aria-labelledby="move">
      <h2 id="move" class="font-display text-4xl leading-tight text-neutral-800 lg:text-5xl lg:leading-tight dark:text-neutral-100">
        Moving from {{ comparison.name }}
      </h2>
      <p class="mt-5 max-w-4xl text-lg/relaxed text-neutral-600 dark:text-neutral-400">
        {{ comparison.moveIntro }}
      </p>
      <div class="mt-10 grid items-start gap-x-16 gap-y-10 lg:grid-cols-2">
        <table class="w-full border-collapse text-left text-base">
          <thead>
            <tr class="text-[15px] text-neutral-600 dark:text-neutral-400">
              <th scope="col" class="py-3 pr-4 font-semibold">On {{ comparison.name }}</th>
              <th scope="col" class="py-3 font-semibold">On Lucity</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="[theirs, ours] in comparison.map" :key="theirs" class="border-t border-default align-top">
              <td class="py-3 pr-4 font-semibold text-neutral-900 dark:text-neutral-100">{{ theirs }}</td>
              <td class="py-3 text-neutral-700 dark:text-neutral-300">{{ ours }}</td>
            </tr>
          </tbody>
        </table>
        <ol class="border-b border-default">
          <li v-for="(step, index) in comparison.steps" :key="step" class="grid grid-cols-[2rem_1fr] gap-1 border-t border-default py-3.5 text-[17px]/normal text-neutral-700 dark:text-neutral-300">
            <span class="font-mono text-[15px] text-neutral-500">{{ index + 1 }}</span>
            <span>{{ step }}</span>
          </li>
        </ol>
      </div>
    </section>

    <section id="history" class="mx-auto mt-28 max-w-content scroll-mt-28 px-6" aria-labelledby="history-title">
      <h2 id="history-title" class="font-display text-3xl leading-tight text-neutral-800 md:text-4xl dark:text-neutral-100">
        Page history
      </h2>
      <ul class="mt-6 border-b border-default">
        <li v-for="entry in comparison.history" :key="entry.date + entry.text" class="flex flex-wrap gap-x-7 gap-y-1 border-t border-default py-3 text-base">
          <span class="w-40 shrink-0 font-mono text-sm text-neutral-600 dark:text-neutral-400">{{ formatComparisonDate(entry.date) }}</span>
          <span class="flex-1 text-neutral-700 dark:text-neutral-300">{{ entry.text }}</span>
        </li>
      </ul>
      <p class="mt-6 text-lg text-neutral-700 dark:text-neutral-300">
        Something here wrong or out of date?
        <a :href="editUrl" target="_blank" rel="noopener" class="font-semibold text-neutral-900 underline decoration-1 underline-offset-4 dark:text-neutral-100">Edit this page on GitHub</a>
        or
        <a :href="issueUrl" target="_blank" rel="noopener" class="font-semibold text-neutral-900 underline decoration-1 underline-offset-4 dark:text-neutral-100">open an issue</a>.
      </p>
    </section>
  </article>
</template>
