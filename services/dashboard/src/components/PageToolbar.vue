<script setup lang="ts">
defineProps<{
  title: string;
  count?: number;
}>();
</script>

<template>
  <div class="flex shrink-0 flex-wrap items-center gap-x-2.5 border-b border-border/70 px-4 pb-3 sm:flex-nowrap sm:pb-0">
    <div class="flex h-12 items-center gap-2.5">
      <h1 class="text-base font-semibold text-foreground">{{ title }}</h1>
      <span
        v-if="count !== undefined"
        class="inline-flex h-5 items-center rounded-md bg-muted px-2 text-xs font-medium text-muted-foreground"
      >
        {{ count }}
      </span>
    </div>
    <div class="flex-1" />
    <slot />
  </div>
</template>
