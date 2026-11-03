<script setup lang="ts">
import { RouterView } from 'vue-router';
import { useMediaQuery } from '@vueuse/core';
import { Sonner } from '@/components/ui/sonner';
import { TooltipProvider } from '@/components/ui/tooltip';
import { Undo2 } from '@lucide/vue';
import { useTheme } from '@/composables/useTheme';
import { useDiscardPrompt } from '@/composables/useUnsavedChanges';
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import GateDialog from '@/components/GateDialog.vue';
import AnnouncementBanner from '@/components/AnnouncementBanner.vue';
import VersionFooter from '@/components/VersionFooter.vue';

const { theme } = useTheme();
const discardPrompt = useDiscardPrompt();
const toastsClearBar = useMediaQuery('(width < 48rem)');
</script>

<template>
  <TooltipProvider :delay-duration="150">
    <AnnouncementBanner />
    <RouterView />
    <VersionFooter />
    <Sonner
      position="bottom-right"
      :theme="theme"
      :offset="toastsClearBar ? { bottom: 80 } : undefined"
      :mobile-offset="{ bottom: 80 }"
    />
    <ConfirmDialog
      :open="discardPrompt.open.value"
      title="Discard unsaved changes?"
      description="You have changes that aren't saved yet. Leaving drops them."
      confirm-label="Discard"
      cancel-label="Keep editing"
      :icon="Undo2"
      @update:open="(value) => { if (!value) discardPrompt.answer(false) }"
      @confirm="discardPrompt.answer(true)"
    />
    <GateDialog />
  </TooltipProvider>
</template>
