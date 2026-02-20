<script setup lang="ts">
import { computed, watch, ref, onMounted } from 'vue';
import { VueFlow, useVueFlow } from '@vue-flow/core';
import { Controls } from '@vue-flow/controls';
import { Background } from '@vue-flow/background';
import { Plus } from 'lucide-vue-next';
import { useEnvironment } from '@/composables/useEnvironment';
import { usePanel } from '@/composables/usePanel';
import ServiceNode from './ServiceNode.vue';
import { Button } from '@/components/ui/button';
import '@vue-flow/core/dist/style.css';
import '@vue-flow/core/dist/theme-default.css';
import '@vue-flow/controls/dist/style.css';

const props = defineProps<{
  services: {
    name: string;
    image: string;
    port: number;
    public: boolean;
    framework?: string;
  }[];
}>();

const emit = defineEmits<{
  (e: 'create'): void;
}>();

const { activeEnvServices } = useEnvironment();
const { openPanel, currentPanel } = usePanel();

const { fitView } = useVueFlow({
  id: 'service-canvas',
});

const nodes = computed(() => {
  return props.services.map((svc, index) => {
    const envService = activeEnvServices.value.find(es => es.name === svc.name);
    return {
      id: svc.name,
      type: 'service',
      position: { x: 0, y: index * 130 },
      data: {
        name: svc.name,
        framework: svc.framework,
        port: svc.port,
        public: svc.public,
        ready: envService?.ready,
        imageTag: envService?.imageTag,
      },
      selected: currentPanel.value?.id === svc.name,
    };
  });
});

const edges = ref([]);

function handleNodeClick(event: { node: { id: string; data: { name: string } } }) {
  openPanel({ type: 'service', id: event.node.id, label: event.node.data.name });
}

// Fit view on mount with max zoom capped at 1x
onMounted(() => {
  setTimeout(() => fitView({ padding: 0.3, maxZoom: 1 }), 200);
});

// Re-fit view when services change
watch(() => props.services.length, () => {
  setTimeout(() => fitView({ padding: 0.3, maxZoom: 1 }), 100);
});
</script>

<template>
  <div class="relative h-full w-full">
    <VueFlow
      :nodes="nodes"
      :edges="edges"
      :default-viewport="{ zoom: 1, x: 0, y: 0 }"
      :min-zoom="0.25"
      :max-zoom="2"
      :snap-to-grid="true"
      :snap-grid="[20, 20]"
      class="canvas-bg"
      @node-click="handleNodeClick"
    >
      <template #node-service="nodeProps">
        <ServiceNode
          :data="nodeProps.data"
          :selected="nodeProps.selected"
          @select="openPanel({ type: 'service', id: nodeProps.id, label: nodeProps.data.name })"
        />
      </template>

      <Background variant="dots" :gap="24" :size="1" class="canvas-dots" />
      <Controls position="top-left" class="!border-border !bg-card !shadow-sm" />
    </VueFlow>

    <!-- Create button (floating top-right) -->
    <div class="absolute right-4 top-4 z-10">
      <Button
        variant="outline"
        size="sm"
        @click="emit('create')"
      >
        <Plus :size="14" class="mr-1" />
        Create
      </Button>
    </div>
  </div>
</template>

<style scoped>
.canvas-bg {
  background-color: transparent;
}

:deep(.canvas-dots pattern circle) {
  fill: color-mix(in oklch, var(--muted-foreground) 25%, transparent);
}

:deep(.vue-flow__controls) {
  display: flex;
  flex-direction: column;
  gap: 2px;
  border-radius: 8px;
  overflow: hidden;
}

:deep(.vue-flow__controls-button) {
  background-color: var(--card);
  border-color: var(--border);
  color: var(--foreground);
  width: 28px;
  height: 28px;
}

:deep(.vue-flow__controls-button:hover) {
  background-color: var(--accent);
}
</style>
