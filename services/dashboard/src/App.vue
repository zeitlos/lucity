<script setup lang="ts">
import { RouterView } from 'vue-router';
import { Sonner } from '@/components/ui/sonner';
import { TooltipProvider } from '@/components/ui/tooltip';
import { useTheme } from '@/composables/useTheme';
import AnnouncementBanner from '@/components/AnnouncementBanner.vue';
import VersionFooter from '@/components/VersionFooter.vue';

const { theme } = useTheme();
</script>

<template>
  <TooltipProvider :delay-duration="150">
    <AnnouncementBanner />
    <RouterView />
    <VersionFooter />
    <Sonner position="bottom-right" rich-colors close-button :theme="theme" />
  </TooltipProvider>
</template>
