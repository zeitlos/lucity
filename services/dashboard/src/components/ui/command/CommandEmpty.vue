<script setup lang="ts">
import { ComboboxEmpty, type ComboboxEmptyProps } from 'reka-ui';

const props = defineProps<ComboboxEmptyProps>();
</script>

<template>
  <ComboboxEmpty v-bind="props" class="py-6 text-center text-sm">
    <slot />
  </ComboboxEmpty>
</template>
