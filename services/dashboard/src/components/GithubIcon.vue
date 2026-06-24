<script setup lang="ts">
import { computed } from 'vue';
import { useTheme } from '@/composables/useTheme';

const props = withDefaults(defineProps<{ size?: number; primary?: boolean }>(), {
  size: 24,
  primary: false,
});

const { theme } = useTheme();

const src = computed(() => {
  const wantWhite = props.primary ? theme.value === 'light' : theme.value === 'dark';
  return wantWhite
    ? 'https://devicons.railway.com/i/github-light.svg'
    : 'https://devicons.railway.com/i/github-dark.svg';
});
</script>

<template>
  <img :src="src" :width="size" :height="size" alt="" />
</template>
