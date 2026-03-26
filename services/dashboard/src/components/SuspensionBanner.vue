<script setup lang="ts">
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { useQuery } from '@vue/apollo-composable';
import { AlertTriangle, Sparkles } from 'lucide-vue-next';
import { SubscriptionQuery } from '@/graphql/billing';

const router = useRouter();
const { result: subResult } = useQuery(SubscriptionQuery, null, { fetchPolicy: 'cache-and-network' });

const isTrial = computed(() => !subResult.value?.subscription?.plan);

function goToBilling() {
  router.push({ name: 'workspace-settings', query: { tab: 'billing' } });
}
</script>

<template>
  <div class="flex items-center justify-center gap-2 bg-destructive px-4 py-2 text-xs font-medium text-destructive-foreground">
    <AlertTriangle :size="14" class="shrink-0" />
    <template v-if="isTrial">
      <span>Your free credits have been used up. Upgrade to a plan to resume your workspace.</span>
      <button
        class="inline-flex items-center gap-1 rounded-full bg-destructive-foreground/15 px-2.5 py-0.5 text-xs font-medium transition-colors hover:bg-destructive-foreground/25"
        @click="goToBilling"
      >
        <Sparkles :size="12" />
        Upgrade
      </button>
    </template>
    <template v-else>
      <span>Your workspace has been suspended due to a payment issue. Builds, deploys, and scaling are disabled.</span>
      <button
        class="inline-flex items-center gap-1 rounded-full bg-destructive-foreground/15 px-2.5 py-0.5 text-xs font-medium transition-colors hover:bg-destructive-foreground/25"
        @click="goToBilling"
      >
        Fix payment
      </button>
    </template>
  </div>
</template>
