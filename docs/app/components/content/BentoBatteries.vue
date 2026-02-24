<script setup lang="ts">
import { ref, watch } from 'vue';
import { useBentoVisible } from './useBentoVisible';

const root = ref<HTMLElement | null>(null);
const visible = useBentoVisible(root);
const rowCount = ref(0);

const columns = ['id', 'name', 'email', 'role'];
const rows = [
  ['1', 'alice', 'alice@acme.com', 'admin'],
  ['2', 'bob', 'bob@acme.com', 'dev'],
  ['3', 'carol', 'carol@acme.com', 'dev'],
  ['4', 'dave', 'dave@acme.com', 'viewer'],
  ['5', 'eve', 'eve@acme.com', 'dev'],
];

watch(visible, (v) => {
  if (!v) return;
  setTimeout(() => { rowCount.value = 1; }, 300);
  setTimeout(() => { rowCount.value = 2; }, 550);
  setTimeout(() => { rowCount.value = 3; }, 800);
  setTimeout(() => { rowCount.value = 4; }, 1050);
  setTimeout(() => { rowCount.value = 5; }, 1300);
});
</script>

<template>
  <div
    ref="root"
    class="bento-batteries"
  >
    <!-- Table positioned to bleed bottom-left -->
    <div class="bento-table-wrap">
      <div class="bento-table">
        <!-- Window chrome bar -->
        <div class="bento-chrome">
          <span class="bento-chrome-dot" />
          <span class="bento-chrome-dot" />
          <span class="bento-chrome-dot" />
          <span class="bento-chrome-title">users</span>
        </div>

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
          :style="{ animationDelay: `${i * 80}ms` }"
        >
          <span
            v-for="(cell, j) in row"
            :key="j"
            class="bento-cell"
          >{{ cell }}</span>
        </div>

        <!-- Footer -->
        <div
          v-if="rowCount >= 5"
          class="bento-footer"
        >
          5 rows &middot; 8ms
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.bento-batteries {
  min-height: 180px;
  position: relative;
  overflow: visible;
  padding: 0;
}

.bento-table-wrap {
  position: absolute;
  bottom: -20px;
  left: -36px;
  z-index: 1;
}

@media (min-width: 640px) {
  .bento-table-wrap {
    bottom: -24px;
    left: -44px;
  }
}

.bento-table {
  width: 380px;
  border: 1px solid var(--ui-border);
  border-radius: 10px;
  background: var(--ui-bg-elevated);
  overflow: hidden;
  box-shadow: 0 4px 24px oklch(0 0 0 / 0.06), 0 1px 4px oklch(0 0 0 / 0.04);
}

@media (min-width: 640px) {
  .bento-table {
    width: 440px;
  }
}

/* Browser-style window chrome */
.bento-chrome {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: var(--ui-bg-muted);
  border-bottom: 1px solid var(--ui-border);
}

.bento-chrome-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--ui-border);
}

.bento-chrome-dot:nth-child(1) { background: oklch(0.65 0.18 25); }
.bento-chrome-dot:nth-child(2) { background: oklch(0.75 0.15 90); }
.bento-chrome-dot:nth-child(3) { background: oklch(0.65 0.16 145); }

.bento-chrome-title {
  margin-left: 8px;
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--ui-text-muted);
}

.bento-row {
  display: grid;
  grid-template-columns: 36px 1fr 1.6fr 0.7fr;
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
  letter-spacing: 0.06em;
}

.bento-row-visible {
  animation: bento-row-in 0.3s ease both;
}

.bento-cell {
  padding: 7px 10px;
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
  padding: 6px 10px;
  font-size: 10px;
  font-family: var(--font-mono);
  color: var(--ui-text-muted);
  border-top: 1px solid var(--ui-border);
  text-align: right;
  animation: bento-fade-in 0.3s ease both;
  animation-delay: 0.3s;
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
