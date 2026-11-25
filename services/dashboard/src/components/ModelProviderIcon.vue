<script setup lang="ts">
import { computed } from 'vue';
import { Server } from '@lucide/vue';
import { brandMarks } from '@/lib/brandMarks';
import { cn } from '@/lib/utils';
import { modelProviderBrand, type ModelProviderBrand, type ModelProviderIdentity } from '@/lib/agents';

const props = withDefaults(defineProps<{
  brand?: ModelProviderBrand;
  provider?: ModelProviderIdentity;
  size?: 'sm' | 'md';
}>(), {
  brand: undefined,
  provider: undefined,
  size: 'md',
});

const resolved = computed<ModelProviderBrand>(() => props.brand ?? (props.provider ? modelProviderBrand(props.provider) : 'generic'));
const mark = computed(() => (resolved.value === 'generic' ? null : brandMarks[resolved.value] ?? null));
</script>

<template>
  <span
    :class="cn(
      'flex shrink-0 items-center justify-center rounded-sm',
      size === 'sm' ? 'size-5' : 'size-8',
      !mark && 'bg-muted text-muted-foreground',
    )"
    :style="mark ? { background: mark.background, color: mark.color } : undefined"
    aria-hidden="true"
  >
    <svg
      v-if="mark"
      :viewBox="mark.viewBox ?? '0 0 24 24'"
      :class="size === 'sm' ? 'size-3' : 'size-4.5'"
      fill="currentColor"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path :d="mark.path" />
    </svg>
    <Server v-else :class="size === 'sm' ? 'size-3' : 'size-4.5'" />
  </span>
</template>
