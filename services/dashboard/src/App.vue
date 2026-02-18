<script setup lang="ts">
import { RouterView } from 'vue-router';
import { Sonner } from '@/components/ui/sonner';
import { useTheme } from '@/composables/useTheme';

const { theme } = useTheme();
</script>

<template>
  <RouterView />
  <Sonner position="bottom-right" rich-colors :theme="theme" />
</template>
