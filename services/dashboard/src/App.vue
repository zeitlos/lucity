<script setup lang="ts">
import { RouterView } from 'vue-router';
import { Sonner } from '@/components/ui/sonner';
import { TooltipProvider } from '@/components/ui/tooltip';
import { Undo2 } from '@lucide/vue';
import { useTheme } from '@/composables/useTheme';
import { useDiscardPrompt } from '@/composables/useUnsavedChanges';
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import AnnouncementBanner from '@/components/AnnouncementBanner.vue';
import VersionFooter from '@/components/VersionFooter.vue';

const { theme } = useTheme();
const discardPrompt = useDiscardPrompt();
</script>

<template>
  <TooltipProvider :delay-duration="150">
    <AnnouncementBanner />
    <RouterView />
    <VersionFooter />
    <Sonner position="bottom-right" rich-colors close-button :theme="theme" />
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
  </TooltipProvider>
</template>
