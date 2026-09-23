<script setup lang="ts">
interface Item {
  title: string;
  text: string;
  icon?: string;
}

withDefaults(defineProps<{
  items?: Item[];
  columns?: number;
}>(), {
  items: () => [],
  columns: 3,
});
</script>

<template>
  <div
    class="feature-cards"
    :style="{ '--columns': columns }"
  >
    <div
      v-for="item in items"
      :key="item.title"
      class="feature-card"
    >
      <span
        v-if="item.icon"
        class="feature-card-icon"
      >
        <UIcon
          :name="item.icon"
          class="size-5"
        />
      </span>
      <p class="feature-card-title">{{ item.title }}</p>
      <p class="feature-card-text">{{ item.text }}</p>
    </div>
  </div>
</template>

<style scoped>
.feature-cards {
  display: grid;
  gap: 1rem;
  margin: 2rem 0;
  grid-template-columns: repeat(1, minmax(0, 1fr));
}

@media (min-width: 640px) {
  .feature-cards {
    grid-template-columns: repeat(min(var(--columns), 2), minmax(0, 1fr));
  }
}

@media (min-width: 1024px) {
  .feature-cards {
    grid-template-columns: repeat(var(--columns), minmax(0, 1fr));
  }
}

.feature-card {
  padding: 1.25rem;
  border: 1px solid var(--ui-border);
  border-radius: 0.875rem;
  background: var(--ui-bg-elevated);
}

.feature-card-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  margin-bottom: 0.875rem;
  border-radius: 0.625rem;
  color: var(--ui-primary);
  background: color-mix(in oklab, var(--ui-primary) 14%, transparent);
}

.feature-card-title {
  margin: 0;
  font-weight: 600;
  color: var(--ui-text-highlighted);
}

.feature-card-text {
  margin: 0.375rem 0 0;
  font-size: 0.9375rem;
  line-height: 1.6;
  color: var(--ui-text-muted);
  text-wrap: pretty;
}
</style>
