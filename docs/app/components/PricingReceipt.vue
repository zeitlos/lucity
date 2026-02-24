<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';

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
  { type: 'separator' },
  { type: 'center',       text: 'LUCITY CLOUD' },
  { type: 'center-muted', text: 'Monthly Usage Receipt' },
  { type: 'separator' },
  { type: 'justify',      left: 'Date: Feb 2026', right: '' },
  { type: 'justify',      left: 'Plan: Starter', right: '5.00' },
  { type: 'dash' },
  { type: 'line' },
  { type: 'label',        text: 'Resources:' },
  { type: 'item',         desc: '1x vCPU    @ CHF 10.00', amount: '10.00' },
  { type: 'item',         desc: '512 MB RAM @ CHF  5/GB', amount: '2.50' },
  { type: 'item',         desc: '10 GB Disk @ CHF  0.10', amount: '1.00' },
  { type: 'item',         desc: '5 GB Egrs  @ CHF  0.02', amount: '0.10' },
  { type: 'subtotal-dash' },
  { type: 'justify',      left: 'Subtotal:', right: '13.60' },
  { type: 'justify',      left: 'Credits:', right: '-5.00', highlight: true },
  { type: 'total-dash' },
  { type: 'total',        left: 'TOTAL:       CHF', right: '8.60' },
  { type: 'separator' },
  { type: 'line' },
  { type: 'center',       text: 'Thank you for shipping!' },
  { type: 'center-muted', text: 'No lock-in. Ever.' },
];

/* ---- Height calculations ---- */

const LINE_H = 20.4;   // 12px × 1.7 line-height
const SEP_H = 12;      // separator/dash lines are shorter
const PAD_TOP = 28;
const PAD_BOTTOM = 32;
const TEAR_H = 12;
const PRINTER_BODY_H = 18;

function heightForLine(type: string): number {
  if (type === 'separator' || type === 'dash' || type === 'subtotal-dash' || type === 'total-dash') return SEP_H;
  if (type === 'line') return LINE_H * 0.5;
  return LINE_H;
}

// Pre-calculate total height so the container is fixed
const totalPaperH = PAD_TOP + lines.reduce((sum, l) => sum + heightForLine(l.type), 0) + PAD_BOTTOM + TEAR_H;
const reservedHeight = `${PRINTER_BODY_H + totalPaperH}px`;

// Current feed height — grows per line (instant, no transition)
const feedHeight = computed(() => {
  if (lineCount.value === 0) return '0px';
  let h = PAD_TOP;
  for (let i = 0; i < lineCount.value && i < lines.length; i++) {
    h += heightForLine(lines[i].type);
  }
  // Include tear + bottom padding when fully printed
  h += lineCount.value >= lines.length ? PAD_BOTTOM + TEAR_H : 8;
  return `${h}px`;
});

/* ---- Printing trigger ---- */

function startPrinting() {
  if (lineCount.value > 0) return;
  lines.forEach((_, i) => {
    setTimeout(() => { lineCount.value = i + 1; }, 300 + i * 100);
  });
}

