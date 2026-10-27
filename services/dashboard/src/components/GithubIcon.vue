<script setup lang="ts">
import { computed } from 'vue';
import { useTheme } from '@/composables/useTheme';
import { devicon } from '@/lib/devicons';

const props = withDefaults(defineProps<{ size?: number; primary?: boolean }>(), {
  size: 24,
  primary: false,
});

const { theme } = useTheme();

const src = computed(() => {
  const wantWhite = props.primary ? theme.value === 'light' : theme.value === 'dark';
  return wantWhite
    ? devicon('github-light')
    : devicon('github-dark');
});
</script>

<template>
  <img :src="src" :width="size" :height="size" alt="" />
</template>
