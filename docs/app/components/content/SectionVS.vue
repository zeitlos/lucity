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
          to="/cloud"
          color="white"
          variant="solid"
          trailing-icon="i-lucide-arrow-right"
        >
          Join the waitlist
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
  min-height: 380px;
  isolation: isolate;
}

/* ── Sides ── */
.vs-side {
  position: relative;
  flex: 1;
  display: flex;
  align-items: flex-end;
  overflow: hidden;
  opacity: 0;
  transition: opacity 0.5s ease, transform 0.6s cubic-bezier(0.16, 1, 0.3, 1);
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
    oklch(0.15 0.06 240 / 0.95) 0%,
    oklch(0.15 0.06 240 / 0.7) 35%,
    transparent 70%
  );
}

.vs-gradient-right {
  background: linear-gradient(
    to top,
    oklch(0.18 0.06 25 / 0.95) 0%,
    oklch(0.18 0.06 25 / 0.7) 35%,
    transparent 70%
  );
}

/* ── Content ── */
.vs-content {
  position: relative;
  z-index: 3;
  padding: 32px 28px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
}

.vs-title {
  font-family: var(--font-serif);
  font-weight: normal;
  font-size: 1.6rem;
  color: white;
  margin: 0;
  text-shadow: 0 2px 12px oklch(0 0 0 / 0.5);
}

.vs-desc {
  font-size: 0.9rem;
  line-height: 1.5;
  color: oklch(0.88 0.02 80);
  margin: 0;
  max-width: 340px;
  text-shadow: 0 1px 6px oklch(0 0 0 / 0.4);
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
.vs-lightning {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  filter: drop-shadow(0 0 6px oklch(0.90 0.18 85 / 0.6));
  transition: opacity 0.3s ease;
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
  font-size: 2rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  color: oklch(0.92 0.18 85);
  text-shadow:
    0 0 20px oklch(0.85 0.20 85 / 0.8),
    0 0 40px oklch(0.80 0.18 85 / 0.4),
    0 2px 4px oklch(0 0 0 / 0.5);
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
      to right,
      oklch(0.18 0.08 240) 0%,
      oklch(0.14 0.04 55) 50%,
      oklch(0.18 0.08 25) 100%
    );
  }

  .vs-lightning {
    display: none;
  }

  .vs-badge {
    font-size: 1.6rem;
  }
}
</style>
