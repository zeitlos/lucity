<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { PaginationPrev, type PaginationPrevProps } from 'reka-ui';
import { ChevronLeft } from 'lucide-vue-next';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

const props = defineProps<PaginationPrevProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <PaginationPrev v-bind="props" as-child>
    <Button :class="cn('h-10 w-10 p-0', props.class)" variant="outline">
      <ChevronLeft class="h-4 w-4" />
      <span class="sr-only">Previous page</span>
    </Button>
  </PaginationPrev>
</template>
