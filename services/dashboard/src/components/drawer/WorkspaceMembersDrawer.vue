<script setup lang="ts">
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import { computed, ref } from 'vue';
import { useMutation, useQuery } from '@vue/apollo-composable';
import { Shield, User as UserIcon, Trash2 } from '@lucide/vue';
import { graphql } from '@/gql';
import { Availability, WorkspaceRole } from '@/gql/graphql';
import { useCapabilities } from '@/composables/useCapabilities';
import { useGate } from '@/composables/useGate';
import GateTag from '@/components/GateTag.vue';
import { planTitle } from '@/lib/capabilities';

const InviteMemberDocument = graphql(`
  mutation InviteMember($input: InviteMemberInput!) {
    inviteMember(input: $input) {
      id
      email
      name
      role
    }
  }
`);

const RemoveMemberDocument = graphql(`
  mutation RemoveMember($userId: ID!) {
    removeMember(userId: $userId)
  }
`);

const MembersPlanDocument = graphql(`
  query MembersPlan {
    subscription {
      plan
    }
  }
`);

const UpdateMemberRoleDocument = graphql(`
  mutation UpdateMemberRole($input: UpdateMemberRoleInput!) {
    updateMemberRole(input: $input) {
      id
      email
      name
      role
    }
  }
`);
import ContextDrawer from '@/components/drawer/ContextDrawer.vue';
import Spinner from '@/components/LoadingSpinner.vue';
import { Button } from '@/components/ui/button';
import IconButton from '@/components/IconButton.vue';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { failureToast, successToast } from '@/components/ui/sonner';
import { useAuth } from '@/composables/useAuth';

const props = defineProps<{
  workspace: {
    name: string;
    members: { id: string; email: string; name?: string | null; role: WorkspaceRole }[];
  };
  isAdmin: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'changed'): void;
}>();

const { user, refreshToken } = useAuth();

const inviteEmail = ref('');
const inviteRole = ref<WorkspaceRole>(WorkspaceRole.User);
const { mutate: inviteMutate, loading: inviting } = useMutation(InviteMemberDocument);

const { capabilities, refetch: refetchCapabilities } = useCapabilities();
const { openFor } = useGate();
const personal = computed(() => capabilities.value?.members.availability === Availability.Unavailable);
const inviteGated = computed(() => capabilities.value?.members.availability === Availability.PlanRequired);

const { result: planResult } = useQuery(MembersPlanDocument, null, () => ({ enabled: inviteGated.value, errorPolicy: 'all' }));
const gatedLine = computed(() => {
  const members = capabilities.value?.members;
  if (!members?.plan) return '';
  const plan = planResult.value?.subscription?.plan;
  const current = plan ? planTitle(plan) : 'Trial';
  const upgrade = members.upgrades.find(u => u.plan === members.plan);
  const room = members.limit === 1 ? `${current} is for one person.` : `${current} includes ${members.limit} members.`;
  const more = upgrade?.limit == null ? 'adds as many members as you need' : `includes ${upgrade.limit} members`;
  return `${room} ${planTitle(members.plan)} ${more}.`;
});

async function handleInvite() {
  const gate = capabilities.value?.members;
  if (inviteGated.value && gate?.plan && openFor(gate, `Inviting members needs the ${planTitle(gate.plan)} plan`)) return;
  if (!inviteEmail.value.trim()) return;
  try {
    await inviteMutate({
      input: { email: inviteEmail.value.trim(), role: inviteRole.value },
    });
    successToast(`Invited ${inviteEmail.value.trim()}`);
    inviteEmail.value = '';
    inviteRole.value = WorkspaceRole.User;
    await refreshToken();
    emit('changed');
    refetchCapabilities();
  } catch (e: unknown) {
    failureToast('Failed to invite member', e);
  }
}

const { mutate: removeMutate, loading: removingMember } = useMutation(RemoveMemberDocument);
const memberToRemove = ref<{ id: string; label: string } | null>(null);

async function handleRemoveMember() {
  if (!memberToRemove.value) return;
  try {
    await removeMutate({ userId: memberToRemove.value.id });
    memberToRemove.value = null;
    await refreshToken();
    emit('changed');
    refetchCapabilities();
  } catch (e: unknown) {
    failureToast('Failed to remove member', e);
  }
}

const { mutate: updateRoleMutate } = useMutation(UpdateMemberRoleDocument);
const roleUpdatingFor = ref<string | null>(null);

async function handleUpdateRole(userId: string, role: WorkspaceRole) {
  roleUpdatingFor.value = userId;
  try {
    await updateRoleMutate({ input: { userId, role } });
    emit('changed');
  } catch (e: unknown) {
    failureToast('Failed to update role', e);
  } finally {
    roleUpdatingFor.value = null;
  }
}

function initial(member: { email: string; name?: string | null }) {
  return (member.name || member.email).charAt(0).toUpperCase();
}
</script>

