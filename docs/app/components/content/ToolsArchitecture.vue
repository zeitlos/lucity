<script setup lang="ts">
import { ref, watch } from 'vue';
import { useBentoVisible } from './useBentoVisible';

const root = ref<HTMLElement | null>(null);
const visible = useBentoVisible(root, 0.15);
const step = ref(0);
const showSubtitle = ref(false);

// Flowing dots along edges
const DOTS_PER_EDGE = 3;
const DOT_DUR = 4; // seconds per loop

interface ArchNode {
  id: string;
  icon: string;
  label: string;
  role: string;
  x: number; // % of container width
  y: number; // % of container height
  phase: number; // animation phase (1-6)
  isLucity?: boolean;
}

interface ArchEdge {
  from: string;
  to: string;
  label?: string;
  phase: number;
}

const nodes: ArchNode[] = [
  { id: 'lucity', icon: '', label: 'Lucity', role: 'Orchestrator', x: 50, y: 30, phase: 1, isLucity: true },
  { id: 'github', icon: 'i-simple-icons-github', label: 'GitHub', role: 'Source code', x: 50, y: 7, phase: 2 },
  { id: 'zot', icon: 'i-lucide-container', label: 'Zot', role: 'Container registry', x: 78, y: 30, phase: 3 },
  { id: 'softserve', icon: 'i-lucide-git-branch', label: 'Soft-serve', role: 'GitOps repo', x: 22, y: 30, phase: 3 },
  { id: 'helm', icon: 'i-simple-icons-helm', label: 'Helm', role: 'Package manager', x: 22, y: 55, phase: 4 },
  { id: 'argocd', icon: 'i-simple-icons-argo', label: 'ArgoCD', role: 'Continuous delivery', x: 50, y: 55, phase: 4 },
  { id: 'kubernetes', icon: 'i-simple-icons-kubernetes', label: 'Kubernetes', role: 'Your cluster', x: 50, y: 76, phase: 5 },
  { id: 'gateway', icon: 'i-lucide-globe', label: 'Gateway API', role: 'Ingress traffic', x: 30, y: 92, phase: 6 },
  { id: 'cnpg', icon: 'i-simple-icons-postgresql', label: 'CloudNativePG', role: 'Managed Postgres', x: 70, y: 92, phase: 6 },
];

const edges: ArchEdge[] = [
  { from: 'github', to: 'lucity', label: 'webhook', phase: 2 },
  { from: 'lucity', to: 'zot', label: 'build & push', phase: 3 },
  { from: 'lucity', to: 'softserve', label: 'values', phase: 3 },
  { from: 'softserve', to: 'helm', phase: 4 },
  { from: 'softserve', to: 'argocd', label: 'sync', phase: 4 },
  { from: 'argocd', to: 'kubernetes', label: 'deploy', phase: 5 },
  { from: 'kubernetes', to: 'gateway', phase: 6 },
  { from: 'kubernetes', to: 'cnpg', phase: 6 },
];

// Mobile node order for vertical pipeline
const mobileNodes = [
  { ...nodes.find(n => n.id === 'github')!, mobilePhase: 1, edgeLabel: 'webhook' },
  { ...nodes.find(n => n.id === 'lucity')!, mobilePhase: 2, edgeLabel: 'build & push' },
  { ...nodes.find(n => n.id === 'zot')!, mobilePhase: 3, edgeLabel: 'values' },
  { ...nodes.find(n => n.id === 'softserve')!, mobilePhase: 3, edgeLabel: 'sync' },
  { ...nodes.find(n => n.id === 'argocd')!, mobilePhase: 4, edgeLabel: 'deploy' },
  { ...nodes.find(n => n.id === 'kubernetes')!, mobilePhase: 5, edgeLabel: undefined },
];

const mobileBottomNodes = [
  { ...nodes.find(n => n.id === 'gateway')!, mobilePhase: 6 },
  { ...nodes.find(n => n.id === 'cnpg')!, mobilePhase: 6 },
];

function getNode(id: string) {
  return nodes.find(n => n.id === id)!;
}

// Icon radius in SVG viewBox units (approximate at ~900px container width)
const ICON_R = 42; // 72px icon → ~80 viewBox units → radius ~42
const LUCITY_R = 52; // 88px icon → ~98 viewBox units → radius ~52

function nodeRadius(id: string): number {
  return id === 'lucity' ? LUCITY_R : ICON_R;
}

