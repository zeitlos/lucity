<script setup lang="ts">
const appConfig = useAppConfig();
const site = useSiteConfig();

const appUrl = '/app';

const route = useRoute();

const inDocs = computed(() => route.meta.layout === 'docs');

const navItems = computed(() => [
  {
    label: 'Docs',
    to: '/quickstart',
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
        <div class="flex items-center gap-5 text-highlighted">
          <img
            src="/logo-mark.svg"
            alt=""
            width="59"
            height="24"
            class="h-6 w-auto shrink-0"
          >
          <span class="wordmark text-2xl leading-none">Lucity Docs</span>
        </div>
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

<style scoped>
.wordmark {
  font-family: 'Redaction 35', Georgia, serif;
  font-weight: 400;
}
</style>
