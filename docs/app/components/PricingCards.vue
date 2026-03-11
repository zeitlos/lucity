<script setup lang="ts">
const baseUi = {
  root: 'pricing-card rounded-2xl bg-gradient-to-b from-[var(--ui-bg-elevated)]/60 to-transparent relative overflow-hidden',
  title: 'font-serif text-3xl sm:text-4xl font-normal',
};

const hobbyUi = {
  root: `${baseUi.root} shadow-[0_2px_20px_-4px_oklch(0_0_0/0.07),0_1px_3px_-1px_oklch(0_0_0/0.04)]`,
  title: baseUi.title,
};

const proUi = {
  root: 'pricing-card rounded-2xl bg-gradient-to-b from-[var(--ui-bg-elevated)] to-[var(--ui-bg)] relative overflow-hidden shadow-[0_4px_32px_-6px_oklch(0_0_0/0.12),0_2px_6px_-2px_oklch(0_0_0/0.06)]',
  title: baseUi.title,
};

const hobby = {
  title: 'Hobby',
  description: 'For side projects and experiments.',
  price: '€5',
  billingCycle: '/month',
  variant: 'outline' as const,
  tagline: '€5 in credits included',
  features: [
    'Just you',
    '1 custom domain',
    'Eco resources',
    'Production resources',
    'Community support',
  ],
  button: {
    label: 'Get Started',
    to: 'https://lucity.cloud/app/login',
    color: 'neutral' as const,
  },
  ui: hobbyUi,
};

const pro = {
  title: 'Pro',
  description: 'For teams shipping to production.',
  price: '€25',
  billingCycle: '/month',
  highlight: true,
  tagline: '€25 in credits included',
  features: [
    'Unlimited collaborators',
    'Unlimited custom domains',
    'Eco resources',
    'Production resources',
    'FADP/nDSG compliance ready',
    'Priority support',
  ],
  button: {
    label: 'Get Started',
    to: 'https://lucity.cloud/app/login',
  },
  ui: proUi,
};
</script>

<template>
  <div class="pricing-stack">
    <div class="pricing-card-wrapper">
      <UPricingPlan v-bind="hobby" />
    </div>
    <div class="pricing-card-wrapper">
      <UPricingPlan v-bind="pro" />
    </div>
  </div>
</template>

<style scoped>
.pricing-stack {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-width: 720px;
  margin: 0 auto;
}

@media (min-width: 768px) {
  .pricing-stack {
    flex-direction: row;
    align-items: stretch;
    gap: 24px;
  }

  .pricing-card-wrapper {
    flex: 1 1 0%;
  }

  .pricing-card-wrapper :deep(.pricing-card) {
    height: 100%;
  }
}

/* Dot pattern overlay on all pricing cards */
:deep(.pricing-card)::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: radial-gradient(circle, currentColor 1px, transparent 1px);
  background-size: 14px 14px;
  opacity: 0.07;
  pointer-events: none;
  z-index: 1;
}
</style>
