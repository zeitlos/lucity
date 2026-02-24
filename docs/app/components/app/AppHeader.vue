<script setup lang="ts">
const appConfig = useAppConfig();
const site = useSiteConfig();

const links = computed(() => appConfig.github && appConfig.github.url
  ? [
      {
        'icon': 'i-simple-icons-github',
        'to': appConfig.github.url,
        'target': '_blank',
        'aria-label': 'GitHub',
      },
    ]
  : []);

const navLinks = [
  { label: 'Docs', to: '/getting-started/quick-start' },
  { label: 'Pricing', to: '/pricing' },
  { label: 'Blog', to: '/blog' },
];
</script>

<template>
  <UHeader
    :ui="{ center: 'flex-1' }"
    to="/"
    :title="appConfig.header?.title || site.name"
  >
    <template #title>
      <div class="flex items-center gap-2">
        <AppHeaderLogo class="h-6 w-auto shrink-0" />
        <span class="font-serif text-xl tracking-tight">Lucity</span>
      </div>
    </template>

    <div class="hidden lg:flex items-center gap-1">
      <UButton
        v-for="link in navLinks"
        :key="link.label"
        :to="link.to"
        :target="link.target"
        color="neutral"
        variant="ghost"
      >
        {{ link.label }}
      </UButton>
    </div>

    <template #right>
      <UContentSearchButton
        :collapsed="false"
        class="hidden lg:inline-flex w-full max-w-56"
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

      <template v-if="links?.length">
        <UButton
          v-for="(link, index) of links"
          :key="index"
          v-bind="{ color: 'neutral', variant: 'ghost', ...link }"
        />
      </template>
    </template>

    <template #toggle="{ open, toggle }">
      <IconMenuToggle
        :open="open"
        class="lg:hidden"
        @click="toggle"
      />
    </template>

    <template #body>
      <div class="flex flex-col gap-1 p-4">
        <UButton
          v-for="link in navLinks"
          :key="link.label"
          :to="link.to"
          :target="link.target"
          color="neutral"
          variant="ghost"
          block
          class="justify-start"
        >
          {{ link.label }}
        </UButton>
      </div>

      <USeparator />

      <AppHeaderBody />
    </template>
  </UHeader>
</template>
