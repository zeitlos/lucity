<script setup lang="ts">
import { AlertDialogRoot, type AlertDialogRootEmits, type AlertDialogRootProps, useForwardPropsEmits } from 'reka-ui';

const props = defineProps<AlertDialogRootProps>();
const emits = defineEmits<AlertDialogRootEmits>();

const forwarded = useForwardPropsEmits(props, emits);
</script>

<template>
  <AlertDialogRoot v-bind="forwarded">
    <slot />
  </AlertDialogRoot>
</template>
