<script setup lang="ts">
import { computed } from 'vue';
import { Container } from 'lucide-vue-next';
import { useTheme } from '@/composables/useTheme';

const props = defineProps<{
  framework?: string | null;
  size?: number;
}>();

const { theme } = useTheme();
const iconSize = computed(() => props.size ?? 16);

// Frameworks that need -light/-dark suffix (monochrome icons)
const THEMED = new Set(['nextjs', 'rust', 'django', 'remix', 'flask', 'deno']);

const DEVICON_MAP: Record<string, string> = {
  nuxt: 'nuxtjs',
  nextjs: 'nextjs',
  vue: 'vuejs',
  vite: 'vitejs',
  react: 'react',
  svelte: 'svelte',
  astro: 'astro',
  angular: 'angularjs',
  node: 'nodejs',
  python: 'python',
  go: 'go',
  rust: 'rust',
  django: 'django',
  rails: 'rails',
  php: 'php',
  laravel: 'laravel',
  elixir: 'elixir',
  remix: 'remix',
  cra: 'react',
  'react-router': 'react',
  fastapi: 'fastapi',
  flask: 'flask',
  java: 'java',
  phoenix: 'phoenix',
  dotnet: 'dotnet',
  deno: 'deno',
  ruby: 'ruby',
};

const iconUrl = computed(() => {
  if (!props.framework) return null;
  const base = DEVICON_MAP[props.framework];
  if (!base) return null;
  const suffix = THEMED.has(base)
    ? theme.value === 'dark' ? '-light' : '-dark'
    : '';
  return `https://devicons.railway.com/i/${base}${suffix}.svg`;
});
</script>

<template>
  <img
    v-if="iconUrl"
    :src="iconUrl"
    :width="iconSize"
    :height="iconSize"
    class="shrink-0"
    alt=""
  />
  <Container v-else :size="iconSize" class="shrink-0 text-muted-foreground" />
</template>
