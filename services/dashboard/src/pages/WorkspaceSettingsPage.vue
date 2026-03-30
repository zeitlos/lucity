<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { ArrowLeft, Trash2, UserPlus, X, Shield, User as UserIcon, CreditCard, ExternalLink } from 'lucide-vue-next';
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
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';
import PlanPicker from '@/components/PlanPicker.vue';
import {
  WorkspacesDocument,
  UpdateWorkspaceDocument,
  DeleteWorkspaceDocument,
  InviteMemberDocument,
  RemoveMemberDocument,
  UpdateMemberRoleDocument,
  SubscriptionDocument,
  SubscriptionStatus,
  UsageSummaryDocument,
  ChangePlanDocument,
  BillingPortalUrlDocument,
  CreatePlanCheckoutDocument,
  WorkspaceDocument,
  WorkspaceRole,
  Plan,
} from '@/gql/graphql';

const router = useRouter();
const { refreshToken, setActiveWorkspace } = useAuth();

const { result, loading, refetch } = useQuery(WorkspaceDocument);
const workspace = computed(() => result.value?.workspace);
const members = computed(() => workspace.value?.members ?? []);
const isAdmin = computed(() => {
  const { user } = useAuth();
  if (!user.value || !workspace.value) return false;
  const membership = user.value.workspaces.find(w => w.workspace === workspace.value!.id);
  return membership?.role === 'admin';
});

