<script setup lang="ts">
defineProps<{
  name: string;
  icon: string;
  lucity?: boolean;
}>();
</script>

<template>
  <span class="inline-flex size-[1.1em] shrink-0 items-center justify-center" aria-hidden="true">
    <LucityMark v-if="lucity" class="h-auto w-full" />
    <UIcon v-else-if="icon" :name="icon" class="size-full" />
    <span
      v-else
      class="flex size-full items-center justify-center rounded-[0.22em] bg-stone-800 text-[0.62em] leading-none font-bold text-stone-50 dark:bg-stone-200 dark:text-stone-900"
    >{{ name.charAt(0) }}</span>
  </span>
</template>
