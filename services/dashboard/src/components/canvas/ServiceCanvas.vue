<script setup lang="ts">
import IconButton from '@/components/IconButton.vue';
import { computed, watch, ref, nextTick, toRef } from 'vue';
import { VueFlow, useVueFlow, getRectOfNodes, type GraphNode, type ViewportTransform } from '@vue-flow/core';
import { Background } from '@vue-flow/background';
import { Plus, Scan, WandSparkles } from '@lucide/vue';
import { usePanel } from '@/composables/usePanel';
import { useCanvasReleaseStatus } from '@/composables/useCanvasReleaseStatus';
import { useCanvasLayout, firstFreeSpot, nearestFreeSpot, gridSize, type NodeRect } from '@/composables/useCanvasLayout';
import type { Service, Database, KeyValueStore, Bucket, Volume } from '@/composables/useEnvironment';
import type { Agent } from '@/lib/agents';
import ServiceNode from './ServiceNode.vue';
import DatabaseNode from './DatabaseNode.vue';
import KeyValueStoreNode from './KeyValueStoreNode.vue';
import BucketNode from './BucketNode.vue';
import VolumeNode from './VolumeNode.vue';
import AgentNode from './AgentNode.vue';
import { Button } from '@/components/ui/button';
import '@vue-flow/core/dist/style.css';
import '@vue-flow/core/dist/theme-default.css';

const props = defineProps<{
  services: Service[];
  databases: Database[];
  keyValueStores: KeyValueStore[];
  buckets: Bucket[];
  volumes: Volume[];
  agents: Agent[];
  environmentId: string;
}>();

const emit = defineEmits<{
  (e: 'create'): void;
  (e: 'deploy-completed'): void;
  (e: 'mount-volume', volumeId: string): void;
  (e: 'pane-click'): void;
}>();

const { openPanel, currentPanel } = usePanel();

const servicesRef = toRef(props, 'services');
const { statusMap, failureMap } = useCanvasReleaseStatus(
  servicesRef,
  () => emit('deploy-completed'),
);

const { getNodes, areNodesInitialized, setViewport, findNode, setCenter, dimensions, viewport, onNodeDragStop } = useVueFlow({
  id: 'service-canvas',
});

const { positionFor, setPosition, reconcile, reset } = useCanvasLayout(() => props.environmentId);

const allNodeIds = computed(() => [
  ...props.services,
  ...props.databases,
  ...props.keyValueStores,
  ...props.buckets,
  ...props.volumes,
  ...props.agents,
].map(node => node.id));

watch(
  () => allNodeIds.value.join('|'),
  () => reconcile(allNodeIds.value),
  { immediate: true },
);

const settlingNodeIds = ref<string[]>([]);

function placement(nodeId: string) {
  const position = positionFor(nodeId);
  const settling = settlingNodeIds.value.includes(nodeId) ? 'settling' : '';
  if (position) return { position, style: {}, class: settling };
  return { position: { x: 0, y: 0 }, style: { visibility: 'hidden' as const }, class: '' };
}

function settle(nodeId: string) {
  settlingNodeIds.value = [...settlingNodeIds.value, nodeId];
  setTimeout(() => {
    settlingNodeIds.value = settlingNodeIds.value.filter(id => id !== nodeId);
  }, 200);
}

