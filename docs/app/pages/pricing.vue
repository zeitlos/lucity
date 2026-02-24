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
  <UPageHero
    title="Simple, transparent pricing"
    description="Usage-based. No seat fees, no hidden costs. Pay for what you use, eject whenever you want."
  />

  <UContainer>
    <PricingCards />
  </UContainer>

  <UPageSection
    title="Resource pricing"
    description="Pay only for what your apps actually use. Here's a sample monthly bill."
  >
    <PricingMenuBoard />

    <PricingReceipt />
  </UPageSection>

  <UPageSection>
    <PricingSelfHost />
  </UPageSection>
</template>
