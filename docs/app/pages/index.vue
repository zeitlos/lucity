<script setup lang="ts">
definePageMeta({
  layout: 'default',
});

useSeo({
  title: 'Lucity: The PaaS you can leave',
  description: 'Open-source PaaS on Kubernetes. Git push to deploy, eject to standard Helm charts and ArgoCD configs at any time. Self-hostable, no lock-in.',
  type: 'website',
});
</script>

<template>
  <UPageHero>
    <template #title>
      The PaaS you can leave.
    </template>

    <template #description>
      Open-source PaaS on Kubernetes. Git push to deploy, environments out of the box, and a real exit door.
      <code>lucity eject</code> gives you standard Helm charts and ArgoCD configs.
    </template>

    <template #links>
      <UButton
        to="/getting-started/quick-start"
        color="neutral"
        size="xl"
        trailing-icon="i-lucide-arrow-right"
      >
        Get Started
      </UButton>

      <UButton
        to="https://github.com/zeitlos/lucity"
        target="_blank"
        variant="outline"
        icon="i-simple-icons-github"
        size="xl"
        color="neutral"
      >
        View on GitHub
      </UButton>
    </template>
  </UPageHero>

  <HeroDemo />

  <UPageSection class="mt-16 sm:mt-24">
    <template #title>
      Everything you need to ship
    </template>

    <template #description>
      All the building blocks for deploying and running your apps on Kubernetes. Built on standard tools, so you can eject whenever you want.
    </template>

    <BentoGrid />
  </UPageSection>

  <UPageSection>
    <template #title>
      Built on tools you already know
    </template>

    <template #description>
      The enterprise way to deploy Kubernetes apps, cleverly automated. No proprietary runtime, no black boxes — just standard tools, loosely coupled, with zero platform state to babysit.
    </template>

    <ToolsArchitecture />
  </UPageSection>

  <UPageSection>
    <SectionVS />
  </UPageSection>
</template>
