<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { graphql } from '@/gql';

const CompletePlanCheckoutDocument = graphql(`
  mutation CompletePlanCheckout($sessionId: String!) {
    completePlanCheckout(sessionId: $sessionId) {
      plan
      status
      currentPeriodEnd
      creditAmountCents
      hasPaymentMethod
    }
  }
`);

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
import { apolloClient } from '@/lib/apollo';
import { AlertCircle } from '@lucide/vue';
import Spinner from '@/components/LoadingSpinner.vue';
import { Button } from '@/components/ui/button';
import { errorMessage } from '@/lib/utils';

const route = useRoute();
const router = useRouter();

const error = ref('');
const { mutate } = useMutation(CompletePlanCheckoutDocument);

onMounted(async () => {
  const sessionId = route.query.session_id as string;
  if (!sessionId) {
    error.value = 'Missing session ID. Please try again from workspace settings.';
    return;
  }

  try {
    const res = await mutate({ sessionId });

    if (!res?.data?.completePlanCheckout) {
      error.value = 'Failed to activate plan. Please try again.';
      return;
    }

    // Refetch subscription data so the billing page reflects the new plan.
    await apolloClient.refetchQueries({ include: [SubscriptionDocument] });
    router.push({ name: 'workspace', params: { drawer: 'billing' } });
  } catch (e: unknown) {
    error.value = errorMessage(e);
  }
});
</script>

<template>
  <div class="flex min-h-screen items-center justify-center">
    <div
      v-if="error"
      class="max-w-md space-y-4 text-center"
    >
      <AlertCircle
        :size="32"
        class="mx-auto text-destructive"
      />
      <h1 class="text-lg font-medium">Something went wrong</h1>
      <p class="text-sm text-muted-foreground">{{ error }}</p>
      <Button @click="router.push({ name: 'workspace', params: { drawer: 'billing' } })">
        Back to billing
      </Button>
    </div>
    <div
      v-else
      class="space-y-4 text-center"
    >
      <Spinner
        :size="32"
        class="mx-auto animate-spin text-muted-foreground"
      />
      <p class="text-sm text-muted-foreground">Activating your plan...</p>
    </div>
  </div>
</template>
