<script setup lang="ts">
const groups = [
  { to: '/features#builds', icon: 'i-lucide-hammer', label: 'Builds' },
  { to: '/features#deploys', icon: 'i-lucide-rocket', label: 'Deploys' },
  { to: '/features#services', icon: 'i-lucide-boxes', label: 'Services' },
  { to: '/features#domains', icon: 'i-lucide-globe', label: 'Domains' },
  { to: '/features#postgres', icon: 'i-lucide-database', label: 'Postgres' },
  { to: '/features#storage', icon: 'i-lucide-hard-drive', label: 'Storage' },
  { to: '/features#observe', icon: 'i-lucide-activity', label: 'Logs and metrics' },
  { to: '/features#tools', icon: 'i-lucide-wrench', label: 'Tools' },
  { to: '/features#team', icon: 'i-lucide-users', label: 'Team' },
  { to: '/features#billing', icon: 'i-lucide-receipt', label: 'Billing' },
  { to: '/features#security', icon: 'i-lucide-shield-check', label: 'Security' },
  { to: '/features#ownership', icon: 'i-lucide-door-open', label: 'Ownership' },
];

const depth = [
  { to: '/security', icon: 'i-lucide-shield-check', title: 'Security', text: 'Guard rails, on by default.' },
  { to: '/no-lock-in', icon: 'i-lucide-door-open', title: 'No lock-in', text: 'How leaving works.' },
  { to: '/agents', icon: 'i-lucide-bot', title: 'Agent hosting', text: 'Agents, next to your app.' },
];
</script>

<template>
  <NavPanel single>
    <div>
      <p class="mb-1.5 px-3 text-sm text-stone-600 dark:text-stone-400">Every part, on one page</p>
      <div class="grid grid-cols-3 gap-x-1">
        <NavMini v-for="item in groups" :key="item.to" v-bind="item" />
      </div>
      <div class="mt-1 border-t border-stone-200 pt-3.5 dark:border-stone-800">
        <p class="mb-1.5 px-3 text-sm text-stone-600 dark:text-stone-400">In depth</p>
        <div class="grid grid-cols-3 gap-x-1">
          <NavRow v-for="item in depth" :key="item.to" v-bind="item" tint="rose" />
        </div>
      </div>
    </div>

    <template #foot>
      <span class="text-stone-700 dark:text-stone-400">New here? The quickstart takes about five minutes.</span>
      <NuxtLink
        to="/docs/quickstart"
        class="rounded-sm text-[15px] font-semibold text-stone-900 underline decoration-1 underline-offset-4 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-(--nav-accent) dark:text-stone-100"
      >
        Open the quickstart
      </NuxtLink>
    </template>

    <template #stage>
      <NavStage to="/features" tint="keep" action="See every feature">
        <img src="/img/landing/sketch-releases.svg" alt="" class="mx-auto mb-2.5 block h-auto w-full max-w-55 dark:hidden">
        <img src="/img/landing/sketch-releases-night.svg" alt="" class="mx-auto mb-2.5 hidden h-auto w-full max-w-55 dark:block">
        <span class="font-display text-[25px]/[1.15]">The parts you need to run what you build.</span>
      </NavStage>
    </template>
  </NavPanel>
</template>
