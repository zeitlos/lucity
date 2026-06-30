<script setup lang="ts">
definePageMeta({
  layout: 'default',
});

useSeo({
  title: 'Pricing',
  description: 'Usage-based pricing that scales with you. Start free with self-hosting, or let us run it.',
  type: 'website',
});
</script>

<template>
  <div class="pricing-page">
    <section class="mx-auto max-w-page px-6 pt-24 pb-12 text-center">
      <h1 class="font-display text-6xl leading-tight text-neutral-800 md:text-7xl dark:text-neutral-100">
        Simple, transparent pricing
      </h1>
      <p class="mx-auto mt-8 max-w-3xl text-2xl leading-relaxed text-neutral-700 dark:text-neutral-300">
        Two plans, usage-based resources: pay only for what you use. Each plan includes a monthly credit allowance applied automatically to your bill everymonth.
      </p>
    </section>

    <section class="mx-auto max-w-page px-6">
      <PricingCards />
    </section>

    <section class="mx-auto max-w-page px-6 mt-28 md:mt-36">
      <div class="mx-auto mb-16 max-w-3xl text-center">
        <h2 class="font-display text-5xl leading-tight text-neutral-800 md:text-6xl dark:text-neutral-100">
          Resource pricing
        </h2>
        <p class="mt-6 text-xl leading-relaxed text-neutral-600 dark:text-neutral-400">
          Two ways to pay for compute. <strong class="font-semibold text-neutral-800 dark:text-neutral-200">On-demand</strong> is fully metered, so you pay only for what you use, with burstable resources that suit development and spiky workloads. <strong class="font-semibold text-neutral-800 dark:text-neutral-200">Guaranteed</strong> reserves your capacity for consistent performance with no noisy neighbors, predictable billing, and 20% off vCPU and memory.
        </p>
      </div>

      <PricingMenuBoard />
    </section>

    <section class="mx-auto max-w-page px-6 mt-28 md:mt-36">
      <div class="mx-auto mb-16 max-w-3xl text-center">
        <h2 class="font-display text-5xl leading-tight text-neutral-800 md:text-6xl dark:text-neutral-100">
          Sample monthly bill
        </h2>
        <p class="mt-6 text-xl leading-relaxed text-neutral-600 dark:text-neutral-400">
          Here's what a typical Hobby plan bill looks like.
        </p>
      </div>

      <PricingReceipt />
    </section>

    <section class="mx-auto max-w-page px-6 mt-28 mb-28 md:mt-36 md:mb-36">
      <PricingSelfHost />
    </section>
  </div>
</template>

<style scoped>
.pricing-page {
  overflow-x: clip;
}
</style>
