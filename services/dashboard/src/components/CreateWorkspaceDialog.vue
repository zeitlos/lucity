<script setup lang="ts">
import { ref, computed } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { CreateWorkspaceCheckoutMutation } from '@/graphql/workspaces';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';
import { isValidSlug } from '@/lib/slug';
import NameSlugField from '@/components/NameSlugField.vue';
import PlanPicker from '@/components/PlanPicker.vue';

defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void;
}>();

const name = ref('');
const id = ref('');
const selectedPlan = ref<'HOBBY' | 'PRO'>('HOBBY');

const isValid = computed(() => name.value.trim().length > 0 && isValidSlug(id.value));

const { mutate, loading } = useMutation(CreateWorkspaceCheckoutMutation);

async function handleCheckout() {
  if (!isValid.value) return;

  try {
    const res = await mutate({
      input: { id: id.value, name: name.value.trim(), plan: selectedPlan.value },
    });

    if (res?.errors?.length) {
      errorToast('Failed to create checkout session', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    const url = res?.data?.createWorkspaceCheckout?.url;
    if (url) {
      window.location.href = url;
    }
  } catch (e: unknown) {
    errorToast('Failed to create checkout session', { description: errorMessage(e) });
  }
}
</script>

<template>
  <Dialog
    :open="open"
    @update:open="emit('update:open', $event)"
  >
    <DialogContent class="sm:max-w-xl">
      <DialogHeader>
        <DialogTitle>New Workspace</DialogTitle>
        <DialogDescription>
          Create a workspace to collaborate with your team.
        </DialogDescription>
      </DialogHeader>

      <form
        class="space-y-4"
        @submit.prevent="handleCheckout"
      >
        <NameSlugField
          v-model:name="name"
          v-model:slug="id"
          :disabled="loading"
          name-placeholder="e.g. My Team"
          slug-description="Used in URLs and API calls. Auto-generated from the name."
        />

        <div class="space-y-2">
          <Label>Plan</Label>
          <PlanPicker
            v-model="selectedPlan"
            :disabled="loading"
          />
        </div>

        <DialogFooter>
          <Button
            type="submit"
            :disabled="!isValid || loading"
          >
            {{ loading ? 'Redirecting...' : 'Subscribe' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
