<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { NavigationMenuList, type NavigationMenuListProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<NavigationMenuListProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <NavigationMenuList :class="cn('group flex flex-1 list-none items-center justify-center gap-1', props.class)">
    <slot />
  </NavigationMenuList>
</template>
