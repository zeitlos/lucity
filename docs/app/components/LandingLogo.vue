<script setup lang="ts">
const props = defineProps<{ tucked?: boolean }>();

type EdgePoint = { x: number; y: number; offset: number };

const PROFILE = 'M 130 0 L 130 76 C 130 102.64, 114.4 112, 70 112 L 60 112 C 15.6 112, 0 121.36, 0 148 L 0 181';
const BEND_STEPS = 12;
const GULP_EASING = 'cubic-bezier(0.3, 1.6, 0.5, 1)';
const SPRING_STIFFNESS = 260;
const SPRING_DAMPING = 0.42;
const REST = logoDefaults();

function cubic(a: number, b: number, c: number, d: number, t: number) {
  const u = 1 - t;
  return u * u * u * a + 3 * u * u * t * b + 3 * u * t * t * c + t * t * t * d;
}

function bend(a: EdgePoint, b: EdgePoint, c: EdgePoint, d: EdgePoint): EdgePoint[] {
  return Array.from({ length: BEND_STEPS }, (_, i) => {
    const t = (i + 1) / BEND_STEPS;
    return {
      x: cubic(a.x, b.x, c.x, d.x, t),
      y: cubic(a.y, b.y, c.y, d.y, t),
      offset: cubic(a.offset, b.offset, c.offset, d.offset, t),
    };
  });
}

const EDGE: EdgePoint[] = [
  { x: 130, y: 0, offset: 120 },
  { x: 130, y: 76, offset: 120 },
  ...bend(
    { x: 130, y: 76, offset: 120 },
    { x: 130, y: 102.64, offset: 120 },
    { x: 114.4, y: 112, offset: 126.37 },
    { x: 70, y: 112, offset: 144.5 },
  ),
  { x: 60, y: 112, offset: 144.5 },
  ...bend(
    { x: 60, y: 112, offset: 144.5 },
    { x: 15.6, y: 112, offset: 162.63 },
    { x: 0, y: 121.36, offset: 169 },
    { x: 0, y: 148, offset: 169 },
  ),
  { x: 0, y: 181, offset: 169 },
];

function face(depth: number) {
  const percent = (value: number) => `${+(value * 100).toFixed(2)}%`;
  const left = EDGE.map((p) => [p.x, p.y] as const);
  const right = EDGE.map((p) => [p.x + p.offset * depth, p.y] as const).reverse();
  return `polygon(${[...left, ...right].map(([x, y]) => `${percent(x / 250)} ${percent(y / 181)}`).join(', ')})`;
}

const faceStyle = { clipPath: face(1), '--face-flat': face(0) };

const root = ref<HTMLElement>();
const mark = ref<HTMLElement>();
const settled = ref(false);
const flex = ref(0);

const markPath = computed(() =>
  logoSheetPath({
    ...REST,
    bendW: REST.bendW + 16 * flex.value,
    flat: REST.flat + 6.4 * flex.value,
    tension: REST.tension - 0.08 * flex.value,
    tension2: REST.tension2 - 0.08 * flex.value,
  }),
);

const spring = { position: 0, velocity: 0, target: 0 };
let frame = 0;
let lastFrame = 0;

