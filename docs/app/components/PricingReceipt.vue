<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';

const root = ref<HTMLElement | null>(null);
const lineCount = ref(0);
let observer: IntersectionObserver | null = null;
let fallbackTimer: ReturnType<typeof setTimeout> | null = null;

interface ReceiptLine {
  type: 'separator' | 'dash' | 'center' | 'center-muted' | 'line' | 'label' | 'justify' | 'item' | 'subtotal-dash' | 'total-dash' | 'total';
  text?: string;
  left?: string;
  right?: string;
  desc?: string;
  amount?: string;
  highlight?: boolean;
}

const lines: ReceiptLine[] = [
  { type: 'separator', text: '================================' },
  { type: 'center',    text: 'LUCITY CLOUD' },
  { type: 'center-muted', text: 'Monthly Usage Receipt' },
  { type: 'separator', text: '================================' },
  { type: 'justify',   left: 'Date: Feb 2026', right: '' },
  { type: 'justify',   left: 'Plan: Starter', right: '5.00' },
  { type: 'dash',      text: '--------------------------------' },
  { type: 'line',      text: '' },
  { type: 'label',     text: 'Resources:' },
  { type: 'item',      desc: '1x vCPU    @ CHF 10.00', amount: '10.00' },
  { type: 'item',      desc: '512 MB RAM @ CHF  5/GB', amount: '2.50' },
  { type: 'item',      desc: '10 GB Disk @ CHF  0.10', amount: '1.00' },
  { type: 'item',      desc: '5 GB Egrs  @ CHF  0.02', amount: '0.10' },
  { type: 'subtotal-dash' },
  { type: 'justify',   left: 'Subtotal:', right: '13.60' },
  { type: 'justify',   left: 'Credits:', right: '-5.00', highlight: true },
  { type: 'total-dash' },
  { type: 'total',     left: 'TOTAL:       CHF', right: '8.60' },
  { type: 'separator', text: '================================' },
  { type: 'line',      text: '' },
  { type: 'center',    text: 'Thank you for shipping!' },
  { type: 'center-muted', text: 'No lock-in. Ever.' },
];

function startPrinting() {
  if (lineCount.value > 0) return;
  lines.forEach((_, i) => {
    setTimeout(() => { lineCount.value = i + 1; }, 150 + i * 80);
  });
}

onMounted(() => {
  if (!root.value) return;

  // Fallback: start printing after 2s even if IntersectionObserver doesn't fire
  fallbackTimer = setTimeout(startPrinting, 2000);

  if (typeof IntersectionObserver === 'undefined') {
    startPrinting();
    return;
  }

  observer = new IntersectionObserver(
    ([entry]) => {
      if (entry.isIntersecting) {
        if (fallbackTimer) clearTimeout(fallbackTimer);
        startPrinting();
        observer?.disconnect();
      }
    },
    { threshold: 0.3 },
  );
  observer.observe(root.value);
});

onUnmounted(() => {
  observer?.disconnect();
  if (fallbackTimer) clearTimeout(fallbackTimer);
});
</script>

<template>
  <div
    ref="root"
    class="receipt-container"
  >
    <div class="receipt-paper">
      <div
        v-for="(line, i) in lines"
        :key="i"
        class="receipt-line"
        :class="[
          `receipt-${line.type}`,
          {
            'receipt-line-visible': lineCount > i,
            'receipt-highlight': line.highlight,
          },
        ]"
      >
        <!-- separator / dash -->
        <template v-if="line.type === 'separator' || line.type === 'dash'">
          {{ line.text }}
        </template>

        <!-- center / center-muted -->
        <template v-else-if="line.type === 'center' || line.type === 'center-muted'">
          {{ line.text }}
        </template>

        <!-- justify / total (left + right) -->
        <template v-else-if="line.type === 'justify' || line.type === 'total'">
          <span>{{ line.left }}</span>
          <span>{{ line.right }}</span>
        </template>

        <!-- item (desc + amount) -->
        <template v-else-if="line.type === 'item'">
          <span class="receipt-item-desc">{{ line.desc }}</span>
          <span>{{ line.amount }}</span>
        </template>

        <!-- subtotal-dash / total-dash (right-aligned rule) -->
        <template v-else-if="line.type === 'subtotal-dash' || line.type === 'total-dash'">
          <span />
          <span>{{ line.type === 'total-dash' ? '======' : '------' }}</span>
        </template>

        <!-- label / line (plain text) -->
        <template v-else>
          {{ line.text }}
        </template>
      </div>

      <!-- Zigzag torn edge -->
      <div
        class="receipt-tear"
        :class="{ 'receipt-tear-visible': lineCount >= lines.length }"
      />
    </div>
  </div>
