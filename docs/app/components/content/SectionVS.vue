<script setup lang="ts">
import { ref, watch } from 'vue';
import { useBentoVisible } from './useBentoVisible';

const root = ref<HTMLElement | null>(null);
const visible = useBentoVisible(root, 0.2);
const phase = ref(0);

watch(visible, (v) => {
  if (!v) return;
  phase.value = 1; // sides slide in
  setTimeout(() => { phase.value = 2; }, 400); // lightning draws
  setTimeout(() => { phase.value = 3; }, 700); // VS bounces in
});
</script>

<template>
  <div
    ref="root"
    class="vs-arena"
  >
    <!-- Left side: Self-host -->
    <div
      class="vs-side vs-side-left"
      :class="{ 'vs-side-visible': phase >= 1 }"
    >
      <div class="vs-bg vs-bg-left" />
      <div class="vs-tint vs-tint-left" />
      <div class="vs-gradient vs-gradient-left" />
      <div class="vs-content">
        <h3 class="vs-title">
          Self-host it
        </h3>
        <p class="vs-desc">
          Run Lucity on your own Kubernetes cluster. One Helm install, full control. Your infrastructure, your rules.
        </p>
        <UButton
          to="/getting-started/quick-start"
          color="white"
          variant="solid"
          icon="i-lucide-server"
          trailing-icon="i-lucide-arrow-right"
        >
          Quick Start
        </UButton>
      </div>
    </div>

    <!-- Center VS divider -->
    <div class="vs-divider">
      <!-- Lightning bolt SVG -->
      <svg
        class="vs-lightning"
        :class="{ 'vs-lightning-visible': phase >= 2 }"
        viewBox="0 0 40 400"
        preserveAspectRatio="none"
      >
        <path
          d="M20 0 L26 60 L14 80 L28 140 L10 170 L30 230 L12 260 L26 320 L16 350 L20 400"
          fill="none"
          stroke="url(#lightning-gradient)"
          stroke-width="3"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="vs-bolt"
        />
        <!-- Glow layer -->
        <path
          d="M20 0 L26 60 L14 80 L28 140 L10 170 L30 230 L12 260 L26 320 L16 350 L20 400"
          fill="none"
          stroke="url(#lightning-gradient)"
          stroke-width="8"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="vs-bolt-glow"
        />
        <defs>
          <linearGradient
            id="lightning-gradient"
            x1="0"
            y1="0"
            x2="0"
            y2="1"
          >
            <stop
              offset="0%"
              stop-color="oklch(0.95 0.15 85)"
            />
            <stop
              offset="50%"
              stop-color="oklch(0.90 0.20 85)"
            />
            <stop
              offset="100%"
              stop-color="oklch(0.95 0.15 85)"
            />
          </linearGradient>
        </defs>
      </svg>

      <!-- Horizontal lightning bolt (mobile only) -->
      <svg
        class="vs-lightning-h"
        :class="{ 'vs-lightning-visible': phase >= 2 }"
        viewBox="0 0 400 40"
        preserveAspectRatio="none"
      >
        <path
          d="M0 20 L60 14 L80 26 L140 12 L170 30 L230 10 L260 28 L320 14 L350 24 L400 20"
          fill="none"
          stroke="url(#lightning-gradient-h)"
          stroke-width="3"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="vs-bolt"
        />
        <path
          d="M0 20 L60 14 L80 26 L140 12 L170 30 L230 10 L260 28 L320 14 L350 24 L400 20"
          fill="none"
          stroke="url(#lightning-gradient-h)"
          stroke-width="8"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="vs-bolt-glow"
        />
        <defs>
          <linearGradient
            id="lightning-gradient-h"
            x1="0"
            y1="0"
            x2="1"
            y2="0"
          >
            <stop
              offset="0%"
              stop-color="oklch(0.95 0.15 85)"
            />
            <stop
              offset="50%"
              stop-color="oklch(0.90 0.20 85)"
            />
            <stop
              offset="100%"
              stop-color="oklch(0.95 0.15 85)"
            />
          </linearGradient>
        </defs>
      </svg>

      <!-- VS badge -->
      <div
        class="vs-badge"
        :class="{ 'vs-badge-visible': phase >= 3 }"
      >
        VS
      </div>
    </div>

    <!-- Right side: Cloud -->
    <div
      class="vs-side vs-side-right"
      :class="{ 'vs-side-visible': phase >= 1 }"
    >
      <div class="vs-bg vs-bg-right" />
      <div class="vs-tint vs-tint-right" />
      <div class="vs-gradient vs-gradient-right" />
      <div class="vs-content">
        <h3 class="vs-title">
          Or let us run it
        </h3>
        <p class="vs-desc">
          Lucity Cloud is the managed version. Same open-source platform, hosted in Switzerland, zero infrastructure to maintain.
        </p>
        <UButton
          to="https://lucity.cloud/app/login"
          color="white"
          variant="solid"
          icon="i-lucide-cloud"
          trailing-icon="i-lucide-arrow-right"
        >
          Start for free
        </UButton>
      </div>
    </div>
  </div>
