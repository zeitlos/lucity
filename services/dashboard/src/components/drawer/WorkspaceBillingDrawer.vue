<script setup lang="ts">
import { subscriptionStatus } from '@/lib/status';
import { formatDate, formatMoney } from '@/lib/format';
import QueryError from '@/components/QueryError.vue';
import { ref, computed } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { ArrowRight, CreditCard, ExternalLink } from '@lucide/vue';
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

const BillingPlansDocument = graphql(`
  query BillingPlans {
    plans {
      plan
      priceCents
      currency
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
import { useCapabilities } from '@/composables/useCapabilities';
import { useGate } from '@/composables/useGate';

defineProps<{
  isAdmin: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const { cardRequired } = useCapabilities();
const { open: openGate } = useGate();

const { result: subResult, loading: subLoading, error: subError, refetch: refetchSubscription } = useQuery(SubscriptionDocument);
const subscription = computed(() => subResult.value?.subscription);

const { result: usageResult, loading: usageLoading, error: usageError, refetch: refetchUsage } = useQuery(UsageSummaryDocument);
const usage = computed(() => usageResult.value?.usageSummary);

const { result: plansResult } = useQuery(BillingPlansDocument);

function planName(plan: Plan | null | undefined): string {
  return plan === Plan.Pro ? 'Pro' : 'Hobby';
}

function planPrice(plan: Plan | null | undefined): string {
  const offer = plansResult.value?.plans.find(candidate => candidate.plan === plan);
  return offer ? `${formatMoney(offer.priceCents / 100, offer.currency)} per month` : '';
}


const billingNotConfigured = computed(() => !!subError.value?.message.includes('not configured'));
const billingAvailable = computed(() => !subError.value);

const { mutate: selectPlanMutate, loading: selectingPlan } = useMutation(SelectPlanDocument, {
  refetchQueries: () => [{ query: SubscriptionDocument }, { query: UsageSummaryDocument }],
});
const { mutate: portalMutate, loading: openingPortal } = useMutation(BillingPortalUrlDocument);
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

function changePlan() {
  const plan = subscription.value?.plan === Plan.Pro ? Plan.Hobby : Plan.Pro;
  openGate({ kind: 'plan', plan, title: 'Change plan' });
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
    <div v-if="subLoading" class="space-y-5">
      <div class="flex flex-col gap-0.5">
        <span class="text-sm text-muted-foreground">Plan</span>
        <Skeleton class="my-1.5 h-5 w-32" />
        <Skeleton class="my-0.5 h-3 w-40" />
      </div>
      <div class="-mx-4 border-t" />
      <div class="flex flex-col gap-0.5">
        <span class="text-sm text-muted-foreground">Total so far</span>
        <Skeleton class="my-1.5 h-5 w-20" />
      </div>
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

    <div v-else class="space-y-5">
      <section v-if="cardRequired && isAdmin" class="space-y-2.5">
        <h3 class="text-sm font-medium text-muted-foreground">Payment method</h3>
        <div class="flex items-center gap-3 rounded-lg border p-3">
          <CreditCard :size="18" class="shrink-0 text-muted-foreground" />
          <div class="min-w-0 flex-1">
            <p class="text-sm font-medium">No card on file</p>
            <p class="text-xs text-status-warn">Needed before you can create anything</p>
          </div>
          <Button variant="outline" class="shrink-0" @click="openGate({ kind: 'card' })">Add card</Button>
        </div>
      </section>
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

      <section v-if="hasPlan" aria-label="Plan">
        <div class="flex items-start justify-between gap-4">
          <div class="flex flex-col gap-0.5">
            <span class="text-sm text-muted-foreground">Plan</span>
            <span class="text-xl font-semibold">
              {{ planName(subscription!.plan) }}
              <span class="text-sm font-normal text-muted-foreground">{{ planPrice(subscription!.plan) }}</span>
            </span>
            <span class="text-xs text-muted-foreground">
              <Status :tone="subscriptionStatus(subscription!.status).tone" variant="text">{{ subscriptionStatus(subscription!.status).label }}</Status>
              · {{ isTrialing ? 'ends' : 'renews' }} on {{ formatDate(subscription!.currentPeriodEnd) }}
            </span>
          </div>
          <Button v-if="isAdmin" variant="outline" class="shrink-0" @click="changePlan">
            Change plan
          </Button>
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

      <template v-if="hasPlan">
        <div class="-mx-4 border-t" />

        <QueryError
          v-if="usageError && !usage"
          title="Couldn't load this period's total"
          :error="usageError"
          @retry="refetchUsage()"
        />
        <section v-else aria-label="This period" class="flex flex-col gap-0.5">
          <span class="text-sm text-muted-foreground">Total so far</span>
          <span v-if="usage" class="text-xl font-semibold tabular-nums">{{ formatCents(usage.estimatedTotalCents) }}</span>
          <Skeleton v-else-if="usageLoading" class="my-1.5 h-5 w-20" />
          <RouterLink
            :to="{ name: 'usage' }"
            class="mt-1.5 inline-flex items-center gap-1.5 self-start text-sm font-medium text-foreground hover:text-muted-foreground"
          >
            See usage and invoice preview
            <ArrowRight :size="14" />
          </RouterLink>
        </section>
      </template>
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

  </ContextDrawer>
</template>