function edgePath(edge: ArchEdge): string {
  const from = getNode(edge.from);
  const to = getNode(edge.to);

  // Map percentage coords to viewBox (1000 x 750)
  let x1 = from.x * 10;
  let y1 = from.y * 7.5;
  let x2 = to.x * 10;
  let y2 = to.y * 7.5;

  // Shorten path to stop at icon edges
  const dx = x2 - x1;
  const dy = y2 - y1;
  const len = Math.sqrt(dx * dx + dy * dy);

  if (len === 0) return `M${x1},${y1}`;

  const nx = dx / len;
  const ny = dy / len;

  const fromR = nodeRadius(from.id);
  const toR = nodeRadius(to.id);

  x1 += nx * fromR;
  y1 += ny * fromR;
  x2 -= nx * toR;
  y2 -= ny * toR;

  // Compute a control point for the quadratic bezier
  const mx = (x1 + x2) / 2;
  const my = (y1 + y2) / 2;

  // Perpendicular offset for curve (small, playful)
  const shortenedLen = Math.sqrt((x2 - x1) ** 2 + (y2 - y1) ** 2);
  const offset = shortenedLen * 0.08;
  const cx = mx + (-ny) * offset;
  const cy = my + (nx) * offset;

  return `M${x1},${y1} Q${cx},${cy} ${x2},${y2}`;
}

function edgeLength(edge: ArchEdge): number {
  // Approximate path length for stroke-dasharray
  const from = getNode(edge.from);
  const to = getNode(edge.to);
  const dx = (to.x - from.x) * 10;
  const dy = (to.y - from.y) * 7.5;
  const fullLen = Math.sqrt(dx * dx + dy * dy);
  // Account for shortened endpoints
  const shortened = fullLen - nodeRadius(from.id) - nodeRadius(to.id);
  return Math.max(shortened, 10) * 1.1;
}

function edgeLabelPos(edge: ArchEdge): { x: number; y: number } {
  const from = getNode(edge.from);
  const to = getNode(edge.to);
  return {
    x: ((from.x + to.x) / 2) * 10,
    y: ((from.y + to.y) / 2) * 7.5 - 8,
  };
}

watch(visible, (v) => {
  if (!v) return;
  const delays = [0, 400, 1000, 1600, 2200, 2800];
  delays.forEach((delay, i) => {
    setTimeout(() => { step.value = i + 1; }, delay);
  });
  setTimeout(() => { showSubtitle.value = true; }, 3400);
});
</script>

