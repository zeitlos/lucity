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
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
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

const ApiKeysDocument = graphql(`
  query ApiKeys {
    apiKeys {
      id
      name
      role
      createdAt
    }
  }
`);

const CreateApiKeyDocument = graphql(`
  mutation CreateApiKey($input: CreateApiKeyInput!) {
    createApiKey(input: $input) {
      apiKey {
        id
        name
        role
        createdAt
      }
      key
    }
  }
`);

const RevokeApiKeyDocument = graphql(`
  mutation RevokeApiKey($id: ID!) {
    revokeApiKey(id: $id)
  }
`);

const { result, loading, refetch } = useQuery(ApiKeysDocument);
const apiKeys = computed(() => result.value?.apiKeys ?? []);

const newName = ref('');
const newRole = ref<WorkspaceRole>(WorkspaceRole.User);
const createdKey = ref<string | null>(null);
const showKeyDialog = ref(false);

const { mutate: createMutate, loading: creating } = useMutation(CreateApiKeyDocument);
const { mutate: revokeMutate } = useMutation(RevokeApiKeyDocument);

async function handleCreate() {
  if (!newName.value.trim()) return;
  try {
    const res = await createMutate({ input: { name: newName.value.trim(), role: newRole.value } });
    if (res?.errors?.length) {
      errorToast('Failed to create API key', {
        description: res.errors.map((e: { message: string }) => e.message).join(', '),
      });
      return;
    }
    createdKey.value = res?.data?.createApiKey?.key ?? null;
    showKeyDialog.value = true;
    newName.value = '';
    newRole.value = WorkspaceRole.User;
    refetch();
  } catch (e: unknown) {
    errorToast('Failed to create API key', { description: errorMessage(e) });
  }
}

async function handleRevoke(id: string) {
  try {
    await revokeMutate({ id });
    toast.success('API key revoked');
    refetch();
  } catch (e: unknown) {
    errorToast('Failed to revoke API key', { description: errorMessage(e) });
  }
}

async function copyKey() {
  if (!createdKey.value) return;
  await navigator.clipboard.writeText(createdKey.value);
  toast.success('Copied to clipboard');
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
}
</script>

<template>
  <section class="space-y-6">
    <div>
      <h2 class="text-lg font-semibold text-foreground">API keys</h2>
      <p class="text-sm text-muted-foreground">
        Machine credentials for CLI and CI automation, scoped to this workspace.
      </p>
    </div>

    <div class="flex items-end gap-2">
      <div class="flex-1 space-y-2">
        <Label for="apikey-name">Name</Label>
        <Input id="apikey-name" v-model="newName" placeholder="ci-deploy" :disabled="creating" />
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
        Create
      </Button>
    </div>

    <div v-if="loading" class="space-y-2">
      <Skeleton class="h-10 w-full" />
      <Skeleton class="h-10 w-full" />
    </div>
    <Table v-else-if="apiKeys.length">
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Role</TableHead>
          <TableHead>Created</TableHead>
          <TableHead class="w-10" />
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="key in apiKeys" :key="key.id">
          <TableCell class="font-medium">{{ key.name }}</TableCell>
          <TableCell>
            <Badge variant="secondary">{{ key.role === 'ADMIN' ? 'Admin' : 'Member' }}</Badge>
          </TableCell>
          <TableCell class="text-muted-foreground">{{ formatDate(key.createdAt) }}</TableCell>
          <TableCell>
            <AlertDialog>
              <AlertDialogTrigger as-child>
                <Button variant="ghost" size="icon">
                  <Trash2 :size="14" />
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Revoke API key?</AlertDialogTitle>
                  <AlertDialogDescription>
                    "{{ key.name }}" will stop working immediately. This cannot be undone.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction @click="handleRevoke(key.id)">Revoke</AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
    <div v-else class="rounded-lg border border-dashed p-8 text-center text-sm text-muted-foreground">
      <KeyRound :size="20" class="mx-auto mb-2 opacity-50" />
      No API keys yet.
    </div>

    <Dialog v-model:open="showKeyDialog">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>API key created</DialogTitle>
          <DialogDescription>Copy it now. You won't be able to see it again.</DialogDescription>
        </DialogHeader>
        <div class="flex min-w-0 items-start gap-2">
          <code class="min-w-0 flex-1 break-all rounded bg-muted px-2 py-1.5 font-mono text-xs">{{ createdKey }}</code>
          <Button variant="outline" size="icon" class="shrink-0" @click="copyKey">
            <Copy :size="14" />
          </Button>
        </div>
        <DialogFooter>
          <Button @click="showKeyDialog = false">Done</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </section>
</template>
