<script setup lang="ts">
import { computed, ref } from 'vue';
import { useMutation, useQuery } from '@vue/apollo-composable';
import { Plug, Trash2 } from '@lucide/vue';
import { graphql } from '@/gql';
import { Approval, type AgentConnectorInput } from '@/gql/graphql';
import { approvals, type Agent } from '@/lib/agents';
import BrandTile from '@/components/BrandTile.vue';
import EmptyState from '@/components/EmptyState.vue';
import IconButton from '@/components/IconButton.vue';
import QueryError from '@/components/QueryError.vue';
import Spinner from '@/components/LoadingSpinner.vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import { Switch } from '@/components/ui/switch';
import { Select, SelectContent, SelectItem, SelectSeparator, SelectTrigger, SelectValue } from '@/components/ui/select';
import { failureToast } from '@/components/ui/sonner';

const WorkspaceConnectorsDocument = graphql(`
  query WorkspaceConnectors {
    connectors {
      id
      name
      url
      status
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

const SetAgentNetworkDocument = graphql(`
  mutation SetAgentNetwork($agent: AgentID!, $allowedHosts: [String!]!, $privateNetwork: Boolean!) {
    setAgentNetwork(agent: $agent, allowedHosts: $allowedHosts, privateNetwork: $privateNetwork) {
      id
      allowedHosts
      privateNetwork
    }
  }
`);

const props = defineProps<{
  agent: Agent;
  serviceNames: string[];
}>();

const OFF = 'off';

const { result, loading, error, refetch } = useQuery(WorkspaceConnectorsDocument);
const connectors = computed(() => result.value?.connectors ?? []);

const { mutate: setConnectors, loading: savingConnectors } = useMutation(SetAgentConnectorsDocument);
const { mutate: setNetwork, loading: savingNetwork } = useMutation(SetAgentNetworkDocument);

const savingConnector = ref<string | null>(null);

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
  } catch (e: unknown) {
    failureToast('Failed to change connector access', e);
  } finally {
    savingConnector.value = null;
  }
}

const hostDraft = ref('');
const removingHost = ref<string | null>(null);

const canAddHost = computed(() => {
  const host = hostDraft.value.trim().toLowerCase();
  return host.length > 0 && !props.agent.allowedHosts.includes(host);
});

async function saveNetwork(allowedHosts: string[], privateNetwork: boolean) {
  await setNetwork({ agent: props.agent.id, allowedHosts, privateNetwork });
}

async function addHost() {
  if (!canAddHost.value) return;
  const host = hostDraft.value.trim().toLowerCase();
  try {
    await saveNetwork([...props.agent.allowedHosts, host], props.agent.privateNetwork);
    hostDraft.value = '';
  } catch (e: unknown) {
    failureToast('Failed to add host', e);
  }
}

async function removeHost(host: string) {
  removingHost.value = host;
  try {
    await saveNetwork(props.agent.allowedHosts.filter(existing => existing !== host), props.agent.privateNetwork);
  } catch (e: unknown) {
    failureToast('Failed to remove host', e);
  } finally {
    removingHost.value = null;
  }
}

const togglingPrivateNetwork = ref(false);

async function togglePrivateNetwork(value: boolean) {
  togglingPrivateNetwork.value = true;
  try {
    await saveNetwork(props.agent.allowedHosts, value);
  } catch (e: unknown) {
    failureToast('Failed to change private network', e);
  } finally {
    togglingPrivateNetwork.value = false;
  }
}

const privateNetworkHint = computed(() => (props.serviceNames.length
  ? `Reach ${props.serviceNames.join(', ')} directly.`
  : 'Reach the services of this environment directly.'));
</script>

<template>
  <div class="space-y-8">
    <section class="space-y-4">
      <div class="flex items-start justify-between gap-4">
        <div class="space-y-1 text-sm text-muted-foreground/80">
          <h3 class="text-sm font-medium text-muted-foreground">Connectors</h3>
          <p>Credentials stay with the workspace. {{ agent.name }} never sees them.</p>
        </div>
        <Button variant="outline" size="xs" class="shrink-0" as-child>
          <RouterLink :to="{ name: 'workspace', params: { drawer: 'connections' } }">Manage connectors</RouterLink>
        </Button>
      </div>

      <div v-if="loading && !result" class="space-y-2">
        <Skeleton v-for="i in 3" :key="i" class="h-12 w-full" />
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
        <div
          v-for="connector in connectors"
          :key="connector.id"
          class="flex items-center gap-3 px-4 py-2.5"
        >
          <BrandTile :name="connector.name" />
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-medium text-foreground">{{ connector.name }}</p>
            <p class="truncate font-mono text-xs text-muted-foreground">{{ connector.url }}</p>
          </div>
          <Spinner v-if="savingConnector === connector.id" :size="14" class="shrink-0 text-muted-foreground" />
          <Select
            :model-value="grantFor(connector.id)?.approval ?? OFF"
            :disabled="savingConnectors"
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
        </div>
      </div>
    </section>

    <section class="space-y-4">
      <div class="space-y-1 text-sm text-muted-foreground/80">
        <h3 class="text-sm font-medium text-muted-foreground">Allowed hosts</h3>
        <p>Hosts on the internet {{ agent.name }} may reach directly, without credentials. Everything else is blocked.</p>
      </div>

      <div v-if="agent.allowedHosts.length" class="-mx-4 divide-y border-y">
        <div
          v-for="host in agent.allowedHosts"
          :key="host"
          class="flex items-center gap-3 py-2 pr-3 pl-4"
        >
          <span class="min-w-0 flex-1 truncate font-mono text-sm text-foreground">{{ host }}</span>
          <IconButton
            :label="`Remove ${host}`"
            size="icon-xs"
            class="shrink-0"
            :pending="removingHost === host"
            :disabled="savingNetwork"
            @click="removeHost(host)"
          >
            <Trash2 />
          </IconButton>
        </div>
      </div>

      <form class="flex items-center gap-2" @submit.prevent="addHost">
        <Input
          v-model="hostDraft"
          placeholder="api.example.com or *.example.com"
          class="h-9 min-w-0 flex-1 font-mono text-sm"
          aria-label="Host"
          autocomplete="off"
          spellcheck="false"
          :disabled="savingNetwork"
        />
        <Button
          type="submit"
          variant="outline"
          class="shrink-0"
          :pending="savingNetwork && !removingHost && !togglingPrivateNetwork"
          :disabled="!canAddHost || savingNetwork"
        >
          Add
        </Button>
      </form>
    </section>

    <section class="space-y-2">
      <h3 class="text-sm font-medium text-muted-foreground">Private network</h3>
      <div class="overflow-hidden rounded-lg border">
        <div class="flex items-center gap-3 px-4 py-3">
          <div class="min-w-0 flex-1">
            <p class="text-sm font-medium text-foreground">Services in this environment</p>
            <p class="text-xs text-muted-foreground">{{ privateNetworkHint }}</p>
          </div>
          <Spinner v-if="togglingPrivateNetwork" :size="14" class="shrink-0 text-muted-foreground" />
          <Switch
            :model-value="agent.privateNetwork"
            :disabled="savingNetwork"
            class="data-[state=unchecked]:bg-border"
            aria-label="Private network"
            @update:model-value="togglePrivateNetwork"
          />
        </div>
      </div>
    </section>
  </div>
</template>
