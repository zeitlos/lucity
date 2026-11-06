<script setup lang="ts">
import { mobileSheetClass } from '@/lib/dialog';
import DialogIcon from '@/components/DialogIcon.vue';
import { Building2 } from '@lucide/vue';
import { ref, computed } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { graphql } from '@/gql';
import { Plan } from '@/gql/graphql';

const CreateWorkspaceCheckoutDocument = graphql(`
  mutation CreateWorkspaceCheckout($input: CreateWorkspaceCheckoutInput!) {
    createWorkspaceCheckout(input: $input) {
      url
    }
  }
`);
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
import { failureToast } from '@/components/ui/sonner';
import NameIdField from '@/components/NameIdField.vue';
import { RESERVED_WORKSPACE_IDS } from '@/lib/routes';
import PlanPicker from '@/components/PlanPicker.vue';

defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void;
}>();

const name = ref('');
const id = ref('');
const selectedPlan = ref<Plan>(Plan.Hobby);

const nameIdRef = ref<InstanceType<typeof NameIdField> | null>(null);
const isValid = computed(() => nameIdRef.value?.isValid ?? false);

const { mutate, loading } = useMutation(CreateWorkspaceCheckoutDocument);

async function handleCheckout() {
  if (!isValid.value) return;

  try {
    const res = await mutate({
      input: { id: id.value, name: name.value.trim(), plan: selectedPlan.value },
    });

    const url = res?.data?.createWorkspaceCheckout?.url;
    if (url) {
      window.location.href = url;
    }
  } catch (e: unknown) {
    failureToast('Failed to create checkout session', e);
  }
}
</script>

<template>
  <Dialog
    :open="open"
    @update:open="emit('update:open', $event)"
  >
    <DialogContent :class="[mobileSheetClass, 'sm:max-w-xl']">
      <DialogHeader>
        <DialogIcon :icon="Building2" />
        <DialogTitle>New workspace</DialogTitle>
        <DialogDescription>
          Create a workspace to collaborate with your team.
        </DialogDescription>
      </DialogHeader>

      <form
        class="space-y-4"
        @submit.prevent="handleCheckout"
      >
        <NameIdField
          ref="nameIdRef"
          v-model:name="name"
          v-model:id="id"
          fallback="team"
          :taken="RESERVED_WORKSPACE_IDS"
          reserved
          :disabled="loading"
          name-placeholder="e.g. My Team"
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
            :pending="loading"
            :disabled="!isValid"
          >
            {{ loading ? 'Redirecting…' : 'Subscribe' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
