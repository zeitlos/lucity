<script setup lang="ts">
defineProps<{
  options: string[];
  label: string;
  panel: string;
}>();

const selected = defineModel<number>({ required: true });
</script>

<template>
  <div class="tabs" role="tablist" :aria-label="label">
    <button
      v-for="(option, index) in options"
      :key="option"
      type="button"
      class="tab"
      :class="{ on: selected === index }"
      role="tab"
      :aria-selected="selected === index"
      :aria-controls="panel"
      @click="selected = index"
    >
      {{ option }}
    </button>
  </div>
</template>

<style scoped>
.tabs {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 4px;
  border-radius: 12px;
  background: oklch(0.25 0.04 55 / .07);
}

.tab {
  min-height: 44px;
  padding: 0 18px;
  border: 0;
  border-radius: 9px;
  background: none;
  font: inherit;
  font-size: 16px;
  font-weight: 600;
  color: oklch(0.40 0.04 55);
  cursor: pointer;
}

.tab.on {
  color: oklch(0.25 0.03 55);
  background: oklch(0.99 0.008 80);
  box-shadow: 0 1px 2px oklch(0.25 0.04 55 / .14), 0 0 0 1px oklch(0.25 0.04 55 / .08);
}

.tab:focus-visible {
  outline: 2px solid oklch(0.55 0.15 300);
  outline-offset: 2px;
}

.dark .tabs {
  background: oklch(1 0 0 / .07);
}

.dark .tab {
  color: oklch(0.72 0.03 70);
}

.dark .tab.on {
  color: oklch(0.95 0.02 80);
  background: oklch(0.32 0.03 55);
  box-shadow: 0 1px 2px oklch(0 0 0 / .3), 0 0 0 1px oklch(1 0 0 / .08);
}

@media (max-width: 640px) {
  .tabs {
    display: flex;
    flex-wrap: nowrap;
    width: 100%;
  }

  .tab {
    flex: 1 1 auto;
    min-height: 48px;
    padding: 0 6px;
    font-size: 15px;
    white-space: nowrap;
  }
}

@media (max-width: 360px) {
  .tab {
    padding: 0 3px;
    font-size: 14px;
  }
}
</style>
