<script setup lang="ts">
import { computed, ref } from 'vue';
import { useMutation, useQuery } from '@vue/apollo-composable';
import { ChevronDown, ChevronRight, Plug, Search } from '@lucide/vue';
import { graphql } from '@/gql';
import { Approval, CredentialStatus, type AgentConnectorInput } from '@/gql/graphql';
import { approvals, type Agent } from '@/lib/agents';
import { credentialStatus } from '@/lib/status';
import BrandTile from '@/components/BrandTile.vue';
import EmptyState from '@/components/EmptyState.vue';
import IconButton from '@/components/IconButton.vue';
import QueryError from '@/components/QueryError.vue';
import Spinner from '@/components/LoadingSpinner.vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import { Status } from '@/components/ui/status';
import { Switch } from '@/components/ui/switch';
import { Select, SelectContent, SelectItem, SelectSeparator, SelectTrigger, SelectValue } from '@/components/ui/select';
import { failureToast } from '@/components/ui/sonner';

const WorkspaceConnectorsDocument = graphql(`
  query WorkspaceConnectors {
    connectors {
      id
      name
      status
      tools {
        name
        description
      }
    }
  }
`);

const SetAgentConnectorsDocument = graphql(`
  mutation SetAgentConnectors($agent: AgentID!, $connectors: [AgentConnectorInput!]!) {
    setAgentConnectors(agent: $agent, connectors: $connectors) {
      id
      connectors {
        approval
        hiddenTools
        connector {
          id
          name
        }
      }
    }
  }
`);

const props = defineProps<{
  agent: Agent;
}>();

const OFF = 'off';

const { result, loading, error, refetch } = useQuery(WorkspaceConnectorsDocument);
const connectors = computed(() => result.value?.connectors ?? []);

const { mutate: setConnectors, loading: saving } = useMutation(SetAgentConnectorsDocument);

const savingConnector = ref<string | null>(null);
const savingTool = ref<string | null>(null);
const openConnector = ref<string | null>(null);
const filter = ref('');

function grantFor(connectorId: string) {
  return props.agent.connectors.find(grant => grant.connector.id === connectorId);
}

function grantInputs(): AgentConnectorInput[] {
  return props.agent.connectors.map(grant => ({
    connector: grant.connector.id,
    approval: grant.approval,
    hiddenTools: grant.hiddenTools,
  }));
}

function isHidden(connectorId: string, tool: string) {
  return grantFor(connectorId)?.hiddenTools.includes(tool) ?? false;
}

function toolSummary(connectorId: string, tools: { name: string }[]) {
  const count = `${tools.length} ${tools.length === 1 ? 'tool' : 'tools'}`;
  const hidden = tools.filter(tool => isHidden(connectorId, tool.name)).length;
  return hidden ? `${count}, ${hidden} hidden` : count;
}

function toggleOpen(connectorId: string) {
  openConnector.value = openConnector.value === connectorId ? null : connectorId;
  filter.value = '';
}

const visibleTools = computed(() => {
  const connector = connectors.value.find(candidate => candidate.id === openConnector.value);
  if (!connector) return [];
  const term = filter.value.trim().toLowerCase();
  if (!term) return connector.tools;
  return connector.tools.filter(tool => tool.name.toLowerCase().includes(term));
});

async function changeApproval(connectorId: string, value: unknown) {
  const next = grantInputs().filter(grant => grant.connector !== connectorId);
  if (value === Approval.Allow || value === Approval.Ask) {
    next.push({
      connector: connectorId,
      approval: value,
      hiddenTools: grantFor(connectorId)?.hiddenTools ?? [],
    });
  }

  savingConnector.value = connectorId;
  try {
    await setConnectors({ agent: props.agent.id, connectors: next });
    if (value === OFF && openConnector.value === connectorId) openConnector.value = null;
  } catch (e: unknown) {
    failureToast('Failed to change connector access', e);
  } finally {
    savingConnector.value = null;
  }
}

