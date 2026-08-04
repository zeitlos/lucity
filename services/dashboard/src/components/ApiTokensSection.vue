<script setup lang="ts">
import { ref, computed } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Plus, Trash2, Copy, KeyRound } from '@lucide/vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';
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
        description: res.errors.map((e: { message: string }) => e.message).join(', '),
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
  <section class="space-y-6">
    <div>
      <h2 class="text-lg font-semibold text-foreground">API tokens</h2>
      <p class="text-sm text-muted-foreground">
        Machine credentials for CLI and CI automation, scoped to this workspace.
      </p>
    </div>

    <div class="flex items-end gap-2">
      <div class="flex-1 space-y-2">
        <Label for="apitoken-name">Name</Label>
        <Input id="apitoken-name" v-model="newName" placeholder="ci-deploy" :disabled="creating" />
      </div>
      <div class="w-32 space-y-2">
        <Label>Role</Label>
        <Select v-model="newRole">
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="USER">Member</SelectItem>
            <SelectItem value="ADMIN">Admin</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <Button :disabled="!newName.trim() || creating" @click="handleCreate">
        <Plus :size="14" />
        {{ creating ? 'Creating...' : 'Create' }}
      </Button>
    </div>

    <div v-if="loading" class="space-y-2">
      <Skeleton class="h-10 w-full" />
      <Skeleton class="h-10 w-full" />
    </div>
    <Table v-else-if="apiTokens.length">
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Role</TableHead>
          <TableHead>Created</TableHead>
          <TableHead class="w-10" />
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="token in apiTokens" :key="token.id">
          <TableCell class="font-medium">{{ token.name }}</TableCell>
          <TableCell>
            <Badge variant="secondary">{{ token.role === 'ADMIN' ? 'Admin' : 'Member' }}</Badge>
          </TableCell>
          <TableCell class="text-muted-foreground">{{ formatDate(token.createdAt) }}</TableCell>
          <TableCell>
            <Button
              variant="ghost"
              size="icon"
              :disabled="revoking"
              @click="tokenToRevoke = { id: token.id, name: token.name }"
            >
              <Trash2 :size="14" />
            </Button>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
    <div v-else class="rounded-lg border border-dashed p-8 text-center text-sm text-muted-foreground">
      <KeyRound :size="20" class="mx-auto mb-2 opacity-50" />
      No API tokens yet.
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
          <Button variant="outline" size="icon" class="shrink-0" @click="copyToken">
            <Copy :size="14" />
          </Button>
        </div>
        <DialogFooter>
          <Button @click="showTokenDialog = false">Done</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </section>
</template>
