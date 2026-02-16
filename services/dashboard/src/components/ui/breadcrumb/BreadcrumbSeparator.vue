<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { ChevronRight } from 'lucide-vue-next';
import { cn } from '@/lib/utils';

const props = defineProps<{ class?: HTMLAttributes['class'] }>();
</script>

<template>
  <li :class="cn('[&>svg]:h-3.5 [&>svg]:w-3.5', props.class)" role="presentation" aria-hidden="true">
    <slot>
      <ChevronRight />
    </slot>
  </li>
</template>
