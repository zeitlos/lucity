<script setup lang="ts">
import { computed } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Clock, CreditCard } from 'lucide-vue-next';
import { SubscriptionQuery, UsageSummaryQuery, BillingPortalUrlMutation } from '@/graphql/billing';
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { toast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const { result: subResult } = useQuery(SubscriptionQuery, null, { fetchPolicy: 'cache-and-network' });
const { result: usageResult } = useQuery(UsageSummaryQuery, null, { fetchPolicy: 'cache-and-network' });
const { mutate: portalMutate, loading: openingPortal } = useMutation(BillingPortalUrlMutation);

const subscription = computed(() => subResult.value?.subscription);
const usage = computed(() => usageResult.value?.usageSummary);
const hasPaymentMethod = computed(() => subscription.value?.hasPaymentMethod ?? false);
const hasCreditExpiry = computed(() => !!subscription.value?.creditExpiry);
const showBadge = computed(() => hasCreditExpiry.value && !hasPaymentMethod.value);

const daysRemaining = computed(() => {
  if (!subscription.value?.creditExpiry) return 0;
  const end = new Date(subscription.value.creditExpiry).getTime();
  const now = Date.now();
  return Math.max(0, Math.ceil((end - now) / 86_400_000));
});

const creditTotalCents = computed(() => subscription.value?.creditAmountCents ?? 0);
const creditsUsedCents = computed(() => usage.value?.creditsCents ?? 0);
const creditsRemainingCents = computed(() => Math.max(0, creditTotalCents.value - creditsUsedCents.value));

const creditsRemainingFormatted = computed(() => `€${(creditsRemainingCents.value / 100).toFixed(2)}`);
const creditTotalFormatted = computed(() => `€${(creditTotalCents.value / 100).toFixed(2)}`);

const creditsPercent = computed(() => {
  if (creditTotalCents.value === 0) return 100;
  return Math.round((creditsRemainingCents.value / creditTotalCents.value) * 100);
});

const daysPercent = computed(() => {
  // Credit grant is 14 days, calculate percentage remaining.
  return Math.round((daysRemaining.value / 14) * 100);
});

// Show the more urgent constraint in the badge.
const badgeText = computed(() => {
  const days = daysRemaining.value;
  const credits = creditsRemainingFormatted.value;
  return `${days}d or ${credits} left`;
});

// Urgency: green > 7 days & > 50% credits, yellow > 3 days & > 20%, red otherwise.
const urgency = computed(() => {
  if (daysRemaining.value <= 3 || creditsPercent.value <= 20) return 'danger';
  if (daysRemaining.value <= 7 || creditsPercent.value <= 50) return 'warn';
  return 'ok';
});

const urgencyClasses = computed(() => {
  switch (urgency.value) {
    case 'danger': return 'border-destructive/40 text-destructive';
    case 'warn': return 'border-yellow-500/40 text-yellow-600 dark:text-yellow-400';
    default: return 'border-primary/40 text-primary';
  }
});

const barColor = computed(() => {
  switch (urgency.value) {
    case 'danger': return 'bg-destructive';
    case 'warn': return 'bg-yellow-500';
    default: return 'bg-primary';
  }
});

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
</script>

<template>
  <Popover v-if="showBadge">
    <PopoverTrigger as-child>
      <button
        class="flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium transition-colors hover:bg-accent/50"
        :class="urgencyClasses"
      >
        <Clock :size="12" />
        {{ badgeText }}
      </button>
    </PopoverTrigger>
    <PopoverContent align="end" class="w-72 p-0">
      <div class="space-y-3 p-4">
        <!-- Days remaining -->
        <div class="space-y-1.5">
          <div class="flex items-center justify-between text-sm">
            <span class="text-muted-foreground">Credits expire in</span>
            <span class="font-medium">{{ daysRemaining }} days</span>
          </div>
          <div class="h-1.5 w-full overflow-hidden rounded-full bg-secondary">
            <div
              class="h-full rounded-full transition-all"
              :class="barColor"
              :style="{ width: `${daysPercent}%` }"
            />
          </div>
        </div>

        <!-- Credits remaining -->
        <div class="space-y-1.5">
          <div class="flex items-center justify-between text-sm">
            <span class="text-muted-foreground">Credits remaining</span>
            <span class="font-medium">{{ creditsRemainingFormatted }}</span>
          </div>
          <div class="h-1.5 w-full overflow-hidden rounded-full bg-secondary">
            <div
              class="h-full rounded-full transition-all"
              :class="barColor"
              :style="{ width: `${creditsPercent}%` }"
            />
          </div>
          <p class="text-xs text-muted-foreground">
            {{ creditTotalFormatted }} included with your plan
          </p>
        </div>
      </div>

      <Separator />

      <div class="p-4">
        <p class="text-sm font-medium">
          Free Credits
        </p>
        <p class="mt-1 text-xs text-muted-foreground">
          Your credits expire in {{ daysRemaining }} days. Add a payment method to continue using the platform.
        </p>

        <Button
          class="mt-3 w-full"
          size="sm"
          :disabled="openingPortal"
          @click="openBillingPortal"
        >
          <CreditCard :size="14" class="mr-1.5" />
          Add payment method
        </Button>
      </div>
    </PopoverContent>
  </Popover>
</template>
