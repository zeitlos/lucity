<script setup lang="ts">
import { computed } from 'vue';
import type { RouteLocationRaw } from 'vue-router';
import { FolderGit2 } from '@lucide/vue';
import ResourceTile from '@/components/ResourceTile.vue';
import { repositoryName, type ResourceKind } from '@/lib/resources';

interface NamedResource {
  id: string;
  name: string;
}

interface Project {
  id: string;
  name: string;
  environments: {
    id: string;
    name: string;
    services: (NamedResource & { sourceUrl: string })[];
    databases: NamedResource[];
    keyValueStores: NamedResource[];
    buckets: NamedResource[];
    volumes: NamedResource[];
  }[];
}

const props = defineProps<{
  projects: Project[];
  linkTo: (project: Project) => RouteLocationRaw;
}>();

const columns = 'grid grid-cols-[minmax(0,1.2fr)_minmax(0,1.4fr)_minmax(0,1.2fr)_minmax(0,1.6fr)] items-center gap-6 px-4';

interface Tile {
  key: string;
  kind: ResourceKind;
  name: string;
}

function uniqueTiles(kind: ResourceKind, items: NamedResource[]): Tile[] {
  const names = [...new Set(items.map(item => item.name))];
  return names.map(name => ({ key: `${kind}/${name}`, kind, name }));
}

function tiles(project: Project) {
  const environments = project.environments;
  const services = environments.flatMap(environment => environment.services);
  return {
    services: [
      ...uniqueTiles('repositoryService', services.filter(service => service.sourceUrl)),
      ...uniqueTiles('imageService', services.filter(service => !service.sourceUrl)),
    ],
    data: [
      ...uniqueTiles('database', environments.flatMap(environment => environment.databases)),
      ...uniqueTiles('keyValueStore', environments.flatMap(environment => environment.keyValueStores)),
      ...uniqueTiles('bucket', environments.flatMap(environment => environment.buckets)),
      ...uniqueTiles('volume', environments.flatMap(environment => environment.volumes)),
    ],
  };
}

function repositories(project: Project) {
  const names = new Set<string>();
  for (const environment of project.environments) {
    for (const service of environment.services) {
      if (service.sourceUrl) names.add(repositoryName(service.sourceUrl));
    }
  }
  return [...names].join(', ');
}
const rows = computed(() => props.projects.map(project => ({
  project,
  repositories: repositories(project),
  tiles: tiles(project),
})));
</script>

<template>
  <div>
    <div class="sticky top-0 z-10 border-b bg-card">
      <div :class="[columns, 'h-8 bg-muted/40 text-xs font-medium text-muted-foreground']">
        <span>Project</span>
        <span>Repository</span>
        <span>Environments</span>
        <span>Resources</span>
      </div>
    </div>
    <RouterLink
      v-for="{ project, repositories: repositoryNames, tiles: resourceTiles } in rows"
      :key="project.id"
      :to="linkTo(project)"
      :class="[columns, 'min-h-13 border-b py-2 transition-colors hover:bg-muted/50']"
    >
      <span class="truncate text-sm font-semibold text-foreground">{{ project.name }}</span>
      <span class="flex min-w-0 items-center gap-1.5 text-xs text-muted-foreground">
        <template v-if="repositoryNames">
          <FolderGit2 :size="13" class="shrink-0" />
          <span class="truncate">{{ repositoryNames }}</span>
        </template>
      </span>
      <span class="flex flex-wrap gap-1.5">
        <span
          v-for="environment in project.environments"
          :key="environment.id"
          class="inline-flex h-5.5 items-center rounded-md bg-muted px-2 text-xs text-foreground"
        >
          {{ environment.name }}
        </span>
      </span>
      <span class="flex flex-wrap items-center gap-1">
        <template v-for="(group, index) in [resourceTiles.services, resourceTiles.data]" :key="index">
          <span v-if="index === 1 && group.length && resourceTiles.services.length" class="w-1" />
          <ResourceTile v-for="tile in group" :key="tile.key" :kind="tile.kind" :name="tile.name" />
        </template>
      </span>
    </RouterLink>
  </div>
</template>
