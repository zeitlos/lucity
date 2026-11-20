<script setup lang="ts">
defineProps<{
  to: string;
  label: string;
  icon?: string;
  path?: string;
}>();
</script>

<template>
  <NuxtLink
    :to="to"
    class="flex min-h-10 items-center gap-[11px] rounded-[10px] px-3 text-[15.5px] font-medium text-stone-900 transition-colors hover:bg-stone-900/5 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-(--nav-accent) dark:text-stone-100 dark:hover:bg-white/6"
  >
    <span class="flex flex-none text-stone-800 dark:text-stone-300">
      <svg v-if="path" class="size-4.5" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path :d="path" /></svg>
      <UIcon v-else-if="icon" :name="icon" class="size-4.5" />
    </span>
    {{ label }}
  </NuxtLink>
</template>
