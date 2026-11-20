<script setup lang="ts">
defineProps<{
  guides: { path: string; title: string }[];
}>();

const icons: Record<string, string> = {
  nextjs: 'i-simple-icons-nextdotjs',
  nuxt: 'i-simple-icons-nuxt',
  sveltekit: 'i-simple-icons-svelte',
  astro: 'i-simple-icons-astro',
  express: 'i-simple-icons-express',
  fastapi: 'i-simple-icons-fastapi',
  django: 'i-simple-icons-django',
  laravel: 'i-simple-icons-laravel',
  go: 'i-simple-icons-go',
};

function iconOf(path: string) {
  return icons[path.split('/').pop() ?? ''] ?? 'i-lucide-book-open';
}
</script>

<template>
  <div class="guides border-b border-stone-300 dark:border-stone-800">
    <NuxtLink
      v-for="guide in guides"
      :key="guide.path"
      :to="guide.path"
      class="flex min-h-11 items-center gap-3.5 border-t border-stone-300 py-4 text-lg font-semibold text-stone-900 underline-offset-4 hover:underline dark:border-stone-800 dark:text-stone-100"
    >
      <UIcon :name="iconOf(guide.path)" class="size-6 shrink-0" />
      <span class="min-w-0">{{ guide.title }}</span>
    </NuxtLink>
    <NuxtLink
      to="/docs/guides"
      class="flex min-h-11 items-center gap-3.5 border-t border-stone-300 py-4 text-lg font-semibold text-stone-700 underline-offset-4 hover:underline dark:border-stone-800 dark:text-stone-400"
    >
      <span>All guides</span>
      <UIcon name="i-lucide-arrow-right" class="size-5 shrink-0" />
    </NuxtLink>
  </div>
</template>

<style scoped>
.guides {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(210px, 100%), 1fr));
  gap: 0 40px;
}

@media (max-width: 640px) {
  .guides {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0 24px;
  }
}
</style>
