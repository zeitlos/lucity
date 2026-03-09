<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';

const root = ref<HTMLElement | null>(null);
const visible = ref(false);
const phase = ref<'idle' | 'rack' | 'cables' | 'done'>('idle');
let observer: IntersectionObserver | null = null;

/* ── Port / SVG dimensions ── */

const PORT_W = 18;
const PORT_H = 13;
const PORT_GAP = 4.5;
const PORT_PAD = 12;

/* Port counts: top row wider-spaced, bottom row tight-centered */
const TOP_COUNT = 20;
const BOT_COUNT = 20;
const PORT_COUNT = 24; // kept for bottom standalone panel

/* SVG width stays based on 24-port layout for consistency */
const SVG_W = PORT_PAD * 2 + PORT_COUNT * PORT_W + (PORT_COUNT - 1) * PORT_GAP;

/* Top row: 20 ports evenly spread across full width */
const TOP_GAP = (SVG_W - 2 * PORT_PAD - TOP_COUNT * PORT_W) / (TOP_COUNT - 1);

/* Bottom row: 20 ports tight (original gap), centered */
const BOT_TOTAL_W = BOT_COUNT * PORT_W + (BOT_COUNT - 1) * PORT_GAP;
const BOT_OFFSET = (SVG_W - BOT_TOTAL_W) / 2;

const ROW_TOP_Y = 8;
const ROW_BOT_Y = 95;
const SVG_H = ROW_BOT_Y + PORT_H + 18; // extra room for LEDs

function topPortX(index: number): number {
  return PORT_PAD + index * (PORT_W + TOP_GAP);
}

function botPortX(index: number): number {
  return BOT_OFFSET + index * (PORT_W + PORT_GAP);
}

/* Original portX kept for standalone bottom panel */
function portX(index: number): number {
  return PORT_PAD + index * (PORT_W + PORT_GAP);
}

/* ── Cable definitions — all 20 ports filled, white/gray cables ── */

interface Cable {
  fromPort: number;
  toPort: number;
  shadow: string;
  body: string;
  highlight: string;
  delay: number;
}

const grayPalette = [
  { shadow: '#909090', body: '#ffffff', highlight: '#ffffff' },
  { shadow: '#888888', body: '#f0f0f0', highlight: '#fafafa' },
  { shadow: '#8a8a8a', body: '#e8e8e8', highlight: '#f5f5f5' },
  { shadow: '#929292', body: '#f5f5f5', highlight: '#ffffff' },
  { shadow: '#868686', body: '#ececec', highlight: '#f8f8f8' },
];

/* Routing: mostly straight down, a few swap-pairs for organic crossovers */
const swapPairs: [number, number][] = [[3, 4], [9, 10], [14, 15]];
const emptyPorts = new Set([7, 13]); // leave unplugged to show port detail
const cableRouting: [number, number][] = Array.from({ length: BOT_COUNT }, (_, i) => {
  if (emptyPorts.has(i)) return null;
  for (const [a, b] of swapPairs) {
    if (i === a) return [i, b] as [number, number];
    if (i === b) return [i, a] as [number, number];
  }
  return [i, i] as [number, number];
}).filter((r): r is [number, number] => r !== null);

const cables: Cable[] = cableRouting.map(([from, to], idx) => ({
  fromPort: from,
  toPort: to,
  ...grayPalette[idx % grayPalette.length],
  delay: idx * 40,
}));

/* Every port is plugged */
const topPluggedSet = new Set(cables.map(c => c.fromPort));
const bottomPluggedSet = new Set(cables.map(c => c.toPort));
const cableColorByTopPort = Object.fromEntries(cables.map(c => [c.fromPort, c.body]));
const cableColorByBottomPort = Object.fromEntries(cables.map(c => [c.toPort, c.body]));

/* ── Cable path generation with perspective ── */

function cablePath(fromIdx: number, toIdx: number): string {
  const x1 = topPortX(fromIdx) + PORT_W / 2;
  const y1 = ROW_TOP_Y + PORT_H - 1;
  const x2 = botPortX(toIdx) + PORT_W / 2;
  const y2 = ROW_BOT_Y + 6;

  const midY = (y1 + y2) / 2;
  const spread = Math.abs(fromIdx - toIdx);
  const droop = 16 + spread * 12;

  /* Perspective: cables fan outward from center */
  const center = (BOT_COUNT - 1) / 2;
  const norm = (toIdx - center) / center; // -1..+1
  const shift = norm * 12;

  return `M ${x1} ${y1} C ${x1 + shift} ${midY + droop}, ${x2 + shift} ${midY + droop}, ${x2} ${y2}`;
}

/* ── LED definitions ── */

const leds = [
  { label: 'PWR', color: '#22C55E', blinkDuration: '0s', alwaysOn: true },
  { label: 'ACT', color: '#F59E0B', blinkDuration: '1.8s', alwaysOn: false },
  { label: 'NET', color: '#22C55E', blinkDuration: '2.4s', alwaysOn: false },
  { label: 'HDD', color: '#F59E0B', blinkDuration: '1.3s', alwaysOn: false },
];

/* ── Animation trigger ── */

function startAnimation() {
  if (phase.value !== 'idle') return;
  phase.value = 'rack';
  setTimeout(() => { phase.value = 'cables'; }, 300);
  setTimeout(() => { phase.value = 'done'; }, 300 + cables.length * 40 + 700 + 300);
}

onMounted(() => {
  if (!root.value) return;
  if (typeof IntersectionObserver === 'undefined') {
    visible.value = true;
    startAnimation();
    return;
  }
  observer = new IntersectionObserver(
    ([entry]) => {
      if (entry.isIntersecting) {
        visible.value = true;
        startAnimation();
        observer?.disconnect();
      }
    },
    { threshold: 0.15 },
  );
  observer.observe(root.value);
});

