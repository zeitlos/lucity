<script setup lang="ts">
interface Step {
  title: string;
  text: string;
}

withDefaults(defineProps<{
  steps?: Step[];
}>(), {
  steps: () => [],
});
</script>

<template>
  <ol class="step-flow">
    <li
      v-for="(step, index) in steps"
      :key="step.title"
      class="step"
    >
      <span class="step-marker">{{ index + 1 }}</span>
      <div class="step-body">
        <p class="step-title">{{ step.title }}</p>
        <p class="step-text">{{ step.text }}</p>
      </div>
    </li>
  </ol>
</template>

<style scoped>
.step-flow {
  list-style: none;
  margin: 2rem 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.step {
  position: relative;
  display: flex;
  gap: 1rem;
  padding-bottom: 1.75rem;
}

.step:last-child {
  padding-bottom: 0;
}

.step::before {
  content: '';
  position: absolute;
  inset-inline-start: 1rem;
  top: 2.25rem;
  bottom: 0.25rem;
  width: 1px;
  background: var(--ui-border);
}

.step:last-child::before {
  display: none;
}

.step-marker {
  position: relative;
  z-index: 1;
  flex: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border-radius: 999px;
  font-size: 0.875rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: var(--ui-primary);
  background: color-mix(in oklab, var(--ui-primary) 14%, var(--ui-bg));
  border: 1px solid color-mix(in oklab, var(--ui-primary) 35%, transparent);
}

.step-body {
  padding-top: 0.1875rem;
}

.step-title {
  margin: 0;
  font-weight: 600;
  color: var(--ui-text-highlighted);
}

.step-text {
  margin: 0.25rem 0 0;
  line-height: 1.65;
  color: var(--ui-text-muted);
  text-wrap: pretty;
}
</style>