function reducedMotion() {
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

function animationsDone() {
  return Promise.allSettled(root.value?.getAnimations({ subtree: true }).map((animation) => animation.finished) ?? []);
}

function step(now: number) {
  const dt = Math.min(0.032, (now - lastFrame) / 1000);
  lastFrame = now;
  const force = SPRING_STIFFNESS * (spring.target - spring.position) - 2 * SPRING_DAMPING * Math.sqrt(SPRING_STIFFNESS) * spring.velocity;
  spring.velocity += force * dt;
  spring.position += spring.velocity * dt;
  if (Math.abs(spring.target - spring.position) < 0.001 && Math.abs(spring.velocity) < 0.001) {
    spring.position = spring.target;
    spring.velocity = 0;
    frame = 0;
  } else {
    frame = requestAnimationFrame(step);
  }
  flex.value = spring.position;
}

function flexTo(target: number) {
  if (!settled.value || reducedMotion()) return;
  spring.target = target;
  if (frame) return;
  lastFrame = performance.now();
  frame = requestAnimationFrame(step);
}

watch(
  () => props.tucked,
  () => {
    if (!settled.value || reducedMotion()) return;
    mark.value?.animate(
      [
        { transform: 'none', easing: GULP_EASING },
        { transform: 'scaleX(1.09)', offset: 0.4, easing: GULP_EASING },
        { transform: 'none' },
      ],
      { duration: 550 },
    );
  },
);

onMounted(async () => {
  await animationsDone();
  settled.value = true;
});

onBeforeUnmount(() => cancelAnimationFrame(frame));
</script>

<template>
  <span
    ref="root"
    class="landing-logo"
    :class="{ 'is-tucked': tucked, 'is-settled': settled }"
    @mouseenter="flexTo(1)"
    @mouseleave="flexTo(0)"
  >
    <span ref="mark" class="mark">
      <span class="mark-draft" aria-hidden="true">
        <span class="mark-face" :style="faceStyle" />
        <svg class="mark-line" viewBox="0 0 250 181">
          <path :d="PROFILE" pathLength="1" />
        </svg>
      </span>
      <svg class="mark-final" viewBox="0 0 250 181" aria-hidden="true">
        <path :d="markPath" />
      </svg>
    </span>
    <span class="word">
      <span class="word-clip">
        <span class="word-text">Lucity</span>
      </span>
    </span>
  </span>
</template>

<style scoped>
.landing-logo {
  display: inline-flex;
  align-items: center;
  line-height: 1;
}

.mark {
  position: relative;
  height: 0.98em;
  aspect-ratio: 250 / 181;
  transform-origin: 0 50%;
}

.mark-draft,
.mark-face,
.mark-final,
.mark-line {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}

.mark-draft {
  animation: solidify 0.55s cubic-bezier(0.65, 0, 0.35, 1) 1.35s both;
}

.mark-face {
  background: var(--brand-mark);
  animation: extrude 0.9s cubic-bezier(0.34, 1.5, 0.55, 1) 1.05s both;
}

.mark-line {
  overflow: visible;
  animation: vanish 0.5s ease 1.8s both;
}

.mark-line path {
  fill: none;
  stroke: var(--brand-mark);
  stroke-width: 11;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-dasharray: 1 2;
  animation: draw 1s cubic-bezier(0.65, 0, 0.35, 1) 0.15s both;
}

.mark-final {
  overflow: visible;
  fill: var(--brand-mark);
  animation: appear 0s 1.95s both;
}

.word {
  display: inline-grid;
  grid-template-columns: 1fr;
  transition: grid-template-columns 0.5s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.word-clip {
  min-width: 0;
  overflow-x: clip;
  padding-left: 0.833em;
  animation: word-in 0.7s cubic-bezier(0.33, 1, 0.68, 1) 1.9s both;
}

.word-text {
  display: inline-block;
  white-space: nowrap;
  font-family: 'Redaction 35', Georgia, serif;
  font-weight: 400;
  transition:
    opacity 0.3s ease 0.08s,
    translate 0.5s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.is-tucked .word {
  grid-template-columns: 0fr;
}

.is-tucked .word-text {
  opacity: 0;
  translate: -0.6em 0;
  transition:
    opacity 0.2s ease,
    translate 0.5s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.is-tucked:hover .word,
:focus-visible > .is-tucked .word {
  grid-template-columns: 1fr;
}

.is-tucked:hover .word-text,
:focus-visible > .is-tucked .word-text {
  opacity: 1;
  translate: none;
  transition:
    opacity 0.3s ease 0.08s,
    translate 0.5s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.is-settled .mark-draft {
  display: none;
}

.is-settled .mark-final,
.is-settled .word-clip {
  animation: none;
}

@keyframes draw {
  from {
    stroke-dashoffset: 1.02;
  }
  to {
    stroke-dashoffset: 0;
  }
}

@keyframes vanish {
  to {
    opacity: 0;
  }
}

@keyframes extrude {
  from {
    clip-path: var(--face-flat);
  }
}

@keyframes solidify {
  from {
    opacity: 0.2;
  }
}

@keyframes appear {
  from {
    opacity: 0;
  }
}

@keyframes word-in {
  from {
    opacity: 0;
    transform: translateX(-0.4em);
  }
}

@media (prefers-reduced-motion: reduce) {
  .mark-draft {
    display: none;
  }

  .mark-final,
  .word-clip {
    animation: none;
  }

  .word,
  .word-text {
    transition: none;
  }
}
</style>
