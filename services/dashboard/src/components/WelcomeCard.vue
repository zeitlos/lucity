<script setup lang="ts">
import { ref, computed } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Github, ArrowRight, FolderPlus, Check, CreditCard } from 'lucide-vue-next';
import { ChangePlanMutation, BillingPortalUrlMutation, SubscriptionQuery } from '@/graphql/billing';
import { useAuth } from '@/composables/useAuth';
import { toast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

defineEmits<{
  (e: 'dismiss'): void;
  (e: 'create-project'): void;
}>();

const { user } = useAuth();
const selectedPlan = ref<'HOBBY' | 'PRO'>('HOBBY');

const { mutate: changePlanMutate, loading: changingPlan } = useMutation(ChangePlanMutation, {
  refetchQueries: () => [{ query: SubscriptionQuery }],
});

async function selectPlan(plan: 'HOBBY' | 'PRO') {
  selectedPlan.value = plan;
  if (plan === 'PRO') {
    try {
      await changePlanMutate({ plan: 'PRO' });
    } catch (e: unknown) {
      toast.error('Failed to switch plan', { description: errorMessage(e) });
      selectedPlan.value = 'HOBBY';
    }
  }
}

const { result: subResult } = useQuery(SubscriptionQuery, null, { fetchPolicy: 'cache-and-network' });
const hasPaymentMethod = computed(() => subResult.value?.subscription?.hasPaymentMethod ?? false);

const { mutate: portalMutate, loading: openingPortal } = useMutation(BillingPortalUrlMutation);

async function openBillingPortal() {
  try {
    const result = await portalMutate();
    const url = result?.data?.billingPortalUrl?.url;
    if (url) {
      window.open(url, '_blank');
    }
  } catch (e: unknown) {
    toast.error('Failed to open billing portal', { description: errorMessage(e) });
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
            You have free credits to get started — no card needed.
          </p>
        </div>

        <!-- Plan selection -->
        <div class="space-y-3">
          <h2 class="text-sm font-medium text-foreground">Choose your plan</h2>
          <div class="grid grid-cols-2 gap-3 max-w-lg">
            <button
              class="rounded-lg border p-4 text-left transition-colors"
              :class="selectedPlan === 'HOBBY'
                ? 'border-primary bg-primary/5'
                : 'hover:border-muted-foreground/50'"
              :disabled="changingPlan"
              @click="selectPlan('HOBBY')"
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
                €5<span class="text-sm font-normal text-muted-foreground">/mo</span>
              </p>
              <p class="mt-1 text-xs text-muted-foreground">
                €5 credit/mo. Great for side projects.
              </p>
            </button>
            <button
              class="rounded-lg border p-4 text-left transition-colors"
              :class="selectedPlan === 'PRO'
                ? 'border-primary bg-primary/5'
                : 'hover:border-muted-foreground/50'"
              :disabled="changingPlan"
              @click="selectPlan('PRO')"
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
                €25<span class="text-sm font-normal text-muted-foreground">/mo</span>
              </p>
              <p class="mt-1 text-xs text-muted-foreground">
                €25 credit/mo. For teams &amp; production.
              </p>
            </button>
          </div>
          <p class="text-xs text-muted-foreground">
            You can change your plan anytime.
          </p>
        </div>

        <!-- Payment method CTA -->
        <div v-if="!hasPaymentMethod" class="max-w-lg">
          <button
            class="flex w-full items-center justify-between rounded-lg border border-dashed p-4 transition-colors hover:bg-accent/50"
            :disabled="openingPortal"
            @click="openBillingPortal"
          >
            <div class="flex items-center gap-3">
              <CreditCard :size="18" class="text-muted-foreground" />
              <div class="text-left">
                <span class="text-sm font-medium">Add a payment method</span>
                <p class="text-xs text-muted-foreground">Avoids interruption when your trial ends. Completely optional.</p>
              </div>
            </div>
            <ArrowRight :size="16" class="text-muted-foreground" />
          </button>
        </div>
        <div v-else class="max-w-lg">
          <div class="flex w-full items-center rounded-lg border border-green-500/30 bg-green-500/5 p-4">
            <div class="flex items-center gap-3">
              <Check :size="18" class="text-green-500" />
              <div class="text-left">
                <span class="text-sm font-medium">Payment method added</span>
                <p class="text-xs text-muted-foreground">Your plan will continue automatically after the trial.</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Action cards -->
        <div class="space-y-3">
          <h2 class="text-sm font-medium text-foreground">What's next?</h2>
          <div class="max-w-lg space-y-2">
            <a
              href="/auth/github/connect"
              class="flex items-center justify-between rounded-lg border p-4 transition-colors hover:bg-accent/50"
            >
              <div class="flex items-center gap-3">
                <Github :size="18" class="text-muted-foreground" />
                <span class="text-sm font-medium">Import from GitHub</span>
              </div>
              <ArrowRight :size="16" class="text-muted-foreground" />
            </a>
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
