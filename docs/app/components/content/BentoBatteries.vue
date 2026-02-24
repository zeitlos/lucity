<script setup lang="ts">
import { ref, watch } from 'vue';
import { useBentoVisible } from './useBentoVisible';

const root = ref<HTMLElement | null>(null);
const visible = useBentoVisible(root);
const rowCount = ref(0);

const columns = ['id', 'name', 'email', 'role', 'created_at'];
const rows = [
  ['1', 'alice', 'alice@acme.com', 'admin', '2025-01-12'],
  ['2', 'bob', 'bob@acme.com', 'dev', '2025-02-03'],
  ['3', 'carol', 'carol@acme.com', 'dev', '2025-02-14'],
  ['4', 'dave', 'dave@acme.com', 'viewer', '2025-03-01'],
  ['5', 'eve', 'eve@acme.com', 'dev', '2025-03-22'],
  ['6', 'frank', 'frank@acme.com', 'dev', '2025-04-08'],
  ['7', 'grace', 'grace@acme.com', 'admin', '2025-05-15'],
  ['8', 'heidi', 'heidi@acme.com', 'viewer', '2025-06-01'],
];

watch(visible, (v) => {
  if (!v) return;
  rows.forEach((_, i) => {
    setTimeout(() => { rowCount.value = i + 1; }, 250 + i * 200);
  });
});
</script>

<template>
  <div
    ref="root"
    class="bento-batteries"
  >
    <!-- Table positioned to bleed bottom-left — most of it is off-screen -->
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
          :style="{ animationDelay: `${i * 60}ms` }"
        >
          <span
            v-for="(cell, j) in row"
            :key="j"
            class="bento-cell"
          >{{ cell }}</span>
        </div>

        <!-- Footer -->
        <div
          v-if="rowCount >= rows.length"
          class="bento-footer"
        >
          {{ rows.length }} rows &middot; 8ms
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.bento-batteries {
  height: 200px;
  position: relative;
  overflow: visible;
  padding: 0;
}

/* Table is positioned so only the top-left corner peeks into the card.
   It bleeds off the right and bottom edges.
   The card's overflow:hidden (on .bento-card in BentoGrid) clips the rest. */
.bento-table-wrap {
  position: absolute;
  bottom: -80px;
  right: -140px;
  z-index: 1;
}

@media (min-width: 640px) {
  .bento-table-wrap {
    bottom: -90px;
    right: -160px;
  }
}

.bento-table {
  width: 520px;
  border: 1px solid var(--ui-border);
  border-radius: 10px;
  background: var(--ui-bg-elevated);
  overflow: hidden;
  box-shadow: 0 4px 24px oklch(0 0 0 / 0.06), 0 1px 4px oklch(0 0 0 / 0.04);
}

@media (min-width: 640px) {
  .bento-table {
    width: 600px;
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
  grid-template-columns: 36px 80px 1.4fr 60px 90px;
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
