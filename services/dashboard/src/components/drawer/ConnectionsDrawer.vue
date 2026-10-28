<script setup lang="ts">
import { computed, ref } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { ChevronRight, Plug, Plus, Sparkles } from '@lucide/vue';
import { graphql } from '@/gql';
import { CredentialStatus } from '@/gql/graphql';
import { modelApiLabel, resourceName } from '@/lib/agents';
import { credentialStatus } from '@/lib/status';
import BrandTile from '@/components/BrandTile.vue';
import EmptyState from '@/components/EmptyState.vue';
import NewConnectionDialog from '@/components/NewConnectionDialog.vue';
import QueryError from '@/components/QueryError.vue';
import ContextDrawer from '@/components/drawer/ContextDrawer.vue';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Status } from '@/components/ui/status';

const ConnectionsDocument = graphql(`
  query Connections {
    modelProviders {
      id
      name
      api
      status
      models
    }
    connectors {
      id
      name
      url
      status
      tools {
        name
      }
    }
  }
`);

defineProps<{
  workspaceName: string;
  isAdmin: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const { result, loading, error, refetch } = useQuery(ConnectionsDocument);

const modelProviders = computed(() => result.value?.modelProviders ?? []);
const connectors = computed(() => result.value?.connectors ?? []);

const newOpen = ref(false);

function plural(count: number, noun: string) {
  return `${count} ${noun}${count === 1 ? '' : 's'}`;
}

function attention(status: CredentialStatus, credential: string) {
  const view = credentialStatus(status, credential);
  return { label: view.label, tone: status === CredentialStatus.Valid ? 'neutral' as const : view.tone };
}
</script>

<template>
  <ContextDrawer title="Connections" @close="emit('close')">
    <template v-if="isAdmin" #actions>
      <Button size="xs" @click="newOpen = true">
        <Plus />
        New connection
      </Button>
    </template>

    <div class="space-y-8">
      <p class="text-sm text-muted-foreground">
        Model providers and connectors for every project in {{ workspaceName }}. Agents use them without ever seeing a key.
      </p>

      <div v-if="loading && !result" class="space-y-8">
        <div v-for="i in 2" :key="i" class="space-y-2">
          <Skeleton class="h-4 w-32" />
          <Skeleton class="h-14 w-full" />
          <Skeleton class="h-14 w-full" />
        </div>
      </div>

      <QueryError
        v-else-if="error && !result"
        title="Couldn't load connections"
        :error="error"
        @retry="refetch()"
      />

      <template v-else>
        <section class="space-y-4">
          <div class="space-y-1 text-sm text-muted-foreground/80">
            <h3 class="text-sm font-medium text-muted-foreground">Model providers</h3>
            <p>Every agent can use their models, as far as its harness supports them.</p>
          </div>

          <EmptyState
            v-if="modelProviders.length === 0"
            size="compact"
            :icon="Sparkles"
            title="No model providers yet"
            :description="isAdmin ? 'Connect Anthropic, OpenAI or any OpenAI-compatible API.' : 'An admin of the workspace connects them.'"
          />
          <div v-else class="-mx-4 divide-y border-y">
            <RouterLink
              v-for="provider in modelProviders"
              :key="provider.id"
              :to="{ name: 'connection', params: { kind: 'model-providers', name: resourceName(provider.id) } }"
              class="flex items-center gap-3 py-2.5 pr-3 pl-4 transition-colors hover:bg-muted/50"
            >
              <BrandTile :name="provider.name" />
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm font-medium text-foreground">{{ provider.name }}</p>
                <p class="truncate text-xs text-muted-foreground">
                  {{ modelApiLabel(provider.api) }} · {{ plural(provider.models.length, 'model') }}
                </p>
              </div>
              <Status variant="text" :tone="attention(provider.status, 'Key').tone" class="shrink-0 text-xs">
                {{ attention(provider.status, 'Key').label }}
              </Status>
              <ChevronRight :size="16" class="shrink-0 text-muted-foreground" />
            </RouterLink>
          </div>
        </section>

        <section class="space-y-4">
          <div class="space-y-1 text-sm text-muted-foreground/80">
            <h3 class="text-sm font-medium text-muted-foreground">Connectors</h3>
            <p>Tools agents call through Lucity, once they are granted them.</p>
          </div>

          <EmptyState
            v-if="connectors.length === 0"
            size="compact"
            :icon="Plug"
            title="No connectors yet"
            :description="isAdmin ? 'Connect a remote MCP server, like Linear or GitHub.' : 'An admin of the workspace connects them.'"
          />
          <div v-else class="-mx-4 divide-y border-y">
            <RouterLink
              v-for="connector in connectors"
              :key="connector.id"
              :to="{ name: 'connection', params: { kind: 'connectors', name: resourceName(connector.id) } }"
              class="flex items-center gap-3 py-2.5 pr-3 pl-4 transition-colors hover:bg-muted/50"
            >
              <BrandTile :name="connector.name" />
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm font-medium text-foreground">{{ connector.name }}</p>
                <p class="truncate text-xs text-muted-foreground">
                  MCP server · {{ plural(connector.tools.length, 'tool') }}
                </p>
              </div>
              <Status variant="text" :tone="attention(connector.status, 'Credential').tone" class="shrink-0 text-xs">
                {{ attention(connector.status, 'Credential').label }}
              </Status>
              <ChevronRight :size="16" class="shrink-0 text-muted-foreground" />
            </RouterLink>
          </div>
        </section>
      </template>
    </div>

    <NewConnectionDialog v-if="isAdmin" v-model:open="newOpen" @created="refetch()" />
  </ContextDrawer>
</template>
