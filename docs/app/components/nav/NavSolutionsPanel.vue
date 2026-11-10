<script setup lang="ts">
const useCases = [
  { to: '/use-cases/startups', icon: 'i-lucide-rocket', title: 'Startups and small teams', text: 'The simplicity you came for, without the rebuild when you outgrow it.' },
  { to: '/use-cases/agencies', icon: 'i-lucide-briefcase', title: 'Agencies', text: 'Ship client projects fast, then hand over infrastructure they own.' },
  { to: '/use-cases/platform-teams', icon: 'i-lucide-layers', title: 'Platform teams', text: 'Self-service for developers, on the cluster you already run.' },
];

const platforms = [
  { to: '/comparisons/vs-heroku', icon: 'i-simple-icons-heroku', label: 'Heroku' },
  { to: '/comparisons/vs-railway', icon: 'i-simple-icons-railway', label: 'Railway' },
  { to: '/comparisons/vs-render', icon: 'i-simple-icons-render', label: 'Render' },
  { to: '/comparisons/vs-vercel', icon: 'i-simple-icons-vercel', label: 'Vercel' },
  { to: '/comparisons/vs-fly-io', icon: 'i-simple-icons-flydotio', label: 'Fly.io' },
  { to: '/comparisons/vs-coolify', icon: 'i-simple-icons-coolify', label: 'Coolify' },
];
</script>

<template>
  <NavPanel>
    <div>
      <p class="mb-1.5 px-3 text-sm text-stone-600 dark:text-stone-400">Use cases</p>
      <NavRow v-for="item in useCases" :key="item.to" v-bind="item" tint="violet" />
    </div>
    <div>
      <p class="mb-1.5 px-3 text-sm text-stone-600 dark:text-stone-400">Switching from</p>
      <div class="grid grid-cols-2 gap-x-1">
        <NavMini v-for="item in platforms" :key="item.to" v-bind="item" />
      </div>
    </div>

    <template #foot>
      <span class="text-stone-700 dark:text-stone-400">Every comparison says where the other platform is the better pick.</span>
      <NuxtLink
        to="/comparisons"
        class="rounded-sm text-[15px] font-semibold text-stone-900 underline decoration-1 underline-offset-4 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-(--nav-accent) dark:text-stone-100"
      >
        All comparisons
      </NuxtLink>
    </template>

    <template #stage>
      <NavStage to="/agents" tint="joy" action="See how it works">
        <img src="/img/landing/robot.svg" alt="" width="132" height="123" class="-mt-1.5 mb-1.5 -ml-2.5 block h-auto w-33">
        <span class="font-display text-[25px]/[1.15]">Agent hosting</span>
        <span class="mt-2 text-[15px]/[1.45] text-[oklch(0.36_0.04_300)] dark:text-[oklch(0.84_0.04_300)]">Host the agents you already run, right next to your app.</span>
      </NavStage>
    </template>
  </NavPanel>
</template>
