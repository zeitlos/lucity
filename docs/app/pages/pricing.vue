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
    <UPageHero
    title="Simple, transparent pricing"
    description="Two plans, usage-based resources. Pay for what you use, eject whenever you want."
  />

  <UContainer>
    <PricingCards />

    <div class="explainer">
      <h2 class="explainer-title">
        How credits work
      </h2>
      <p class="explainer-body">
        Every plan includes a monthly credit allowance that covers your resource usage.
        Hobby comes with €5, Pro with €25. Credits are applied automatically to your
        resource bill each month. If your usage is lower than your credits, you pay
        nothing extra. No rollover, no surprises.
      </p>
    </div>
  </UContainer>

  <UPageSection>
    <template #title>
      Resource pricing
    </template>
    <template #description>
      Two tiers, pick what fits. Eco for experiments, Production for the real thing.
    </template>

    <PricingMenuBoard />

    <div class="explainer explainer-cols">
      <div class="explainer-col">
        <h2 class="explainer-title">
          Eco
        </h2>
        <p class="explainer-body">
          Shared infrastructure in the EU. Great for side projects, staging
          environments, and anything that doesn't need guaranteed resources.
          Metered per vCPU and GB, you only pay for what you use.
        </p>
      </div>
      <div class="explainer-col">
        <h2 class="explainer-title">
          Production
        </h2>
        <p class="explainer-body">
          Dedicated resources with predictable monthly pricing. No metering
          surprises, no noisy neighbors. Perfect for agencies and teams that
          need stable, pass-through-friendly bills. Swiss data residency coming soon.
        </p>
      </div>
    </div>
  </UPageSection>

  <UPageSection>
    <template #title>
      Sample monthly bill
    </template>
    <template #description>
      Here's what a typical Hobby plan bill looks like.
    </template>

    <PricingReceipt />
  </UPageSection>

  <UPageSection>
    <PricingSelfHost />
  </UPageSection>
  </div>
</template>

<style scoped>
.pricing-page {
  overflow-x: clip;
}

.explainer {
  max-width: 640px;
  margin: 2.5rem auto 0;
  padding: 1.5rem 2rem;
  border: 1px solid var(--ui-border);
  border-radius: 12px;
  background: var(--ui-bg-elevated);
  text-align: left;
}

.explainer-title {
  font-family: var(--font-serif);
  font-size: 1.25rem;
  font-weight: 600;
  margin-bottom: 0.75rem;
  color: var(--ui-text-highlighted);
}

.explainer-body {
  font-size: 0.9375rem;
  line-height: 1.7;
  color: var(--ui-text-muted);
  margin-bottom: 0.75rem;
}

.explainer-body:last-child {
  margin-bottom: 0;
}

.explainer-body :deep(strong) {
  color: var(--ui-text-highlighted);
  font-weight: 600;
}

.explainer-cols {
  max-width: 760px;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  text-align: left;
}

@media (min-width: 640px) {
  .explainer-cols {
    flex-direction: row;
    gap: 2rem;
  }

  .explainer-col {
    flex: 1 1 0%;
  }

  .explainer-col + .explainer-col {
    border-left: 1px solid var(--ui-border);
    padding-left: 2rem;
  }
}
</style>
