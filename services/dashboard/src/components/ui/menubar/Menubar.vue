<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { MenubarRoot, type MenubarRootEmits, type MenubarRootProps, useForwardPropsEmits } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<MenubarRootProps & { class?: HTMLAttributes['class'] }>();
const emits = defineEmits<MenubarRootEmits>();

const forwarded = useForwardPropsEmits(props, emits);
</script>

<template>
  <MenubarRoot v-bind="forwarded" :class="cn('flex h-10 items-center gap-1 rounded-md border bg-background p-1', props.class)">
    <slot />
  </MenubarRoot>
</template>
