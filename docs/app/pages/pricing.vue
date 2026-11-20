<script setup lang="ts">
definePageMeta({
  layout: 'landing',
  header: false,
  footer: false,
});

useSeo({
  title: 'Pricing',
  description: 'Usage-based pricing that scales with you. Start free with self-hosting, or let us run it.',
  type: 'website',
});
</script>

<template>
  <div class="pricing-page">
    <PricingHero />

    <PricingPlans class="mt-24 sm:mt-38" />

    <PricingRates class="mt-24 sm:mt-38" />

    <PricingEstimator class="mt-24 sm:mt-38" />

    <PricingBills class="mt-24 sm:mt-38" />

    <PricingSelfHosted class="mt-24 sm:mt-38" />

    <PricingFaq class="mt-24 sm:mt-34" />

    <PricingEnding class="mt-24 pb-20 sm:mt-38 sm:pb-26" />
  </div>
</template>

<style scoped>
.pricing-page {
  overflow-x: clip;
}

.pricing-page :deep(.beats) {
  display: none;
}
</style>
