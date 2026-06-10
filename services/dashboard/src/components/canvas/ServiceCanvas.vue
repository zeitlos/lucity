<script setup lang="ts">
import { computed, watch, ref, onMounted, toRef } from 'vue';
import { VueFlow, useVueFlow, Panel, PanOnScrollMode } from '@vue-flow/core';
import { Background } from '@vue-flow/background';
import { Plus, Maximize2 } from 'lucide-vue-next';
import { usePanel } from '@/composables/usePanel';
import { useCanvasBuildStatus } from '@/composables/useCanvasBuildStatus';
import type { Service, Database } from '@/composables/useEnvironment';
import ServiceNode from './ServiceNode.vue';
import DatabaseNode from './DatabaseNode.vue';
import { Button } from '@/components/ui/button';
import '@vue-flow/core/dist/style.css';
import '@vue-flow/core/dist/theme-default.css';

const props = defineProps<{
  services: Service[];
  databases: Database[];
}>();

const emit = defineEmits<{
  (e: 'create'): void;
  (e: 'deploy-completed'): void;
}>();

const { openPanel, currentPanel } = usePanel();

const servicesRef = toRef(props, 'services');
const { statusMap } = useCanvasBuildStatus(
  servicesRef,
  () => emit('deploy-completed'),
);

const { fitView, findNode, setCenter, dimensions } = useVueFlow({
  id: 'service-canvas',
});

const nodes = computed(() => {
  const serviceNodes = props.services.map((svc, index) => {
    const buildInfo = statusMap.value[svc.id];
    return {
      id: svc.id,
      type: 'service',
      position: { x: 0, y: index * 180 },
      data: {
        name: svc.name,
        sourceUrl: svc.sourceUrl,
        endpoints: svc.endpoints,
        status: svc.status,
        replicas: svc.replicas,
        activeBuildStatus: buildInfo?.status ?? null,
        activeBuildStartedAt: buildInfo?.startedAt ?? null,
      },
      selected: currentPanel.value?.id === svc.id && currentPanel.value?.type === 'service',
    };
  });

  const databaseNodes = props.databases.map((db, index) => {
    return {
      id: db.id,
      type: 'database',
      position: { x: 340, y: index * 220 },
      data: {
        name: db.name,
        version: db.version,
        instances: db.instances,
        size: db.size,
        status: db.status,
      },
      selected: currentPanel.value?.id === db.id && currentPanel.value?.type === 'database',
    };
  });

  return [...serviceNodes, ...databaseNodes];
});

const edges = ref([]);

function handleNodeClick(event: { node: { id: string; type: string; data: { name: string } } }) {
  if (event.node.type === 'database') {
    openPanel({ type: 'database', id: event.node.id, label: event.node.data.name });
  } else {
    openPanel({ type: 'service', id: event.node.id, label: event.node.data.name });
  }
}

function handleFitView() {
  fitView({ padding: 0.3, maxZoom: 1 });
}

onMounted(() => {
  setTimeout(() => {
    const panel = currentPanel.value;
    if (panel?.type === 'service' || panel?.type === 'database') {
      const node = findNode(panel.id);
      if (node) {
        const nodeCenterX = node.position.x + (node.dimensions.width / 2);
        const nodeCenterY = node.position.y + (node.dimensions.height / 2);
        const panelOffset = (dimensions.value.width * 0.55) / 2;
        setCenter(nodeCenterX + panelOffset, nodeCenterY, { zoom: 1 });
        return;
      }
    }
    handleFitView();
  }, 200);
});

const totalNodes = computed(() => props.services.length + props.databases.length);
watch(totalNodes, () => {
  setTimeout(() => handleFitView(), 100);
});

watch(
  () => currentPanel.value,
  (panel, oldPanel) => {
    if (panel?.type === 'service' || panel?.type === 'database') {
      const node = findNode(panel.id);
      if (node) {
        const nodeCenterX = node.position.x + (node.dimensions.width / 2);
        const nodeCenterY = node.position.y + (node.dimensions.height / 2);
        const panelOffset = (dimensions.value.width * 0.55) / 2;
        setCenter(nodeCenterX + panelOffset, nodeCenterY, { zoom: 1 });
      }
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
      :pan-on-scroll-mode="PanOnScrollMode.Vertical"
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

      <template #node-database="nodeProps">
        <DatabaseNode
          :data="nodeProps.data"
          :selected="nodeProps.selected"
          @select="openPanel({ type: 'database', id: nodeProps.id, label: nodeProps.data.name })"
        />
      </template>

      <Background variant="dots" :gap="24" :size="1" />

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
