<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/lib/utils';

const props = defineProps<{ class?: HTMLAttributes['class']; colspan?: number }>();
</script>

<template>
  <tr>
    <td :colspan="colspan" :class="cn('p-4 text-center text-muted-foreground', props.class)">
      <slot />
    </td>
  </tr>
</template>
