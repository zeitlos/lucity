<script setup lang="ts">
import { ref, computed } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Copy, KeyRound } from '@lucide/vue';
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
import IconButton from '@/components/IconButton.vue';
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
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const { result, loading, refetch } = useQuery(ApiTokensDocument);
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
    if (res?.errors?.length) {
      errorToast('Failed to create API token', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }
    createdToken.value = res?.data?.createApiToken?.token ?? null;
    showTokenDialog.value = true;
    newName.value = '';
    newRole.value = WorkspaceRole.User;
    refetch();
  } catch (e: unknown) {
    errorToast('Failed to create API token', { description: errorMessage(e) });
  }
}

async function handleRevoke() {
  if (!tokenToRevoke.value) return;
  try {
    await revokeMutate({ id: tokenToRevoke.value.id });
    toast.success('API token revoked');
    tokenToRevoke.value = null;
    refetch();
  } catch (e: unknown) {
    errorToast('Failed to revoke API token', { description: errorMessage(e) });
  }
}

async function copyToken() {
  if (!createdToken.value) return;
  await navigator.clipboard.writeText(createdToken.value);
  toast.success('Copied to clipboard');
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
}
</script>

<template>
  <ContextDrawer
    title="API tokens"
    @close="emit('close')"
  >
    <div class="space-y-5">
      <form class="space-y-2" @submit.prevent="handleCreate">
        <Label for="apitoken-name" class="text-[13px] text-muted-foreground">New token</Label>
        <div class="flex items-center gap-2">
          <Input
            id="apitoken-name"
            v-model="newName"
            placeholder="e.g. github-actions"
            class="h-9 min-w-0 flex-1 font-mono text-xs"
            :disabled="creating"
          />
          <Select v-model="newRole">
            <SelectTrigger class="h-9 w-28 shrink-0 text-[13px]" aria-label="Role">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem :value="WorkspaceRole.User">Member</SelectItem>
              <SelectItem :value="WorkspaceRole.Admin">Admin</SelectItem>
            </SelectContent>
          </Select>
          <Button type="submit" class="shrink-0" :disabled="!newName.trim() || creating">
            {{ creating ? 'Creating...' : 'Create' }}
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
      <div v-else-if="apiTokens.length" class="divide-y overflow-hidden rounded-lg border">
        <div
          v-for="token in apiTokens"
          :key="token.id"
          class="flex items-center gap-3 py-3 pl-3.5 pr-2"
        >
          <div class="min-w-0 flex-1">
            <p class="truncate font-mono text-[13px] text-foreground">{{ token.name }}</p>
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
      <div v-else class="rounded-lg border border-dashed p-8 text-center text-sm text-muted-foreground">
        <KeyRound :size="20" class="mx-auto mb-2 opacity-50" />
        No API tokens yet.
      </div>
    </div>

    <AlertDialog :open="!!tokenToRevoke">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Revoke API token?</AlertDialogTitle>
          <AlertDialogDescription>
            "{{ tokenToRevoke?.name }}" will stop working immediately. This cannot be undone.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel :disabled="revoking" @click="tokenToRevoke = null">Cancel</AlertDialogCancel>
          <Button variant="destructive" :disabled="revoking" @click="handleRevoke">
            {{ revoking ? 'Revoking...' : 'Revoke' }}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

    <Dialog v-model:open="showTokenDialog">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>API token created</DialogTitle>
          <DialogDescription>Copy it now. You won't be able to see it again.</DialogDescription>
        </DialogHeader>
        <div class="flex min-w-0 items-start gap-2">
          <code class="min-w-0 flex-1 break-all rounded bg-muted px-2 py-1.5 font-mono text-xs">{{ createdToken }}</code>
          <IconButton label="Copy token" variant="outline" size="icon" class="shrink-0" @click="copyToken">
            <Copy />
          </IconButton>
        </div>
        <DialogFooter>
          <Button @click="showTokenDialog = false">Done</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </ContextDrawer>
</template>
