<script setup lang="ts">
const rails = [
  { name: 'Leaked keys', text: 'Every deploy checks your repository and its history, and flags keys that still work.' },
  { name: 'Vulnerabilities', text: 'Apps Lucity built are rechecked every two hours while they run.' },
  { name: 'Isolation', text: "Apps in one workspace can't reach another's apps or databases." },
];
</script>

<template>
  <section class="band" aria-labelledby="guard">
    <div class="sunset" aria-hidden="true" />
    <img class="road" src="/img/landing/guard-rail-bridge.svg" alt="A van has slid wide in the bend of a bridge, and the guard rail keeps it on the road" width="640" height="540">
    <div class="relative mx-auto flex max-w-content flex-wrap items-end gap-x-16 gap-y-12 px-6">
      <div class="min-w-0 flex-[1_1_400px] lg:pb-28">
        <p class="mb-3.5 text-base font-semibold text-[oklch(0.84_0.12_62)]">Comes with every project</p>
        <h2 id="guard" class="font-display text-4xl/tight text-[oklch(0.97_0.015_80)] lg:text-5xl/tight">Guard rails, on by default.</h2>
        <p class="mt-4.5 max-w-[520px] text-lg text-[oklch(0.84_0.02_78)]">Nothing to install and nothing to configure, on every plan.</p>
        <div class="mt-7 max-w-[560px] border-b border-white/15">
          <div v-for="rail in rails" :key="rail.name" class="border-t border-white/15 py-4.5">
            <b class="block text-lg font-semibold text-[oklch(0.96_0.015_80)]">{{ rail.name }}</b>
            <span class="mt-1 block text-base text-[oklch(0.80_0.02_78)]">{{ rail.text }}</span>
          </div>
        </div>
        <p class="mt-6 text-lg">
          <NuxtLink to="/features#security" class="inline-flex items-center gap-2 font-semibold text-[oklch(0.97_0.015_80)] underline decoration-1 underline-offset-4">
            How the guard rails work
            <UIcon name="i-lucide-arrow-right" class="size-4.5 shrink-0" />
          </NuxtLink>
        </p>
      </div>
      <div class="relative min-w-0 flex-[1.45_1_520px]">
        <div class="bezel">
          <div>
            <img src="/img/landing/security.webp" alt="The Security tab of a service: vulnerabilities found in the deployed image by severity, and the result of the latest scan for leaked secrets">
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.band {
  position: relative;
  padding-top: 112px;
  overflow: hidden;
  background-color: oklch(0.20 0.03 55);
  background-image: radial-gradient(circle at 1px 1px, oklch(1 0 0 / .07) 1px, transparent 0);
  background-size: 22px 22px;
  color: oklch(0.84 0.02 78);
}

.road {
  position: absolute;
  top: 0;
  right: 0;
  display: block;
  width: min(62vw, 840px);
  height: auto;
  pointer-events: none;
}

.sunset {
  position: absolute;
  right: 0;
  bottom: 0;
  width: min(66vw, 950px);
  height: 640px;
  pointer-events: none;
  filter: blur(30px);
  background:
    radial-gradient(ellipse 30% 30% at 80% 50%, oklch(0.82 0.14 58 / .8), oklch(0.82 0.14 58 / 0) 100%),
    radial-gradient(ellipse 34% 30% at 52% 52%, oklch(0.70 0.21 0 / .62), oklch(0.70 0.21 0 / 0) 100%),
    radial-gradient(ellipse 25% 30% at 31% 55%, oklch(0.46 0.23 282 / .85), oklch(0.46 0.23 282 / 0) 100%);
}

.bezel {
  position: relative;
  box-sizing: border-box;
  width: 100%;
  max-width: 800px;
  margin: 0 auto;
  padding: 8px 8px 0;
  border: 1px solid oklch(1 0 0 / .26);
  border-bottom: 0;
  border-radius: 20px 20px 0 0;
  background: oklch(1 0 0 / .08);
}

.bezel > div {
  max-height: 350px;
  overflow: hidden;
  border-radius: 12px 12px 0 0;
  background: oklch(0.19 0.02 55);
  box-shadow: 0 0 0 1px oklch(1 0 0 / .14);
}

.bezel img {
  display: block;
  width: 100%;
  height: auto;
}

@media (max-width: 900px) {
  .band {
    padding-top: calc(66vw + 16px);
  }

  .road {
    width: 100vw;
  }
}
</style>