<template>
  <div
    ref="root"
    class="arch-diagram"
  >
    <!-- Desktop layout -->
    <div class="arch-canvas">
      <!-- SVG connections -->
      <svg
        class="arch-connections"
        viewBox="0 0 1000 750"
        preserveAspectRatio="xMidYMid meet"
        fill="none"
        role="img"
        aria-label="Architecture diagram showing how Lucity orchestrates Kubernetes deployments using standard open-source tools"
      >

        <!-- Edge paths (faint ghost trail) -->
        <path
          v-for="(edge, i) in edges"
          :key="`edge-${edge.from}-${edge.to}`"
          :d="edgePath(edge)"
          class="arch-edge"
          :class="{ 'arch-edge-active': step >= edge.phase }"
          :stroke="`var(--arch-${edge.to}-color)`"
        />

        <!-- Edge labels -->
        <text
          v-for="edge in edges.filter(e => e.label)"
          :key="`label-${edge.from}-${edge.to}`"
          :x="edgeLabelPos(edge).x"
          :y="edgeLabelPos(edge).y"
          class="arch-edge-label"
          :class="{ 'arch-edge-label-visible': step >= edge.phase }"
          text-anchor="middle"
        >
          {{ edge.label }}
        </text>

        <!-- Flowing dots along edges -->
        <template
          v-for="(edge, i) in edges"
          :key="`dots-${i}`"
        >
          <circle
            v-for="d in DOTS_PER_EDGE"
            :key="`dot-${i}-${d}`"
            :r="2 + (d % 2)"
            class="arch-flow-dot"
            :class="{ 'arch-flow-dot-active': step >= edge.phase }"
            :fill="`var(--arch-${edge.to}-color)`"
          >
            <animateMotion
              :path="edgePath(edge)"
              :dur="`${DOT_DUR}s`"
              :begin="`${(d - 1) * (DOT_DUR / DOTS_PER_EDGE)}s`"
              repeatCount="indefinite"
              calcMode="linear"
            />
          </circle>
        </template>
      </svg>

      <!-- Node cards -->
      <div
        v-for="node in nodes"
        :key="node.id"
        class="arch-node"
        :class="{
          'arch-node-visible': step >= node.phase,
          'arch-node-lucity': node.isLucity,
        }"
        :style="{ left: `${node.x}%`, top: `${node.y}%` }"
      >
        <!-- Glow ring -->
        <div
          class="arch-node-glow"
          :style="{ background: `var(--arch-${node.id}-color)` }"
        />

        <!-- Icon -->
        <div class="arch-node-icon">
          <svg
            v-if="node.isLucity"
            class="size-10"
            viewBox="-68.70 -26.26 121.05 121.05"
            xmlns="http://www.w3.org/2000/svg"
          >
            <defs>
              <mask id="lucity-mask">
                <circle
                  cx="-8.18"
                  cy="34.27"
                  r="60.52"
                  fill="white"
                />
                <polygon
                  points="-34.64,20 -17.32,30 -34.64,40 -51.96,30"
                  fill="black"
                  stroke="black"
                  stroke-width="0.5"
                  stroke-linejoin="round"
                />
                <polygon
                  points="-17.32,30 0,40 -17.32,50 -34.64,40"
                  fill="black"
                  stroke="black"
                  stroke-width="0.5"
                  stroke-linejoin="round"
                />
                <polygon
                  points="0,40 17.32,50 0,60 -17.32,50"
                  fill="black"
                  stroke="black"
                  stroke-width="0.5"
                  stroke-linejoin="round"
                />
                <polygon
                  points="17.32,30 34.64,40 17.32,50 0,40"
                  fill="black"
                  stroke="black"
                  stroke-width="0.5"
                  stroke-linejoin="round"
                />
                <polygon
                  points="-34.64,3.67 34.64,3.67 0,23.67"
                  fill="black"
                />
              </mask>
            </defs>
            <circle
              cx="-8.18"
              cy="34.27"
              r="60.52"
              fill="currentColor"
              mask="url(#lucity-mask)"
            />
          </svg>
          <UIcon
            v-else
            :name="node.icon"
            class="size-7"
          />
        </div>

        <!-- Label -->
        <span class="arch-node-label">{{ node.label }}</span>
        <span class="arch-node-role">{{ node.role }}</span>
      </div>
    </div>

    <!-- Mobile layout -->
    <div class="arch-mobile flex flex-col items-center">
      <div
        v-for="(node, i) in mobileNodes"
        :key="node.id"
        class="arch-mobile-step"
        :class="{ 'arch-mobile-step-visible': step >= node.mobilePhase }"
      >
        <div class="arch-mobile-node">
          <div
            class="arch-mobile-icon"
            :class="{ 'arch-mobile-icon-lucity': node.isLucity }"
          >
            <svg
              v-if="node.isLucity"
              class="size-5"
              viewBox="-68.70 -26.26 121.05 121.05"
              xmlns="http://www.w3.org/2000/svg"
            >
              <defs>
                <mask id="lucity-mask-mobile">
                  <circle
                    cx="-8.18"
                    cy="34.27"
                    r="60.52"
                    fill="white"
                  />
                  <polygon
                    points="-34.64,20 -17.32,30 -34.64,40 -51.96,30"
                    fill="black"
                    stroke="black"
                    stroke-width="0.5"
                    stroke-linejoin="round"
                  />
                  <polygon
                    points="-17.32,30 0,40 -17.32,50 -34.64,40"
                    fill="black"
                    stroke="black"
                    stroke-width="0.5"
                    stroke-linejoin="round"
                  />
                  <polygon
                    points="0,40 17.32,50 0,60 -17.32,50"
                    fill="black"
                    stroke="black"
                    stroke-width="0.5"
                    stroke-linejoin="round"
                  />
                  <polygon
                    points="17.32,30 34.64,40 17.32,50 0,40"
                    fill="black"
                    stroke="black"
                    stroke-width="0.5"
                    stroke-linejoin="round"
                  />
                  <polygon
                    points="-34.64,3.67 34.64,3.67 0,23.67"
                    fill="black"
                  />
                </mask>
              </defs>
              <circle
                cx="-8.18"
                cy="34.27"
                r="60.52"
                fill="currentColor"
                mask="url(#lucity-mask-mobile)"
              />
            </svg>
            <UIcon
              v-else
              :name="node.icon"
              class="size-5"
            />
          </div>
          <div class="arch-mobile-text">
            <div class="arch-mobile-label">
              {{ node.label }}
            </div>
            <div class="arch-mobile-role">
              {{ node.role }}
            </div>
          </div>
        </div>

        <!-- Connector -->
        <div
          v-if="i < mobileNodes.length - 1 || mobileBottomNodes.length"
          class="arch-mobile-connector"
        >
          <div class="arch-mobile-line" />
          <span
            v-if="node.edgeLabel"
            class="arch-mobile-edge-label"
          >
            {{ node.edgeLabel }}
          </span>
        </div>
      </div>

      <!-- Bottom row: Gateway API + CloudNativePG side by side -->
      <div
        class="arch-mobile-bottom"
        :class="{ 'arch-mobile-step-visible': step >= 6 }"
      >
        <div
          v-for="node in mobileBottomNodes"
          :key="node.id"
          class="arch-mobile-node arch-mobile-node-compact"
        >
          <div class="arch-mobile-icon">
            <UIcon
              :name="node.icon"
              class="size-5"
            />
          </div>
          <div class="arch-mobile-text">
            <div class="arch-mobile-label">
              {{ node.label }}
            </div>
            <div class="arch-mobile-role">
              {{ node.role }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ── Color variables ── */
.arch-diagram {
  --arch-github-color: oklch(0.40 0.01 0);
  --arch-lucity-color: oklch(0.75 0.18 160);
  --arch-zot-color: oklch(0.55 0.16 30);
  --arch-softserve-color: oklch(0.50 0.16 140);
  --arch-helm-color: oklch(0.50 0.16 250);
  --arch-argocd-color: oklch(0.52 0.18 30);
  --arch-kubernetes-color: oklch(0.50 0.16 250);
  --arch-gateway-color: oklch(0.50 0.14 200);
  --arch-cnpg-color: oklch(0.50 0.16 250);
  padding: 20px 0;
}

.dark .arch-diagram {
  --arch-github-color: oklch(0.78 0.01 0);
  --arch-lucity-color: oklch(0.72 0.16 160);
  --arch-zot-color: oklch(0.72 0.14 30);
  --arch-softserve-color: oklch(0.70 0.14 140);
  --arch-helm-color: oklch(0.70 0.14 250);
  --arch-argocd-color: oklch(0.70 0.16 30);
  --arch-kubernetes-color: oklch(0.70 0.14 250);
  --arch-gateway-color: oklch(0.68 0.12 200);
  --arch-cnpg-color: oklch(0.70 0.14 250);
}

/* ── Desktop canvas ── */
.arch-canvas {
  display: none;
  position: relative;
  width: 100%;
  max-width: 900px;
  margin: 0 auto;
  aspect-ratio: 4 / 3;
  background-image: radial-gradient(circle, var(--ui-border) 1px, transparent 1px);
  background-size: 28px 28px;
  background-position: center center;
  border-radius: 20px;
}

@media (width >= 64rem) {
  .arch-canvas { display: block; }
  .arch-mobile { display: none; }
}

/* ── SVG layer ── */
.arch-connections {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 0;
}

.arch-edge {
  fill: none;
  stroke-width: 1.5;
  stroke-linecap: round;
  opacity: 0;
  transition: opacity 0.6s ease;
}

.arch-edge-active {
  opacity: 0.15;
}

.dark .arch-edge-active {
  opacity: 0.2;
}

.arch-edge-label {
  font-size: 11px;
  font-family: var(--font-mono);
  fill: var(--ui-text-muted);
  opacity: 0;
  transition: opacity 0.4s ease 0.3s;
}

.arch-edge-label-visible {
  opacity: 1;
}

.arch-flow-dot {
  opacity: 0;
  transition: opacity 0.6s ease;
}

.arch-flow-dot-active {
  opacity: 0.6;
}

.dark .arch-flow-dot-active {
  opacity: 0.7;
}

/* ── Nodes ── */
.arch-node {
  position: absolute;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 7px;
  transform: translate(-50%, -50%) scale(0.8);
  opacity: 0;
  transition:
    opacity 0.5s cubic-bezier(0.16, 1, 0.3, 1),
    transform 0.5s cubic-bezier(0.16, 1, 0.3, 1);
  z-index: 1;
}

.arch-node-visible {
  opacity: 1;
  transform: translate(-50%, -50%) scale(1);
}

.arch-node-glow {
  position: absolute;
  width: 84px;
  height: 84px;
  border-radius: 50%;
  opacity: 0;
  filter: blur(18px);
  pointer-events: none;
  transition: opacity 0.6s ease;
}

.arch-node-visible .arch-node-glow {
  opacity: 0.25;
  animation: arch-pulse 4s ease-in-out infinite;
}

.arch-node-lucity .arch-node-glow {
  width: 100px;
  height: 100px;
  opacity: 0;
}

.arch-node-lucity.arch-node-visible .arch-node-glow {
  opacity: 0.35;
}

.arch-node-icon {
  width: 72px;
  height: 72px;
  border-radius: 20px;
  border: 1.5px solid var(--ui-border);
  background: var(--ui-bg-elevated);
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  z-index: 1;
  color: var(--ui-text);
  transition: border-color 0.3s ease, box-shadow 0.3s ease;
}

.arch-node-lucity .arch-node-icon {
  width: 88px;
  height: 88px;
  border-radius: 24px;
  border-color: var(--arch-lucity-color);
  box-shadow: 0 0 24px oklch(0.75 0.18 160 / 0.15);
  color: var(--arch-lucity-color);
}

.dark .arch-node-lucity .arch-node-icon {
  box-shadow: 0 0 24px oklch(0.72 0.16 160 / 0.2);
}

.arch-node-label {
  font-size: 14px;
  font-weight: 600;
  color: var(--ui-text);
  white-space: nowrap;
}

.arch-node-role {
  font-size: 12px;
  color: var(--ui-text-muted);
  white-space: nowrap;
}

/* Stagger glow pulse per node */
.arch-node:nth-child(1) .arch-node-glow { animation-delay: 0s; }
.arch-node:nth-child(2) .arch-node-glow { animation-delay: 0.5s; }
.arch-node:nth-child(3) .arch-node-glow { animation-delay: 1s; }
.arch-node:nth-child(4) .arch-node-glow { animation-delay: 1.5s; }
.arch-node:nth-child(5) .arch-node-glow { animation-delay: 2s; }
.arch-node:nth-child(6) .arch-node-glow { animation-delay: 2.5s; }
.arch-node:nth-child(7) .arch-node-glow { animation-delay: 3s; }
.arch-node:nth-child(8) .arch-node-glow { animation-delay: 3.5s; }
.arch-node:nth-child(9) .arch-node-glow { animation-delay: 4s; }

@keyframes arch-pulse {
  0%, 100% { opacity: 0.15; transform: scale(1); }
  50% { opacity: 0.35; transform: scale(1.12); }
}

/* ── Mobile layout ── */
.arch-mobile {
  flex-direction: column;
  gap: 0;
  padding: 0 16px;
  max-width: 360px;
  margin: 0 auto;
}

.arch-mobile-step {
  display: flex;
  flex-direction: column;
  align-items: center;
  opacity: 0;
  transform: translateY(8px);
  transition:
    opacity 0.4s cubic-bezier(0.16, 1, 0.3, 1),
    transform 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.arch-mobile-step-visible {
  opacity: 1;
  transform: translateY(0);
}

.arch-mobile-node {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  border-radius: 12px;
  border: 1.5px solid var(--ui-border);
  background: var(--ui-bg-elevated);
  min-width: 200px;
}

.arch-mobile-node-compact {
  min-width: 0;
  flex: 1;
}

.arch-mobile-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: var(--ui-bg-muted);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--ui-text);
}

