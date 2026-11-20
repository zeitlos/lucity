<script setup lang="ts">
defineProps<{ single?: boolean }>();
</script>

<template>
  <div class="grid grid-cols-[minmax(0,2.05fr)_minmax(0,1fr)] gap-2.5 text-left text-[15px]/[1.45]">
    <div class="flex flex-col px-2 pt-4 pb-1.5">
      <div
        class="mb-4 grid content-start gap-x-4 gap-y-2.5"
        :class="single ? 'grid-cols-1' : 'grid-cols-[minmax(0,1.12fr)_minmax(0,1fr)]'"
      >
        <slot />
      </div>
      <div class="mt-auto flex flex-wrap items-center justify-between gap-x-6 gap-y-2.5 border-t border-stone-200 px-3 pt-3.5 pb-3.5 dark:border-stone-800">
        <slot name="foot" />
      </div>
    </div>
    <slot name="stage" />
  </div>
</template>
