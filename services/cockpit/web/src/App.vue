<script setup lang="ts">
import { ChevronsUpDown } from '@lucide/vue';
import { useQuery } from '@vue/apollo-composable';

import WorkspacesView from '@/components/WorkspacesView.vue';
import { graphql } from '@/gql';
import { apolloClient } from '@/lib/apollo';
import { environment, environments } from '@/lib/environment';

const EnvironmentDocument = graphql(`
  query Environment {
    environment
  }
`);

const { error, onResult } = useQuery(EnvironmentDocument);

onResult((query) => {
  if (query.data && !environment.value) environment.value = query.data.environment;
});

async function switchTo(name: string) {
  if (environment.value === name) return;

  await apolloClient.clearStore();
  environment.value = name;
}
</script>

<template>
  <div class="flex h-screen flex-col">
    <header class="flex h-13 shrink-0 items-center gap-2 border-b bg-sidebar pr-4 pl-24">
      <svg viewBox="0 0 250 181" class="h-[13px] w-[18px]" aria-hidden="true">
        <path
          class="fill-brand-mark"
          d="M 130 0 L 130 76 C 130 102.64, 114.4 112, 70 112 L 60 112 C 15.6 112, 0 121.36, 0 148 L 0 181 L 169 181 L 169 148 C 169 121.36, 178.23 112, 204.5 112 L 214.5 112 C 240.77 112, 250 102.64, 250 76 L 250 0 Z"
        />
      </svg>
      <h1 class="font-semibold">Workspaces</h1>
      <label class="relative ml-auto flex items-center">
        <span class="sr-only">Environment</span>
        <select
          :value="environment"
          :disabled="!environment"
          class="h-7 appearance-none rounded-sm border pr-7 pl-2.5 font-mono font-medium shadow-xs"
          :class="environment === 'prod'
            ? 'border-status-danger/40 bg-status-danger-background text-status-danger'
            : 'bg-card'"
          @change="switchTo(($event.target as HTMLSelectElement).value)"
        >
          <option v-for="name in environments" :key="name" :value="name">lucity-{{ name }}</option>
        </select>
        <ChevronsUpDown
          class="pointer-events-none absolute right-2 size-3.5"
          :class="environment === 'prod' ? 'text-status-danger' : 'text-muted-foreground'"
          aria-hidden="true"
        />
      </label>
    </header>
    <WorkspacesView v-if="environment" :key="environment" />
    <main v-else class="flex-1 bg-card p-4">
      <div v-if="error" class="rounded-sm border border-status-danger/40 px-4 py-3">
        <p class="font-medium text-status-danger">Could not connect</p>
        <p class="mt-1 text-muted-foreground">{{ error.message }}</p>
      </div>
      <p v-else class="text-muted-foreground">Connecting…</p>
    </main>
  </div>
</template>
