<script setup lang="ts">
import { subscriptionStatus } from '@/lib/status';
import { formatDate, formatMoney } from '@/lib/format';
import QueryError from '@/components/QueryError.vue';
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import { ref, computed } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { CreditCard, ExternalLink } from '@lucide/vue';
import { graphql } from '@/gql';
import { Plan, SubscriptionStatus } from '@/gql/graphql';

const SubscriptionDocument = graphql(`
  query Subscription {
    subscription {
      plan
      status
      currentPeriodEnd
      creditAmountCents
      creditExpiry
      hasPaymentMethod
    }
  }
`);

const UsageSummaryDocument = graphql(`
  query UsageSummary {
    usageSummary {
      resourceCostCents
      creditsCents
      estimatedTotalCents
    }
  }
`);

const SelectPlanDocument = graphql(`
  mutation SelectPlan($plan: Plan!) {
    selectPlan(plan: $plan) {
      checkoutUrl
      subscription {
        plan
        status
        currentPeriodEnd
        creditAmountCents
        creditExpiry
        hasPaymentMethod
      }
    }
  }
`);

const BillingPortalUrlDocument = graphql(`
  mutation BillingPortalUrl {
    billingPortalUrl {
      url
    }
  }
`);
import ContextDrawer from '@/components/drawer/ContextDrawer.vue';
import PlanPicker from '@/components/PlanPicker.vue';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Status } from '@/components/ui/status';
import { failureToast } from '@/components/ui/sonner';

