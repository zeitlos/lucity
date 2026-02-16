<script setup lang="ts">
import { type DrawerRootEmits, type DrawerRootProps, DrawerRoot } from 'vaul-vue';
import { useForwardPropsEmits } from 'reka-ui';

const props = withDefaults(defineProps<DrawerRootProps>(), {
  shouldScaleBackground: true,
});
const emits = defineEmits<DrawerRootEmits>();

const forwarded = useForwardPropsEmits(props, emits);
</script>

<template>
  <DrawerRoot v-bind="forwarded">
    <slot />
  </DrawerRoot>
</template>
