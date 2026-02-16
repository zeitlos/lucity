<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { SelectSeparator, type SelectSeparatorProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<SelectSeparatorProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <SelectSeparator :class="cn('-mx-1 my-1 h-px bg-muted', props.class)" />
</template>