defineProps<{
  isAdmin: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const { result: subResult, loading: subLoading, error: subError, refetch: refetchSubscription } = useQuery(SubscriptionDocument);
const subscription = computed(() => subResult.value?.subscription);

const { result: usageResult, loading: usageLoading } = useQuery(UsageSummaryDocument);
const usage = computed(() => usageResult.value?.usageSummary);

const billingNotConfigured = computed(() => !!subError.value?.message.includes('not configured'));
const billingAvailable = computed(() => !subError.value);

const { mutate: selectPlanMutate, loading: selectingPlan } = useMutation(SelectPlanDocument, {
  refetchQueries: () => [{ query: SubscriptionDocument }, { query: UsageSummaryDocument }],
});
const { mutate: portalMutate, loading: openingPortal } = useMutation(BillingPortalUrlDocument);
const confirmPlan = ref<Plan | null>(null);
const trialSelectedPlan = ref<Plan>(Plan.Hobby);
const subscriptionEnded = computed(() => subscription.value?.status === SubscriptionStatus.Canceled);
const hasPlan = computed(() => billingAvailable.value && !!subscription.value?.plan && !subscriptionEnded.value);
const isTrialing = computed(() => billingAvailable.value && subscription.value?.status === SubscriptionStatus.Trialing);

function formatCents(cents: number): string {
  return formatMoney(cents / 100, 'EUR');
}

async function selectPlan(plan: Plan): Promise<boolean> {
  try {
    const res = await selectPlanMutate({ plan });
    const url = res?.data?.selectPlan?.checkoutUrl;
    if (url) {
      window.location.href = url;
      return false;
    }
    return true;
  } catch (e: unknown) {
    failureToast('Failed to select the plan', e);
    return false;
  }
}

async function handleChangePlan() {
  if (!confirmPlan.value) return;
  try {
    await selectPlan(confirmPlan.value);
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
    failureToast('Failed to open billing portal', e);
  }
}
</script>

<template>
  <ContextDrawer
    title="Billing"
    @close="emit('close')"
  >
    <div v-if="subLoading" class="space-y-3">
      <Skeleton class="h-28 w-full" />
      <Skeleton class="h-24 w-full" />
    </div>

    <QueryError
      v-else-if="subError && !billingNotConfigured"
      title="Couldn't load billing"
      :error="subError"
      @retry="refetchSubscription()"
    />

    <div v-else-if="!billingAvailable" class="flex items-center gap-3 rounded-lg border p-3">
      <CreditCard :size="18" class="shrink-0 text-muted-foreground" />
      <div>
        <p class="text-sm font-medium">Billing is not configured</p>
        <p class="text-xs text-muted-foreground">
          Billing is not available for this workspace. Contact your administrator to set up billing.
        </p>
      </div>
    </div>

    <div v-else class="space-y-6">
      <section v-if="!hasPlan && isAdmin" class="space-y-4 rounded-lg border p-4">
        <div>
          <h3 class="text-sm font-medium text-foreground">Choose a plan to continue</h3>
          <p v-if="!subscription" class="mt-1 text-xs text-muted-foreground">
            Pick a plan and add a payment method to start using your workspace.
          </p>
          <p v-else-if="subscriptionEnded" class="mt-1 text-xs text-muted-foreground">
            Your subscription has ended. Pick a plan to reactivate your workspace.
          </p>
          <p v-else-if="isTrialing" class="mt-1 text-xs text-muted-foreground">
            You're on a trial with &euro;5 in free credits<template v-if="subscription.creditExpiry">, expiring {{ formatDate(subscription.creditExpiry) }}</template>.
            Pick a plan and add a payment method to keep your workspace running.
          </p>
          <p v-else class="mt-1 text-xs text-muted-foreground">
            Your trial has ended. Pick a plan and add a payment method to reactivate your workspace.
          </p>
        </div>
        <PlanPicker v-model="trialSelectedPlan" :disabled="selectingPlan" />
        <Button :pending="selectingPlan" @click="selectPlan(trialSelectedPlan)">
          {{ selectingPlan ? 'Continuing…' : `Continue with ${trialSelectedPlan === Plan.Pro ? 'Pro' : 'Hobby'}` }}
        </Button>
      </section>

      <section v-if="hasPlan" class="space-y-2.5">
        <h3 class="text-sm font-medium text-muted-foreground">Subscription</h3>
        <div class="space-y-2.5 rounded-lg border p-3">
          <div class="flex items-center justify-between text-sm">
            <span class="text-muted-foreground">Status</span>
            <Status :tone="subscriptionStatus(subscription!.status).tone" variant="text">
              {{ subscriptionStatus(subscription!.status).label }}
            </Status>
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
            <span class="text-muted-foreground">Credit balance</span>
            <span class="tabular-nums">{{ formatCents(subscription!.creditAmountCents) }}</span>
          </div>
        </div>
      </section>

      <div
        v-if="hasPlan && subscription!.creditExpiry && !subscription!.hasPaymentMethod"
        class="rounded-lg border border-primary/30 bg-primary/5 p-3"
      >
        <p class="text-sm font-medium text-foreground">
          Your credits expire on {{ formatDate(subscription!.creditExpiry) }}. Add a payment method to continue using the platform.
        </p>
        <p class="mt-1 text-xs text-muted-foreground">
          After credits expire, your workspace will be suspended until a payment method is added.
        </p>
      </div>
      <div
        v-else-if="hasPlan && subscription!.creditExpiry && subscription!.hasPaymentMethod"
        class="rounded-lg border border-status-ok/30 bg-status-ok/5 p-3"
      >
        <p class="text-sm font-medium text-foreground">
          You're all set! Your payment method is on file and billing will continue automatically after your credits expire on {{ formatDate(subscription!.creditExpiry) }}.
        </p>
      </div>

      <section class="space-y-2.5">
        <h3 class="text-sm font-medium text-muted-foreground">This period</h3>
        <div class="space-y-2.5 rounded-lg border p-3">
          <template v-if="usageLoading">
            <Skeleton class="h-4 w-full" />
            <Skeleton class="h-4 w-full" />
            <Skeleton class="h-4 w-full" />
          </template>
          <template v-else-if="usage">
            <div class="flex items-center justify-between text-sm">
              <span class="text-muted-foreground">Resources</span>
              <span class="tabular-nums">{{ formatCents(usage.resourceCostCents) }}</span>
            </div>
            <div class="flex items-center justify-between text-sm">
              <span class="text-muted-foreground">Credits</span>
              <span class="text-status-ok tabular-nums">{{ formatCents(-usage.creditsCents) }}</span>
            </div>
            <div class="h-px bg-border/60" />
            <div class="flex items-center justify-between text-sm font-semibold">
              <span>Estimated total</span>
              <span class="tabular-nums">{{ formatCents(usage.estimatedTotalCents) }}</span>
            </div>
          </template>
        </div>
      </section>

      <section v-if="isAdmin && hasPlan" class="space-y-2.5">
        <h3 class="text-sm font-medium text-muted-foreground">Plan</h3>
        <PlanPicker
          :model-value="subscription!.plan ?? Plan.Hobby"
          @update:model-value="confirmPlan = $event"
        />
      </section>
    </div>

    <template v-if="isAdmin && billingAvailable && !subLoading" #footer>
      <div class="flex items-center justify-between gap-4">
        <p class="text-xs leading-4.5 text-muted-foreground">
          Invoices and payment methods live in the billing portal.
        </p>
        <Button variant="outline" class="shrink-0" :pending="openingPortal" @click="handleOpenPortal">
          <ExternalLink />
          {{ openingPortal ? 'Opening…' : 'Open portal' }}
        </Button>
      </div>
    </template>

    <ConfirmDialog
      :open="!!confirmPlan"
      :title="`Switch to ${confirmPlan === Plan.Pro ? 'Pro' : 'Hobby'}?`"
      :description="`Your plan changes to ${confirmPlan === Plan.Pro ? 'Pro (€25/mo)' : 'Hobby (€5/mo)'} right away, with prorated billing.`"
      confirm-label="Switch plan"
      :icon="CreditCard"
      pending-label="Switching…"
      :pending="selectingPlan"
      @update:open="(value) => { if (!value) confirmPlan = null }"
      @confirm="handleChangePlan"
    />
  </ContextDrawer>
</template>
