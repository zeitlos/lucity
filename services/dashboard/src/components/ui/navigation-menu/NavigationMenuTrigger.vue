<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { NavigationMenuTrigger, type NavigationMenuTriggerProps } from 'reka-ui';
import { ChevronDown } from 'lucide-vue-next';
import { cn } from '@/lib/utils';

const navigationMenuTriggerStyle = 'group inline-flex h-10 w-max items-center justify-center rounded-md bg-background px-4 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground focus:outline-none disabled:pointer-events-none disabled:opacity-50 data-[active]:bg-accent/50 data-[state=open]:bg-accent/50';

const props = defineProps<NavigationMenuTriggerProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <NavigationMenuTrigger
    v-bind="props"
    :class="cn(navigationMenuTriggerStyle, 'group', props.class)"
  >
    <slot />
    <ChevronDown class="relative top-px ml-1 h-3 w-3 transition duration-200 group-data-[state=open]:rotate-180" aria-hidden="true" />
  </NavigationMenuTrigger>
</template>