onUnmounted(() => { observer?.disconnect(); });
</script>

<template>
  <div
    ref="root"
    class="rack-wrapper"
  >
    <div
      class="rack"
      :class="{ 'rack-visible': visible }"
    >
      <!-- ── Rack rails ── -->
      <div class="rack-rail rack-rail-left">
        <div class="screw" />
        <div class="screw" />
        <div class="screw" />
        <div class="screw" />
        <div class="screw" />
        <div class="screw" />
        <div class="screw" />
      </div>
      <div class="rack-rail rack-rail-right">
        <div class="screw" />
        <div class="screw" />
        <div class="screw" />
        <div class="screw" />
        <div class="screw" />
        <div class="screw" />
        <div class="screw" />
      </div>

      <!-- ── 1U: Top patch panel (SVG) ── -->
      <div class="rack-unit patch-panel">
        <div class="faceplate patch-faceplate">
          <svg
            class="patch-svg"
            :viewBox="`0 0 ${SVG_W} ${SVG_H}`"
            preserveAspectRatio="xMidYMid meet"
          >
            <defs />

            <!-- Top row of RJ45 ports (20, wide spacing) -->
            <g
              v-for="i in TOP_COUNT"
              :key="`top-${i}`"
            >
              <rect
                :x="topPortX(i - 1)"
                :y="ROW_TOP_Y"
                :width="PORT_W"
                :height="PORT_H"
                rx="1.5"
                ry="1.5"
                fill="#0a0a0a"
                stroke="#2a2a2a"
                stroke-width="0.8"
              />
              <line
                :x1="topPortX(i - 1) + 1.5"
                :y1="ROW_TOP_Y + 0.5"
                :x2="topPortX(i - 1) + PORT_W - 1.5"
                :y2="ROW_TOP_Y + 0.5"
                stroke="#3a3a3a"
                stroke-width="0.6"
                stroke-linecap="round"
              />
              <rect
                :x="topPortX(i - 1) + 2"
                :y="ROW_TOP_Y + 2"
                :width="PORT_W - 4"
                :height="PORT_H - 3"
                rx="0.8"
                fill="#020202"
              />
              <rect
                :x="topPortX(i - 1) + 2"
                :y="ROW_TOP_Y + 2"
                :width="PORT_W - 4"
                height="2"
                rx="0.5"
                fill="#000000"
                opacity="0.4"
              />
              <line
                v-for="pin in 4"
                :key="`tpin-${i}-${pin}`"
                :x1="topPortX(i - 1) + 4 + (pin - 1) * 3"
                :y1="ROW_TOP_Y + 3"
                :x2="topPortX(i - 1) + 4 + (pin - 1) * 3"
                :y2="ROW_TOP_Y + 7"
                stroke="#222222"
                stroke-width="0.7"
              />
              <rect
                :x="topPortX(i - 1) + PORT_W / 2 - 3.5"
                :y="ROW_TOP_Y + PORT_H - 0.5"
                width="7"
                height="3"
                rx="0.5"
                fill="#0a0a0a"
                stroke="#2a2a2a"
                stroke-width="0.5"
              />
              <rect
                v-if="topPluggedSet.has(i - 1)"
                :x="topPortX(i - 1) + 2.5"
                :y="ROW_TOP_Y + 2.5"
                :width="PORT_W - 5"
                :height="PORT_H - 4"
                rx="1"
                :fill="cableColorByTopPort[i - 1]"
                opacity="0.8"
                class="port-plug"
              />
            </g>

            <!-- Bottom row of RJ45 ports (20, tight centered) -->
            <g
              v-for="i in BOT_COUNT"
              :key="`bot-${i}`"
            >
              <rect
                :x="botPortX(i - 1)"
                :y="ROW_BOT_Y"
                :width="PORT_W"
                :height="PORT_H"
                rx="1.5"
                ry="1.5"
                fill="#0a0a0a"
                stroke="#2a2a2a"
                stroke-width="0.8"
              />
              <!-- Metallic top-edge highlight -->
              <line
                :x1="botPortX(i - 1) + 1.5"
                :y1="ROW_BOT_Y + 0.5"
                :x2="botPortX(i - 1) + PORT_W - 1.5"
                :y2="ROW_BOT_Y + 0.5"
                stroke="#3a3a3a"
                stroke-width="0.6"
                stroke-linecap="round"
              />
              <!-- Inner cavity (darker) -->
              <rect
                :x="botPortX(i - 1) + 2"
                :y="ROW_BOT_Y + 2"
                :width="PORT_W - 4"
                :height="PORT_H - 3"
                rx="0.8"
                fill="#020202"
              />
              <!-- Inset shadow overlay -->
              <rect
                :x="botPortX(i - 1) + 2"
                :y="ROW_BOT_Y + 2"
                :width="PORT_W - 4"
                height="2"
                rx="0.5"
                fill="#000000"
                opacity="0.4"
              />
              <line
                v-for="pin in 4"
                :key="`bpin-${i}-${pin}`"
                :x1="botPortX(i - 1) + 4 + (pin - 1) * 3"
                :y1="ROW_BOT_Y + 3"
                :x2="botPortX(i - 1) + 4 + (pin - 1) * 3"
                :y2="ROW_BOT_Y + 7"
                stroke="#222222"
                stroke-width="0.7"
              />
              <!-- Clip tab -->
              <rect
                :x="botPortX(i - 1) + PORT_W / 2 - 3.5"
                :y="ROW_BOT_Y + PORT_H - 0.5"
                width="7"
                height="3"
                rx="0.5"
                fill="#0a0a0a"
                stroke="#2a2a2a"
                stroke-width="0.5"
              />
              <!-- Plugged cable fill -->
              <rect
                v-if="bottomPluggedSet.has(i - 1)"
                :x="botPortX(i - 1) + 2.5"
                :y="ROW_BOT_Y + 2.5"
                :width="PORT_W - 5"
                :height="PORT_H - 4"
                rx="1"
                :fill="cableColorByBottomPort[i - 1]"
                opacity="0.8"
                class="port-plug"
              />
              <!-- LED indicator below port -->
              <circle
                :cx="botPortX(i - 1) + PORT_W / 2"
                :cy="ROW_BOT_Y + PORT_H + 8"
                r="2"
                fill="#22C55E"
                opacity="0.3"
                class="port-led-off"
              />
              <circle
                :cx="botPortX(i - 1) + PORT_W / 2"
                :cy="ROW_BOT_Y + PORT_H + 8"
                r="2"
                fill="#22C55E"
                class="port-led"
                :class="{ 'port-led-on': bottomPluggedSet.has(i - 1) && (phase === 'cables' || phase === 'done') }"
                :style="{ '--port-led-delay': `${(i - 1) * 60}ms` }"
              />
            </g>

            <!-- Patch cables (3 layers each: shadow → body → highlight) -->
            <g
              v-for="(cable, idx) in cables"
              :key="`cable-${idx}`"
              class="cable-group"
              :class="{ 'cable-drawn': phase === 'cables' || phase === 'done' }"
              :style="{ '--cable-delay': `${cable.delay}ms` }"
            >
              <!-- Shadow layer — offset down-right -->
              <path
                :d="cablePath(cable.fromPort, cable.toPort)"
                :stroke="cable.shadow"
                stroke-width="3.5"
                stroke-linecap="round"
                fill="none"
                opacity="0.35"
                transform="translate(1, 1)"
                class="cable-layer cable-shadow"
              />
              <!-- Body layer — centered -->
              <path
                :d="cablePath(cable.fromPort, cable.toPort)"
                :stroke="cable.body"
                stroke-width="2.5"
                stroke-linecap="round"
                fill="none"
                class="cable-layer cable-body"
              />
              <!-- Highlight layer — offset up-left -->
              <path
                :d="cablePath(cable.fromPort, cable.toPort)"
                :stroke="cable.highlight"
                stroke-width="0.8"
                stroke-linecap="round"
                fill="none"
                opacity="0.5"
                transform="translate(-0.5, -0.5)"
                class="cable-layer cable-highlight"
              />
            </g>
          </svg>
        </div>
      </div>

      <!-- ── 1U: Server / LCD console ── -->
      <div class="rack-unit server-unit">
        <div class="faceplate server-faceplate">
          <div class="server-layout">
            <div class="lcd-bezel">
              <div class="lcd-screen">
                <div class="lcd-scanlines" />
                <div class="lcd-content">
                  <div class="lcd-title">
                    Prefer to self-host?
                  </div>
                  <div class="lcd-body">
                    AGPL-3.0 &middot; All features
                  </div>
                  <div class="lcd-body">
                    No limits &middot; No fees
                  </div>
                </div>
              </div>
            </div>

            <div class="server-controls">
              <div class="led-grid">
                <div
                  v-for="led in leds"
                  :key="led.label"
                  class="led-row"
                >
                  <div
                    class="led"
                    :class="{
                      'led-blink': !led.alwaysOn && (phase === 'cables' || phase === 'done'),
                      'led-on': led.alwaysOn && phase !== 'idle',
                    }"
                    :style="{
                      '--led-color': led.color,
                      '--led-blink-duration': led.blinkDuration,
                    }"
                  />
                  <span class="led-label">{{ led.label }}</span>
                </div>
              </div>

              <NuxtLink
                to="/getting-started/quick-start"
                class="rack-button"
              >
                <span class="rack-button-face">
                  Self-host guide
                  <UIcon
                    name="i-lucide-arrow-right"
                    class="size-3.5"
                  />
                </span>
              </NuxtLink>
            </div>
          </div>
        </div>
      </div>

      <!-- ── 2U: NVMe storage (2×5 grid, orange tabs on right) ── -->
      <div class="rack-unit nvme-unit">
        <div class="faceplate storage-faceplate">
          <div class="nvme-layout">
            <!-- Left: service port + status LEDs -->
            <div class="nvme-service">
              <div class="nvme-service-leds">
                <div
                  class="nvme-sled"
                  :class="{ 'nvme-sled-on': phase !== 'idle' }"
                />
                <div
                  class="nvme-sled nvme-sled-act"
                  :class="{ 'nvme-sled-blink': phase === 'cables' || phase === 'done' }"
                />
              </div>
              <div class="nvme-usb-port" />
            </div>
            <!-- Drive grid: 2 rows × 5 -->
            <div class="nvme-grid">
              <div
                v-for="i in 10"
                :key="`nvme-${i}`"
                class="nvme-bay"
              >
                <div class="nvme-slot">
                  <div class="nvme-vent-lines">
                    <div
                      v-for="v in 3"
                      :key="v"
                      class="nvme-vent"
                    />
                  </div>
                </div>
                <div class="nvme-tab" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ── 2U: HDD storage (3×4 grid, dark tabs, status LEDs) ── -->
      <div class="rack-unit hdd-unit">
        <div class="faceplate storage-faceplate">
          <div class="hdd-grid">
            <div
              v-for="i in 12"
              :key="`hdd-${i}`"
              class="hdd-bay"
            >
              <div class="hdd-slot">
                <div class="hdd-vent-lines">
                  <div
                    v-for="v in 5"
                    :key="v"
                    class="hdd-vent"
                  />
                </div>
              </div>
              <div class="hdd-bottom">
                <div
                  class="hdd-led"
                  :class="{ 'hdd-led-blink': phase === 'cables' || phase === 'done' }"
                  :style="{ '--led-blink-duration': `${1.2 + (i % 5) * 0.4}s` }"
                />
                <div class="hdd-tab" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ── 1U: Compute server (branded handle front panel) ── -->
      <div class="rack-unit compute-unit">
        <div class="faceplate compute-faceplate">
          <div class="compute-layout">
            <!-- Left ear -->
            <div class="compute-ear compute-ear-left">
              <div class="power-button">
                <div class="power-icon">⏻</div>
              </div>
              <div class="compute-leds">
                <div class="compute-led compute-led-on" />
                <div class="compute-led compute-led-blink" />
                <div class="compute-led compute-led-blink2" />
              </div>
            </div>

            <!-- Center: hex mesh background + tapered handle -->
            <div class="compute-center">
              <div class="compute-hex-mesh" />
              <div class="compute-handle">
                <div class="compute-handle-inner">
                  <span class="compute-brand-text">lucity</span>
                </div>
              </div>
            </div>

            <!-- Right ear -->
            <div class="compute-ear compute-ear-right">
              <div class="compute-port compute-port-usb" />
              <div class="compute-port compute-port-usb" />
              <div class="compute-port compute-port-idrac" />
            </div>
          </div>
        </div>
      </div>

      <!-- ── 1U: Bottom panel ── -->
      <div class="rack-unit bottom-panel">
        <div class="faceplate patch-faceplate">
          <svg
            class="patch-svg bottom-svg"
            :viewBox="`0 0 ${SVG_W} ${PORT_H + 12}`"
            preserveAspectRatio="xMidYMid meet"
          >
            <g
              v-for="i in PORT_COUNT"
              :key="`bp-${i}`"
            >
              <rect
                :x="portX(i - 1)"
                y="4"
                :width="PORT_W"
                :height="PORT_H"
                rx="1.5"
                ry="1.5"
                fill="#0a0a0a"
                stroke="#2a2a2a"
                stroke-width="0.8"
              />
              <!-- Metallic top-edge highlight -->
              <line
                :x1="portX(i - 1) + 1.5"
                y1="4.5"
                :x2="portX(i - 1) + PORT_W - 1.5"
                y2="4.5"
                stroke="#3a3a3a"
                stroke-width="0.6"
                stroke-linecap="round"
              />
              <rect
                :x="portX(i - 1) + 2"
                y="6"
                :width="PORT_W - 4"
                :height="PORT_H - 3"
                rx="0.8"
                fill="#020202"
              />
              <!-- Inset shadow -->
              <rect
                :x="portX(i - 1) + 2"
                y="6"
                :width="PORT_W - 4"
                height="2"
                rx="0.5"
                fill="#000000"
                opacity="0.4"
              />
              <line
                v-for="pin in 4"
                :key="`bppin-${i}-${pin}`"
                :x1="portX(i - 1) + 4 + (pin - 1) * 3"
                y1="7"
                :x2="portX(i - 1) + 4 + (pin - 1) * 3"
                y2="10.5"
                stroke="#222222"
                stroke-width="0.7"
              />
              <!-- Clip tab (wider, taller) -->
              <rect
                :x="portX(i - 1) + PORT_W / 2 - 3.5"
                :y="PORT_H + 3.5"
                width="7"
                height="3"
                rx="0.5"
                fill="#0a0a0a"
                stroke="#2a2a2a"
                stroke-width="0.5"
              />
            </g>
          </svg>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ── Rack wrapper ── */