onMounted(() => {
  if (!root.value) return;

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
    { threshold: 0.15 },
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
    class="receipt-printer"
    :style="{ minHeight: reservedHeight }"
  >
    <!-- Paper area — fixed-height wrapper, paper grows upward inside -->
    <div
      class="paper-area"
      :style="{ height: `${totalPaperH}px` }"
    >
      <div
        class="paper-feed"
        :style="{ maxHeight: feedHeight }"
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
            <!-- separator (full-width solid rule) -->
            <template v-if="line.type === 'separator'" />

            <!-- dash (full-width dashed rule) -->
            <template v-else-if="line.type === 'dash'" />

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

            <!-- subtotal-dash / total-dash (right-aligned partial rule) -->
            <template v-else-if="line.type === 'subtotal-dash' || line.type === 'total-dash'">
              <span />
              <span :class="line.type === 'total-dash' ? 'rule-double' : 'rule-single'" />
            </template>

            <!-- label / line (plain text) -->
            <template v-else>
              {{ line.text }}
            </template>
          </div>
        </div>

        <!-- Zigzag torn edge — right above the printer slit -->
        <div
          class="receipt-tear"
          :class="{ 'receipt-tear-visible': lineCount >= lines.length }"
        />
      </div>
    </div>

    <!-- Printer body at bottom -->
    <div class="printer-body">
      <div class="printer-slit" />
    </div>
  </div>
</template>

<style scoped>
.receipt-printer {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0 16px;
}

/* ---- Paper area (fixed-height wrapper) ---- */

.paper-area {
  position: relative;
  width: 100%;
  max-width: 380px;
}

/* ---- Paper feed — grows upward from the printer slit ---- */

.paper-feed {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  max-height: 0;
  overflow: hidden;
  /* No transition — instant height snaps for old-school printer feel */
}

/* ---- Printer body (at bottom) ---- */

.printer-body {
  --printer-bg: oklch(0.28 0.01 55);
  width: 100%;
  max-width: 380px;
  height: 18px;
  background: linear-gradient(0deg,
    oklch(0.34 0.01 55) 0%,
    var(--printer-bg) 100%
  );
  border-radius: 0 0 10px 10px;
  position: relative;
  z-index: 2;
  box-shadow:
    inset 0 -1px 0 oklch(0.42 0.01 55 / 0.4),
    0 2px 8px oklch(0 0 0 / 0.2);
}

.printer-slit {
  position: absolute;
  top: -1px;
  left: 8%;
  right: 8%;
  height: 3px;
  background: oklch(0.10 0.005 55);
  border-radius: 3px 3px 0 0;
  box-shadow:
    inset 0 -1px 2px oklch(0 0 0 / 0.6),
    0 -1px 3px oklch(0 0 0 / 0.3);
}

/* ---- Receipt paper ---- */

.receipt-paper {
  --receipt-bg: oklch(0.97 0.005 80);
  --receipt-text: oklch(0.25 0.02 55);
  --receipt-muted: oklch(0.50 0.03 55);

  background: var(--receipt-bg);
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.7;
  color: var(--receipt-text);
  padding: 28px 24px 32px;
  width: 100%;
  max-width: 340px;
  margin: 0 auto;
  box-shadow:
    2px 0 6px oklch(0.3 0.02 55 / 0.06),
    -2px 0 6px oklch(0.3 0.02 55 / 0.06);
}

.dark .receipt-paper {
  --receipt-bg: oklch(0.94 0.005 80);
  --receipt-text: oklch(0.22 0.02 55);
  --receipt-muted: oklch(0.42 0.03 55);
  box-shadow:
    2px 0 12px oklch(0 0 0 / 0.25),
    -2px 0 12px oklch(0 0 0 / 0.25);
}

/* ---- Line animation (instant snap — dot-matrix style) ---- */

.receipt-line {
  opacity: 0;
  white-space: pre;
  min-height: 1.7em;
}

.receipt-line-visible {
  opacity: 1;
}

/* ---- Separator lines (full-width CSS rules) ---- */

.receipt-separator {
  min-height: 0;
  height: 0;
  padding: 4px 0;
  border-top: 1.5px solid var(--receipt-text);
}

.receipt-dash {
  min-height: 0;
  height: 0;
  padding: 4px 0;
  border-top: 1px dashed var(--receipt-muted);
}

/* ---- Line types ---- */

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
  min-height: 0;
  height: 0;
  padding: 4px 0;
}

.rule-single {
  width: 60px;
  border-top: 1px solid var(--receipt-muted);
}

.rule-double {
  width: 60px;
  border-top: 2.5px double var(--receipt-text);
}

.receipt-total {
  font-weight: 700;
  font-size: 13px;
}

.receipt-highlight span:last-child {
  color: var(--ui-primary);
  font-weight: 500;
}

/* ---- Torn edge ---- */

.receipt-tear {
  width: 100%;
  max-width: 340px;
  margin: 0 auto;
  height: 12px;
  background: oklch(0.97 0.005 80);
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

.dark .receipt-tear {
  background: oklch(0.94 0.005 80);
}

.receipt-tear-visible {
  opacity: 1;
}

/* ---- Reduced motion ---- */

@media (prefers-reduced-motion: reduce) {
  .paper-feed {
    position: static !important;
    max-height: none !important;
    overflow: visible !important;
  }
  .receipt-line {
    opacity: 1 !important;
  }
  .receipt-tear {
    opacity: 1 !important;
  }
}
</style>
