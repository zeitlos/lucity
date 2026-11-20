<script setup lang="ts">
defineProps<{ suffix?: string }>();
</script>

<template>
  <span class="inline-flex items-baseline gap-[0.61em] leading-none">
    <LucityMark class="h-[1em] w-auto" />
    <span class="inline-flex items-baseline gap-2">
      <LucityWordmark class="h-[1em] w-auto" />
      <span v-if="suffix" class="text-[1.372em] font-medium text-muted">{{ suffix }}</span>
    </span>
  </span>
</template>