.rack-wrapper {
  display: flex;
  justify-content: center;
  padding: 0 16px;
}

.rack {
  position: relative;
  width: 100%;
  max-width: 700px;
  background: linear-gradient(
    180deg,
    oklch(0.88 0.005 80) 0%,
    oklch(0.84 0.005 80) 50%,
    oklch(0.80 0.005 80) 100%
  );
  border-radius: 6px;
  overflow: hidden;
  box-shadow:
    0 2px 20px oklch(0 0 0 / 0.15),
    inset 0 1px 0 oklch(1 0 0 / 0.3);

  opacity: 0;
  transform: translateY(12px);
  transition: opacity 0.5s ease, transform 0.5s ease;
}

.rack-visible {
  opacity: 1;
  transform: translateY(0);
}

/* ── Rack rails ── */

.rack-rail {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 24px;
  background: linear-gradient(
    90deg,
    oklch(0.78 0.005 80) 0%,
    oklch(0.82 0.005 80) 50%,
    oklch(0.78 0.005 80) 100%
  );
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: space-evenly;
  padding: 8px 0;
  z-index: 2;
}

.rack-rail-left {
  left: 0;
  border-right: 1px solid oklch(0.72 0.005 80);
}

.rack-rail-right {
  right: 0;
  border-left: 1px solid oklch(0.72 0.005 80);
}