</template>

<style scoped>
.vs-arena {
  position: relative;
  display: flex;
  border-radius: 20px;
  overflow: hidden;
  min-height: 520px;
  isolation: isolate;
  box-shadow:
    0 8px 60px oklch(0 0 0 / 0.25),
    0 2px 20px oklch(0 0 0 / 0.15);
}

/* ── Sides ── */
.vs-side {
  position: relative;
  flex: 1;
  display: flex;
  align-items: flex-end;
  overflow: hidden;
  opacity: 0;
  cursor: pointer;
  transition: opacity 0.5s ease, transform 0.6s cubic-bezier(0.16, 1, 0.3, 1);
}

.vs-side-visible:hover .vs-bg {
  transform: scale(1.04);
  filter: brightness(1.1);
}

.vs-side-visible:hover .vs-tint {
  opacity: 0.7;
}

.vs-side-visible:hover .vs-content {
  transform: translateY(-3px);
}

.vs-side-left {
  transform: translateX(-30px);
}

.vs-side-right {
  transform: translateX(30px);
}

.vs-side-visible {
  opacity: 1;
  transform: translateX(0);
}

/* ── Background images ── */
.vs-bg {
  position: absolute;
  inset: 0;
  background-size: cover;
  background-position: center;
  z-index: 0;
  transition: transform 0.5s cubic-bezier(0.16, 1, 0.3, 1), filter 0.5s ease;
}

.vs-bg-left {
  background-image: url('/img/mountain_city_night.webp');
}

.vs-bg-right {
  background-image: url('/img/mountain_city.webp');
}

/* ── Color tints ── */
.vs-tint {
  position: absolute;
  inset: 0;
  z-index: 1;
  pointer-events: none;
  transition: opacity 0.5s ease;
}

.vs-tint-left {
  background: oklch(0.25 0.12 240 / 0.6);
}

.vs-tint-right {
  background: oklch(0.30 0.14 25 / 0.55);
}

/* ── Bottom gradient for text readability ── */
.vs-gradient {
  position: absolute;
  inset: 0;
  z-index: 2;
  pointer-events: none;
}

.vs-gradient-left {
  background: linear-gradient(
    to top,
    oklch(0.15 0.06 240 / 0.97) 0%,
    oklch(0.15 0.06 240 / 0.8) 40%,
    transparent 75%
  );
}

.vs-gradient-right {
  background: linear-gradient(
    to top,
    oklch(0.18 0.06 25 / 0.97) 0%,
    oklch(0.18 0.06 25 / 0.8) 40%,
    transparent 75%
  );
}

