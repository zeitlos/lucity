<script setup lang="ts">
const baseUi = {
  root: 'pricing-card rounded-2xl bg-gradient-to-b from-[var(--ui-bg-elevated)]/60 to-transparent relative overflow-hidden',
  title: 'font-serif text-3xl sm:text-4xl font-normal',
};

const sideUi = {
  root: `${baseUi.root} shadow-[0_2px_20px_-4px_oklch(0_0_0/0.07),0_1px_3px_-1px_oklch(0_0_0/0.04)]`,
  title: baseUi.title,
};

const proUi = {
  root: 'pricing-card rounded-2xl bg-[var(--ui-bg-elevated)] relative overflow-hidden shadow-[0_4px_32px_-6px_oklch(0_0_0/0.12),0_2px_6px_-2px_oklch(0_0_0/0.06)]',
  title: baseUi.title,
};

const starter = {
  title: 'Starter',
  description: 'For side projects and MVPs.',
  price: 'CHF 5',
  billingCycle: '/month',
  variant: 'outline' as const,
  tagline: 'CHF 5 in credits included',
  features: [
    'Up to 3 projects',
    'EU region',
    'Email support (best-effort)',
    'Community GitHub access',
  ],
  button: {
    label: 'Join the waitlist',
    to: '/cloud',
    color: 'neutral' as const,
  },
  ui: sideUi,
};

const pro = {
  title: 'Pro',
  description: 'For small teams and agencies.',
  price: 'CHF 20',
  billingCycle: '/month',
  highlight: true,
  tagline: 'CHF 20 in credits included',
  features: [
    'Unlimited projects',
    'EU region',
    'Email support (1 business day)',
    'Priority issue handling',
  ],
  button: {
    label: 'Join the waitlist',
    to: '/cloud',
  },
  ui: proUi,
};

const business = {
  title: 'Business',
  description: 'For regulated industries and compliance needs.',
  price: 'CHF 50',
  billingCycle: '/month',
  variant: 'outline' as const,
  tagline: 'CHF 50 in credits included',
  features: [
    'Unlimited projects',
    'EU region + Swiss data residency (coming soon)',
    'Email + Slack support (4h response)',
    'FADP/nDSG compliance ready',
  ],
  button: {
    label: 'Contact us',
    to: 'mailto:hello@zeitlos.software',
    color: 'neutral' as const,
  },
  ui: sideUi,
};
</script>

<template>
  <div class="pricing-stack">
    <div class="pricing-side pricing-left">
      <UPricingPlan v-bind="starter" />
    </div>
    <div class="pricing-center">
      <UPricingPlan v-bind="pro" />
    </div>
    <div class="pricing-side pricing-right">
      <UPricingPlan v-bind="business" />
    </div>
  </div>
</template>

<style scoped>
.pricing-stack {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

@media (min-width: 1024px) {
  .pricing-stack {
    flex-direction: row;
    align-items: stretch;
    justify-content: center;
    gap: 0;
  }

  /* Cards fill their wrapper height */
  .pricing-side :deep(.pricing-card),
  .pricing-center :deep(.pricing-card) {
    height: 100%;
  }

  .pricing-side {
    flex: 1 1 0%;
    position: relative;
    z-index: 1;
  }

  .pricing-left {
    margin-right: -16px;
  }

  .pricing-right {
    margin-left: -16px;
  }

  .pricing-center {
    flex: 1 1 0%;
    position: relative;
    z-index: 2;
    transform: scale(1.05);
  }
}

/* Dot pattern overlay on all pricing cards */
:deep(.pricing-card)::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: radial-gradient(circle, oklch(1 0 0) 0.7px, transparent 0.7px);
  background-size: 16px 16px;
  opacity: 0.07;
  pointer-events: none;
  z-index: 1;
}
</style>
