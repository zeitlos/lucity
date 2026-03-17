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
import { toast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';
import { isValidSlug } from '@/lib/slug';
import { Check } from 'lucide-vue-next';
import NameSlugField from '@/components/NameSlugField.vue';

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
      toast.error('Failed to create checkout session', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    const url = res?.data?.createWorkspaceCheckout?.url;
    if (url) {
      window.location.href = url;
    }
  } catch (e: unknown) {
    toast.error('Failed to create checkout session', { description: errorMessage(e) });
  }
}
</script>

<template>
  <Dialog
    :open="open"
    @update:open="emit('update:open', $event)"
  >
    <DialogContent class="sm:max-w-md">
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
          <div class="grid grid-cols-2 gap-3">
            <button
              type="button"
              class="rounded-lg border p-4 text-left transition-colors"
              :class="selectedPlan === 'HOBBY'
                ? 'border-primary bg-primary/5'
                : 'hover:border-muted-foreground/50'"
              :disabled="loading"
              @click="selectedPlan = 'HOBBY'"
            >
              <div class="flex items-center justify-between">
                <p class="text-sm font-medium">Hobby</p>
                <Check
                  v-if="selectedPlan === 'HOBBY'"
                  :size="14"
                  class="text-primary"
                />
              </div>
              <p class="text-lg font-semibold">
                &euro;5<span class="text-sm font-normal text-muted-foreground">/mo</span>
              </p>
              <p class="mt-1 text-xs text-muted-foreground">
                &euro;5 credit/mo. Great for side projects.
              </p>
            </button>
            <button
              type="button"
              class="rounded-lg border p-4 text-left transition-colors"
              :class="selectedPlan === 'PRO'
                ? 'border-primary bg-primary/5'
                : 'hover:border-muted-foreground/50'"
              :disabled="loading"
              @click="selectedPlan = 'PRO'"
            >
              <div class="flex items-center justify-between">
                <p class="text-sm font-medium">Pro</p>
                <Check
                  v-if="selectedPlan === 'PRO'"
                  :size="14"
                  class="text-primary"
                />
              </div>
              <p class="text-lg font-semibold">
                &euro;25<span class="text-sm font-normal text-muted-foreground">/mo</span>
              </p>
              <p class="mt-1 text-xs text-muted-foreground">
                &euro;25 credit/mo. For teams &amp; production.
              </p>
            </button>
          </div>
          <p class="text-xs text-muted-foreground">
            You can change your plan anytime.
          </p>
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
