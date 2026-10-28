<script setup lang="ts">
import { formatDate } from '@/lib/format';
import CopyButton from '@/components/CopyButton.vue';
import QueryError from '@/components/QueryError.vue';
import EmptyState from '@/components/EmptyState.vue';
import DialogIcon from '@/components/DialogIcon.vue';
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import { ref, computed } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { KeyRound } from '@lucide/vue';
import { graphql } from '@/gql';
import { WorkspaceRole } from '@/gql/graphql';

const ApiTokensDocument = graphql(`
  query ApiTokens {
    apiTokens {
      id
      name
      role
      createdAt
    }
  }
`);

const CreateApiTokenDocument = graphql(`
  mutation CreateApiToken($input: CreateApiTokenInput!) {
    createApiToken(input: $input) {
      apiToken {
        id
        name
        role
        createdAt
      }
      token
    }
  }
`);

const RevokeApiTokenDocument = graphql(`
  mutation RevokeApiToken($id: ID!) {
    revokeApiToken(id: $id)
  }
`);
import ContextDrawer from '@/components/drawer/ContextDrawer.vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { failureToast } from '@/components/ui/sonner';

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const { result, loading, error, refetch } = useQuery(ApiTokensDocument);
const apiTokens = computed(() => result.value?.apiTokens ?? []);

const newName = ref('');
const newRole = ref<WorkspaceRole>(WorkspaceRole.User);
const createdToken = ref<string | null>(null);
const showTokenDialog = ref(false);

const { mutate: createMutate, loading: creating } = useMutation(CreateApiTokenDocument);
const { mutate: revokeMutate, loading: revoking } = useMutation(RevokeApiTokenDocument);
const tokenToRevoke = ref<{ id: string; name: string } | null>(null);

async function handleCreate() {
  if (!newName.value.trim()) return;
  try {
    const res = await createMutate({ input: { name: newName.value.trim(), role: newRole.value } });
    createdToken.value = res?.data?.createApiToken?.token ?? null;
    showTokenDialog.value = true;
    newName.value = '';
    newRole.value = WorkspaceRole.User;
    refetch();
  } catch (e: unknown) {
    failureToast('Failed to create API token', e);
  }
}

async function handleRevoke() {
  if (!tokenToRevoke.value) return;
  try {
    await revokeMutate({ id: tokenToRevoke.value.id });
    tokenToRevoke.value = null;
    refetch();
  } catch (e: unknown) {
    failureToast('Failed to revoke API token', e);
  }
}
</script>

<template>
  <ContextDrawer
    title="API tokens"
    @close="emit('close')"
  >
    <div class="space-y-5">
      <form class="space-y-2" @submit.prevent="handleCreate">
        <Label for="apitoken-name" class="text-sm text-muted-foreground">New token</Label>
        <div class="flex items-center gap-2">
          <Input
            id="apitoken-name"
            v-model="newName"
            placeholder="e.g. github-actions"
            class="h-9 min-w-0 flex-1 font-mono text-xs"
            :disabled="creating"
          />
          <Select v-model="newRole">
            <SelectTrigger class="h-9 w-28 shrink-0" aria-label="Role">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem :value="WorkspaceRole.User">Member</SelectItem>
              <SelectItem :value="WorkspaceRole.Admin">Admin</SelectItem>
            </SelectContent>
          </Select>
          <Button type="submit" class="shrink-0" :pending="creating" :disabled="!newName.trim()">
            {{ creating ? 'Creating…' : 'Create' }}
          </Button>
        </div>
        <p class="text-xs text-muted-foreground">
          For the CLI and CI. A token acts with the role you give it.
        </p>
      </form>

      <div v-if="loading" class="space-y-2">
        <Skeleton class="h-12 w-full" />
        <Skeleton class="h-12 w-full" />
      </div>
      <QueryError
        v-else-if="error && !result"
        title="Couldn't load API tokens"
        :error="error"
        @retry="refetch()"
      />
      <div v-else-if="apiTokens.length" class="-mx-4 divide-y border-y">
        <div
          v-for="token in apiTokens"
          :key="token.id"
          class="flex items-center gap-3 py-3 pr-3 pl-4"
        >
          <div class="min-w-0 flex-1">
            <p class="truncate font-mono text-sm text-foreground">{{ token.name }}</p>
            <p class="text-xs text-muted-foreground">
              {{ token.role === WorkspaceRole.Admin ? 'Admin' : 'Member' }} · created {{ formatDate(token.createdAt) }}
            </p>
          </div>
          <Button
            variant="ghost"
            size="xs"
            class="shrink-0 text-destructive hover:text-destructive"
            :aria-label="`Revoke ${token.name}`"
            :disabled="revoking"
            @click="tokenToRevoke = { id: token.id, name: token.name }"
          >
            Revoke
          </Button>
        </div>
      </div>
      <EmptyState
        v-else
        size="compact"
        :icon="KeyRound"
        title="No API tokens yet"
        description="Create one for the CLI or a CI pipeline."
      />
    </div>

    <ConfirmDialog
      :open="!!tokenToRevoke"
      :title="`Revoke API token &quot;${tokenToRevoke?.name}&quot;?`"
      description="It stops working right away. This can't be undone."
      confirm-label="Revoke token"
      pending-label="Revoking…"
      :pending="revoking"
      destructive
      @update:open="(value) => { if (!value) tokenToRevoke = null }"
      @confirm="handleRevoke"
    />

    <Dialog v-model:open="showTokenDialog">
      <DialogContent>
        <DialogHeader>
          <DialogIcon :icon="KeyRound" />
          <DialogTitle>API token created</DialogTitle>
          <DialogDescription>Copy it now. You won't be able to see it again.</DialogDescription>
        </DialogHeader>
        <div class="flex min-w-0 items-start gap-2">
          <code class="min-w-0 flex-1 break-all rounded bg-muted px-2 py-1.5 font-mono text-xs">{{ createdToken }}</code>
          <CopyButton label="Copy token" variant="outline" size="icon" class="shrink-0" :value="createdToken ?? ''" />
        </div>
        <DialogFooter>
          <Button @click="showTokenDialog = false">Done</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </ContextDrawer>
</template>
