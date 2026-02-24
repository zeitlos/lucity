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

  <UPageSection
    title="Lucity Cloud"
    description="Managed hosting in the EU. Same open-source platform, zero infrastructure to maintain."
  >
    <PricingCards />
  </UPageSection>

  <UPageSection
    title="Resource pricing"
    description="Credits are consumed based on resource usage. Simple per-unit rates, no surprises."
  >
    <div class="mx-auto max-w-lg">
      <table class="w-full text-left">
        <thead>
          <tr class="border-b border-default">
            <th class="pb-3 font-medium text-muted">
              Resource
            </th>
            <th class="pb-3 font-medium text-muted text-right">
              Rate
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-default">
          <tr>
            <td class="py-3 font-medium">
              vCPU
            </td>
            <td class="py-3 text-right text-muted">
              CHF 10 / vCPU / month
            </td>
          </tr>
          <tr>
            <td class="py-3 font-medium">
              RAM
            </td>
            <td class="py-3 text-right text-muted">
              CHF 5 / GB / month
            </td>
          </tr>
          <tr>
            <td class="py-3 font-medium">
              Storage
            </td>
            <td class="py-3 text-right text-muted">
              CHF 0.10 / GB / month
            </td>
          </tr>
          <tr>
            <td class="py-3 font-medium">
              Egress
            </td>
            <td class="py-3 text-right text-muted">
              CHF 0.02 / GB
            </td>
          </tr>
        </tbody>
      </table>

      <p class="mt-6 text-sm text-muted text-center text-balance">
        A typical small app (0.5 vCPU, 512 MB RAM, 10 GB storage) runs for about <strong class="text-highlighted">CHF 8/month</strong>.
        <br>
        Swiss region pricing will be available with the Business tier launch.
      </p>
    </div>
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
