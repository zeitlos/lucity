<script setup lang="ts">
import { SplitterPanel, type SplitterPanelProps } from 'reka-ui';

const props = defineProps<SplitterPanelProps>();
</script>

<template>
  <SplitterPanel v-bind="props">
    <slot />
  </SplitterPanel>
</template>
