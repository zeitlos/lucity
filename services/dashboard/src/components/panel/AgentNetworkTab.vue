<script setup lang="ts">
import { computed, ref } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { Globe, Trash2 } from '@lucide/vue';
import { graphql } from '@/gql';
import type { Agent } from '@/lib/agents';
import EmptyState from '@/components/EmptyState.vue';
import IconButton from '@/components/IconButton.vue';
import Spinner from '@/components/LoadingSpinner.vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import { failureToast } from '@/components/ui/sonner';

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

const { mutate: setNetwork, loading: saving } = useMutation(SetAgentNetworkDocument);

const hostDraft = ref('');
const removingHost = ref<string | null>(null);
const togglingPrivateNetwork = ref(false);

const draftHost = computed(() => hostDraft.value.trim().toLowerCase());
const duplicateHost = computed(() => props.agent.allowedHosts.includes(draftHost.value));
const canAddHost = computed(() => draftHost.value.length > 0 && !duplicateHost.value);

async function saveNetwork(allowedHosts: string[], privateNetwork: boolean) {
  await setNetwork({ agent: props.agent.id, allowedHosts, privateNetwork });
}

async function addHost() {
  if (!canAddHost.value) return;
  try {
    await saveNetwork([...props.agent.allowedHosts, draftHost.value], props.agent.privateNetwork);
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
      <div class="space-y-1 text-sm text-muted-foreground/80">
        <h3 class="text-sm font-medium text-muted-foreground">Allowed hosts</h3>
        <p>Everything else on the internet is blocked.</p>
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
            :disabled="saving"
            @click="removeHost(host)"
          >
            <Trash2 />
          </IconButton>
        </div>
      </div>
      <EmptyState
        v-else
        size="compact"
        :icon="Globe"
        title="No allowed hosts"
        description="This agent reaches nothing on the internet directly."
      />

      <div class="space-y-2">
        <form class="flex items-center gap-2" @submit.prevent="addHost">
          <Input
            v-model="hostDraft"
            placeholder="api.example.com or *.example.com"
            class="h-9 min-w-0 flex-1 font-mono text-sm"
            aria-label="Host"
            autocomplete="off"
            spellcheck="false"
            :disabled="saving"
          />
          <Button
            type="submit"
            variant="outline"
            class="shrink-0"
            :pending="saving && !removingHost && !togglingPrivateNetwork"
            :disabled="!canAddHost || saving"
          >
            Add
          </Button>
        </form>
        <p v-if="duplicateHost" class="text-xs text-muted-foreground">"{{ draftHost }}" is already allowed.</p>
      </div>
    </section>

    <section class="space-y-4">
      <div class="-mx-4 flex items-center gap-3 border-y px-4 py-3">
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium text-foreground">Private network</p>
          <p class="text-xs text-muted-foreground">{{ privateNetworkHint }}</p>
        </div>
        <Spinner v-if="togglingPrivateNetwork" :size="14" class="shrink-0 text-muted-foreground" />
        <Switch
          :model-value="agent.privateNetwork"
          :disabled="saving"
          class="data-[state=unchecked]:bg-border"
          aria-label="Private network"
          @update:model-value="togglePrivateNetwork"
        />
      </div>
      <p class="text-sm text-muted-foreground/80">Models and connectors are reached through Lucity and need no host here.</p>
    </section>
  </div>
</template>
