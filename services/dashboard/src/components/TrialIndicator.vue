<script setup lang="ts">
import { formatMoney } from '@/lib/format';
import { computed, ref } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { useRouter } from 'vue-router';
import { Clock, Sparkles, AlertTriangle } from '@lucide/vue';
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { graphql } from '@/gql';
import { SubscriptionStatus } from '@/gql/graphql';

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

const router = useRouter();
const { result: subResult } = useQuery(SubscriptionDocument, null, { fetchPolicy: 'cache-and-network' });
const { result: usageResult } = useQuery(UsageSummaryDocument, null, { fetchPolicy: 'cache-and-network' });

const popoverOpen = ref(false);

const subscription = computed(() => subResult.value?.subscription);
const usage = computed(() => usageResult.value?.usageSummary);
const showBadge = computed(() => subscription.value?.status === SubscriptionStatus.Trialing);

const daysRemaining = computed(() => {
  if (!subscription.value?.creditExpiry) return 0;
  const end = new Date(subscription.value.creditExpiry).getTime();
  const now = Date.now();
  return Math.max(0, Math.ceil((end - now) / 86_400_000));
});

const creditTotalCents = computed(() => subscription.value?.creditAmountCents ?? 0);
const creditsUsedCents = computed(() => usage.value?.creditsCents ?? 0);
const creditsRemainingCents = computed(() => Math.max(0, creditTotalCents.value - creditsUsedCents.value));

const creditsRemainingFormatted = computed(() => formatMoney(creditsRemainingCents.value / 100, 'EUR'));
const creditTotalFormatted = computed(() => formatMoney(creditTotalCents.value / 100, 'EUR'));

const creditsExhausted = computed(() => creditsRemainingCents.value <= 0);
const daysExpired = computed(() => daysRemaining.value <= 0);
const willSuspend = computed(() => creditsExhausted.value || daysExpired.value);

const creditsPercent = computed(() => {
  if (creditTotalCents.value === 0) return 0;
  return Math.round((creditsRemainingCents.value / creditTotalCents.value) * 100);
});

const daysPercent = computed(() => {
  return Math.round((daysRemaining.value / 14) * 100);
});

const badgeText = computed(() => {
  if (willSuspend.value) return 'Suspension pending';
  const days = daysRemaining.value;
  const credits = creditsRemainingFormatted.value;
  return `${days}d or ${credits} left`;
});

// Urgency: green > 7 days & > 50% credits, yellow > 3 days & > 20%, red otherwise.
const urgency = computed(() => {
  if (willSuspend.value) return 'danger';
  if (daysRemaining.value <= 3 || creditsPercent.value <= 20) return 'danger';
  if (daysRemaining.value <= 7 || creditsPercent.value <= 50) return 'warn';
  return 'ok';
});

const urgencyClasses = computed(() => {
  switch (urgency.value) {
    case 'danger': return 'border-status-danger/40 text-status-danger';
    case 'warn': return 'border-status-warn/40 text-status-warn';
    default: return 'border-status-ok/40 text-status-ok';
  }
});

const barColor = computed(() => {
  switch (urgency.value) {
    case 'danger': return 'bg-status-danger';
    case 'warn': return 'bg-status-warn';
    default: return 'bg-status-ok';
  }
});

function goToBilling() {
  popoverOpen.value = false;
  router.push({ name: 'workspace', params: { drawer: 'billing' } });
}
</script>

<template>
  <Popover v-if="showBadge" v-model:open="popoverOpen">
    <PopoverTrigger as-child>
      <button
        class="flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium transition-colors hover:bg-accent/50"
        :class="urgencyClasses"
      >
        <AlertTriangle v-if="willSuspend" :size="12" />
        <Clock v-else :size="12" />
        {{ badgeText }}
      </button>
    </PopoverTrigger>
    <PopoverContent align="end" class="w-72 p-0">
      <div class="space-y-3 p-4">
        <!-- Suspension warning -->
        <div v-if="willSuspend" class="rounded-md bg-destructive/10 p-2.5">
          <p class="text-xs font-medium text-destructive">
            Your workspace will be suspended shortly. Upgrade to a plan to keep it running.
          </p>
        </div>

        <template v-if="!willSuspend">
          <!-- Days remaining -->
          <div class="space-y-1.5">
            <div class="flex items-center justify-between text-sm">
              <span class="text-muted-foreground">Credits expire in</span>
              <span class="font-medium">{{ daysRemaining }} {{ daysRemaining === 1 ? 'day' : 'days' }}</span>
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
        </template>
      </div>

      <Separator />

      <div class="p-4">
        <p class="text-sm font-medium">
          {{ willSuspend ? 'Trial ended' : 'Free credits' }}
        </p>
        <p class="mt-1 text-xs text-muted-foreground">
          <template v-if="willSuspend">
            {{ creditsExhausted ? 'Your free credits have been used up.' : 'Your free credits have expired.' }}
            Upgrade to a plan to keep your workspace running.
          </template>
          <template v-else>
            Your credits expire in {{ daysRemaining }} {{ daysRemaining === 1 ? 'day' : 'days' }}. Upgrade to a plan to continue using the platform.
          </template>
        </p>

        <Button
          class="mt-3 w-full"
          @click="goToBilling"
        >
          <Sparkles />
          Upgrade to a plan
        </Button>
      </div>
    </PopoverContent>
  </Popover>
</template>
