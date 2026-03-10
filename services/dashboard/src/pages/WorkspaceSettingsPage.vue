<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useQuery, useMutation, useLazyQuery } from '@vue/apollo-composable';
import { ArrowLeft, Trash2, Github, UserPlus, X, Shield, User as UserIcon, CreditCard, ExternalLink, Plus } from 'lucide-vue-next';
import {
  WorkspaceQuery,
  WorkspacesQuery,
  UpdateWorkspaceMutation,
  DeleteWorkspaceMutation,
  InviteMemberMutation,
  RemoveMemberMutation,
  UpdateMemberRoleMutation,
  LinkGithubInstallationMutation,
  UnlinkGithubInstallationMutation,
} from '@/graphql/workspaces';
import { GitHubInstallationsQuery } from '@/graphql/github';
import {
  SubscriptionQuery,
  UsageSummaryQuery,
  ChangePlanMutation,
  BillingPortalUrlMutation,
} from '@/graphql/billing';
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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar';
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
    { id: 'billing', label: 'Billing' },
  ];
  if (isAdmin.value && !workspace.value?.personal) {
    s.push({ id: 'danger', label: 'Danger Zone' });
  }
  return s;
});

// Update workspace name
const editName = ref('');
const nameInitialized = ref(false);

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

// Installation picker
const pickerOpen = ref(false);
const {
  result: installationsResult,
  loading: installationsLoading,
  load: loadInstallations,
  error: installationsError,
} = useLazyQuery(GitHubInstallationsQuery);
const installations = computed(() => installationsResult.value?.githubInstallations ?? []);

watch(pickerOpen, (open) => {
  if (open) loadInstallations();
});

const { mutate: linkMutate, loading: linking } = useMutation(LinkGithubInstallationMutation);

async function handleLinkInstallation(installationId: string) {
  try {
    await linkMutate({ installationId });
    toast.success('GitHub connected');
    pickerOpen.value = false;
    refetch();
  } catch (e: unknown) {
    toast.error('Failed to connect GitHub', { description: errorMessage(e) });
  }
}

const githubManageUrl = computed(() => {
  if (!workspace.value?.githubLinked) return null;
  const id = workspace.value.githubInstallationId;
  const login = workspace.value.githubAccountLogin;
  const type = workspace.value.githubAccountType;
  if (!id) return null;
  if (type === 'ORGANIZATION' && login) {
    return `https://github.com/organizations/${login}/settings/installations/${id}`;
  }
  return `https://github.com/settings/installations/${id}`;
});

// Billing
const { result: subResult, loading: subLoading, error: subError } = useQuery(SubscriptionQuery);
const subscription = computed(() => subResult.value?.subscription);

const { result: usageResult, loading: usageLoading } = useQuery(UsageSummaryQuery);
const usage = computed(() => usageResult.value?.usageSummary);

const billingAvailable = computed(() => !subError.value && subscription.value);

const { mutate: changePlanMutate, loading: changingPlan } = useMutation(ChangePlanMutation, {
  refetchQueries: () => [{ query: SubscriptionQuery }, { query: UsageSummaryQuery }],
});
const { mutate: portalMutate, loading: openingPortal } = useMutation(BillingPortalUrlMutation);
const confirmPlan = ref<string | null>(null);

function formatCents(cents: number): string {
  return `€${(cents / 100).toFixed(2)}`;
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString(undefined, { year: 'numeric', month: 'long', day: 'numeric' });
}

async function handleChangePlan() {
  if (!confirmPlan.value) return;
  try {
    await changePlanMutate({ plan: confirmPlan.value });
    toast.success(`Switched to ${confirmPlan.value === 'PRO' ? 'Pro' : 'Hobby'} plan`);
  } catch (e: unknown) {
    toast.error('Failed to change plan', { description: errorMessage(e) });
  } finally {
    confirmPlan.value = null;
  }
}

