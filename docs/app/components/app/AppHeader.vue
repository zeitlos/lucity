<script setup lang="ts">
const appConfig = useAppConfig();
const site = useSiteConfig();

const appUrl = 'https://lucity.cloud/app';

const navItems = [
  {
    label: 'Features',
    children: [
      {
        label: 'Push to deploy',
        description: 'Connect GitHub and ship with git push.',
        icon: 'i-lucide-rocket',
        to: '/features/builds',
      },
      {
        label: 'Environments',
        description: 'Dev, staging, production, and PR previews.',
        icon: 'i-lucide-git-branch',
        to: '/features/environments',
      },
      {
        label: 'Databases',
        description: 'Managed PostgreSQL via CloudNativePG.',
        icon: 'i-lucide-database',
        to: '/infrastructure/databases',
      },
      {
        label: 'Eject anytime',
        description: 'Standard Helm charts and ArgoCD configs.',
        icon: 'i-lucide-door-open',
        to: '/features/eject',
      },
    ],
  },
  {
    label: 'Pricing',
    to: '/pricing',
  },
  {
    label: 'Docs',
    to: '/getting-started/quick-start',
  },
];

const githubLink = computed(() =>
  appConfig.github?.url
    ? { to: appConfig.github.url, target: '_blank' }
    : null,
);
</script>

<template>
  <UHeader
    :ui="{
      center: 'flex-[3] justify-center',
    }"
    to="/"
    :title="appConfig.header?.title || site.name"
  >
    <template #title>
      <div class="flex items-center gap-2">
        <img src="/logo-light.svg" alt="Lucity" width="24" height="24" class="dark:hidden h-6 w-6 shrink-0">
        <img src="/logo-dark.svg" alt="Lucity" width="24" height="24" class="hidden dark:block h-6 w-6 shrink-0">
        <span class="font-serif text-2xl tracking-tight">Lucity</span>
      </div>
    </template>

    <UNavigationMenu
      :items="navItems"
      variant="link"
      content-orientation="horizontal"
      class="hidden lg:flex w-full justify-center"
    />

    <template #right>
      <UContentSearchButton
        :collapsed="false"
        class="hidden lg:inline-flex w-full max-w-40"
        variant="soft"
        :ui="{ leadingIcon: 'size-4 mr-1' }"
      />

      <UContentSearchButton class="lg:hidden" />

      <ClientOnly>
        <UColorModeButton />

        <template #fallback>
          <div class="h-8 w-8 animate-pulse bg-neutral-200 dark:bg-neutral-800 rounded-md" />
        </template>
      </ClientOnly>

      <UButton
        v-if="githubLink"
        v-bind="githubLink"
        icon="i-simple-icons-github"
        color="neutral"
        variant="ghost"
        aria-label="GitHub"
      />

      <UButton
        :to="`${appUrl}/login`"
        color="primary"
        class="whitespace-nowrap"
      >
        Get Started
      </UButton>
    </template>

    <template #toggle="{ open, toggle }">
      <IconMenuToggle
        :open="open"
        class="lg:hidden"
        aria-label="Toggle navigation menu"
        @click="toggle"
      />
    </template>

    <template #body>
      <UNavigationMenu
        :items="navItems"
        orientation="vertical"
        class="p-2"
      />

      <USeparator class="my-2" />

      <div class="px-4 pb-4">
        <UButton
          :to="`${appUrl}/login`"
          color="primary"
          block
        >
          Get Started
        </UButton>
      </div>

      <AppHeaderBody />
    </template>
  </UHeader>
</template>
