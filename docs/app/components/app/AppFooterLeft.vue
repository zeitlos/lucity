<script setup lang="ts">
const year = new Date().getFullYear();
</script>

<template>
  <div class="flex items-center gap-1.5 text-sm text-muted">
    <span>© {{ year }}</span>
    <a
      href="https://zeitlos.software"
      target="_blank"
      class="underline decoration-muted/30 underline-offset-2 transition-colors hover:text-foreground"
    >zeitlos.software</a>
    <span class="text-muted/40">·</span>
    <span>cooked with care in Switzerland</span>
  </div>
</template>
