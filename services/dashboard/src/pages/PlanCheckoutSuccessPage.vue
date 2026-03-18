<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { CompletePlanCheckoutMutation, SubscriptionQuery } from '@/graphql/billing';
import { apolloClient } from '@/lib/apollo';
import { Loader2, AlertCircle } from 'lucide-vue-next';
import { Button } from '@/components/ui/button';
import { errorMessage } from '@/lib/utils';

const route = useRoute();
const router = useRouter();

const error = ref('');
const { mutate } = useMutation(CompletePlanCheckoutMutation);

onMounted(async () => {
  const sessionId = route.query.session_id as string;
  if (!sessionId) {
    error.value = 'Missing session ID. Please try again from workspace settings.';
    return;
  }

  try {
    const res = await mutate({ sessionId });

    if (res?.errors?.length) {
      error.value = res.errors.map(e => e.message).join(', ');
      return;
    }

    if (!res?.data?.completePlanCheckout) {
      error.value = 'Failed to activate plan. Please try again.';
      return;
    }

    // Refetch subscription data so the billing page reflects the new plan.
    await apolloClient.refetchQueries({ include: [SubscriptionQuery] });
    router.push({ name: 'workspace-settings' });
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
      <Button @click="router.push({ name: 'workspace-settings' })">
        Back to settings
      </Button>
    </div>
    <div
      v-else
      class="space-y-4 text-center"
    >
      <Loader2
        :size="32"
        class="mx-auto animate-spin text-muted-foreground"
      />
      <p class="text-sm text-muted-foreground">Activating your plan...</p>
    </div>
  </div>
</template>
