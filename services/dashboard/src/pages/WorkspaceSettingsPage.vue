<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { ArrowLeft, Trash2, Github, UserPlus, X, Shield, User as UserIcon } from 'lucide-vue-next';
import {
  WorkspaceQuery,
  WorkspacesQuery,
  UpdateWorkspaceMutation,
  DeleteWorkspaceMutation,
  InviteMemberMutation,
  RemoveMemberMutation,
  UpdateMemberRoleMutation,
  UnlinkGithubInstallationMutation,
} from '@/graphql/workspaces';
import { useAuth } from '@/composables/useAuth';
import { apolloClient } from '@/lib/apollo';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { toast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const router = useRouter();
const { refreshToken, setActiveWorkspace } = useAuth();

const { result, loading, refetch } = useQuery(WorkspaceQuery);
const workspace = computed(() => result.value?.workspace);
const members = computed(() => workspace.value?.members ?? []);
const isAdmin = computed(() => {
  const { user } = useAuth();
  if (!user.value || !workspace.value) return false;
  const membership = user.value.workspaces.find(w => w.workspace === workspace.value!.id);
  return membership?.role === 'admin';
});

// Settings sections
const activeSection = ref('general');
const sections = computed(() => {
  const s = [
    { id: 'general', label: 'General' },
    { id: 'members', label: 'Members' },
    { id: 'github', label: 'GitHub' },
  ];
  if (isAdmin.value && !workspace.value?.personal) {
    s.push({ id: 'danger', label: 'Danger Zone' });
  }
  return s;
});

// Update workspace name
const editName = ref('');
const nameInitialized = ref(false);

import { watch } from 'vue';
watch(
  () => workspace.value?.name,
  (name) => {
    if (name && !nameInitialized.value) {
      editName.value = name;
      nameInitialized.value = true;
    }
  },
  { immediate: true },
);

const { mutate: updateMutate, loading: updating } = useMutation(UpdateWorkspaceMutation);

async function handleUpdateName() {
  if (!editName.value.trim() || editName.value.trim() === workspace.value?.name) return;
  try {
    await updateMutate({ input: { name: editName.value.trim() } });
    toast.success('Workspace name updated');
    refetch();
  } catch (e: unknown) {
    toast.error('Failed to update workspace', { description: errorMessage(e) });
  }
}

// Invite member
const inviteEmail = ref('');
const inviteRole = ref('USER');
const { mutate: inviteMutate, loading: inviting } = useMutation(InviteMemberMutation);

async function handleInvite() {
  if (!inviteEmail.value.trim()) return;
  try {
    const res = await inviteMutate({
      input: { email: inviteEmail.value.trim(), role: inviteRole.value },
    });
    if (res?.errors?.length) {
      toast.error('Failed to invite member', {
        description: res.errors.map((e: { message: string }) => e.message).join(', '),
      });
      return;
    }
    toast.success(`Invited ${inviteEmail.value.trim()}`);
    inviteEmail.value = '';
    inviteRole.value = 'USER';
    await refreshToken();
    refetch();
  } catch (e: unknown) {
    toast.error('Failed to invite member', { description: errorMessage(e) });
  }
}

// Remove member
const { mutate: removeMutate } = useMutation(RemoveMemberMutation);

async function handleRemoveMember(userId: string) {
  try {
    await removeMutate({ userId });
    toast.success('Member removed');
    await refreshToken();
    refetch();
  } catch (e: unknown) {
    toast.error('Failed to remove member', { description: errorMessage(e) });
  }
}

// Update member role
const { mutate: updateRoleMutate } = useMutation(UpdateMemberRoleMutation);

async function handleUpdateRole(userId: string, role: string) {
  try {
    await updateRoleMutate({ input: { userId, role } });
    toast.success('Member role updated');
    refetch();
  } catch (e: unknown) {
    toast.error('Failed to update role', { description: errorMessage(e) });
  }
}

// GitHub
const { mutate: unlinkGithubMutate, loading: unlinking } = useMutation(UnlinkGithubInstallationMutation);

async function handleUnlinkGithub() {
  try {
    await unlinkGithubMutate();
    toast.success('GitHub disconnected');
    refetch();
  } catch (e: unknown) {
    toast.error('Failed to disconnect GitHub', { description: errorMessage(e) });
  }
}

// Delete workspace
const { mutate: deleteMutate, loading: deleting } = useMutation(DeleteWorkspaceMutation, {
  refetchQueries: () => [{ query: WorkspacesQuery }],
});

async function handleDelete() {
  try {
    const res = await deleteMutate();
    if (res?.errors?.length) {
      toast.error('Failed to delete workspace', {
        description: res.errors.map((e: { message: string }) => e.message).join(', '),
      });
      return;
    }
    await refreshToken();
    const { user } = useAuth();
    const firstWs = user.value?.workspaces[0]?.workspace;
    if (firstWs) {
      setActiveWorkspace(firstWs);
    }
    apolloClient.resetStore();
    toast.success('Workspace deleted');
    router.push({ name: 'projects' });
  } catch (e: unknown) {
    toast.error('Failed to delete workspace', { description: errorMessage(e) });
  }
}
</script>

<template>
  <div class="mx-auto max-w-4xl p-6">
    <div class="mb-6">
      <Button
        variant="ghost"
        size="sm"
        class="mb-4 text-muted-foreground"
        @click="router.push({ name: 'projects' })"
      >
        <ArrowLeft :size="14" class="mr-1.5" />
        Back
      </Button>
      <h1 class="text-2xl font-semibold">Workspace Settings</h1>
      <p class="text-sm text-muted-foreground">
        Manage your workspace, members, and integrations.
      </p>
    </div>

    <template v-if="loading">
      <Skeleton class="mb-4 h-8 w-64" />
      <Skeleton class="h-48 w-full" />
    </template>

    <template v-else-if="workspace">
      <div class="flex gap-8">
        <!-- Sidebar -->
        <nav class="w-40 shrink-0">
          <ul class="space-y-1">
            <li v-for="section in sections" :key="section.id">
              <button
                class="w-full rounded-md px-3 py-1.5 text-left text-sm transition-colors"
                :class="activeSection === section.id ? 'bg-accent font-medium' : 'text-muted-foreground hover:bg-accent/50'"
                @click="activeSection = section.id"
              >
                {{ section.label }}
              </button>
            </li>
          </ul>
        </nav>

        <!-- Content -->
        <div class="min-w-0 flex-1">
          <!-- General -->
          <section v-if="activeSection === 'general'" class="space-y-6">
            <div>
              <h2 class="text-lg font-medium">General</h2>
              <p class="text-sm text-muted-foreground">Basic workspace information.</p>
            </div>

            <div class="space-y-4">
              <div class="space-y-2">
                <Label>Workspace ID</Label>
                <div class="flex items-center gap-2">
                  <code class="rounded bg-muted px-2 py-1 text-sm">{{ workspace.id }}</code>
                  <Badge v-if="workspace.personal" variant="secondary">Personal</Badge>
                </div>
              </div>

              <div class="space-y-2">
                <Label for="ws-name-edit">Name</Label>
                <div class="flex items-center gap-2">
                  <Input
                    id="ws-name-edit"
                    v-model="editName"
                    :disabled="!isAdmin || updating"
                    class="max-w-sm"
                  />
                  <Button
                    v-if="isAdmin"
                    size="sm"
                    :disabled="!editName.trim() || editName.trim() === workspace.name || updating"
                    @click="handleUpdateName"
                  >
                    {{ updating ? 'Saving...' : 'Save' }}
                  </Button>
                </div>
              </div>
            </div>
          </section>

          <!-- Members -->
          <section v-if="activeSection === 'members'" class="space-y-6">
            <div>
              <h2 class="text-lg font-medium">Members</h2>
              <p class="text-sm text-muted-foreground">Manage who has access to this workspace.</p>
            </div>

            <!-- Invite form -->
            <div v-if="isAdmin" class="flex items-end gap-2">
              <div class="flex-1 space-y-2">
                <Label for="invite-email">Invite by email</Label>
                <Input
                  id="invite-email"
                  v-model="inviteEmail"
                  type="email"
                  placeholder="user@example.com"
                  :disabled="inviting"
                />
              </div>
              <div class="w-32 space-y-2">
                <Label>Role</Label>
                <Select v-model="inviteRole">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="USER">Member</SelectItem>
                    <SelectItem value="ADMIN">Admin</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <Button
                :disabled="!inviteEmail.trim() || inviting"
                @click="handleInvite"
              >
                <UserPlus :size="14" class="mr-1.5" />
                {{ inviting ? 'Inviting...' : 'Invite' }}
              </Button>
            </div>

            <Separator />

            <!-- Members table -->
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Member</TableHead>
                  <TableHead>Role</TableHead>
                  <TableHead v-if="isAdmin" class="w-24" />
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="member in members" :key="member.id">
                  <TableCell>
                    <div>
                      <p class="text-sm font-medium">{{ member.name || member.email }}</p>
                      <p v-if="member.name" class="text-xs text-muted-foreground">{{ member.email }}</p>
                    </div>
                  </TableCell>
                  <TableCell>
                    <template v-if="isAdmin">
                      <Select
                        :model-value="member.role"
                        @update:model-value="handleUpdateRole(member.id, $event as string)"
                      >
                        <SelectTrigger class="w-28">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="USER">
                            <div class="flex items-center gap-1.5">
                              <UserIcon :size="12" />
                              Member
                            </div>
                          </SelectItem>
                          <SelectItem value="ADMIN">
                            <div class="flex items-center gap-1.5">
                              <Shield :size="12" />
                              Admin
                            </div>
                          </SelectItem>
                        </SelectContent>
                      </Select>
                    </template>
                    <template v-else>
                      <Badge :variant="member.role === 'ADMIN' ? 'default' : 'secondary'">
                        {{ member.role === 'ADMIN' ? 'Admin' : 'Member' }}
                      </Badge>
                    </template>
                  </TableCell>
                  <TableCell v-if="isAdmin">
                    <AlertDialog>
                      <AlertDialogTrigger as-child>
                        <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground hover:text-destructive">
                          <X :size="14" />
                        </Button>
                      </AlertDialogTrigger>
                      <AlertDialogContent>
                        <AlertDialogHeader>
                          <AlertDialogTitle>Remove member?</AlertDialogTitle>
                          <AlertDialogDescription>
                            {{ member.name || member.email }} will lose access to this workspace.
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel>Cancel</AlertDialogCancel>
                          <AlertDialogAction @click="handleRemoveMember(member.id)">
                            Remove
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </section>

          <!-- GitHub -->
          <section v-if="activeSection === 'github'" class="space-y-6">
            <div>
              <h2 class="text-lg font-medium">GitHub</h2>
              <p class="text-sm text-muted-foreground">Connect a GitHub App installation to access repositories.</p>
            </div>

            <div class="rounded-lg border p-4">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <Github :size="20" />
                  <div>
                    <p class="text-sm font-medium">
                      {{ workspace.githubLinked ? 'GitHub connected' : 'GitHub not connected' }}
                    </p>
                    <p class="text-xs text-muted-foreground">
                      {{ workspace.githubLinked ? 'Your workspace has access to repositories from the installed GitHub App.' : 'Connect GitHub to import repositories and enable automated deployments.' }}
                    </p>
                  </div>
                </div>
                <div>
                  <template v-if="workspace.githubLinked">
                    <AlertDialog v-if="isAdmin">
                      <AlertDialogTrigger as-child>
                        <Button variant="outline" size="sm" :disabled="unlinking">
                          Disconnect
                        </Button>
                      </AlertDialogTrigger>
                      <AlertDialogContent>
                        <AlertDialogHeader>
                          <AlertDialogTitle>Disconnect GitHub?</AlertDialogTitle>
                          <AlertDialogDescription>
                            This will remove the GitHub App link. Existing deployments won't be affected, but you won't be able to import new repositories.
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel>Cancel</AlertDialogCancel>
                          <AlertDialogAction @click="handleUnlinkGithub">
                            Disconnect
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                    <Badge v-else variant="default">Connected</Badge>
                  </template>
                  <a
                    v-else-if="workspace.githubInstallUrl && isAdmin"
                    :href="`/auth/github/install?workspace=${workspace.id}`"
                    class="inline-flex"
                  >
                    <Button size="sm">
                      <Github :size="14" class="mr-1.5" />
                      Connect GitHub
                    </Button>
                  </a>
                </div>
              </div>
            </div>
          </section>

          <!-- Danger Zone -->
          <section v-if="activeSection === 'danger'" class="space-y-6">
            <div>
              <h2 class="text-lg font-medium text-destructive">Danger Zone</h2>
              <p class="text-sm text-muted-foreground">Irreversible actions.</p>
            </div>

            <div class="rounded-lg border border-destructive/50 p-4">
              <div class="flex items-center justify-between">
                <div>
                  <p class="text-sm font-medium">Delete workspace</p>
                  <p class="text-xs text-muted-foreground">
                    Permanently delete this workspace and all its data. This cannot be undone.
                  </p>
                </div>
                <AlertDialog>
                  <AlertDialogTrigger as-child>
                    <Button variant="destructive" size="sm" :disabled="deleting">
                      <Trash2 :size="14" class="mr-1.5" />
                      Delete
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>Delete workspace "{{ workspace.name }}"?</AlertDialogTitle>
                      <AlertDialogDescription>
                        This will permanently delete the workspace, remove all members, and unlink GitHub. All projects must be deleted first.
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>Cancel</AlertDialogCancel>
                      <AlertDialogAction
                        class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                        @click="handleDelete"
                      >
                        Delete workspace
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </div>
            </div>
          </section>
        </div>
      </div>
    </template>
  </div>
</template>
