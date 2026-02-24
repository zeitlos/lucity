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
    <div class="resource-prices">
      <div class="resource-item">
        <span class="resource-label">vCPU</span>
        <span class="resource-price">CHF 10 <span class="resource-unit">/ core / mo</span></span>
      </div>
      <div class="resource-item">
        <span class="resource-label">Memory</span>
        <span class="resource-price">CHF 5 <span class="resource-unit">/ GB / mo</span></span>
      </div>
      <div class="resource-item">
        <span class="resource-label">Disk</span>
        <span class="resource-price">CHF 0.10 <span class="resource-unit">/ GB / mo</span></span>
      </div>
      <div class="resource-item">
        <span class="resource-label">Egress</span>
        <span class="resource-price">CHF 0.02 <span class="resource-unit">/ GB</span></span>
      </div>
    </div>

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

<style scoped>
.resource-prices {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px 16px;
  max-width: 420px;
  margin: 0 auto 32px;
  font-family: var(--font-mono);
}

@media (min-width: 640px) {
  .resource-prices {
    grid-template-columns: repeat(4, 1fr);
    max-width: 640px;
  }
}

.resource-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 12px 8px;
  border-radius: 8px;
  background: var(--ui-bg-elevated);
  border: 1px solid var(--ui-border);
}

.resource-label {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--ui-text-muted);
}

.resource-price {
  font-size: 15px;
  font-weight: 600;
  color: var(--ui-text);
}

.resource-unit {
  font-size: 11px;
  font-weight: 400;
  color: var(--ui-text-muted);
}
</style>
