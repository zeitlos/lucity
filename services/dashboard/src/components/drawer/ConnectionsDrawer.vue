<script setup lang="ts">
import { computed, ref } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { ChevronRight, MessageSquare, Plug, Plus, Sparkles } from '@lucide/vue';
import { graphql } from '@/gql';
import { CredentialStatus } from '@/gql/graphql';
import { chatAppKindLabel, modelApiLabel, resourceName } from '@/lib/agents';
import { credentialStatus } from '@/lib/status';
import BrandTile from '@/components/BrandTile.vue';
import EmptyState from '@/components/EmptyState.vue';
import NewChatAppDialog from '@/components/NewChatAppDialog.vue';
import NewConnectorDialog from '@/components/NewConnectorDialog.vue';
import NewModelProviderDialog from '@/components/NewModelProviderDialog.vue';
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
    chatApps {
      id
      name
      kind
      account
      url
      status
    }
  }
`);

defineProps<{
  isAdmin: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const { result, loading, error, refetch } = useQuery(ConnectionsDocument);

const modelProviders = computed(() => result.value?.modelProviders ?? []);
const connectors = computed(() => result.value?.connectors ?? []);
const chatApps = computed(() => result.value?.chatApps ?? []);

const skeletonSections = [
  {
    title: 'Model providers',
    description: 'Every agent can use their models, as far as its harness supports them.',
    rows: ['w-24', 'w-16'],
  },
  {
    title: 'Connectors',
    description: 'Tools agents call through Lucity, once they are granted them.',
    rows: ['w-20', 'w-28'],
  },
  {
    title: 'Chat apps',
    description: 'Bots agents talk through, in the chats they are added to.',
    rows: ['w-24'],
  },
];

const newModelProviderOpen = ref(false);
const newConnectorOpen = ref(false);
const newChatAppOpen = ref(false);

function chatAppHost(url: string | null | undefined) {
  if (!url) return '';
  try {
    return new URL(url).host;
  } catch {
    return url;
  }
}

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

    <div class="space-y-8">
      <p class="text-sm text-muted-foreground">
        Model providers, connectors and chat apps for every project in this workspace. Agents use them without ever seeing a key.
      </p>

      <template v-if="loading && !result">
        <section v-for="section in skeletonSections" :key="section.title" class="space-y-4">
          <div class="space-y-1 text-sm text-muted-foreground/80">
            <h3 class="text-sm font-medium text-muted-foreground">{{ section.title }}</h3>
            <p>{{ section.description }}</p>
          </div>
          <div class="-mx-4 divide-y border-y">
            <div v-for="width in section.rows" :key="width" class="flex items-center gap-3 py-2.5 pr-3 pl-4">
              <Skeleton class="size-8 rounded-sm" />
              <div class="min-w-0 flex-1">
                <div class="flex h-5 items-center">
                  <Skeleton :class="['h-3', width]" />
                </div>
                <div class="flex h-4 items-center">
                  <Skeleton class="h-2.5 w-32" />
                </div>
              </div>
              <Skeleton class="h-2.5 w-10" />
              <Skeleton class="size-4 rounded-sm" />
            </div>
          </div>
        </section>
      </template>

      <QueryError
        v-else-if="error && !result"
        title="Couldn't load connections"
        :error="error"
        @retry="refetch()"
      />

      <template v-else>
        <section class="space-y-4">
          <div class="flex items-start justify-between gap-4">
            <div class="space-y-1 text-sm text-muted-foreground/80">
              <h3 class="text-sm font-medium text-muted-foreground">Model providers</h3>
              <p>Every agent can use their models, as far as its harness supports them.</p>
            </div>
            <Button v-if="isAdmin" variant="outline" size="xs" class="shrink-0" @click="newModelProviderOpen = true">
              <Plus />
              New model provider
            </Button>
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
          <div class="flex items-start justify-between gap-4">
            <div class="space-y-1 text-sm text-muted-foreground/80">
              <h3 class="text-sm font-medium text-muted-foreground">Connectors</h3>
              <p>Tools agents call through Lucity, once they are granted them.</p>
            </div>
            <Button v-if="isAdmin" variant="outline" size="xs" class="shrink-0" @click="newConnectorOpen = true">
              <Plus />
              New connector
            </Button>
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

        <section class="space-y-4">
          <div class="flex items-start justify-between gap-4">
            <div class="space-y-1 text-sm text-muted-foreground/80">
              <h3 class="text-sm font-medium text-muted-foreground">Chat apps</h3>
              <p>Bots agents talk through, in the chats they are added to.</p>
            </div>
            <Button v-if="isAdmin" variant="outline" size="xs" class="shrink-0" @click="newChatAppOpen = true">
              <Plus />
              New chat app
            </Button>
          </div>

          <EmptyState
            v-if="chatApps.length === 0"
            size="compact"
            :icon="MessageSquare"
            title="No chat apps yet"
            :description="isAdmin ? 'Connect a Telegram or Mattermost bot.' : 'An admin of the workspace connects them.'"
          />
          <div v-else class="-mx-4 divide-y border-y">
            <RouterLink
              v-for="app in chatApps"
              :key="app.id"
              :to="{ name: 'connection', params: { kind: 'chat-apps', name: resourceName(app.id) } }"
              class="flex items-center gap-3 py-2.5 pr-3 pl-4 transition-colors hover:bg-muted/50"
            >
              <BrandTile :name="chatAppKindLabel(app.kind)" />
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm font-medium text-foreground">{{ app.name }}</p>
                <p class="truncate text-xs text-muted-foreground">
                  {{ chatAppKindLabel(app.kind) }} bot · {{ app.account }}<template v-if="app.url"> on {{ chatAppHost(app.url) }}</template>
                </p>
              </div>
              <Status variant="text" :tone="attention(app.status, 'Token').tone" class="shrink-0 text-xs">
                {{ attention(app.status, 'Token').label }}
              </Status>
              <ChevronRight :size="16" class="shrink-0 text-muted-foreground" />
            </RouterLink>
          </div>
        </section>
      </template>
    </div>

    <template v-if="isAdmin">
      <NewModelProviderDialog v-model:open="newModelProviderOpen" @created="refetch()" />
      <NewConnectorDialog v-model:open="newConnectorOpen" @created="refetch()" />
      <NewChatAppDialog v-model:open="newChatAppOpen" @created="refetch()" />
    </template>
  </ContextDrawer>
</template>
