<script setup lang="ts">
defineProps<{
  title?: string;
}>();
</script>

<template>
  <div class="terminal">
    <div v-if="title" class="terminal-bar" aria-hidden="true"><i /><i /><i /><span>{{ title }}</span></div>
    <div class="terminal-body"><slot /></div>
  </div>
</template>

<style scoped>
.terminal {
  overflow: hidden;
  border-radius: 14px;
  background: oklch(0.20 0.03 55);
  box-shadow: 0 28px 56px -22px oklch(0.25 0.04 55 / .42);
}

.terminal-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 13px 18px;
  border-bottom: 1px solid oklch(1 0 0 / .08);
  font-size: 13px;
  line-height: 1.55;
  color: oklch(0.74 0.03 70);
}

.terminal-bar i {
  flex: none;
  width: 11px;
  height: 11px;
  border-radius: 50%;
  background: oklch(0.40 0.03 55);
}

.terminal-bar span {
  margin-left: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.terminal-body {
  padding: 20px 24px 24px;
  font-family: 'Fira Code', ui-monospace, monospace;
  font-size: 13.5px;
  line-height: 1.72;
  color: oklch(0.90 0.03 80);
}

.terminal-body :deep(.c-dim) {
  color: oklch(0.70 0.03 70);
}

.terminal-body :deep(.c-ok) {
  color: oklch(0.80 0.16 160);
}

.terminal-body :deep(.c-acc) {
  color: oklch(0.80 0.11 300);
}

.terminal-body :deep(.c-warm) {
  color: oklch(0.84 0.11 70);
}

@media (max-width: 520px) {
  .terminal-body {
    padding: 16px 16px 18px;
    font-size: 12.5px;
  }
}
</style>
