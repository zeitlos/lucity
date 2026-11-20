<script setup lang="ts">
import { computed } from 'vue';
import { ArrowUp, CreditCard } from '@lucide/vue';
import type { Plan } from '@/gql/graphql';
import { planTitle } from '@/lib/capabilities';

const props = defineProps<{
  plan?: Plan | null;
  card?: boolean;
  label?: string;
  inverse?: boolean;
}>();

const text = computed(() => props.label ?? (props.card ? 'Add card' : props.plan ? planTitle(props.plan) : ''));
</script>

<template>
  <span
    class="inline-flex shrink-0 items-center gap-1 text-xs font-medium"
    :class="inverse ? 'bg-white/20 px-1.5 py-0.5 text-white' : 'text-violet-600 dark:text-violet-400'"
  >
    <CreditCard v-if="card" :size="12" :stroke-width="2.25" class="size-3!" />
    <ArrowUp v-else :size="12" :stroke-width="2.25" class="size-3!" />
    {{ text }}
  </span>
</template>
