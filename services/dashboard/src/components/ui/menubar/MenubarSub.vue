<script setup lang="ts">
import { MenubarSub, type MenubarSubEmits, type MenubarSubProps, useForwardPropsEmits } from 'reka-ui';

const props = defineProps<MenubarSubProps>();
const emits = defineEmits<MenubarSubEmits>();

const forwarded = useForwardPropsEmits(props, emits);
</script>

<template>
  <MenubarSub v-bind="forwarded">
    <slot />
  </MenubarSub>
</template>
