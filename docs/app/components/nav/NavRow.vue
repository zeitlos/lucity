<script setup lang="ts">
defineProps<{
  to: string;
  icon: string;
  title: string;
  text: string;
  tint: 'violet' | 'orange' | 'rose';
}>();

const tints = {
  violet: 'bg-[oklch(0.93_0.035_300)] dark:bg-[oklch(0.34_0.06_300)]',
  orange: 'bg-[oklch(0.93_0.045_68)] dark:bg-[oklch(0.35_0.05_68)]',
  rose: 'bg-[oklch(0.93_0.035_8)] dark:bg-[oklch(0.34_0.06_8)]',
};
</script>

<template>
  <NuxtLink
    :to="to"
    class="flex items-start gap-3.5 rounded-xl px-3 py-2.5 text-stone-900 transition-colors hover:bg-stone-900/5 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-(--nav-accent) dark:text-stone-100 dark:hover:bg-white/6"
  >
    <span
      class="flex size-10 flex-none items-center justify-center rounded-[11px]"
      :class="tints[tint]"
    >
      <UIcon :name="icon" class="size-5" />
    </span>
    <span>
      <b class="block text-[16.5px]/[1.3] font-semibold">{{ title }}</b>
      <span class="mt-0.5 block text-[14.5px]/[1.4] text-stone-700 dark:text-stone-400">{{ text }}</span>
    </span>
  </NuxtLink>
</template>