.screw {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: radial-gradient(
    circle at 40% 35%,
    oklch(0.82 0.01 80) 0%,
    oklch(0.72 0.01 80) 50%,
    oklch(0.65 0.01 80) 100%
  );
  box-shadow:
    inset 0 0.5px 1px oklch(1 0 0 / 0.3),
    0 1px 2px oklch(0 0 0 / 0.15);
  position: relative;
}

.screw::before,
.screw::after {
  content: '';
  position: absolute;
  background: oklch(0.55 0.01 80);
  border-radius: 0.5px;
}

.screw::before {
  width: 6px;
  height: 1px;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.screw::after {
  width: 1px;
  height: 6px;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

/* ── Rack unit (1U) ── */

.rack-unit {
  margin: 0 24px;
  border-bottom: 1px solid oklch(0.72 0.005 80);
}

.rack-unit:last-child {
  border-bottom: none;
}

.faceplate {
  background: oklch(0.90 0.005 80);
  border: 1px solid oklch(0.76 0.005 80);
  border-radius: 3px;
  margin: 6px 8px;
  position: relative;
}

/* ── Patch panel SVG ── */

.patch-faceplate {
  padding: 4px 0;
}

.patch-svg {
  display: block;
  width: 100%;
  height: auto;
}

.bottom-svg {
  max-height: 28px;
}

/* ── Cable 3D layers ── */

.cable-group {
  --cable-delay: 0ms;
}

.cable-layer {
  stroke-dasharray: 300;
  stroke-dashoffset: 300;
  transition: stroke-dashoffset 0.5s cubic-bezier(0.16, 1, 0.3, 1);
  transition-delay: var(--cable-delay);
}

.cable-drawn .cable-layer {
  stroke-dashoffset: 0;
}

/* ── Server unit (LCD console) ── */

.server-unit .faceplate {
  padding: 12px;
}

.server-faceplate {
  background: linear-gradient(
    180deg,
    oklch(0.92 0.005 80) 0%,
    oklch(0.86 0.005 80) 100%
  ) !important;
}

.server-layout {
  display: flex;
  gap: 16px;
  align-items: stretch;
}

/* ── LCD screen ── */

.lcd-bezel {
  flex: 1;
  min-width: 0;
  background: oklch(0.08 0.01 55);
  border-radius: 4px;
  padding: 3px;
  box-shadow:
    inset 0 1px 3px oklch(0 0 0 / 0.6),
    inset 0 0 1px oklch(0 0 0 / 0.4);
}

.lcd-screen {
  --lcd-glow: oklch(0.82 0.16 170);

  position: relative;
  background: oklch(0.10 0.04 180);
  border-radius: 3px;
  padding: 12px 14px;
  overflow: hidden;
  box-shadow:
    inset 0 0 8px oklch(0 0 0 / 0.3),
    0 0 12px oklch(0.82 0.16 170 / 0.08),
    0 0 30px oklch(0.82 0.16 170 / 0.03);
}

.lcd-scanlines {
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    0deg,
    oklch(0 0 0 / 0.06) 0px,
    oklch(0 0 0 / 0.06) 1px,
    transparent 1px,
    transparent 3px
  );
  pointer-events: none;
  z-index: 1;
}

.lcd-content {
  position: relative;
  z-index: 2;
  font-family: var(--font-lcd);
  color: var(--lcd-glow);
  text-shadow:
    0 0 4px oklch(0.82 0.16 170),
    0 0 10px oklch(0.82 0.16 170 / 0.5),
    0 0 20px oklch(0.82 0.16 170 / 0.2);
}

.lcd-title {
  font-size: 20px;
  line-height: 1.3;
  margin-bottom: 4px;
}

.lcd-body {
  font-size: 15px;
  line-height: 1.4;
  opacity: 0.7;
}

/* ── LEDs + controls ── */

.server-controls {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 10px;
  flex-shrink: 0;
  min-width: 120px;
}

.led-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px 12px;
}

