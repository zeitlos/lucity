<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { TagsInputItem, type TagsInputItemProps, TagsInputItemDelete, TagsInputItemText } from 'reka-ui';
import { X } from 'lucide-vue-next';
import { cn } from '@/lib/utils';

const props = defineProps<TagsInputItemProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <TagsInputItem
    v-bind="props"
    :class="cn('inline-flex items-center gap-1.5 rounded-sm bg-secondary px-1.5 py-0.5 text-sm', props.class)"
  >
    <TagsInputItemText />
    <TagsInputItemDelete class="rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2">
      <X class="h-3 w-3" />
    </TagsInputItemDelete>
  </TagsInputItem>
</template>
