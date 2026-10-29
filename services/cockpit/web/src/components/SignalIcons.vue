<script setup lang="ts">
import { signalIcons, signalLabel, type Signal } from '@/lib/workspaces';

defineProps<{
  signals: Signal[];
  muted?: boolean;
}>();
</script>

<template>
  <span class="flex items-center gap-1.5">
    <span
      v-for="signal in signals"
      :key="signal.kind + signal.detail"
      role="img"
      :aria-label="signalLabel(signal)"
      :title="signalLabel(signal)"
      :class="muted ? 'text-muted-foreground' : signal.strong ? 'text-status-danger' : 'text-status-warn'"
    >
      <component :is="signalIcons[signal.kind]" class="size-4" :stroke-width="2.25" aria-hidden="true" />
    </span>
  </span>
</template>
