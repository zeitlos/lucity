<script setup lang="ts">
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import { computed, ref } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { Earth, Laptop, Plus, ShieldAlert, Trash2 } from '@lucide/vue';
import { graphql } from '@/gql';
import { Button } from '@/components/ui/button';
import IconButton from '@/components/IconButton.vue';
import { Input } from '@/components/ui/input';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const ANY_ADDRESS = '0.0.0.0/0';

const AddDatabaseAllowRuleDocument = graphql(`
  mutation AddDatabaseAllowRule($database: DatabaseID!, $rule: DatabaseAllowRuleInput!) {
    addDatabaseAllowRule(database: $database, rule: $rule) {
      id
      allowRules {
        range
        description
      }
    }
  }
`);

const RemoveDatabaseAllowRuleDocument = graphql(`
  mutation RemoveDatabaseAllowRule($database: DatabaseID!, $range: String!) {
    removeDatabaseAllowRule(database: $database, range: $range) {
      id
      allowRules {
        range
        description
      }
    }
  }
`);

const props = defineProps<{
  databaseId: string;
  rules: { range: string; description: string }[];
  clientAddress: string | null;
}>();

const { mutate: addRule, loading: adding } = useMutation(AddDatabaseAllowRuleDocument);
const { mutate: removeRule } = useMutation(RemoveDatabaseAllowRuleDocument);

const removing = ref<string | null>(null);
const newRange = ref('');
const newDescription = ref('');
const anyAddressDialogOpen = ref(false);

const clientRange = computed(() => (props.clientAddress ? `${props.clientAddress}/32` : null));
const clientAllowed = computed(() => props.rules.some(rule => rule.range === clientRange.value));
const anyAllowed = computed(() => props.rules.some(rule => rule.range === ANY_ADDRESS));

async function allow(range: string, description: string) {
  try {
    await addRule({ database: props.databaseId, rule: { range, description } });
    return true;
  } catch (e: unknown) {
    errorToast('Failed to allow address', { description: errorMessage(e) });
    return false;
  }
}

async function allowClient() {
  if (!props.clientAddress) return;
  if (await allow(props.clientAddress, 'Added from the dashboard')) {
    toast.success(`${props.clientAddress} can now connect`);
  }
}

async function allowAnyAddress() {
  if (await allow(ANY_ADDRESS, 'Any address')) {
    anyAddressDialogOpen.value = false;
    toast.success('Any address can now connect');
  }
}

async function allowEntered() {
  const range = newRange.value.trim();
  if (!range) return;
  if (await allow(range, newDescription.value.trim())) {
    newRange.value = '';
    newDescription.value = '';
  }
}

async function remove(range: string) {
  removing.value = range;
  try {
    await removeRule({ database: props.databaseId, range });
  } catch (e: unknown) {
    errorToast('Failed to remove address', { description: errorMessage(e) });
  } finally {
    removing.value = null;
  }
}
</script>

<template>
  <div class="space-y-2">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <span class="text-xs font-medium text-muted-foreground">Allowed addresses</span>
      <div class="flex flex-wrap gap-1.5">
        <Button
          v-if="clientAddress && !clientAllowed"
          variant="outline"
          size="xs"
          :disabled="adding"
          @click="allowClient"
        >
          <Laptop />
          Add my IP ({{ clientAddress }})
        </Button>
        <Button
          v-if="!anyAllowed"
          variant="outline"
          size="xs"
          :disabled="adding"
          @click="anyAddressDialogOpen = true"
        >
          <Earth />
          Allow connections from anywhere
        </Button>
      </div>
    </div>

    <p
      v-if="rules.length === 0"
      class="rounded-md border border-dashed px-3 py-2 text-xs text-muted-foreground"
    >
      Nobody can connect over the public hostname yet. Allow your own address or a range below.
    </p>

    <div
      v-for="rule in rules"
      :key="rule.range"
      class="group flex items-center gap-2 rounded-md bg-muted/40 px-3 py-2"
    >
      <template v-if="rule.range === ANY_ADDRESS">
        <ShieldAlert :size="12" class="shrink-0 text-muted-foreground" />
        <span class="font-mono text-xs text-foreground">{{ rule.range }}</span>
        <span class="min-w-0 flex-1 truncate text-xs text-muted-foreground">
          Any address on the internet
        </span>
      </template>
      <template v-else>
        <span class="font-mono text-xs text-foreground">{{ rule.range }}</span>
        <span class="min-w-0 flex-1 truncate text-xs text-muted-foreground">{{ rule.description }}</span>
      </template>
      <IconButton
        :label="`Remove ${rule.range}`"
        size="icon-xs"
        class="shrink-0"
        :disabled="removing === rule.range"
        @click="remove(rule.range)"
      >
        <Trash2 />
      </IconButton>
    </div>

    <form class="flex gap-1.5" @submit.prevent="allowEntered">
      <fieldset :disabled="adding" class="flex min-w-0 flex-1 gap-1.5">
        <Input
          v-model="newRange"
          placeholder="203.0.113.7 or 203.0.113.0/24"
          class="flex-1 font-mono text-xs"
          aria-label="Address or range"
        />
        <Input
          v-model="newDescription"
          placeholder="Description (optional)"
          maxlength="64"
          class="flex-1 text-xs"
          aria-label="Description"
        />
      </fieldset>
      <Button type="submit" :pending="adding" :disabled="!newRange.trim()">
        <Plus />
        {{ adding ? 'Allowing…' : 'Allow' }}
      </Button>
    </form>

    <ConfirmDialog
      v-model:open="anyAddressDialogOpen"
      title="Allow connections from anywhere?"
      description="Anyone on the internet can try to connect, and only the database password keeps them out. Prefer allowing just the addresses you connect from."
      confirm-label="Allow any address"
      pending-label="Allowing…"
      :pending="adding"
      :icon="ShieldAlert"
      @confirm="allowAnyAddress"
    />
  </div>
</template>
