<script setup lang="ts">
import { computed, watch, ref, onMounted, toRef } from 'vue';
import { VueFlow, useVueFlow, Panel } from '@vue-flow/core';
import { Background } from '@vue-flow/background';
import { Plus, Maximize2 } from 'lucide-vue-next';
import { useEnvironment } from '@/composables/useEnvironment';
import { usePanel } from '@/composables/usePanel';
import { useCanvasDeployStatus } from '@/composables/useCanvasDeployStatus';
import { useDatabaseAutoConnect } from '@/composables/useDatabaseAutoConnect';
import ServiceNode from './ServiceNode.vue';
import DatabaseNode from './DatabaseNode.vue';
import { Button } from '@/components/ui/button';
import '@vue-flow/core/dist/style.css';
import '@vue-flow/core/dist/theme-default.css';

const props = defineProps<{
  projectId: string;
  services: {
    name: string;
    image: string;
    port: number;
    framework?: string;
    sourceUrl?: string;
  }[];
  databases?: {
    name: string;
    version: string;
    instances: number;
    size: string;
  }[];
}>();

const emit = defineEmits<{
  (e: 'create'): void;
  (e: 'deploy-completed'): void;
}>();

const { activeEnvServices, activeEnvDatabases, activeEnvironment } = useEnvironment();
const { openPanel, currentPanel } = usePanel();

const serviceNames = computed(() => props.services.map(s => s.name));
const envName = computed(() => activeEnvironment.value?.name ?? null);
const { statusMap } = useCanvasDeployStatus(
  toRef(props, 'projectId'), envName, serviceNames,
  () => emit('deploy-completed'),
);

// Auto-create DATABASE_URL shared variables when databases become ready.
useDatabaseAutoConnect(toRef(props, 'projectId'), envName, activeEnvDatabases);

const { fitView, findNode, setCenter, dimensions } = useVueFlow({
  id: 'service-canvas',
});

const nodes = computed(() => {
  const serviceNodes = props.services.map((svc, index) => {
    const envService = activeEnvServices.value.find(es => es.name === svc.name);
    const deployInfo = statusMap.value[svc.name];
    return {
      id: svc.name,
      type: 'service',
      position: { x: 0, y: index * 180 },
      data: {
        name: svc.name,
        framework: svc.framework,
        port: svc.port,
        sourceUrl: svc.sourceUrl,
        host: envService?.host,
        ready: envService?.ready,
        imageTag: envService?.imageTag,
        replicas: envService?.replicas,
        activeDeployPhase: deployInfo?.phase ?? null,
        activeDeployStartedAt: deployInfo?.startedAt ?? null,
      },
      selected: currentPanel.value?.id === svc.name && currentPanel.value?.type === 'service',
    };
  });

  const databaseNodes = (props.databases ?? []).map((db, index) => {
    const envDb = activeEnvDatabases.value.find(ed => ed.name === db.name);
    return {
      id: `db-${db.name}`,
      type: 'database',
      position: { x: 340, y: index * 220 },
      data: {
        name: db.name,
        version: db.version,
        instances: envDb?.instances ?? db.instances,
        size: envDb?.size ?? db.size,
        ready: envDb?.ready,
        volume: envDb?.volume ?? null,
      },
      selected: currentPanel.value?.id === db.name && currentPanel.value?.type === 'database',
    };
  });

  return [...serviceNodes, ...databaseNodes];
});

const edges = ref([]);

function handleNodeClick(event: { node: { id: string; type: string; data: { name: string } } }) {
  if (event.node.type === 'database') {
    openPanel({ type: 'database', id: event.node.data.name, label: event.node.data.name });
  } else {
    openPanel({ type: 'service', id: event.node.id, label: event.node.data.name });
  }
}

function handleFitView() {
  fitView({ padding: 0.3, maxZoom: 1 });
}

// Fit view on mount
onMounted(() => {
  setTimeout(() => handleFitView(), 200);
});

// Re-fit view when services or databases change
const totalNodes = computed(() => props.services.length + (props.databases?.length ?? 0));
watch(totalNodes, () => {
  setTimeout(() => handleFitView(), 100);
});

// Center selected card in the visible left portion (panel overlays 55% from right)
watch(
  () => currentPanel.value,
  (panel, oldPanel) => {
    if (panel?.type === 'service' || panel?.type === 'database') {
      const nodeId = panel.type === 'database' ? `db-${panel.id}` : panel.id;
      const node = findNode(nodeId);
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

      <template #node-database="nodeProps">
        <DatabaseNode
          :data="nodeProps.data"
          :selected="nodeProps.selected"
          @select="openPanel({ type: 'database', id: nodeProps.data.name, label: nodeProps.data.name })"
          @select-volume="openPanel({ type: 'volume', id: $event, label: 'Volume' })"
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
