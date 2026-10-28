<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/lib/utils';

const props = defineProps<{ class?: HTMLAttributes['class'] }>();
</script>

<template>
  <td :class="cn('px-3 py-2 align-middle first:pl-4 last:pr-4 [&:has([role=checkbox])]:pr-0', props.class)">
    <slot />
  </td>
</template>
