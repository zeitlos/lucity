<script setup lang="ts">
const year = new Date().getFullYear();
</script>

<template>
  <div class="text-sm text-muted">
    <span>© {{ year }} </span>
    <a
      href="https://zeitlos.software"
      target="_blank"
      class="underline decoration-muted/30 underline-offset-2 transition-colors hover:text-foreground"
    >zeitlos.software</a>
    <span class="mx-1.5 text-muted/40">·</span>
    <span>cooked with care in Switzerland</span>
  </div>
</template>
