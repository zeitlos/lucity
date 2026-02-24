<script setup lang="ts">
const plans = [
  {
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
  },
  {
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
  },
  {
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
  },
];
</script>

<template>
  <UPricingPlans scale>
    <UPricingPlan
      v-for="plan in plans"
      :key="plan.title"
      v-bind="plan"
    />
  </UPricingPlans>
</template>
