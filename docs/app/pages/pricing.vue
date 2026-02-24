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
    description="Here's what a typical small app costs per month."
  >
    <PricingReceipt />
  </UPageSection>

  <UPageSection>
    <UPageCard variant="soft">
      <template #title>
        Prefer to self-host?
      </template>

      <template #description>
        Lucity is open source under the AGPL-3.0 license. Run it on your own Kubernetes cluster with all features included. No limits, no license fees, no strings attached.
        <div class="mt-4">
          <UButton
            to="/getting-started/quick-start"
            color="neutral"
            trailing-icon="i-lucide-arrow-right"
          >
            Self-hosting guide
          </UButton>
        </div>
      </template>
    </UPageCard>
  </UPageSection>
</template>
