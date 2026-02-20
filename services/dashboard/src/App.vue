<script setup lang="ts">
import { RouterView } from 'vue-router';
import { Sonner } from '@/components/ui/sonner';
import { useTheme } from '@/composables/useTheme';
import BackgroundGrid from '@/components/BackgroundGrid.vue';

const { theme } = useTheme();
</script>

<template>
  <BackgroundGrid />
  <RouterView />
  <Sonner position="bottom-right" rich-colors close-button :theme="theme" />
</template>
