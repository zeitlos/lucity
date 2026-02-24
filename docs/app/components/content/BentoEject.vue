<script setup lang="ts">
import { ref, watch, nextTick } from 'vue';
import { useBentoVisible } from './useBentoVisible';

const root = ref<HTMLElement | null>(null);
const rightRef = ref<HTMLElement | null>(null);
const visible = useBentoVisible(root);
const phase = ref<'idle' | 'door' | 'lines' | 'done'>('idle');
const artifactCount = ref(0);

const outputs = [
  { icon: 'i-simple-icons-helm', label: 'Helm chart', path: 'ejected/chart/' },
  { icon: 'i-simple-icons-argo', label: 'ArgoCD apps', path: 'ejected/argocd/' },
  { icon: 'i-lucide-settings', label: 'Env values', path: 'ejected/environments/' },
  { icon: 'i-lucide-file-text', label: 'README', path: 'ejected/README.md' },
];

/* SVG path data — computed from actual DOM positions */
interface ConnectorPath { d: string; endX: number; endY: number; len: number }
const svgHeight = ref(200);
const connectorPaths = ref<ConnectorPath[]>([]);

function computePaths() {
  const container = rightRef.value;
  if (!container) return;

  const cards = container.querySelectorAll('.bento-output');
  if (cards.length === 0) return;

  const rect = container.getBoundingClientRect();
  svgHeight.value = rect.height;
  const centerY = rect.height / 2;
  const endX = 70;

  connectorPaths.value = Array.from(cards).map((el) => {
    const cardRect = el.getBoundingClientRect();
    const y = cardRect.top - rect.top + cardRect.height / 2;
    /* Cubic bezier: horizontal departure from origin, horizontal arrival at endpoint */
    const cp1x = endX * 0.5;
    const cp2x = endX * 0.65;
    const d = `M 0,${centerY} C ${cp1x},${centerY} ${cp2x},${y} ${endX},${y}`;
    return { d, endX, endY: y, len: 0 };
  });

  /* Measure path lengths for stroke-dasharray animation */
  nextTick(() => {
    const svg = container.querySelector('.bento-connectors');
    if (!svg) return;
    const pathEls = svg.querySelectorAll('path');
    pathEls.forEach((p, i) => {
      if (connectorPaths.value[i]) {
        connectorPaths.value[i].len = (p as SVGPathElement).getTotalLength();
      }
    });
  });
}

watch(visible, async (v) => {
  if (!v) return;
  /* Compute connector paths from DOM positions */
  await nextTick();
  computePaths();
  /* Animation sequence */
  setTimeout(() => { phase.value = 'door'; }, 300);
  setTimeout(() => { phase.value = 'lines'; }, 900);
  setTimeout(() => { artifactCount.value = 1; }, 1100);
  setTimeout(() => { artifactCount.value = 2; }, 1350);
  setTimeout(() => { artifactCount.value = 3; }, 1600);
  setTimeout(() => { artifactCount.value = 4; }, 1850);
  setTimeout(() => { phase.value = 'done'; }, 2200);
});
</script>

