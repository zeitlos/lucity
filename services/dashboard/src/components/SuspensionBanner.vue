<script setup lang="ts">
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { AlertTriangle, Sparkles } from 'lucide-vue-next';
import { SubscriptionQuery, BillingPortalUrlMutation } from '@/graphql/billing';
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const router = useRouter();
const { result: subResult } = useQuery(SubscriptionQuery, null, { fetchPolicy: 'cache-and-network' });
const { mutate: portalMutate, loading: openingPortal } = useMutation(BillingPortalUrlMutation);

const isTrial = computed(() => !subResult.value?.subscription?.plan);

async function openBillingPortal() {
  try {
    const result = await portalMutate();
    const url = result?.data?.billingPortalUrl?.url;
    if (url) {
      window.open(url, '_blank');
    }
  } catch (e: unknown) {
    errorToast('Failed to open billing portal', { description: errorMessage(e) });
  }
}

function goToBilling() {
  router.push({ name: 'workspace-settings', query: { tab: 'billing' } });
}
</script>

<template>
  <Alert variant="destructive" class="mt-3">
    <AlertTriangle class="h-4 w-4" />
    <AlertTitle>Workspace suspended</AlertTitle>
    <AlertDescription class="flex items-center justify-between">
      <template v-if="isTrial">
        <span>Your free credits have been used up. Upgrade to a plan to resume your workspace.</span>
        <Button
          variant="destructive"
          size="sm"
          @click="goToBilling"
        >
          <Sparkles :size="14" class="mr-1.5" />
          Upgrade to a plan
        </Button>
      </template>
      <template v-else>
        <span>Your workspace has been suspended due to a payment issue. Builds, deploys, and scaling are disabled.</span>
        <Button
          variant="destructive"
          size="sm"
          :disabled="openingPortal"
          @click="openBillingPortal"
        >
          Fix payment
        </Button>
      </template>
    </AlertDescription>
  </Alert>
</template>
