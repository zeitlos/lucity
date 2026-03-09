<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';

const root = ref<HTMLElement | null>(null);
const paperSlide = ref<HTMLElement | null>(null);
const offset = ref(-999); // starts off-screen (real value set after mount)
const printing = ref(false);
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
  { type: 'justify',      left: 'Date: Mar 2026', right: '' },
  { type: 'justify',      left: 'Plan: Hobby', right: '5.00' },
  { type: 'dash' },
  { type: 'line' },
  { type: 'label',        text: 'Resources (Eco):' },
  { type: 'item',         desc: '0.25 vCPU @ €20/core', amount: '5.00' },
  { type: 'item',         desc: '256 MB    @ €10/GB',   amount: '2.56' },
  { type: 'item',         desc: '1 GB Disk @ €0.10/GB', amount: '0.10' },
  { type: 'item',         desc: '0.5 GB    @ €0.02/GB', amount: '0.01' },
  { type: 'subtotal-dash' },
  { type: 'justify',      left: 'Subtotal:', right: '7.67' },
  { type: 'justify',      left: 'Credits:', right: '-5.00', highlight: true },
  { type: 'total-dash' },
  { type: 'total',        left: 'TOTAL:         €', right: '2.67' },
  { type: 'separator' },
  { type: 'line' },
  { type: 'center',       text: 'Thank you for shipping!' },
  { type: 'center-muted', text: 'No lock-in. Ever.' },
];

const STEP = 20; // pixels per feed step

function randomPause(): number {
  return 20 + Math.random() * 180; // 20–200ms
}

async function startPrinting() {
  if (printing.value) return;
  printing.value = true;

  const el = paperSlide.value;
  if (!el) return;

  const totalHeight = el.scrollHeight;
  offset.value = -totalHeight;

  // Step through in ~20px increments
  while (offset.value < 0) {
    await new Promise(r => setTimeout(r, randomPause()));
    offset.value = Math.min(offset.value + STEP, 0);
  }
}

onMounted(() => {
  if (!root.value || !paperSlide.value) return;

  // Set initial offset to hide paper above mask
  offset.value = -paperSlide.value.scrollHeight;

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
  >
    <!-- Printer body at top -->
    <div class="printer-body">
      <div class="printer-slit" />
    </div>

    <!-- Paper mask — clips everything, only shows what's below the slit -->
    <div class="paper-mask">
      <div
        ref="paperSlide"
        class="paper-slide"
        :style="{ transform: `perspective(600px) rotateX(-6deg) translateY(${offset}px)` }"
      >
        <div class="receipt-paper">
          <div
            v-for="(line, i) in lines"
            :key="i"
            class="receipt-line"
            :class="[
              `receipt-${line.type}`,
              { 'receipt-highlight': line.highlight },
            ]"
          >
            <template v-if="line.type === 'separator'" />
            <template v-else-if="line.type === 'dash'" />

            <template v-else-if="line.type === 'center' || line.type === 'center-muted'">
              {{ line.text }}
            </template>

            <template v-else-if="line.type === 'justify' || line.type === 'total'">
              <span>{{ line.left }}</span>
              <span>{{ line.right }}</span>
            </template>

            <template v-else-if="line.type === 'item'">
              <span class="receipt-item-desc">{{ line.desc }}</span>
              <span>{{ line.amount }}</span>
            </template>

            <template v-else-if="line.type === 'subtotal-dash' || line.type === 'total-dash'">
              <span />
              <span :class="line.type === 'total-dash' ? 'rule-double' : 'rule-single'" />
            </template>

            <template v-else>
              {{ line.text }}
            </template>
          </div>
        </div>

        <!-- Zigzag torn edge — trailing edge, last to emerge -->
        <div class="receipt-tear" />
      </div>
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

/* ---- Printer body (at top) ---- */

.printer-body {
  --printer-bg: oklch(0.28 0.01 55);
  width: 100%;
  max-width: 380px;
  height: 18px;
  background: linear-gradient(180deg,
    var(--printer-bg) 0%,
    oklch(0.34 0.01 55) 100%
  );
  border-radius: 10px 10px 0 0;
  position: relative;
  z-index: 2;
  box-shadow:
    inset 0 1px 0 oklch(0.42 0.01 55 / 0.4),
    0 2px 8px oklch(0 0 0 / 0.2);
}

.printer-slit {
  position: absolute;
  bottom: -1px;
  left: 8%;
  right: 8%;
  height: 3px;
  background: oklch(0.10 0.005 55);
  border-radius: 0 0 3px 3px;
  box-shadow:
    inset 0 1px 2px oklch(0 0 0 / 0.6),
    0 1px 3px oklch(0 0 0 / 0.3);
}

/* ---- Paper mask — clips to reveal area below slit ---- */

.paper-mask {
  width: 100%;
  max-width: 380px;
  overflow: hidden;
}

/* ---- Paper slide — moves downward in steps ---- */

.paper-slide {
  transform-origin: top center;
}

/* ---- Torn edge (leading edge — first to emerge) ---- */

.receipt-tear {
  width: 100%;
  max-width: 340px;
  margin: 0 auto;
  height: 12px;
  background: oklch(0.97 0.005 80);
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

/* ---- Lines ---- */

.receipt-line {
  white-space: pre;
  min-height: 1.7em;
}

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

/* ---- Reduced motion ---- */

@media (prefers-reduced-motion: reduce) {
  .paper-slide {
    transform: none !important;
  }
}
</style>
