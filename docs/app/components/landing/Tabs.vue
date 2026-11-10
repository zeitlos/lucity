<script setup lang="ts">
defineProps<{
  options: string[];
  label: string;
  idPrefix?: string;
}>();

const selected = defineModel<number>({ required: true });
</script>

<template>
  <div class="tabs" :role="idPrefix ? 'tablist' : 'group'" :aria-label="label">
    <button
      v-for="(option, index) in options"
      :id="idPrefix ? `${idPrefix}-tab-${index}` : undefined"
      :key="option"
      type="button"
      class="tab"
      :class="{ on: selected === index }"
      :role="idPrefix ? 'tab' : undefined"
      :aria-selected="idPrefix ? selected === index : undefined"
      :aria-controls="idPrefix ? `${idPrefix}-panel` : undefined"
      :aria-pressed="idPrefix ? undefined : selected === index"
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
  color: var(--tab-off, oklch(0.40 0.04 55));
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
</style>
