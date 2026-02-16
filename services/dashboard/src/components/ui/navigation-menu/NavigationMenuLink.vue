<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { NavigationMenuLink, type NavigationMenuLinkEmits, type NavigationMenuLinkProps, useForwardPropsEmits } from 'reka-ui';
import { cn } from '@/lib/utils';

const navigationMenuTriggerStyle = 'group inline-flex h-10 w-max items-center justify-center rounded-md bg-background px-4 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground focus:outline-none disabled:pointer-events-none disabled:opacity-50 data-[active]:bg-accent/50 data-[state=open]:bg-accent/50';

const props = defineProps<NavigationMenuLinkProps & { class?: HTMLAttributes['class'] }>();
const emits = defineEmits<NavigationMenuLinkEmits>();

const forwarded = useForwardPropsEmits(props, emits);
</script>

<template>
  <NavigationMenuLink v-bind="forwarded" :class="cn(navigationMenuTriggerStyle, props.class)">
    <slot />
  </NavigationMenuLink>
</template>
