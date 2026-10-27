<script setup lang="ts">
import PanelShell from '@/components/PanelShell.vue';

defineProps<{
  title: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();
</script>

<template>
  <PanelShell
    as="aside"
    :title="title"
    :guard="false"
    @close="emit('close')"
  >
    <div class="min-h-0 flex-1 overflow-y-auto p-4">
      <slot />
    </div>

    <template v-if="$slots.footer" #footer>
      <slot name="footer" />
    </template>
  </PanelShell>
</template>
