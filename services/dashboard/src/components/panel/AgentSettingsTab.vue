<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { Trash2 } from '@lucide/vue';
import { graphql } from '@/gql';
import type { Harness } from '@/gql/graphql';
import { harnesses, type Agent } from '@/lib/agents';
import { useUnsavedChanges } from '@/composables/useUnsavedChanges';
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import FormFooter from '@/components/FormFooter.vue';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { failureToast } from '@/components/ui/sonner';

const UpdateAgentDocument = graphql(`
  mutation UpdateAgent($agent: AgentID!, $harness: Harness!, $instructions: String!) {
    updateAgent(agent: $agent, harness: $harness, instructions: $instructions) {
      id
      harness
      instructions
      status
    }
  }
`);

const DeleteAgentDocument = graphql(`
  mutation DeleteAgent($agent: AgentID!) {
    deleteAgent(agent: $agent)
  }
`);

const props = defineProps<{
  agent: Agent;
}>();

const emit = defineEmits<{
  (e: 'removed'): void;
}>();

const harness = ref<Harness>(props.agent.harness);
const instructions = ref(props.agent.instructions);

watch(() => [props.agent.harness, props.agent.instructions] as const, ([nextHarness, nextInstructions]) => {
  harness.value = nextHarness;
  instructions.value = nextInstructions;
});

const dirty = computed(() => harness.value !== props.agent.harness || instructions.value !== props.agent.instructions);

useUnsavedChanges(dirty);

const { mutate: updateAgent, loading: saving } = useMutation(UpdateAgentDocument);
const { mutate: deleteAgent, loading: deleting } = useMutation(DeleteAgentDocument);

function discard() {
  harness.value = props.agent.harness;
  instructions.value = props.agent.instructions;
}

async function save() {
  try {
    await updateAgent({ agent: props.agent.id, harness: harness.value, instructions: instructions.value });
  } catch (e: unknown) {
    failureToast('Failed to save agent', e);
  }
}

async function handleDelete() {
  try {
    await deleteAgent({ agent: props.agent.id });
    emit('removed');
  } catch (e: unknown) {
    failureToast('Failed to delete agent', e);
  }
}
</script>

<template>
  <div class="space-y-6">
    <section class="space-y-2">
      <h3 class="text-sm font-medium text-muted-foreground">Agent</h3>

      <div class="overflow-hidden rounded-lg border">
        <fieldset :disabled="saving" class="min-w-0 divide-y">
          <div class="flex items-center gap-4 px-4 py-3">
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">Harness</p>
              <p class="text-xs text-muted-foreground">The coding agent that runs the conversation.</p>
            </div>
            <Select v-model="harness">
              <SelectTrigger class="w-44 shrink-0" aria-label="Harness">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="option in harnesses" :key="option.value" :value="option.value">
                  {{ option.label }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div class="space-y-2 px-4 py-3">
            <Label for="agent-instructions" class="text-sm font-medium text-foreground">Instructions</Label>
            <Textarea
              id="agent-instructions"
              v-model="instructions"
              rows="8"
              class="text-sm"
              placeholder="What the agent is for, and what it should and shouldn't do."
            />
            <p class="text-xs text-muted-foreground">Added to the harness's own instruction file, next to its system prompt.</p>
          </div>
        </fieldset>

        <FormFooter
          :dirty="dirty"
          :saving="saving"
          :hint="`Saving restarts ${agent.name}. Running sessions end.`"
          @save="save"
          @discard="discard"
        />
      </div>
    </section>

    <div class="flex items-center justify-between gap-4 rounded-lg border border-destructive/30 p-3">
      <div>
        <p class="text-sm font-medium text-foreground">Delete agent</p>
        <p class="text-xs text-muted-foreground">Its sessions and files are deleted. Connectors stay in the workspace.</p>
      </div>
      <ConfirmDialog
        :title="`Delete agent &quot;${agent.name}&quot;?`"
        description="Every session and file of the agent is deleted with it. This can't be undone."
        confirm-label="Delete agent"
        pending-label="Deleting…"
        :pending="deleting"
        :confirm-text="agent.name"
        destructive
        @confirm="handleDelete"
      >
        <template #trigger>
          <Button variant="destructive">
            <Trash2 />
            Delete
          </Button>
        </template>
      </ConfirmDialog>
    </div>
  </div>
</template>
