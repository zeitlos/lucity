<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { AlertCircle } from '@lucide/vue';
import { graphql } from '@/gql';
import Spinner from '@/components/LoadingSpinner.vue';
import { Button } from '@/components/ui/button';
import { CapabilitiesDocument } from '@/composables/useCapabilities';
import { gateFromError, takePendingCreate, useGate } from '@/composables/useGate';
import { apolloClient } from '@/lib/apollo';
import { errorMessage } from '@/lib/utils';

const CompletePaymentMethodCheckoutDocument = graphql(`
  mutation CompletePaymentMethodCheckout($sessionId: String!) {
    completePaymentMethodCheckout(sessionId: $sessionId) {
      id
    }
  }
`);

const route = useRoute();
const router = useRouter();
const { open } = useGate();

const error = ref('');
const { mutate } = useMutation(CompletePaymentMethodCheckoutDocument);

function returnPath(): string {
  const pending = takePendingCreate();
  const path = (route.query.return as string | undefined) ?? pending ?? '/';
  return path.startsWith('/') && !path.startsWith('//') ? path : '/';
}

onMounted(async () => {
  const sessionId = route.query.session_id as string | undefined;
  const back = returnPath();

  if (!sessionId) {
    error.value = 'The card checkout came back without a session. Please add the card again.';
    return;
  }

  try {
    await mutate({ sessionId });
    await apolloClient.refetchQueries({ include: [CapabilitiesDocument] });
    router.replace(back);
  } catch (e: unknown) {
    const gate = gateFromError(e);

    if (gate?.kind === 'card') {
      await router.replace(back);
      open({ ...gate, returnPath: back });
      return;
    }

    error.value = errorMessage(e);
  }
});
</script>

<template>
  <div class="flex min-h-screen items-center justify-center">
    <div v-if="error" class="max-w-md space-y-4 text-center">
      <AlertCircle :size="32" class="mx-auto text-destructive" />
      <h1 class="text-lg font-medium">Something went wrong</h1>
      <p class="text-sm text-muted-foreground">{{ error }}</p>
      <Button @click="router.replace('/')">Back to the dashboard</Button>
    </div>
    <div v-else class="space-y-4 text-center">
      <Spinner :size="32" class="mx-auto text-muted-foreground" />
      <p class="text-sm text-muted-foreground">Saving your card…</p>
    </div>
  </div>
</template>