.led-row {
  display: flex;
  align-items: center;
  gap: 5px;
}

.led {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: oklch(0.70 0.01 80);
  box-shadow: inset 0 0.5px 1px oklch(0 0 0 / 0.2);
  flex-shrink: 0;
}

.led-on {
  background: var(--led-color);
  box-shadow:
    0 0 4px var(--led-color),
    0 0 8px color-mix(in oklch, var(--led-color) 40%, transparent);
}

.led-blink {
  background: var(--led-color);
  animation: led-pulse var(--led-blink-duration) ease-in-out infinite;
  box-shadow:
    0 0 4px var(--led-color),
    0 0 8px color-mix(in oklch, var(--led-color) 40%, transparent);
}

@keyframes led-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.25; }
}

.led-label {
  font-family: var(--font-mono);
  font-size: 9px;
  font-weight: 500;
  color: oklch(0.45 0.01 55);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

/* ── Rack button (CTA) ── */

.rack-button {
  display: block;
  text-decoration: none;
}

.rack-button-face {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 500;
  color: oklch(0.25 0.01 55);
  background: linear-gradient(
    180deg,
    oklch(0.82 0.005 80) 0%,
    oklch(0.74 0.005 80) 100%
  );
  border: 1px solid oklch(0.68 0.005 80);
  border-radius: 4px;
  padding: 7px 14px;
  cursor: pointer;
  box-shadow:
    0 1px 3px oklch(0 0 0 / 0.1),
    inset 0 1px 0 oklch(1 0 0 / 0.25);
  transition: all 0.15s ease;
  white-space: nowrap;
}

.rack-button:hover .rack-button-face {
  background: linear-gradient(
    180deg,
    oklch(0.86 0.005 80) 0%,
    oklch(0.78 0.005 80) 100%
  );
  box-shadow:
    0 1px 2px oklch(0 0 0 / 0.08),
    inset 0 1px 0 oklch(1 0 0 / 0.3);
  color: oklch(0.18 0.01 55);
}

.rack-button:active .rack-button-face {
  background: linear-gradient(
    180deg,
    oklch(0.70 0.005 80) 0%,
    oklch(0.74 0.005 80) 100%
  );
  box-shadow: inset 0 1px 3px oklch(0 0 0 / 0.15);
  transform: translateY(1px);
}

/* ── NVMe storage (2×5 grid, orange tabs on right side) ── */

.storage-faceplate {
  padding: 8px 10px !important;
}

.nvme-layout {
  display: flex;
  gap: 8px;
  align-items: stretch;
}

.nvme-service {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 4px 6px;
  flex-shrink: 0;
}

.nvme-service-leds {
  display: flex;
  gap: 4px;
}

.nvme-sled {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: oklch(0.30 0.01 55);
}

.nvme-sled-on {
  background: #22C55E;
  box-shadow: 0 0 3px #22C55E;
}

.nvme-sled-act {
  background: oklch(0.30 0.01 55);
}

.nvme-sled-blink {
  background: #F59E0B !important;
  box-shadow: 0 0 3px #F59E0B;
  animation: led-pulse 1.4s ease-in-out infinite;
}

.nvme-usb-port {
  width: 7px;
  height: 10px;
  border-radius: 1.5px;
  background: oklch(0.08 0.01 55);
  border: 1px solid oklch(0.25 0.01 55);
}

.nvme-grid {
  flex: 1;
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  grid-template-rows: 1fr 1fr;
  gap: 3px;
}

.nvme-bay {
  display: flex;
  border-radius: 2px;
  overflow: hidden;
  border: 1px solid oklch(0.68 0.005 80);
  background: oklch(0.84 0.005 80);
}

.nvme-slot {
  flex: 1;
  padding: 3px 4px;
  background: oklch(0.18 0.01 55);
  display: flex;
  align-items: center;
}

.nvme-vent-lines {
  display: flex;
  flex-direction: column;
  gap: 3px;
  width: 100%;
}

.nvme-vent {
  height: 1.5px;
  background: oklch(0.14 0.01 55);
  border-radius: 1px;
  box-shadow: 0 0.5px 0 oklch(0.22 0.01 55);
}

.nvme-tab {
  width: 8px;
  flex-shrink: 0;
  background: linear-gradient(180deg, #F97316 0%, #EA580C 100%);
  position: relative;
  cursor: pointer;
}

.nvme-tab::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 2px;
  height: 8px;
  border-radius: 1px;
  background: oklch(1 0 0 / 0.25);
}

/* ── HDD storage (3×4 grid, dark tabs, status LEDs) ── */

.hdd-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  grid-template-rows: repeat(3, 1fr);
  gap: 3px;
}