<template>
  <ContextDrawer
    title="Members"
    @close="emit('close')"
  >
    <div class="space-y-5">
      <p v-if="personal" class="text-sm text-muted-foreground">
        This is your personal workspace, so only you can be in it. To work with others, create a new workspace from the workspace menu.
      </p>
      <p v-else-if="!isAdmin" class="text-sm text-muted-foreground">
        Only admins can invite people.
      </p>
      <div v-else-if="inviteGated" class="flex items-center justify-between gap-4">
        <p class="text-sm text-muted-foreground">{{ gatedLine }}</p>
        <Button variant="upgrade" class="shrink-0" @click="handleInvite">
          Invite
          <GateTag inverse :plan="capabilities?.members.plan" />
        </Button>
      </div>
      <form v-else class="space-y-2" @submit.prevent="handleInvite">
        <Label for="invite-email" class="text-sm text-muted-foreground">Invite by email</Label>
        <div
          class="flex h-9 items-center rounded-md border border-input bg-background pr-1 ring-offset-background focus-within:ring-2 focus-within:ring-ring focus-within:ring-offset-2"
          :class="inviting && 'opacity-50'"
        >
          <Input
            id="invite-email"
            v-model="inviteEmail"
            type="email"
            placeholder="user@example.com"
            class="h-full min-w-0 flex-1 border-0 bg-transparent focus-visible:ring-0 focus-visible:ring-offset-0 disabled:opacity-100"
            :disabled="inviting"
          />
          <Select v-model="inviteRole" :disabled="inviting">
            <SelectTrigger class="h-7 w-auto shrink-0 gap-1 rounded-sm border-0 bg-muted px-2 text-xs focus:ring-0 focus:ring-offset-0 disabled:opacity-100" aria-label="Role">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem :value="WorkspaceRole.User">Member</SelectItem>
              <SelectItem :value="WorkspaceRole.Admin">Admin</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div class="flex justify-end">
          <Button type="submit" :pending="inviting" :disabled="!inviteEmail.trim()">
            {{ inviting ? 'Inviting…' : 'Invite' }}
          </Button>
        </div>
      </form>

      <div class="-mx-4 divide-y border-y">
        <div
          v-for="member in props.workspace.members"
          :key="member.id"
          class="flex items-center gap-3 py-3 pr-3 pl-4"
        >
          <span class="flex size-8 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-semibold">
            {{ initial(member) }}
          </span>
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-medium text-foreground">
              {{ member.name || member.email }}
              <span v-if="member.email === user?.email" class="font-normal text-muted-foreground">(you)</span>
            </p>
            <p v-if="member.name" class="truncate text-xs text-muted-foreground">{{ member.email }}</p>
          </div>
          <template v-if="isAdmin">
            <Select
              :model-value="member.role"
              :disabled="roleUpdatingFor === member.id"
              @update:model-value="handleUpdateRole(member.id, $event as WorkspaceRole)"
            >
              <SelectTrigger class="h-8 w-26 shrink-0 text-xs" :aria-label="`Role of ${member.name || member.email}`">
                <Spinner v-if="roleUpdatingFor === member.id" :size="12" />
                <SelectValue v-else />
              </SelectTrigger>
              <SelectContent>
                <SelectItem :value="WorkspaceRole.User">
                  <div class="flex items-center gap-1.5">
                    <UserIcon :size="12" />
                    Member
                  </div>
                </SelectItem>
                <SelectItem :value="WorkspaceRole.Admin">
                  <div class="flex items-center gap-1.5">
                    <Shield :size="12" />
                    Admin
                  </div>
                </SelectItem>
              </SelectContent>
            </Select>
            <IconButton
              :label="`Remove ${member.name || member.email}`"
              class="shrink-0 text-muted-foreground hover:text-destructive"
              :disabled="removingMember"
              @click="memberToRemove = { id: member.id, label: member.name || member.email }"
            >
              <Trash2 />
            </IconButton>
          </template>
          <span
            v-else
            class="shrink-0 rounded-full px-2.5 py-0.5 text-xs font-medium"
            :class="member.role === WorkspaceRole.Admin ? 'bg-primary/15 text-foreground' : 'bg-muted text-muted-foreground'"
          >
            {{ member.role === WorkspaceRole.Admin ? 'Admin' : 'Member' }}
          </span>
        </div>
      </div>
    </div>

    <ConfirmDialog
      :open="!!memberToRemove"
      :title="`Remove member &quot;${memberToRemove?.label}&quot;?`"
      description="They lose access to this workspace right away. You can invite them again later."
      confirm-label="Remove member"
      pending-label="Removing…"
      :pending="removingMember"
      destructive
      @update:open="(value) => { if (!value) memberToRemove = null }"
      @confirm="handleRemoveMember"
    />
  </ContextDrawer>
</template>
