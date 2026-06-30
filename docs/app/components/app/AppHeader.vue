<script setup lang="ts">
const appConfig = useAppConfig();
const site = useSiteConfig();

const appUrl = 'https://lucity.cloud/app';

const navItems = [
  {
    label: 'Pricing',
    to: '/pricing',
  },
  {
    label: 'Docs',
    to: '/getting-started/concepts',
  },
  {
    label: 'Blog',
    to: '/blog',
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
      root: 'docs-nav border-b-0 bg-transparent backdrop-blur-none',
      center: 'flex-[3] justify-center',
    }"
    to="/"
    :title="appConfig.header?.title || site.name"
  >
    <template #title>
      <div class="flex items-center gap-2">
        <img src="/logo-light.svg" alt="Lucity" width="36" height="36" class="dark:hidden h-9 w-9 shrink-0">
        <img src="/logo-dark.svg" alt="Lucity" width="36" height="36" class="hidden dark:block h-9 w-9 shrink-0">
        <span class="wordmark text-3xl leading-none">Lucity</span>
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

<style scoped>
.wordmark {
  font-family: 'Redaction 35', Georgia, serif;
  font-weight: 400;
}
</style>