.hdd-bay {
  display: flex;
  flex-direction: column;
  border-radius: 2px;
  overflow: hidden;
  border: 1px solid oklch(0.68 0.005 80);
  background: oklch(0.84 0.005 80);
}

.hdd-slot {
  flex: 1;
  padding: 4px 5px;
  background: oklch(0.18 0.01 55);
  display: flex;
  align-items: center;
  min-height: 22px;
}

.hdd-vent-lines {
  display: flex;
  flex-direction: column;
  gap: 2.5px;
  width: 100%;
}

.hdd-vent {
  height: 1.5px;
  background: oklch(0.14 0.01 55);
  border-radius: 1px;
  box-shadow: 0 0.5px 0 oklch(0.22 0.01 55);
}

.hdd-bottom {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 2px 4px;
  background: oklch(0.84 0.005 80);
}

.hdd-led {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: oklch(0.30 0.01 55);
}

.hdd-led-blink {
  background: #22C55E;
  box-shadow: 0 0 3px #22C55E;
  animation: led-pulse var(--led-blink-duration) ease-in-out infinite;
}

.hdd-tab {
  width: 16px;
  height: 5px;
  border-radius: 1px;
  background: linear-gradient(180deg, oklch(0.38 0.01 55) 0%, oklch(0.28 0.01 55) 100%);
  box-shadow: inset 0 0.5px 0 oklch(1 0 0 / 0.08);
}

/* ── Compute server (branded handle front panel) ── */

.compute-faceplate {
  padding: 0 !important;
  overflow: hidden;
  background: oklch(0.14 0.01 55) !important;
  border-color: oklch(0.22 0.01 55) !important;
}

.compute-layout {
  display: flex;
  align-items: stretch;
  min-height: 88px;
  background: oklch(0.14 0.01 55);
}

.compute-ear {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: oklch(0.14 0.01 55);
  flex-shrink: 0;
}

.compute-ear-left {
  border-right: 1px solid oklch(0.10 0.01 55);
}

.compute-ear-right {
  border-left: 1px solid oklch(0.10 0.01 55);
}

.power-button {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: radial-gradient(
    circle at 45% 40%,
    oklch(0.22 0.01 55) 0%,
    oklch(0.14 0.01 55) 100%
  );
  border: 1px solid oklch(0.28 0.01 55);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow:
    inset 0 1px 2px oklch(0 0 0 / 0.3),
    0 1px 0 oklch(1 0 0 / 0.05);
}

.power-icon {
  font-size: 9px;
  color: oklch(0.50 0.01 55);
  line-height: 1;
}

