<script setup lang="ts">
defineProps<{
  items: { name: string; text: string }[];
}>();
</script>

<template>
  <div class="grid grid-cols-[repeat(auto-fit,minmax(min(250px,100%),1fr))] gap-x-10 border-b border-stone-300 dark:border-stone-800">
    <div v-for="item in items" :key="item.name" class="row-span-2 grid grid-rows-subgrid content-start gap-y-0 border-t border-stone-300 py-5.5 dark:border-stone-800">
      <b class="block text-lg font-semibold text-balance text-stone-900 dark:text-stone-100">{{ item.name }}</b>
      <span class="mt-1 block text-base text-stone-700 dark:text-stone-400">{{ item.text }}</span>
    </div>
  </div>
</template>
