<script setup lang="ts">
import { ref } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { Shield, User as UserIcon, Trash2 } from '@lucide/vue';
import { graphql } from '@/gql';
import { WorkspaceRole } from '@/gql/graphql';

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
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';
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

async function handleInvite() {
  if (!inviteEmail.value.trim()) return;
  try {
    const res = await inviteMutate({
      input: { email: inviteEmail.value.trim(), role: inviteRole.value },
    });
    if (res?.errors?.length) {
      errorToast('Failed to invite member', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }
    toast.success(`Invited ${inviteEmail.value.trim()}`);
    inviteEmail.value = '';
    inviteRole.value = WorkspaceRole.User;
    await refreshToken();
    emit('changed');
  } catch (e: unknown) {
    errorToast('Failed to invite member', { description: errorMessage(e) });
  }
}

const { mutate: removeMutate, loading: removingMember } = useMutation(RemoveMemberDocument);
const memberToRemove = ref<{ id: string; label: string } | null>(null);

async function handleRemoveMember() {
  if (!memberToRemove.value) return;
  try {
    await removeMutate({ userId: memberToRemove.value.id });
    toast.success('Member removed');
    memberToRemove.value = null;
    await refreshToken();
    emit('changed');
  } catch (e: unknown) {
    errorToast('Failed to remove member', { description: errorMessage(e) });
  }
}

const { mutate: updateRoleMutate } = useMutation(UpdateMemberRoleDocument);
const roleUpdatingFor = ref<string | null>(null);

async function handleUpdateRole(userId: string, role: WorkspaceRole) {
  roleUpdatingFor.value = userId;
  try {
    await updateRoleMutate({ input: { userId, role } });
    toast.success('Member role updated');
    emit('changed');
  } catch (e: unknown) {
    errorToast('Failed to update role', { description: errorMessage(e) });
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
    :description="isAdmin
      ? `Everyone with access to ${props.workspace.name}. Admins can invite people and change roles.`
      : `Everyone with access to ${props.workspace.name}.`"
    @close="emit('close')"
  >
    <div class="space-y-5">
      <form v-if="isAdmin" class="space-y-2" @submit.prevent="handleInvite">
        <Label for="invite-email" class="text-[13px] text-muted-foreground">Invite by email</Label>
        <div class="flex items-center gap-2">
          <Input
            id="invite-email"
            v-model="inviteEmail"
            type="email"
            placeholder="user@example.com"
            class="h-9 min-w-0 flex-1 text-[13px]"
            :disabled="inviting"
          />
          <Select v-model="inviteRole">
            <SelectTrigger class="h-9 w-28 shrink-0 text-[13px]" aria-label="Role">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem :value="WorkspaceRole.User">Member</SelectItem>
              <SelectItem :value="WorkspaceRole.Admin">Admin</SelectItem>
            </SelectContent>
          </Select>
          <Button type="submit" class="shrink-0" :disabled="!inviteEmail.trim() || inviting">
            {{ inviting ? 'Inviting...' : 'Invite' }}
          </Button>
        </div>
      </form>

      <div class="divide-y overflow-hidden rounded-lg border">
        <div
          v-for="member in props.workspace.members"
          :key="member.id"
          class="flex items-center gap-3 py-3 pl-3.5 pr-2"
        >
          <span class="flex size-[30px] shrink-0 items-center justify-center rounded-full bg-muted text-xs font-semibold">
            {{ initial(member) }}
          </span>
          <div class="min-w-0 flex-1">
            <p class="truncate text-[13px] font-medium text-foreground">
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
              <SelectTrigger class="h-8 w-[104px] shrink-0 text-xs" :aria-label="`Role of ${member.name || member.email}`">
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
            class="shrink-0 rounded-full px-2.5 py-0.5 text-[11px] font-medium"
            :class="member.role === WorkspaceRole.Admin ? 'bg-primary/15 text-foreground' : 'bg-muted text-muted-foreground'"
          >
            {{ member.role === WorkspaceRole.Admin ? 'Admin' : 'Member' }}
          </span>
        </div>
      </div>
    </div>

    <AlertDialog :open="!!memberToRemove">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Remove member?</AlertDialogTitle>
          <AlertDialogDescription>
            {{ memberToRemove?.label }} will lose access to this workspace.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel :disabled="removingMember" @click="memberToRemove = null">Cancel</AlertDialogCancel>
          <Button variant="destructive" :disabled="removingMember" @click="handleRemoveMember">
            {{ removingMember ? 'Removing...' : 'Remove' }}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </ContextDrawer>
</template>
