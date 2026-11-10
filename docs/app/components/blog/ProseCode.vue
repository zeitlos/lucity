<script setup lang="ts">
defineOptions({ inheritAttrs: false });

defineProps<{
  code?: string;
  language?: string;
  filename?: string;
  highlights?: number[];
  meta?: string;
  class?: string;
}>();
</script>

<template>
  <div class="term">
    <div class="term-bar" aria-hidden="true">
      <i /><i /><i />
      <span>{{ filename || language }}</span>
    </div>
    <pre :class="$props.class" tabindex="0"><slot /></pre>
  </div>
</template>

<style scoped>
.term {
  margin-top: 30px;
  border-radius: 14px;
  overflow: hidden;
  background: oklch(0.20 0.03 55);
  box-shadow: 0 28px 56px -22px oklch(0.25 0.04 55 / .42);
}

.dark .term {
  background: oklch(0.155 0.025 55);
  box-shadow: 0 0 0 1px oklch(1 0 0 / .08);
}

.term-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 13px 18px;
  border-bottom: 1px solid oklch(1 0 0 / .08);
  font-size: 13px;
  line-height: 1.55;
  color: oklch(0.74 0.03 70);
}

.term-bar i {
  width: 11px;
  height: 11px;
  border-radius: 50%;
  background: oklch(0.40 0.03 55);
}

.term-bar span {
  margin-left: 8px;
}

.term pre {
  margin: 0;
  padding: 20px 24px 24px;
  font-family: var(--font-mono);
  font-size: 15px;
  line-height: 1.72;
  color: oklch(0.90 0.03 80);
  white-space: pre;
  overflow-x: auto;
}

.term pre:focus-visible {
  outline: 2px solid oklch(0.72 0.12 300);
  outline-offset: -2px;
}

@media (max-width: 640px) {
  .term pre {
    padding: 16px 18px 20px;
    font-size: 13.5px;
  }
}
</style>
