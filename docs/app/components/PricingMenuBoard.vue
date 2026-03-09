<script setup lang="ts">
interface ResourcePrice {
  label: string;
  value: string;
  unit: string;
}

interface Tier {
  name: string;
  detail: string;
  prices: ResourcePrice[];
}

const tiers: Tier[] = [
  {
    name: 'Eco',
    detail: 'Metered · Shared EU',
    prices: [
      { label: 'vCPU', value: '€20', unit: '/ core / mo' },
      { label: 'Memory', value: '€10', unit: '/ GB / mo' },
      { label: 'Disk', value: '€0.10', unit: '/ GB / mo' },
      { label: 'Egress', value: '€0.02', unit: '/ GB' },
    ],
  },
  {
    name: 'Production',
    detail: 'Predictable · Dedicated EU',
    prices: [
      { label: 'vCPU', value: '€50', unit: '/ core / mo' },
      { label: 'Memory', value: '€25', unit: '/ GB / mo' },
      { label: 'Disk', value: '€0.25', unit: '/ GB / mo' },
      { label: 'Egress', value: '€0.05', unit: '/ GB' },
    ],
  },
];
</script>

<template>
  <div class="board-scene">
    <div class="board-anchor">
      <div class="board">
        <div class="board-title">
          Resources
        </div>

        <hr class="board-separator">

        <div class="board-columns">
          <div
            v-for="(tier, i) in tiers"
            :key="i"
            class="board-column"
          >
            <div class="board-tier-header">
              <span class="board-tier-name">{{ tier.name }}</span>
              <span class="board-tier-detail">{{ tier.detail }}</span>
            </div>

            <div
              v-for="(price, j) in tier.prices"
              :key="`${i}-${j}`"
              class="board-row"
            >
              <span class="board-name">{{ price.label }}</span>
              <span class="board-dots" />
              <span class="board-price">{{ price.value }}</span>
              <span class="board-unit">{{ price.unit }}</span>
            </div>
          </div>
        </div>

        <hr class="board-separator">

        <div class="board-wordmark">
          LUCITY
        </div>
        <div class="board-tagline">
          May your deploys be boring.
        </div>
      </div>

      <div class="cable" aria-hidden="true" />
    </div>
  </div>
</template>

<style scoped>
.board-scene {
  display: flex;
  justify-content: center;
  padding: 0 16px;
  margin-bottom: 32px;
}

.board-anchor {
  position: relative;
  max-width: 760px;
  width: 100%;
  overflow: visible;
}

/* ---- Board ---- */

.board {
  position: relative;
  z-index: 2;
  background: oklch(0.12 0.005 55);
  border: 1px solid oklch(0.22 0.01 55);
  border-radius: 10px;
  padding: 20px 24px;
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 1.8;
  color: oklch(0.88 0.02 80);
  display: flex;
  flex-direction: column;
  gap: 2px;
  box-shadow:
    inset 0 1px 0 oklch(0.2 0.01 55 / 0.5),
    inset 0 -1px 0 oklch(0.05 0.005 55),
    0 8px 30px -4px oklch(0 0 0 / 0.5),
    0 2px 6px -1px oklch(0 0 0 / 0.3);
}

/* ---- Backlit glow ---- */

.board::before {
  content: '';
  position: absolute;
  inset: -12px;
  border-radius: 20px;
  background: oklch(0.7 0.16 65 / 0.12);
  filter: blur(25px);
  z-index: -1;
  animation: backlight-flicker 10s ease-in-out infinite;
}

.dark .board::before {
  background: oklch(0.7 0.16 65 / 0.18);
}

/* ---- Title ---- */

.board-title {
  text-transform: uppercase;
  letter-spacing: 0.12em;
  font-size: 11px;
  font-weight: 500;
  color: oklch(0.5 0.01 80);
}

/* ---- Separators ---- */

.board-separator {
  border: none;
  border-top: 1px solid oklch(0.22 0.01 55);
  margin: 6px 0;
}

/* ---- Two-column layout ---- */