/* ── Content ── */
.vs-content {
  position: relative;
  z-index: 3;
  padding: 40px 36px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
  transition: transform 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.vs-title {
  font-family: var(--font-serif);
  font-weight: normal;
  font-size: 2.8rem;
  line-height: 1.1;
  color: white;
  margin: 0;
  text-shadow: 0 2px 16px oklch(0 0 0 / 0.6);
}

.vs-desc {
  font-size: 1.1rem;
  line-height: 1.6;
  color: oklch(0.90 0.02 80);
  margin: 0;
  max-width: 400px;
  text-shadow: 0 1px 8px oklch(0 0 0 / 0.5);
}

/* ── Arcade buttons ── */
.vs-content :deep(a),
.vs-content :deep(button) {
  width: fit-content !important;
  color: white !important;
  font-weight: 600 !important;
  letter-spacing: 0.02em;
  border: none !important;
  border-top: 1.5px solid oklch(1 0 0 / 0.4) !important;
  border-radius: 10px !important;
  padding: 10px 20px !important;
  box-shadow:
    0 4px 12px oklch(0 0 0 / 0.5),
    0 1px 0 oklch(1 0 0 / 0.1) inset,
    0 -2px 6px oklch(0 0 0 / 0.3) inset;
  backdrop-filter: blur(6px);
  transition: transform 0.15s ease, box-shadow 0.15s ease, filter 0.15s ease;
}

.vs-content :deep(a:hover),
.vs-content :deep(button:hover) {
  transform: translateY(-1px);
  filter: brightness(1.15);
  box-shadow:
    0 6px 20px oklch(0 0 0 / 0.5),
    0 1px 0 oklch(1 0 0 / 0.15) inset,
    0 -2px 6px oklch(0 0 0 / 0.3) inset;
}

.vs-content :deep(a:active),
.vs-content :deep(button:active) {
  transform: translateY(1px);
  box-shadow:
    0 2px 6px oklch(0 0 0 / 0.5),
    0 1px 0 oklch(1 0 0 / 0.05) inset,
    0 -1px 4px oklch(0 0 0 / 0.4) inset;
}

/* Left side: blue arcade button */
.vs-side-left .vs-content :deep(a),
.vs-side-left .vs-content :deep(button) {
  background: linear-gradient(
    to bottom,
    oklch(0.45 0.14 240) 0%,
    oklch(0.32 0.12 240) 100%
  ) !important;
}

/* Right side: red/orange arcade button */
.vs-side-right .vs-content :deep(a),
.vs-side-right .vs-content :deep(button) {
  background: linear-gradient(
    to bottom,
    oklch(0.50 0.18 25) 0%,
    oklch(0.36 0.15 25) 100%
  ) !important;
}

/* ── Center divider ── */
.vs-divider {
  position: absolute;
  left: 50%;
  top: 0;
  bottom: 0;
  width: 40px;
  transform: translateX(-50%);
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;
}

/* ── Lightning ── */
.vs-lightning,
.vs-lightning-h {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  filter: drop-shadow(0 0 6px oklch(0.90 0.18 85 / 0.6));
  transition: opacity 0.3s ease;
}

/* Desktop: vertical visible, horizontal hidden */
.vs-lightning-h {
  display: none;
}

.vs-lightning-visible {
  opacity: 1;
}

.vs-bolt {
  stroke-dasharray: 800;
  stroke-dashoffset: 800;
}

.vs-lightning-visible .vs-bolt {
  animation: vs-draw-bolt 0.5s ease forwards;
}

.vs-bolt-glow {
  opacity: 0.3;
  stroke-dasharray: 800;
  stroke-dashoffset: 800;
}

.vs-lightning-visible .vs-bolt-glow {
  animation: vs-draw-bolt 0.5s ease forwards, vs-pulse 2s ease-in-out 0.5s infinite;
}

@keyframes vs-draw-bolt {
  to { stroke-dashoffset: 0; }
}

@keyframes vs-pulse {
  0%, 100% { opacity: 0.3; }
  50% { opacity: 0.6; }
}

/* ── VS badge ── */
.vs-badge {
  position: relative;
  z-index: 11;
  font-family: var(--font-sans);
  font-size: 3.4rem;
  font-weight: 900;
  letter-spacing: 0.02em;
  line-height: 1;
  /* Fiery red-to-yellow gradient text */
  background: linear-gradient(
    to top,
    oklch(0.55 0.25 25) 0%,
    oklch(0.65 0.28 30) 30%,
    oklch(0.75 0.22 45) 60%,
    oklch(0.88 0.18 80) 100%
  );
  background-clip: text;
  -webkit-background-clip: text;
  color: transparent;
  -webkit-text-stroke: 1.5px oklch(0.30 0.12 25 / 0.6);
  paint-order: stroke fill;
  filter:
    drop-shadow(0 0 12px oklch(0.65 0.25 30 / 0.7))
    drop-shadow(0 0 30px oklch(0.60 0.22 25 / 0.4))
    drop-shadow(0 2px 3px oklch(0 0 0 / 0.6));
  opacity: 0;
  transform: scale(0);
}

.vs-badge-visible {
  animation: vs-badge-in 0.4s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

@keyframes vs-badge-in {
  0% { opacity: 0; transform: scale(0); }
  70% { opacity: 1; transform: scale(1.2); }
  100% { opacity: 1; transform: scale(1); }
}

/* ── Mobile: stack vertically ── */
@media (max-width: 639px) {
  .vs-arena {
    flex-direction: column;
    min-height: auto;
    border-radius: 16px;
  }

  .vs-side {
    min-height: 280px;
  }

  .vs-side-left {
    transform: translateY(-20px);
    border-radius: 16px 16px 0 0;
  }

  .vs-side-right {
    transform: translateY(20px);
    border-radius: 0 0 16px 16px;
  }

  .vs-side-visible {
    transform: translateY(0);
  }

  .vs-divider {
    position: relative;
    left: auto;
    top: auto;
    bottom: auto;
    width: 100%;
    height: 48px;
    transform: none;
    background: linear-gradient(
      to bottom,
      oklch(0.18 0.08 240) 0%,
      oklch(0.14 0.04 55) 50%,
      oklch(0.18 0.08 25) 100%
    );
  }

  .vs-lightning {
    display: none;
  }

  .vs-lightning-h {
    display: block;
  }

  .vs-badge {
    font-size: 1.6rem;
  }
}
</style>
