<script setup lang="ts">
import { computed } from 'vue';
import type { RouteLocationRaw } from 'vue-router';
import { FolderGit2 } from '@lucide/vue';
import ResourceTile from '@/components/ResourceTile.vue';
import { repositoryName, serviceKind, type ResourceKind } from '@/lib/resources';

interface NamedResource {
  id: string;
  name: string;
}

interface ProjectEnvironment {
  id: string;
  name: string;
  services: (NamedResource & { sourceUrl: string })[];
  databases: NamedResource[];
  keyValueStores: NamedResource[];
  buckets: NamedResource[];
  volumes: NamedResource[];
}

const props = defineProps<{
  project: {
    id: string;
    name: string;
    environments: ProjectEnvironment[];
  };
  to: RouteLocationRaw;
}>();

interface Tile {
  id: string;
  kind: ResourceKind;
  name: string;
}

const repositories = computed(() => {
  const names = new Set<string>();
  for (const environment of props.project.environments) {
    for (const service of environment.services) {
      if (service.sourceUrl) names.add(repositoryName(service.sourceUrl));
    }
  }
  return [...names].join(', ');
});

const rows = computed(() => props.project.environments.map(environment => ({
  environment,
  services: environment.services.map<Tile>(service => ({ id: service.id, kind: serviceKind(service.sourceUrl), name: service.name })),
  data: [
    ...environment.databases.map<Tile>(database => ({ id: database.id, kind: 'database', name: database.name })),
    ...environment.keyValueStores.map<Tile>(store => ({ id: store.id, kind: 'keyValueStore', name: store.name })),
    ...environment.buckets.map<Tile>(bucket => ({ id: bucket.id, kind: 'bucket', name: bucket.name })),
    ...environment.volumes.map<Tile>(volume => ({ id: volume.id, kind: 'volume', name: volume.name })),
  ],
})));
</script>

<template>
  <div class="relative flex flex-col gap-3.5 rounded-lg border bg-card px-5 py-4.5 shadow-sm transition-colors hover:border-primary/50">
    <div class="min-w-0 space-y-1.5">
      <RouterLink
        :to="to"
        class="block truncate text-base font-semibold text-foreground after:absolute after:inset-0 after:rounded-lg"
      >
        {{ project.name }}
      </RouterLink>
      <p v-if="repositories" class="flex min-w-0 items-center gap-1.5 text-xs text-muted-foreground">
        <FolderGit2 :size="13" class="shrink-0" />
        <span class="truncate">{{ repositories }}</span>
      </p>
    </div>

    <div class="space-y-0.5 border-t border-border/60 pt-3">
      <RouterLink
        v-for="row in rows"
        :key="row.environment.id"
        :to="{ name: 'environment', params: { projectId: project.id, environmentId: row.environment.id } }"
        class="relative z-10 -mx-1.5 grid min-h-7.5 grid-cols-[104px_minmax(0,1fr)] items-center gap-2.5 rounded-md px-1.5 py-1 transition-colors hover:bg-muted/60"
      >
        <span class="truncate text-xs text-foreground">{{ row.environment.name }}</span>
        <span v-if="row.services.length + row.data.length === 0" class="text-xs text-muted-foreground">No resources yet</span>
        <span v-else class="flex flex-wrap items-center gap-1">
          <template v-for="(group, index) in [row.services, row.data]" :key="index">
            <span v-if="index === 1 && row.services.length && row.data.length" class="w-1" />
            <ResourceTile v-for="tile in group" :key="tile.id" :kind="tile.kind" :name="tile.name" />
          </template>
        </span>
      </RouterLink>
    </div>
  </div>
</template>