.board-columns {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

@media (min-width: 640px) {
  .board-columns {
    flex-direction: row;
    gap: 24px;
  }

  .board-column {
    flex: 1 1 0%;
  }
}

/* ---- Divider between columns on desktop ---- */

@media (min-width: 640px) {
  .board-column + .board-column {
    border-left: 1px solid oklch(0.22 0.01 55);
    padding-left: 24px;
  }
}

/* ---- Tier headers ---- */

.board-tier-header {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 2px;
}

.board-tier-name {
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  font-size: 12px;
  color: oklch(0.82 0.08 65);
  text-shadow: 0 0 12px oklch(0.8 0.1 65 / 0.2);
}

.board-tier-detail {
  font-size: 10px;
  color: oklch(0.45 0.01 80);
  letter-spacing: 0.02em;
}

/* ---- Rows ---- */

.board-row {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.board-name {
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  white-space: nowrap;
  text-shadow: 0 0 10px oklch(0.9 0.05 80 / 0.15);
}

.board-dots {
  flex: 1;
  min-width: 12px;
  border-bottom: 1px dotted oklch(0.28 0.005 55);
  position: relative;
  top: -3px;
}

.board-price {
  font-weight: 600;
  white-space: nowrap;
  text-shadow: 0 0 10px oklch(0.9 0.05 80 / 0.15);
}

.board-unit {
  font-size: 10px;
  color: oklch(0.5 0.01 80);
  white-space: nowrap;
}

/* ---- Wordmark ---- */

.board-wordmark {
  font-family: var(--font-sans);
  font-weight: 900;
  font-size: 2.5rem;
  letter-spacing: 0.12em;
  text-align: center;
  color: oklch(0.9 0.02 80);
  text-shadow: 0 0 20px oklch(0.9 0.08 80 / 0.25);
  margin-top: 4px;
  line-height: 1.2;
}

.board-tagline {
  text-align: center;
  font-size: 10px;
  color: oklch(0.45 0.01 80);
  letter-spacing: 0.04em;
  margin-top: 2px;
}

/* ---- Cable ---- */

.cable {
  display: none;
}

@media (min-width: 768px) {
  .cable {
    display: block;
    position: absolute;
    top: 50%;
    left: 100%;
    transform: translateY(-50%);
    width: 200vw;
    height: 10px;
    border-radius: 2px;
    z-index: 1;

    background: linear-gradient(180deg,
      oklch(0.70 0.006 250) 0%,
      oklch(0.56 0.006 250) 28%,
      oklch(0.78 0.006 250) 48%,
      oklch(0.53 0.006 250) 72%,
      oklch(0.63 0.006 250) 100%
    );

    box-shadow:
      0 1px 3px oklch(0 0 0 / 0.25),
      0 3px 8px oklch(0 0 0 / 0.12),
      inset 0 1px 0 oklch(1 0 0 / 0.12);
  }

  .dark .cable {
    background: linear-gradient(180deg,
      oklch(0.35 0.006 250) 0%,
      oklch(0.25 0.006 250) 28%,
      oklch(0.40 0.006 250) 48%,
      oklch(0.22 0.006 250) 72%,
      oklch(0.30 0.006 250) 100%
    );
    box-shadow:
      0 1px 3px oklch(0 0 0 / 0.4),
      0 3px 8px oklch(0 0 0 / 0.2),
      inset 0 1px 0 oklch(1 0 0 / 0.06);
  }

  .cable::before {
    content: '';
    position: absolute;
    left: -2px;
    top: 50%;
    transform: translateY(-50%);
    width: 6px;
    height: 18px;
    border-radius: 2px;

    background: linear-gradient(180deg,
      oklch(0.68 0.006 250) 0%,
      oklch(0.52 0.006 250) 30%,
      oklch(0.75 0.006 250) 50%,
      oklch(0.50 0.006 250) 70%,
      oklch(0.60 0.006 250) 100%
    );

    box-shadow:
      0 1px 2px oklch(0 0 0 / 0.3),
      inset 0 1px 0 oklch(1 0 0 / 0.1);
  }

  .dark .cable::before {
    background: linear-gradient(180deg,
      oklch(0.32 0.006 250) 0%,
      oklch(0.22 0.006 250) 30%,
      oklch(0.38 0.006 250) 50%,
      oklch(0.20 0.006 250) 70%,
      oklch(0.28 0.006 250) 100%
    );
    box-shadow:
      0 1px 2px oklch(0 0 0 / 0.5),
      inset 0 1px 0 oklch(1 0 0 / 0.04);
  }
}

/* ---- Flicker ---- */

@keyframes backlight-flicker {
  0%, 100% { opacity: 1; }
  15% { opacity: 1; }
  15.3% { opacity: 0.72; }
  15.6% { opacity: 0.88; }
  15.9% { opacity: 0.65; }
  16.3% { opacity: 1; }
  52% { opacity: 1; }
  52.2% { opacity: 0.82; }
  52.5% { opacity: 1; }
  78% { opacity: 1; }
  78.2% { opacity: 0.78; }
  78.5% { opacity: 0.85; }
  78.8% { opacity: 0.68; }
  79.2% { opacity: 0.92; }
  79.6% { opacity: 1; }
}

@media (prefers-reduced-motion: reduce) {
  .board::before {
    animation: none;
  }
}
</style>
