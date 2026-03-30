<script setup lang="ts">
import { ref } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { Github, ArrowRight, FolderPlus } from 'lucide-vue-next';
import { CreatePlanCheckoutDocument, Plan } from '@/gql/graphql';
import { useAuth } from '@/composables/useAuth';
import { errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';
import PlanPicker from '@/components/PlanPicker.vue';

defineEmits<{
  (e: 'dismiss'): void;
  (e: 'create-project'): void;
  (e: 'import-github'): void;
}>();

const { user } = useAuth();
const selectedPlan = ref<Plan>(Plan.Hobby);

const { mutate: planCheckoutMutate, loading: checkingOut } = useMutation(CreatePlanCheckoutDocument);

async function continueWithPlan() {
  try {
    const result = await planCheckoutMutate({ plan: selectedPlan.value });
    const url = result?.data?.createPlanCheckout?.url;
    if (url) {
      window.location.href = url;
    }
  } catch (e: unknown) {
    errorToast('Failed to start checkout', { description: errorMessage(e) });
  }
}

const firstName = ref(
  user.value?.name?.split(' ')[0] || '',
);
</script>

<template>
  <div class="overflow-hidden rounded-lg border border-border bg-card">
    <div class="relative px-8 py-10">
      <!-- Pattern texture -->
      <div class="pattern-crosshatch pointer-events-none absolute inset-0 opacity-[0.02]" />

      <div class="relative space-y-8">
        <!-- Header -->
        <div>
          <h1 class="font-serif text-3xl text-foreground">
            Welcome{{ firstName ? `, ${firstName}` : '' }}
          </h1>
          <p class="mt-2 text-sm text-muted-foreground">
            You have free credits to get started, no card needed.
          </p>
        </div>

        <!-- Plan selection -->
        <div class="space-y-3">
          <h2 class="text-sm font-medium text-foreground">Choose your plan</h2>
          <div class="max-w-lg">
            <PlanPicker v-model="selectedPlan" />
          </div>
        </div>

        <!-- Continue CTA -->
        <div class="max-w-lg">
          <button
            class="flex w-full items-center justify-center gap-2 rounded-lg bg-primary px-4 py-3 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
            :disabled="checkingOut"
            @click="continueWithPlan"
          >
            Continue with {{ selectedPlan === Plan.Pro ? 'Pro' : 'Hobby' }}
            <ArrowRight :size="16" />
          </button>
        </div>

        <!-- Action cards -->
        <div class="space-y-3">
          <h2 class="text-sm font-medium text-foreground">What's next?</h2>
          <div class="max-w-lg space-y-2">
            <button
              class="flex w-full items-center justify-between rounded-lg border p-4 transition-colors hover:bg-accent/50"
              @click="$emit('import-github')"
            >
              <div class="flex items-center gap-3">
                <Github :size="18" class="text-muted-foreground" />
                <span class="text-sm font-medium">Import from GitHub</span>
              </div>
              <ArrowRight :size="16" class="text-muted-foreground" />
            </button>
            <button
              class="flex w-full items-center justify-between rounded-lg border p-4 transition-colors hover:bg-accent/50"
              @click="$emit('create-project')"
            >
              <div class="flex items-center gap-3">
                <FolderPlus :size="18" class="text-muted-foreground" />
                <span class="text-sm font-medium">Start with an empty project</span>
              </div>
              <ArrowRight :size="16" class="text-muted-foreground" />
            </button>
          </div>
        </div>

        <!-- Skip -->
        <button
          class="text-sm text-muted-foreground transition-colors hover:text-foreground"
          @click="$emit('dismiss')"
        >
          Skip for now
        </button>
      </div>
    </div>
  </div>
</template>