async function handleOpenPortal() {
  try {
    const res = await portalMutate();
    const url = res?.data?.billingPortalUrl?.url;
    if (url) {
      window.open(url, '_blank');
    }
  } catch (e: unknown) {
    toast.error('Failed to open billing portal', { description: errorMessage(e) });
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
                  <Avatar
                    v-if="workspace.githubLinked && workspace.githubAccountAvatarUrl"
                    class="h-10 w-10"
                  >
                    <AvatarImage
                      :src="workspace.githubAccountAvatarUrl"
                      :alt="workspace.githubAccountLogin"
                    />
                    <AvatarFallback>
                      <Github :size="20" />
                    </AvatarFallback>
                  </Avatar>
                  <Github v-else :size="20" />

                  <div>
                    <p class="text-sm font-medium">
                      <template v-if="workspace.githubLinked && workspace.githubAccountLogin">
                        Connected to {{ workspace.githubAccountLogin }}
                      </template>
                      <template v-else-if="workspace.githubLinked">
                        GitHub connected
                      </template>
                      <template v-else>
                        GitHub not connected
                      </template>
                    </p>
                    <p class="text-xs text-muted-foreground">
                      <template v-if="workspace.githubLinked">
                        Your workspace has access to repositories from this installation.
                      </template>
                      <template v-else>
                        Connect GitHub to import repositories and enable automated deployments.
                      </template>
                    </p>
                  </div>
                </div>

                <div class="flex items-center gap-2">
                  <template v-if="workspace.githubLinked">
                    <a
                      v-if="githubManageUrl"
                      :href="githubManageUrl"
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      <Button variant="outline" size="sm">
                        <ExternalLink :size="14" class="mr-1.5" />
                        Manage
                      </Button>
                    </a>
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
                    <Badge v-else-if="!isAdmin" variant="default">Connected</Badge>
                  </template>

                  <template v-else-if="isAdmin">
                    <Dialog v-model:open="pickerOpen">
                      <DialogTrigger as-child>
                        <Button size="sm">
                          <Github :size="14" class="mr-1.5" />
                          Connect GitHub
                        </Button>
                      </DialogTrigger>
                      <DialogContent class="sm:max-w-md">
                        <DialogHeader>
                          <DialogTitle>Connect GitHub</DialogTitle>
                          <DialogDescription>
                            Select a GitHub account or organization to connect.
                          </DialogDescription>
                        </DialogHeader>

                        <div v-if="installationsLoading" class="space-y-3 py-4">
                          <Skeleton class="h-12 w-full" />
                          <Skeleton class="h-12 w-full" />
                        </div>

                        <div
                          v-else-if="installationsError"
                          class="py-6 text-center"
                        >
                          <p class="text-sm text-muted-foreground">
                            Failed to load installations.
                          </p>
                        </div>

                        <div
                          v-else-if="installations.length === 0"
                          class="py-6 text-center"
                        >
                          <Github
                            :size="24"
                            class="mx-auto mb-3 text-muted-foreground"
                          />
                          <p class="text-sm font-medium">
                            No installations found
                          </p>
                          <p class="mt-1 text-xs text-muted-foreground">
                            Install the Lucity GitHub App on your organization or account first.
                          </p>
                          <a
                            :href="`/auth/github/install?workspace=${workspace.id}`"
                            class="mt-3 inline-flex"
                          >
                            <Button size="sm">
                              <Github :size="14" class="mr-1.5" />
                              Install GitHub App
                            </Button>
                          </a>
                        </div>

                        <div v-else class="max-h-64 space-y-1 overflow-y-auto py-2">
                          <button
                            v-for="inst in installations"
                            :key="inst.id"
                            class="flex w-full items-center gap-3 rounded-md px-3 py-2.5 text-left transition-colors hover:bg-accent"
                            :disabled="linking"
                            @click="handleLinkInstallation(inst.id)"
                          >
                            <Avatar class="h-8 w-8">
                              <AvatarImage
                                :src="inst.accountAvatarUrl"
                                :alt="inst.accountLogin"
                              />
                              <AvatarFallback>
                                {{ inst.accountLogin[0]?.toUpperCase() }}
                              </AvatarFallback>
                            </Avatar>
                            <div>
                              <p class="text-sm font-medium">
                                {{ inst.accountLogin }}
                              </p>
                              <p class="text-xs text-muted-foreground">
                                {{ inst.accountType === 'ORGANIZATION' ? 'Organization' : 'Personal account' }}
                              </p>
                            </div>
                          </button>

                          <Separator class="my-2" />
                          <a
                            :href="`/auth/github/install?workspace=${workspace.id}`"
                            class="flex w-full items-center gap-3 rounded-md px-3 py-2 text-left text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                          >
                            <Plus :size="16" class="ml-1.5" />
                            Install on a different account
                          </a>
                        </div>
                      </DialogContent>
                    </Dialog>
                  </template>
                </div>
              </div>
            </div>
          </section>

          <!-- Billing -->
          <section v-if="activeSection === 'billing'" class="space-y-6">
            <div>
              <h2 class="text-lg font-medium">Billing</h2>
              <p class="text-sm text-muted-foreground">Manage your subscription, plan, and usage.</p>
            </div>

            <template v-if="subLoading">
              <Skeleton class="h-32 w-full" />
              <Skeleton class="h-24 w-full" />
            </template>

            <template v-else-if="!billingAvailable">
              <div class="rounded-lg border p-4">
                <div class="flex items-center gap-3">
                  <CreditCard :size="20" class="text-muted-foreground" />
                  <div>
                    <p class="text-sm font-medium">Billing is not configured</p>
                    <p class="text-xs text-muted-foreground">
                      Billing is not available for this workspace. Contact your administrator to set up billing.
                    </p>
                  </div>
                </div>
              </div>
            </template>

            <template v-else>
              <!-- Subscription -->
              <div class="rounded-lg border p-4 space-y-3">
                <div class="flex items-center justify-between">
                  <h3 class="text-sm font-medium">Subscription</h3>
                  <Badge :variant="subscription!.status === 'ACTIVE' ? 'default' : 'destructive'">
                    {{ subscription!.status === 'ACTIVE' ? 'Active' : subscription!.status === 'PAST_DUE' ? 'Past Due' : subscription!.status }}
                  </Badge>
                </div>
                <div class="flex items-center justify-between text-sm">
                  <span class="text-muted-foreground">Current plan</span>
                  <span class="font-medium">{{ subscription!.plan === 'PRO' ? 'Pro' : 'Hobby' }}</span>
                </div>
                <div class="flex items-center justify-between text-sm">
                  <span class="text-muted-foreground">Current period ends</span>
                  <span>{{ formatDate(subscription!.currentPeriodEnd) }}</span>
                </div>
                <div class="flex items-center justify-between text-sm">
                  <span class="text-muted-foreground">Plan credit</span>
                  <span>{{ formatCents(subscription!.creditAmountCents) }}/mo</span>
                </div>
              </div>

              <!-- Plan switcher (admin only) -->
              <div v-if="isAdmin" class="space-y-3">
                <h3 class="text-sm font-medium">Plan</h3>
                <div class="grid grid-cols-2 gap-3">
                  <button
                    class="rounded-lg border p-4 text-left transition-colors"
                    :class="subscription!.plan === 'HOBBY'
                      ? 'border-primary bg-primary/5'
                      : 'hover:border-muted-foreground/50'"
                    :disabled="subscription!.plan === 'HOBBY' || changingPlan"
                    @click="confirmPlan = 'HOBBY'"
                  >
                    <p class="text-sm font-medium">Hobby</p>
                    <p class="text-lg font-semibold">€5<span class="text-sm font-normal text-muted-foreground">/mo</span></p>
                    <p class="mt-1 text-xs text-muted-foreground">€5 resource credit included</p>
                  </button>
                  <button
                    class="rounded-lg border p-4 text-left transition-colors"
                    :class="subscription!.plan === 'PRO'
                      ? 'border-primary bg-primary/5'
                      : 'hover:border-muted-foreground/50'"
                    :disabled="subscription!.plan === 'PRO' || changingPlan"
                    @click="confirmPlan = 'PRO'"
                  >
                    <p class="text-sm font-medium">Pro</p>
                    <p class="text-lg font-semibold">€25<span class="text-sm font-normal text-muted-foreground">/mo</span></p>
                    <p class="mt-1 text-xs text-muted-foreground">€25 resource credit included</p>
                  </button>
                </div>
              </div>

              <!-- Plan change confirmation -->
              <AlertDialog :open="!!confirmPlan">
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>
                      Switch to {{ confirmPlan === 'PRO' ? 'Pro' : 'Hobby' }}?
                    </AlertDialogTitle>
                    <AlertDialogDescription>
                      Your plan will be changed to {{ confirmPlan === 'PRO' ? 'Pro (€25/mo)' : 'Hobby (€5/mo)' }}.
                      The change takes effect immediately with prorated billing.
                    </AlertDialogDescription>
                  </AlertDialogHeader>
                  <AlertDialogFooter>
                    <AlertDialogCancel @click="confirmPlan = null">Cancel</AlertDialogCancel>
                    <AlertDialogAction :disabled="changingPlan" @click="handleChangePlan">
                      {{ changingPlan ? 'Switching...' : 'Confirm' }}
                    </AlertDialogAction>
                  </AlertDialogFooter>
                </AlertDialogContent>
              </AlertDialog>

              <!-- Usage summary -->
              <div class="rounded-lg border p-4 space-y-3">
                <h3 class="text-sm font-medium">Current Period Usage</h3>
                <template v-if="usageLoading">
                  <Skeleton class="h-4 w-full" />
                  <Skeleton class="h-4 w-full" />
                  <Skeleton class="h-4 w-full" />
                </template>
                <template v-else-if="usage">
                  <div class="flex items-center justify-between text-sm">
                    <span class="text-muted-foreground">Resource costs</span>
                    <span>{{ formatCents(usage.resourceCostCents) }}</span>
                  </div>
                  <div class="flex items-center justify-between text-sm">
                    <span class="text-muted-foreground">Credits applied</span>
                    <span class="text-green-600">-{{ formatCents(usage.creditsCents) }}</span>
                  </div>
                  <Separator />
                  <div class="flex items-center justify-between text-sm font-medium">
                    <span>Estimated total</span>
                    <span>{{ formatCents(usage.estimatedTotalCents) }}</span>
                  </div>
                </template>
              </div>

              <!-- Billing portal -->
              <div v-if="isAdmin" class="rounded-lg border p-4">
                <div class="flex items-center justify-between">
                  <div>
                    <p class="text-sm font-medium">Billing portal</p>
                    <p class="text-xs text-muted-foreground">
                      Manage payment methods, view invoices, and update billing details.
                    </p>
                  </div>
                  <Button variant="outline" size="sm" :disabled="openingPortal" @click="handleOpenPortal">
                    <ExternalLink :size="14" class="mr-1.5" />
                    {{ openingPortal ? 'Opening...' : 'Open portal' }}
                  </Button>
                </div>
              </div>
            </template>
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
