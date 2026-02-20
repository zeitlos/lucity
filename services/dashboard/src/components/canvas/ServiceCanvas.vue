<script setup lang="ts">
import { computed, watch, ref, onMounted } from 'vue';
import { VueFlow, useVueFlow, Panel } from '@vue-flow/core';
import { Background } from '@vue-flow/background';
import { Plus, Maximize2 } from 'lucide-vue-next';
import { useEnvironment } from '@/composables/useEnvironment';
import { usePanel } from '@/composables/usePanel';
import ServiceNode from './ServiceNode.vue';
import { Button } from '@/components/ui/button';
import '@vue-flow/core/dist/style.css';
import '@vue-flow/core/dist/theme-default.css';

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
      position: { x: 0, y: index * 180 },
      data: {
        name: svc.name,
        framework: svc.framework,
        port: svc.port,
        public: svc.public,
        ready: envService?.ready,
        imageTag: envService?.imageTag,
        replicas: envService?.replicas,
      },
      selected: currentPanel.value?.id === svc.name,
    };
  });
});

const edges = ref([]);

function handleNodeClick(event: { node: { id: string; data: { name: string } } }) {
  openPanel({ type: 'service', id: event.node.id, label: event.node.data.name });
}

function handleFitView() {
  fitView({ padding: 0.3, maxZoom: 1 });
}

// Fit view on mount
onMounted(() => {
  setTimeout(() => handleFitView(), 200);
});

// Re-fit view when services change
watch(() => props.services.length, () => {
  setTimeout(() => handleFitView(), 100);
});

// Center selected card when panel opens, re-fit all when it closes
watch(
  () => currentPanel.value,
  (panel, oldPanel) => {
    if (panel?.type === 'service') {
      fitView({ nodes: [panel.id], padding: 0.5, maxZoom: 1 });
    } else if (!panel && oldPanel) {
      handleFitView();
    }
  },
);
</script>

<template>
  <div class="relative h-full w-full">
    <VueFlow
      :nodes="nodes"
      :edges="edges"
      :default-viewport="{ zoom: 1, x: 0, y: 0 }"
      :min-zoom="1"
      :max-zoom="1"
      :zoom-on-scroll="false"
      :zoom-on-double-click="false"
      :zoom-on-pinch="false"
      :pan-on-scroll="true"
      :pan-on-scroll-mode="'vertical'"
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

      <Panel position="top-left" class="!m-3">
        <button
          class="flex h-8 w-8 items-center justify-center rounded-lg border border-border bg-card text-muted-foreground shadow-sm transition-colors hover:bg-accent hover:text-foreground"
          title="Fit view"
          @click="handleFitView"
        >
          <Maximize2 :size="14" />
        </button>
      </Panel>
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
</style>
