<script setup lang="ts">
import { ref, watch } from 'vue';
import { useBentoVisible } from './useBentoVisible';

const root = ref<HTMLElement | null>(null);
const visible = useBentoVisible(root);
const rowCount = ref(0);

const columns = ['id', 'name', 'email'];
const rows = [
  ['1', 'alice', 'alice@acme.com'],
  ['2', 'bob', 'bob@acme.com'],
  ['3', 'carol', 'carol@acme.com'],
];

watch(visible, (v) => {
  if (!v) return;
  setTimeout(() => { rowCount.value = 1; }, 300);
  setTimeout(() => { rowCount.value = 2; }, 600);
  setTimeout(() => { rowCount.value = 3; }, 900);
});
</script>

<template>
  <div
    ref="root"
    class="bento-batteries"
  >
    <div class="bento-table">
      <!-- Header -->
      <div class="bento-row bento-header">
        <span
          v-for="col in columns"
          :key="col"
          class="bento-cell"
        >{{ col }}</span>
      </div>

      <!-- Rows -->
      <div
        v-for="(row, i) in rows"
        :key="i"
        class="bento-row"
        :class="{ 'bento-row-visible': rowCount > i }"
        :style="{ animationDelay: `${i * 100}ms` }"
      >
        <span
          v-for="(cell, j) in row"
          :key="j"
          class="bento-cell"
        >{{ cell }}</span>
      </div>

      <!-- Footer -->
      <div
        v-if="rowCount >= 3"
        class="bento-footer"
      >
        3 rows &middot; 12ms
      </div>
    </div>
  </div>
</template>

<style scoped>
.bento-batteries {
  min-height: 140px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.bento-table {
  width: 100%;
  max-width: 280px;
  border: 1px solid var(--ui-border);
  border-radius: 8px;
  background: var(--ui-bg-elevated);
  overflow: hidden;
}

.bento-row {
  display: grid;
  grid-template-columns: 28px 1fr 1.5fr;
  gap: 0;
  opacity: 0;
}

.bento-header {
  opacity: 1;
  background: var(--ui-bg-muted);
  border-bottom: 1px solid var(--ui-border);
}

.bento-header .bento-cell {
  font-weight: 600;
  color: var(--ui-text-muted);
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.bento-row-visible {
  animation: bento-row-in 0.3s ease both;
}

.bento-cell {
  padding: 6px 8px;
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--ui-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  border-right: 1px solid var(--ui-border);
}

.bento-cell:last-child {
  border-right: none;
}

.bento-row:not(.bento-header) + .bento-row {
  border-top: 1px solid var(--ui-border);
}

.bento-footer {
  padding: 4px 8px;
  font-size: 10px;
  color: var(--ui-text-muted);
  border-top: 1px solid var(--ui-border);
  text-align: right;
  animation: bento-fade-in 0.3s ease both;
  animation-delay: 0.4s;
}

@keyframes bento-row-in {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes bento-fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>
