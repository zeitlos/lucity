<script setup lang="ts">
import { useMutation } from '@vue/apollo-composable';
import { AlertTriangle } from 'lucide-vue-next';
import { BillingPortalUrlMutation } from '@/graphql/billing';
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { toast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const { mutate: portalMutate, loading } = useMutation(BillingPortalUrlMutation);

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
  <Alert variant="destructive" class="mt-3">
    <AlertTriangle class="h-4 w-4" />
    <AlertTitle>Workspace suspended</AlertTitle>
    <AlertDescription class="flex items-center justify-between">
      <span>Your workspace has been suspended due to a payment issue. Builds, deploys, and scaling are disabled.</span>
      <Button
        variant="destructive"
        size="sm"
        :disabled="loading"
        @click="openBillingPortal"
      >
        Fix payment
      </Button>
    </AlertDescription>
  </Alert>
</template>