.compute-leds {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.compute-led {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: oklch(0.12 0.01 55);
}

.compute-led-on {
  background: #22C55E;
  box-shadow: 0 0 4px #22C55E;
}

.compute-led-blink {
  background: #3B82F6;
  box-shadow: 0 0 3px #3B82F6;
  animation: led-pulse 2.2s ease-in-out infinite;
}

.compute-led-blink2 {
  background: #F59E0B;
  box-shadow: 0 0 3px #F59E0B;
  animation: led-pulse 1.6s ease-in-out infinite;
}

/* Center area: holds hex mesh + handle */
.compute-center {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

/* Dark hex/honeycomb mesh behind the handle */
.compute-hex-mesh {
  position: absolute;
  inset: 0;
  /* Layered radial gradients to approximate hex dot pattern */
  --hex-c: oklch(0.26 0.01 55);
  --hex-bg: oklch(0.09 0.01 55);
  background-color: var(--hex-bg);
  background-image:
    radial-gradient(circle at 50% 50%, var(--hex-c) 1.8px, transparent 1.8px),
    radial-gradient(circle at 50% 50%, var(--hex-c) 1.8px, transparent 1.8px);
  background-size: 10px 17.32px;
  background-position: 0 0, 5px 8.66px;
}

/*
  Handle shape: >===<
  Full height at edges, short taper, then straight through the middle.
  ~12% taper zone on each side, flat 76% in the center at ~36% height.
*/
.compute-handle {
  position: relative;
  z-index: 1;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background:
    linear-gradient(180deg,
      oklch(0.72 0.005 80) 0%,
      oklch(0.84 0.005 80) 20%,
      oklch(0.88 0.005 80) 50%,
      oklch(0.84 0.005 80) 80%,
      oklch(0.72 0.005 80) 100%
    );
  clip-path: polygon(
    /* Top edge: full height at sides, short taper, then straight */
    0% 0%,
    10% 32%,
    90% 32%,
    100% 0%,
    /* Bottom edge: mirror */
    100% 100%,
    90% 68%,
    10% 68%,
    0% 100%
  );
}

.compute-handle-inner {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 20px;
  width: 100%;
  height: 100%;
}

.compute-brand-text {
  font-family: var(--font-sans);
  font-size: 20px;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: oklch(0.40 0.01 55);
  text-shadow: 0 1px 0 oklch(1 0 0 / 0.3);
}

.compute-port {
  border-radius: 2px;
  background: oklch(0.08 0.01 55);
  border: 1px solid oklch(0.22 0.01 55);
}

.compute-port-usb {
  width: 7px;
  height: 11px;
}

.compute-port-idrac {
  width: 13px;
  height: 11px;
  position: relative;
}

.compute-port-idrac::after {
  content: '';
  position: absolute;
  inset: 2px;
  background: oklch(0.12 0.01 55);
  border-radius: 1px;
}

/* ── Port LED indicators ── */

.port-led-off {
  opacity: 0.15;
}

.port-led {
  opacity: 0;
  transition: opacity 0.3s ease;
  transition-delay: var(--port-led-delay, 0ms);
}

.port-led-on {
  opacity: 1;
  animation: port-led-blink 1.8s ease-in-out infinite;
  animation-delay: var(--port-led-delay, 0ms);
  filter: drop-shadow(0 0 2px #22C55E);
}

@keyframes port-led-blink {
  0%, 100% { opacity: 1; }
  40% { opacity: 1; }
  50% { opacity: 0.3; }
  60% { opacity: 1; }
}

/* ── Responsive ── */

@media (max-width: 639px) {
  .rack-rail {
    width: 16px;
  }

  .screw {
    width: 7px;
    height: 7px;
  }

  .screw::before { width: 4px; }
  .screw::after { height: 4px; }

  .server-layout {
    flex-direction: column;
    gap: 10px;
  }

  .server-controls {
    flex-direction: row;
    align-items: center;
    min-width: 0;
  }

  .lcd-title { font-size: 17px; }
  .lcd-body { font-size: 13px; }
}

/* ── Reduced motion ── */

@media (prefers-reduced-motion: reduce) {
  .rack {
    opacity: 1 !important;
    transform: none !important;
    transition: none !important;
  }

  .cable-layer {
    stroke-dashoffset: 0 !important;
    transition: none !important;
  }

  .led-blink,
  .hdd-led-blink,
  .nvme-sled-blink,
  .compute-led-blink,
  .compute-led-blink2,
  .port-led-on {
    animation: none !important;
    opacity: 1;
  }
}
</style>

<!-- Non-scoped dark mode: :global(.dark) doesn't work with Nuxt UI / Tailwind v4 -->
<style>
.dark .rack-wrapper .rack {
  background: linear-gradient(
    180deg,
    oklch(0.22 0.01 55) 0%,
    oklch(0.20 0.01 55) 50%,
    oklch(0.18 0.01 55) 100%
  );
  box-shadow:
    0 2px 20px oklch(0 0 0 / 0.4),
    inset 0 1px 0 oklch(1 0 0 / 0.04);
}

.dark .rack-wrapper .rack-rail {
  background: linear-gradient(
    90deg,
    oklch(0.28 0.01 55) 0%,
    oklch(0.32 0.01 55) 50%,
    oklch(0.28 0.01 55) 100%
  );
}

.dark .rack-wrapper .rack-rail-left {
  border-right-color: oklch(0.35 0.01 55);
}

.dark .rack-wrapper .rack-rail-right {
  border-left-color: oklch(0.35 0.01 55);
}

.dark .rack-wrapper .screw {
  background: radial-gradient(
    circle at 40% 35%,
    oklch(0.42 0.01 55) 0%,
    oklch(0.32 0.01 55) 50%,
    oklch(0.26 0.01 55) 100%
  );
  box-shadow:
    inset 0 0.5px 1px oklch(1 0 0 / 0.08),
    0 1px 2px oklch(0 0 0 / 0.3);
}

.dark .rack-wrapper .screw::before,
.dark .rack-wrapper .screw::after {
  background: oklch(0.22 0.01 55);
}

.dark .rack-wrapper .rack-unit {
  border-bottom-color: oklch(0.12 0.005 55);
}

.dark .rack-wrapper .faceplate {
  background: oklch(0.16 0.01 55);
  border-color: oklch(0.24 0.01 55);
}

.dark .rack-wrapper .server-faceplate {
  background: linear-gradient(
    180deg,
    oklch(0.18 0.01 55) 0%,
    oklch(0.14 0.01 55) 100%
  ) !important;
}

.dark .rack-wrapper .led {
  background: oklch(0.15 0.01 55);
  box-shadow: inset 0 0.5px 1px oklch(0 0 0 / 0.4);
}

.dark .rack-wrapper .led.led-on {
  background: var(--led-color);
  box-shadow:
    0 0 4px var(--led-color),
    0 0 8px color-mix(in oklch, var(--led-color) 40%, transparent);
}

.dark .rack-wrapper .led.led-blink {
  background: var(--led-color);
  box-shadow:
    0 0 4px var(--led-color),
    0 0 8px color-mix(in oklch, var(--led-color) 40%, transparent);
}

.dark .rack-wrapper .led-label {
  color: oklch(0.50 0.01 55);
}

.dark .rack-wrapper .rack-button-face {
  color: oklch(0.85 0.01 55);
  background: linear-gradient(
    180deg,
    oklch(0.28 0.01 55) 0%,
    oklch(0.22 0.01 55) 100%
  );
  border-color: oklch(0.35 0.01 55);
  box-shadow:
    0 1px 3px oklch(0 0 0 / 0.3),
    inset 0 1px 0 oklch(1 0 0 / 0.06);
}

.dark .rack-wrapper .rack-button:hover .rack-button-face {
  background: linear-gradient(
    180deg,
    oklch(0.32 0.01 55) 0%,
    oklch(0.26 0.01 55) 100%
  );
  box-shadow:
    0 1px 2px oklch(0 0 0 / 0.2),
    inset 0 1px 0 oklch(1 0 0 / 0.08);
  color: oklch(0.92 0.01 55);
}

.dark .rack-wrapper .rack-button:active .rack-button-face {
  background: linear-gradient(
    180deg,
    oklch(0.20 0.01 55) 0%,
    oklch(0.24 0.01 55) 100%
  );
  box-shadow: inset 0 1px 3px oklch(0 0 0 / 0.4);
}

/* ── Dark: NVMe storage ── */

.dark .rack-wrapper .nvme-bay {
  border-color: oklch(0.24 0.01 55);
  background: oklch(0.20 0.01 55);
}

.dark .rack-wrapper .nvme-slot {
  background: oklch(0.12 0.01 55);
}

.dark .rack-wrapper .nvme-vent {
  background: oklch(0.08 0.01 55);
  box-shadow: 0 0.5px 0 oklch(0.16 0.01 55);
}

.dark .rack-wrapper .nvme-sled {
  background: oklch(0.20 0.01 55);
}

.dark .rack-wrapper .nvme-sled.nvme-sled-on {
  background: #22C55E;
  box-shadow: 0 0 3px #22C55E;
}

.dark .rack-wrapper .nvme-sled.nvme-sled-blink {
  background: #F59E0B;
  box-shadow: 0 0 3px #F59E0B;
}

.dark .rack-wrapper .nvme-usb-port {
  background: oklch(0.06 0.01 55);
  border-color: oklch(0.20 0.01 55);
}

/* ── Dark: HDD storage ── */

.dark .rack-wrapper .hdd-bay {
  border-color: oklch(0.24 0.01 55);
  background: oklch(0.20 0.01 55);
}

.dark .rack-wrapper .hdd-slot {
  background: oklch(0.12 0.01 55);
}

.dark .rack-wrapper .hdd-vent {
  background: oklch(0.08 0.01 55);
  box-shadow: 0 0.5px 0 oklch(0.16 0.01 55);
}

.dark .rack-wrapper .hdd-bottom {
  background: oklch(0.20 0.01 55);
}

.dark .rack-wrapper .hdd-led {
  background: oklch(0.12 0.01 55);
}

.dark .rack-wrapper .hdd-led.hdd-led-blink {
  background: #22C55E;
  box-shadow: 0 0 3px #22C55E;
}

.dark .rack-wrapper .hdd-tab {
  background: linear-gradient(180deg, oklch(0.32 0.01 55) 0%, oklch(0.22 0.01 55) 100%);
}

/* ── Dark: Compute server ── */

.dark .rack-wrapper .compute-faceplate {
  background: oklch(0.10 0.01 55) !important;
  border-color: oklch(0.16 0.01 55) !important;
}

.dark .rack-wrapper .compute-layout {
  background: oklch(0.10 0.01 55);
}

.dark .rack-wrapper .compute-center {
  background: oklch(0.10 0.01 55);
}

.dark .rack-wrapper .compute-hex-mesh {
  --hex-c: oklch(0.20 0.01 55);
  --hex-bg: oklch(0.06 0.01 55);
}

.dark .rack-wrapper .compute-ear {
  background: oklch(0.10 0.01 55);
}

.dark .rack-wrapper .compute-ear-left {
  border-right-color: oklch(0.06 0.01 55);
}

.dark .rack-wrapper .compute-ear-right {
  border-left-color: oklch(0.06 0.01 55);
}

.dark .rack-wrapper .compute-handle {
  background: linear-gradient(180deg,
    oklch(0.22 0.01 55) 0%,
    oklch(0.30 0.01 55) 20%,
    oklch(0.34 0.01 55) 50%,
    oklch(0.30 0.01 55) 80%,
    oklch(0.22 0.01 55) 100%
  );
}

.dark .rack-wrapper .compute-brand-text {
  color: oklch(0.62 0.01 55);
  text-shadow: 0 1px 0 oklch(0 0 0 / 0.5);
}

.dark .rack-wrapper .compute-led {
  background: oklch(0.06 0.01 55);
}

.dark .rack-wrapper .compute-led.compute-led-on {
  background: #22C55E;
  box-shadow: 0 0 4px #22C55E;
}

.dark .rack-wrapper .compute-led.compute-led-blink {
  background: #3B82F6;
  box-shadow: 0 0 3px #3B82F6;
}

.dark .rack-wrapper .compute-led.compute-led-blink2 {
  background: #F59E0B;
  box-shadow: 0 0 3px #F59E0B;
}

.dark .rack-wrapper .power-button {
  background: radial-gradient(
    circle at 45% 40%,
    oklch(0.16 0.01 55) 0%,
    oklch(0.08 0.01 55) 100%
  );
  border-color: oklch(0.20 0.01 55);
}

.dark .rack-wrapper .compute-port {
  background: oklch(0.04 0.01 55);
  border-color: oklch(0.14 0.01 55);
}

/* ── Dark: Storage faceplate ── */

.dark .rack-wrapper .storage-faceplate {
  background: oklch(0.16 0.01 55);
  border-color: oklch(0.24 0.01 55);
}

/* ── Dark: Cables — darker gray ── */

.dark .rack-wrapper .cable-shadow {
  stroke: #404040 !important;
}

.dark .rack-wrapper .cable-body {
  stroke: #808080 !important;
}

.dark .rack-wrapper .cable-highlight {
  stroke: #999999 !important;
}

.dark .rack-wrapper .port-plug {
  fill: #808080 !important;
}
</style>
