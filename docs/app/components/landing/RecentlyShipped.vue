<script setup lang="ts">
import releases from '~/assets/landing/releases.json';
</script>

<template>
  <section class="desk py-20 sm:pt-28 sm:pb-31" aria-label="Recently shipped, and a note from the founder">
    <div class="mx-auto flex max-w-content flex-wrap items-start gap-x-18 gap-y-14 px-6">
      <div class="min-w-0 flex-[1_1_380px]">
        <h2 class="font-display text-3xl/tight text-stone-900 lg:text-4xl/tight dark:text-stone-100">Recently shipped</h2>
        <div class="timeline mt-7">
          <div v-for="release in releases" :key="release.date" class="timeline-group">
            <div class="timeline-date text-sm font-semibold text-stone-900 dark:text-stone-100">{{ release.date }}</div>
            <div v-for="item in release.items" :key="item.name" class="pb-4">
              <b class="block text-lg/6 font-semibold text-stone-900 dark:text-stone-100">{{ item.name }}</b>
              <span class="mt-0.5 block text-base text-stone-700 dark:text-stone-400">{{ item.text }}</span>
            </div>
          </div>
        </div>
        <a href="https://github.com/zeitlos/lucity/releases" class="mt-2.5 ml-26 inline-block text-base font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">All releases</a>
      </div>

      <div class="letter min-w-0 flex-[1.25_1_520px] -rotate-[0.5deg] bg-white px-6 py-7 sm:px-13 sm:py-12">
        <h2 class="mb-4.5 font-display text-2xl/tight text-stone-900">Why I built Lucity</h2>
        <div class="flex flex-col gap-3.5 text-lg/relaxed font-light text-[oklch(0.22_0.02_55)] lg:text-xl/relaxed">
          <p>In 13 years as a software engineer I've deployed software three ways. Each got one thing right, and each made me give something up.</p>
          <p>Copying files to a server under my desk was tedious, but <b class="font-bold text-[oklch(0.58_0.20_0)]">it was truly mine.</b></p>
          <p>Heroku and Vercel shipped my code straight from the repository. <b class="font-bold text-[oklch(0.55_0.15_300)]">A joy to use</b>, paid for with control over my own setup.</p>
          <p>Kubernetes ran thousands of containers. <b class="font-bold text-[oklch(0.58_0.17_35)]">Built to grow</b>, and too complicated for anyone who just wants to ship.</p>
          <p class="font-semibold">I built Lucity to be all three at once.</p>
          <p>Deploy your first app, use it for a bit, then write and tell me what you think.</p>
        </div>
        <img src="/img/signature-christian.png" alt="Christian's signature" class="mt-6 block h-13 w-auto">
        <div class="mt-3 flex items-center gap-3">
          <img src="/img/founder-avatar.png" alt="" width="44" height="44" class="size-11 rounded-full">
          <div class="text-lg/snug text-[oklch(0.22_0.02_55)]">
            <b>Christian Blättler</b>, <a href="mailto:christian@zeitlos.software" class="underline underline-offset-3">christian<wbr>@zeitlos.software</a><br>
            <span class="text-stone-700 italic">Creator of Lucity</span>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.desk {
  background-color: oklch(0.915 0.027 76);
  background-image: radial-gradient(circle at 1px 1px, oklch(0.25 0.04 55 / .13) 1px, transparent 0);
  background-size: 22px 22px;
  box-shadow: inset 0 1px 0 oklch(0.25 0.04 55 / .10), inset 0 -1px 0 oklch(0.25 0.04 55 / .10);
  --tick: oklch(0.70 0.03 70);
  --tick-major: oklch(0.60 0.03 65);
  --tick-mark: oklch(0.25 0.03 55);
}

.dark .desk {
  background-color: oklch(0.17 0.025 55);
  background-image: radial-gradient(circle at 1px 1px, oklch(1 0 0 / .07) 1px, transparent 0);
  box-shadow: inset 0 1px 0 oklch(1 0 0 / .06), inset 0 -1px 0 oklch(1 0 0 / .06);
  --tick: oklch(0.45 0.03 55);
  --tick-major: oklch(0.55 0.03 60);
  --tick-mark: oklch(0.90 0.02 80);
}

.timeline {
  position: relative;
  padding-left: 104px;
}

.timeline::before {
  content: "";
  position: absolute;
  left: 64px;
  top: 0;
  bottom: 0;
  width: 10px;
  border-right: 1px solid var(--tick);
  background: repeating-linear-gradient(to bottom, var(--tick) 0 1px, transparent 1px 8px);
}

.timeline::after {
  content: "";
  position: absolute;
  left: 56px;
  top: 0;
  bottom: 0;
  width: 18px;
  background: repeating-linear-gradient(to bottom, var(--tick-major) 0 1px, transparent 1px 40px);
}

.timeline-group {
  position: relative;
  padding-bottom: 28px;
}

.timeline-group:last-child {
  padding-bottom: 8px;
}

.timeline-group::before {
  content: "";
  position: absolute;
  z-index: 1;
  left: -56px;
  top: 11px;
  width: 26px;
  height: 2px;
  background: var(--tick-mark);
}

.timeline-date {
  position: absolute;
  left: -104px;
  top: 0;
  width: 42px;
  text-align: right;
  line-height: 24px;
  white-space: nowrap;
}

.letter {
  box-shadow: 0 24px 40px -24px oklch(0.25 0.04 55 / .35);
}
</style>
