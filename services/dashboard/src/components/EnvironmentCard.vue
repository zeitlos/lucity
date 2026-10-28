<script setup lang="ts">
import { computed } from 'vue';
import type { RouteLocationRaw } from 'vue-router';
import ResourceIcon from '@/components/ResourceIcon.vue';
import { BucketStatus, DatabaseStatus, ResourceTier, ServiceStatus } from '@/gql/graphql';
import { serviceKind, type ResourceKind } from '@/lib/resources';
import { bucketStatus, databaseStatus, serviceStatus, type StatusTone } from '@/lib/status';

const props = defineProps<{
  environment: {
    id: string;
    name: string;
    resourceTier: ResourceTier;
    services: { id: string; name: string; status: ServiceStatus; sourceUrl: string }[];
    databases: { id: string; name: string; status: DatabaseStatus }[];
    keyValueStores: { id: string; name: string; status: DatabaseStatus }[];
    buckets: { id: string; name: string; status: BucketStatus }[];
    volumes: { id: string; name: string }[];
  };
  to: RouteLocationRaw;
}>();

interface Tile {
  id: string;
  kind: ResourceKind;
  name: string;
  tone?: StatusTone;
}

const tiles = computed<Tile[]>(() => [
  ...props.environment.services.map(service => ({ id: service.id, kind: serviceKind(service.sourceUrl), name: service.name, tone: serviceStatus(service.status).tone })),
  ...props.environment.databases.map(database => ({ id: database.id, kind: 'database' as const, name: database.name, tone: databaseStatus(database.status).tone })),
  ...props.environment.keyValueStores.map(store => ({ id: store.id, kind: 'keyValueStore' as const, name: store.name, tone: databaseStatus(store.status).tone })),
  ...props.environment.buckets.map(bucket => ({ id: bucket.id, kind: 'bucket' as const, name: bucket.name, tone: bucketStatus(bucket.status).tone })),
  ...props.environment.volumes.map(volume => ({ id: volume.id, kind: 'volume' as const, name: volume.name })),
]);
</script>

<template>
  <RouterLink
    :to="to"
    class="flex flex-col gap-3.5 rounded-lg border bg-card px-5 py-4.5 shadow-sm transition-colors hover:border-primary/50"
  >
    <div class="flex min-w-0 items-center gap-2.5">
      <span class="truncate text-base font-semibold text-foreground">{{ environment.name }}</span>
      <span class="shrink-0 rounded-full border px-2 py-px text-xs text-muted-foreground">
        {{ environment.resourceTier === ResourceTier.Production ? 'Production' : 'Eco' }}
      </span>
    </div>

    <div class="canvas-preview grid min-h-30 grid-cols-3 content-start gap-2 rounded-lg border border-border/60 p-3.5">
      <span
        v-for="tile in tiles"
        :key="tile.id"
        class="flex h-8 min-w-0 items-center gap-1.5 rounded-md border bg-card px-2 text-xs font-medium text-foreground shadow-2xs"
      >
        <span class="text-muted-foreground"><ResourceIcon :kind="tile.kind" :size="13" /></span>
        <span class="min-w-0 flex-1 truncate">{{ tile.name }}</span>
        <span
          v-if="tile.tone"
          class="size-1.5 shrink-0 rounded-full"
          :style="{ background: `var(--status-${tile.tone})` }"
        />
      </span>
      <span v-if="tiles.length === 0" class="col-span-full text-xs text-muted-foreground">No resources yet</span>
    </div>
  </RouterLink>
</template>

<style scoped>
.canvas-preview {
  background-color: color-mix(in oklab, var(--color-muted) 40%, transparent);
  background-image: radial-gradient(var(--color-border) 1px, transparent 1px);
  background-size: 14px 14px;
}
</style>