const nodes = computed(() => {
  const serviceNodes = props.services.map(svc => {
    const releaseInfo = statusMap.value[svc.id];
    const mountedVolume = props.volumes.find(vol => vol.mount?.service === svc.id);
    return {
      id: svc.id,
      type: 'service',
      ...placement(svc.id),
      data: {
        name: svc.name,
        sourceUrl: svc.sourceUrl,
        endpoints: svc.endpoints,
        status: svc.status,
        replicas: svc.replicas,
        activeReleasePhase: releaseInfo?.phase ?? null,
        activeReleaseStartedAt: releaseInfo?.startedAt ?? null,
        releaseFailure: failureMap.value[svc.id] ?? null,
        live: !!svc.activeDeployment,
        volume: mountedVolume
          ? {
            id: mountedVolume.id,
            name: mountedVolume.name,
            path: mountedVolume.mount!.path,
            usagePercent: mountedVolume.usagePercent ?? null,
            selected: currentPanel.value?.id === mountedVolume.id && currentPanel.value?.type === 'volume',
          }
          : null,
      },
      selected: currentPanel.value?.id === svc.id && currentPanel.value?.type === 'service',
    };
  });

  const databaseNodes = props.databases.map(db => {
    return {
      id: db.id,
      type: 'database',
      ...placement(db.id),
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

  const keyValueStoreNodes = props.keyValueStores.map(kv => {
    return {
      id: kv.id,
      type: 'keyValueStore',
      ...placement(kv.id),
      data: {
        name: kv.name,
        version: kv.version,
        size: kv.size,
        status: kv.status,
      },
      selected: currentPanel.value?.id === kv.id && currentPanel.value?.type === 'keyValueStore',
    };
  });

  const bucketNodes = props.buckets.map(bucket => {
    return {
      id: bucket.id,
      type: 'bucket',
      ...placement(bucket.id),
      data: {
        name: bucket.name,
        region: bucket.region,
        sizeBytes: bucket.sizeBytes,
        status: bucket.status,
      },
      selected: currentPanel.value?.id === bucket.id && currentPanel.value?.type === 'bucket',
    };
  });

  const volumeNodes = props.volumes
    .filter(volume => !volume.mount)
    .map(volume => {
      return {
        id: volume.id,
        type: 'volume',
        ...placement(volume.id),
        data: {
          name: volume.name,
        },
        selected: currentPanel.value?.id === volume.id && currentPanel.value?.type === 'volume',
      };
    });

  const agentNodes = props.agents.map(agent => {
    return {
      id: agent.id,
      type: 'agent',
      ...placement(agent.id),
      data: {
        name: agent.name,
        harness: agent.harness,
        status: agent.status,
        connectors: agent.connectors.map(grant => grant.connector.name),
      },
      selected: currentPanel.value?.id === agent.id && currentPanel.value?.type === 'agent',
    };
  });

  return [...serviceNodes, ...databaseNodes, ...keyValueStoreNodes, ...bucketNodes, ...volumeNodes, ...agentNodes];
});

const edges = ref<never[]>([]);

function handleNodeClick(event: { node: { id: string; type: string; data: { name: string } } }) {
  if (event.node.type === 'database') {
    openPanel({ type: 'database', id: event.node.id, label: event.node.data.name });
  } else if (event.node.type === 'keyValueStore') {
    openPanel({ type: 'keyValueStore', id: event.node.id, label: event.node.data.name });
  } else if (event.node.type === 'bucket') {
    openPanel({ type: 'bucket', id: event.node.id, label: event.node.data.name });
  } else if (event.node.type === 'volume') {
    openPanel({ type: 'volume', id: event.node.id, label: event.node.data.name });
  } else if (event.node.type === 'agent') {
    openPanel({ type: 'agent', id: event.node.id, label: event.node.data.name });
  } else {
    openPanel({ type: 'service', id: event.node.id, label: event.node.data.name });
  }
}

const canvasInset = { x: 16, y: 68 };

function axisOffset(start: number, size: number, viewport: number, inset: number) {
  if (size + 2 * inset <= viewport) return (viewport - size) / 2 - start;
  return inset - start;
}

function handleFitView(duration = 0) {
  const bounds = getRectOfNodes(getNodes.value);
  const { width, height } = dimensions.value;
  setViewport({
    x: axisOffset(bounds.x, bounds.width, width, canvasInset.x),
    y: axisOffset(bounds.y, bounds.height, height, canvasInset.y),
    zoom: 1,
  }, { duration, interpolate: 'linear' });
}

function handleResetLayout() {
  reset();
  setTimeout(() => handleFitView(), 50);
}

function rectOf(node: GraphNode): NodeRect {
  return { ...node.position, ...node.dimensions };
}

const unplacedNodeIds = computed(() => getNodes.value.filter(node => !positionFor(node.id)).map(node => node.id));

function placeNewNodes() {
  if (!areNodesInitialized.value) return;
  const occupied = getNodes.value.filter(node => positionFor(node.id)).map(rectOf);
  for (const node of getNodes.value) {
    if (positionFor(node.id)) continue;
    const position = firstFreeSpot(node.dimensions, occupied);
    setPosition(node.id, position);
    occupied.push({ ...position, ...node.dimensions });
  }
}

watch([() => unplacedNodeIds.value.join('|'), areNodesInitialized], placeNewNodes);

onNodeDragStop(({ nodes: draggedNodes }) => {
  for (const node of draggedNodes) {
    const occupied = getNodes.value.filter(other => other.id !== node.id).map(rectOf);
    const spot = nearestFreeSpot(rectOf(node), occupied);
    if (spot.x !== node.position.x || spot.y !== node.position.y) settle(node.id);
    setPosition(node.id, spot);
  }
});

const stopInitialView = watch(
  () => areNodesInitialized.value && unplacedNodeIds.value.length === 0,
  async (settled) => {
    if (!settled) return;
    stopInitialView();
    await nextTick();
    const panel = currentPanel.value;
    if (panel?.type === 'service' || panel?.type === 'database' || panel?.type === 'keyValueStore' || panel?.type === 'bucket' || panel?.type === 'volume' || panel?.type === 'agent') {
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
  },
);

const totalNodes = computed(() => props.services.length + props.databases.length + props.keyValueStores.length + props.buckets.length + props.volumes.length + props.agents.length);
watch(totalNodes, () => {
  setTimeout(() => handleFitView(), 100);
});

let viewportBeforePanel: ViewportTransform | null = null;

watch(
  () => currentPanel.value,
  (panel, oldPanel) => {
    if (panel?.type === 'service' || panel?.type === 'database' || panel?.type === 'keyValueStore' || panel?.type === 'bucket' || panel?.type === 'volume' || panel?.type === 'agent') {
      const node = findNode(panel.id);
      if (node) {
        if (!oldPanel) viewportBeforePanel = { ...viewport.value };
        const nodeCenterX = node.position.x + (node.dimensions.width / 2);
        const nodeCenterY = node.position.y + (node.dimensions.height / 2);
        const panelOffset = (dimensions.value.width * 0.55) / 2;
        setCenter(nodeCenterX + panelOffset, nodeCenterY, { zoom: 1, duration: 300, interpolate: 'linear' });
      }
    } else if (!panel && oldPanel) {
      if (viewportBeforePanel) {
        setViewport(viewportBeforePanel, { duration: 300, interpolate: 'linear' });
        viewportBeforePanel = null;
      } else {
        handleFitView(300);
      }
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
      fit-view-on-init
      :min-zoom="1"
      :max-zoom="1"
      :zoom-on-scroll="false"
      :zoom-on-double-click="false"
      :zoom-on-pinch="false"
      :pan-on-scroll="true"
      :snap-to-grid="true"
      :snap-grid="[gridSize, gridSize]"
      class="canvas-bg"
      @node-click="handleNodeClick"
      @pane-click="emit('pane-click')"
    >
      <template #node-service="nodeProps">
        <ServiceNode
          :data="nodeProps.data"
          :selected="nodeProps.selected"
          @select="openPanel({ type: 'service', id: nodeProps.id, label: nodeProps.data.name })"
          @select-volume="openPanel({ type: 'volume', id: nodeProps.data.volume.id, label: nodeProps.data.volume.name })"
        />
      </template>

      <template #node-database="nodeProps">
        <DatabaseNode
          :data="nodeProps.data"
          :selected="nodeProps.selected"
          @select="openPanel({ type: 'database', id: nodeProps.id, label: nodeProps.data.name })"
        />
      </template>

      <template #node-keyValueStore="nodeProps">
        <KeyValueStoreNode
          :data="nodeProps.data"
          :selected="nodeProps.selected"
          @select="openPanel({ type: 'keyValueStore', id: nodeProps.id, label: nodeProps.data.name })"
        />
      </template>

      <template #node-bucket="nodeProps">
        <BucketNode
          :data="nodeProps.data"
          :selected="nodeProps.selected"
          @select="openPanel({ type: 'bucket', id: nodeProps.id, label: nodeProps.data.name })"
        />
      </template>

      <template #node-volume="nodeProps">
        <VolumeNode
          :data="nodeProps.data"
          :selected="nodeProps.selected"
          @select="openPanel({ type: 'volume', id: nodeProps.id, label: nodeProps.data.name })"
          @mount="emit('mount-volume', nodeProps.id)"
        />
      </template>

      <template #node-agent="nodeProps">
        <AgentNode
          :data="nodeProps.data"
          :selected="nodeProps.selected"
          @select="openPanel({ type: 'agent', id: nodeProps.id, label: nodeProps.data.name })"
        />
      </template>

      <Background variant="dots" :gap="gridSize" :size="1" :offset="0.5" />
    </VueFlow>

    <div class="absolute right-4 top-4 flex items-center gap-2">
      <IconButton
        label="Fit view"
        variant="outline"
        size="icon"
        class="text-muted-foreground"
        @click="handleFitView()"
      >
        <Scan />
      </IconButton>
      <IconButton
        label="Reset layout"
        variant="outline"
        size="icon"
        class="text-muted-foreground"
        @click="handleResetLayout"
      >
        <WandSparkles />
      </IconButton>
      <Button
        variant="outline"
        @click="emit('create')"
      >
        <Plus />
        New
      </Button>
    </div>
  </div>
</template>

<style scoped>
.canvas-bg :deep(.vue-flow__node > *) {
  transition: rotate 150ms ease, opacity 150ms ease;
}

.canvas-bg :deep(.vue-flow__node.dragging > *) {
  rotate: 2deg;
  opacity: 0.9;
}

.canvas-bg :deep(.vue-flow__node.settling) {
  transition: transform 200ms ease-out;
}
</style>
