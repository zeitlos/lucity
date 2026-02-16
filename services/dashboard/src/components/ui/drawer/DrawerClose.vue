<script setup lang="ts">
import { type DrawerCloseProps, DrawerClose } from 'vaul-vue';

const props = defineProps<DrawerCloseProps>();
</script>

<template>
  <DrawerClose v-bind="props">
    <slot />
  </DrawerClose>
</template>
