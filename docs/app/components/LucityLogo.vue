<script setup lang="ts">
const props = defineProps<{ suffix?: string }>();

const wordmark = computed(() => (props.suffix ? `Lucity ${props.suffix}` : 'Lucity'));
</script>

<template>
  <span class="inline-flex items-center gap-[0.833em] leading-none">
    <LucityMark class="h-[0.98em] w-auto" />
    <span class="wordmark">{{ wordmark }}</span>
  </span>
</template>

<style scoped>
.wordmark {
  font-family: 'Redaction 35', Georgia, serif;
  font-weight: 400;
}
</style>
