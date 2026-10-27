<script setup lang="ts">
defineProps<{
  title: string;
  count?: number;
}>();
</script>

<template>
  <div class="flex h-12 shrink-0 items-center gap-2.5 border-b border-border/70 px-4">
    <h1 class="text-[15px] font-semibold text-foreground">{{ title }}</h1>
    <span
      v-if="count !== undefined"
      class="inline-flex h-5 items-center rounded-[6px] bg-muted px-[7px] text-xs font-medium text-muted-foreground"
    >
      {{ count }}
    </span>
    <div class="flex-1" />
    <slot />
  </div>
</template>
