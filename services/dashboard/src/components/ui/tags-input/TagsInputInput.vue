<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { TagsInputInput, type TagsInputInputProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<TagsInputInputProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <TagsInputInput
    v-bind="props"
    :class="cn('flex-1 bg-transparent outline-none placeholder:text-muted-foreground', props.class)"
  />
</template>
