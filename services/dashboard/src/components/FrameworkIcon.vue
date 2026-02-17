<script setup lang="ts">
import { computed } from 'vue';
import { Container } from 'lucide-vue-next';

const props = defineProps<{
  framework?: string | null;
  size?: number;
}>();

const iconSize = computed(() => props.size ?? 16);

const frameworkConfig = computed(() => {
  const configs: Record<string, { label: string; color: string }> = {
    nextjs: { label: 'N', color: 'bg-black text-white dark:bg-white dark:text-black' },
    vite: { label: 'V', color: 'bg-purple-500 text-white' },
    nuxt: { label: 'Nx', color: 'bg-green-600 text-white' },
    astro: { label: 'A', color: 'bg-orange-500 text-white' },
    remix: { label: 'R', color: 'bg-blue-600 text-white' },
    svelte: { label: 'S', color: 'bg-red-500 text-white' },
    angular: { label: 'Ng', color: 'bg-red-600 text-white' },
    cra: { label: 'CR', color: 'bg-cyan-500 text-white' },
    'react-router': { label: 'RR', color: 'bg-red-400 text-white' },
    node: { label: 'JS', color: 'bg-green-500 text-white' },
    django: { label: 'Dj', color: 'bg-green-800 text-white' },
    fastapi: { label: 'FA', color: 'bg-teal-600 text-white' },
    flask: { label: 'Fl', color: 'bg-gray-600 text-white' },
    python: { label: 'Py', color: 'bg-yellow-500 text-black' },
    go: { label: 'Go', color: 'bg-cyan-600 text-white' },
    rust: { label: 'Rs', color: 'bg-orange-700 text-white' },
    rails: { label: 'Rb', color: 'bg-red-700 text-white' },
    ruby: { label: 'Rb', color: 'bg-red-600 text-white' },
    laravel: { label: 'La', color: 'bg-red-500 text-white' },
    php: { label: 'PH', color: 'bg-indigo-500 text-white' },
    java: { label: 'Jv', color: 'bg-orange-600 text-white' },
    phoenix: { label: 'Ph', color: 'bg-orange-500 text-white' },
    elixir: { label: 'Ex', color: 'bg-purple-600 text-white' },
    dotnet: { label: '.N', color: 'bg-purple-700 text-white' },
    deno: { label: 'De', color: 'bg-gray-800 text-white' },
  };

  return props.framework ? configs[props.framework] : null;
});
</script>

<template>
  <span
    v-if="frameworkConfig"
    class="inline-flex items-center justify-center rounded text-[10px] font-bold leading-none"
    :class="frameworkConfig.color"
    :style="{ width: `${iconSize}px`, height: `${iconSize}px` }"
  >
    {{ frameworkConfig.label }}
  </span>
  <Container v-else :size="iconSize" class="text-muted-foreground" />
</template>