.arch-mobile-icon-lucity {
  background: oklch(0.75 0.18 160 / 0.12);
  color: var(--arch-lucity-color);
}

.arch-mobile-text {
  min-width: 0;
}

.arch-mobile-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--ui-text);
}

.arch-mobile-role {
  font-size: 11px;
  color: var(--ui-text-muted);
}

.arch-mobile-connector {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0;
  padding: 4px 0;
}

.arch-mobile-line {
  width: 1.5px;
  height: 20px;
  background: var(--ui-border);
  border-radius: 1px;
}

.arch-mobile-edge-label {
  font-size: 10px;
  font-family: var(--font-mono);
  color: var(--ui-text-muted);
  padding: 2px 8px;
  border-radius: 4px;
  background: var(--ui-bg-muted);
  margin: 2px 0;
}

.arch-mobile-bottom {
  display: flex;
  gap: 8px;
  margin-top: 4px;
  opacity: 0;
  transform: translateY(8px);
  transition:
    opacity 0.4s cubic-bezier(0.16, 1, 0.3, 1),
    transform 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.arch-mobile-bottom.arch-mobile-step-visible {
  opacity: 1;
  transform: translateY(0);
}

/* ── Reduced motion ── */
@media (prefers-reduced-motion: reduce) {
  .arch-node,
  .arch-mobile-step,
  .arch-mobile-bottom {
    transition: none !important;
    opacity: 1 !important;
    transform: translate(-50%, -50%) scale(1) !important;
  }

  .arch-mobile-step,
  .arch-mobile-bottom {
    transform: translateY(0) !important;
  }

  .arch-edge {
    transition: none !important;
    opacity: 0.15 !important;
  }

  .arch-edge-label {
    transition: none !important;
    opacity: 0.7 !important;
  }

  .arch-node-glow {
    animation: none !important;
  }

  .arch-flow-dot {
    display: none;
  }
}
</style>
