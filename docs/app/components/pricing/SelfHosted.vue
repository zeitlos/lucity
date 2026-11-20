<script setup lang="ts">
const facts = [
  { name: 'Features', value: 'All of them' },
  { name: 'Fees', value: 'None' },
  { name: 'Who runs it', value: 'You do' },
  { name: 'Support', value: 'Community' },
];
</script>

<template>
  <section class="mx-auto max-w-content px-6" aria-labelledby="self">
    <div class="night flex flex-wrap items-center gap-x-20 gap-y-10 px-6 py-11 sm:px-16 sm:py-15">
      <div class="min-w-0 flex-[1.2_1_480px]">
        <h2 id="self" class="font-display text-4xl/11 text-balance text-[oklch(0.97_0.015_80)] md:text-5xl/14 lg:text-6xl/17">Or run it yourself, for free.</h2>
        <p class="mt-5.5 max-w-[600px] text-lg text-[oklch(0.88_0.02_80)] sm:text-[19px]">
          Lucity is open source under AGPL-3.0. The self-hosted platform has every feature, charges no fees and reports nothing back. You bring a Kubernetes cluster, and an afternoon for the first install.
        </p>
        <p class="mt-7 text-lg">
          <NuxtLink to="/docs/self-hosting" class="inline-flex items-center gap-2 font-semibold text-[oklch(0.97_0.015_80)] underline decoration-1 underline-offset-4">
            Read the self-hosting guide
            <UIcon name="i-lucide-arrow-right" class="size-4.5 shrink-0" />
          </NuxtLink>
        </p>
      </div>
      <ul class="facts min-w-0 flex-[1_1_360px]">
        <li v-for="fact in facts" :key="fact.name"><span>{{ fact.name }}</span><em>{{ fact.value }}</em></li>
      </ul>
    </div>
  </section>
</template>

<style scoped>
.night {
  position: relative;
  overflow: hidden;
  border-radius: 24px;
  background-color: oklch(0.20 0.03 55);
  background-image: radial-gradient(circle at 1px 1px, oklch(1 0 0 / .07) 1px, transparent 0);
  background-size: 22px 22px;
  box-shadow: 0 28px 56px -22px oklch(0.25 0.04 55 / .42);
  color: oklch(0.88 0.02 80);
}

.dark .night {
  background-color: oklch(0.24 0.03 55);
  box-shadow: inset 0 0 0 1px oklch(1 0 0 / .08);
}

.facts li {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 0;
  border-top: 1px solid oklch(0.30 0.03 55);
  font-size: 17px;
  color: oklch(0.97 0.01 80);
}

.facts li:last-child {
  border-bottom: 1px solid oklch(0.30 0.03 55);
}

.dark .facts li {
  border-color: oklch(0.34 0.03 55);
}

.facts em {
  font-size: 16px;
  font-style: normal;
  font-weight: 500;
  color: oklch(0.88 0.02 80);
  white-space: nowrap;
}
</style>
