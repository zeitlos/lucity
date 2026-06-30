<script setup lang="ts">
import { ref, watch } from 'vue';
import { useBentoVisible } from './useBentoVisible';

const root = ref<HTMLElement | null>(null);
const visible = useBentoVisible(root);
const filled = ref(false);
const amount = ref(0);

const meters = [
  { label: 'vCPU', fill: 58 },
  { label: 'Memory', fill: 41 },
  { label: 'Egress', fill: 24 },
];

const target = 7.42;

watch(visible, (v) => {
  if (!v) return;
  setTimeout(() => { filled.value = true; }, 150);

  const start = performance.now();
  const duration = 1100;
  const tick = (now: number) => {
    const k = Math.min(1, (now - start) / duration);
    const eased = 1 - Math.pow(1 - k, 3);
    amount.value = target * eased;
    if (k < 1) requestAnimationFrame(tick);
  };
  setTimeout(() => requestAnimationFrame(tick), 300);
});
</script>

<template>
  <div
    ref="root"
    class="bento-pricing"
  >
    <div class="bento-pricing-meters">
      <div
        v-for="(m, i) in meters"
        :key="m.label"
        class="bento-pricing-row"
      >
        <span class="bento-pricing-label">{{ m.label }}</span>
        <span class="bento-pricing-track">
          <span
            class="bento-pricing-fill"
            :style="{ width: filled ? `${m.fill}%` : '0%', transitionDelay: `${i * 140}ms` }"
          />
        </span>
      </div>
    </div>

    <div
      class="bento-pricing-total"
      :class="{ 'bento-pricing-total-in': filled }"
    >
      <span class="bento-pricing-total-label">This month</span>
      <span class="bento-pricing-total-amount">CHF {{ amount.toFixed(2) }}</span>
    </div>
  </div>
</template>

<style scoped>
.bento-pricing {
  min-height: 130px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 18px;
  padding: 20px 28px 8px;
}

.bento-pricing-meters {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.bento-pricing-row {
  display: flex;
  align-items: center;
  gap: 14px;
}

.bento-pricing-label {
  width: 64px;
  flex-shrink: 0;
  font-family: var(--font-mono);
  font-size: 12px;
  letter-spacing: 0.04em;
  color: var(--ui-text-muted);
}

.bento-pricing-track {
  flex: 1;
  height: 8px;
  border-radius: 999px;
  background: var(--ui-bg-muted);
  overflow: hidden;
}

.bento-pricing-fill {
  display: block;
  height: 100%;
  width: 0;
  border-radius: 999px;
  background: var(--bento-accent);
  transition: width 0.9s cubic-bezier(0.16, 1, 0.3, 1);
}

.bento-pricing-total {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  opacity: 0;
  transform: translateY(6px);
  transition: opacity 0.5s ease 0.4s, transform 0.5s cubic-bezier(0.16, 1, 0.3, 1) 0.4s;
}

.bento-pricing-total-in {
  opacity: 1;
  transform: translateY(0);
}

.bento-pricing-total-label {
  font-family: var(--font-mono);
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--ui-text-muted);
}

.bento-pricing-total-amount {
  font-family: var(--font-display);
  font-size: 1.75rem;
  line-height: 1;
  color: var(--bento-accent);
  font-variant-numeric: tabular-nums;
}

@media (prefers-reduced-motion: reduce) {
  .bento-pricing-fill,
  .bento-pricing-total {
    transition: none;
  }
}
</style>
