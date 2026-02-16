<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { AvatarImage, type AvatarImageProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<AvatarImageProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <AvatarImage v-bind="props" :class="cn('aspect-square h-full w-full', props.class)" />
</template>
