<script setup lang="ts">
import { comparisonCheckedOn, formatComparisonDate } from '~/utils/comparisons';
import type { ComparisonPlatform, MatrixGroup } from '~/utils/comparisons/types';

defineProps<{
  groups: MatrixGroup[];
  platforms: ComparisonPlatform[];
}>();
</script>

<template>
  <div class="overflow-x-auto border-t border-default">
    <table class="w-full min-w-[1180px] table-fixed border-collapse text-left text-[14.5px]/snug">
      <colgroup>
        <col class="w-[200px]">
        <col v-for="platform in platforms" :key="platform.key">
      </colgroup>
      <thead>
        <tr>
          <th scope="col" class="sticky left-0 z-10 bg-(--ui-bg) py-4 pr-3 align-bottom text-sm font-medium text-neutral-600 dark:text-neutral-400">
            Checked {{ formatComparisonDate(comparisonCheckedOn) }}
          </th>
          <th
            v-for="platform in platforms"
            :key="platform.key"
            scope="col"
            class="border-b border-stone-300 px-3 py-4 align-top dark:border-stone-700"
            :class="platform.key === 'lucity' ? 'bg-white/55 dark:bg-white/4' : ''"
          >
            <span class="flex items-center gap-2 text-base font-semibold text-neutral-900 dark:text-neutral-100">
              <ComparePlatformIcon :name="platform.name" :icon="platform.icon" :lucity="platform.key === 'lucity'" />
              {{ platform.name }}
            </span>
            <NuxtLink
              v-if="platform.to"
              :to="platform.to"
              class="mt-2 block text-[13px] font-medium text-neutral-700 underline decoration-stone-400 decoration-1 underline-offset-4 hover:decoration-current dark:text-neutral-300"
            >
              Read the comparison
            </NuxtLink>
            <span v-else class="mt-2 block text-[13px] font-medium text-neutral-600 dark:text-neutral-400">Us</span>
          </th>
        </tr>
      </thead>
      <tbody v-for="group in groups" :key="group.title">
        <tr>
          <th
            scope="colgroup"
            :colspan="platforms.length + 1"
            class="border-b border-stone-300 pt-8 pb-2.5 font-display text-2xl font-normal text-neutral-800 dark:border-stone-700 dark:text-neutral-100"
          >
            <span class="sticky left-0">{{ group.title }}</span>
          </th>
        </tr>
        <tr v-for="row in group.rows" :key="row.name" class="border-b border-default align-top">
          <th scope="row" class="sticky left-0 z-10 bg-(--ui-bg) py-3 pr-3 font-semibold text-neutral-900 dark:text-neutral-100">
            {{ row.name }}
          </th>
          <td
            v-for="platform in platforms"
            :key="platform.key"
            class="px-3 py-3"
            :class="platform.key === 'lucity' ? 'bg-white/55 dark:bg-white/4' : ''"
          >
            <span v-if="row.cells[platform.key]" class="flex gap-2">
              <CompareMark :mark="row.cells[platform.key]!.mark" />
              <span
                :class="row.cells[platform.key]!.mark === 'info' ? 'font-medium text-neutral-900 dark:text-neutral-100' : row.cells[platform.key]!.mark === 'no' ? 'text-neutral-600 dark:text-neutral-400' : 'text-neutral-800 dark:text-neutral-200'"
              >{{ row.cells[platform.key]!.text }}</span>
            </span>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