async function toggleTool(connectorId: string, tool: string, shown: boolean) {
  const next = grantInputs().map((grant) => {
    if (grant.connector !== connectorId) return grant;
    const hiddenTools = (grant.hiddenTools ?? []).filter(name => name !== tool);
    if (!shown) hiddenTools.push(tool);
    return { ...grant, hiddenTools };
  });

  savingTool.value = tool;
  try {
    await setConnectors({ agent: props.agent.id, connectors: next });
  } catch (e: unknown) {
    failureToast(shown ? 'Failed to show tool' : 'Failed to hide tool', e);
  } finally {
    savingTool.value = null;
  }
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between gap-4">
      <p class="text-sm text-muted-foreground/80">Credentials stay with the workspace.</p>
      <Button variant="outline" size="xs" class="shrink-0" as-child>
        <RouterLink :to="{ name: 'workspace', params: { drawer: 'connections' } }">Manage connectors</RouterLink>
      </Button>
    </div>

    <div v-if="loading && !result" class="-mx-4 divide-y border-y">
      <div v-for="width in ['w-16', 'w-40', 'w-24']" :key="width" class="flex items-center gap-3 py-2.5 pr-3 pl-4">
        <Skeleton class="size-8 rounded-sm" />
        <div class="min-w-0 flex-1">
          <div class="flex h-5 items-center">
            <Skeleton :class="['h-3', width]" />
          </div>
          <div class="flex h-4 items-center">
            <Skeleton class="h-2.5 w-16" />
          </div>
        </div>
        <div class="flex h-8 w-32 shrink-0 items-center justify-between rounded-md border border-input bg-background px-3">
          <Skeleton class="h-3 w-14" />
          <Skeleton class="size-3 rounded-sm" />
        </div>
        <span class="size-6 shrink-0" />
      </div>
    </div>
    <QueryError
      v-else-if="error && !result"
      title="Couldn't load connectors"
      :error="error"
      @retry="refetch()"
    />
    <EmptyState
      v-else-if="connectors.length === 0"
      size="compact"
      :icon="Plug"
      title="No connectors yet"
      description="An admin connects tools like Linear or GitHub under Connections."
    />
    <div v-else class="-mx-4 divide-y border-y">
      <template v-for="connector in connectors" :key="connector.id">
        <div class="flex items-center gap-3 py-2.5 pr-3 pl-4">
          <BrandTile :name="connector.name" />
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-medium text-foreground">{{ connector.name }}</p>
            <Status
              v-if="connector.status !== CredentialStatus.Valid"
              variant="text"
              :tone="credentialStatus(connector.status, 'Token').tone"
              class="text-xs"
            >
              {{ credentialStatus(connector.status, 'Token').label }}
            </Status>
            <p v-else class="truncate text-xs text-muted-foreground">{{ toolSummary(connector.id, connector.tools) }}</p>
          </div>
          <Spinner v-if="savingConnector === connector.id" :size="14" class="shrink-0 text-muted-foreground" />
          <Select
            :model-value="grantFor(connector.id)?.approval ?? OFF"
            :disabled="saving"
            @update:model-value="changeApproval(connector.id, $event)"
          >
            <SelectTrigger class="h-8 w-32 shrink-0 text-sm" :aria-label="`Access to ${connector.name}`">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="option in approvals" :key="option.value" :value="option.value">
                {{ option.label }}
              </SelectItem>
              <SelectSeparator />
              <SelectItem :value="OFF">Off</SelectItem>
            </SelectContent>
          </Select>
          <IconButton
            v-if="grantFor(connector.id)"
            :label="openConnector === connector.id ? `Hide tools of ${connector.name}` : `Show tools of ${connector.name}`"
            size="icon-xs"
            class="shrink-0"
            :aria-expanded="openConnector === connector.id"
            @click="toggleOpen(connector.id)"
          >
            <ChevronDown v-if="openConnector === connector.id" />
            <ChevronRight v-else />
          </IconButton>
          <span v-else class="size-6 shrink-0" />
        </div>

        <template v-if="openConnector === connector.id && grantFor(connector.id)">
          <p v-if="connector.tools.length === 0" class="py-2.5 pr-4 pl-15 text-sm text-muted-foreground">
            The server listed no tools at the last check.
          </p>
          <template v-else>
            <div class="relative py-2 pr-4 pl-15">
              <Search :size="14" class="pointer-events-none absolute top-1/2 left-17.5 -translate-y-1/2 text-muted-foreground" />
              <Input
                v-model="filter"
                placeholder="Filter tools"
                aria-label="Filter tools"
                class="h-8 pl-8 text-sm"
                autocomplete="off"
                spellcheck="false"
              />
            </div>
            <p v-if="visibleTools.length === 0" class="py-2.5 pr-4 pl-15 text-sm text-muted-foreground">
              No tool matches "{{ filter.trim() }}".
            </p>
            <div
              v-for="tool in visibleTools"
              :key="tool.name"
              class="flex items-center gap-3 py-2 pr-4 pl-15"
            >
              <div class="min-w-0 flex-1">
                <p :class="['truncate font-mono text-sm', isHidden(connector.id, tool.name) ? 'text-muted-foreground' : 'text-foreground']">
                  {{ tool.name }}
                </p>
                <p v-if="tool.description" class="truncate text-xs text-muted-foreground">{{ tool.description }}</p>
              </div>
              <Spinner v-if="savingTool === tool.name" :size="14" class="shrink-0 text-muted-foreground" />
              <Switch
                :model-value="!isHidden(connector.id, tool.name)"
                :disabled="saving"
                class="shrink-0 data-[state=unchecked]:bg-border"
                :aria-label="`Show ${tool.name} to this agent`"
                @update:model-value="toggleTool(connector.id, tool.name, $event)"
              />
            </div>
          </template>
        </template>
      </template>
    </div>
  </div>
</template>
