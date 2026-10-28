<script setup lang="ts">
import { computed, ref } from 'vue';
import LogViewer from '@/components/LogViewer.vue';
import { useEscapeToClose } from '@/composables/useEscapeToClose';
import { useServiceLogs } from '@/composables/useServiceLogs';

const props = defineProps<{
  serviceId: string;
  serviceName: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

useEscapeToClose(() => emit('close'));

const serviceIdRef = computed(() => props.serviceId);
const { lines, error, retry } = useServiceLogs(serviceIdRef, ref(true));

const entries = computed(() => lines.value.map(entry => ({ text: entry.line.replace(/^\[[\w.-]+\] /, ''), pod: entry.pod })));
const state = computed(() => {
  if (error.value) return 'error' as const;
  return lines.value.length ? 'streaming' as const : 'connecting' as const;
});
</script>

<template>
  <LogViewer
    title="Runtime logs"
    :lines="entries"
    :state="state"
    :error="error"
    waiting-hint="Lines appear as the service writes them."
    :file-name="`${serviceName}-logs.txt`"
    @close="emit('close')"
    @retry="retry"
  />
</template>
