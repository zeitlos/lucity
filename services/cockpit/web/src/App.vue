<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted } from 'vue';
import { useGlobalMutationLoading, useGlobalQueryLoading, useQuery } from '@vue/apollo-composable';

import BrandMark from '@/components/BrandMark.vue';
import EnvironmentPill from '@/components/EnvironmentPill.vue';
import PlatformView from '@/components/PlatformView.vue';
import ViewDock from '@/components/ViewDock.vue';
import WorkspacesView from '@/components/WorkspacesView.vue';
import { graphql } from '@/gql';
import { apolloClient } from '@/lib/apollo';
import { dock, views } from '@/lib/dock';
import { environment } from '@/lib/environment';

const EnvironmentDocument = graphql(`
  query Environment {
    environment
  }
`);

const queryLoading = useGlobalQueryLoading();
const mutationLoading = useGlobalMutationLoading();
const loading = computed(() => queryLoading.value || mutationLoading.value);

const { error, onResult } = useQuery(EnvironmentDocument);

onResult((query) => {
  if (query.data && !environment.value) environment.value = query.data.environment;
});

const title = computed(() => views.find((view) => view.key === dock.view)!.name);

async function switchTo(name: string) {
  if (environment.value === name) return;

  await apolloClient.clearStore();
  environment.value = name;
}

function onKeydown(event: KeyboardEvent) {
  const view = views[Number(event.key) - 1];

  if (!event.metaKey || event.altKey || event.shiftKey || event.ctrlKey || !view) return;

  event.preventDefault();
  dock.view = view.key;
}

onMounted(() => window.addEventListener('keydown', onKeydown));
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown));
</script>

<template>
  <div class="flex h-screen flex-col">
    <div
      v-if="environment"
      class="pointer-events-none fixed inset-x-0 top-0 z-50 h-0.75"
      :class="environment === 'prod'
        ? 'bg-linear-to-r from-environment-prod to-status-danger shadow-[0_1px_8px_oklch(0.56_0.21_25/0.45)]'
        : 'bg-linear-to-r from-environment-dev to-[oklch(0.52_0.13_245)] shadow-[0_1px_8px_oklch(0.56_0.13_245/0.4)]'"
      aria-hidden="true"
    />
    <header class="flex h-13 shrink-0 items-center gap-2 border-b bg-sidebar pr-4 pl-24">
      <BrandMark :loading="loading" />
      <span v-if="loading" class="sr-only">Loading</span>
      <span class="mx-1 h-4 w-px bg-border" aria-hidden="true" />
      <h1 class="font-semibold">{{ title }}</h1>
      <div class="ml-auto">
        <EnvironmentPill :current="environment" @switch="switchTo" />
      </div>
    </header>
    <div v-if="environment" :key="environment" class="flex min-h-0 flex-1 flex-col">
      <KeepAlive>
        <WorkspacesView v-if="dock.view === 'workspaces'" />
        <PlatformView v-else />
      </KeepAlive>
      <ViewDock />
    </div>
    <main v-else class="flex-1 bg-card p-4">
      <div v-if="error" class="rounded-sm border border-status-danger/40 px-4 py-3">
        <p class="font-medium text-status-danger">Could not connect</p>
        <p class="mt-1 text-muted-foreground">{{ error.message }}</p>
      </div>
      <p v-else class="text-muted-foreground">Connecting…</p>
    </main>
  </div>
</template>
