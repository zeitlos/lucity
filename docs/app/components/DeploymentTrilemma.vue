<script setup lang="ts">
type AxisScores = { simplicity: number; scalability: number; sovereignty: number };

const profiles: Record<string, { color: string; scores: AxisScores }> = {
  paas: { color: '#7f38f8', scores: { simplicity: 0.96, scalability: 0.85, sovereignty: 0.34 } },
  selfhost: { color: '#ff5199', scores: { simplicity: 0.8, scalability: 0.24, sovereignty: 0.82 } },
  k8s: { color: '#fc715e', scores: { simplicity: 0.18, scalability: 0.95, sovereignty: 0.85 } },
};

const active = ref<keyof typeof profiles>('paas');
const activeColor = computed(() => profiles[active.value].color);
const displayed = reactive<AxisScores>({ ...profiles.paas.scores });

let rafId = 0;
function select(key: keyof typeof profiles) {
  active.value = key;
  const target = profiles[key].scores;
  if (import.meta.client && window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
    Object.assign(displayed, target);
    return;
  }
  const start = { ...displayed };
  const t0 = performance.now();
  const dur = 450;
  cancelAnimationFrame(rafId);
  const step = (now: number) => {
    const k = Math.min(1, (now - t0) / dur);
    const e = 1 - Math.pow(1 - k, 3);
    displayed.simplicity = start.simplicity + (target.simplicity - start.simplicity) * e;
    displayed.scalability = start.scalability + (target.scalability - start.scalability) * e;
    displayed.sovereignty = start.sovereignty + (target.sovereignty - start.sovereignty) * e;
    if (k < 1) rafId = requestAnimationFrame(step);
  };
  rafId = requestAnimationFrame(step);
}

onBeforeUnmount(() => cancelAnimationFrame(rafId));

const cx = 220;
const cy = 190;
const R = 150;
const axes = {
  simplicity: { ux: 0, uy: -1 },
  sovereignty: { ux: 0.8660254, uy: 0.5 },
  scalability: { ux: -0.8660254, uy: 0.5 },
};

function coord(axis: keyof typeof axes, s: number) {
  const a = axes[axis];
  return { x: cx + R * s * a.ux, y: cy + R * s * a.uy };
}
function tri(s: number) {
  return [coord('simplicity', s), coord('sovereignty', s), coord('scalability', s)]
    .map(p => `${p.x.toFixed(1)},${p.y.toFixed(1)}`)
    .join(' ');
}
const dataPoly = computed(() =>
  [coord('simplicity', displayed.simplicity), coord('sovereignty', displayed.sovereignty), coord('scalability', displayed.scalability)]
    .map(p => `${p.x.toFixed(1)},${p.y.toFixed(1)}`)
    .join(' '),
);
const dots = computed(() => [
  coord('simplicity', displayed.simplicity),
  coord('sovereignty', displayed.sovereignty),
  coord('scalability', displayed.scalability),
]);
const axisTips = [coord('simplicity', 1), coord('sovereignty', 1), coord('scalability', 1)];

function sectionStyle(key: keyof typeof profiles) {
  return {
    color: profiles[key].color,
    paddingLeft: active.value === key ? '1.75rem' : '0',
  };
}
</script>

<template>
  <section class="px-6">
    <div class="mx-auto grid max-w-[88rem] items-center gap-12 lg:grid-cols-2 lg:gap-20">
      <div>
        <h2 class="font-display text-5xl leading-tight text-neutral-800 md:text-6xl">
          We’re not saying the competition is bad…
        </h2>

        <p class="mt-8 text-2xl leading-relaxed text-neutral-700">
          …it just seems there are obvious tradeoffs for each alternative.
        </p>

        <div class="mt-8 space-y-5">
          <button type="button" class="block w-full cursor-pointer text-left text-2xl leading-relaxed transition-all duration-300" :style="sectionStyle('paas')" @click="select('paas')">
            <i class="devicon-vercel-plain mr-1.5 align-middle text-[1.1em]" aria-hidden="true" />Vercel,
            <i class="devicon-heroku-plain mr-1.5 align-middle text-[1.1em]" aria-hidden="true" />Heroku and
            <i class="devicon-railway-plain mr-1.5 align-middle text-[1.1em]" aria-hidden="true" />Railway are easy to use, but really the anti-christ of sovereignty.
          </button>

          <button type="button" class="block w-full cursor-pointer text-left text-2xl leading-relaxed transition-all duration-300" :style="sectionStyle('selfhost')" @click="select('selfhost')">
            Self-hosting tools like <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" class="mr-1.5 inline-block size-[1.05em] align-middle"><path d="M4.364 4.364V0h17.454v4.364zm0 13.09H0V4.365h4.364zm0 0h17.454v4.364H4.364ZM6.545 6.546v-1.7H22.3V2.182H24v4.363zm0 0v10.4h-1.7v-10.4Zm-2.663 11.39v1.7h-1.7v-1.7ZM24 24H6.545v-1.7H22.3v-2.664H24Z" /></svg>Coolify provide sovereignty but reach their limits once you need to scale beyond one server or grow your team.
          </button>

          <button type="button" class="block w-full cursor-pointer text-left text-2xl leading-relaxed transition-all duration-300" :style="sectionStyle('k8s')" @click="select('k8s')">
            <i class="devicon-kubernetes-plain mr-1.5 align-middle text-[1.1em]" aria-hidden="true" />Kubernetes scales, but requires you to have a dedicated platform team to run in production.
          </button>
        </div>

        <p class="mt-8 text-2xl leading-relaxed text-neutral-700">
          We call this the Software Deployment Trilemma.
          <NuxtLink to="/blog/the-software-deployment-trilemma" class="font-semibold text-neutral-800 underline decoration-2 underline-offset-4 hover:text-neutral-950">
            Read the blog post.
          </NuxtLink>
        </p>
      </div>

      <div class="flex justify-center">
        <div class="relative w-full max-w-lg">
          <svg viewBox="0 0 440 400" class="w-full" role="img" aria-label="Deployment trilemma chart">
            <g fill="none" stroke="#d2b78b">
              <polygon :points="tri(1)" stroke-width="1.25" />
              <polygon :points="tri(0.66)" stroke-opacity="0.55" />
              <polygon :points="tri(0.33)" stroke-opacity="0.55" />
              <line v-for="(tip, i) in axisTips" :key="i" :x1="cx" :y1="cy" :x2="tip.x" :y2="tip.y" stroke-opacity="0.55" />
            </g>

            <g :style="{ color: activeColor }">
              <polygon
                :points="dataPoly"
                fill="currentColor"
                fill-opacity="0.2"
                stroke="currentColor"
                stroke-width="2.5"
                stroke-linejoin="round"
              />
              <circle v-for="(d, i) in dots" :key="i" :cx="d.x" :cy="d.y" r="5" fill="currentColor" />
            </g>
          </svg>

          <span class="absolute left-1/2 top-[1%] -translate-x-1/2 text-2xl font-medium text-neutral-800">Simplicity</span>
          <span class="absolute left-[20.5%] top-[70%] -translate-x-1/2 text-2xl font-medium text-neutral-800">Scalability</span>
          <span class="absolute left-[79.5%] top-[70%] -translate-x-1/2 text-2xl font-medium text-neutral-800">Sovereignty</span>
        </div>
      </div>
    </div>
  </section>
</template>
