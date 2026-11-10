<script setup lang="ts">
defineProps<{
  id: string;
  promise: string;
  rest: string;
}>();
</script>

<template>
  <div class="flex flex-wrap items-end justify-between gap-x-18 gap-y-6">
    <div class="min-w-0 flex-[1_1_560px]">
      <h2 :id="id" class="font-display text-4xl/11 text-balance text-stone-900 md:text-5xl/14 lg:text-6xl/17 dark:text-stone-100">
        <span class="relative whitespace-nowrap">{{ promise }}<img src="/img/underline-keep.svg" alt="" class="pointer-events-none absolute bottom-0 left-0 h-[0.32em] w-full"></span>,<br>
        <span class="text-stone-500">{{ rest }}</span>
      </h2>
    </div>
    <p class="max-w-[540px] flex-[1_1_380px] text-xl leading-relaxed text-pretty text-stone-700 dark:text-stone-300">
      <slot />
    </p>
  </div>
</template>