</template>

<style scoped>
.receipt-container {
  display: flex;
  justify-content: center;
  padding: 0 16px;
}

.receipt-paper {
  --receipt-bg: oklch(0.97 0.005 80);
  --receipt-text: oklch(0.25 0.02 55);
  --receipt-muted: oklch(0.50 0.03 55);

  position: relative;
  background: var(--receipt-bg);
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.7;
  color: var(--receipt-text);
  padding: 28px 24px 32px;
  width: 100%;
  max-width: 340px;
  border-radius: 4px 4px 0 0;
  box-shadow:
    0 2px 8px oklch(0.3 0.02 55 / 0.08),
    0 1px 2px oklch(0.3 0.02 55 / 0.04);
}

/* Receipt stays paper-colored in dark mode — like a real receipt */
.dark .receipt-paper {
  --receipt-bg: oklch(0.94 0.005 80);
  --receipt-text: oklch(0.22 0.02 55);
  --receipt-muted: oklch(0.42 0.03 55);
  box-shadow:
    0 4px 16px oklch(0 0 0 / 0.3),
    0 2px 4px oklch(0 0 0 / 0.2);
}

/* ---- Line animation ---- */

.receipt-line {
  opacity: 0;
  transform: translateY(-3px);
  white-space: pre;
  min-height: 1.7em;
}

.receipt-line-visible {
  animation: receipt-line-in 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

@keyframes receipt-line-in {
  from {
    opacity: 0;
    transform: translateY(-3px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* ---- Line types ---- */

.receipt-separator,
.receipt-dash {
  color: var(--receipt-muted);
  letter-spacing: -0.5px;
}

.receipt-center {
  text-align: center;
  font-weight: 600;
}

.receipt-center-muted {
  text-align: center;
  color: var(--receipt-muted);
  font-size: 11px;
}

.receipt-label {
  font-weight: 500;
  margin-top: 2px;
}

.receipt-justify,
.receipt-total {
  display: flex;
  justify-content: space-between;
  gap: 8px;
}

.receipt-item {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  padding-left: 2ch;
}

.receipt-item-desc {
  color: var(--receipt-muted);
}

.receipt-subtotal-dash,
.receipt-total-dash {
  display: flex;
  justify-content: flex-end;
  color: var(--receipt-muted);
  padding-right: 0;
}

.receipt-total {
  font-weight: 700;
  font-size: 13px;
}

.receipt-highlight span:last-child {
  color: var(--ui-primary);
  font-weight: 500;
}

.receipt-line.receipt-line:is([class*="line"]) {
  min-height: 0.85em;
}

/* ---- Torn edge ---- */

.receipt-tear {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 10px;
  background: var(--receipt-bg);
  transform: translateY(100%);
  opacity: 0;
  clip-path: polygon(
    0% 0%, 2.5% 100%, 5% 0%, 7.5% 100%, 10% 0%,
    12.5% 100%, 15% 0%, 17.5% 100%, 20% 0%, 22.5% 100%,
    25% 0%, 27.5% 100%, 30% 0%, 32.5% 100%, 35% 0%,
    37.5% 100%, 40% 0%, 42.5% 100%, 45% 0%, 47.5% 100%,
    50% 0%, 52.5% 100%, 55% 0%, 57.5% 100%, 60% 0%,
    62.5% 100%, 65% 0%, 67.5% 100%, 70% 0%, 72.5% 100%,
    75% 0%, 77.5% 100%, 80% 0%, 82.5% 100%, 85% 0%,
    87.5% 100%, 90% 0%, 92.5% 100%, 95% 0%, 97.5% 100%, 100% 0%
  );
}

.receipt-tear-visible {
  animation: receipt-tear-in 0.4s ease forwards;
}

@keyframes receipt-tear-in {
  from { opacity: 0; }
  to   { opacity: 1; }
}

/* ---- Reduced motion ---- */

@media (prefers-reduced-motion: reduce) {
  .receipt-line {
    opacity: 1 !important;
    transform: none !important;
    animation: none !important;
  }
  .receipt-tear {
    opacity: 1 !important;
    animation: none !important;
  }
}
</style>