// Settings sections — sync with ?tab= query param for deep linking.
const route = useRoute();
const validSections = ['general', 'members', 'billing', 'danger'];
const initialTab = validSections.includes(route.query.tab as string) ? (route.query.tab as string) : 'general';
const activeSection = ref(initialTab);
const sections = computed(() => {
  const s = [
    { id: 'general', label: 'General' },
    { id: 'members', label: 'Members' },
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

const { mutate: updateMutate, loading: updating } = useMutation(UpdateWorkspaceDocument);

async function handleUpdateName() {
  if (!editName.value.trim() || editName.value.trim() === workspace.value?.name) return;
  try {
    await updateMutate({ input: { name: editName.value.trim() } });
    toast.success('Workspace name updated');
    refetch();
  } catch (e: unknown) {
    errorToast('Failed to update workspace', { description: errorMessage(e) });
  }
}

// Invite member
const inviteEmail = ref('');
const inviteRole = ref<WorkspaceRole>(WorkspaceRole.User);
const { mutate: inviteMutate, loading: inviting } = useMutation(InviteMemberDocument);

async function handleInvite() {
  if (!inviteEmail.value.trim()) return;
  try {
    const res = await inviteMutate({
      input: { email: inviteEmail.value.trim(), role: inviteRole.value },
    });
    if (res?.errors?.length) {
      errorToast('Failed to invite member', {
        description: res.errors.map((e: { message: string }) => e.message).join(', '),
      });
      return;
    }
    toast.success(`Invited ${inviteEmail.value.trim()}`);
    inviteEmail.value = '';
    inviteRole.value = WorkspaceRole.User;
    await refreshToken();
    refetch();
  } catch (e: unknown) {
    errorToast('Failed to invite member', { description: errorMessage(e) });
  }
}

// Remove member
const { mutate: removeMutate } = useMutation(RemoveMemberDocument);

async function handleRemoveMember(userId: string) {
  try {
    await removeMutate({ userId });
    toast.success('Member removed');
    await refreshToken();
    refetch();
  } catch (e: unknown) {
    errorToast('Failed to remove member', { description: errorMessage(e) });
  }
}

// Update member role
const { mutate: updateRoleMutate } = useMutation(UpdateMemberRoleDocument);

async function handleUpdateRole(userId: string, role: WorkspaceRole) {
  try {
    await updateRoleMutate({ input: { userId, role } });
    toast.success('Member role updated');
    refetch();
  } catch (e: unknown) {
    errorToast('Failed to update role', { description: errorMessage(e) });
  }
}

// Billing
const { result: subResult, loading: subLoading, error: subError } = useQuery(SubscriptionDocument);
const subscription = computed(() => subResult.value?.subscription);

const { result: usageResult, loading: usageLoading } = useQuery(UsageSummaryDocument);
const usage = computed(() => usageResult.value?.usageSummary);

const billingAvailable = computed(() => !subError.value && subscription.value);

const { mutate: changePlanMutate, loading: changingPlan } = useMutation(ChangePlanDocument, {
  refetchQueries: () => [{ query: SubscriptionDocument }, { query: UsageSummaryDocument }],
});
const { mutate: portalMutate, loading: openingPortal } = useMutation(BillingPortalUrlDocument);
const { mutate: planCheckoutMutate, loading: startingPlanCheckout } = useMutation(CreatePlanCheckoutDocument);
const confirmPlan = ref<Plan | null>(null);
const trialSelectedPlan = ref<Plan>(Plan.Hobby);
const isTrial = computed(() => billingAvailable.value && subscription.value?.status === SubscriptionStatus.Trialing);

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
    toast.success(`Switched to ${confirmPlan.value === Plan.Pro ? 'Pro' : 'Hobby'} plan`);
  } catch (e: unknown) {
    errorToast('Failed to change plan', { description: errorMessage(e) });
  } finally {
    confirmPlan.value = null;
  }
}

async function handlePlanCheckout(plan: Plan) {
  try {
    const res = await planCheckoutMutate({ plan });
    const url = res?.data?.createPlanCheckout?.url;
    if (url) {
      window.location.href = url;
    }
  } catch (e: unknown) {
    errorToast('Failed to start checkout', { description: errorMessage(e) });
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
    errorToast('Failed to open billing portal', { description: errorMessage(e) });
  }
}

// Delete workspace
const { mutate: deleteMutate, loading: deleting } = useMutation(DeleteWorkspaceDocument, {
  refetchQueries: () => [{ query: WorkspacesDocument }],
});

async function handleDelete() {
  try {
    const res = await deleteMutate();
    if (res?.errors?.length) {
      errorToast('Failed to delete workspace', {
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
    errorToast('Failed to delete workspace', { description: errorMessage(e) });
  }
}
</script>

<template>
  <div class="flex h-[calc(100vh-52px-0.75rem)] flex-col">
    <!-- Loading -->
    <div v-if="loading" class="flex flex-1 items-center justify-center">
      <div class="space-y-4 text-center">
        <Skeleton class="mx-auto h-8 w-48" />
        <Skeleton class="mx-auto h-4 w-64" />
      </div>
    </div>

    <template v-else-if="workspace">
      <div class="flex flex-1 overflow-hidden p-3">
        <div class="mx-auto flex w-full max-w-4xl gap-6 overflow-hidden rounded-lg border bg-card/80 shadow-sm backdrop-blur-sm [background-image:var(--gradient-card)]">
          <!-- Sidebar -->
          <nav class="w-48 shrink-0 border-r p-4">
            <div class="mb-4">
              <button
                class="flex items-center gap-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground"
                @click="router.push({ name: 'projects' })"
              >
                <ArrowLeft :size="12" />
                Back
              </button>
            </div>
            <h2 class="mb-3 text-sm font-semibold text-foreground">Settings</h2>
            <ul class="space-y-1">
              <li v-for="section in sections" :key="section.id">
                <button
                  class="w-full rounded-md px-3 py-1.5 text-left text-sm transition-colors"
                  :class="activeSection === section.id
                    ? 'bg-accent text-accent-foreground font-medium'
                    : 'text-muted-foreground hover:text-foreground hover:bg-accent/50'"
                  @click="activeSection = section.id"
                >
                  {{ section.label }}
                </button>
              </li>
            </ul>
          </nav>

          <!-- Content -->
          <div class="flex-1 overflow-y-auto p-6">
          <!-- General -->
          <section v-if="activeSection === 'general'" class="space-y-6">
            <div>
              <h2 class="text-lg font-semibold text-foreground">General</h2>
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
              <h2 class="text-lg font-semibold text-foreground">Members</h2>
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
                        @update:model-value="handleUpdateRole(member.id, $event as WorkspaceRole)"
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
                      <Badge :variant="member.role === WorkspaceRole.Admin ? 'default' : 'secondary'">
                        {{ member.role === WorkspaceRole.Admin ? 'Admin' : 'Member' }}
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

          <!-- Billing -->
          <section v-if="activeSection === 'billing'" class="space-y-6">
            <div>
              <h2 class="text-lg font-semibold text-foreground">Billing</h2>
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
              <!-- Trial: upgrade prompt -->
              <div
                v-if="isTrial && isAdmin"
                class="rounded-lg border p-6 space-y-5"
              >
                <div class="flex items-start gap-4">
                  <div class="flex size-10 shrink-0 items-center justify-center rounded-full bg-primary/10">
                    <CreditCard :size="20" class="text-primary" />
                  </div>
                  <div>
                    <h3 class="text-sm font-medium text-foreground">
                      Choose a plan to continue
                    </h3>
                    <p class="mt-1 text-sm text-muted-foreground">
                      You're on a trial with &euro;5 in free credits<template v-if="subscription!.creditExpiry">, expiring {{ formatDate(subscription!.creditExpiry) }}</template>.
                      Pick a plan and add a payment method to keep your workspace running.
                    </p>
                  </div>
                </div>

                <PlanPicker
                  v-model="trialSelectedPlan"
                  :disabled="startingPlanCheckout"
                />

                <Button
                  :disabled="startingPlanCheckout"
                  @click="handlePlanCheckout(trialSelectedPlan)"
                >
                  {{ startingPlanCheckout ? 'Redirecting...' : `Continue with ${trialSelectedPlan === Plan.Pro ? 'Pro' : 'Hobby'}` }}
                </Button>
              </div>

              <!-- Non-trial: subscription summary -->
              <div v-if="!isTrial" class="rounded-lg border p-4 space-y-3">
                <div class="flex items-center justify-between">
                  <h3 class="text-sm font-medium">Subscription</h3>
                  <Badge :variant="subscription!.status === SubscriptionStatus.Active ? 'default' : 'destructive'">
                    {{ subscription!.status === SubscriptionStatus.Active ? 'Active' : subscription!.status === SubscriptionStatus.PastDue ? 'Past Due' : subscription!.status }}
                  </Badge>
                </div>
                <div class="flex items-center justify-between text-sm">
                  <span class="text-muted-foreground">Current plan</span>
                  <span class="font-medium">{{ subscription!.plan === Plan.Pro ? 'Pro' : 'Hobby' }}</span>
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

              <!-- Credits banner: no payment method yet (non-trial, e.g. workspace checkout) -->
              <div
                v-if="!isTrial && subscription!.creditExpiry && !subscription!.hasPaymentMethod"
                class="rounded-lg border border-primary/30 bg-primary/5 p-4"
              >
                <p class="text-sm font-medium text-foreground">
                  Your credits expire on {{ formatDate(subscription!.creditExpiry) }}. Add a payment method to continue using the platform.
                </p>
                <p class="mt-1 text-xs text-muted-foreground">
                  After credits expire, your workspace will be suspended until a payment method is added.
                </p>
                <Button
                  v-if="isAdmin"
                  variant="outline"
                  size="sm"
                  class="mt-3"
                  :disabled="openingPortal"
                  @click="handleOpenPortal"
                >
                  <ExternalLink :size="14" class="mr-1.5" />
                  {{ openingPortal ? 'Opening...' : 'Open Billing Portal' }}
                </Button>
              </div>

              <!-- Credits banner: payment method set -->
              <div
                v-else-if="!isTrial && subscription!.creditExpiry && subscription!.hasPaymentMethod"
                class="rounded-lg border border-green-500/30 bg-green-500/5 p-4"
              >
                <p class="text-sm font-medium text-foreground">
                  You're all set! Your payment method is on file and billing will continue automatically after your credits expire on {{ formatDate(subscription!.creditExpiry) }}.
                </p>
              </div>

              <!-- Plan switcher (admin only, only when already on a plan) -->
              <div v-if="isAdmin && !isTrial" class="space-y-3">
                <h3 class="text-sm font-medium">Plan</h3>
                <PlanPicker
                  :model-value="subscription!.plan ?? Plan.Hobby"
                  @update:model-value="confirmPlan = $event"
                />
              </div>

              <!-- Plan change confirmation -->
              <AlertDialog :open="!!confirmPlan">
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>
                      Switch to {{ confirmPlan === Plan.Pro ? 'Pro' : 'Hobby' }}?
                    </AlertDialogTitle>
                    <AlertDialogDescription>
                      Your plan will be changed to {{ confirmPlan === Plan.Pro ? 'Pro (&euro;25/mo)' : 'Hobby (&euro;5/mo)' }}.
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
              <h2 class="text-lg font-semibold text-destructive">Danger Zone</h2>
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
                        This will permanently delete the workspace and remove all members. All projects must be deleted first.
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
      </div>
    </template>
  </div>
</template>
