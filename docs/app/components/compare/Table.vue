<script setup lang="ts">
import type { Cell, CompareGroup } from '~/utils/comparisons/types';

const props = defineProps<{
  groups: CompareGroup[];
  name: string;
  icon: string;
}>();

const onlyDifferences = ref(true);

const total = computed(() => props.groups.reduce((sum, group) => sum + group.rows.length, 0));
const differing = computed(() => props.groups.reduce((sum, group) => sum + group.rows.filter(row => row.differs).length, 0));

const visible = computed(() => props.groups
  .map(group => ({ ...group, rows: group.rows.filter(row => !onlyDifferences.value || row.differs) }))
  .filter(group => group.rows.length > 0));

const defaultLabels = { yes: 'Yes', partly: 'Partly', no: 'No', info: '' };

function label(cell: Cell) {
  return cell.label ?? defaultLabels[cell.mark];
}
</script>

<template>
  <div>
    <div class="mt-8 flex flex-wrap items-center gap-x-7 gap-y-3 border-y border-default py-3.5 text-[15px]">
      <span class="text-neutral-600 dark:text-neutral-400">
        {{ onlyDifferences ? `Showing the ${differing} rows that differ, out of ${total}.` : `Showing all ${total} rows.` }}
      </span>
      <span class="flex flex-wrap items-center gap-x-5 gap-y-2">
        <span class="inline-flex items-center gap-1.5"><CompareMark mark="yes" /><span class="font-semibold">Yes</span></span>
        <span class="inline-flex items-center gap-1.5"><CompareMark mark="partly" /><span class="font-semibold">Partly</span></span>
        <span class="inline-flex items-center gap-1.5"><CompareMark mark="no" /><span class="font-semibold">No</span></span>
      </span>
      <button
        type="button"
        :aria-pressed="onlyDifferences"
        class="ml-auto inline-flex min-h-11 items-center rounded-lg border px-4 font-semibold transition-colors"
        :class="onlyDifferences ? 'border-neutral-800 bg-neutral-800 text-neutral-50 dark:border-neutral-100 dark:bg-neutral-100 dark:text-neutral-900' : 'border-stone-300 bg-white/50 text-neutral-800 hover:bg-white dark:border-stone-700 dark:bg-white/5 dark:text-neutral-100'"
        @click="onlyDifferences = !onlyDifferences"
      >
        Only differences
      </button>
    </div>

    <div class="md:hidden">
      <section v-for="group in visible" :key="group.title" :aria-label="group.title">
        <h3 class="border-t border-default pt-8 pb-2 font-display text-2xl text-neutral-800 dark:text-neutral-100">
          {{ group.title }}
        </h3>
        <dl class="border-b border-default">
          <div v-for="row in group.rows" :key="row.key" class="border-t border-default py-3.5">
            <dt class="font-semibold text-neutral-900 dark:text-neutral-100">{{ row.name }}</dt>
            <dd v-for="side in [{ key: 'lucity', name: 'Lucity', icon: '', cell: row.lucity }, { key: 'other', name, icon, cell: row.other }]" :key="side.key" class="mt-2 flex items-start gap-2.5 text-[15px]/snug">
              <span class="inline-flex w-22 shrink-0 items-center gap-1.5 text-neutral-600 dark:text-neutral-400">
                <ComparePlatformIcon :name="side.name" :icon="side.icon" :lucity="side.key === 'lucity'" />{{ side.name }}
              </span>
              <CompareMark :mark="side.cell.mark" />
              <span>
                <span class="font-semibold text-neutral-900 dark:text-neutral-100">{{ label(side.cell) }}</span>
                <span v-if="side.cell.text" class="text-neutral-600 dark:text-neutral-400"><template v-if="label(side.cell)"> · </template>{{ side.cell.text }}</span>
                <a
                  v-if="side.cell.href"
                  :href="side.cell.href"
                  target="_blank"
                  rel="noopener"
                  class="ml-1.5 inline-flex items-center gap-0.5 text-[13px] whitespace-nowrap text-neutral-600 underline decoration-stone-400 decoration-1 underline-offset-4 dark:text-neutral-400"
                >source<UIcon name="i-lucide-arrow-up-right" class="size-3.5" /></a>
              </span>
            </dd>
          </div>
        </dl>
      </section>
    </div>

    <div class="hidden overflow-x-auto md:block">
      <table class="w-full min-w-3xl table-fixed border-collapse text-base/snug">
        <colgroup>
          <col class="w-[24%]">
          <col class="w-[38%]">
          <col class="w-[38%]">
        </colgroup>
        <thead>
          <tr class="text-left text-[15px] text-neutral-600 dark:text-neutral-400">
            <th scope="col" class="py-3.5 pr-4 font-semibold">Feature</th>
            <th scope="col" class="py-3.5 pr-4 font-semibold">
              <span class="inline-flex items-center gap-2"><ComparePlatformIcon name="Lucity" icon="" lucity />Lucity</span>
            </th>
            <th scope="col" class="py-3.5 pr-4 font-semibold">
              <span class="inline-flex items-center gap-2"><ComparePlatformIcon :name="name" :icon="icon" />{{ name }}</span>
            </th>
          </tr>
        </thead>
        <tbody v-for="group in visible" :key="group.title">
          <tr>
            <th colspan="3" scope="colgroup" class="border-t border-default pt-10 pb-3 text-left font-display text-[28px] font-normal text-neutral-800 dark:text-neutral-100">
              {{ group.title }}
            </th>
          </tr>
          <tr v-for="row in group.rows" :key="row.key" class="border-t border-default align-top">
            <th scope="row" class="py-3.5 pr-4 text-left font-semibold text-neutral-900 dark:text-neutral-100">
              {{ row.name }}
            </th>
            <td v-for="(cell, side) in [row.lucity, row.other]" :key="side" class="py-3.5 pr-5">
              <span class="flex gap-2.5">
                <CompareMark :mark="cell.mark" />
                <span>
                  <span class="font-semibold text-neutral-900 dark:text-neutral-100">{{ label(cell) }}</span>
                  <span v-if="cell.text" class="text-neutral-600 dark:text-neutral-400"><template v-if="label(cell)"> · </template>{{ cell.text }}</span>
                  <a
                    v-if="cell.href"
                    :href="cell.href"
                    target="_blank"
                    rel="noopener"
                    class="ml-1.5 inline-flex items-center gap-0.5 text-[13px] whitespace-nowrap text-neutral-600 underline decoration-stone-400 decoration-1 underline-offset-4 hover:decoration-current dark:text-neutral-400"
                  >source<UIcon name="i-lucide-arrow-up-right" class="size-3.5" /></a>
                </span>
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