<template>
  <div
    ref="root"
    class="bento-eject"
  >
    <div class="bento-eject-layout">
      <!-- Left: door + command -->
      <div class="bento-door-section">
        <div
          class="bento-door"
          :class="{ 'bento-door-visible': phase !== 'idle' }"
        >
          <UIcon
            name="i-lucide-door-open"
            class="size-8 sm:size-10"
          />
        </div>
        <code
          v-if="phase !== 'idle'"
          class="bento-cmd"
        >lucity eject</code>
      </div>

      <!-- Right: connector SVG + artifact cards -->
      <div
        ref="rightRef"
        class="bento-right"
      >
        <!-- Curved connector paths — single SVG, all lines originate from door center -->
        <svg
          class="bento-connectors"
          :viewBox="`0 0 80 ${svgHeight}`"
          fill="none"
        >
          <path
            v-for="(p, i) in connectorPaths"
            :key="i"
            :d="p.d"
            :stroke="artifactCount > i ? 'var(--bento-accent)' : 'var(--ui-border)'"
            stroke-width="1.5"
            stroke-linecap="round"
            :stroke-dasharray="p.len || 'none'"
            :stroke-dashoffset="phase === 'idle' || phase === 'door' ? (p.len || 0) : 0"
            class="bento-path"
          />
          <circle
            v-for="(p, i) in connectorPaths"
            :key="`dot-${i}`"
            :cx="p.endX"
            :cy="p.endY"
            r="3"
            :fill="artifactCount > i ? 'var(--bento-accent)' : 'var(--ui-border)'"
            class="bento-endpoint"
            :class="{ 'bento-endpoint-active': artifactCount > i }"
          />
        </svg>

        <!-- Artifact cards -->
        <div class="bento-artifacts">
          <div
            v-for="(item, i) in outputs"
            :key="item.label"
            class="bento-output"
            :class="{ 'bento-output-visible': artifactCount > i }"
            :style="{ animationDelay: `${i * 60}ms` }"
          >
            <UIcon
              :name="item.icon"
              class="size-4 shrink-0 text-(--bento-accent)"
            />
            <div class="min-w-0">
              <div class="text-xs font-medium text-(--ui-text)">
                {{ item.label }}
              </div>
              <div class="truncate font-mono text-[10px] text-(--ui-text-muted)">
                {{ item.path }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Tagline — always rendered (opacity-only) so it doesn't shift centering -->
    <div
      class="bento-tagline"
      :class="{ 'bento-tagline-visible': phase === 'done' }"
    >
      No lock-in. Standard tools.
    </div>
  </div>
</template>

<style scoped>
/* Fixed height — prevents the grid from shifting when the
   tagline and artifact cards animate in. */
.bento-eject {
  height: 275px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 44px 20px 16px;
  gap: 14px;
}

.bento-eject-layout {
  display: flex;
  align-items: center;
  gap: 0;
  width: 100%;
  max-width: 460px;
}

/* Left section: door icon + command.
   z-index ensures the door renders above the connector SVG
   that extends behind it. */
.bento-door-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  width: 100px;
  position: relative;
  z-index: 2;
}

@media (min-width: 640px) {
  .bento-door-section {
    width: 120px;
  }
}

.bento-door {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  border: 1.5px solid var(--bento-accent);
  background: var(--ui-bg-elevated);
  color: var(--bento-accent);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transform: scale(0.7);
  transition: all 0.5s cubic-bezier(0.16, 1, 0.3, 1);
}

@media (min-width: 640px) {
  .bento-door {
    width: 64px;
    height: 64px;
  }
}

.bento-door-visible {
  opacity: 1;
  transform: scale(1);
}

.bento-cmd {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--bento-accent);
  font-weight: 500;
  animation: bento-fade-in 0.4s ease both;
  white-space: nowrap;
  /* Inline code styling */
  padding: 3px 8px;
  border-radius: 6px;
  border: 1px solid var(--ui-border);
  background: var(--ui-bg-muted);
}

/* Right section: connector SVG + artifact cards */
.bento-right {
  position: relative;
  flex: 1;
  min-width: 0;
}

/* Single SVG for all connector curves.
   Extends left behind the door icon via negative margin + extra width. */
.bento-connectors {
  position: absolute;
  left: -50px;
  top: 0;
  width: 80px;
  height: 100%;
  pointer-events: none;
  z-index: 1;
}

@media (min-width: 640px) {
  .bento-connectors {
    left: -60px;
  }
}

.bento-path {
  transition: stroke 0.3s ease, stroke-dashoffset 0.5s cubic-bezier(0.16, 1, 0.3, 1);
}

.bento-endpoint {
  transition: fill 0.3s ease;
}

.bento-endpoint-active {
  filter: drop-shadow(0 0 4px var(--bento-accent-glow));
}

/* Stacked artifact cards */
.bento-artifacts {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding-left: 16px;
}

/* Artifact cards */
.bento-output {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 10px;
  border-radius: 8px;
  border: 1px solid var(--ui-border);
  background: var(--ui-bg-elevated);
  white-space: nowrap;
  opacity: 0;
  transform: translateX(-6px);
  min-width: 0;
}

.bento-output-visible {
  animation: bento-artifact-in 0.35s cubic-bezier(0.16, 1, 0.3, 1) both;
}

/* Tagline — always in DOM for stable centering, shown via opacity */
.bento-tagline {
  text-align: center;
  font-size: 11px;
  font-weight: 500;
  color: var(--bento-accent);
  opacity: 0;
}

.bento-tagline-visible {
  animation: bento-fade-in 0.6s ease both;
}

@keyframes bento-artifact-in {
  from { opacity: 0; transform: translateX(-6px); }
  to { opacity: 1; transform: translateX(0); }
}

@keyframes bento-fade-in {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
