<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { PaginationNext, type PaginationNextProps } from 'reka-ui';
import { ChevronRight } from 'lucide-vue-next';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

const props = defineProps<PaginationNextProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <PaginationNext v-bind="props" as-child>
    <Button :class="cn('h-10 w-10 p-0', props.class)" variant="outline">
      <ChevronRight class="h-4 w-4" />
      <span class="sr-only">Next page</span>
    </Button>
  </PaginationNext>
</template>
