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
    <div class="menu-board">
      <div class="menu-row">
        <span class="menu-name">vCPU</span>
        <span class="menu-dots" />
        <span class="menu-price">CHF 10<span class="menu-unit"> / core / mo</span></span>
      </div>
      <div class="menu-row">
        <span class="menu-name">Memory</span>
        <span class="menu-dots" />
        <span class="menu-price">CHF 5<span class="menu-unit"> / GB / mo</span></span>
      </div>
      <div class="menu-row">
        <span class="menu-name">Disk</span>
        <span class="menu-dots" />
        <span class="menu-price">CHF 0.10<span class="menu-unit"> / GB / mo</span></span>
      </div>
      <div class="menu-row">
        <span class="menu-name">Egress</span>
        <span class="menu-dots" />
        <span class="menu-price">CHF 0.02<span class="menu-unit"> / GB</span></span>
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
/* ---- Menu board — restaurant-style price list ---- */

.menu-board {
  max-width: 380px;
  margin: 0 auto 32px;
  font-family: var(--font-mono);
  font-size: 13px;
  line-height: 1.7;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.menu-row {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.menu-name {
  font-weight: 600;
  white-space: nowrap;
  color: var(--ui-text);
}

.menu-dots {
  flex: 1;
  border-bottom: 1.5px dotted var(--ui-text-muted);
  min-width: 20px;
  position: relative;
  top: -3px;
  opacity: 0.4;
}

.menu-price {
  white-space: nowrap;
  font-weight: 600;
  color: var(--ui-text);
}

.menu-unit {
  font-weight: 400;
  font-size: 11px;
  color: var(--ui-text-muted);
}
</style>
