<script setup lang="ts">
const appConfig = useAppConfig();
const site = useSiteConfig();

const appUrl = '/app';

const route = useRoute();

const inDocs = computed(() => route.path === '/docs' || route.path.startsWith('/docs/'));

provide('navigation', await useContentSectionNavigation());

const navItems = computed(() => [
  {
    label: 'Docs',
    to: '/docs/quickstart',
    active: inDocs.value,
  },
  {
    label: 'Blog',
    to: '/blog',
  },
  {
    label: 'Pricing',
    to: '/pricing',
  },
]);

const { panelOpen } = useLogoGenerator();

const logoMenu = [[{
  label: 'Open logo generator',
  icon: 'i-lucide-pen-tool',
  onSelect: () => { panelOpen.value = true; },
}]];

const githubLink = computed(() =>
  appConfig.github?.url
    ? { to: appConfig.github.url, target: '_blank' }
    : null,
);
</script>

<template>
  <AnnouncementBanner />

  <UHeader
    :ui="{
      root: 'docs-nav border-b-0 bg-transparent backdrop-blur-none',
      center: 'flex-[3] justify-center',
    }"
    to="/"
    :title="appConfig.header?.title || site.name"
  >
    <template #title>
      <UContextMenu :items="logoMenu">
        <LucityLogo
          :suffix="inDocs ? 'Docs' : undefined"
          class="text-lg sm:text-xl text-highlighted"
        />
      </UContextMenu>
    </template>

    <template #right>
      <UNavigationMenu
        :items="navItems"
        variant="link"
        content-orientation="horizontal"
        class="hidden lg:flex"
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
        external
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
          external
          color="primary"
          block
        >
          Get Started
        </UButton>
      </div>

      <AppHeaderBody />
    </template>
  </UHeader>

  <LogoGeneratorPanel />
</template>
