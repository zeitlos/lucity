<script setup lang="ts">
import { ContextMenuTrigger, type ContextMenuTriggerProps } from 'reka-ui';

const props = defineProps<ContextMenuTriggerProps>();
</script>

<template>
  <ContextMenuTrigger v-bind="props">
    <slot />
  </ContextMenuTrigger>
</template>
